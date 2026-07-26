// Status: PROPOSED
// authority_effect: NONE
// Package envelope handles the wrapping and validation of CyExchange payloads.
package envelope

import (
	"fmt"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/crypto"
)

// CyExchangeEnvelope represents the unmarshaled proto3 structure.
type CyExchangeEnvelope struct {
	EnvelopeID       string   `json:"envelope_id"`
	Sender           NodeCard `json:"sender"`
	Receiver         NodeCard `json:"receiver"`
	RequiredState    string   `json:"required_state"`
	PayloadSchemaURI string   `json:"payload_schema_uri"`
	PayloadBytes     []byte   `json:"payload_bytes"`
	PayloadHash      string   `json:"payload_hash"`
	SignatureEd25519 string   `json:"signature_ed25519"`
}

type NodeCard struct {
	DID              string `json:"did"`
	PublicKeyEd25519 string `json:"public_key_ed25519"`
	CapabilityScope  string `json:"capability_scope"`
}

// Validate routes the envelope through the 010 invariant checks.
func Validate(env *CyExchangeEnvelope) error {
	if env.RequiredState != "010" && env.RequiredState != "101" {
		return fmt.Errorf("invalid required_state: must be 010 or 101, got %s", env.RequiredState)
	}

	// 1. Verify Payload Hash
	expectedHash, err := HashPayload(env.PayloadBytes)
	if err != nil {
		return fmt.Errorf("failed to canonicalize and hash payload: %w", err)
	}
	if expectedHash != env.PayloadHash {
		return fmt.Errorf("payload hash mismatch: expected %s, got %s", expectedHash, env.PayloadHash)
	}

	// 2. Verify Sender Signature
	// The signature must cover (envelope_id + payload_hash)
	msgHex := fmt.Sprintf("%x%x", []byte(env.EnvelopeID), []byte(env.PayloadHash))
	valid, err := crypto.VerifyEd25519Signature(env.Sender.PublicKeyEd25519, msgHex, env.SignatureEd25519)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid ed25519 signature from sender")
	}

	// 3. 010 Observer constraints
	// The envelope is syntactically valid and intact.
	// We do NOT execute any 101 logic here.
	return nil
}
