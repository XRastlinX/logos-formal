import { parentPort, workerData } from 'node:worker_threads';
import { ArkBoundaryStore, BoundarySecrets } from '../../src/boundary';

const input = workerData as {
  databasePath: string;
  secrets: BoundarySecrets;
  entryHash: string;
};
const store = new ArkBoundaryStore(input.databasePath, input.secrets);

store.db.exec('BEGIN IMMEDIATE;');
store.db
  .prepare(
    `INSERT INTO shadow_ledger
     (entry_hash, status, source, entry_json, created_at)
     VALUES (?, 'ACTIVE', 'CRASH_ASSAY', ?, ?)`,
  )
  .run(
    input.entryHash,
    JSON.stringify({ uncommitted: true }),
    new Date().toISOString(),
  );
store.db
  .prepare(
    `INSERT INTO consumed_nonces
     (nonce, lease_id, target_resource, payload_digest, status)
     VALUES (?, ?, '/api/test/crash', ?, 'CLAIMED')`,
  )
  .run('f'.repeat(32), 'lease_crash_assay', input.entryHash);
parentPort?.postMessage({ type: 'UNCOMMITTED_WRITE_READY' });

// The parent terminates this worker without allowing COMMIT, ROLLBACK, or close.
setInterval(() => undefined, 1_000);
