import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import { DatabaseSync } from 'node:sqlite';

export const LEASE_TYPES = [
  'INGEST_WINDOW',
  'EVALUATE_WINDOW',
  'APPLY_WINDOW',
  'FEDERATION_WINDOW',
] as const;

export type LeaseType = (typeof LEASE_TYPES)[number];

export interface LeasePayload {
  id: string;
  nonce: string;
  type: LeaseType;
  targetResource: string;
  payloadDigest: string;
  scope: string;
  notBefore: string;
  expiresAt: string;
  keyId: string;
  issuer: string;
  parentNodeId?: string;
  topologyHash?: string;
}

export interface AuthenticatedLease extends LeasePayload {
  algorithm: 'HMAC-SHA256';
  mac: string;
}

export interface BoundarySecrets {
  authSecret: string;
  signingSecret: string;
  receiptSecret: string;
  registrySecret: string;
  activeKeyId: string;
  registryKeyId: string;
}

export interface LeaseExpectation {
  type: LeaseType;
  targetResource: string;
  payloadDigest: string;
  scope: string;
  topologyHash?: string;
}

export interface TransactionResult<T> {
  response: T;
  replayed: boolean;
  receipt: AuthenticatedReceipt;
}

export interface AuthenticatedReceipt {
  receiptId: string;
  algorithm: 'HMAC-SHA256';
  keyId: string;
  operation: string;
  targetResource: string;
  payloadDigest: string;
  resultDigest: string;
  nonce: string;
  createdAt: string;
  mac: string;
}

export type MatrixObligationState =
  | 'SUSPENDED_STATE'
  | 'CLEARED_BY_RECEIPT'
  | 'RESOLVED_BY_WITNESS'
  | 'TERMINAL_TIMEOUT';

export interface MatrixObligation {
  obligationId: string;
  sourceNodeId: string;
  targetClass: string;
  mediator: string;
  graphHash: string | null;
  graphBindingStatus: 'LEGACY_UNBOUND' | 'BOUND';
  currentState: MatrixObligationState;
  opCost: 'INFINITY' | string;
  initiatedAt: string;
}

export interface ClosureReceipt_v02 {
  profile: 'CLOSURE_RECEIPT_v0.2';
  receiptId: string;
  obligationId: string;
  resolvedByNode: string;
  witnessDigest: string;
  graphHash: string;
  verifierAlgorithm: 'HMAC-SHA256';
  verifierKeyId: string;
  propagationHopLimit: number;
  resolutionState: Exclude<MatrixObligationState, 'SUSPENDED_STATE'>;
  finiteCost: number;
  recordedAt: string;
  authorityEffect: 'NONE';
  verifierMac: string;
}

export interface ClosureReceipt_v03
  extends Omit<ClosureReceipt_v02, 'profile'> {
  profile: 'CLOSURE_RECEIPT_v0.3';
  resultDigest: string;
}

export type ClosureReceipt = ClosureReceipt_v02 | ClosureReceipt_v03;

export interface RegistryUpdate_v01 {
  profile: 'REGISTRY_UPDATE_v0.1';
  updateId: string;
  keyId: string;
  purpose: string;
  keyVersion: number;
  revoked: true;
  revokedAt: string;
  reason: string;
  issuerKeyId: string;
  algorithm: 'HMAC-SHA256';
  macDomain: 'ARK_REGISTRY_UPDATE_V1';
  mac: string;
}

const SHA256_HEX = /^[a-f0-9]{64}$/;
const NONCE = /^[a-f0-9]{32,128}$/;
const RECEIPT_DOMAIN = 'ARK_RECEIPT_V1';
const LEASE_DOMAIN = 'ARK_LEASE_V1';
const CLOSURE_RECEIPT_DOMAIN_V2 = 'ARK_CLOSURE_RECEIPT_V2';
const CLOSURE_RECEIPT_DOMAIN_V3 = 'ARK_CLOSURE_RECEIPT_V3';

function requireEnv(name: string): string {
  const value = process.env[name]?.trim();
  if (!value) {
    throw new Error(`CONFIGURATION_ERROR: ${name} is required.`);
  }
  return value;
}

export function loadBoundarySecrets(): BoundarySecrets {
  return {
    authSecret: requireEnv('AUTH_SECRET'),
    signingSecret: requireEnv('SIGNING_SECRET'),
    receiptSecret: requireEnv('RECEIPT_SECRET'),
    registrySecret: requireEnv('REGISTRY_SECRET'),
    activeKeyId: requireEnv('ACTIVE_KEY_ID'),
    registryKeyId: requireEnv('REGISTRY_KEY_ID'),
  };
}

function canonicalValue(value: unknown): string {
  if (value === null) return 'null';
  if (typeof value === 'string') return JSON.stringify(value);
  if (typeof value === 'boolean') return value ? 'true' : 'false';
  if (typeof value === 'number') {
    if (!Number.isFinite(value)) {
      throw new Error('CANONICALIZATION_ERROR: non-finite number.');
    }
    return JSON.stringify(Object.is(value, -0) ? 0 : value);
  }
  if (Array.isArray(value)) {
    return `[${value.map((item) => canonicalValue(item)).join(',')}]`;
  }
  if (typeof value === 'object') {
    const record = value as Record<string, unknown>;
    const keys = Object.keys(record).sort();
    return `{${keys
      .filter((key) => record[key] !== undefined)
      .map((key) => `${JSON.stringify(key)}:${canonicalValue(record[key])}`)
      .join(',')}}`;
  }
  throw new Error(`CANONICALIZATION_ERROR: unsupported ${typeof value}.`);
}

export function canonicalJson(value: unknown): string {
  return canonicalValue(value);
}

export function sha256Bytes(value: string | Buffer): string {
  return crypto.createHash('sha256').update(value).digest('hex');
}

export function sha256Canonical(value: unknown): string {
  return sha256Bytes(canonicalJson(value));
}

function hmacHex(secret: string, domain: string, value: unknown): string {
  return crypto
    .createHmac('sha256', secret)
    .update(domain)
    .update('\0')
    .update(canonicalJson(value))
    .digest('hex');
}

function constantTimeTextEqual(left: string, right: string): boolean {
  const leftDigest = crypto.createHash('sha256').update(left).digest();
  const rightDigest = crypto.createHash('sha256').update(right).digest();
  return crypto.timingSafeEqual(leftDigest, rightDigest);
}

function constantTimeHexEqual(left: string, right: string): boolean {
  if (!SHA256_HEX.test(left) || !SHA256_HEX.test(right)) return false;
  return crypto.timingSafeEqual(Buffer.from(left, 'hex'), Buffer.from(right, 'hex'));
}

export function authenticateOperator(
  suppliedSecret: string | string[] | undefined,
  expectedSecret: string,
): void {
  if (
    typeof suppliedSecret !== 'string' ||
    !constantTimeTextEqual(suppliedSecret, expectedSecret)
  ) {
    throw new BoundaryError(401, 'AUTHENTICATION_REQUIRED');
  }
}

export class BoundaryError extends Error {
  constructor(
    public readonly statusCode: number,
    message: string,
  ) {
    super(message);
  }
}

function leaseMacBody(lease: LeasePayload): LeasePayload {
  return lease;
}

function validateIsoWindow(notBefore: string, expiresAt: string, nowMs: number): void {
  const start = Date.parse(notBefore);
  const end = Date.parse(expiresAt);
  if (!Number.isFinite(start) || !Number.isFinite(end) || start >= end) {
    throw new BoundaryError(400, 'INVALID_TEMPORAL_WINDOW');
  }
  if (nowMs < start || nowMs >= end) {
    throw new BoundaryError(403, 'LEASE_OUTSIDE_TEMPORAL_WINDOW');
  }
}

export class LeaseRegistry {
  constructor(private readonly secrets: BoundarySecrets, private readonly db: DatabaseSync) {}

