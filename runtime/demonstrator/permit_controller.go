package main

import (
	"fmt"
)

// PermitController sits directly in front of the Actuator. It receives Permits,
// verifies their cryptosignatures, and checks for nonce-reuse before passing
// the validated limits to the Actuator.
// Capability: 001 (Effectuation gate)
type PermitController struct {
	ExpectedTargetID string
	UsedNonces       map[string]bool
}

func NewPermitController(targetID string) *PermitController {
	return &PermitController{
		ExpectedTargetID: targetID,
		UsedNonces:       make(map[string]bool),
	}
}

// Verify enforces the PIC-4 execution invariant (Permit verification)
func (pc *PermitController) Verify(permit Permit, art Artifact) error {
	fmt.Printf("[PermitController] Verifying Permit for Artifact %s...\n", permit.ArtifactRoot[:8])

	if pc.UsedNonces[permit.Nonce] {
		return fmt.Errorf("SECURITY FAULT: Replay attack detected. Nonce %s already used", permit.Nonce)
	}

	if permit.TargetID != pc.ExpectedTargetID {
		return fmt.Errorf("SECURITY FAULT: Permit target mismatch. Expected %s, got %s", pc.ExpectedTargetID, permit.TargetID)
	}

	expectedRoot := HashArtifact(art)
	if permit.ArtifactRoot != expectedRoot {
		return fmt.Errorf("SECURITY FAULT: Artifact root mismatch. Artifact tampered with after Permit issuance")
	}

	if permit.Signature != "VALID_SIG" { // In reality: ECDSA/RSA verify
		return fmt.Errorf("SECURITY FAULT: Invalid cryptographic signature on Permit")
	}

	// Mark nonce as used to prevent replay
	pc.UsedNonces[permit.Nonce] = true
	fmt.Printf("[PermitController] Permit verified. Authorization granted.\n")
	return nil
}
