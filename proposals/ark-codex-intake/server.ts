import express from 'express';
import { GoogleGenAI } from '@google/genai';
import dotenv from 'dotenv';
import fs from 'node:fs';
import path from 'node:path';
import {
  ArkBoundaryStore,
  AuthenticatedLease,
  BoundaryError,
  LeaseRegistry,
  LeaseType,
  RegistryUpdate_v01,
  authenticateOperator,
  defaultArkDatabasePath,
  loadBoundarySecrets,
  sha256Canonical,
} from './src/boundary';

dotenv.config();

interface Frame {
  nodeId: string;
  parent: string;
  call: string;
  outputs: string[];
}

interface FederationEnvelope {
  entryHash: string;
  payload: unknown;
  source: string;
}

const ARK_PLACE = process.env.ARK_PLACE || 'BEACON';
const secrets = loadBoundarySecrets();
const store = new ArkBoundaryStore(defaultArkDatabasePath(), secrets);
const leases = new LeaseRegistry(secrets, store.db);
const gammaStack: Frame[] = [];
const app = express();

app.use(express.json({ limit: '10mb', strict: true }));

const vocabularyPath = path.join(process.cwd(), 'limited_vocabulary_model.json');
const vocabularyModel = JSON.parse(
  fs.readFileSync(vocabularyPath, 'utf8'),
).limited_vocabulary_model;

function getGeminiClient(): GoogleGenAI | null {
  const apiKey = process.env.GEMINI_API_KEY?.trim();
  return apiKey ? new GoogleGenAI({ apiKey }) : null;
}

function requireFoundry(
  _req: express.Request,
  res: express.Response,
  next: express.NextFunction,
): void {
  if (ARK_PLACE !== 'FOUNDRY') {
    res.status(403).json({ error: 'FOUNDRY_MODE_REQUIRED' });
    return;
  }
  next();
}

function requireOperator(
  req: express.Request,
  res: express.Response,
  next: express.NextFunction,
): void {
  try {
    authenticateOperator(req.headers['x-sovereign-token'], secrets.authSecret);
    next();
  } catch (error) {
    sendError(res, error);
  }
}

function sendError(res: express.Response, error: unknown): void {
  if (error instanceof BoundaryError) {
    res.status(error.statusCode).json({ error: error.message });
    return;
  }
  const message = error instanceof Error ? error.message : 'INTERNAL_ERROR';
  res.status(500).json({ error: message });
}

function leaseFromBody(body: Record<string, unknown>): unknown {
  return body.lease;
}

function verifyEndpointLease(
  candidate: unknown,
  type: LeaseType,
  targetResource: string,
  scope: string,
  payload: unknown,
): AuthenticatedLease {
  return leases.verify(candidate, {
    type,
    targetResource,
    scope,
    payloadDigest: sha256Canonical(payload),
  });
}

function normalizeSensory(modality: unknown, data: unknown): Record<string, unknown> {
  if (modality !== 'ME-VISION-01' && modality !== 'ME-YOUTUBE-01') {
    throw new BoundaryError(400, 'UNKNOWN_MODALITY');
  }
  if (data === undefined || data === null) {
    throw new BoundaryError(400, 'SENSORY_DATA_REQUIRED');
  }
  const record =
    typeof data === 'object' && !Array.isArray(data)
      ? (data as Record<string, unknown>)
      : { value: data };
  const confidence = Object.keys(record).length > 0 ? 1 : 0;
  return {
    SigmaText:
      typeof record.text === 'string'
        ? record.text
        : typeof record.value === 'string'
          ? record.value
          : '',
    SigmaScene: modality === 'ME-VISION-01' ? record.scene ?? null : null,
    SigmaObjects: Array.isArray(record.objects) ? record.objects : [],
    SigmaIntent: 'ROUTE_SENSORY_DATA',
    SigmaMeta: {
      confidence,
      provenance: modality,
      sourceDigest: sha256Canonical(data),
    },
  };
}