  issue(input: {
    type: LeaseType;
    targetResource: string;
    payloadDigest: string;
    scope: string;
    notBefore: string;
    expiresAt: string;
    parentNodeId?: string;
    topologyHash?: string;
  }): AuthenticatedLease {
    if (!LEASE_TYPES.includes(input.type)) {
      throw new BoundaryError(400, 'UNKNOWN_LEASE_TYPE');
    }
    if (!input.targetResource.startsWith('/api/')) {
      throw new BoundaryError(400, 'INVALID_TARGET_RESOURCE');
    }
    if (!SHA256_HEX.test(input.payloadDigest)) {
      throw new BoundaryError(400, 'INVALID_PAYLOAD_DIGEST');
    }
    validateIsoWindow(input.notBefore, input.expiresAt, Date.now());

    const keyRow = this.db.prepare(
      `SELECT revoked FROM trust_keys WHERE key_id = ? AND purpose = 'LEASE_HMAC'`
    ).get(this.secrets.activeKeyId) as { revoked: number } | undefined;
    if (!keyRow || keyRow.revoked) {
      throw new BoundaryError(403, 'ACTIVE_KEY_REVOKED');
    }

    const payload: LeasePayload = {
      id: `lease_${crypto.randomBytes(16).toString('hex')}`,
      nonce: crypto.randomBytes(16).toString('hex'),
      type: input.type,
      targetResource: input.targetResource,
      payloadDigest: input.payloadDigest,
      scope: input.scope,
      notBefore: new Date(input.notBefore).toISOString(),
      expiresAt: new Date(input.expiresAt).toISOString(),
      keyId: this.secrets.activeKeyId,
      issuer: 'SOVEREIGN_OPERATOR',
      ...(input.parentNodeId ? { parentNodeId: input.parentNodeId } : {}),
      ...(input.topologyHash ? { topologyHash: input.topologyHash } : {}),
    };

    return {
      ...payload,
      algorithm: 'HMAC-SHA256',
      mac: hmacHex(this.secrets.signingSecret, LEASE_DOMAIN, leaseMacBody(payload)),
    };
  }

  verify(
    candidate: unknown,
    expected: LeaseExpectation,
    nowMs = Date.now(),
  ): AuthenticatedLease {
    if (!candidate || typeof candidate !== 'object') {
      throw new BoundaryError(403, 'LEASE_REQUIRED');
    }
    const lease = candidate as AuthenticatedLease;
    if (lease.algorithm !== 'HMAC-SHA256') {
      throw new BoundaryError(403, 'LEASE_KEY_REJECTED');
    }

    const keyRow = this.db.prepare(
      `SELECT revoked FROM trust_keys WHERE key_id = ? AND purpose = 'LEASE_HMAC'`
    ).get(lease.keyId) as { revoked: number } | undefined;

    if (!keyRow || keyRow.revoked) {
      throw new BoundaryError(403, 'LEASE_KEY_REJECTED');
    }

    if (!NONCE.test(lease.nonce) || !SHA256_HEX.test(lease.payloadDigest)) {
      throw new BoundaryError(403, 'MALFORMED_LEASE');
    }
    if (
      lease.type !== expected.type ||
      lease.targetResource !== expected.targetResource ||
      lease.payloadDigest !== expected.payloadDigest ||
      lease.scope !== expected.scope ||
      (expected.topologyHash !== undefined &&
        lease.topologyHash !== expected.topologyHash)
    ) {
      throw new BoundaryError(403, 'LEASE_BINDING_MISMATCH');
    }
    validateIsoWindow(lease.notBefore, lease.expiresAt, nowMs);

    const {
      algorithm: _algorithm,
      mac,
      ...payload
    } = lease;
    const expectedMac = hmacHex(
      this.secrets.signingSecret,
      LEASE_DOMAIN,
      leaseMacBody(payload),
    );
    if (!constantTimeHexEqual(mac, expectedMac)) {
      throw new BoundaryError(403, 'LEASE_MAC_INVALID');
    }
    return lease;
  }
}

function ensureParentDirectory(filePath: string): void {
  fs.mkdirSync(path.dirname(filePath), { recursive: true });
}

export class ArkBoundaryStore {
  readonly db: DatabaseSync;

  constructor(
    databasePath: string,
    private readonly secrets: BoundarySecrets,
  ) {
    ensureParentDirectory(databasePath);
    this.db = new DatabaseSync(databasePath);
    // Install the busy handler before requesting WAL mode. Multiple workers may
    // open the same database simultaneously, and journal-mode negotiation can
    // itself require the write lock.
    this.db.exec('PRAGMA busy_timeout = 10000;');
    this.db.exec(`
      PRAGMA journal_mode = WAL;
      PRAGMA synchronous = NORMAL;
      PRAGMA foreign_keys = ON;
      PRAGMA trusted_schema = OFF;
    `);

    const userVersionRow = this.db.prepare('PRAGMA user_version').get() as { user_version: number };
    const userVersion = userVersionRow.user_version;

    if (userVersion === 2) {
      this.db.exec('PRAGMA foreign_keys = OFF;');
      try {
        this.db.exec('BEGIN EXCLUSIVE;');
        this.db.exec(`
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

          ALTER TABLE trust_keys RENAME TO trust_keys_old;
          ALTER TABLE matrix_obligations RENAME TO matrix_obligations_old;
          ALTER TABLE closure_receipts RENAME TO closure_receipts_old;
        `);
        this.createSchemaV4();
        this.db.exec(`
          INSERT INTO trust_keys
            (key_id, purpose, registry_version, revoked, last_update_id, created_at)
          SELECT key_id, purpose, 0, revoked, NULL, created_at
          FROM trust_keys_old;

          INSERT INTO matrix_obligations
            (obligation_id, source_node_id, target_class, mediator,
             graph_hash, graph_binding_status, current_state, op_cost, initiated_at)
          SELECT obligation_id, source_node_id, target_class, mediator,
                 NULL, 'LEGACY_UNBOUND', current_state, op_cost, initiated_at
          FROM matrix_obligations_old;

          INSERT INTO closure_receipts
            (receipt_id, obligation_id, resolved_by_node, witness_digest,
             graph_hash, graph_binding_status, verifier_algorithm,
             verifier_key_id, verifier_key_purpose, verifier_mac,
             propagation_hop_limit, resolution_state, finite_cost,
             receipt_payload_json, recorded_at, result_digest)
          SELECT receipt_id, obligation_id, resolved_by_node, witness_digest,
                 NULL, 'LEGACY_UNBOUND', verifier_algorithm,
                 verifier_key_id, verifier_key_purpose, verifier_mac,
                 propagation_hop_limit, resolution_state, finite_cost,
                 receipt_payload_json, recorded_at, NULL
          FROM closure_receipts_old;

          DROP TABLE closure_receipts_old;
          DROP TABLE matrix_obligations_old;
          DROP TABLE trust_keys_old;
          PRAGMA user_version = 4;
        `);
        this.db.exec('COMMIT;');
      } catch (error) {
        try {
          this.db.exec('ROLLBACK;');
        } catch {
          // Preserve the original migration failure.
        }
        throw error;
      } finally {
        this.db.exec('PRAGMA foreign_keys = ON;');
      }
    } else if (userVersion === 3) {
      this.db.exec('PRAGMA foreign_keys = OFF;');
      try {
        this.db.exec('BEGIN EXCLUSIVE;');
        this.db.exec(`
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

          ALTER TABLE closure_receipts RENAME TO closure_receipts_v3;
        `);
        this.createSchemaV4();
        this.db.exec(`
          INSERT INTO closure_receipts
            (receipt_id, obligation_id, resolved_by_node, witness_digest,
             graph_hash, graph_binding_status, verifier_algorithm,
             verifier_key_id, verifier_key_purpose, verifier_mac,
             propagation_hop_limit, resolution_state, finite_cost,
             receipt_payload_json, recorded_at, result_digest)
          SELECT receipt_id, obligation_id, resolved_by_node, witness_digest,
                 graph_hash, graph_binding_status, verifier_algorithm,
                 verifier_key_id, verifier_key_purpose, verifier_mac,
                 propagation_hop_limit, resolution_state, finite_cost,
                 receipt_payload_json, recorded_at, NULL
          FROM closure_receipts_v3;

          DROP TABLE closure_receipts_v3;
          PRAGMA user_version = 4;
        `);
        this.db.exec('COMMIT;');
      } catch (error) {
        try {
          this.db.exec('ROLLBACK;');
        } catch {
          // Preserve the original migration failure.
        }
        throw error;
      } finally {
        this.db.exec('PRAGMA foreign_keys = ON;');
      }
    } else if (userVersion === 0) {
      this.db.exec('PRAGMA foreign_keys = ON;');
      this.db.exec('BEGIN EXCLUSIVE;');
      this.createSchemaV4();
      this.seedTrustRegistry();
      this.db.exec('PRAGMA user_version = 4;');
      this.db.exec('COMMIT;');
    } else if (userVersion !== 4) {
      throw new Error(`UNSUPPORTED_USER_VERSION: ${userVersion}`);
    } else {
      this.db.exec('PRAGMA foreign_keys = ON;');
      this.createSchemaV4();
    }

    this.db.exec('PRAGMA trusted_schema = OFF;');
    this.seedTrustRegistry();
    this.assertFoundation();
  }

