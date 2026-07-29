#!/usr/bin/env python3
"""TRH-0 reference xi evaluator and frozen design-matrix layer.

Status: PROPOSED
Authority effect: NONE
Canonical effect: NONE

This module implements only the reference target, atom support, design matrices,
weighting, real-system conversion, column normalization, and diagnostics. It
does not fit coefficients, inspect zeros, or establish a mathematical claim.
"""

from __future__ import annotations

import argparse
import json
import math
from dataclasses import dataclass
from typing import Iterable, Sequence

import mpmath as mp
import numpy as np

MODEL = "TRH-0_RESIDUAL_ATOMIC_PILOT"
STATUS = "PROPOSED"
AUTHORITY_EFFECT = "NONE"
CANONICAL_EFFECT = "NONE"
DEFAULT_DPS = 80
EFFECTIVE_RANK_RELATIVE_CUTOFF = 1.0e-12

PRIMES_1_MOD_6 = (
    7, 13, 19, 31, 37, 43, 61, 67, 73, 79, 97, 103,
    109, 127, 139, 151, 157, 163, 181, 193, 199, 211, 223, 229,
)
SIGMAS = tuple(round(0.05 * i, 2) for i in range(1, 11))
TRAIN_HEIGHTS = tuple(round(0.25 * i, 2) for i in range(73))
VALIDATION_HEIGHTS = tuple(round(19.0 + 0.25 * i, 2) for i in range(33))
TEST_HEIGHTS = tuple(round(28.0 + 0.25 * i, 2) for i in range(49))


class TRH0Error(ValueError):
    """Typed fail-closed error for the proposal implementation."""


@dataclass(frozen=True)
class RealLinearSystem:
    matrix: np.ndarray
    target: np.ndarray
    row_weights: np.ndarray


@dataclass(frozen=True)
class ColumnNormalizedSystem:
    matrix: np.ndarray
    target: np.ndarray
    column_scales: np.ndarray

    def restore_coefficients(self, normalized_coefficients: Sequence[float]) -> np.ndarray:
        beta = np.asarray(normalized_coefficients, dtype=np.float64)
        if beta.shape != self.column_scales.shape:
            raise TRH0Error("coefficient shape does not match column scales")
        return beta / self.column_scales


def _is_prime(value: int) -> bool:
    if value < 2:
        return False
    if value % 2 == 0:
        return value == 2
    for factor in range(3, math.isqrt(value) + 1, 2):
        if value % factor == 0:
            return False
    return True


def validate_atom_family(primes: Sequence[int] = PRIMES_1_MOD_6) -> None:
    if len(primes) != 24:
        raise TRH0Error(f"expected 24 atoms, received {len(primes)}")
    if tuple(sorted(primes)) != tuple(primes) or len(set(primes)) != len(primes):
        raise TRH0Error("atom family must be strictly increasing and unique")
    for atom in primes:
        if not _is_prime(atom) or atom % 6 != 1:
            raise TRH0Error(f"invalid 1 mod 6 prime atom: {atom}")


def embed_atom(atom: int) -> float:
    if atom <= 0:
        raise TRH0Error("atom must be positive")
    value = (atom - 1.0) / (2.0 * (atom + 1.0))
    if not 0.0 < value < 0.5:
        raise TRH0Error("embedded atom escaped the monadic half-cell")
    return value


def frozen_support(primes: Sequence[int] = PRIMES_1_MOD_6) -> np.ndarray:
    validate_atom_family(primes)
    support = np.asarray([embed_atom(atom) for atom in primes], dtype=np.float64)
    support.setflags(write=False)
    return support


def build_grid(sigmas: Sequence[float], heights: Sequence[float]) -> np.ndarray:
    sigma_array = np.asarray(sigmas, dtype=np.float64)
    height_array = np.asarray(heights, dtype=np.float64)
    if sigma_array.ndim != 1 or height_array.ndim != 1:
        raise TRH0Error("sigmas and heights must be one-dimensional")
    if sigma_array.size == 0 or height_array.size == 0:
        raise TRH0Error("grid axes must be non-empty")
    return np.asarray(
        [complex(sigma, height) for height in height_array for sigma in sigma_array],
        dtype=np.complex128,
    )


