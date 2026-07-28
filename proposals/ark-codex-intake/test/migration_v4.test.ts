import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { DatabaseSync } from 'node:sqlite';
import {
  ArkBoundaryStore,
  BoundarySecrets,
  canonicalJson,
  sha256Canonical,
} from '../src/boundary';

const secrets: BoundarySecrets = {
  authSecret: 'migration-v4-auth-secret',
  signingSecret: 'migration-v4-signing-secret',
  receiptSecret: 'migration-v4-receipt-secret',
  registrySecret: 'migration-v4-registry-secret',
  activeKeyId: 'migration-v4-key-v1',
  registryKeyId: 'migration-v4-registry-key-v1',
};

function createV3Fixture(databasePath: string) {
  const initial = new ArkBoundaryStore(databasePath, secrets);
  const graphHash = sha256Canonical({ graph: 'v3-fixture' });
  const binding = {
    sourceNodeId: 'V3_Node',
    targetClass: 'SWARM_NODE',
    mediator: 'GOSSIP_PROTOCOL',
    graphHash,
  };
  const obligation = initial.openMatrixObligation(binding);
  initial.db.close();

  const db = new DatabaseSync(databasePath);
  db.exec('PRAGMA foreign_keys = OFF;');
  db.exec('BEGIN EXCLUSIVE;');
  db.exec(`
    DROP TRIGGER IF EXISTS matrix_obligation_identity_immutable;
    DROP TRIGGER IF EXISTS matrix_obligation_transition_guard;
    DROP TRIGGER IF EXISTS closure_receipt_graph_binding_guard;
    DROP TRIGGER IF EXISTS closure_receipt_applies_transition;
    DROP TRIGGER IF EXISTS closure_receipt_immutable_update;
    DROP TRIGGER IF EXISTS closure_receipt_immutable_delete;
    DROP TRIGGER IF EXISTS resolved_obligation_trace_immutable;
    DROP TRIGGER IF EXISTS trust_key_registry_update_guard;
    DROP TRIGGER IF EXISTS registry_update_immutable_update;
    DROP TRIGGER IF EXISTS registry_update_immutable_delete;

    ALTER TABLE closure_receipts RENAME TO closure_receipts_v4;
    CREATE TABLE closure_receipts (
      receipt_id TEXT PRIMARY KEY,
      obligation_id TEXT NOT NULL UNIQUE,
      resolved_by_node TEXT NOT NULL,
      witness_digest TEXT NOT NULL,
      graph_hash TEXT,
      graph_binding_status TEXT NOT NULL,
      verifier_algorithm TEXT NOT NULL,
      verifier_key_id TEXT NOT NULL,
      verifier_key_purpose TEXT NOT NULL,
      verifier_mac TEXT NOT NULL,
      propagation_hop_limit INTEGER NOT NULL,
      resolution_state TEXT NOT NULL,
      finite_cost INTEGER NOT NULL,
      receipt_payload_json TEXT NOT NULL,
      recorded_at TEXT NOT NULL,
      FOREIGN KEY(obligation_id)
        REFERENCES matrix_obligations(obligation_id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT,
      FOREIGN KEY(verifier_key_id, verifier_key_purpose)
        REFERENCES trust_keys(key_id, purpose)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
    ) STRICT;
    DROP TABLE closure_receipts_v4;
    PRAGMA user_version = 3;
  `);

  const body = {
    profile: 'CLOSURE_RECEIPT_v0.2',
    obligationId: obligation.obligationId,
    resolvedByNode: 'V3_Resolver',
    witnessDigest: sha256Canonical({ witness: 'v3-fixture' }),
    graphHash,
    verifierAlgorithm: 'HMAC-SHA256',
    verifierKeyId: secrets.activeKeyId,
    propagationHopLimit: 1,
    resolutionState: 'RESOLVED_BY_WITNESS',
    finiteCost: 1,
    recordedAt: '2026-07-28T20:00:00.000Z',
    authorityEffect: 'NONE',
  };
  const receiptId = sha256Canonical(body);
  const receipt = {
    ...body,
    receiptId,
    verifierMac: 'a'.repeat(64),
  };
  const receiptJson = canonicalJson(receipt);
  db.prepare(
    `INSERT INTO closure_receipts
     (receipt_id, obligation_id, resolved_by_node, witness_digest,
      graph_hash, graph_binding_status, verifier_algorithm,
      verifier_key_id, verifier_key_purpose, verifier_mac,
      propagation_hop_limit, resolution_state, finite_cost,
      receipt_payload_json, recorded_at)
     VALUES (?, ?, ?, ?, ?, 'BOUND', ?, ?, 'RECEIPT_HMAC', ?, ?, ?, ?, ?, ?)`,
  ).run(
    receiptId,
    obligation.obligationId,
    receipt.resolvedByNode,
    receipt.witnessDigest,
    graphHash,
    receipt.verifierAlgorithm,
    receipt.verifierKeyId,
    receipt.verifierMac,
    receipt.propagationHopLimit,
    receipt.resolutionState,
    receipt.finiteCost,
    receiptJson,
    receipt.recordedAt,
  );
  db.prepare(
    `UPDATE matrix_obligations
     SET current_state = 'RESOLVED_BY_WITNESS', op_cost = '1'
     WHERE obligation_id = ?`,
  ).run(obligation.obligationId);
  db.exec('COMMIT;');
  db.close();
  return { obligationId: obligation.obligationId, receiptId, receiptJson };
}

test('v3 to v4 migration preserves v0.2 bytes and binds result_digest to new v0.3 receipts', () => {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'ark-v4-migration-'));
  const databasePath = path.join(directory, 'ark.sqlite');
  const fixture = createV3Fixture(databasePath);
  const store = new ArkBoundaryStore(databasePath, secrets);
  try {
    assert.equal(store.foundationStatus().userVersion, 4);
    assert.equal(store.foundationStatus().requiredTriggersPresent, true);
    const row = store.db
      .prepare(
        `SELECT receipt_payload_json, result_digest
         FROM closure_receipts WHERE receipt_id = ?`,
      )
      .get(fixture.receiptId) as {
      receipt_payload_json: string;
      result_digest: string | null;
    };
    assert.equal(row.receipt_payload_json, fixture.receiptJson);
    assert.equal(row.result_digest, null);
    assert.equal(
      store.db
        .prepare(
          `SELECT COUNT(*) AS count FROM sqlite_master
           WHERE type = 'table' AND name = 'closure_receipts_v3'`,
        )
        .get()!.count,
      0,
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