  private createSchemaV4(): void {
    this.db.exec(`
      CREATE TABLE IF NOT EXISTS trust_keys (
        key_id TEXT NOT NULL,
        purpose TEXT NOT NULL,
        registry_version INTEGER NOT NULL DEFAULT 0,
        revoked INTEGER NOT NULL DEFAULT 0 CHECK (revoked IN (0, 1)),
        last_update_id TEXT,
        created_at TEXT NOT NULL,
        PRIMARY KEY (key_id, purpose)
      );
      CREATE TABLE IF NOT EXISTS registry_updates (
        update_id TEXT PRIMARY KEY
          CHECK (
            length(update_id) = 64
            AND update_id = lower(update_id)
            AND update_id NOT GLOB '*[^0-9a-f]*'
          ),
        key_id TEXT NOT NULL CHECK (length(key_id) BETWEEN 1 AND 128),
        purpose TEXT NOT NULL
          CHECK (purpose IN ('LEASE_HMAC', 'RECEIPT_HMAC', 'REGISTRY_HMAC')),
        key_version INTEGER NOT NULL CHECK (key_version >= 1),
        revoked INTEGER NOT NULL CHECK (revoked = 1),
        revoked_at TEXT NOT NULL,
        reason TEXT NOT NULL CHECK (length(reason) BETWEEN 1 AND 512),
        issuer_key_id TEXT NOT NULL CHECK (length(issuer_key_id) BETWEEN 1 AND 128),
        algorithm TEXT NOT NULL CHECK (algorithm = 'HMAC-SHA256'),
        mac_domain TEXT NOT NULL CHECK (mac_domain = 'ARK_REGISTRY_UPDATE_V1'),
        mac TEXT NOT NULL
          CHECK (
            length(mac) = 64
            AND mac = lower(mac)
            AND mac NOT GLOB '*[^0-9a-f]*'
          ),
        update_json TEXT NOT NULL CHECK (json_valid(update_json)),
        applied_at TEXT NOT NULL,
        UNIQUE (key_id, purpose, key_version),
        CHECK (
          json_extract(update_json, '$.profile') = 'REGISTRY_UPDATE_v0.1'
          AND json_extract(update_json, '$.updateId') = update_id
          AND json_extract(update_json, '$.keyId') = key_id
          AND json_extract(update_json, '$.purpose') = purpose
          AND json_extract(update_json, '$.keyVersion') = key_version
          AND json_extract(update_json, '$.revoked') = 1
          AND json_extract(update_json, '$.revokedAt') = revoked_at
          AND json_extract(update_json, '$.reason') = reason
          AND json_extract(update_json, '$.issuerKeyId') = issuer_key_id
          AND json_extract(update_json, '$.algorithm') = algorithm
          AND json_extract(update_json, '$.macDomain') = mac_domain
          AND json_extract(update_json, '$.mac') = mac
        )
      ) STRICT;
      CREATE TABLE IF NOT EXISTS shadow_ledger (
        entry_hash TEXT PRIMARY KEY,
        status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'QUARANTINE')),
        source TEXT NOT NULL,
        entry_json TEXT NOT NULL,
        created_at TEXT NOT NULL
      );
      CREATE TABLE IF NOT EXISTS consumed_nonces (
        nonce TEXT PRIMARY KEY,
        lease_id TEXT NOT NULL UNIQUE,
        target_resource TEXT NOT NULL,
        payload_digest TEXT NOT NULL,
        status TEXT NOT NULL CHECK (status IN ('CLAIMED', 'COMMITTED')),
        response_json TEXT,
        receipt_id TEXT,
        committed_at TEXT
      );
      CREATE TABLE IF NOT EXISTS audit_receipts (
        receipt_id TEXT PRIMARY KEY,
        operation TEXT NOT NULL,
        target_resource TEXT NOT NULL,
        payload_digest TEXT NOT NULL,
        result_digest TEXT NOT NULL,
        nonce TEXT NOT NULL UNIQUE,
        key_id TEXT NOT NULL,
        algorithm TEXT NOT NULL,
        receipt_json TEXT NOT NULL,
        mac TEXT NOT NULL,
        created_at TEXT NOT NULL
      );
      CREATE TABLE IF NOT EXISTS external_operations (
        nonce TEXT PRIMARY KEY,
        lease_id TEXT NOT NULL UNIQUE,
        target_resource TEXT NOT NULL,
        payload_digest TEXT NOT NULL,
        status TEXT NOT NULL CHECK (status IN ('PENDING', 'COMMITTED', 'FAILED')),
        response_json TEXT,
        error_code TEXT,
        updated_at TEXT NOT NULL
      );
      CREATE TABLE IF NOT EXISTS matrix_obligations (
        obligation_id TEXT PRIMARY KEY
          CHECK (
            length(obligation_id) = 64
            AND obligation_id = lower(obligation_id)
            AND obligation_id NOT GLOB '*[^0-9a-f]*'
          ),
        source_node_id TEXT NOT NULL CHECK (length(source_node_id) BETWEEN 1 AND 256),
        target_class TEXT NOT NULL CHECK (length(target_class) BETWEEN 1 AND 128),
        mediator TEXT NOT NULL CHECK (length(mediator) BETWEEN 1 AND 128),
        graph_hash TEXT CHECK (
          (graph_binding_status = 'LEGACY_UNBOUND' AND graph_hash IS NULL) OR
          (graph_binding_status = 'BOUND' AND length(graph_hash) = 64 AND graph_hash = lower(graph_hash) AND graph_hash NOT GLOB '*[^0-9a-f]*')
        ),
        graph_binding_status TEXT NOT NULL DEFAULT 'BOUND'
          CHECK (graph_binding_status IN ('LEGACY_UNBOUND', 'BOUND')),
        current_state TEXT NOT NULL DEFAULT 'SUSPENDED_STATE'
          CHECK (
            current_state IN (
              'SUSPENDED_STATE',
              'CLEARED_BY_RECEIPT',
              'RESOLVED_BY_WITNESS',
              'TERMINAL_TIMEOUT'
            )
          ),
        op_cost TEXT NOT NULL DEFAULT 'INFINITY',
        initiated_at TEXT NOT NULL,
        CHECK (
          (current_state = 'SUSPENDED_STATE' AND op_cost = 'INFINITY')
          OR (current_state = 'CLEARED_BY_RECEIPT' AND op_cost = '0')
          OR (current_state = 'TERMINAL_TIMEOUT' AND op_cost = '0')
          OR (
            current_state = 'RESOLVED_BY_WITNESS'
            AND length(op_cost) > 0
            AND op_cost NOT GLOB '*[^0-9]*'
            AND (op_cost = '0' OR substr(op_cost, 1, 1) != '0')
          )
        )
      ) STRICT;
      CREATE TABLE IF NOT EXISTS closure_receipts (
        receipt_id TEXT PRIMARY KEY
          CHECK (
            length(receipt_id) = 64
            AND receipt_id = lower(receipt_id)
            AND receipt_id NOT GLOB '*[^0-9a-f]*'
          ),
        obligation_id TEXT NOT NULL UNIQUE,
        resolved_by_node TEXT NOT NULL CHECK (length(resolved_by_node) BETWEEN 1 AND 256),
        witness_digest TEXT NOT NULL
          CHECK (
            length(witness_digest) = 64
            AND witness_digest = lower(witness_digest)
            AND witness_digest NOT GLOB '*[^0-9a-f]*'
          ),
        graph_hash TEXT CHECK (
          (graph_binding_status = 'LEGACY_UNBOUND' AND graph_hash IS NULL) OR
          (graph_binding_status = 'BOUND' AND length(graph_hash) = 64 AND graph_hash = lower(graph_hash) AND graph_hash NOT GLOB '*[^0-9a-f]*')
        ),
        graph_binding_status TEXT NOT NULL DEFAULT 'BOUND'
          CHECK (graph_binding_status IN ('LEGACY_UNBOUND', 'BOUND')),
        verifier_algorithm TEXT NOT NULL CHECK (verifier_algorithm = 'HMAC-SHA256'),
        verifier_key_id TEXT NOT NULL CHECK (length(verifier_key_id) BETWEEN 1 AND 128),
        verifier_key_purpose TEXT NOT NULL DEFAULT 'RECEIPT_HMAC'
          CHECK (verifier_key_purpose = 'RECEIPT_HMAC'),
        verifier_mac TEXT NOT NULL
          CHECK (
            length(verifier_mac) = 64
            AND verifier_mac = lower(verifier_mac)
            AND verifier_mac NOT GLOB '*[^0-9a-f]*'
          ),
        propagation_hop_limit INTEGER NOT NULL
          CHECK (propagation_hop_limit BETWEEN 0 AND 64),
        resolution_state TEXT NOT NULL
          CHECK (resolution_state IN ('CLEARED_BY_RECEIPT', 'RESOLVED_BY_WITNESS')),
        finite_cost INTEGER NOT NULL CHECK (finite_cost >= 0),
        receipt_payload_json TEXT NOT NULL CHECK (json_valid(receipt_payload_json)),
        recorded_at TEXT NOT NULL,
        result_digest TEXT CHECK (
          (
            json_extract(receipt_payload_json, '$.profile')
              IN ('CLOSURE_RECEIPT_v0.1', 'CLOSURE_RECEIPT_v0.2')
            AND result_digest IS NULL
          )
          OR
          (
            graph_binding_status = 'BOUND'
            AND json_extract(receipt_payload_json, '$.profile')
              = 'CLOSURE_RECEIPT_v0.3'
            AND result_digest IS NOT NULL
            AND length(result_digest) = 64
            AND result_digest = lower(result_digest)
            AND result_digest NOT GLOB '*[^0-9a-f]*'
            AND json_type(receipt_payload_json, '$.resultDigest') = 'text'
            AND json_extract(receipt_payload_json, '$.resultDigest')
              = result_digest
          )
        ),
        FOREIGN KEY(obligation_id)
          REFERENCES matrix_obligations(obligation_id)
          ON DELETE RESTRICT
          ON UPDATE RESTRICT,
        FOREIGN KEY(verifier_key_id, verifier_key_purpose)
          REFERENCES trust_keys(key_id, purpose)
          ON DELETE RESTRICT
          ON UPDATE RESTRICT,
        CHECK (
          resolution_state != 'CLEARED_BY_RECEIPT' OR finite_cost = 0
        ),
        CHECK (
          (graph_binding_status = 'LEGACY_UNBOUND' AND json_extract(receipt_payload_json, '$.profile') = 'CLOSURE_RECEIPT_v0.1' AND json_extract(receipt_payload_json, '$.graphHash') IS NULL)
          OR
          (graph_binding_status = 'BOUND' AND json_extract(receipt_payload_json, '$.profile') = 'CLOSURE_RECEIPT_v0.2' AND json_extract(receipt_payload_json, '$.graphHash') = graph_hash)
          OR
          (graph_binding_status = 'BOUND' AND json_extract(receipt_payload_json, '$.profile') = 'CLOSURE_RECEIPT_v0.3' AND json_extract(receipt_payload_json, '$.graphHash') = graph_hash AND json_extract(receipt_payload_json, '$.resultDigest') = result_digest)
        ),
        CHECK (
          json_extract(receipt_payload_json, '$.receiptId') = receipt_id
          AND json_extract(receipt_payload_json, '$.obligationId') = obligation_id
          AND json_extract(receipt_payload_json, '$.resolvedByNode') = resolved_by_node
          AND json_extract(receipt_payload_json, '$.witnessDigest') = witness_digest
          AND json_extract(receipt_payload_json, '$.verifierAlgorithm') = verifier_algorithm
          AND json_extract(receipt_payload_json, '$.verifierKeyId') = verifier_key_id
          AND json_extract(receipt_payload_json, '$.verifierMac') = verifier_mac
          AND json_extract(receipt_payload_json, '$.propagationHopLimit') = propagation_hop_limit
          AND json_extract(receipt_payload_json, '$.resolutionState') = resolution_state
          AND json_extract(receipt_payload_json, '$.finiteCost') = finite_cost
          AND json_extract(receipt_payload_json, '$.recordedAt') = recorded_at
          AND json_extract(receipt_payload_json, '$.authorityEffect') = 'NONE'
        )
      ) STRICT;
      CREATE TRIGGER IF NOT EXISTS matrix_obligation_identity_immutable
      BEFORE UPDATE OF
        obligation_id, source_node_id, target_class, mediator, graph_hash, graph_binding_status, initiated_at
      ON matrix_obligations
      BEGIN
        SELECT RAISE(ABORT, 'MATRIX_OBLIGATION_IDENTITY_IMMUTABLE');
      END;
      CREATE TRIGGER IF NOT EXISTS matrix_obligation_transition_guard
      BEFORE UPDATE OF current_state, op_cost
      ON matrix_obligations
      WHEN NOT (
        OLD.current_state = NEW.current_state
        AND OLD.op_cost = NEW.op_cost
      )
      BEGIN
        SELECT CASE
          WHEN OLD.graph_binding_status = 'LEGACY_UNBOUND'
            THEN RAISE(ABORT, 'LEGACY_OBLIGATION_READ_ONLY')
          WHEN OLD.current_state != 'SUSPENDED_STATE'
            THEN RAISE(ABORT, 'MATRIX_OBLIGATION_TERMINAL')
          WHEN NEW.current_state NOT IN ('CLEARED_BY_RECEIPT', 'RESOLVED_BY_WITNESS', 'TERMINAL_TIMEOUT')
            THEN RAISE(ABORT, 'MATRIX_OBLIGATION_TRANSITION_INVALID')
          WHEN NEW.current_state != 'TERMINAL_TIMEOUT' AND NOT EXISTS (
            SELECT 1 FROM closure_receipts AS receipt
            WHERE receipt.obligation_id = OLD.obligation_id
              AND receipt.resolution_state = NEW.current_state
              AND CAST(receipt.finite_cost AS TEXT) = NEW.op_cost
          )
            THEN RAISE(ABORT, 'CLOSURE_RECEIPT_REQUIRED_FOR_TRANSITION')
        END;
      END;
      CREATE TRIGGER IF NOT EXISTS closure_receipt_graph_binding_guard
      BEFORE INSERT ON closure_receipts
      WHEN NEW.graph_binding_status = 'BOUND'
        AND EXISTS (
          SELECT 1 FROM matrix_obligations
          WHERE obligation_id = NEW.obligation_id
        )
      BEGIN
        SELECT CASE
          WHEN NOT EXISTS (
            SELECT 1
            FROM matrix_obligations AS obligation
            WHERE obligation.obligation_id = NEW.obligation_id
              AND obligation.graph_binding_status = 'BOUND'
              AND obligation.graph_hash = NEW.graph_hash
          )
            THEN RAISE(ABORT, 'CLOSURE_RECEIPT_GRAPH_MISMATCH')
        END;
      END;
      CREATE TRIGGER IF NOT EXISTS closure_receipt_applies_transition
      AFTER INSERT ON closure_receipts
      WHEN NEW.graph_binding_status = 'BOUND'
      BEGIN
        UPDATE matrix_obligations
        SET
          current_state = NEW.resolution_state,
          op_cost = CAST(NEW.finite_cost AS TEXT)
        WHERE obligation_id = NEW.obligation_id;
      END;
      CREATE TRIGGER IF NOT EXISTS closure_receipt_immutable_update
      BEFORE UPDATE ON closure_receipts
      BEGIN
        SELECT RAISE(ABORT, 'CLOSURE_RECEIPT_IMMUTABLE');
      END;
      CREATE TRIGGER IF NOT EXISTS closure_receipt_immutable_delete
      BEFORE DELETE ON closure_receipts
      BEGIN
        SELECT RAISE(ABORT, 'CLOSURE_RECEIPT_IMMUTABLE');
      END;
      CREATE TRIGGER IF NOT EXISTS resolved_obligation_trace_immutable
      BEFORE DELETE ON matrix_obligations
      WHEN OLD.graph_binding_status = 'LEGACY_UNBOUND'
        OR OLD.current_state != 'SUSPENDED_STATE'
        OR EXISTS (
          SELECT 1 FROM closure_receipts
          WHERE obligation_id = OLD.obligation_id
        )
      BEGIN
        SELECT RAISE(ABORT, 'RESOLVED_OBLIGATION_TRACE_IMMUTABLE');
      END;
      CREATE TRIGGER IF NOT EXISTS trust_key_registry_update_guard
      BEFORE UPDATE OF registry_version, revoked, last_update_id
      ON trust_keys
      WHEN OLD.registry_version != NEW.registry_version
        OR OLD.revoked != NEW.revoked
        OR OLD.last_update_id IS NOT NEW.last_update_id
      BEGIN
        SELECT CASE
          WHEN NOT EXISTS (
            SELECT 1 FROM registry_updates AS update_record
            WHERE update_record.update_id = NEW.last_update_id
              AND update_record.key_id = NEW.key_id
              AND update_record.purpose = NEW.purpose
              AND update_record.key_version = NEW.registry_version
              AND update_record.revoked = NEW.revoked
              AND NEW.registry_version > OLD.registry_version
          )
            THEN RAISE(ABORT, 'GOVERNED_REGISTRY_UPDATE_REQUIRED')
        END;
      END;
      CREATE TRIGGER IF NOT EXISTS registry_update_immutable_update
      BEFORE UPDATE ON registry_updates
      BEGIN
        SELECT RAISE(ABORT, 'REGISTRY_UPDATE_IMMUTABLE');
      END;
      CREATE TRIGGER IF NOT EXISTS registry_update_immutable_delete
      BEFORE DELETE ON registry_updates
      BEGIN
        SELECT RAISE(ABORT, 'REGISTRY_UPDATE_IMMUTABLE');
      END;
    `);
  }