function assertEvaluationOutput(result: unknown): {
  output: string;
  pComplete: boolean;
  pError: boolean;
  pRetry: boolean;
} {
  if (!result || typeof result !== 'object') {
    throw new BoundaryError(502, 'INVALID_MODEL_RESPONSE');
  }
  const record = result as Record<string, unknown>;
  if (
    typeof record.output !== 'string' ||
    typeof record.p_complete !== 'boolean' ||
    typeof record.p_error !== 'boolean' ||
    typeof record.p_retry !== 'boolean'
  ) {
    throw new BoundaryError(502, 'INVALID_MODEL_RESPONSE_SCHEMA');
  }
  const forbiddenTerms: string[] =
    vocabularyModel.grammar.forbidden_surface_terms ?? [];
  const escaped = forbiddenTerms
    .map((term) => term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'))
    .sort((left, right) => right.length - left.length);
  if (escaped.length > 0) {
    const forbiddenRegex = new RegExp(
      `(?:^|\\W)(?:${escaped.join('|')})(?:$|\\W)`,
      'i',
    );
    if (forbiddenRegex.test(record.output)) {
      throw new BoundaryError(400, 'LEXICAL_BOUNDARY_ERROR');
    }
  }
  return {
    output: record.output,
    pComplete: record.p_complete,
    pError: record.p_error,
    pRetry: record.p_retry,
  };
}

function parseFederationEnvelopes(candidate: unknown): FederationEnvelope[] {
  if (!Array.isArray(candidate) || candidate.length === 0 || candidate.length > 1000) {
    throw new BoundaryError(400, 'INVALID_FEDERATION_BATCH');
  }
  return candidate.map((item) => {
    if (!item || typeof item !== 'object' || Array.isArray(item)) {
      throw new BoundaryError(400, 'INVALID_FEDERATION_ENVELOPE');
    }
    const record = item as Record<string, unknown>;
    const keys = Object.keys(record).sort();
    if (
      keys.join(',') !== 'entryHash,payload,source' ||
      typeof record.entryHash !== 'string' ||
      typeof record.source !== 'string' ||
      record.entryHash !== sha256Canonical(record.payload)
    ) {
      throw new BoundaryError(400, 'FEDERATION_INTEGRITY_FAILURE');
    }
    return {
      entryHash: record.entryHash,
      payload: record.payload,
      source: record.source,
    };
  });
}

app.post(
  '/api/foundry/lease/issue',
  requireFoundry,
  requireOperator,
  (req, res) => {
    try {
      const body = req.body as Record<string, unknown>;
      const lease = leases.issue({
        type: body.type as LeaseType,
        targetResource: String(body.targetResource ?? body.target_resource ?? ''),
        payloadDigest: String(body.payloadDigest ?? body.payload_digest ?? ''),
        scope: String(body.scope ?? ''),
        notBefore: String(body.notBefore ?? body.not_before ?? ''),
        expiresAt: String(body.expiresAt ?? body.expires_at ?? ''),
        parentNodeId:
          typeof body.parentNodeId === 'string'
            ? body.parentNodeId
            : typeof body.parent_node_id === 'string'
              ? body.parent_node_id
              : undefined,
        topologyHash: store.topologyHash(),
      });
      res.json({ lease });
    } catch (error) {
      sendError(res, error);
    }
  },
);

app.post(
  '/api/foundry/sensory/route',
  requireFoundry,
  requireOperator,
  (req, res) => {
    try {
      const body = req.body as Record<string, unknown>;
      const payload = { modality: body.modality, data: body.data };
      const lease = verifyEndpointLease(
        leaseFromBody(body),
        'INGEST_WINDOW',
        '/api/foundry/sensory/route',
        'sensory:route',
        payload,
      );
      const result = store.executeOnce(lease, 'SENSORY_ROUTE', () => {
        const Sigma = normalizeSensory(body.modality, body.data);
        return {
          Sigma: {
            ...Sigma,
            SigmaMeta: {
              ...(Sigma.SigmaMeta as Record<string, unknown>),
              leaseId: lease.id,
            },
          },
          routedTo:
            (Sigma.SigmaMeta as { confidence: number }).confidence < 0.7
              ? 'CLARIFY'
              : 'PUSHDOWN_QUEUE',
        };
      });
      res.json({ ...result.response, replayed: result.replayed, receipt: result.receipt });
    } catch (error) {
      sendError(res, error);
    }
  },
);

app.post(
  '/api/foundry/pushdown/evaluate',
  requireFoundry,
  requireOperator,
  async (req, res) => {
    const body = req.body as Record<string, unknown>;
    let lease: AuthenticatedLease | undefined;
    let externalClaimed = false;
    let activeFrame: Frame | undefined;
    try {
      const payload = { Sigma: body.Sigma, pCall: body.p_call };
      lease = verifyEndpointLease(
        leaseFromBody(body),
        'EVALUATE_WINDOW',
        '/api/foundry/pushdown/evaluate',
        'pushdown:evaluate',
        payload,
      );

      const prior = store.beginExternal(lease);
      if (prior.state === 'COMMITTED') {
        res.json({ ...(prior.response as object), replayed: true });
        return;
      }
      externalClaimed = true;

      const parent = gammaStack.at(-1)?.nodeId ?? 'ROOT';
      if (lease.parentNodeId && lease.parentNodeId !== parent) {
        throw new BoundaryError(409, 'PARENT_NODE_MISMATCH');
      }
      if (typeof body.p_call !== 'string' || body.p_call.length === 0) {
        throw new BoundaryError(400, 'CALL_VECTOR_REQUIRED');
      }
      const ai = getGeminiClient();
      if (!ai) {
        throw new BoundaryError(503, 'GEMINI_NOT_CONFIGURED');
      }

      const frame: Frame = {
        nodeId: `v_delta_${lease.nonce.slice(0, 16)}`,
        parent,
        call: body.p_call,
        outputs: [],
      };
      gammaStack.push(frame);
      activeFrame = frame;

      const prompt = [
        'Execute Pushdown Branch.',
        `Substrate Context: ${JSON.stringify(body.Sigma)}`,
        `Call Vector: ${body.p_call}`,
        'Use the declared limited vocabulary constraints.',
        'Return JSON with output, p_complete, p_error, and p_retry.',
      ].join('\n');
      const modelResponse = await ai.models.generateContent({
        model: 'gemini-3.6-flash',
        contents: prompt,
        config: { responseMimeType: 'application/json' },
      });
      let parsed: unknown;
      try {
        parsed = JSON.parse(modelResponse.text || '{}');
      } catch {
        throw new BoundaryError(502, 'MODEL_JSON_PARSE_ERROR');
      }
      const evaluation = assertEvaluationOutput(parsed);
      frame.outputs.push(evaluation.output);
      if (evaluation.pComplete || evaluation.pError) gammaStack.pop();

      const response = {
        status: evaluation.pComplete ? 'EVALUATED_AND_POPPED' : 'STACK_PENDING',
        evaluation,
        frame,
      };
      const completed = store.completeExternal(lease, 'PUSHDOWN_EVALUATE', response);
      res.json({ ...completed.response, replayed: false, receipt: completed.receipt });
    } catch (error) {
      if (activeFrame && gammaStack.at(-1)?.nodeId === activeFrame.nodeId) {
        gammaStack.pop();
      }
      if (lease && externalClaimed) {
        store.failExternal(
          lease,
          error instanceof Error ? error.message : 'EXTERNAL_EVALUATION_FAILED',
        );
      }
      sendError(res, error);
    }
  },
);

app.post(
  '/api/foundry/apply',
  requireFoundry,
  requireOperator,
  (req, res) => {
    try {
      const body = req.body as Record<string, unknown>;
      const payload = {
        frameData: body.frame_data,
        structuralModification: body.structural_modification,
      };
      const lease = verifyEndpointLease(
        leaseFromBody(body),
        'APPLY_WINDOW',
        '/api/foundry/apply',
        'ledger:apply',
        payload,
      );
      const result = store.executeOnce(lease, 'SOVEREIGN_APPLY', () => {
        const modification = body.structural_modification;
        if (
          modification &&
          typeof modification === 'object' &&
          (modification as Record<string, unknown>).type === 'PROMOTE_QUARANTINE'
        ) {
          const hashes = (modification as Record<string, unknown>).entryHashes;
          if (!Array.isArray(hashes) || hashes.some((hash) => typeof hash !== 'string')) {
            throw new BoundaryError(400, 'INVALID_PROMOTION_REQUEST');
          }
          return {
            status: 'QUARANTINE_PROMOTED',
            promoted: store.promoteQuarantine(hashes as string[]),
          };
        }
        if (gammaStack.length > 0 && !modification) {
          throw new BoundaryError(409, 'STACK_PENDING');
        }
        const entry = {
          timestamp: new Date().toISOString(),
          event: modification ? 'STRUCTURAL_MUTATION' : 'TERMINAL_SETTLEMENT',
          frameData: body.frame_data ?? null,
          structuralModification: modification ?? null,
          promotiveAuthority: 'AUTHENTICATED_OPERATOR',
        };
        const entryHash = store.appendLedgerEntry(entry, 'ACTIVE', 'LOCAL_APPLY');
        return { status: 'APPLY_COMMITTED', entryHash, entry };
      });
      res.json({ ...result.response, replayed: result.replayed, receipt: result.receipt });
    } catch (error) {
      sendError(res, error);
    }
  },
);

app.get('/api/matrix/obligations', (_req, res) => {
  res.json({
    status: 'MATRIX_OBLIGATION_LIST',
    obligations: store.listMatrixObligations(),
  });
});

app.post(
  '/api/matrix/obligations/open',
  requireFoundry,
  requireOperator,
  (req, res) => {
    try {
      const body = req.body as Record<string, unknown>;
      const payload = {
        obligationId: body.obligationId,
        sourceNodeId: body.sourceNodeId,
        targetClass: body.targetClass,
        mediator: body.mediator,
        graphHash: body.graphHash,
      };
      const lease = verifyEndpointLease(
        leaseFromBody(body),
        'APPLY_WINDOW',
        '/api/matrix/obligations/open',
        'matrix:obligation:open',
        payload,
      );
      const result = store.executeOnce(lease, 'MATRIX_OBLIGATION_OPEN', () => ({
        status: 'SUSPENDED_STATE',
        obligation: store.openMatrixObligation({
          obligationId: String(body.obligationId ?? ''),
          sourceNodeId: String(body.sourceNodeId ?? ''),
          targetClass: String(body.targetClass ?? ''),
          mediator: String(body.mediator ?? ''),
          graphHash: String(body.graphHash ?? ''),
        }),
      }));
      res.json({
        ...result.response,
        replayed: result.replayed,
        receipt: result.receipt,
      });
    } catch (error) {
      sendError(res, error);
    }
  },
);

app.post(
  '/api/matrix/obligations/close',
  requireFoundry,
  requireOperator,
  (req, res) => {
    try {
      const body = req.body as Record<string, unknown>;
      const payload = {
        obligationId: body.obligationId,
        resolvedByNode: body.resolvedByNode,
        witnessDigest: body.witnessDigest,
        propagationHopLimit: body.propagationHopLimit,
        resolutionState: body.resolutionState,
        finiteCost: body.finiteCost,
        graphHash: body.graphHash,
      };
      const lease = verifyEndpointLease(
        leaseFromBody(body),
        'APPLY_WINDOW',
        '/api/matrix/obligations/close',
        'matrix:closure:apply',
        payload,
      );
      const result = store.executeOnce(lease, 'MATRIX_CLOSURE_APPLY', () => {
        const closure = store.applyClosureReceipt({
          obligationId: String(body.obligationId ?? ''),
          resolvedByNode: String(body.resolvedByNode ?? ''),
          witnessDigest: String(body.witnessDigest ?? ''),
          propagationHopLimit: Number(body.propagationHopLimit),
          resolutionState: body.resolutionState as
            | 'CLEARED_BY_RECEIPT'
            | 'RESOLVED_BY_WITNESS',
          finiteCost: Number(body.finiteCost),
          graphHash: String(body.graphHash ?? ''),
        });
        return {
          status: body.resolutionState,
          closureReceipt: closure.receipt,
          obligation: closure.obligation,
        };
      });
      res.json({
        ...result.response,
        replayed: result.replayed,
        receipt: result.receipt,
      });
    } catch (error) {
      sendError(res, error);
    }
  },
);

app.post(
  '/api/trust/keys/revoke',
  requireFoundry,
  requireOperator,
  (req, res) => {
    try {
      const body = req.body as Record<string, unknown>;
      const payload = { update: body.update };
      const lease = verifyEndpointLease(
        leaseFromBody(body),
        'APPLY_WINDOW',
        '/api/trust/keys/revoke',
        'trust:key:revoke',
        payload,
      );
      const result = store.executeOnce(lease, 'REGISTRY_KEY_REVOKE', () => ({
        status: store.applyRegistryUpdate(
          body.update as RegistryUpdate_v01,
        ).status,
      }));
      res.json({
        ...result.response,
        replayed: result.replayed,
        receipt: result.receipt,
      });
    } catch (error) {
      sendError(res, error);
    }
  },
);

app.get('/api/federation/sync', (_req, res) => {
  res.json({
    status: 'SYNC_EXPORT',
    topologyHash: store.topologyHash(),
    ledger: store.listLedger(false),
  });
});

app.post(
  '/api/federation/sync',
  requireFoundry,
  requireOperator,
  (req, res) => {
    try {
      const body = req.body as Record<string, unknown>;
      const payload = { externalLedger: body.external_ledger };
      const lease = verifyEndpointLease(
        leaseFromBody(body),
        'FEDERATION_WINDOW',
        '/api/federation/sync',
        'federation:import',
        payload,
      );
      const envelopes = parseFederationEnvelopes(body.external_ledger);
      const result = store.executeOnce(lease, 'FEDERATION_IMPORT', () => {
        const quarantined = envelopes.map((envelope) =>
          store.appendLedgerEntry(
            {
              externalEntryHash: envelope.entryHash,
              payload: envelope.payload,
              source: envelope.source,
            },
            'QUARANTINE',
            envelope.source,
          ),
        );
        return {
          status: 'FEDERATION_QUARANTINED',
          importedCount: quarantined.length,
          quarantineHashes: quarantined,
        };
      });
      res.json({ ...result.response, replayed: result.replayed, receipt: result.receipt });
    } catch (error) {
      sendError(res, error);
    }
  },
);

app.use((_req, res) => {
  res.status(404).json({ error: 'ENDPOINT_NOT_FOUND' });
});

export { app, leases, store };

export async function startServer(): Promise<void> {
  const port = Number(process.env.PORT || 3000);
  app.listen(port, '127.0.0.1', () => {
    console.log(`[Cyphon Runtime] Active place: ${ARK_PLACE}`);
    console.log(`[Cyphon Runtime] Listening on http://127.0.0.1:${port}`);
  });
}

if (
  process.argv[1] &&
  ['server.ts', 'server.cjs', 'server.js'].includes(path.basename(process.argv[1]))
) {
  void startServer();
}
