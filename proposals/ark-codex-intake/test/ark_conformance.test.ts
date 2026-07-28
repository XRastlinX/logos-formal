import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import {
  ArkBoundaryStore,
  BoundaryError,
  BoundarySecrets,
  LeaseRegistry,
  canonicalJson,
  sha256Canonical,
} from '../src/boundary';

const secrets: BoundarySecrets = {
  authSecret: 'auth-test-secret',
  signingSecret: 'signing-test-secret',
  receiptSecret: 'receipt-test-secret',
  registrySecret: 'registry-test-secret',
  activeKeyId: 'test-key-v1',
  registryKeyId: 'registry-test-key-v1',
};

function fixture() {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'ark-boundary-'));
  const store = new ArkBoundaryStore(path.join(directory, 'ark.sqlite'), secrets);
  const leases = new LeaseRegistry(secrets, store.db);
  const cleanup = () => {
    store.db.close();
    fs.rmSync(directory, { recursive: true, force: true });
  };
  return { store, leases, cleanup };
}

function issue(
  leases: LeaseRegistry,
  overrides: Partial<Parameters<LeaseRegistry['issue']>[0]> = {},
) {
  const payload = overrides.payloadDigest
    ? { value: 'unused' }
    : { value: 'fixture' };
  return leases.issue({
    type: 'APPLY_WINDOW',
    targetResource: '/api/foundry/apply',
    payloadDigest: overrides.payloadDigest ?? sha256Canonical(payload),
    scope: 'ledger:apply',
    notBefore: new Date(Date.now() - 1_000).toISOString(),
    expiresAt: new Date(Date.now() + 60_000).toISOString(),
    topologyHash: sha256Canonical([]),
    ...overrides,
  });
}

test('canonical JSON is stable across object key order', () => {
  assert.equal(canonicalJson({ b: 2, a: 1 }), canonicalJson({ a: 1, b: 2 }));
  assert.equal(sha256Canonical({ b: 2, a: 1 }), sha256Canonical({ a: 1, b: 2 }));
});

test('lease verification rejects wrong route, type, payload, and expired windows', () => {
  const { leases, cleanup } = fixture();
  try {
    const payloadDigest = sha256Canonical({ value: 'fixture' });
    const lease = issue(leases, { payloadDigest });
    assert.equal(
      leases.verify(lease, {
        type: 'APPLY_WINDOW',
        targetResource: '/api/foundry/apply',
        payloadDigest,
        scope: 'ledger:apply',
      }).id,
      lease.id,
    );
    for (const expected of [
      {
        type: 'INGEST_WINDOW' as const,
        targetResource: '/api/foundry/apply',
        payloadDigest,
        scope: 'ledger:apply',
      },
      {
        type: 'APPLY_WINDOW' as const,
        targetResource: '/api/federation/sync',
        payloadDigest,
        scope: 'ledger:apply',
      },
      {
        type: 'APPLY_WINDOW' as const,
        targetResource: '/api/foundry/apply',
        payloadDigest: sha256Canonical({ value: 'tampered' }),
        scope: 'ledger:apply',
      },
    ]) {
      assert.throws(
        () => leases.verify(lease, expected),
        (error: unknown) =>
          error instanceof BoundaryError && error.message === 'LEASE_BINDING_MISMATCH',
      );
    }
    assert.throws(
      () =>
        leases.verify(
          lease,
          {
            type: 'APPLY_WINDOW',
            targetResource: '/api/foundry/apply',
            payloadDigest,
            scope: 'ledger:apply',
          },
          Date.parse(lease.expiresAt),
        ),
      /LEASE_OUTSIDE_TEMPORAL_WINDOW/,
    );
  } finally {
    cleanup();
  }
});

test('tampered canonical lease bytes fail MAC verification', () => {
  const { leases, cleanup } = fixture();
  try {
    const lease = issue(leases);
    const tampered = { ...lease, issuer: 'SELF_DECLARED' };
    assert.throws(
      () =>
        leases.verify(tampered, {
          type: 'APPLY_WINDOW',
          targetResource: '/api/foundry/apply',
          payloadDigest: lease.payloadDigest,
          scope: 'ledger:apply',
        }),
      /LEASE_MAC_INVALID/,
    );
  } finally {
    cleanup();
  }
});

