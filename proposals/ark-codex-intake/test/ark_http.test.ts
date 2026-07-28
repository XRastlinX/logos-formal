import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import type { Server } from 'node:http';
import { sha256Canonical } from '../src/boundary';

test('HTTP apply requires authentication, exact permit binding, and replays prior result', async () => {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'ark-http-'));
  process.env.AUTH_SECRET = 'http-auth-secret';
  process.env.SIGNING_SECRET = 'http-signing-secret';
  process.env.RECEIPT_SECRET = 'http-receipt-secret';
  process.env.ACTIVE_KEY_ID = 'http-key-v1';
  process.env.REGISTRY_SECRET = 'http-registry-secret';
  process.env.REGISTRY_KEY_ID = 'http-registry-key-v1';
  process.env.ARK_PLACE = 'FOUNDRY';
  process.env.ARK_DB_PATH = path.join(directory, 'ark.sqlite');

  const { app, store } = await import('../server');
  const server = await new Promise<Server>((resolve) => {
    const listening = app.listen(0, '127.0.0.1', () => resolve(listening));
  });
  try {
    const address = server.address();
    assert(address && typeof address === 'object');
    const base = `http://127.0.0.1:${address.port}`;
    const payload = { frameData: null, structuralModification: null };
    const requestBody = {
      frame_data: null,
      structural_modification: null,
    };

    const unauthorized = await fetch(`${base}/api/foundry/lease/issue`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({}),
    });
    assert.equal(unauthorized.status, 401);

    const leaseResponse = await fetch(`${base}/api/foundry/lease/issue`, {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
        'x-sovereign-token': process.env.AUTH_SECRET,
      },
      body: JSON.stringify({
        type: 'APPLY_WINDOW',
        targetResource: '/api/foundry/apply',
        payloadDigest: sha256Canonical(payload),
        scope: 'ledger:apply',
        notBefore: new Date(Date.now() - 1_000).toISOString(),
        expiresAt: new Date(Date.now() + 60_000).toISOString(),
      }),
    });
    assert.equal(leaseResponse.status, 200);
    const { lease } = (await leaseResponse.json()) as { lease: unknown };

    const apply = () =>
      fetch(`${base}/api/foundry/apply`, {
        method: 'POST',
        headers: {
          'content-type': 'application/json',
          'x-sovereign-token': process.env.AUTH_SECRET,
        },
        body: JSON.stringify({ ...requestBody, lease }),
      });

    const first = await apply();
    assert.equal(first.status, 200);
    const firstBody = (await first.json()) as Record<string, unknown>;
    assert.equal(firstBody.replayed, false);
    assert.equal(firstBody.status, 'APPLY_COMMITTED');

    const replay = await apply();
    assert.equal(replay.status, 200);
    const replayBody = (await replay.json()) as Record<string, unknown>;
    assert.equal(replayBody.replayed, true);
    assert.equal(replayBody.entryHash, firstBody.entryHash);

    const binding = {
      sourceNodeId: 'Node_A',
      targetClass: 'SWARM_NODE',
      mediator: 'GOSSIP_PROTOCOL',
      graphHash: '0'.repeat(64),
    };
    const openPayload = {
      obligationId: sha256Canonical(binding),
      ...binding,
    };
    const issueLease = async (
      type: string,
      targetResource: string,
      scope: string,
      boundPayload: unknown,
    ) => {
      const response = await fetch(`${base}/api/foundry/lease/issue`, {
        method: 'POST',
        headers: {
          'content-type': 'application/json',
          'x-sovereign-token': process.env.AUTH_SECRET,
        },
        body: JSON.stringify({
          type,
          targetResource,
          payloadDigest: sha256Canonical(boundPayload),
          scope,
          notBefore: new Date(Date.now() - 1_000).toISOString(),
          expiresAt: new Date(Date.now() + 60_000).toISOString(),
        }),
      });
      assert.equal(response.status, 200);
      return ((await response.json()) as { lease: unknown }).lease;
    };
    const openLease = await issueLease(
      'APPLY_WINDOW',
      '/api/matrix/obligations/open',
      'matrix:obligation:open',
      openPayload,
    );
    const opened = await fetch(`${base}/api/matrix/obligations/open`, {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
        'x-sovereign-token': process.env.AUTH_SECRET,
      },
      body: JSON.stringify({ ...openPayload, lease: openLease }),
    });
    assert.equal(opened.status, 200);
    assert.equal(
      ((await opened.json()) as { obligation: { opCost: string } }).obligation
        .opCost,
      'INFINITY',
    );

    const closePayload = {
      obligationId: openPayload.obligationId,
      resolvedByNode: 'Node_C',
      witnessDigest: sha256Canonical({ witness: 'http' }),
      propagationHopLimit: 2,
      resolutionState: 'RESOLVED_BY_WITNESS',
      finiteCost: 1,
      graphHash: '0'.repeat(64),
    };
    const closeLease = await issueLease(
      'APPLY_WINDOW',
      '/api/matrix/obligations/close',
      'matrix:closure:apply',
      closePayload,
    );
    const close = () =>
      fetch(`${base}/api/matrix/obligations/close`, {
        method: 'POST',
        headers: {
          'content-type': 'application/json',
          'x-sovereign-token': process.env.AUTH_SECRET,
        },
        body: JSON.stringify({ ...closePayload, lease: closeLease }),
      });
    const closed = await close();
    assert.equal(closed.status, 200);
    const closedBody = (await closed.json()) as {
      obligation: { currentState: string; opCost: string };
      closureReceipt: { receiptId: string };
      replayed: boolean;
    };
    assert.equal(closedBody.obligation.currentState, 'RESOLVED_BY_WITNESS');
    assert.equal(closedBody.obligation.opCost, '1');
    assert.equal(closedBody.replayed, false);
    assert.equal(
      store.verifyClosureReceipt(
        store.getClosureReceipt(closedBody.closureReceipt.receiptId),
      ),
      true,
    );
    const closedReplay = await close();
    assert.equal(closedReplay.status, 200);
    assert.equal(
      ((await closedReplay.json()) as { replayed: boolean }).replayed,
      true,
    );

    const registryUpdate = store.createRegistryUpdate({
      keyId: process.env.ACTIVE_KEY_ID!,
      purpose: 'RECEIPT_HMAC',
      keyVersion: 1,
      revokedAt: new Date().toISOString(),
      reason: 'HTTP registry route assay',
    });
    const registryPayload = { update: registryUpdate };
    const registryLease = await issueLease(
      'APPLY_WINDOW',
      '/api/trust/keys/revoke',
      'trust:key:revoke',
      registryPayload,
    );
    const revoke = () =>
      fetch(`${base}/api/trust/keys/revoke`, {
        method: 'POST',
        headers: {
          'content-type': 'application/json',
          'x-sovereign-token': process.env.AUTH_SECRET,
        },
        body: JSON.stringify({ ...registryPayload, lease: registryLease }),
      });
    const revoked = await revoke();
    assert.equal(revoked.status, 200);
    assert.equal(
      ((await revoked.json()) as { status: string; replayed: boolean }).status,
      'APPLIED',
    );
    const revokedReplay = await revoke();
    const revokedReplayBody = (await revokedReplay.json()) as {
      status: string;
      replayed: boolean;
    };
    assert.equal(revokedReplay.status, 200);
    assert.equal(revokedReplayBody.status, 'APPLIED');
    assert.equal(revokedReplayBody.replayed, true);
  } finally {
    await new Promise<void>((resolve, reject) =>
      server.close((error) => (error ? reject(error) : resolve())),
    );
    store.db.close();
    fs.rmSync(directory, { recursive: true, force: true });
  }
});
