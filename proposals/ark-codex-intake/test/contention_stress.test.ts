import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { Worker } from 'node:worker_threads';
import {
  ArkBoundaryStore,
  AuthenticatedLease,
  BoundarySecrets,
  LeaseRegistry,
  sha256Canonical,
} from '../src/boundary';

const secrets: BoundarySecrets = {
  authSecret: 'contention-auth-secret',
  signingSecret: 'contention-signing-secret',
  receiptSecret: 'contention-receipt-secret',
  registrySecret: 'contention-registry-secret',
  activeKeyId: 'contention-key-v1',
  registryKeyId: 'contention-registry-key-v1',
};

interface WorkerResult {
  response: Record<string, unknown>;
  replayed: boolean;
  receipt: { receiptId: string; nonce: string };
}

function binding(index: number) {
  const value = {
    sourceNodeId: `Node_${index}`,
    targetClass: 'SWARM_NODE',
    mediator: 'GOSSIP_PROTOCOL',
    graphHash: '0'.repeat(64),
  };
  return { ...value, obligationId: sha256Canonical(value) };
}

function issue(leases: LeaseRegistry, index: number): AuthenticatedLease {
  const payloadDigest = sha256Canonical({ contentionFixture: index });
  return leases.issue({
    type: 'APPLY_WINDOW',
    targetResource: '/api/test/concurrent-contention',
    payloadDigest,
    scope: 'test:concurrent-contention',
    notBefore: new Date(Date.now() - 1_000).toISOString(),
    expiresAt: new Date(Date.now() + 60_000).toISOString(),
    topologyHash: sha256Canonical([]),
  });
}

async function runBurst(
  databasePath: string,
  inputs: Array<{
    lease: AuthenticatedLease;
    operation: string;
    response: Record<string, unknown>;
    binding: ReturnType<typeof binding>;
  }>,
): Promise<WorkerResult[]> {
  const workerUrl = new URL('./support/contention-worker.ts', import.meta.url);
  const workers = inputs.map(
    (input) =>
      new Worker(workerUrl, {
        workerData: {
          databasePath,
          secrets: {
            ...secrets,
          },
          ...input,
        },
      }),
  );

  try {
    await Promise.all(
      workers.map(
        (worker) =>
          new Promise<void>((resolve, reject) => {
            const onMessage = (message: { type?: string; error?: string }) => {
              if (message.type === 'READY') {
                worker.off('error', reject);
                worker.off('message', onMessage);
                resolve();
              } else if (message.type === 'ERROR') {
                reject(new Error(message.error));
              }
            };
            worker.on('message', onMessage);
            worker.once('error', reject);
          }),
      ),
    );

    const results = workers.map(
      (worker) =>
        new Promise<WorkerResult>((resolve, reject) => {
          worker.on(
            'message',
            (message: {
              type?: string;
              result?: WorkerResult;
              error?: string;
            }) => {
              if (message.type === 'RESULT' && message.result) {
                resolve(message.result);
              } else if (message.type === 'ERROR') {
                reject(new Error(message.error));
              }
            },
          );
          worker.once('error', reject);
        }),
    );
    for (const worker of workers) worker.postMessage({ type: 'START' });
    return await Promise.all(results);
  } finally {
    await Promise.all(workers.map((worker) => worker.terminate()));
  }
}