  foundationStatus(): {
    journalMode: string;
    synchronous: number;
    foreignKeys: number;
    trustedSchema: number;
    userVersion: number;
    strictTables: string[];
    closureForeignKeyRestricted: boolean;
    requiredTriggersPresent: boolean;
  } {
    const journal = this.db.prepare('PRAGMA journal_mode').get() as { journal_mode: string };
    const synchronous = this.db.prepare('PRAGMA synchronous').get() as { synchronous: number };
    const foreignKeys = this.db.prepare('PRAGMA foreign_keys').get() as { foreign_keys: number };
    const trustedSchema = this.db.prepare('PRAGMA trusted_schema').get() as { trusted_schema: number };
    const userVersion = this.db.prepare('PRAGMA user_version').get() as { user_version: number };
    const strictTables = (
      this.db
        .prepare(
          `SELECT name FROM pragma_table_list
           WHERE strict = 1
             AND name IN ('matrix_obligations', 'closure_receipts')
           ORDER BY name`
        )
        .all() as Array<{ name: string }>
    ).map((row) => row.name);
    const foreignKeysRows = this.db
      .prepare(`PRAGMA foreign_key_list('closure_receipts')`)
      .all() as Array<{ table: string; from: string; on_delete: string; on_update: string }>;
    const triggerRows = this.db
      .prepare(
        `SELECT name FROM sqlite_master
         WHERE type = 'trigger'
           AND name IN (
             'matrix_obligation_identity_immutable',
             'matrix_obligation_transition_guard',
             'closure_receipt_graph_binding_guard',
             'closure_receipt_applies_transition',
             'closure_receipt_immutable_update',
             'closure_receipt_immutable_delete',
             'resolved_obligation_trace_immutable',
             'trust_key_registry_update_guard',
             'registry_update_immutable_update',
             'registry_update_immutable_delete'
           )`,
      )
      .all() as Array<{ name: string }>;
    return {
      journalMode: journal.journal_mode.toLowerCase(),
      synchronous: Number(synchronous.synchronous),
      foreignKeys: Number(foreignKeys.foreign_keys),
      trustedSchema: Number(trustedSchema.trusted_schema),
      userVersion: Number(userVersion.user_version),
      strictTables,
      closureForeignKeyRestricted: foreignKeysRows.some(
        (row) =>
          row.table === 'matrix_obligations' &&
          row.from === 'obligation_id' &&
          row.on_delete === 'RESTRICT' &&
          row.on_update === 'RESTRICT',
      ),
      requiredTriggersPresent: triggerRows.length === 10,
    };
  }

