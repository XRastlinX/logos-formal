#!/usr/bin/env python3
"""TRH-0 SVD ridge, matched-control, and residual-zero pipeline.

Status: PROPOSED. Authority/canonical effect: NONE. Results are OBSERVE_ONLY.
"""
from __future__ import annotations

import argparse
import json
import math
import platform
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable, Sequence

import mpmath as mp
import numpy as np
import trh0_reference as ref

MODEL = ref.MODEL
STATUS = "PROPOSED"
AUTHORITY_EFFECT = CANONICAL_EFFECT = "NONE"
RESULT_MODE = "OBSERVE_ONLY"
MASTER_RANDOM_SEED = 20260729
DEFAULT_RANDOM_CONTROLS = 100
DEFAULT_DPS = 80
RIDGE_RELATIVE_CUTOFF = 1e-12
ZERO_STABILITY_DISTANCE = 0.05
OFF_CENTER_TOLERANCE = 1e-7
PRIMES_5_MOD_6 = (5, 11, 17, 23, 29, 41, 47, 53, 59, 71, 83, 89,
                  101, 107, 113, 131, 137, 149, 167, 173, 179, 191, 197, 227)
RIDGE_LAMBDAS = tuple(float(10 ** (-12 + .5 * i)) for i in range(29))


@dataclass(frozen=True)
class GridPartition:
    points: np.ndarray
    target: np.ndarray
    weights: np.ndarray


@dataclass(frozen=True)
class ExperimentData:
    train: GridPartition
    validation: GridPartition
    test: GridPartition
    dps: int
    training_step: float


@dataclass(frozen=True)
class RidgeSolution:
    lambda_value: float
    coefficients: np.ndarray
    normalized_coefficients: np.ndarray
    singular_values: np.ndarray
    effective_rank: int
    condition_number: float
    coefficient_norm: float


@dataclass(frozen=True)
class FitResult:
    family: str
    support: np.ndarray
    selected_lambda: float
    admissible_lambdas: tuple[float, ...]
    coefficients: np.ndarray
    validation_path: tuple[dict[str, float], ...]
    diagnostics: dict[str, object]
    metrics: dict[str, dict[str, float]]


@dataclass(frozen=True)
class Rectangle:
    sigma_min: float
    sigma_max: float
    t_min: float
    t_max: float

    def validate(self):
        v = (self.sigma_min, self.sigma_max, self.t_min, self.t_max)
        if not all(math.isfinite(x) for x in v) or self.sigma_min >= self.sigma_max or self.t_min >= self.t_max:
            raise ref.TRH0Error("invalid rectangle")

    @property
    def width(self): return self.sigma_max - self.sigma_min

    @property
    def height(self): return self.t_max - self.t_min

    def contains(self, z, tol=1e-8):
        return (self.sigma_min-tol <= z.real <= self.sigma_max+tol and
                self.t_min-tol <= z.imag <= self.t_max+tol)


def _support(values, count=24):
    x = np.sort(np.asarray(values, dtype=np.float64))
    if x.shape != (count,) or np.unique(x).size != count or not np.all(np.isfinite(x)) or not np.all((x > 0) & (x < .5)):
        raise ref.TRH0Error("invalid support")
    return x


def opposite_residue_support():
    if any(not ref._is_prime(p) or p % 6 != 5 for p in PRIMES_5_MOD_6):
        raise ref.TRH0Error("invalid C1 atoms")
    return _support([ref.embed_atom(p) for p in PRIMES_5_MOD_6])


def support_frequency_interval(support):
    w = -np.log(_support(support))
    return float(w.min()), float(w.max())


def chebyshev_support(base_support):
    lo, hi = support_frequency_interval(base_support)
    k = np.arange(1, 25, dtype=float)
    roots = np.cos((2*k-1)*math.pi/48)
    return _support(np.exp(-((lo+hi)/2 + (hi-lo)*roots/2)))


