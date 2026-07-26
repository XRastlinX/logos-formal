package main

import (
	"strings"
	"testing"
)

func TestBleedOverOperator(t *testing.T) {
	op := &BleedOverOperator{}

	sourceClaim := ClaimPacket{
		ID:            "C1",
		Proposition:   "Test",
		SourceContext: "Src",
		Authority:     1,
		Effect:        1,
		Uncertainty:   0.5,
		Provenance:    []string{"P1"},
	}

	targetCtx := TargetContext{
		FieldID:            "Target",
		AllowedVocab:       []string{"Test"},
		RequiresProvenance: true,
	}

	t.Run("Valid translation preserves uncertainty and provenance", func(t *testing.T) {
		bo, err := op.Translate(sourceClaim, targetCtx, "Test Translation", "Delta", 0.5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bo.Uncertainty != 0.5 {
			t.Errorf("expected uncertainty 0.5, got %v", bo.Uncertainty)
		}
		if len(bo.ProvenancePath) == 0 {
			t.Errorf("expected provenance to be preserved, got empty")
		}
	})

	t.Run("Decreasing uncertainty is rejected", func(t *testing.T) {
		_, err := op.Translate(sourceClaim, targetCtx, "Test", "Delta", 0.1)
		if err == nil {
			t.Errorf("expected error for certainty inflation")
		} else if !strings.Contains(err.Error(), "Uncertainty inflation detected") {
			t.Errorf("expected uncertainty error, got: %v", err)
		}
	})

	t.Run("Stripped provenance is rejected", func(t *testing.T) {
		noProvClaim := sourceClaim
		noProvClaim.Provenance = []string{}
		_, err := op.Translate(noProvClaim, targetCtx, "Test", "Delta", 0.5)
		if err == nil {
			t.Errorf("expected error for missing provenance")
		} else if !strings.Contains(err.Error(), "Claim lacks provenance") {
			t.Errorf("expected provenance error, got: %v", err)
		}
	})

	t.Run("Authority remains 0 after translation", func(t *testing.T) {
		bo, err := op.Translate(sourceClaim, targetCtx, "Test", "Delta", 0.6)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bo.Authority != 0 {
			t.Errorf("expected Authority to be 0, got %d", bo.Authority)
		}
		if bo.Effect != 0 {
			t.Errorf("expected Effect to be 0, got %d", bo.Effect)
		}
	})
}