  private assertFoundation(): void {
    const status = this.foundationStatus();
    if (
      status.journalMode !== 'wal' ||
      status.synchronous !== 1 ||
      status.foreignKeys !== 1 ||
      status.trustedSchema !== 0 ||
      status.userVersion !== 4 ||
      status.strictTables.join(',') !==
        'closure_receipts,matrix_obligations' ||
      !status.closureForeignKeyRestricted ||
      !status.requiredTriggersPresent
    ) {
      throw new Error('SQLITE_FOUNDATION_INVARIANT_FAILURE');
    }
    const violations = this.db.prepare('PRAGMA foreign_key_check').all();
    if (violations.length > 0) {
      throw new Error('SQLITE_FOREIGN_KEY_CHECK_FAILURE');
    }
  }

  private seedTrustRegistry(): void {
    const now = new Date().toISOString();
    const insert = this.db.prepare(`
      INSERT INTO trust_keys (key_id, purpose, registry_version, revoked, created_at)
      VALUES (?, ?, ?, ?, ?)
      ON CONFLICT(key_id, purpose) DO NOTHING
    `);
    insert.run(this.secrets.activeKeyId, 'LEASE_HMAC', 0, 0, now);
    insert.run(this.secrets.activeKeyId, 'RECEIPT_HMAC', 0, 0, now);
    insert.run(this.secrets.registryKeyId, 'REGISTRY_HMAC', 0, 0, now);
  }

  topologyHash(): string {
    const rows = this.db
      .prepare(
        `SELECT entry_hash, status, source, entry_json, created_at
         FROM shadow_ledger WHERE status = 'ACTIVE' ORDER BY entry_hash`
      )
      .all();
    return sha256Canonical(rows);
  }

  listLedger(includeQuarantine = false): unknown[] {
    const rows = this.db
      .prepare(
        `SELECT entry_json FROM shadow_ledger
         ${includeQuarantine ? '' : "WHERE status = 'ACTIVE'"}
         ORDER BY created_at, entry_hash`
      )
      .all() as Array<{ entry_json: string }>;
    return rows.map((row) => JSON.parse(row.entry_json));
  }

  openMatrixObligation(input: {
    obligationId?: string;
    sourceNodeId: string;
    targetClass: string;
    mediator: string;
    graphHash: string;
  }): MatrixObligation {
    if (!SHA256_HEX.test(input.graphHash)) {
      throw new BoundaryError(400, 'GRAPH_HASH_INVALID');
    }
    const binding = {
      sourceNodeId: input.sourceNodeId,
      targetClass: input.targetClass,
      mediator: input.mediator,
      graphHash: input.graphHash,
    };
    const obligationId = sha256Canonical(binding);
    if (input.obligationId && input.obligationId !== obligationId) {
      throw new BoundaryError(400, 'OBLIGATION_DIGEST_MISMATCH');
    }
    const initiatedAt = new Date().toISOString();
    this.db
      .prepare(
        `INSERT INTO matrix_obligations
         (obligation_id, source_node_id, target_class, mediator, graph_hash, graph_binding_status,
          current_state, op_cost, initiated_at)
         VALUES (?, ?, ?, ?, ?, 'BOUND', 'SUSPENDED_STATE', 'INFINITY', ?)`
      )
      .run(
        obligationId,
        input.sourceNodeId,
        input.targetClass,
        input.mediator,
        input.graphHash,
        initiatedAt,
      );
    return {
      obligationId,
      ...binding,
      graphBindingStatus: 'BOUND',
      currentState: 'SUSPENDED_STATE',
      opCost: 'INFINITY',
      initiatedAt,
    };
  }