def _as_mpc(value: complex | float | int | mp.mpc | mp.mpf) -> mp.mpc:
    if isinstance(value, mp.mpc):
        return value
    if isinstance(value, mp.mpf):
        return mp.mpc(value, 0)
    parsed = complex(value)
    return mp.mpc(parsed.real, parsed.imag)


def xi_direct_mp(s, dps: int = DEFAULT_DPS) -> mp.mpc:
    """Completed xi from zeta, gamma, and pi at arbitrary precision."""
    if dps < 30:
        raise TRH0Error("reference evaluation requires at least 30 digits")
    with mp.workdps(dps):
        z = _as_mpc(s)
        if z == 0 or z == 1:
            return mp.mpc(mp.mpf("0.5"), 0)
        return (
            mp.mpf("0.5") * z * (z - 1) * mp.power(mp.pi, -z / 2)
            * mp.gamma(z / 2) * mp.zeta(z)
        )


def _theta_remainder_mp(x: mp.mpf, tolerance: mp.mpf) -> mp.mpf:
    if x < 1:
        raise TRH0Error("theta evaluator requires x >= 1")
    total = mp.mpf("0")
    n = 1
    while True:
        term = mp.exp(-mp.pi * n * n * x)
        total += term
        if term <= tolerance:
            return total
        n += 1
        if n > 100000:
            raise TRH0Error("theta tail failed to converge")


def xi_theta_mp(s, dps: int = DEFAULT_DPS) -> mp.mpc:
    """Independent symmetric theta/Mellin evaluator of completed xi."""
    if dps < 30:
        raise TRH0Error("theta evaluation requires at least 30 digits")
    with mp.workdps(dps):
        z = _as_mpc(s)
        tolerance = mp.power(10, -(dps + 15))
        x_max = mp.mpf(dps + 25) * mp.log(10) / mp.pi
        u_max = mp.log(x_max)

        def integrand(u):
            x = mp.exp(u)
            psi = _theta_remainder_mp(x, tolerance)
            return psi * (mp.exp(z * u / 2) + mp.exp((1 - z) * u / 2))

        cuts = [mp.mpf("0")]
        cursor = mp.mpf("0.25")
        while cursor < u_max:
            cuts.append(cursor)
            cursor += mp.mpf("0.25")
        cuts.append(u_max)
        integral = mp.quad(integrand, cuts)
        return mp.mpf("0.5") + z * (z - 1) * integral / 2


def xi_grid_float64(points: Sequence[complex], dps: int = DEFAULT_DPS) -> np.ndarray:
    values = []
    for point in points:
        value = xi_direct_mp(point, dps=dps)
        values.append(complex(float(mp.re(value)), float(mp.im(value))))
    result = np.asarray(values, dtype=np.complex128)
    if not np.all(np.isfinite(result.real)) or not np.all(np.isfinite(result.imag)):
        raise TRH0Error("reference grid contains non-finite values")
    return result


def _validate_support_and_points(points, support):
    s = np.asarray(points, dtype=np.complex128)
    x = np.asarray(support, dtype=np.float64)
    if s.ndim != 1 or x.ndim != 1 or s.size == 0 or x.size == 0:
        raise TRH0Error("points and support must be non-empty vectors")
    if not np.all((x > 0.0) & (x < 0.5)):
        raise TRH0Error("support must lie in (0, 1/2)")
    return s, x


def build_coupled_design(points, support) -> np.ndarray:
    """B_0=2 and B_k=x_k^(-s)+x_k^(s-1)."""
    s, x = _validate_support_and_points(points, support)
    logs = np.log(x)
    matrix = np.empty((s.size, x.size + 1), dtype=np.complex128)
    matrix[:, 0] = 2.0 + 0.0j
    matrix[:, 1:] = (
        np.exp(-s[:, None] * logs[None, :])
        + np.exp((s[:, None] - 1.0) * logs[None, :])
    )
    return matrix


