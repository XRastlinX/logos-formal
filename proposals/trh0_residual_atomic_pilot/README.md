# TRH-0 Residual Atomic Pilot

```yaml
proposal: TRH-0_RESIDUAL_ATOMIC_PILOT
status: PROPOSED
authority_effect: NONE
canonical_effect: NONE
repository_lane: non-canonical proposals/
implemented_stage: SVD_RIDGE_CONTROLS_AND_ZERO_SCANNER
```

This directory contains a frozen, non-canonical numerical pilot. It does not
modify the canonical archive, establish Monadic Spectral Theory, prove the
Transverse Residual Hypothesis, or prove the Riemann Hypothesis.

## Artifacts

- `MST_Glossary.md` — reconciled proposal vocabulary and claim boundary.
- `mst_axiomatic_charter_v1.json` — proposed axioms I–V with explicit governance.
- `trh0_reference.py` — 80-digit xi target, theta/Mellin cross-check, atom support,
  weighting, and design matrices.
- `trh_0_evaluator.py` — SVD ridge path, validation selection, matched controls,
  argument-principle zero isolation, stability filtering, and receipts.
- `test_trh0_reference.py` and `test_trh_0_evaluator.py` — deterministic invariant
  and synthetic zero-isolation tests.
- `receipts/` — execution records; receipts are observations, not promotion records.

## Frozen model

The primary support consists of the first 24 primes congruent to `1 mod 6`,
embedded by

```text
x_k = (a_k - 1) / (2(a_k + 1))
```

The fitted coupled object is

```text
xi_hat(s) = 2*c0 + sum c_k * (x_k^(-s) + x_k^(s-1))
```

with real coefficients. The raw residual transform used for zero inspection is

```text
F_24(s) = c0 + sum c_k * x_k^(-s)
```

All support coordinates are real and contained in `(0, 1/2)`. Complex spectral
behavior comes from the exponent `s`, not from pretending that the atoms are
complex.

## Environment

```text
Python >= 3.11
numpy >= 2.0
mpmath >= 1.3
```

## Verify

```bash
python -m unittest -v test_trh0_reference.py test_trh_0_evaluator.py
python trh0_reference.py verify --dps 80
```

## Run the reconstruction assay

```bash
python trh_0_evaluator.py fit \
  --dps 80 \
  --random-controls 100 \
  --seed 20260729 \
  --output receipts/reconstruction_assay_20260729.json
```

The fitter uses an augmented SVD, algebraically equivalent to

```text
min ||B*c-y||^2 + lambda*||c[1:]||^2
```

on the column-normalized real system. The intercept remains unpenalized. The
selected lambda is the largest value within 1% of the minimum validation loss.

## Run residual-zero stability

One support family:

```bash
python trh_0_evaluator.py zeros \
  --family primary \
  --dps 80 \
  --output receipts/primary_zero_stability_20260729.json
```

Full matched-control zero assay:

```bash
python trh_0_evaluator.py zero-assay \
  --dps 80 \
  --random-controls 100 \
  --seed 20260729 \
  --output receipts/zero_control_assay_20260729.json
```

The zero scanner counts zeros by contour winding, recursively subdivides the
search rectangle, refines isolated roots by Newton iteration, and verifies the
residual at 80 digits. A contour that cannot be resolved is recorded as
`NUMERICALLY_UNRESOLVED`; it is not silently treated as zero evidence.

## Current observed results

The frozen reconstruction run produced:

```text
primary selected lambda:                    1.0
primary unregularized effective rank:       8 of 25
primary test height-balanced MSE:           110444.5829758669
best deterministic control MSE:             199685.09334920358
observed primary/control ratio:              0.553093779427614
random-control fifth percentile MSE:         162258.5162670578
reconstruction relative threshold:          PASSED
```

This is a relative approximation result under the pre-registered metric. It is
not evidence for TRH by itself, particularly because the design remains
numerically degenerate.

The full zero-control run produced:

```text
primary stable off-center zeros:             0
C1 stable off-center zeros:                  0
C2 zero assay:                               NUMERICALLY_UNRESOLVED
resolved random controls:                    87 of 100
unresolved random controls:                  13 of 100
resolved controls with stable off-center:    0
zero-control threshold:                      NOT_EVALUABLE
```

The current zero result therefore does not support the transverse-zero claim.
The next valid step is to resolve the contour ambiguities without changing the
support families, fit objective, validation rule, or evidence thresholds.

## Claim boundary

```text
functional symmetry alone:        NOT EVIDENCE
training error alone:              NOT EVIDENCE
relative reconstruction pass:      PILOT OBSERVATION ONLY
stable transverse-zero evidence:   NOT ESTABLISHED
TRH:                               NOT ESTABLISHED
Riemann Hypothesis:                NOT ESTABLISHED
authority effect:                  NONE
```