  getMatrixObligation(obligationId: string): MatrixObligation {
    const row = this.db
      .prepare(
        `SELECT obligation_id, source_node_id, target_class, mediator, graph_hash, graph_binding_status,
                current_state, op_cost, initiated_at
         FROM matrix_obligations WHERE obligation_id = ?`
      )
      .get(obligationId) as
      | {
          obligation_id: string;
          source_node_id: string;
          target_class: string;
          mediator: string;
          graph_hash: string | null;
          graph_binding_status: 'LEGACY_UNBOUND' | 'BOUND';
          current_state: MatrixObligationState;
          op_cost: string;
          initiated_at: string;
        }
      | undefined;
    if (!row) throw new BoundaryError(404, 'MATRIX_OBLIGATION_NOT_FOUND');
    return {
      obligationId: row.obligation_id,
      sourceNodeId: row.source_node_id,
      targetClass: row.target_class,
      mediator: row.mediator,
      graphHash: row.graph_hash,
      graphBindingStatus: row.graph_binding_status,
      currentState: row.current_state,
      opCost: row.op_cost,
      initiatedAt: row.initiated_at,
    };
  }

  listMatrixObligations(): MatrixObligation[] {
    const rows = this.db
      .prepare(
        `SELECT obligation_id FROM matrix_obligations
         ORDER BY initiated_at, obligation_id`
      )
      .all() as Array<{ obligation_id: string }>;
    return rows.map((row) => this.getMatrixObligation(row.obligation_id));
  }

  // NOTE: applyClosureReceipt must be called inside executeOnce
  // since it does NOT start its own transaction but relies on executeOnce's BEGIN IMMEDIATE.
  applyClosureReceipt(input: {
    obligationId: string;
    resolvedByNode: string;
    witnessDigest: string;
    resultDigest: string;
    graphHash: string;
    propagationHopLimit: number;
    resolutionState: Exclude<MatrixObligationState, 'SUSPENDED_STATE'>;
    finiteCost: number;
  }): { receipt: ClosureReceipt_v03; obligation: MatrixObligation } {
    if (!this.db.isTransaction) {
      throw new BoundaryError(500, 'TRANSACTION_CONTEXT_REQUIRED');
    }
    const obligation = this.getMatrixObligation(input.obligationId);
    if (obligation.currentState !== 'SUSPENDED_STATE') {
      throw new BoundaryError(409, 'MATRIX_OBLIGATION_ALREADY_TERMINAL');
    }
    if (obligation.graphBindingStatus === 'LEGACY_UNBOUND') {
      throw new BoundaryError(409, 'LEGACY_OBLIGATION_READ_ONLY');
    }
    if (obligation.graphHash !== input.graphHash) {
      throw new BoundaryError(409, 'CLOSURE_RECEIPT_GRAPH_MISMATCH');
    }
    if (
      !SHA256_HEX.test(input.witnessDigest) ||
      !SHA256_HEX.test(input.resultDigest) ||
      !Number.isSafeInteger(input.propagationHopLimit) ||
      input.propagationHopLimit < 0 ||
      input.propagationHopLimit > 64 ||
      !Number.isSafeInteger(input.finiteCost) ||
      input.finiteCost < 0 ||
      (input.resolutionState !== 'CLEARED_BY_RECEIPT' &&
        input.resolutionState !== 'RESOLVED_BY_WITNESS') ||
      (input.resolutionState === 'CLEARED_BY_RECEIPT' && input.finiteCost !== 0)
    ) {
      throw new BoundaryError(400, 'CLOSURE_RECEIPT_INPUT_INVALID');
    }
    if (
      typeof input.resolvedByNode !== 'string' ||
      input.resolvedByNode.length < 1 ||
      input.resolvedByNode.length > 256
    ) {
      throw new BoundaryError(400, 'RESOLVER_NODE_INVALID');
    }

    const keyRow = this.db.prepare(
      `SELECT revoked FROM trust_keys WHERE key_id = ? AND purpose = 'RECEIPT_HMAC'`
    ).get(this.secrets.activeKeyId) as { revoked: number } | undefined;
    if (!keyRow || keyRow.revoked) {
      throw new BoundaryError(403, 'RECEIPT_KEY_REVOKED');
    }

    const body = {
      profile: 'CLOSURE_RECEIPT_v0.3' as const,
      obligationId: input.obligationId,
      resolvedByNode: input.resolvedByNode,
      witnessDigest: input.witnessDigest,
      resultDigest: input.resultDigest,
      graphHash: input.graphHash,
      verifierAlgorithm: 'HMAC-SHA256' as const,
      verifierKeyId: this.secrets.activeKeyId,
      propagationHopLimit: input.propagationHopLimit,
      resolutionState: input.resolutionState,
      finiteCost: input.finiteCost,
      recordedAt: new Date().toISOString(),
      authorityEffect: 'NONE' as const,
    };
    const receiptId = sha256Canonical(body);
    const receipt: ClosureReceipt_v03 = {
      ...body,
      receiptId,
      verifierMac: hmacHex(
        this.secrets.receiptSecret,
        CLOSURE_RECEIPT_DOMAIN_V3,
        { receiptId, ...body },
      ),
    };
    this.db
      .prepare(
        `INSERT INTO closure_receipts
         (receipt_id, obligation_id, resolved_by_node, witness_digest, graph_hash, graph_binding_status,
          verifier_algorithm, verifier_key_id, verifier_key_purpose, verifier_mac,
          propagation_hop_limit, resolution_state, finite_cost,
          receipt_payload_json, recorded_at, result_digest)
         VALUES (?, ?, ?, ?, ?, 'BOUND', ?, ?, 'RECEIPT_HMAC', ?, ?, ?, ?, ?, ?, ?)`
      )
      .run(
        receipt.receiptId,
        receipt.obligationId,
        receipt.resolvedByNode,
        receipt.witnessDigest,
        receipt.graphHash,
        receipt.verifierAlgorithm,
        receipt.verifierKeyId,
        receipt.verifierMac,
        receipt.propagationHopLimit,
        receipt.resolutionState,
        receipt.finiteCost,
        canonicalJson(receipt),
        receipt.recordedAt,
        receipt.resultDigest,
      );
    return {
      receipt,
      obligation: this.getMatrixObligation(input.obligationId),
    };
  }

  createRegistryUpdate(input: {
    keyId: string;
    purpose: 'LEASE_HMAC' | 'RECEIPT_HMAC' | 'REGISTRY_HMAC';
    keyVersion: number;
    revokedAt: string;
    reason: string;
  }): RegistryUpdate_v01 {
    const body = {
      profile: 'REGISTRY_UPDATE_v0.1' as const,
      keyId: input.keyId,
      purpose: input.purpose,
      keyVersion: input.keyVersion,
      revoked: true as const,
      revokedAt: input.revokedAt,
      reason: input.reason,
      issuerKeyId: this.secrets.registryKeyId,
      algorithm: 'HMAC-SHA256' as const,
      macDomain: 'ARK_REGISTRY_UPDATE_V1' as const,
    };
    const updateId = sha256Canonical(body);
    return {
      ...body,
      updateId,
      mac: hmacHex(this.secrets.registrySecret, body.macDomain, {
        updateId,
        ...body,
      }),
    };
  }