def build_residual_design(points, support) -> np.ndarray:
    """R_0=1 and R_k=x_k^(-s)."""
    s, x = _validate_support_and_points(points, support)
    matrix = np.empty((s.size, x.size + 1), dtype=np.complex128)
    matrix[:, 0] = 1.0 + 0.0j
    matrix[:, 1:] = np.exp(-s[:, None] * np.log(x)[None, :])
    return matrix


def height_balancing_weights(
    points, targets, *, xi_half: complex | None = None, tau_scale: float = 1.0e-6
) -> np.ndarray:
    s = np.asarray(points, dtype=np.complex128)
    y = np.asarray(targets, dtype=np.complex128)
    if s.shape != y.shape or s.ndim != 1:
        raise TRH0Error("points and targets must have matching vector shapes")
    if xi_half is None:
        half = xi_direct_mp(0.5, dps=DEFAULT_DPS)
        xi_half = complex(float(mp.re(half)), float(mp.im(half)))
    tau = tau_scale * abs(xi_half)
    rounded_heights = np.round(s.imag, 12)
    weights = np.empty(s.size, dtype=np.float64)
    for height in np.unique(rounded_heights):
        mask = rounded_heights == height
        maximum = float(np.max(np.abs(y[mask])))
        weights[mask] = 1.0 / math.sqrt(maximum * maximum + tau * tau)
    return weights


def complex_to_real_system(design, target, row_weights=None) -> RealLinearSystem:
    matrix = np.asarray(design, dtype=np.complex128)
    y = np.asarray(target, dtype=np.complex128)
    if matrix.ndim != 2 or y.ndim != 1 or matrix.shape[0] != y.size:
        raise TRH0Error("design and target dimensions are incompatible")
    weights = (
        np.ones(y.size, dtype=np.float64)
        if row_weights is None
        else np.asarray(row_weights, dtype=np.float64)
    )
    if weights.shape != y.shape or np.any(weights <= 0) or not np.all(np.isfinite(weights)):
        raise TRH0Error("row weights must be finite, positive, and aligned")
    weighted_matrix = matrix * weights[:, None]
    weighted_target = y * weights
    return RealLinearSystem(
        np.vstack((weighted_matrix.real, weighted_matrix.imag)).astype(np.float64),
        np.concatenate((weighted_target.real, weighted_target.imag)).astype(np.float64),
        np.concatenate((weights, weights)),
    )


def normalize_penalized_columns(system: RealLinearSystem) -> ColumnNormalizedSystem:
    matrix = np.asarray(system.matrix, dtype=np.float64)
    if matrix.ndim != 2 or matrix.shape[1] < 2:
        raise TRH0Error("system requires intercept and atom columns")
    scales = np.ones(matrix.shape[1], dtype=np.float64)
    norms = np.linalg.norm(matrix[:, 1:], axis=0)
    if np.any(norms == 0) or not np.all(np.isfinite(norms)):
        raise TRH0Error("penalized column has invalid norm")
    scales[1:] = norms
    return ColumnNormalizedSystem(matrix / scales[None, :], system.target.copy(), scales)


def svd_diagnostics(matrix, relative_cutoff=EFFECTIVE_RANK_RELATIVE_CUTOFF):
    values = np.linalg.svd(np.asarray(matrix, dtype=np.float64), compute_uv=False)
    if values.size == 0 or values[0] == 0:
        raise TRH0Error("cannot diagnose zero-rank matrix")
    effective_rank = int(np.count_nonzero(values >= relative_cutoff * values[0]))
    condition = float(values[0] / values[-1]) if values[-1] > 0 else math.inf
    return {
        "singular_values": [float(value) for value in values],
        "effective_rank": effective_rank,
        "relative_cutoff": relative_cutoff,
        "condition_number": condition,
    }


