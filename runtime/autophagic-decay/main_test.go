package main

import (
	"testing"
)

func TestCalculateEffectiveUncertainty(t *testing.T) {
	obj := BoundaryObject{
		ID:              "TEST",
		Content:         "Test",
		BaseUncertainty: 0.1,
	}
	lambda := 0.2

	t.Run("D=0 returns base uncertainty", func(t *testing.T) {
		obj.GenerationalDepth = 0
		u := CalculateEffectiveUncertainty(obj, lambda)
		if u != 0.1 {
			t.Errorf("expected D=0 to return base uncertainty 0.1, got %v", u)
		}
	})

	t.Run("U_eff is monotonically increasing with D", func(t *testing.T) {
		var prev float64 = -1
		for d := 0; d < 10; d++ {
			obj.GenerationalDepth = d
			u := CalculateEffectiveUncertainty(obj, lambda)
			if u <= prev {
				t.Errorf("expected U_eff to increase at D=%d, got %v (prev: %v)", d, u, prev)
			}
			prev = u
		}
	})

	t.Run("U_eff approaches 1.0 as D grows large", func(t *testing.T) {
		obj.GenerationalDepth = 1000
		u := CalculateEffectiveUncertainty(obj, lambda)
		if u < 0.99 {
			t.Errorf("expected U_eff to approach 1.0 for large D, got %v", u)
		}
	})

	t.Run("U_eff never exceeds 1.0", func(t *testing.T) {
		obj.GenerationalDepth = 10000
		u := CalculateEffectiveUncertainty(obj, lambda)
		if u > 1.0 {
			t.Errorf("expected U_eff to never exceed 1.0, got %v", u)
		}
	})

	t.Run("Different lambda values", func(t *testing.T) {
		obj.GenerationalDepth = 5
		u1 := CalculateEffectiveUncertainty(obj, 0.1)
		u2 := CalculateEffectiveUncertainty(obj, 0.5)
		if u1 >= u2 {
			t.Errorf("expected larger lambda (0.5) to result in larger U_eff than smaller lambda (0.1)")
		}
	})
}
