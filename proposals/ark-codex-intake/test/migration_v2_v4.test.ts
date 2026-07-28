import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { DatabaseSync } from 'node:sqlite';
import {
  ArkBoundaryStore,
  BoundarySecrets,
  LeaseRegistry,
  canonicalJson,
  sha256Canonical,
} from '../src/boundary';

const secrets: BoundarySecrets = {
  authSecret: 'migration-auth-secret',
  signingSecret: 'migration-signing-secret',
  receiptSecret: 'migration-receipt-secret',
  registrySecret: 'migration-registry-secret',
  activeKeyId: 'migration-key-v1',
  registryKeyId: 'migration-registry-key-v1',
};

function createV2Database(databasePath: string) {
  const db = new DatabaseSync(databasePath);
  db.exec(`
    PRAGMA foreign_keys = ON;
    CREATE TABLE trust_keys (
      key_id TEXT NOT NULL,
      purpose TEXT NOT NULL,
      revoked INTEGER NOT NULL DEFAULT 0 CHECK (revoked IN (0, 1)),
      created_at TEXT NOT NULL,
      PRIMARY KEY (key_id, purpose)
    );
    CREATE TABLE matrix_obligations (
      obligation_id TEXT PRIMARY KEY,
      source_node_id TEXT NOT NULL,
      target_class TEXT NOT NULL,
      mediator TEXT NOT NULL,
      current_state TEXT NOT NULL,
      op_cost TEXT NOT NULL,
      initiated_at TEXT NOT NULL
    ) STRICT;
    CREATE TABLE closure_receipts (
      receipt_id TEXT PRIMARY KEY,
      obligation_id TEXT NOT NULL UNIQUE,
      resolved_by_node TEXT NOT NULL,
      witness_digest TEXT NOT NULL,
      verifier_algorithm TEXT NOT NULL,
      verifier_key_id TEXT NOT NULL,
      verifier_key_purpose TEXT NOT NULL,
      verifier_mac TEXT NOT NULL,
      propagation_hop_limit INTEGER NOT NULL,
      resolution_state TEXT NOT NULL,
      finite_cost INTEGER NOT NULL,
      receipt_payload_json TEXT NOT NULL,
      recorded_at TEXT NOT NULL,
      FOREIGN KEY (obligation_id) REFERENCES matrix_obligations(obligation_id)
        ON DELETE RESTRICT ON UPDATE RESTRICT,
      FOREIGN KEY (verifier_key_id, verifier_key_purpose)
        REFERENCES trust_keys(key_id, purpose)
        ON DELETE RESTRICT ON UPDATE RESTRICT
    ) STRICT;
    CREATE TRIGGER matrix_obligation_identity_immutable
    BEFORE UPDATE OF obligation_id ON matrix_obligations
    BEGIN
      SELECT RAISE(ABORT, 'OLD_TRIGGER');
    END;
    CREATE TRIGGER closure_receipt_applies_transition
    AFTER INSERT ON closure_receipts
    BEGIN
      SELECT 1;
    END;
    PRAGMA user_version = 2;
  `);
  const recordedAt = '2026-07-28T00:00:00.000Z';
  const binding = {
    sourceNodeId: 'Legacy_Node',
    targetClass: 'SWARM_NODE',
    mediator: 'GOSSIP_PROTOCOL',
  };
  const obligationId = sha256Canonical(binding);
  const receiptBody = {
    profile: 'CLOSURE_RECEIPT_v0.1',
    obligationId,
    resolvedByNode: 'Legacy_Resolver',
    witnessDigest: sha256Canonical({ witness: 'legacy' }),
    verifierAlgorithm: 'HMAC-SHA256',
    verifierKeyId: secrets.activeKeyId,
    propagationHopLimit: 1,
    resolutionState: 'RESOLVED_BY_WITNESS',
    finiteCost: 1,
    recordedAt,
    authorityEffect: 'NONE',
  };
  const receiptId = sha256Canonical(receiptBody);
  const receipt = {
    ...receiptBody,
    receiptId,
    verifierMac: 'a'.repeat(64),
  };
  const receiptJson = canonicalJson(receipt);

  const insertKey = db.prepare(
    `INSERT INTO trust_keys (key_id, purpose, revoked, created_at)
     VALUES (?, ?, 0, ?)`,
  );
  insertKey.run(secrets.activeKeyId, 'LEASE_HMAC', recordedAt);
  insertKey.run(secrets.activeKeyId, 'RECEIPT_HMAC', recordedAt);
  db.prepare(
    `INSERT INTO matrix_obligations
     (obligation_id, source_node_id, target_class, mediator,
      current_state, op_cost, initiated_at)
     VALUES (?, ?, ?, ?, 'RESOLVED_BY_WITNESS', '1', ?)`,
  ).run(
    obligationId,
    binding.sourceNodeId,
    binding.targetClass,
    binding.mediator,
    recordedAt,
  );
  db.prepare(
    `INSERT INTO closure_receipts
     (receipt_id, obligation_id, resolved_by_node, witness_digest,
      verifier_algorithm, verifier_key_id, verifier_key_purpose, verifier_mac,
      propagation_hop_limit, resolution_state, finite_cost,
      receipt_payload_json, recorded_at)
     VALUES (?, ?, ?, ?, ?, ?, 'RECEIPT_HMAC', ?, ?, ?, ?, ?, ?)`,
  ).run(
    receiptId,
    obligationId,
    receipt.resolvedByNode,
    receipt.witnessDigest,
    receipt.verifierAlgorithm,
    receipt.verifierKeyId,
    receipt.verifierMac,
    receipt.propagationHopLimit,
    receipt.resolutionState,
    receipt.finiteCost,
    receiptJson,
    recordedAt,
  );
  db.close();
  return { obligationId, receiptId, receiptJson };
}