def stratified_random_supports(base_support, count=100, seed=MASTER_RANDOM_SEED):
    if count <= 0 or seed < 0: raise ref.TRH0Error("invalid random control declaration")
    lo, hi = support_frequency_interval(base_support)
    edges = np.linspace(lo, hi, 25)
    rng = np.random.Generator(np.random.PCG64(seed))
    return tuple(_support(np.exp(-rng.uniform(edges[:-1], edges[1:]))) for _ in range(count))


def training_heights(step):
    if not math.isfinite(step) or step <= 0: raise ref.TRH0Error("invalid training step")
    n = round(18/step)
    if not math.isclose(n*step, 18, abs_tol=1e-10): raise ref.TRH0Error("step must divide 18")
    return tuple(round(i*step, 12) for i in range(n+1))


def _partition(heights, dps):
    points = ref.build_grid(ref.SIGMAS, heights)
    target = ref.xi_grid_float64(points, dps=dps)
    return GridPartition(points, target, ref.height_balancing_weights(points, target))


def build_experiment_data(dps=DEFAULT_DPS, train_step=.25):
    if dps < 30: raise ref.TRH0Error("dps must be >=30")
    return ExperimentData(_partition(training_heights(train_step), dps),
                          _partition(ref.VALIDATION_HEIGHTS, dps),
                          _partition(ref.TEST_HEIGHTS, dps), dps, float(train_step))


def solve_column_normalized_ridge_svd(system, lambda_value, relative_cutoff=RIDGE_RELATIVE_CUTOFF):
    if not math.isfinite(lambda_value) or lambda_value < 0: raise ref.TRH0Error("invalid lambda")
    A, y = np.asarray(system.matrix, float), np.asarray(system.target, float)
    if lambda_value:
        P = np.zeros((A.shape[1]-1, A.shape[1])); P[:, 1:] = np.eye(A.shape[1]-1)
        A = np.vstack((A, math.sqrt(lambda_value)*P)); y = np.r_[y, np.zeros(P.shape[0])]
    U, sv, Vh = np.linalg.svd(A, full_matrices=False)
    threshold = relative_cutoff*sv[0]
    keep = sv >= threshold
    inv = np.zeros_like(sv); inv[keep] = 1/sv[keep]
    beta = Vh.T @ (inv*(U.T@y))
    c = system.restore_coefficients(beta)
    return RidgeSolution(float(lambda_value), c, beta, sv, int(keep.sum()),
                         float(sv[0]/sv[-1]) if sv[-1] else math.inf,
                         float(np.linalg.norm(c)))


def complex_metrics(design, coefficients, target, weights):
    r = np.asarray(design) @ np.asarray(coefficients) - np.asarray(target)
    wr = np.asarray(weights)*r
    return {"height_balanced_mse": float(np.mean(abs(wr)**2)),
            "height_balanced_rmse": float(np.sqrt(np.mean(abs(wr)**2))),
            "unweighted_rmse": float(np.sqrt(np.mean(abs(r)**2))),
            "unweighted_max_residual": float(np.max(abs(r)))}


def _training_system(support, data):
    B = ref.build_coupled_design(data.train.points, support)
    real = ref.complex_to_real_system(B, data.train.target, data.train.weights)
    normalized = ref.normalize_penalized_columns(real)
    return normalized, ref.svd_diagnostics(normalized.matrix, RIDGE_RELATIVE_CUTOFF)