  // NOTE: applyRegistryUpdate must be called inside executeOnce as well
  applyRegistryUpdate(update: RegistryUpdate_v01): { status: string } {
    if (!this.db.isTransaction) {
      throw new BoundaryError(500, 'TRANSACTION_CONTEXT_REQUIRED');
    }
    if (update.profile !== 'REGISTRY_UPDATE_v0.1') throw new BoundaryError(400, 'INVALID_PROFILE');
    if (update.algorithm !== 'HMAC-SHA256') throw new BoundaryError(400, 'INVALID_ALGORITHM');
    if (update.macDomain !== 'ARK_REGISTRY_UPDATE_V1') throw new BoundaryError(400, 'INVALID_MAC_DOMAIN');
    if (update.revoked !== true) throw new BoundaryError(400, 'REVOCATION_MUST_BE_TRUE');
    if (
      typeof update.keyId !== 'string' ||
      update.keyId.length < 1 ||
      update.keyId.length > 128 ||
      !['LEASE_HMAC', 'RECEIPT_HMAC', 'REGISTRY_HMAC'].includes(
        update.purpose,
      ) ||
      !Number.isSafeInteger(update.keyVersion) ||
      update.keyVersion < 1 ||
      typeof update.reason !== 'string' ||
      update.reason.length < 1 ||
      update.reason.length > 512 ||
      typeof update.issuerKeyId !== 'string' ||
      update.issuerKeyId.length < 1 ||
      update.issuerKeyId.length > 128 ||
      !Number.isFinite(Date.parse(update.revokedAt)) ||
      new Date(update.revokedAt).toISOString() !== update.revokedAt ||
      !SHA256_HEX.test(update.updateId) ||
      !SHA256_HEX.test(update.mac)
    ) {
      throw new BoundaryError(400, 'REGISTRY_UPDATE_INPUT_INVALID');
    }

    const { mac, updateId, ...body } = update;
    const expectedUpdateId = sha256Canonical(body);
    if (updateId !== expectedUpdateId) throw new BoundaryError(400, 'INVALID_UPDATE_ID');

    const expectedMac = hmacHex(this.secrets.registrySecret, update.macDomain, { updateId, ...body });
    if (!constantTimeHexEqual(mac, expectedMac)) throw new BoundaryError(403, 'INVALID_MAC');

    if (update.issuerKeyId !== this.secrets.registryKeyId) {
      throw new BoundaryError(403, 'REGISTRY_ISSUER_REJECTED');
    }
    const issuer = this.db
      .prepare(
        `SELECT revoked FROM trust_keys
         WHERE key_id = ? AND purpose = 'REGISTRY_HMAC'`,
      )
      .get(update.issuerKeyId) as { revoked: number } | undefined;
    if (!issuer || issuer.revoked) {
      throw new BoundaryError(403, 'REGISTRY_ISSUER_REVOKED');
    }
    if (
      update.issuerKeyId === update.keyId &&
      update.purpose === 'REGISTRY_HMAC'
    ) {
      throw new BoundaryError(403, 'REGISTRY_KEY_CANNOT_REVOKE_ITSELF');
    }

    const current = this.db.prepare(
      `SELECT registry_version, last_update_id FROM trust_keys WHERE key_id = ? AND purpose = ?`
    ).get(update.keyId, update.purpose) as { registry_version: number, last_update_id: string | null } | undefined;

    const currentVersion = current ? current.registry_version : -1;
    if (update.keyVersion < currentVersion) {
      throw new BoundaryError(409, 'STALE_REGISTRY_UPDATE');
    }
    if (update.keyVersion === currentVersion) {
      if (current?.last_update_id === updateId) {
        return { status: 'ALREADY_APPLIED' };
      }
      throw new BoundaryError(409, 'REGISTRY_UPDATE_CONFLICT');
    }

    const appliedAt = new Date().toISOString();
    this.db
      .prepare(
        `INSERT INTO registry_updates
         (update_id, key_id, purpose, key_version, revoked, revoked_at, reason,
          issuer_key_id, algorithm, mac_domain, mac, update_json, applied_at)
         VALUES (?, ?, ?, ?, 1, ?, ?, ?, ?, ?, ?, ?, ?)`,
      )
      .run(
        update.updateId,
        update.keyId,
        update.purpose,
        update.keyVersion,
        update.revokedAt,
        update.reason,
        update.issuerKeyId,
        update.algorithm,
        update.macDomain,
        update.mac,
        canonicalJson(update),
        appliedAt,
      );
    this.db.prepare(
      `INSERT INTO trust_keys (key_id, purpose, registry_version, revoked, last_update_id, created_at)
       VALUES (?, ?, ?, 1, ?, ?)
       ON CONFLICT(key_id, purpose) DO UPDATE SET
         registry_version = excluded.registry_version,
         revoked = excluded.revoked,
         last_update_id = excluded.last_update_id`
    ).run(
      update.keyId,
      update.purpose,
      update.keyVersion,
      updateId,
      appliedAt,
    );

    return { status: 'APPLIED' };
  }

  getClosureReceipt(receiptId: string): unknown {
    const row = this.db
      .prepare(
        `SELECT receipt_payload_json FROM closure_receipts WHERE receipt_id = ?`
      )
      .get(receiptId) as { receipt_payload_json: string } | undefined;
    if (!row) throw new BoundaryError(404, 'CLOSURE_RECEIPT_NOT_FOUND');
    return JSON.parse(row.receipt_payload_json);
  }

  verifyClosureReceipt(candidate: unknown): boolean {
    if (!candidate || typeof candidate !== 'object') return false;
    const receipt = candidate as ClosureReceipt;
    if (
      (receipt.profile !== 'CLOSURE_RECEIPT_v0.2' &&
        receipt.profile !== 'CLOSURE_RECEIPT_v0.3') ||
      receipt.verifierAlgorithm !== 'HMAC-SHA256'
    ) {
      return false;
    }
    if (
      receipt.profile === 'CLOSURE_RECEIPT_v0.3' &&
      !SHA256_HEX.test(receipt.resultDigest)
    ) {
      return false;
    }

    const keyRow = this.db.prepare(
      `SELECT revoked FROM trust_keys WHERE key_id = ? AND purpose = 'RECEIPT_HMAC'`
    ).get(receipt.verifierKeyId) as { revoked: number } | undefined;

    if (!keyRow || keyRow.revoked) {
      return false;
    }

    const { verifierMac, receiptId, ...body } = receipt;
    const domain =
      receipt.profile === 'CLOSURE_RECEIPT_v0.3'
        ? CLOSURE_RECEIPT_DOMAIN_V3
        : CLOSURE_RECEIPT_DOMAIN_V2;
    return (
      receiptId === sha256Canonical(body) &&
      constantTimeHexEqual(
        verifierMac,
        hmacHex(
          this.secrets.receiptSecret,
          domain,
          { receiptId, ...body },
        ),
      )
    );
  }

  appendLedgerEntry(
    entry: unknown,
    status: 'ACTIVE' | 'QUARANTINE',
    source: string,
  ): string {
    const entryJson = canonicalJson(entry);
    const entryHash = sha256Bytes(entryJson);
    this.db
      .prepare(
        `INSERT INTO shadow_ledger
         (entry_hash, status, source, entry_json, created_at)
         VALUES (?, ?, ?, ?, ?)`
      )
      .run(entryHash, status, source, entryJson, new Date().toISOString());
    return entryHash;
  }

  promoteQuarantine(entryHashes: string[]): number {
    if (
      entryHashes.length === 0 ||
      entryHashes.some((entryHash) => !SHA256_HEX.test(entryHash))
    ) {
      throw new BoundaryError(400, 'INVALID_QUARANTINE_PROMOTION_SET');
    }
    const update = this.db.prepare(
      `UPDATE shadow_ledger SET status = 'ACTIVE'
       WHERE entry_hash = ? AND status = 'QUARANTINE'`
    );
    let promoted = 0;
    for (const entryHash of [...new Set(entryHashes)].sort()) {
      const result = update.run(entryHash);
      promoted += Number(result.changes);
    }
    if (promoted !== new Set(entryHashes).size) {
      throw new BoundaryError(409, 'QUARANTINE_PROMOTION_TARGET_MISSING');
    }
    return promoted;
  }

