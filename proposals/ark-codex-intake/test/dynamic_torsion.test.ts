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
  authSecret: 'torsion-auth-secret',
  signingSecret: 'torsion-signing-secret',
  receiptSecret: 'torsion-receipt-secret',
  registrySecret: 'torsion-registry-secret',
  activeKeyId: 'torsion-key-v1',
  registryKeyId: 'torsion-registry-key-v1',
};

const NODE_COUNT = 50;
const HOP_LIMIT = 4;
const MAX_LOGICAL_TICKS = 200;
const SOURCE_NODE = 'Node_01';
const PARTITIONED_EDGE = 'Node_01->Node_02';
const STATIC_DEFORMED_EDGES = new Set([
  'Node_01->Node_02',
  'Node_11->Node_12',
  'Node_21->Node_22',
  'Node_31->Node_32',
  'Node_41->Node_42',
]);

type NodeClass = 'SWARM_NODE' | 'AUX_NODE';
type RoutingMode = 'FIXED_PATH' | 'SCOPED_GOSSIP' | 'ALTERNATE_PATH';

interface GraphDefinition {
  nodes: Array<{ id: string; nodeClass: NodeClass }>;
  edges: Array<{ source: string; target: string; order: number }>;
}

interface ClosureInput {
  obligationId: string;
  resolvedByNode: string;
  witnessDigest: string;
  graphHash: string;
  propagationHopLimit: number;
  resolutionState: 'RESOLVED_BY_WITNESS';
  finiteCost: number;
}

interface RoutedClosure {
  messageId: string;
  source: string;
  target: string;
  hop: number;
  attempt: number;
  deliverAt: number;
  closure: ClosureInput;
}

function nodeId(index: number): string {
  return `Node_${String(index + 1).padStart(2, '0')}`;
}

function buildGraph(): GraphDefinition {
  const nodes = Array.from({ length: NODE_COUNT }, (_, index) => ({
    id: nodeId(index),
    nodeClass: (index + 1) % 13 === 0 ? 'AUX_NODE' : 'SWARM_NODE',
  })) satisfies GraphDefinition['nodes'];
  const edges: GraphDefinition['edges'] = [];
  for (let index = 0; index < NODE_COUNT; index += 1) {
    edges.push({
      source: nodeId(index),
      target: nodeId((index + 1) % NODE_COUNT),
      order: 0,
    });
    edges.push({
      source: nodeId(index),
      target: nodeId((index + 7) % NODE_COUNT),
      order: 1,
    });
  }
  return { nodes, edges };
}

function adjacency(graph: GraphDefinition): Map<string, string[]> {
  const result = new Map<string, string[]>();
  for (const node of graph.nodes) result.set(node.id, []);
  for (const edge of graph.edges) result.get(edge.source)?.push(edge.target);
  return result;
}

function nodeClasses(graph: GraphDefinition): Map<string, NodeClass> {
  return new Map(graph.nodes.map((node) => [node.id, node.nodeClass]));
}

function routeTargets(
  graphAdjacency: Map<string, string[]>,
  mode: RoutingMode,
  source: string,
): string[] {
  const targets = graphAdjacency.get(source) ?? [];
  if (mode === 'SCOPED_GOSSIP') return targets;
  if (
    mode === 'ALTERNATE_PATH' &&
    targets[0] &&
    STATIC_DEFORMED_EDGES.has(`${source}->${targets[0]}`)
  ) {
    return targets[1] ? [targets[1]] : [];
  }
  return targets[0] ? [targets[0]] : [];
}

function expectedAffectedSet(
  graph: GraphDefinition,
  mode: RoutingMode,
  source: string,
  targetClass: NodeClass,
  hopLimit: number,
): Set<string> {
  const graphAdjacency = adjacency(graph);
  const classes = nodeClasses(graph);
  const affected = new Set<string>();
  const queue: Array<{ id: string; hop: number }> = [{ id: source, hop: 0 }];
  const bestHop = new Map<string, number>();
  while (queue.length > 0) {
    const current = queue.shift()!;
    const priorHop = bestHop.get(current.id);
    if (priorHop !== undefined && priorHop <= current.hop) continue;
    bestHop.set(current.id, current.hop);
    if (classes.get(current.id) !== targetClass) continue;
    affected.add(current.id);
    if (current.hop >= hopLimit) continue;
    for (const target of routeTargets(graphAdjacency, mode, current.id)) {
      queue.push({ id: target, hop: current.hop + 1 });
    }
  }
  return affected;
}

