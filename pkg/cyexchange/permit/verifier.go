// Status: PROPOSED
// authority_effect: NONE
// Package permit evaluates the Sovereign Apply Permit state.
package permit

import (
	"fmt"
	"time"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/crypto"
)

// SovereignApplyPermit represents the unmarshaled proto3 structure.
type SovereignApplyPermit struct {
	PermitID         string `json:"permit_id"`
	IssuerDID        string `json:"issuer_did"`
	TargetDID        string `json:"target_did"`
	ArtifactHash     string `json:"artifact_hash"`
	ActionScope      string `json:"action_scope"`
	ValidFromUTC     int64  `json:"valid_from_utc"`
	ExpiresAtUTC     int64  `json:"expires_at_utc"`
	Nonce            string `json:"nonce"`
	SignatureEd25519 string `json:"signature_ed25519"`
}

// Verify evaluates the permit strictly as a 010 observer.
// It returns true if the permit is mathematically valid at the specified time.
func Verify(permit *SovereignApplyPermit, currentTimeUTC int64, issuerPubKeyHex string) error {
	// 1. Time bounds check
	if currentTimeUTC < permit.ValidFromUTC {
		return fmt.Errorf("permit is not yet valid (valid_from: %d, current: %d)", permit.ValidFromUTC, currentTimeUTC)
	}
	if currentTimeUTC > permit.ExpiresAtUTC {
		return fmt.Errorf("permit has expired (expires_at: %d, current: %d)", permit.ExpiresAtUTC, currentTimeUTC)
	}

	// 2. Structural checks
	if permit.ArtifactHash == "" {
		return fmt.Errorf("permit lacks artifact_hash binding")
	}

	// 3. Cryptographic Signature Validation
	// The message covers PermitID + TargetDID + ArtifactHash + Nonce
	msgHex := fmt.Sprintf("%x%x%x%x", []byte(permit.PermitID), []byte(permit.TargetDID), []byte(permit.ArtifactHash), []byte(permit.Nonce))
	
	valid, err := crypto.VerifyEd25519Signature(issuerPubKeyHex, msgHex, permit.SignatureEd25519)
	if err != nil {
		return fmt.Errorf("permit signature check failed: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid sovereign signature on permit")
	}

	return nil
}

// EnsureTime enforces a monotonic clock check (simulated).
func EnsureTime() int64 {
	return time.Now().Unix()
}