def fit_support(family, support, data, lambdas=RIDGE_LAMBDAS):
    x = _support(support); lambdas = tuple(float(v) for v in lambdas)
    if not lambdas or tuple(sorted(set(lambdas))) != lambdas: raise ref.TRH0Error("invalid lambda path")
    normalized, diag = _training_system(x, data)
    partitions = {"train": data.train, "validation": data.validation, "test": data.test}
    designs = {k: ref.build_coupled_design(v.points, x) for k, v in partitions.items()}
    path, solutions = [], {}
    for lam in lambdas:
        sol = solve_column_normalized_ridge_svd(normalized, lam); solutions[lam] = sol
        val = complex_metrics(designs["validation"], sol.coefficients, data.validation.target, data.validation.weights)
        path.append({"lambda": lam, "validation_height_balanced_mse": val["height_balanced_mse"],
                     "validation_unweighted_rmse": val["unweighted_rmse"],
                     "coefficient_norm": sol.coefficient_norm,
                     "effective_rank": float(sol.effective_rank), "condition_number": sol.condition_number})
    minimum = min(v["validation_height_balanced_mse"] for v in path)
    band = tuple(v["lambda"] for v in path if v["validation_height_balanced_mse"] <= 1.01*minimum)
    selected = solutions[max(band)]
    metrics = {k: complex_metrics(designs[k], selected.coefficients, partitions[k].target, partitions[k].weights)
               for k in partitions}
    diagnostics = {"unregularized_singular_values": diag["singular_values"],
                   "unregularized_effective_rank": diag["effective_rank"],
                   "unregularized_condition_number": diag["condition_number"],
                   "selected_augmented_singular_values": [float(v) for v in selected.singular_values],
                   "selected_augmented_effective_rank": selected.effective_rank,
                   "selected_augmented_condition_number": selected.condition_number,
                   "selected_coefficient_norm": selected.coefficient_norm,
                   "training_step": data.training_step}
    return FitResult(family, x, selected.lambda_value, band, selected.coefficients,
                     tuple(path), diagnostics, metrics)


def coefficients_for_lambda(support, data, lambda_value):
    normalized, _ = _training_system(support, data)
    return solve_column_normalized_ridge_svd(normalized, lambda_value).coefficients


def _summary(fit, coefficients=True):
    out = {"family": fit.family, "support": fit.support.tolist(),
           "selected_lambda": fit.selected_lambda, "admissible_lambdas": list(fit.admissible_lambdas),
           "metrics": fit.metrics, "diagnostics": fit.diagnostics, "validation_path": list(fit.validation_path)}
    if coefficients: out["coefficients"] = fit.coefficients.tolist()
    return out


def run_reconstruction_assay(dps=DEFAULT_DPS, random_control_count=100, seed=MASTER_RANDOM_SEED):
    data = build_experiment_data(dps)
    primary_support = ref.frozen_support()
    primary = fit_support("PRIMARY_PRIMES_1_MOD_6", primary_support, data)
    controls = [fit_support("C1_PRIMES_5_MOD_6", opposite_residue_support(), data),
                fit_support("C2_CHEBYSHEV_MATCHED", chebyshev_support(primary_support), data)]
    randoms = [fit_support(f"C3_STRATIFIED_RANDOM_{i:03d}", x, data)
               for i, x in enumerate(stratified_random_supports(primary_support, random_control_count, seed))]
    pe = primary.metrics["test"]["height_balanced_mse"]
    best = min(c.metrics["test"]["height_balanced_mse"] for c in controls)
    re = np.array([c.metrics["test"]["height_balanced_mse"] for c in randoms])
    p05 = float(np.percentile(re, 5))
    return {"proposal": MODEL, "status": STATUS, "authority_effect": AUTHORITY_EFFECT,
            "canonical_effect": CANONICAL_EFFECT, "result": RESULT_MODE,
            "implemented_stage": "SVD_RIDGE_AND_MATCHED_CONTROLS", "dps": dps,
            "python_version": platform.python_version(), "numpy_version": np.__version__,
            "mpmath_version": mp.__version__, "random_generator": "PCG64", "master_seed": seed,
            "random_control_count": random_control_count, "primary": _summary(primary),
            "deterministic_controls": [_summary(c) for c in controls],
            "random_controls": [{"family": c.family, "selected_lambda": c.selected_lambda,
                                 "test_height_balanced_mse": c.metrics["test"]["height_balanced_mse"],
                                 "test_unweighted_rmse": c.metrics["test"]["unweighted_rmse"],
                                 "effective_rank": c.diagnostics["unregularized_effective_rank"],
                                 "coefficient_norm": c.diagnostics["selected_coefficient_norm"]} for c in randoms],
            "evidence_threshold": {"primary_test_height_balanced_mse": pe,
                                   "best_deterministic_control_mse": best,
                                   "required_deterministic_ratio": .8,
                                   "observed_deterministic_ratio": pe/best if best else math.inf,
                                   "random_fifth_percentile_mse": p05,
                                   "primary_below_random_fifth_percentile": pe < p05,
                                   "reconstruction_threshold_passed": pe < .8*best and pe < p05},
            "claim_boundary": {"functional_symmetry_counts_as_evidence": False,
                               "training_error_counts_as_evidence": False,
                               "trh_established": False, "riemann_hypothesis_established": False}}