class DeterministicTransport {
  private queue: RoutedClosure[] = [];
  private sequence = 0;
  ticks = 0;
  readonly counters = {
    simulatedLosses: 0,
    duplicates: 0,
    delayedMessages: 0,
    partitionHolds: 0,
    delivered: 0,
  };

  enqueue(
    source: string,
    target: string,
    closure: ClosureInput,
    hop: number,
  ): void {
    const sequence = this.sequence;
    this.sequence += 1;
    const messageId = sha256Canonical({
      seed: 'DYNAMIC_TORSION_V2',
      source,
      target,
      obligationId: closure.obligationId,
      hop,
      sequence,
    });
    const delay = 1 + (sequence % 4);
    this.counters.delayedMessages += 1;
    const message: RoutedClosure = {
      messageId,
      source,
      target,
      hop,
      attempt: 0,
      deliverAt: this.ticks + delay,
      closure,
    };

    // Deterministically lose selected first attempts, then retry them. This
    // models loss without violating the assay's declared fair-delivery bound.
    if (sequence % 7 === 0) {
      this.counters.simulatedLosses += 1;
      this.queue.push({
        ...message,
        attempt: 1,
        deliverAt: message.deliverAt + 2,
      });
    } else {
      this.queue.push(message);
    }
    if (sequence % 5 === 0) {
      this.counters.duplicates += 1;
      this.queue.push({
        ...message,
        messageId: `${message.messageId}:duplicate`,
        attempt: 1,
        deliverAt: message.deliverAt + 1,
      });
    }
  }

  run(nodes: Map<string, SimulatedNode>): void {
    while (this.queue.length > 0 && this.ticks < MAX_LOGICAL_TICKS) {
      this.ticks += 1;
      const ready = this.queue
        .filter((message) => message.deliverAt <= this.ticks)
        .sort((left, right) =>
          left.messageId < right.messageId
            ? 1
            : left.messageId === right.messageId
              ? 0
              : -1,
        );
      this.queue = this.queue.filter(
        (message) => message.deliverAt > this.ticks,
      );
      for (const message of ready) {
        if (
          `${message.source}->${message.target}` === PARTITIONED_EDGE &&
          this.ticks <= 6
        ) {
          this.counters.partitionHolds += 1;
          this.queue.push({ ...message, deliverAt: this.ticks + 1 });
          continue;
        }
        const node = nodes.get(message.target);
        if (!node) throw new Error(`UNKNOWN_SIMULATION_NODE:${message.target}`);
        node.receiveClosure(message);
        this.counters.delivered += 1;
      }
    }
    assert.equal(
      this.queue.length,
      0,
      `transport failed to reach quiescence by tick ${MAX_LOGICAL_TICKS}`,
    );
  }
}

class SimulatedNode {
  readonly store: ArkBoundaryStore;
  private readonly leases: LeaseRegistry;
  private readonly closureLeases = new Map<string, AuthenticatedLease>();

  constructor(
    readonly id: string,
    readonly nodeClass: NodeClass,
    databasePath: string,
    private readonly graphAdjacency: Map<string, string[]>,
    private readonly classes: Map<string, NodeClass>,
    private readonly mode: RoutingMode,
    private readonly transport: DeterministicTransport,
  ) {
    this.store = new ArkBoundaryStore(databasePath, secrets);
    this.leases = new LeaseRegistry(secrets, this.store.db);
  }

  open(binding: {
    sourceNodeId: string;
    targetClass: string;
    mediator: string;
    graphHash: string;
  }): string {
    const payload = {
      ...binding,
      obligationId: sha256Canonical(binding),
    };
    const lease = this.issue(
      '/api/matrix/obligations/open',
      'matrix:obligation:open',
      payload,
    );
    return this.store.executeOnce(lease, 'SIMULATION_OPEN', () =>
      this.store.openMatrixObligation(payload),
    ).response.obligationId;
  }

