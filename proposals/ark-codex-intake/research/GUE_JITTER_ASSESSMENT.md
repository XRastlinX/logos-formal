# Proposed GUE-Inspired Cadence Jitter Experiment

```text
artifact:              gue-jitter-assessment
classification:        PROPOSED_ENGINEERING_EXPERIMENT
lifecycle:              PROPOSED
runtime_integration:    NOT_PERFORMED
evidence:               NOT_ESTABLISHED
authority_effect:       NONE
```

## Question

Can a deterministic schedule inspired by Gaussian Unitary Ensemble (GUE)
nearest-neighbor spacings reduce simultaneous cadence strikes and SQLite
write-lock contention compared with bounded uniform and exponential baselines?

## Required correction

The Wigner-surmise density

```text
p(s) = (32 / pi^2) * s^2 * exp(-(4 / pi) * s^2),  s >= 0
```

has level repulsion near `s = 0` and unit mean under the usual normalization.
That property describes a spacing distribution. It does **not** follow that
independently drawing one absolute offset per node produces repulsion between
the resulting node schedules. Two independent offsets can still be arbitrarily
close.

Consequently, these mechanisms must be tested separately:

1. `INDEPENDENT_GUE_OFFSETS`: each node independently draws an absolute offset;
2. `GUE_RENEWAL_SPACINGS`: ordered strike times are constructed from sampled
   spacings and then scaled into a bounded epoch;
3. `COORDINATED_REPULSIVE_SCHEDULE`: a scheduler constructs a jointly separated
   set of offsets; and
4. `KEYED_DETERMINISTIC_SCHEDULE`: the schedule is derived from a deployment
   secret, node identity, and epoch.

Only the second and third mechanisms encode separation directly. The third
introduces coordination and therefore has a different trust and availability
model.

## Independent bounded default

For independent offsets with a probability density `f` on a fixed interval of
length `L`, the close-pair approximation is governed by:

```text
integral(f(x)^2 dx)
```

Since `integral(f(x) dx) = 1`, Cauchy-Schwarz gives:

```text
integral(f(x)^2 dx) >= 1 / L
```

with equality for the uniform density. Uniform therefore minimizes this
close-pair density criterion among independent densities on the same fixed
window. This is a mathematical statement about the declared model, not proof
that uniform minimizes real SQLite `BUSY` events for every workload.

The present engineering default is consequently:

```text
seed = HASH(scheduleDomain, nodeId, epoch, configurationVersion)
offset = Uniform(0, 2 * meanJitterMs; seed)
```

This provides replay without a centralized phase-assignment round. If node
identities are attacker-controlled, a deployment-keyed derivation is required
to prevent seed selection attacks.

## Baselines

- bounded uniform absolute offsets;
- exponential inter-arrival spacings (the precise meaning of “Poisson jitter”);
- deterministic hash-derived bounded offsets; and
- the four candidate mechanisms above.

“Poisson jitter” must not be used without stating whether the model is a
Poisson counting process, exponential inter-arrival time, or independently
sampled absolute offsets.

## Invariants

- Jitter changes wake-up scheduling only; it never changes obligation TTL,
  closure eligibility, graph binding, receipt validation, or authority.
- A schedule outside its configured window is rejected.
- The same configuration, secret version, node set, and epoch reproduces the
  same schedule.
- Node identity alone must not let an untrusted participant select another
  node’s offset.
- Sampling failure is observable and uses a declared fallback.
- Administrative clearance cost `0` remains legal and trace-preserving under
  the existing cost algebra. This experiment must not silently replace it with
  a `c > 0` rule.

## Measurements

- pairwise minimum and quantile inter-strike spacing;
- collision count within an explicitly declared collision window;
- SQLite busy/locked count, total lock-wait time, and retry count;
- completed obligations per logical interval;
- deadline misses;
- sampler attempts and fallback count; and
- audit-trace and state-transition conformance.

## Experimental matrix

Run at node counts `50`, `200`, and `1000`, with declared obligation load,
cadence period, jitter window, loss model, latency model, partition schedule,
SQLite configuration, seed material, and collision threshold. Each comparison
uses the same workload and fault trace. Report confidence intervals across
multiple independent epochs.

## Falsifiers

- Independent GUE offsets do not materially reduce collisions relative to
  uniform offsets.
- Any observed reduction disappears when workload and schedule windows are
  controlled.
- Lock contention moves rather than decreases.
- Deadline misses or tail latency increase beyond the declared budget.
- Results depend on an unreported seed or scheduler coordination.
- Sampling or clipping destroys the claimed spacing distribution.
- Audit or protected-state invariants differ between scheduling strategies.

## Claim boundary

The proposed “greater than 90% contention reduction” is a benchmark target, not
evidence or a pass condition established in advance. A result is reportable
only with the exact workload, baseline, number of trials, uncertainty, and raw
measurements.

The ready-to-run simulator is `research/gue-jitter-simulator.ts`; its
falsification tests are in `test/gue_jitter_math.test.ts`.

Run the default comparison with:

```text
npm run research:jitter -- 1000 1785263704 500 2
```

The arguments are node count, epoch, mean jitter in milliseconds, and collision
window in milliseconds. The output is JSON.

The fixed default fixture is recorded in
`research/gue-jitter-fixture-result.json`. It observed:

| Strategy | Adjacent collisions within 2 ms | Minimum spacing |
| --- | ---: | ---: |
| uniform absolute offsets | 858 | 0.000513 ms |
| exponential absolute offsets | 744 | 0.000312 ms |
| independent GUE offsets | 880 | 0.000334 ms |
| coordinated GUE spacings | 980 | 0.114743 ms |

For this fixture, independent GUE offsets increased rather than reduced the
adjacent-collision count relative to both baselines. The greater-than-90%
reduction claim is therefore `FALSIFIED_FOR_THIS_FIXTURE`.

The coordinated schedule improved the minimum spacing but packed 1,000 strikes
into an approximately 1,000 ms window, whose average gap is about 1 ms.
Consequently, most adjacent gaps still fell inside a 2 ms collision window.
Repulsive spacing does not override the density imposed by node count, window
size, and collision threshold.

No GUE-inspired scheduler is integrated into `CadenceEngine` in this proposal.
The current deterministic cadence and its tested semantics remain unchanged.

## Fact / hype disposition

```text
independent Wigner marginal creates level repulsion: FALSE
seeded bounded uniform as independent default:       SUPPORTED_BY_MODEL
uniform optimal for all real WAL workloads:           NOT_ESTABLISHED
greater-than-90-percent WAL reduction:                NOT_ESTABLISHED
WAL contention measurement:                           NOT_PERFORMED
zeta / nuclear / physical correspondence:             NONE
Frak as scheduling mechanism:                         NOT_PART_OF_ENGINEERING_CLAIM
authority_effect:                                     NONE
```

The checked-in sampler implements the Wigner-surmise marginal directly. It is
not classified as an Erlang sampler. Results from any earlier two-sample or
Erlang experiment must be recorded as a separate fixture.