def residual_values(points, coefficients, support):
    s, c, x = np.asarray(points, complex), np.asarray(coefficients, float), np.asarray(support, float)
    if c.shape != (x.size+1,): raise ref.TRH0Error("coefficient/support mismatch")
    return np.asarray(c[0] + np.exp(-s[..., None]*np.log(x)) @ c[1:], complex)


def residual_derivative(point, coefficients, support):
    c, x = np.asarray(coefficients, float), np.asarray(support, float); lx = np.log(x)
    return complex(np.sum(-c[1:]*lx*np.exp(-complex(point)*lx)))


def _boundary(r, n):
    r.validate()
    b = np.linspace(r.sigma_min, r.sigma_max, n, endpoint=False)+1j*r.t_min
    right = r.sigma_max+1j*np.linspace(r.t_min, r.t_max, n, endpoint=False)
    top = np.linspace(r.sigma_max, r.sigma_min, n, endpoint=False)+1j*r.t_max
    left = r.sigma_min+1j*np.linspace(r.t_max, r.t_min, n, endpoint=False)
    p = np.r_[b, right, top, left]; return np.r_[p, p[:1]]


def argument_principle_count(rectangle, coefficients, support, initial_samples_per_edge=32,
                             maximum_samples_per_edge=4096, boundary_relative_tolerance=1e-12):
    previous = None; n = initial_samples_per_edge
    while n <= maximum_samples_per_edge:
        v = residual_values(_boundary(rectangle, n), coefficients, support); a = abs(v)
        if not np.isfinite(a.max()) or a.max() == 0: raise ref.TRH0Error("invalid contour values")
        if a.min() <= boundary_relative_tolerance*a.max(): raise ref.TRH0Error("BOUNDARY_ZERO_AMBIGUITY")
        steps = np.angle(v[1:]/v[:-1]); winding = round(float(steps.sum()/(2*math.pi)))
        if winding < 0: raise ref.TRH0Error("negative winding")
        if previous == winding and abs(steps).max() < math.pi/2: return int(winding)
        previous = winding; n *= 2
    raise ref.TRH0Error("argument count did not stabilize before sample limit")


def _newton(rectangle, coefficients, support):
    re = np.linspace(rectangle.sigma_min, rectangle.sigma_max, 7)
    im = np.linspace(rectangle.t_min, rectangle.t_max, 7)
    seeds = np.array([complex(x, y) for y in im for x in re]); z = complex(seeds[np.argmin(abs(residual_values(seeds, coefficients, support)))])
    scale = max(rectangle.width, rectangle.height, 1)
    for _ in range(100):
        f = complex(residual_values([z], coefficients, support)[0])
        if abs(f) <= 1e-12: return z if rectangle.contains(z, 1e-6) else None
        fp = residual_derivative(z, coefficients, support)
        if abs(fp) <= 1e-15: return None
        step = f/fp
        if abs(step) > scale: step *= scale/abs(step)
        candidate = z-step
        for _ in range(12):
            if rectangle.contains(candidate, .05*scale): break
            step *= .5; candidate = z-step
        z = candidate
    return None