test('revoked active key cannot issue leases', () => {
  const { store, leases, cleanup } = fixture();
  try {
    const update = store.createRegistryUpdate({
      keyId: secrets.activeKeyId,
      purpose: 'LEASE_HMAC',
      keyVersion: 1,
      revokedAt: new Date().toISOString(),
      reason: 'test revocation',
    });
    const lease = issue(leases, {
      targetResource: '/api/trust/keys/revoke',
      scope: 'trust:key:revoke',
      payloadDigest: sha256Canonical(update),
    });
    const applied = store.executeOnce(lease, 'REGISTRY_KEY_REVOKE', () =>
      store.applyRegistryUpdate(update),
    );
    assert.deepEqual(applied.response, { status: 'APPLIED' });
    assert.throws(() => issue(leases), /ACTIVE_KEY_REVOKED/);
  } finally {
    cleanup();
  }
});

test('registry updates are authenticated, monotonic, and idempotent', () => {
  const { store, leases, cleanup } = fixture();
  try {
    const update = store.createRegistryUpdate({
      keyId: secrets.activeKeyId,
      purpose: 'RECEIPT_HMAC',
      keyVersion: 1,
      revokedAt: new Date().toISOString(),
      reason: 'compromised receipt key',
    });
    const registryLease = (payload: unknown) =>
      issue(leases, {
        targetResource: '/api/trust/keys/revoke',
        scope: 'trust:key:revoke',
        payloadDigest: sha256Canonical({ update: payload }),
      });
    const lease = registryLease(update);
    const first = store.executeOnce(lease, 'REGISTRY_KEY_REVOKE', () =>
      store.applyRegistryUpdate(update),
    );
    const replay = store.executeOnce(lease, 'REGISTRY_KEY_REVOKE', () => ({
      status: 'MUST_NOT_RUN',
    }));
    assert.deepEqual(first.response, { status: 'APPLIED' });
    assert.deepEqual(replay.response, { status: 'APPLIED' });
    assert.equal(replay.replayed, true);

    const conflict = store.createRegistryUpdate({
      keyId: secrets.activeKeyId,
      purpose: 'RECEIPT_HMAC',
      keyVersion: 1,
      revokedAt: new Date().toISOString(),
      reason: 'different statement at the same version',
    });
    assert.throws(
      () =>
        store.executeOnce(
          registryLease(conflict),
          'REGISTRY_KEY_REVOKE',
          () => store.applyRegistryUpdate(conflict),
        ),
      /REGISTRY_UPDATE_CONFLICT/,
    );

    const tampered = { ...conflict, reason: 'tampered after authentication' };
    assert.throws(
      () =>
        store.executeOnce(
          registryLease(tampered),
          'REGISTRY_KEY_REVOKE',
          () => store.applyRegistryUpdate(tampered),
        ),
      /INVALID_UPDATE_ID/,
    );
    assert.throws(
      () =>
        store.db
          .prepare(
            `UPDATE trust_keys
             SET registry_version = 2, last_update_id = ?
             WHERE key_id = ? AND purpose = 'RECEIPT_HMAC'`,
          )
          .run('f'.repeat(64), secrets.activeKeyId),
      /GOVERNED_REGISTRY_UPDATE_REQUIRED/,
    );
    assert.throws(
      () =>
        store.db
          .prepare(`DELETE FROM registry_updates WHERE update_id = ?`)
          .run(update.updateId),
      /REGISTRY_UPDATE_IMMUTABLE/,
    );
  } finally {
    cleanup();
  }
});

test('transaction commits mutation, receipt, response, and nonce exactly once', () => {
  const { store, leases, cleanup } = fixture();
  try {
    const lease = issue(leases);
    let mutationCount = 0;
    const first = store.executeOnce(lease, 'TEST_APPLY', () => {
      mutationCount += 1;
      const entry = { value: 'settled' };
      store.appendLedgerEntry(entry, 'ACTIVE', 'TEST');
      return { ok: true };
    });
    const replay = store.executeOnce(lease, 'TEST_APPLY', () => {
      mutationCount += 1;
      return { ok: false };
    });

    assert.equal(mutationCount, 1);
    assert.equal(first.replayed, false);
    assert.equal(replay.replayed, true);
    assert.deepEqual(replay.response, { ok: true });
    assert.equal(store.verifyReceipt(first.receipt), true);
    const nonce = store.db
      .prepare(`SELECT status, receipt_id FROM consumed_nonces WHERE nonce = ?`)
      .get(lease.nonce) as { status: string; receipt_id: string };
    assert.equal(nonce.status, 'COMMITTED');
    assert.equal(nonce.receipt_id, first.receipt.receiptId);
  } finally {
    cleanup();
  }
});

