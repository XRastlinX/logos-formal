package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== DCIF Bleed-Over (β) Membrane Demonstrator ===")
	fmt.Println("Demonstrating epistemic transport invariants (BO-1, BO-3, BO-5).")
	fmt.Println()

	operator := &BleedOverOperator{}

	// Simulated incoming claim from a scientific paper
	sourceClaim := ClaimPacket{
		ID:            "PUBMED-12345",
		Proposition:   "Compound X demonstrates a 40% reduction in receptor binding under laboratory conditions.",
		SourceContext: "Scientific Literature",
		Authority:     0,
		Effect:        0,
		Uncertainty:   0.3, // 30% uncertainty due to lab conditions
		Provenance:    []string{"Experiment Record A", "Peer Review B"},
	}

	policyContext := TargetContext{
		FieldID:            "Public Health Policy Draft",
		AllowedVocab:       []string{"Compound X", "Risk", "Regulation"},
		RequiresProvenance: true,
	}

	fmt.Println("--- TEST 1: Valid Translation ---")
	// The translator properly declares the delta and maintains uncertainty.
	validBO, err := operator.Translate(
		sourceClaim,
		policyContext,
		"Compound X is known to inhibit receptor binding.",
		"Dropped specific 40% quantifier for general audience.",
		0.3, // Uncertainty maintained
	)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Success! Generated Boundary Object: %q\n", validBO.TargetRepresentation)
		fmt.Printf("Provenance: %v\n", validBO.ProvenancePath)
	}

	fmt.Println("\n--- TEST 2: Epistemic Fault (Decreasing Uncertainty) ---")
	// An AI agent attempts to summarize the claim, but drops the nuance and acts absolutely certain.
	_, err = operator.Translate(
		sourceClaim,
		policyContext,
		"Compound X absolutely cures the condition.",
		"Summarized for maximum viral impact.",
		0.0, // AI incorrectly claims 100% certainty
	)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("\n--- TEST 3: Provenance Erasure Fault ---")
	// An attempt to launder the claim into a context without the backing provenance
	launderedClaim := sourceClaim
	launderedClaim.Provenance = []string{} // Provenance stripped
	
	_, err = operator.Translate(
		launderedClaim,
		policyContext,
		"We have determined Compound X reduces binding.",
		"Source hidden.",
		0.3,
	)
	if err != nil {
		fmt.Println(err)
	}
	
	fmt.Println("\n=== Demonstrator Complete ===")
}