def _children(r, rf, tf):
    x = r.sigma_min+rf*r.width; y = r.t_min+tf*r.height
    return (Rectangle(r.sigma_min, x, r.t_min, y), Rectangle(x, r.sigma_max, r.t_min, y),
            Rectangle(r.sigma_min, x, y, r.t_max), Rectangle(x, r.sigma_max, y, r.t_max))


def locate_residual_zeros(coefficients, support, rectangle=Rectangle(0, 1, 0, 40),
                          maximum_depth=18, minimum_width=1e-5, minimum_height=1e-4,
                          verification_dps=DEFAULT_DPS):
    x = _support(support); c = np.asarray(coefficients, float); roots = []
    splits = ((.5,.5),(.47,.53),(.53,.47),(.38196601125,.61803398875),(.61803398875,.38196601125))
    def walk(cell, count=None, depth=0):
        count = argument_principle_count(cell, c, x) if count is None else count
        if not count: return
        if count == 1:
            z = _newton(cell, c, x)
            if z is not None: roots.append((z, cell)); return
        if depth >= maximum_depth or (cell.width <= minimum_width and cell.height <= minimum_height):
            raise ref.TRH0Error(f"ZERO_ISOLATION_UNRESOLVED:{count}:{depth}")
        last = None
        for rf, tf in splits:
            children = _children(cell, rf, tf)
            try: counts = tuple(argument_principle_count(ch, c, x) for ch in children)
            except ref.TRH0Error as e: last = e; continue
            if sum(counts) != count: last = ref.TRH0Error("child count mismatch"); continue
            for ch, n in zip(children, counts):
                if n: walk(ch, n, depth+1)
            return
        raise ref.TRH0Error(f"ZERO_SUBDIVISION_FAILED:{last}")
    walk(rectangle)
    unique = []
    for z, cell in sorted(roots, key=lambda v: (v[0].imag, v[0].real)):
        if not any(abs(z-u[0]) <= 1e-7 for u in unique): unique.append((z, cell))
    out = []
    with mp.workdps(verification_dps):
        mx = [mp.mpf(str(v)) for v in x]; mc = [mp.mpf(str(v)) for v in c]
        for z, cell in unique:
            mz = mp.mpc(str(z.real), str(z.imag)); total = mc[0]
            for coeff, value in zip(mc[1:], mx): total += coeff*mp.power(value, -mz)
            out.append({"real": float(mp.re(mz)), "imag": float(mp.im(mz)),
                        "distance_from_critical_line": abs(float(mp.re(mz))-.5),
                        "off_center": abs(float(mp.re(mz))-.5) > OFF_CENTER_TOLERANCE,
                        "multiplicity_estimate": 1, "verified_residual_abs": float(abs(total)),
                        "isolating_rectangle": {"sigma_min": cell.sigma_min, "sigma_max": cell.sigma_max,
                                                "t_min": cell.t_min, "t_max": cell.t_max},
                        "distance_from_search_boundary": min(z.real-rectangle.sigma_min, rectangle.sigma_max-z.real,
                                                             z.imag-rectangle.t_min, rectangle.t_max-z.imag)})
    return tuple(out)


def stable_zero_clusters(variants, distance=ZERO_STABILITY_DISTANCE):
    if not variants: return ()
    stable = []
    for root in variants[0]:
        center = complex(root["real"], root["imag"]); matches = [center]
        for variant in variants[1:]:
            candidates = [complex(v["real"], v["imag"]) for v in variant]
            if not candidates: break
            nearest = min(candidates, key=lambda v: abs(v-center))
            if abs(nearest-center) > distance: break
            matches.append(nearest)
        else:
            mean = sum(matches)/len(matches)
            stable.append({"real": mean.real, "imag": mean.imag,
                           "distance_from_critical_line": abs(mean.real-.5),
                           "off_center": abs(mean.real-.5) > OFF_CENTER_TOLERANCE,
                           "maximum_variant_displacement": max(abs(v-mean) for v in matches),
                           "variant_count": len(matches)})
    return tuple(stable)


