import sys
import unittest
from pathlib import Path

import mpmath as mp
import numpy as np

sys.path.insert(0, str(Path(__file__).resolve().parent))
import trh0_reference as trh0


class TRH0ReferenceTests(unittest.TestCase):
    def test_frozen_atom_family_and_support(self):
        trh0.validate_atom_family()
        support = trh0.frozen_support()
        self.assertEqual((24,), support.shape)
        self.assertTrue(np.all(support > 0.0))
        self.assertTrue(np.all(support < 0.5))
        self.assertTrue(np.all(np.diff(support) > 0.0))
        self.assertFalse(support.flags.writeable)

    def test_frozen_grid_shapes(self):
        train = trh0.build_grid(trh0.SIGMAS, trh0.TRAIN_HEIGHTS)
        validation = trh0.build_grid(trh0.SIGMAS, trh0.VALIDATION_HEIGHTS)
        test = trh0.build_grid(trh0.SIGMAS, trh0.TEST_HEIGHTS)
        self.assertEqual((730,), train.shape)
        self.assertEqual((330,), validation.shape)
        self.assertEqual((490,), test.shape)
        self.assertEqual(0.05 + 0.0j, train[0])
        self.assertEqual(0.50 + 18.0j, train[-1])

    def test_xi_direct_obeys_functional_equation_and_conjugation(self):
        with mp.workdps(70):
            s = mp.mpc("0.21", "12.75")
            value = trh0.xi_direct_mp(s, dps=70)
            reflected = trh0.xi_direct_mp(1 - s, dps=70)
            conjugated = trh0.xi_direct_mp(mp.conj(s), dps=70)
            self.assertLess(abs(value - reflected), mp.mpf("1e-60"))
            self.assertLess(abs(conjugated - mp.conj(value)), mp.mpf("1e-60"))

    def test_theta_mellin_cross_checks_direct_xi(self):
        for point in (0.5 + 0.0j, 0.25 + 3.0j, 0.5 + 14.0j):
            direct = trh0.xi_direct_mp(point, dps=65)
            theta = trh0.xi_theta_mp(point, dps=65)
            self.assertLess(abs(direct - theta), mp.mpf("1e-52"))

    def test_design_matrix_exact_structure(self):
        support = trh0.frozen_support()
        points = np.asarray([0.05 + 0.0j, 0.5 + 14.0j, 0.2 + 40.0j])
        design = trh0.build_coupled_design(points, support)
        self.assertEqual((3, 25), design.shape)
        np.testing.assert_array_equal(design[:, 0], np.full(3, 2.0 + 0.0j))
        reflected = trh0.build_coupled_design(1.0 - points, support)
        np.testing.assert_allclose(design, reflected, rtol=1e-13, atol=1e-13)

    def test_design_is_real_on_critical_line(self):
        support = trh0.frozen_support()
        points = np.asarray([0.5 + 0.0j, 0.5 + 14.0j, 0.5 + 40.0j])
        design = trh0.build_coupled_design(points, support)
        self.assertLess(float(np.max(np.abs(design.imag))), 1.0e-13)

    def test_real_system_and_column_normalization(self):
        support = trh0.frozen_support()
        points = trh0.build_grid(trh0.SIGMAS, (0.0, 0.25))
        design = trh0.build_coupled_design(points, support)
        target = trh0.xi_grid_float64(points, dps=60)
        weights = trh0.height_balancing_weights(points, target)
        system = trh0.complex_to_real_system(design, target, weights)
        normalized = trh0.normalize_penalized_columns(system)
        self.assertEqual((40, 25), normalized.matrix.shape)
        self.assertAlmostEqual(
            float(np.linalg.norm(normalized.matrix[:, 1])), 1.0, places=13
        )
        self.assertEqual(1.0, normalized.column_scales[0])

    def test_residual_design_has_unit_intercept(self):
        support = trh0.frozen_support()
        points = np.asarray([0.1 + 1.0j, 0.5 + 2.0j])
        residual = trh0.build_residual_design(points, support)
        np.testing.assert_array_equal(residual[:, 0], np.ones(2, dtype=np.complex128))


if __name__ == "__main__":
    unittest.main()