test('v2 to v4 migration preserves legacy bytes and restores v4 enforcement', () => {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'ark-migration-'));
  const databasePath = path.join(directory, 'ark.sqlite');
  const legacy = createV2Database(databasePath);
  const store = new ArkBoundaryStore(databasePath, secrets);
  try {
    assert.equal(store.foundationStatus().userVersion, 4);
    assert.equal(store.foundationStatus().requiredTriggersPresent, true);
    const obligation = store.getMatrixObligation(legacy.obligationId);
    assert.equal(obligation.graphHash, null);
    assert.equal(obligation.graphBindingStatus, 'LEGACY_UNBOUND');

    const row = store.db
      .prepare(
        `SELECT graph_hash, graph_binding_status, receipt_payload_json
         FROM closure_receipts WHERE receipt_id = ?`,
      )
      .get(legacy.receiptId) as {
      graph_hash: string | null;
      graph_binding_status: string;
      receipt_payload_json: string;
    };
    assert.equal(row.graph_hash, null);
    assert.equal(row.graph_binding_status, 'LEGACY_UNBOUND');
    assert.equal(row.receipt_payload_json, legacy.receiptJson);

    assert.throws(
      () =>
        store.db
          .prepare(
            `UPDATE matrix_obligations
             SET current_state = 'SUSPENDED_STATE', op_cost = 'INFINITY'
             WHERE obligation_id = ?`,
          )
          .run(legacy.obligationId),
      /LEGACY_OBLIGATION_READ_ONLY/,
    );
    assert.throws(
      () =>
        store.db
          .prepare(`DELETE FROM matrix_obligations WHERE obligation_id = ?`)
          .run(legacy.obligationId),
      /RESOLVED_OBLIGATION_TRACE_IMMUTABLE/,
    );
    assert.throws(
      () =>
        store.db
          .prepare(`DELETE FROM closure_receipts WHERE receipt_id = ?`)
          .run(legacy.receiptId),
      /CLOSURE_RECEIPT_IMMUTABLE/,
    );

    const graphHash = sha256Canonical({ graph: 'v3' });
    const opened = store.openMatrixObligation({
      sourceNodeId: 'New_Node',
      targetClass: 'SWARM_NODE',
      mediator: 'GOSSIP_PROTOCOL',
      graphHash,
    });
    const closureInput = {
      obligationId: opened.obligationId,
      resolvedByNode: 'New_Resolver',
      witnessDigest: sha256Canonical({ witness: 'v3' }),
      resultDigest: sha256Canonical({ result: 'v4' }),
      graphHash,
      propagationHopLimit: 1,
      resolutionState: 'RESOLVED_BY_WITNESS' as const,
      finiteCost: 1,
    };
    const leases = new LeaseRegistry(secrets, store.db);
    const lease = leases.issue({
      type: 'APPLY_WINDOW',
      targetResource: '/api/matrix/obligations/close',
      payloadDigest: sha256Canonical(closureInput),
      scope: 'matrix:closure:apply',
      notBefore: new Date(Date.now() - 1_000).toISOString(),
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
      topologyHash: store.topologyHash(),
    });
    const closed = store.executeOnce(lease, 'MIGRATION_V4_CLOSURE', () =>
      store.applyClosureReceipt(closureInput),
    );
    assert.equal(
      closed.response.obligation.currentState,
      'RESOLVED_BY_WITNESS',
    );
    assert.equal(store.db.prepare('PRAGMA foreign_key_check').all().length, 0);
    assert.equal(
      (
        store.db.prepare('PRAGMA integrity_check').get() as {
          integrity_check: string;
        }
      ).integrity_check,
      'ok',
    );
  } finally {
    store.db.close();
    fs.rmSync(directory, { recursive: true, force: true });
  }
});
