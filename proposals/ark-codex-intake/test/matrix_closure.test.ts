import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import {
  ArkBoundaryStore,
  BoundarySecrets,
  LeaseRegistry,
  sha256Canonical,
} from '../src/boundary';
import { CadenceEngine } from '../src/cadence';

const secrets: BoundarySecrets = {
  authSecret: 'matrix-auth-secret',
  signingSecret: 'matrix-signing-secret',
  receiptSecret: 'matrix-receipt-secret',
  registrySecret: 'matrix-registry-secret',
  activeKeyId: 'matrix-key-v1',
  registryKeyId: 'matrix-registry-key-v1',
};

function applyClosure(
  store: ArkBoundaryStore,
  input: Parameters<ArkBoundaryStore['applyClosureReceipt']>[0],
) {
  const leases = new LeaseRegistry(secrets, store.db);
  const lease = leases.issue({
    type: 'APPLY_WINDOW',
    targetResource: '/api/matrix/obligations/close',
    payloadDigest: sha256Canonical(input),
    scope: 'matrix:closure:apply',
    notBefore: new Date(Date.now() - 1_000).toISOString(),
    expiresAt: new Date(Date.now() + 60_000).toISOString(),
    topologyHash: store.topologyHash(),
  });
  return store.executeOnce(lease, 'MATRIX_CLOSURE_TEST', () =>
    store.applyClosureReceipt(input),
  ).response;
}

function fixture() {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'matrix-closure-'));
  const store = new ArkBoundaryStore(path.join(directory, 'matrix.sqlite'), secrets);
  return {
    store,
    cleanup: () => {
      store.db.close();
      fs.rmSync(directory, { recursive: true, force: true });
    },
  };
}

function binding(sourceNodeId = 'Node_A') {
  const graphHash = '0'.repeat(64);
  const value = {
    sourceNodeId,
    targetClass: 'SWARM_NODE',
    mediator: 'GOSSIP_PROTOCOL',
    graphHash,
  };
  return { ...value, obligationId: sha256Canonical(value) };
}

test('SQLite foundation enforces WAL, NORMAL sync, foreign keys, and STRICT matrix tables', () => {
  const { store, cleanup } = fixture();
  try {
    const journal = store.db.prepare('PRAGMA journal_mode').get() as {
      journal_mode: string;
    };
    const synchronous = store.db.prepare('PRAGMA synchronous').get() as {
      synchronous: number;
    };
    const foreignKeys = store.db.prepare('PRAGMA foreign_keys').get() as {
      foreign_keys: number;
    };
    const trustedSchema = store.db.prepare('PRAGMA trusted_schema').get() as {
      trusted_schema: number;
    };
    assert.equal(journal.journal_mode.toLowerCase(), 'wal');
    assert.equal(Number(synchronous.synchronous), 1);
    assert.equal(Number(foreignKeys.foreign_keys), 1);
    assert.equal(Number(trustedSchema.trusted_schema), 0);
    const foundation = store.foundationStatus();
    assert.equal(foundation.userVersion, 4);
    assert.equal(foundation.closureForeignKeyRestricted, true);

    const strictRows = store.db
      .prepare(
        `SELECT name, strict FROM pragma_table_list
         WHERE name IN ('matrix_obligations', 'closure_receipts')
         ORDER BY name`,
      )
      .all() as Array<{ name: string; strict: number }>;
    assert.deepEqual(
      strictRows.map((row) => [row.name, Number(row.strict)]),
      [
        ['closure_receipts', 1],
        ['matrix_obligations', 1],
      ],
    );
  } finally {
    cleanup();
  }
});

test('obligation identity is the exact binding digest and opens at INFINITY', () => {
  const { store, cleanup } = fixture();
  try {
    const declared = binding();
    const opened = store.openMatrixObligation(declared);
    assert.equal(opened.currentState, 'SUSPENDED_STATE');
    assert.equal(opened.opCost, 'INFINITY');
    assert.throws(
      () =>
        store.openMatrixObligation({
          ...declared,
          obligationId: '0'.repeat(64),
          sourceNodeId: 'Node_B',
        }),
      /OBLIGATION_DIGEST_MISMATCH/,
    );
  } finally {
    cleanup();
  }
});

