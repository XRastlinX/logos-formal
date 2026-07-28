import assert from 'node:assert/strict';
import test from 'node:test';
import {
  runComparison,
  sampleGueSpacing,
  simulateStrategy,
  SimulationConfig,
} from '../research/gue-jitter-simulator';

const fixture: SimulationConfig = {
  nodeCount: 1_000,
  epoch: 1_785_263_704,
  meanJitterMs: 500,
  collisionWindowMs: 2,
};

test('GUE sampler is deterministic and has approximately unit mean', () => {
  const sampleCount = 25_000;
  let sum = 0;
  let fallbacks = 0;
  for (let index = 0; index < sampleCount; index += 1) {
    const first = sampleGueSpacing(index);
    const second = sampleGueSpacing(index);
    assert.deepEqual(first, second);
    sum += first.value;
    fallbacks += Number(first.fallback);
  }
  assert.equal(fallbacks, 0);
  assert.ok(Math.abs(sum / sampleCount - 1) < 0.025);
});

test('each strategy replays exactly for the same epoch', () => {
  assert.deepEqual(runComparison(fixture), runComparison(fixture));
  assert.notDeepEqual(
    runComparison(fixture),
    runComparison({ ...fixture, epoch: fixture.epoch + 1 }),
  );
});

test('independent GUE offsets do not establish inter-node level repulsion', () => {
  const gue = simulateStrategy(fixture, 'INDEPENDENT_GUE_OFFSETS');
  const exponential = simulateStrategy(fixture, 'EXPONENTIAL_ABSOLUTE');

  assert.ok(gue.adjacentCollisionCount > 0);
  assert.ok(gue.minimumSpacingMs < fixture.collisionWindowMs);
  assert.ok(
    gue.adjacentCollisionCount >= exponential.adjacentCollisionCount * 0.1,
    'the fixed fixture must falsify the proposed greater-than-90% reduction',
  );
});

test('coordinated spacings are classified separately from independent offsets', () => {
  const independent = simulateStrategy(fixture, 'INDEPENDENT_GUE_OFFSETS');
  const coordinated = simulateStrategy(fixture, 'COORDINATED_GUE_SPACINGS');

  assert.equal(independent.samplerFallbacks, 0);
  assert.equal(coordinated.samplerFallbacks, 0);
  assert.notEqual(
    coordinated.minimumSpacingMs,
    independent.minimumSpacingMs,
  );
  assert.ok(coordinated.spanMs <= fixture.meanJitterMs * 2);
});
