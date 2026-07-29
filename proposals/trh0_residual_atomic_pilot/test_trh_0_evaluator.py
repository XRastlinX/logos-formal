import math
import sys
import unittest
from pathlib import Path

import numpy as np

sys.path.insert(0, str(Path(__file__).resolve().parent))
import trh0_reference as ref
import trh_0_evaluator as evaluator


class TRH0EvaluatorTests(unittest.TestCase):
    def synthetic_data(self):
        support = ref.frozen_support()
        coefficients = np.linspace(-0.2, 0.2, 25, dtype=np.float64)
        coefficients[0] = 0.35

        def partition(heights):
            points = ref.build_grid((0.1, 0.3, 0.5), heights)
            design = ref.build_coupled_design(points, support)
            target = design @ coefficients
            weights = np.ones(points.size, dtype=np.float64)
            return evaluator.GridPartition(points, target, weights)

        data = evaluator.ExperimentData(
            train=partition((0.0, 0.5, 1.0, 1.5)),
            validation=partition((2.0, 2.5)),
            test=partition((3.0, 3.5)),
            dps=50,
            training_step=0.5,
        )
        return support, coefficients, data

    def test_ridge_lambda_grid_is_frozen(self):
        self.assertEqual(29, len(evaluator.RIDGE_LAMBDAS))
        self.assertAlmostEqual(1.0e-12, evaluator.RIDGE_LAMBDAS[0])
        self.assertAlmostEqual(1.0e2, evaluator.RIDGE_LAMBDAS[-1])
        self.assertTrue(
            all(
                left < right
                for left, right in zip(
                    evaluator.RIDGE_LAMBDAS, evaluator.RIDGE_LAMBDAS[1:]
                )
            )
        )

    def test_control_supports_are_genuinely_different(self):
        primary = ref.frozen_support()
        c1 = evaluator.opposite_residue_support()
        c2 = evaluator.chebyshev_support(primary)
        self.assertEqual((24,), c1.shape)
        self.assertEqual((24,), c2.shape)
        self.assertFalse(np.array_equal(primary, c1))
        self.assertFalse(np.array_equal(primary, c2))
        self.assertFalse(np.array_equal(c1, c2))

    def test_stratified_random_controls_are_deterministic(self):
        primary = ref.frozen_support()
        first = evaluator.stratified_random_supports(primary, count=3, seed=20260729)
        second = evaluator.stratified_random_supports(primary, count=3, seed=20260729)
        for left, right in zip(first, second):
            np.testing.assert_array_equal(left, right)
        self.assertFalse(np.array_equal(first[0], first[1]))

    def test_augmented_svd_ridge_preserves_unpenalized_intercept(self):
        matrix = np.asarray(
            [
                [1.0, 0.0],
                [1.0, 1.0],
                [1.0, 2.0],
                [1.0, 3.0],
            ],
            dtype=np.float64,
        )
        target = np.full(4, 4.25, dtype=np.float64)
        system = ref.RealLinearSystem(matrix, target, np.ones(4))
        normalized = ref.normalize_penalized_columns(system)
        solution = evaluator.solve_column_normalized_ridge_svd(
            normalized, lambda_value=1.0e6
        )
        self.assertAlmostEqual(4.25, solution.coefficients[0], places=5)
        self.assertAlmostEqual(0.0, solution.coefficients[1], places=5)

    def test_fit_support_applies_validation_tie_rule(self):
        support, _, data = self.synthetic_data()
        fit = evaluator.fit_support(
            "SYNTHETIC",
            support,
            data,
            lambdas=(1.0e-12, 1.0e-8, 1.0e-4, 1.0),
        )
        self.assertEqual(max(fit.admissible_lambdas), fit.selected_lambda)
        self.assertLess(fit.metrics["test"]["unweighted_rmse"], 1.0e-3)
        self.assertEqual((25,), fit.coefficients.shape)

    def test_argument_principle_finds_known_residual_zero(self):
        support = ref.frozen_support()
        coefficients = np.zeros(25, dtype=np.float64)
        x = float(support[0])
        coefficients[0] = 1.0
        coefficients[1] = -math.sqrt(x)
        period = 2.0 * math.pi / (-math.log(x))
        rectangle = evaluator.Rectangle(0.1, 0.9, period - 0.45, period + 0.45)
        count = evaluator.argument_principle_count(
            rectangle, coefficients, support
        )
        self.assertEqual(1, count)
        roots = evaluator.locate_residual_zeros(
            coefficients,
            support,
            rectangle=rectangle,
            verification_dps=60,
        )
        self.assertEqual(1, len(roots))
        self.assertAlmostEqual(0.5, roots[0]["real"], places=7)
        self.assertAlmostEqual(period, roots[0]["imag"], places=7)
        self.assertLess(roots[0]["verified_residual_abs"], 1.0e-10)

    def test_stable_zero_clustering_requires_every_variant(self):
        variants = (
            ({"real": 0.4, "imag": 10.0}, {"real": 0.7, "imag": 11.0}),
            ({"real": 0.41, "imag": 10.01},),
            ({"real": 0.39, "imag": 9.99},),
        )
        stable = evaluator.stable_zero_clusters(variants, distance=0.05)
        self.assertEqual(1, len(stable))
        self.assertTrue(stable[0]["off_center"])


if __name__ == "__main__":
    unittest.main()
