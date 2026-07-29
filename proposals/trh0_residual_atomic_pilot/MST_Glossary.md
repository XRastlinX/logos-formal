# Monadic Spectral Theory — Unified Glossary

```yaml
status: PROPOSED
repository_lane: non-canonical proposals/
canonical_effect: NONE
authority_effect: NONE
```

## Framework

- **Monadic Spectral Theory (MST)**: The full analytic–geometric framework spanning monadic geometry, ontological primes, monadic measures, transforms, couplings, pilots, assays, and TRH.

## Geometric Primitives

- **Origin Point**: The dimensionless locus zero at the center of the Monadic Disk. Extension 0, locus count 1.
- **Monadic Disk**: The closed interval `[-1/2, 1/2]` (real) or its complexified disc.
- **Axial Diameter**: The real segment `(-1/2, 1/2)`; the domain of real-atom support.
- **Monadic Boundary**: The endpoints `±1/2` or the circle `|z|=1/2`; the outer polarity.
- **Monadic Polarity**: The irreducible two-valued structure: Origin Point ↔ Monadic Boundary. The “two-valuedness of 1.”

## Number System and Primes

- **Ur-Prime**: The prime 1, carrying Monadic Polarity internally (values 0 and 1).
- **Ontological Primes**: The prime sequence starting with the Ur-Prime: `1, 2, 3, 5, …`.
- **Star-Primes**: Primes congruent to `1 mod 6` (`7, 13, 19, …`). Used as atom indices in the Axial Pilot.

## Measure and Transform

- **Monadic Measure** (`μ`): A measure supported on the Monadic Disk.
- **Anchor Constant** (`c₀`): Coefficient of the constant basis function (character 1).
- **Mellin–Monad Kernel**: `K(s, x) = |x|^{-s}`, used for non-zero atoms.
- **Monadic Transform** (`F_μ`): The integral or sum of the kernel against `μ`.
- **Monadic Reflection** (`ξ`-Coupling): `A_ξ F(s) = F(s) + F(1-s)`. Enforces `s ↔ 1-s` symmetry.
- **Theta Anchor**: The exact theta/Mellin representation of `ξ(s)`; the classical truth target.

## Trials and Branches

- **Axial Pilot**: The 24-atom real trial using Star-Primes and the `ξ`-Coupling.
- **Explicit Formula Branch**: The von Mangoldt/Jacobi alternative model.
- **Theta Primary Assay**: The main classical-compatible fitting trial against the Theta Anchor.
- **Null Families**: Matched control atom families: opposite residue, Chebyshev, and stratified random.

## TRH and Spectral Structure

- **Transverse Residual Hypothesis (TRH)**: For a suitable Monadic Measure, zeros of `F_μ` include modes off the Axial Diameter, and a constrained coupling maps them to the critical line `Re(s)=1/2`.
- **Spectral Transverse Zeros**: Zeros of `F_μ(s)` with `Re(s) ≠ 1/2`.
- **Transverse Support**: Atom positions `z_k = x_k e^{iθ_k}` with `θ_k ≠ 0`; required for TRH discrimination.
- **TRH-Gated**: Status indicating transverse support is not yet enabled.

## Validation and Status Codes

- **Axial Reconstruction Score**: Metric comparing approximation quality against the Null Families.
- **Pilot Statuses**: `PILOT_DECLARED`, `PILOT_READY`, `TRH-GATED`, `NONCANONICAL`, `NO_AUTHORITY_EFFECT`.

## Claim Boundary

The glossary records the vocabulary of a non-canonical research proposal. Its ontological definitions do not replace conventional arithmetic or establish the Riemann Hypothesis. Empirical status is determined only by the frozen TRH-0 controls, held-out evaluation, and residual-zero assay.
