import crypto from 'node:crypto';
import { pathToFileURL } from 'node:url';
import { mulberry32 } from '../src/jitter';

export type JitterStrategy =
  | 'UNIFORM_ABSOLUTE'
  | 'EXPONENTIAL_ABSOLUTE'
  | 'INDEPENDENT_GUE_OFFSETS'
  | 'COORDINATED_GUE_SPACINGS';

export interface SimulationConfig {
  nodeCount: number;
  epoch: number;
  meanJitterMs: number;
  collisionWindowMs: number;
}

export interface StrikeMetrics {
  strategy: JitterStrategy;
  nodeCount: number;
  epoch: number;
  meanJitterMs: number;
  collisionWindowMs: number;
  meanOffsetMs: number;
  spanMs: number;
  minimumSpacingMs: number;
  adjacentCollisionCount: number;
  pairCollisionCount: number;
  samplerAttempts: number;
  samplerFallbacks: number;
}

interface Sample {
  value: number;
  attempts: number;
  fallback: boolean;
}

function seed32(...parts: Array<string | number>): number {
  const digest = crypto
    .createHash('sha256')
    .update(parts.join('\u001f'))
    .digest();
  return digest.readUInt32LE(0);
}

export function sampleGueSpacing(
  initialSeed: number,
  maxAttempts = 1_000,
): Sample {
  const random = mulberry32(initialSeed);
  const sigma = 1;
  const envelope = 3;
  const targetConstant = 32 / (Math.PI * Math.PI);

  for (let attempts = 1; attempts <= maxAttempts; attempts += 1) {
    const uniform = Math.max(Number.EPSILON, random());
    const spacing = sigma * Math.sqrt(-2 * Math.log(uniform));
    const target =
      targetConstant *
      spacing *
      spacing *
      Math.exp(-(4 / Math.PI) * spacing * spacing);
    const proposal =
      (spacing / (sigma * sigma)) *
      Math.exp(-(spacing * spacing) / (2 * sigma * sigma));
    const acceptance = proposal === 0 ? 0 : target / (envelope * proposal);
    if (random() < acceptance) {
      return { value: spacing, attempts, fallback: false };
    }
  }

  return { value: 0.5, attempts: maxAttempts, fallback: true };
}

function independentOffsets(
  config: SimulationConfig,
  strategy: Exclude<JitterStrategy, 'COORDINATED_GUE_SPACINGS'>,
): { strikes: number[]; attempts: number; fallbacks: number } {
  const strikes: number[] = [];
  let attempts = 0;
  let fallbacks = 0;

  for (let index = 0; index < config.nodeCount; index += 1) {
    const nodeId = `NODE_${index}`;
    const random = mulberry32(seed32(strategy, nodeId, config.epoch));
    if (strategy === 'UNIFORM_ABSOLUTE') {
      strikes.push(random() * config.meanJitterMs * 2);
      continue;
    }
    if (strategy === 'EXPONENTIAL_ABSOLUTE') {
      strikes.push(
        -Math.log(Math.max(Number.EPSILON, 1 - random())) *
          config.meanJitterMs,
      );
      continue;
    }

    const sample = sampleGueSpacing(
      seed32(strategy, nodeId, config.epoch, 'spacing'),
    );
    attempts += sample.attempts;
    fallbacks += Number(sample.fallback);
    strikes.push(sample.value * config.meanJitterMs);
  }

  strikes.sort((left, right) => left - right);
  return { strikes, attempts, fallbacks };
}

function coordinatedGueOffsets(config: SimulationConfig): {
  strikes: number[];
  attempts: number;
  fallbacks: number;
} {
  const spacings: number[] = [];
  let attempts = 0;
  let fallbacks = 0;

  for (let index = 0; index < config.nodeCount + 1; index += 1) {
    const sample = sampleGueSpacing(
      seed32('COORDINATED_GUE_SPACINGS', config.epoch, index),
    );
    attempts += sample.attempts;
    fallbacks += Number(sample.fallback);
    spacings.push(sample.value);
  }

  const total = spacings.reduce((sum, spacing) => sum + spacing, 0);
  const windowMs = config.meanJitterMs * 2;
  const strikes: number[] = [];
  let position = 0;
  for (let index = 0; index < config.nodeCount; index += 1) {
    position += spacings[index];
    strikes.push((position / total) * windowMs);
  }
  return { strikes, attempts, fallbacks };
}

export function measureStrikes(
  strikes: number[],
  collisionWindowMs: number,
): Omit<
  StrikeMetrics,
  | 'strategy'
  | 'nodeCount'
  | 'epoch'
  | 'meanJitterMs'
  | 'collisionWindowMs'
  | 'samplerAttempts'
  | 'samplerFallbacks'
> {
  if (strikes.length === 0) {
    return {
      meanOffsetMs: 0,
      spanMs: 0,
      minimumSpacingMs: Number.POSITIVE_INFINITY,
      adjacentCollisionCount: 0,
      pairCollisionCount: 0,
    };
  }

  let adjacentCollisionCount = 0;
  let pairCollisionCount = 0;
  let minimumSpacingMs = Number.POSITIVE_INFINITY;
  let left = 0;

  for (let right = 0; right < strikes.length; right += 1) {
    if (right > 0) {
      const spacing = strikes[right] - strikes[right - 1];
      minimumSpacingMs = Math.min(minimumSpacingMs, spacing);
      adjacentCollisionCount += Number(spacing <= collisionWindowMs);
    }
    while (strikes[right] - strikes[left] > collisionWindowMs) {
      left += 1;
    }
    pairCollisionCount += right - left;
  }

  return {
    meanOffsetMs:
      strikes.reduce((sum, strike) => sum + strike, 0) / strikes.length,
    spanMs: strikes[strikes.length - 1] - strikes[0],
    minimumSpacingMs,
    adjacentCollisionCount,
    pairCollisionCount,
  };
}

export function simulateStrategy(
  config: SimulationConfig,
  strategy: JitterStrategy,
): StrikeMetrics {
  const generated =
    strategy === 'COORDINATED_GUE_SPACINGS'
      ? coordinatedGueOffsets(config)
      : independentOffsets(config, strategy);
  return {
    strategy,
    nodeCount: config.nodeCount,
    epoch: config.epoch,
    meanJitterMs: config.meanJitterMs,
    collisionWindowMs: config.collisionWindowMs,
    ...measureStrikes(generated.strikes, config.collisionWindowMs),
    samplerAttempts: generated.attempts,
    samplerFallbacks: generated.fallbacks,
  };
}

export function runComparison(config: SimulationConfig): StrikeMetrics[] {
  const strategies: JitterStrategy[] = [
    'UNIFORM_ABSOLUTE',
    'EXPONENTIAL_ABSOLUTE',
    'INDEPENDENT_GUE_OFFSETS',
    'COORDINATED_GUE_SPACINGS',
  ];
  return strategies.map((strategy) => simulateStrategy(config, strategy));
}

const isMain =
  process.argv[1] !== undefined &&
  import.meta.url === pathToFileURL(process.argv[1]).href;

if (isMain) {
  const config: SimulationConfig = {
    nodeCount: Number(process.argv[2] ?? 1_000),
    epoch: Number(process.argv[3] ?? 0),
    meanJitterMs: Number(process.argv[4] ?? 500),
    collisionWindowMs: Number(process.argv[5] ?? 2),
  };
  console.log(JSON.stringify({ config, results: runComparison(config) }, null, 2));
}
