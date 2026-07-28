export function cyrb53(value: string, seed = 0): number {
  let high = 0xdeadbeef ^ seed;
  let low = 0x41c6ce57 ^ seed;
  for (let index = 0; index < value.length; index += 1) {
    const code = value.charCodeAt(index);
    high = Math.imul(high ^ code, 2_654_435_761);
    low = Math.imul(low ^ code, 1_597_334_677);
  }
  high =
    Math.imul(high ^ (high >>> 16), 2_246_822_507) ^
    Math.imul(low ^ (low >>> 13), 3_266_489_909);
  low =
    Math.imul(low ^ (low >>> 16), 2_246_822_507) ^
    Math.imul(high ^ (high >>> 13), 3_266_489_909);
  return 4_294_967_296 * (2_097_151 & low) + (high >>> 0);
}

export function mulberry32(initialSeed: number): () => number {
  let seed = initialSeed >>> 0;
  return () => {
    let value = (seed += 0x6d2b79f5);
    value = Math.imul(value ^ (value >>> 15), value | 1);
    value ^= value + Math.imul(value ^ (value >>> 7), value | 61);
    return ((value ^ (value >>> 14)) >>> 0) / 4_294_967_296;
  };
}

/**
 * Returns a replayable independent offset on [0, 2 * meanJitterMs).
 *
 * This function is a scheduling primitive, not a cryptographic RNG. If node
 * identifiers are attacker-controlled, supply a deployment-keyed schedule
 * domain that untrusted nodes cannot choose.
 */
export function computeUniformJitterMs(
  nodeId: string,
  epoch: number,
  meanJitterMs: number,
  scheduleDomain = 'cadence-uniform-v1',
): number {
  if (nodeId.length === 0 || scheduleDomain.length === 0) {
    throw new TypeError('JITTER_IDENTITY_REQUIRED');
  }
  if (!Number.isSafeInteger(epoch)) {
    throw new TypeError('JITTER_EPOCH_INVALID');
  }
  if (!Number.isFinite(meanJitterMs) || meanJitterMs < 0) {
    throw new TypeError('JITTER_MEAN_INVALID');
  }

  const seed = cyrb53(`${scheduleDomain}\u001f${nodeId}\u001f${epoch}`) >>> 0;
  return mulberry32(seed)() * (2 * meanJitterMs);
}
