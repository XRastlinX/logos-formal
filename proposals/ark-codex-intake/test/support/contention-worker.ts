import { parentPort, workerData } from 'node:worker_threads';
import {
  ArkBoundaryStore,
  AuthenticatedLease,
  BoundarySecrets,
} from '../../src/boundary';

interface WorkerInput {
  databasePath: string;
  secrets: BoundarySecrets;
  lease: AuthenticatedLease;
  operation: string;
  response: Record<string, unknown>;
  binding: {
    obligationId: string;
    sourceNodeId: string;
    targetClass: string;
    mediator: string;
    graphHash: string;
  };
}

const input = workerData as WorkerInput;
const secrets: BoundarySecrets = input.secrets;
const store = new ArkBoundaryStore(input.databasePath, secrets);

if (!parentPort) {
  throw new Error('CONTENTION_WORKER_PARENT_PORT_MISSING');
}

parentPort.postMessage({ type: 'READY' });
parentPort.once('message', (message: { type?: string }) => {
  if (message.type !== 'START') return;
  try {
    const result = store.executeOnce(input.lease, input.operation, () => {
      store.openMatrixObligation(input.binding);
      return input.response;
    });
    parentPort?.postMessage({ type: 'RESULT', result });
  } catch (error) {
    parentPort?.postMessage({
      type: 'ERROR',
      error: error instanceof Error ? error.message : String(error),
    });
  } finally {
    store.db.close();
  }
});