def verify_reference_layer(dps: int = DEFAULT_DPS):
    support = frozen_support()
    samples = (0.5 + 0.0j, 0.25 + 3.0j, 0.5 + 14.0j, 0.1 + 27.0j)
    theta_errors = []
    functional_errors = []
    conjugation_errors = []
    for sample in samples:
        direct = xi_direct_mp(sample, dps=dps)
        theta = xi_theta_mp(sample, dps=dps)
        theta_errors.append(float(abs(direct - theta)))
        functional_errors.append(float(abs(direct - xi_direct_mp(1 - sample, dps=dps))))
        conjugation_errors.append(
            float(abs(xi_direct_mp(sample.conjugate(), dps=dps) - mp.conj(direct)))
        )

    probe_points = build_grid(SIGMAS, (0.0, 14.0, 28.0, 40.0))
    design = build_coupled_design(probe_points, support)
    symmetry_error = float(np.max(np.abs(design - build_coupled_design(1.0 - probe_points, support))))
    line_points = np.asarray([0.5 + 1j * t for t in (0.0, 14.0, 28.0, 40.0)])
    line_error = float(np.max(np.abs(build_coupled_design(line_points, support).imag)))

    train_points = build_grid(SIGMAS, TRAIN_HEIGHTS)
    train_targets = xi_grid_float64(train_points, dps=dps)
    train_design = build_coupled_design(train_points, support)
    weights = height_balancing_weights(train_points, train_targets)
    real_system = complex_to_real_system(train_design, train_targets, weights)
    normalized = normalize_penalized_columns(real_system)
    diagnostics = svd_diagnostics(normalized.matrix)

    tolerance = 10.0 ** (-(min(dps, 80) - 20))
    if max(theta_errors) > tolerance:
        raise TRH0Error("theta/Mellin cross-check failed")
    if max(functional_errors) > tolerance or max(conjugation_errors) > tolerance:
        raise TRH0Error("xi symmetry invariant failed")
    if symmetry_error > 1.0e-12 or line_error > 1.0e-12:
        raise TRH0Error("design-matrix invariant failed")

    rank = int(diagnostics["effective_rank"])
    return {
        "model": MODEL,
        "status": STATUS,
        "authority_effect": AUTHORITY_EFFECT,
        "canonical_effect": CANONICAL_EFFECT,
        "result": "OBSERVE_ONLY",
        "dps": dps,
        "atom_count": int(support.size),
        "training_grid_rows": int(train_points.size),
        "complex_design_shape": list(train_design.shape),
        "real_system_shape": list(real_system.matrix.shape),
        "max_direct_theta_error": max(theta_errors),
        "max_functional_equation_error": max(functional_errors),
        "max_conjugation_error": max(conjugation_errors),
        "max_design_symmetry_error_float64": symmetry_error,
        "max_critical_line_imaginary_error_float64": line_error,
        "effective_rank": rank,
        "condition_number": diagnostics["condition_number"],
        "numerical_status": (
            "FULL_EFFECTIVE_RANK" if rank == normalized.matrix.shape[1]
            else "NUMERICALLY_DEGENERATE"
        ),
    }


def _command_verify(args):
    print(json.dumps(verify_reference_layer(args.dps), indent=2, sort_keys=True, allow_nan=False))
    return 0


def build_parser():
    parser = argparse.ArgumentParser(description=__doc__)
    subparsers = parser.add_subparsers(dest="command", required=True)
    verify = subparsers.add_parser("verify")
    verify.add_argument("--dps", type=int, default=DEFAULT_DPS)
    verify.set_defaults(handler=_command_verify)
    return parser


def main(argv: Iterable[str] | None = None) -> int:
    args = build_parser().parse_args(list(argv) if argv is not None else None)
    try:
        return int(args.handler(args))
    except TRH0Error as error:
        print(json.dumps({
            "model": MODEL,
            "status": STATUS,
            "authority_effect": AUTHORITY_EFFECT,
            "result": "REJECT",
            "reason": str(error),
        }, sort_keys=True))
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