test('failed local mutation rolls back ledger, receipt, and CLAIMED nonce', () => {
  const { store, leases, cleanup } = fixture();
  try {
    const lease = issue(leases);
    assert.throws(
      () =>
        store.executeOnce(lease, 'TEST_ROLLBACK', () => {
          store.appendLedgerEntry({ value: 'must-roll-back' }, 'ACTIVE', 'TEST');
          throw new Error('fixture failure');
        }),
      /fixture failure/,
    );
    assert.equal(store.listLedger().length, 0);
    const nonceCount = store.db
      .prepare(`SELECT COUNT(*) AS count FROM consumed_nonces WHERE nonce = ?`)
      .get(lease.nonce) as { count: number };
    assert.equal(Number(nonceCount.count), 0);
    const receiptCount = store.db
      .prepare(`SELECT COUNT(*) AS count FROM audit_receipts WHERE nonce = ?`)
      .get(lease.nonce) as { count: number };
    assert.equal(Number(receiptCount.count), 0);
  } finally {
    cleanup();
  }
});

test('federated data remains quarantined until explicit promotion', () => {
  const { store, leases, cleanup } = fixture();
  try {
    const importLease = issue(leases, {
      type: 'FEDERATION_WINDOW',
      targetResource: '/api/federation/sync',
      scope: 'federation:import',
      payloadDigest: sha256Canonical({ batch: 1 }),
    });
    let quarantineHash = '';
    store.executeOnce(importLease, 'FEDERATION_IMPORT', () => {
      quarantineHash = store.appendLedgerEntry(
        { external: true },
        'QUARANTINE',
        'peer-a',
      );
      return { quarantineHash };
    });
    assert.equal(store.listLedger(false).length, 0);
    assert.equal(store.listLedger(true).length, 1);

    const promotionLease = issue(leases, {
      payloadDigest: sha256Canonical({ promote: quarantineHash }),
    });
    store.executeOnce(promotionLease, 'PROMOTE', () => ({
      promoted: store.promoteQuarantine([quarantineHash]),
    }));
    assert.equal(store.listLedger(false).length, 1);
  } finally {
    cleanup();
  }
});

test('topology drift rejects first use without consuming the nonce', () => {
  const { store, leases, cleanup } = fixture();
  try {
    const staleLease = issue(leases);
    store.appendLedgerEntry({ changed: true }, 'ACTIVE', 'TEST');
    assert.throws(
      () => store.executeOnce(staleLease, 'STALE_APPLY', () => ({ ok: true })),
      /TOPOLOGY_DRIFT_ERROR/,
    );
    const nonceCount = store.db
      .prepare(`SELECT COUNT(*) AS count FROM consumed_nonces WHERE nonce = ?`)
      .get(staleLease.nonce) as { count: number };
    assert.equal(Number(nonceCount.count), 0);
  } finally {
    cleanup();
  }
});

test('committed nonce and deterministic response survive database restart', () => {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'ark-restart-'));
  const databasePath = path.join(directory, 'ark.sqlite');
  const firstStore = new ArkBoundaryStore(databasePath, secrets);
  const leases = new LeaseRegistry(secrets, firstStore.db);
  const lease = issue(leases);
  const first = firstStore.executeOnce(lease, 'RESTART_FIXTURE', () => ({
    persisted: true,
  }));
  firstStore.db.close();

  const reopened = new ArkBoundaryStore(databasePath, secrets);
  try {
    const replay = reopened.executeOnce(lease, 'RESTART_FIXTURE', () => ({
      persisted: false,
    }));
    assert.equal(replay.replayed, true);
    assert.deepEqual(replay.response, first.response);
    assert.equal(reopened.verifyReceipt(replay.receipt), true);
  } finally {
    reopened.db.close();
    fs.rmSync(directory, { recursive: true, force: true });
  }
});

test('SQLite BEGIN IMMEDIATE permits only one concurrent writer', () => {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'ark-lock-'));
  const databasePath = path.join(directory, 'ark.sqlite');
  const first = new ArkBoundaryStore(databasePath, secrets);
  const second = new ArkBoundaryStore(databasePath, secrets);
  try {
    second.db.exec('PRAGMA busy_timeout = 1;');
    first.db.exec('BEGIN IMMEDIATE;');
    assert.throws(() => second.db.exec('BEGIN IMMEDIATE;'), /locked/i);
    first.db.exec('ROLLBACK;');
    assert.doesNotThrow(() => second.db.exec('BEGIN IMMEDIATE;'));
    second.db.exec('ROLLBACK;');
  } finally {
    first.db.close();
    second.db.close();
    fs.rmSync(directory, { recursive: true, force: true });
  }
});
