# TRH-0 Reference Xi and Design Matrix

```yaml
proposal: TRH-0_RESIDUAL_ATOMIC_PILOT
status: PROPOSED
authority_effect: NONE
canonical_effect: NONE
repository_lane: non-canonical proposals/
implemented_stage: REFERENCE_XI_AND_DESIGN_MATRIX
```

This directory implements the first executable layer of the frozen TRH-0 pilot.
It does **not** fit coefficients, inspect residual zeros, or assert support for
TRH or the Riemann Hypothesis.

## Implemented

- the exact 24-prime `1 mod 6` atom family;
- the embedding `x=(a-1)/(2(a+1))`;
- an 80-digit direct evaluator of the completed Riemann xi function;
- an independent symmetric theta/Mellin xi evaluator for cross-checking;
- the coupled matrix
  `B_0=2`, `B_k=x_k^(-s)+x_k^(s-1)`;
- the raw residual matrix `R_0=1`, `R_k=x_k^(-s)`;
- height-balanced complex-to-real stacking for real coefficients;
- normalization of penalized columns only;
- effective-rank and condition diagnostics;
- invariant tests for functional symmetry, conjugation, and real-on-line
  behavior.

## Environment

```text
Python >= 3.11
numpy >= 2.0
mpmath >= 1.3
```

## Verify

From this directory:

```bash
python -m unittest -v test_trh0_reference.py
python trh0_reference.py verify --dps 80
```

A successful CLI run emits a JSON receipt with:

```text
status: PROPOSED
authority_effect: NONE
result: OBSERVE_ONLY
```

## Frozen row order

Grid rows are height-major and sigma-minor:

```text
for t in heights:
    for sigma in sigmas:
        s = sigma + i*t
```

The training matrix therefore has `730 x 25` complex entries and becomes a
`1460 x 25` real system after real/imaginary stacking.

## Numerical boundary

Target values are evaluated with `mpmath` at 80 decimal digits and cast once to
`complex128` for the declared float64 fit. The theta/Mellin route is a slower,
independent check and is not used to generate every grid target.