def run_support_zero_stability(family, support, dps=DEFAULT_DPS, training_steps=(.2,.25,.3),
                               rectangle=Rectangle(0,1,0,40), data_by_step=None):
    variants, labels, fits = [], [], []
    for step in training_steps:
        data = build_experiment_data(dps, step) if data_by_step is None else data_by_step[float(step)]
        fit = fit_support(family, support, data)
        fits.append({"training_step": step, "selected_lambda": fit.selected_lambda,
                     "admissible_lambdas": list(fit.admissible_lambdas)})
        for lam in fit.admissible_lambdas:
            c = coefficients_for_lambda(support, data, lam)
            try: zeros = locate_residual_zeros(c, support, rectangle, verification_dps=dps)
            except ref.TRH0Error as e:
                labels.append({"training_step": step, "lambda": lam, "status": "NUMERICALLY_UNRESOLVED",
                               "reason": str(e), "zero_count": None, "off_center_zero_count": None, "zeros": []})
                continue
            variants.append(zeros)
            labels.append({"training_step": step, "lambda": lam, "status": "RESOLVED",
                           "zero_count": len(zeros), "off_center_zero_count": sum(bool(z["off_center"]) for z in zeros),
                           "zeros": list(zeros)})
    unresolved = sum(v["status"] != "RESOLVED" for v in labels)
    resolved = unresolved == 0 and len(variants) == len(labels)
    stable = stable_zero_clusters(variants) if resolved else ()
    return {"family": family,
            "search_rectangle": {"sigma_min": rectangle.sigma_min, "sigma_max": rectangle.sigma_max,
                                 "t_min": rectangle.t_min, "t_max": rectangle.t_max},
            "fit_variants": fits, "root_variants": labels,
            "stability_status": "RESOLVED" if resolved else "NUMERICALLY_UNRESOLVED",
            "unresolved_variant_count": unresolved, "stable_zeros": list(stable),
            "stable_zero_count": len(stable) if resolved else None,
            "stable_off_center_zero_count": sum(bool(z["off_center"]) for z in stable) if resolved else None,
            "stability_distance": ZERO_STABILITY_DISTANCE}


def run_zero_control_assay(dps=DEFAULT_DPS, random_control_count=100, seed=MASTER_RANDOM_SEED,
                           rectangle=Rectangle(0,1,0,40)):
    primary_support = ref.frozen_support()
    shared = {step: build_experiment_data(dps, step) for step in (.2,.25,.3)}
    primary = run_support_zero_stability("PRIMARY_PRIMES_1_MOD_6", primary_support, dps=dps,
                                         rectangle=rectangle, data_by_step=shared)
    controls = [run_support_zero_stability("C1_PRIMES_5_MOD_6", opposite_residue_support(), dps=dps,
                                           rectangle=rectangle, data_by_step=shared),
                run_support_zero_stability("C2_CHEBYSHEV_MATCHED", chebyshev_support(primary_support), dps=dps,
                                           rectangle=rectangle, data_by_step=shared)]
    randoms = [run_support_zero_stability(f"C3_STRATIFIED_RANDOM_{i:03d}", x, dps=dps,
                                          rectangle=rectangle, data_by_step=shared)
               for i, x in enumerate(stratified_random_supports(primary_support, random_control_count, seed))]
    all_controls = controls+randoms
    unresolved = [v["family"] for v in [primary]+all_controls if v["stability_status"] != "RESOLVED"]
    if unresolved:
        p95 = None; primary_count = primary["stable_off_center_zero_count"]; passed = False; status = "NOT_EVALUABLE"
    else:
        counts = np.array([v["stable_off_center_zero_count"] for v in all_controls], float)
        p95 = float(np.percentile(counts, 95)); primary_count = int(primary["stable_off_center_zero_count"])
        passed = primary_count > p95; status = "EVALUATED"
    return {"proposal": MODEL, "status": STATUS, "authority_effect": AUTHORITY_EFFECT,
            "canonical_effect": CANONICAL_EFFECT, "result": RESULT_MODE,
            "implemented_stage": "RESIDUAL_ZERO_CONTROL_ASSAY", "primary": primary,
            "deterministic_controls": controls, "random_controls": randoms,
            "evidence_threshold": {"status": status, "primary_stable_off_center_zero_count": primary_count,
                                   "control_95th_percentile": p95, "zero_threshold_passed": passed,
                                   "unresolved_families": unresolved},
            "claim_boundary": {"trh_established": False, "riemann_hypothesis_established": False}}