test(
  'concurrent unique writers serialize without lost commits or broken receipts',
  { timeout: 30_000 },
  async () => {
    const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'ark-contention-'));
    const databasePath = path.join(directory, 'ark.sqlite');
    const bootstrap = new ArkBoundaryStore(databasePath, secrets);
    const leases = new LeaseRegistry(secrets, bootstrap.db);
    const writerCount = 16;

    // Pre-issue all leases before closing the bootstrap DB
    const inputs = Array.from({ length: writerCount }, (_, index) => ({
      lease: issue(leases, index),
      operation: 'CONTENTION_UNIQUE_WRITE',
      response: { index, committed: true },
      binding: binding(index),
    }));
    bootstrap.db.close();

    try {
      const results = await runBurst(
        databasePath,
        inputs
      );
      assert.equal(results.length, writerCount);
      assert.equal(results.filter((result) => result.replayed).length, 0);
      assert.equal(new Set(results.map((result) => result.receipt.receiptId)).size, writerCount);

      const audit = new ArkBoundaryStore(databasePath, secrets);
      try {
        const counts = audit.db
          .prepare(
            `SELECT
               (SELECT COUNT(*) FROM consumed_nonces WHERE status = 'COMMITTED') AS nonces,
               (SELECT COUNT(*) FROM audit_receipts) AS receipts,
               (SELECT COUNT(*) FROM matrix_obligations
                WHERE current_state = 'SUSPENDED_STATE' AND op_cost = 'INFINITY') AS obligations`,
          )
          .get() as { nonces: number; receipts: number; obligations: number };
        assert.deepEqual(
          {
            nonces: Number(counts.nonces),
            receipts: Number(counts.receipts),
            obligations: Number(counts.obligations),
          },
          {
            nonces: writerCount,
            receipts: writerCount,
            obligations: writerCount,
          },
        );
        assert.equal(
          audit.db.prepare('PRAGMA integrity_check').get()!.integrity_check,
          'ok',
        );
        for (const result of results) {
          assert.equal(audit.verifyReceipt(audit.getReceipt(result.receipt.receiptId)), true);
        }
      } finally {
        audit.db.close();
      }
    } finally {
      fs.rmSync(directory, { recursive: true, force: true });
    }
  },
);

test(
  'concurrent replay of one nonce commits one mutation and returns one deterministic result',
  { timeout: 30_000 },
  async () => {
    const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'ark-replay-race-'));
    const databasePath = path.join(directory, 'ark.sqlite');
    const bootstrap = new ArkBoundaryStore(databasePath, secrets);
    const leases = new LeaseRegistry(secrets, bootstrap.db);
    const sharedLease = issue(leases, 999);
    bootstrap.db.close();

    const sharedBinding = binding(999);
    const contenderCount = 12;

    try {
      const results = await runBurst(
        databasePath,
        Array.from({ length: contenderCount }, (_, index) => ({
          lease: sharedLease,
          operation: 'CONTENTION_REPLAY_WRITE',
          response: { winningContender: index },
          binding: sharedBinding,
        })),
      );
      assert.equal(results.filter((result) => !result.replayed).length, 1);
      assert.equal(results.filter((result) => result.replayed).length, contenderCount - 1);
      assert.equal(new Set(results.map((result) => JSON.stringify(result.response))).size, 1);
      assert.equal(new Set(results.map((result) => result.receipt.receiptId)).size, 1);

      const audit = new ArkBoundaryStore(databasePath, secrets);
      try {
        const counts = audit.db
          .prepare(
            `SELECT
               (SELECT COUNT(*) FROM consumed_nonces WHERE nonce = ?) AS nonces,
               (SELECT COUNT(*) FROM audit_receipts WHERE nonce = ?) AS receipts,
               (SELECT COUNT(*) FROM matrix_obligations WHERE obligation_id = ?) AS obligations`,
          )
          .get(sharedLease.nonce, sharedLease.nonce, sharedBinding.obligationId) as {
          nonces: number;
          receipts: number;
          obligations: number;
        };
        assert.deepEqual(
          {
            nonces: Number(counts.nonces),
            receipts: Number(counts.receipts),
            obligations: Number(counts.obligations),
          },
          { nonces: 1, receipts: 1, obligations: 1 },
        );
        assert.equal(
          audit.db.prepare('PRAGMA integrity_check').get()!.integrity_check,
          'ok',
        );
      } finally {
        audit.db.close();
      }
    } finally {
      fs.rmSync(directory, { recursive: true, force: true });
    }
  },
);