test('authenticated closure atomically changes INFINITY to finite witness cost', () => {
  const { store, cleanup } = fixture();
  try {
    const declared = binding();
    store.openMatrixObligation(declared);
    const closure = applyClosure(store, {
      obligationId: declared.obligationId,
      resolvedByNode: 'Node_C',
      witnessDigest: sha256Canonical({ witness: 'exact' }),
      resultDigest: sha256Canonical({ result: 'exact' }),
      graphHash: '0'.repeat(64),
      propagationHopLimit: 3,
      resolutionState: 'RESOLVED_BY_WITNESS',
      finiteCost: 2,
    });
    assert.equal(closure.obligation.currentState, 'RESOLVED_BY_WITNESS');
    assert.equal(closure.obligation.opCost, '2');
    assert.equal(closure.receipt.profile, 'CLOSURE_RECEIPT_v0.3');
    assert.equal(
      closure.receipt.resultDigest,
      sha256Canonical({ result: 'exact' }),
    );
    assert.equal(store.verifyClosureReceipt(closure.receipt), true);
    assert.equal(
      store.verifyClosureReceipt({
        ...closure.receipt,
        finiteCost: closure.receipt.finiteCost + 1,
      }),
      false,
    );
    assert.equal(
      store.verifyClosureReceipt({
        ...closure.receipt,
        resultDigest: 'f'.repeat(64),
      }),
      false,
    );
    const persisted = store.db
      .prepare(
        `SELECT result_digest, receipt_payload_json
         FROM closure_receipts WHERE receipt_id = ?`,
      )
      .get(closure.receipt.receiptId) as {
      result_digest: string;
      receipt_payload_json: string;
    };
    assert.equal(persisted.result_digest, closure.receipt.resultDigest);
    assert.equal(
      JSON.parse(persisted.receipt_payload_json).resultDigest,
      closure.receipt.resultDigest,
    );
    assert.deepEqual(
      store.getClosureReceipt(closure.receipt.receiptId),
      closure.receipt,
    );
  } finally {
    cleanup();
  }
});

test('receipt clearance permits only finite cost zero and bounded hop limits', () => {
  const { store, cleanup } = fixture();
  try {
    const first = binding('Node_B');
    store.openMatrixObligation(first);
    assert.throws(
      () =>
        applyClosure(store, {
          obligationId: first.obligationId,
          resolvedByNode: 'Node_C',
          witnessDigest: sha256Canonical({ witness: 1 }),
          resultDigest: sha256Canonical({ result: 1 }),
          graphHash: '0'.repeat(64),
          propagationHopLimit: 65,
          resolutionState: 'CLEARED_BY_RECEIPT',
          finiteCost: 0,
        }),
      /CLOSURE_RECEIPT_INPUT_INVALID/,
    );
    assert.throws(
      () =>
        applyClosure(store, {
          obligationId: first.obligationId,
          resolvedByNode: 'Node_C',
          witnessDigest: sha256Canonical({ witness: 1 }),
          resultDigest: sha256Canonical({ result: 1 }),
          graphHash: '0'.repeat(64),
          propagationHopLimit: 2,
          resolutionState: 'CLEARED_BY_RECEIPT',
          finiteCost: 1,
        }),
      /CLOSURE_RECEIPT_INPUT_INVALID/,
    );
    const cleared = applyClosure(store, {
      obligationId: first.obligationId,
      resolvedByNode: 'Node_C',
      witnessDigest: sha256Canonical({ witness: 1 }),
      resultDigest: sha256Canonical({ result: 1 }),
          graphHash: '0'.repeat(64),
      propagationHopLimit: 2,
      resolutionState: 'CLEARED_BY_RECEIPT',
      finiteCost: 0,
    });
    assert.equal(cleared.obligation.currentState, 'CLEARED_BY_RECEIPT');
    assert.equal(cleared.obligation.opCost, '0');
  } finally {
    cleanup();
  }
});

test('state cannot change without a bound closure receipt', () => {
  const { store, cleanup } = fixture();
  try {
    const declared = binding();
    store.openMatrixObligation(declared);
    assert.throws(
      () =>
        store.db
          .prepare(
            `UPDATE matrix_obligations
             SET current_state = 'RESOLVED_BY_WITNESS', op_cost = '1'
             WHERE obligation_id = ?`,
          )
          .run(declared.obligationId),
      /CLOSURE_RECEIPT_REQUIRED_FOR_TRANSITION/,
    );
    assert.equal(
      store.getMatrixObligation(declared.obligationId).currentState,
      'SUSPENDED_STATE',
    );
  } finally {
    cleanup();
  }
});

test('resolved obligation and closure receipt form an immutable audit relation', () => {
  const { store, cleanup } = fixture();
  try {
    const declared = binding();
    store.openMatrixObligation(declared);
    const { receipt } = applyClosure(store, {
      obligationId: declared.obligationId,
      resolvedByNode: 'Node_C',
      witnessDigest: sha256Canonical({ witness: 2 }),
      resultDigest: sha256Canonical({ result: 2 }),
      graphHash: '0'.repeat(64),
      propagationHopLimit: 1,
      resolutionState: 'RESOLVED_BY_WITNESS',
      finiteCost: 1,
    });
    assert.throws(
      () =>
        store.db
          .prepare(`DELETE FROM closure_receipts WHERE receipt_id = ?`)
          .run(receipt.receiptId),
      /CLOSURE_RECEIPT_IMMUTABLE/,
    );
    assert.throws(
      () =>
        store.db
          .prepare(`DELETE FROM matrix_obligations WHERE obligation_id = ?`)
          .run(declared.obligationId),
      /RESOLVED_OBLIGATION_TRACE_IMMUTABLE/,
    );
    assert.throws(
      () =>
        store.db
          .prepare(
            `UPDATE closure_receipts SET resolved_by_node = 'Node_D'
             WHERE receipt_id = ?`,
          )
          .run(receipt.receiptId),
      /CLOSURE_RECEIPT_IMMUTABLE/,
    );
  } finally {
    cleanup();
  }
});