  receiveClosure(message: RoutedClosure): void {
    if (this.nodeClass !== 'SWARM_NODE') {
      throw new Error(`TARGET_CLASS_FILTER_FAILURE:${this.id}`);
    }
    let lease = this.closureLeases.get(message.closure.obligationId);
    if (!lease) {
      lease = this.issue(
        '/api/matrix/obligations/close',
        'matrix:closure:apply',
        message.closure,
      );
      this.closureLeases.set(message.closure.obligationId, lease);
    }
    const result = this.store.executeOnce(
      lease,
      'SIMULATION_CLOSURE',
      () => this.store.applyClosureReceipt(message.closure),
    );
    if (result.replayed || message.hop >= message.closure.propagationHopLimit) {
      return;
    }
    for (const target of routeTargets(
      this.graphAdjacency,
      this.mode,
      this.id,
    )) {
      if (this.classes.get(target) !== 'SWARM_NODE') continue;
      this.transport.enqueue(
        this.id,
        target,
        message.closure,
        message.hop + 1,
      );
    }
  }

  private issue(
    targetResource: string,
    scope: string,
    payload: unknown,
  ): AuthenticatedLease {
    return this.leases.issue({
      type: 'APPLY_WINDOW',
      targetResource,
      payloadDigest: sha256Canonical(payload),
      scope,
      notBefore: new Date(Date.now() - 1_000).toISOString(),
      expiresAt: new Date(Date.now() + 300_000).toISOString(),
      topologyHash: this.store.topologyHash(),
    });
  }

  close(): void {
    this.store.db.close();
  }
}

function runRoutingScenario(mode: RoutingMode) {
  const directory = fs.mkdtempSync(
    path.join(os.tmpdir(), `ark-torsion-${mode.toLowerCase()}-`),
  );
  const graph = buildGraph();
  const graphHash = sha256Canonical(graph);
  const graphAdjacency = adjacency(graph);
  const classes = nodeClasses(graph);
  const transport = new DeterministicTransport();
  const nodes = new Map<string, SimulatedNode>();
  const primaryBinding = {
    sourceNodeId: SOURCE_NODE,
    targetClass: 'SWARM_NODE',
    mediator: 'GOSSIP_PROTOCOL',
    graphHash,
  };
  const orthogonalBinding = {
    sourceNodeId: 'Orthogonal_Source',
    targetClass: 'SWARM_NODE',
    mediator: 'ORTHOGONAL_MEDIATOR',
    graphHash,
  };
  try {
    for (const node of graph.nodes) {
      nodes.set(
        node.id,
        new SimulatedNode(
          node.id,
          node.nodeClass,
          path.join(directory, `${node.id}.sqlite`),
          graphAdjacency,
          classes,
          mode,
          transport,
        ),
      );
    }

    let primaryObligationId = '';
    let orthogonalObligationId = '';
    for (const node of nodes.values()) {
      primaryObligationId = node.open(primaryBinding);
      orthogonalObligationId = node.open(orthogonalBinding);
    }
    const closure: ClosureInput = {
      obligationId: primaryObligationId,
      resolvedByNode: 'Node_50',
      witnessDigest: sha256Canonical({ witness: mode }),
      graphHash,
      propagationHopLimit: HOP_LIMIT,
      resolutionState: 'RESOLVED_BY_WITNESS',
      finiteCost: 1,
    };
    transport.enqueue('HARNESS', SOURCE_NODE, closure, 0);
    transport.run(nodes);

    const expected = expectedAffectedSet(
      graph,
      mode,
      SOURCE_NODE,
      'SWARM_NODE',
      HOP_LIMIT,
    );
    const actual = new Set<string>();
    for (const [id, node] of nodes) {
      const primary = node.store.getMatrixObligation(primaryObligationId);
      const orthogonal =
        node.store.getMatrixObligation(orthogonalObligationId);
      if (primary.currentState === 'RESOLVED_BY_WITNESS') actual.add(id);
      else assert.equal(primary.currentState, 'SUSPENDED_STATE');
      assert.equal(orthogonal.currentState, 'SUSPENDED_STATE');
      assert.equal(orthogonal.opCost, 'INFINITY');

      const receiptCount = node.store.db
        .prepare(
          `SELECT COUNT(*) AS count FROM closure_receipts
           WHERE obligation_id = ?`,
        )
        .get(primaryObligationId) as { count: number };
      assert.equal(Number(receiptCount.count), actual.has(id) ? 1 : 0);
      assert.equal(
        (
          node.store.db.prepare('PRAGMA integrity_check').get() as {
            integrity_check: string;
          }
        ).integrity_check,
        'ok',
      );
    }
    assert.deepEqual([...actual].sort(), [...expected].sort());

    const terminalNode = nodes.get([...expected][0])!;
    const terminalReceipt = terminalNode.store.db
      .prepare(
        `SELECT receipt_id FROM closure_receipts WHERE obligation_id = ?`,
      )
      .get(primaryObligationId) as { receipt_id: string };
    assert.throws(
      () =>
        terminalNode.store.db
          .prepare(`DELETE FROM closure_receipts WHERE receipt_id = ?`)
          .run(terminalReceipt.receipt_id),
      /CLOSURE_RECEIPT_IMMUTABLE/,
    );

    assert.ok(transport.counters.simulatedLosses > 0);
    assert.ok(transport.counters.duplicates > 0);
    assert.ok(transport.counters.delayedMessages > 0);
    if (mode !== 'ALTERNATE_PATH') {
      assert.ok(transport.counters.partitionHolds > 0);
    }
    return {
      mode,
      expectedCount: expected.size,
      actualCount: actual.size,
      affectedNodes: [...actual].sort(),
      ticks: transport.ticks,
      counters: transport.counters,
    };
  } finally {
    for (const node of nodes.values()) node.close();
    fs.rmSync(directory, { recursive: true, force: true });
  }
}