def _write(value, output):
    text = json.dumps(value, indent=2, sort_keys=True, allow_nan=False)
    if output:
        p = Path(output); p.parent.mkdir(parents=True, exist_ok=True); p.write_text(text+"\n", encoding="utf-8")
    print(text)


def _family(name, index, seed):
    primary = ref.frozen_support()
    if name == "primary": return "PRIMARY_PRIMES_1_MOD_6", primary
    if name == "c1": return "C1_PRIMES_5_MOD_6", opposite_residue_support()
    if name == "c2": return "C2_CHEBYSHEV_MATCHED", chebyshev_support(primary)
    if index < 0: raise ref.TRH0Error("negative random index")
    return f"C3_STRATIFIED_RANDOM_{index:03d}", stratified_random_supports(primary, index+1, seed)[index]


def build_parser():
    p = argparse.ArgumentParser(description=__doc__); subs = p.add_subparsers(dest="command", required=True)
    fit = subs.add_parser("fit"); fit.add_argument("--dps", type=int, default=80); fit.add_argument("--random-controls", type=int, default=100); fit.add_argument("--seed", type=int, default=MASTER_RANDOM_SEED); fit.add_argument("--output")
    zero = subs.add_parser("zeros"); zero.add_argument("--family", choices=("primary","c1","c2","random"), default="primary"); zero.add_argument("--random-index", type=int, default=0); zero.add_argument("--dps", type=int, default=80); zero.add_argument("--seed", type=int, default=MASTER_RANDOM_SEED); zero.add_argument("--output")
    assay = subs.add_parser("zero-assay"); assay.add_argument("--dps", type=int, default=80); assay.add_argument("--random-controls", type=int, default=100); assay.add_argument("--seed", type=int, default=MASTER_RANDOM_SEED); assay.add_argument("--output")
    return p


def main(argv: Iterable[str] | None = None):
    args = build_parser().parse_args(list(argv) if argv is not None else None)
    try:
        if args.command == "fit": result = run_reconstruction_assay(args.dps, args.random_controls, args.seed)
        elif args.command == "zeros":
            name, support = _family(args.family, args.random_index, args.seed)
            result = run_support_zero_stability(name, support, dps=args.dps)
            result.update({"proposal": MODEL, "status": STATUS, "authority_effect": AUTHORITY_EFFECT,
                           "canonical_effect": CANONICAL_EFFECT, "result": RESULT_MODE})
        else: result = run_zero_control_assay(args.dps, args.random_controls, args.seed)
        _write(result, args.output); return 0
    except ref.TRH0Error as e:
        print(json.dumps({"proposal": MODEL, "status": STATUS, "authority_effect": AUTHORITY_EFFECT,
                          "canonical_effect": CANONICAL_EFFECT, "result": "REJECT", "reason": str(e)})); return 2


if __name__ == "__main__": raise SystemExit(main())