  executeOnce<T>(
    lease: AuthenticatedLease,
    operation: string,
    mutate: () => T,
  ): TransactionResult<T> {
    this.db.exec('BEGIN IMMEDIATE;');
    try {
      const existing = this.db
        .prepare(
          `SELECT status, response_json, receipt_id
           FROM consumed_nonces WHERE nonce = ?`
        )
        .get(lease.nonce) as
        | { status: string; response_json: string | null; receipt_id: string | null }
        | undefined;

      if (existing) {
        this.db.exec('ROLLBACK;');
        if (
          existing.status === 'COMMITTED' &&
          existing.response_json &&
          existing.receipt_id
        ) {
          const receipt = this.getReceipt(existing.receipt_id);
          return {
            response: JSON.parse(existing.response_json) as T,
            replayed: true,
            receipt,
          };
        }
        throw new BoundaryError(409, 'NONCE_ALREADY_CLAIMED');
      }

      if (!lease.topologyHash || lease.topologyHash !== this.topologyHash()) {
        throw new BoundaryError(409, 'TOPOLOGY_DRIFT_ERROR');
      }

      this.db
        .prepare(
          `INSERT INTO consumed_nonces
           (nonce, lease_id, target_resource, payload_digest, status)
           VALUES (?, ?, ?, ?, 'CLAIMED')`
        )
        .run(lease.nonce, lease.id, lease.targetResource, lease.payloadDigest);

      const response = mutate();
      const responseJson = canonicalJson(response);
      const receipt = this.createReceipt(
        operation,
        lease.targetResource,
        lease.payloadDigest,
        sha256Bytes(responseJson),
        lease.nonce,
      );
      this.insertReceipt(receipt);
      this.db
        .prepare(
          `UPDATE consumed_nonces
           SET status = 'COMMITTED', response_json = ?, receipt_id = ?, committed_at = ?
           WHERE nonce = ? AND status = 'CLAIMED'`
        )
        .run(
          responseJson,
          receipt.receiptId,
          new Date().toISOString(),
          lease.nonce,
        );
      this.db.exec('COMMIT;');
      return { response, replayed: false, receipt };
    } catch (error) {
      try {
        this.db.exec('ROLLBACK;');
      } catch {
        // The duplicate replay branch already rolled back.
      }
      throw error;
    }
  }

  beginExternal(
    lease: AuthenticatedLease,
  ): { state: 'NEW' | 'COMMITTED'; response?: unknown } {
    this.db.exec('BEGIN IMMEDIATE;');
    try {
      const existing = this.db
        .prepare(
          `SELECT status, response_json FROM external_operations WHERE nonce = ?`
        )
        .get(lease.nonce) as
        | { status: 'PENDING' | 'COMMITTED' | 'FAILED'; response_json: string | null }
        | undefined;
      if (existing) {
        this.db.exec('ROLLBACK;');
        if (existing.status === 'COMMITTED' && existing.response_json) {
          return { state: 'COMMITTED', response: JSON.parse(existing.response_json) };
        }
        throw new BoundaryError(
          409,
          existing.status === 'PENDING'
            ? 'EXTERNAL_OPERATION_PENDING'
            : 'EXTERNAL_OPERATION_FAILED',
        );
      }
      if (!lease.topologyHash || lease.topologyHash !== this.topologyHash()) {
        throw new BoundaryError(409, 'TOPOLOGY_DRIFT_ERROR');
      }
      this.db
        .prepare(
          `INSERT INTO external_operations
           (nonce, lease_id, target_resource, payload_digest, status, updated_at)
           VALUES (?, ?, ?, ?, 'PENDING', ?)`
        )
        .run(
          lease.nonce,
          lease.id,
          lease.targetResource,
          lease.payloadDigest,
          new Date().toISOString(),
        );
      this.db.exec('COMMIT;');
      return { state: 'NEW' };
    } catch (error) {
      try {
        this.db.exec('ROLLBACK;');
      } catch {
        // Already rolled back.
      }
      throw error;
    }
  }

  completeExternal<T>(
    lease: AuthenticatedLease,
    operation: string,
    response: T,
  ): TransactionResult<T> {
    this.db.exec('BEGIN IMMEDIATE;');
    try {
      const external = this.db
        .prepare(
          `SELECT status FROM external_operations
           WHERE nonce = ? AND lease_id = ?`
        )
        .get(lease.nonce, lease.id) as { status: string } | undefined;
      if (!external || external.status !== 'PENDING') {
        throw new BoundaryError(409, 'EXTERNAL_OPERATION_NOT_PENDING');
      }

      this.db
        .prepare(
          `INSERT INTO consumed_nonces
           (nonce, lease_id, target_resource, payload_digest, status)
           VALUES (?, ?, ?, ?, 'CLAIMED')`
        )
        .run(lease.nonce, lease.id, lease.targetResource, lease.payloadDigest);
      const responseJson = canonicalJson(response);
      const receipt = this.createReceipt(
        operation,
        lease.targetResource,
        lease.payloadDigest,
        sha256Bytes(responseJson),
        lease.nonce,
      );
      this.insertReceipt(receipt);
      this.db
        .prepare(
          `UPDATE consumed_nonces
           SET status = 'COMMITTED', response_json = ?, receipt_id = ?, committed_at = ?
           WHERE nonce = ?`
        )
        .run(
          responseJson,
          receipt.receiptId,
          new Date().toISOString(),
          lease.nonce,
        );
      this.db
        .prepare(
          `UPDATE external_operations
           SET status = 'COMMITTED', response_json = ?, updated_at = ?
           WHERE nonce = ?`
        )
        .run(responseJson, new Date().toISOString(), lease.nonce);
      this.db.exec('COMMIT;');
      return { response, replayed: false, receipt };
    } catch (error) {
      this.db.exec('ROLLBACK;');
      throw error;
    }
  }

  failExternal(lease: AuthenticatedLease, errorCode: string): void {
    this.db
      .prepare(
        `UPDATE external_operations
         SET status = 'FAILED', error_code = ?, updated_at = ?
         WHERE nonce = ? AND lease_id = ? AND status = 'PENDING'`
      )
      .run(errorCode, new Date().toISOString(), lease.nonce, lease.id);
  }

  private createReceipt(
    operation: string,
    targetResource: string,
    payloadDigest: string,
    resultDigest: string,
    nonce: string,
  ): AuthenticatedReceipt {
    const body = {
      algorithm: 'HMAC-SHA256' as const,
      keyId: this.secrets.activeKeyId,
      operation,
      targetResource,
      payloadDigest,
      resultDigest,
      nonce,
      createdAt: new Date().toISOString(),
    };
    const receiptId = sha256Canonical(body);
    return {
      receiptId,
      ...body,
      mac: hmacHex(this.secrets.receiptSecret, RECEIPT_DOMAIN, {
        receiptId,
        ...body,
      }),
    };
  }

  private insertReceipt(receipt: AuthenticatedReceipt): void {
    this.db
      .prepare(
        `INSERT INTO audit_receipts
         (receipt_id, operation, target_resource, payload_digest, result_digest,
          nonce, key_id, algorithm, receipt_json, mac, created_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
      )
      .run(
        receipt.receiptId,
        receipt.operation,
        receipt.targetResource,
        receipt.payloadDigest,
        receipt.resultDigest,
        receipt.nonce,
        receipt.keyId,
        receipt.algorithm,
        canonicalJson(receipt),
        receipt.mac,
        receipt.createdAt,
      );
  }

  getReceipt(receiptId: string): AuthenticatedReceipt {
    const row = this.db
      .prepare(`SELECT receipt_json FROM audit_receipts WHERE receipt_id = ?`)
      .get(receiptId) as { receipt_json: string } | undefined;
    if (!row) throw new BoundaryError(500, 'AUDIT_RECEIPT_MISSING');
    return JSON.parse(row.receipt_json) as AuthenticatedReceipt;
  }

  verifyReceipt(receipt: AuthenticatedReceipt): boolean {
    if (
      receipt.algorithm !== 'HMAC-SHA256'
    ) {
      return false;
    }

    const keyRow = this.db.prepare(
      `SELECT revoked FROM trust_keys WHERE key_id = ? AND purpose = 'RECEIPT_HMAC'`
    ).get(receipt.keyId) as { revoked: number } | undefined;

    if (!keyRow || keyRow.revoked) {
      return false;
    }

    const { mac, ...body } = receipt;
    return constantTimeHexEqual(
      mac,
      hmacHex(this.secrets.receiptSecret, RECEIPT_DOMAIN, body),
    );
  }
}

export function defaultArkDatabasePath(): string {
  return process.env.ARK_DB_PATH || path.join(process.cwd(), 'data', 'ark-boundary.sqlite');
}