test(
  '50 independent stores preserve exact bounded locality under deterministic faults',
  { timeout: 60_000 },
  () => {
    const results = (
      ['FIXED_PATH', 'SCOPED_GOSSIP', 'ALTERNATE_PATH'] as RoutingMode[]
    ).map(runRoutingScenario);
    assert.equal(results.length, 3);
    for (const result of results) {
      assert.equal(result.actualCount, result.expectedCount);
      assert.ok(result.ticks < MAX_LOGICAL_TICKS);
    }
    const fixed = results.find((result) => result.mode === 'FIXED_PATH')!;
    const gossip = results.find((result) => result.mode === 'SCOPED_GOSSIP')!;
    const alternate = results.find(
      (result) => result.mode === 'ALTERNATE_PATH',
    )!;
    assert.ok(gossip.actualCount > fixed.actualCount);
    assert.notDeepEqual(alternate.affectedNodes, fixed.affectedNodes);
  },
);

test('replicated registry revocation blocks a delayed closure at each node', () => {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'ark-revocation-'));
  const stores = [0, 1].map(
    (index) =>
      new ArkBoundaryStore(path.join(directory, `Node_${index}.sqlite`), secrets),
  );
  try {
    const graphHash = sha256Canonical({ graph: 'revocation' });
    const binding = {
      sourceNodeId: SOURCE_NODE,
      targetClass: 'SWARM_NODE',
      mediator: 'GOSSIP_PROTOCOL',
      graphHash,
    };
    const closureLeases: AuthenticatedLease[] = [];
    const closureInputs: ClosureInput[] = [];
    for (const store of stores) {
      const obligation = store.openMatrixObligation(binding);
      const closure: ClosureInput = {
        obligationId: obligation.obligationId,
        resolvedByNode: 'Delayed_Resolver',
        witnessDigest: sha256Canonical({ witness: 'delayed' }),
        graphHash,
        propagationHopLimit: 1,
        resolutionState: 'RESOLVED_BY_WITNESS',
        finiteCost: 1,
      };
      closureInputs.push(closure);
      closureLeases.push(
        new LeaseRegistry(secrets, store.db).issue({
          type: 'APPLY_WINDOW',
          targetResource: '/api/matrix/obligations/close',
          payloadDigest: sha256Canonical(closure),
          scope: 'matrix:closure:apply',
          notBefore: new Date(Date.now() - 1_000).toISOString(),
          expiresAt: new Date(Date.now() + 60_000).toISOString(),
          topologyHash: store.topologyHash(),
        }),
      );
    }

    const update = stores[0].createRegistryUpdate({
      keyId: secrets.activeKeyId,
      purpose: 'RECEIPT_HMAC',
      keyVersion: 1,
      revokedAt: new Date().toISOString(),
      reason: 'replicated revocation assay',
    });
    stores.forEach((store, index) => {
      const leases = new LeaseRegistry(secrets, store.db);
      const updateLease = leases.issue({
        type: 'APPLY_WINDOW',
        targetResource: '/api/trust/keys/revoke',
        payloadDigest: sha256Canonical({ update }),
        scope: 'trust:key:revoke',
        notBefore: new Date(Date.now() - 1_000).toISOString(),
        expiresAt: new Date(Date.now() + 60_000).toISOString(),
        topologyHash: store.topologyHash(),
      });
      store.executeOnce(updateLease, 'REGISTRY_KEY_REVOKE', () =>
        store.applyRegistryUpdate(update),
      );
      assert.throws(
        () =>
          store.executeOnce(
            closureLeases[index],
            'DELAYED_CLOSURE',
            () => store.applyClosureReceipt(closureInputs[index]),
          ),
        /RECEIPT_KEY_REVOKED/,
      );
      assert.equal(
        store.getMatrixObligation(closureInputs[index].obligationId)
          .currentState,
        'SUSPENDED_STATE',
      );
      const nonceCount = store.db
        .prepare(
          `SELECT COUNT(*) AS count FROM consumed_nonces WHERE nonce = ?`,
        )
        .get(closureLeases[index].nonce) as { count: number };
      assert.equal(Number(nonceCount.count), 0);
    });
  } finally {
    stores.forEach((store) => store.db.close());
    fs.rmSync(directory, { recursive: true, force: true });
  }
});

