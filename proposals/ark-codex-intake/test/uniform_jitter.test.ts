import assert from 'node:assert/strict';
import test from 'node:test';
import {
  computeUniformJitterMs,
  cyrb53,
  mulberry32,
} from '../src/jitter';

test('uniform jitter replays for an exact node, epoch, and domain', () => {
  const first = computeUniformJitterMs('NODE_7', 42, 500, 'fixture-v1');
  const second = computeUniformJitterMs('NODE_7', 42, 500, 'fixture-v1');
  assert.equal(first, second);
  assert.ok(first >= 0 && first < 1_000);
});

test('uniform jitter separates declared scheduling identities', () => {
  const baseline = computeUniformJitterMs('NODE_7', 42, 500, 'fixture-v1');
  assert.notEqual(
    baseline,
    computeUniformJitterMs('NODE_8', 42, 500, 'fixture-v1'),
  );
  assert.notEqual(
    baseline,
    computeUniformJitterMs('NODE_7', 43, 500, 'fixture-v1'),
  );
  assert.notEqual(
    baseline,
    computeUniformJitterMs('NODE_7', 42, 500, 'fixture-v2'),
  );
});

test('uniform jitter rejects malformed scheduling inputs', () => {
  assert.throws(() => computeUniformJitterMs('', 42, 500), /JITTER_IDENTITY_REQUIRED/);
  assert.throws(() => computeUniformJitterMs('NODE', 1.5, 500), /JITTER_EPOCH_INVALID/);
  assert.throws(() => computeUniformJitterMs('NODE', 42, -1), /JITTER_MEAN_INVALID/);
  assert.throws(
    () => computeUniformJitterMs('NODE', 42, 500, ''),
    /JITTER_IDENTITY_REQUIRED/,
  );
});

test('cyrb53 and mulberry32 remain deterministic primitives', () => {
  assert.equal(cyrb53('fixture'), cyrb53('fixture'));
  const first = mulberry32(cyrb53('fixture') >>> 0);
  const second = mulberry32(cyrb53('fixture') >>> 0);
  assert.deepEqual(
    [first(), first(), first()],
    [second(), second(), second()],
  );
});
