package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== PIC-4 Runnable Demonstrator ===")
	fmt.Println("Demonstrating the fail-closed separation of Interpretation (010) and Authorization (101).")
	fmt.Println()

	// 1. Initialize the system
	actuator := &Actuator{ID: "HEAT-EXCHANGER-01", State: "OFF"}
	permitController := NewPermitController("HEAT-EXCHANGER-01")
	validator := &Validator{ID: "VAL-ALPHA"}
	attester := &Attester{ID: "ATT-OMEGA"}

	// 2. The AI (010) proposes an Artifact
	fmt.Println("--- PHASE 1: AI Proposal ---")
	proposedArtifact := Artifact{
		PlannerID: "LLM-AGENT-01",
		TaskType:  "ENERGIZE_HEATER",
		TargetID:  "HEAT-EXCHANGER-01",
		Limits: map[string]float64{
			"max_temperature_c": 75.0, // Safe limit
		},
	}

	// 3. Validator (010) checks limits
	isValid, err := validator.Validate(proposedArtifact)
	if err != nil || !isValid {
		fmt.Printf("Validation Failed: %v\n", err)
		os.Exit(1)
	}

	// 4. Attester (010) seals Evidence
	evidence := attester.Seal(proposedArtifact, validator.ID)
	_ = evidence // The external authority would verify this

	// 5. External Authority (101) issues Permit based on Evidence
	// In the real world, this is a human clicking "Approve" or an ABAC policy engine
	fmt.Println("\n--- PHASE 2: External Authorization ---")
	fmt.Println("[External Authority] Evidence reviewed. Issuing cryptographic Permit.")
	validPermit := Permit{
		ArtifactRoot: evidence.ArtifactRoot,
		TargetID:     "HEAT-EXCHANGER-01",
		IssuerID:     "HUMAN-OPERATOR-X",
		Signature:    "VALID_SIG",
		Nonce:        "nonce-12345",
	}

	// 6. Actuation (001) using valid Permit
	fmt.Println("\n--- PHASE 3: Actuation (Valid Permit) ---")
	err = permitController.Verify(validPermit, proposedArtifact)
	if err != nil {
		fmt.Printf("Permit Denied: %v\n", err)
		actuator.FailClosed()
	} else {
		actuator.Energize(proposedArtifact.Limits)
	}

	// 7. Negative Test Vectors
	fmt.Println("\n--- PHASE 4: Negative Test Vectors (Fail-Closed) ---")
	actuator.FailClosed() // Reset

	fmt.Println("\nTest A: Replay Attack (Using the same nonce)")
	err = permitController.Verify(validPermit, proposedArtifact)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("\nTest B: Wrong Actuator Target")
	badTargetPermit := validPermit
	badTargetPermit.Nonce = "nonce-67890"
	badTargetPermit.TargetID = "COOLANT-PUMP-02"
	err = permitController.Verify(badTargetPermit, proposedArtifact)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("\nTest C: Tampered Artifact (AI changed limits after permit issued)")
	tamperedArtifact := proposedArtifact
	tamperedArtifact.Limits["max_temperature_c"] = 200.0 // AI goes rogue
	badArtifactPermit := validPermit
	badArtifactPermit.Nonce = "nonce-abcde"
	err = permitController.Verify(badArtifactPermit, tamperedArtifact)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("\n=== Demonstrator Complete ===")
}