test(
  'abrupt worker termination rolls back an open WAL transaction',
  { timeout: 30_000 },
  async () => {
    const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'ark-crash-'));
    const databasePath = path.join(directory, 'node.sqlite');
    const bootstrap = new ArkBoundaryStore(databasePath, secrets);
    bootstrap.db.close();
    const entryHash = sha256Canonical({ crash: 'uncommitted' });
    const worker = new Worker(
      new URL('./support/crash-worker.ts', import.meta.url),
      { workerData: { databasePath, secrets, entryHash } },
    );
    try {
      await new Promise<void>((resolve, reject) => {
        worker.on('message', (message: { type?: string }) => {
          if (message.type === 'UNCOMMITTED_WRITE_READY') resolve();
        });
        worker.once('error', reject);
      });
      await worker.terminate();

      const reopened = new ArkBoundaryStore(databasePath, secrets);
      try {
        const ledgerCount = reopened.db
          .prepare(
            `SELECT COUNT(*) AS count FROM shadow_ledger WHERE entry_hash = ?`,
          )
          .get(entryHash) as { count: number };
        const claimedCount = reopened.db
          .prepare(
            `SELECT COUNT(*) AS count FROM consumed_nonces
             WHERE nonce = ? OR status = 'CLAIMED'`,
          )
          .get('f'.repeat(32)) as { count: number };
        assert.equal(Number(ledgerCount.count), 0);
        assert.equal(Number(claimedCount.count), 0);
        assert.equal(
          (
            reopened.db.prepare('PRAGMA integrity_check').get() as {
              integrity_check: string;
            }
          ).integrity_check,
          'ok',
        );
      } finally {
        reopened.db.close();
      }
    } finally {
      await worker.terminate();
      fs.rmSync(directory, { recursive: true, force: true });
    }
  },
);