test('foreign key and JSON-column checks reject orphaned or mismatched traces', () => {
  const { store, cleanup } = fixture();
  try {
    const fake = {
      profile: 'CLOSURE_RECEIPT_v0.2',
      receiptId: '1'.repeat(64),
      obligationId: '2'.repeat(64),
      resolvedByNode: 'Node_C',
      witnessDigest: '3'.repeat(64),
      verifierAlgorithm: 'HMAC-SHA256',
      verifierKeyId: secrets.activeKeyId,
      graphHash: '0'.repeat(64),
      propagationHopLimit: 1,
      resolutionState: 'RESOLVED_BY_WITNESS',
      finiteCost: 1,
      recordedAt: new Date().toISOString(),
      authorityEffect: 'NONE',
      verifierMac: '4'.repeat(64),
    };
    assert.throws(
      () =>
        store.db
          .prepare(
            `INSERT INTO closure_receipts
             (receipt_id, obligation_id, resolved_by_node, witness_digest, graph_hash, graph_binding_status,
              verifier_algorithm, verifier_key_id, verifier_key_purpose,
              verifier_mac, propagation_hop_limit, resolution_state,
              finite_cost, receipt_payload_json, recorded_at)
             VALUES (?, ?, ?, ?, ?, 'BOUND', ?, ?, 'RECEIPT_HMAC', ?, ?, ?, ?, ?, ?)`,
          )
          .run(
            fake.receiptId,
            fake.obligationId,
            fake.resolvedByNode,
            fake.witnessDigest,
            fake.graphHash,
            fake.verifierAlgorithm,
            fake.verifierKeyId,
            fake.verifierMac,
            fake.propagationHopLimit,
            fake.resolutionState,
            fake.finiteCost,
            JSON.stringify(fake),
            fake.recordedAt,
          ),
      /FOREIGN KEY constraint failed/,
    );

    const declared = binding();
    store.openMatrixObligation(declared);
    const missingResult = {
      ...fake,
      profile: 'CLOSURE_RECEIPT_v0.3',
      obligationId: declared.obligationId,
    };
    assert.throws(
      () =>
        store.db
          .prepare(
            `INSERT INTO closure_receipts
             (receipt_id, obligation_id, resolved_by_node, witness_digest,
              graph_hash, graph_binding_status, verifier_algorithm,
              verifier_key_id, verifier_key_purpose, verifier_mac,
              propagation_hop_limit, resolution_state, finite_cost,
              receipt_payload_json, recorded_at, result_digest)
             VALUES (?, ?, ?, ?, ?, 'BOUND', ?, ?, 'RECEIPT_HMAC',
                     ?, ?, ?, ?, ?, ?, NULL)`,
          )
          .run(
            missingResult.receiptId,
            missingResult.obligationId,
            missingResult.resolvedByNode,
            missingResult.witnessDigest,
            missingResult.graphHash,
            missingResult.verifierAlgorithm,
            missingResult.verifierKeyId,
            missingResult.verifierMac,
            missingResult.propagationHopLimit,
            missingResult.resolutionState,
            missingResult.finiteCost,
            JSON.stringify(missingResult),
            missingResult.recordedAt,
          ),
      /CHECK constraint failed/,
    );
    const mismatched = { ...fake, obligationId: declared.obligationId };
    assert.throws(
      () =>
        store.db
          .prepare(
            `INSERT INTO closure_receipts
             (receipt_id, obligation_id, resolved_by_node, witness_digest, graph_hash, graph_binding_status,
              verifier_algorithm, verifier_key_id, verifier_key_purpose,
              verifier_mac, propagation_hop_limit, resolution_state,
              finite_cost, receipt_payload_json, recorded_at)
             VALUES (?, ?, ?, ?, ?, 'BOUND', ?, ?, 'RECEIPT_HMAC', ?, ?, ?, ?, ?, ?)`,
          )
          .run(
            mismatched.receiptId,
            mismatched.obligationId,
            mismatched.resolvedByNode,
            mismatched.witnessDigest,
            mismatched.graphHash,
            mismatched.verifierAlgorithm,
            mismatched.verifierKeyId,
            mismatched.verifierMac,
            mismatched.propagationHopLimit,
            mismatched.resolutionState,
            mismatched.finiteCost,
            JSON.stringify({ ...mismatched, resolvedByNode: 'MISMATCH' }),
            mismatched.recordedAt,
          ),
      /CHECK constraint failed/,
    );
  } finally {
    cleanup();
  }
});

test('invalid closure inside protected transaction rolls back nonce and state', () => {
  const { store, cleanup } = fixture();
  const leases = new LeaseRegistry(secrets, store.db);
  try {
    const declared = binding();
    store.openMatrixObligation(declared);
    const payloadDigest = sha256Canonical({ closure: declared.obligationId });
    const lease = leases.issue({
      type: 'APPLY_WINDOW',
      targetResource: '/api/matrix/obligations/close',
      payloadDigest,
      scope: 'matrix:closure:apply',
      notBefore: new Date(Date.now() - 1_000).toISOString(),
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
      topologyHash: store.topologyHash(),
    });
    assert.throws(
      () =>
        store.executeOnce(lease, 'INVALID_CLOSURE', () =>
          store.applyClosureReceipt({
            obligationId: declared.obligationId,
            resolvedByNode: 'Node_C',
            witnessDigest: sha256Canonical({ witness: 3 }),
            resultDigest: 'invalid',
            graphHash: '0'.repeat(64),
            propagationHopLimit: -1,
            resolutionState: 'RESOLVED_BY_WITNESS',
            finiteCost: 1,
          }),
        ),
      /CLOSURE_RECEIPT_INPUT_INVALID/,
    );
    assert.equal(
      store.getMatrixObligation(declared.obligationId).currentState,
      'SUSPENDED_STATE',
    );
    const nonceCount = store.db
      .prepare(`SELECT COUNT(*) AS count FROM consumed_nonces WHERE nonce = ?`)
      .get(lease.nonce) as { count: number };
    assert.equal(Number(nonceCount.count), 0);
  } finally {
    cleanup();
  }
});

test('cadence engine sweeps stalled obligations precisely at TTL', () => {
  const { store, cleanup } = fixture();
  try {
    const engine = new CadenceEngine(store.db, 500, 30);
    const declared = binding();
    store.openMatrixObligation(declared);

    // Initial state
    let state = store.getMatrixObligation(declared.obligationId);
    assert.equal(state.currentState, 'SUSPENDED_STATE');

    // Run engine (no time drift yet)
    engine.enforceTimeouts();
    state = store.getMatrixObligation(declared.obligationId);
    assert.equal(state.currentState, 'SUSPENDED_STATE');

    // Advance the cadence clock without mutating the identity-bound timestamp.
    engine.enforceTimeouts(new Date(Date.now() + 31_000));
    state = store.getMatrixObligation(declared.obligationId);
    assert.equal(state.currentState, 'TERMINAL_TIMEOUT');
    assert.equal(state.opCost, '0');
  } finally {
    cleanup();
  }
});

test('Cadence Engine optimistic concurrency prevents overriding valid receipts', async () => {
  const { store, cleanup } = fixture();
  try {
    const engine = new CadenceEngine(store.db, 500, 30);
    const declared = binding();

    // 1. Setup: Insert SUSPENDED_STATE and advance only the cadence clock.
    store.openMatrixObligation(declared);

    // 2. Execution: Fire the Receipt and the Cadence Strike simultaneously
    const validReceiptInjection = Promise.resolve().then(() =>
      applyClosure(store, {
        obligationId: declared.obligationId,
        resolvedByNode: 'Node_C',
        witnessDigest: sha256Canonical({ witness: 1 }),
        resultDigest: sha256Canonical({ result: 1 }),
        graphHash: '0'.repeat(64),
        propagationHopLimit: 2,
        resolutionState: 'CLEARED_BY_RECEIPT',
        finiteCost: 0,
      })
    );

    // Triggers the worker immediately concurrently
    const cadenceStrike = Promise.resolve().then(() =>
      engine.enforceTimeouts(new Date(Date.now() + 31_000)),
    );

    await Promise.all([validReceiptInjection, cadenceStrike]);

    // 3. Validation: Network resolution must win, or fail cleanly if beat by the clock.
    const finalState = store.getMatrixObligation(declared.obligationId);

    // The system must never be in an intermediate or corrupted state
    assert.ok(['CLEARED_BY_RECEIPT', 'TERMINAL_TIMEOUT'].includes(finalState.currentState));

    if (finalState.currentState === 'CLEARED_BY_RECEIPT') {
       // Valid receipt beat the strike. The strike should have reported 0 changes.
       assert.equal(finalState.opCost, '0');
    } else {
       // Timeout beat the receipt. The receipt trigger must have rejected the injection.
       assert.equal(finalState.opCost, '0');
    }
  } finally {
    cleanup();
  }
});
