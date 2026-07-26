// Status: PROPOSED
// authority_effect: NONE
package envelope

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestValidator_Valid010(t *testing.T) {
	// Generate valid Ed25519 keys
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	pubHex := hex.EncodeToString(pub)

	payload := []byte(`{"request":"observe"}`)
	payloadHash, _ := HashPayload(payload)

	envID := "ENV-1234"
	msgHex := fmt.Sprintf("%x%x", []byte(envID), []byte(payloadHash))
	msgBytes, _ := hex.DecodeString(msgHex)

	sig := ed25519.Sign(priv, msgBytes)
	sigHex := hex.EncodeToString(sig)

	env := &CyExchangeEnvelope{
		EnvelopeID: envID,
		Sender: NodeCard{
			PublicKeyEd25519: pubHex,
		},
		RequiredState: "010",
		PayloadBytes:  payload,
		PayloadHash:   payloadHash,
		SignatureEd25519: sigHex,
	}

	if err := Validate(env); err != nil {
		t.Fatalf("expected valid envelope, got error: %v", err)
	}
}

func TestValidator_RejectsMutatedHash(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	payload := []byte(`{"request":"observe"}`)
	payloadHash, _ := HashPayload(payload)

	envID := "ENV-1234"
	msgHex := fmt.Sprintf("%x%x", []byte(envID), []byte(payloadHash))
	msgBytes, _ := hex.DecodeString(msgHex)
	sig := ed25519.Sign(priv, msgBytes)

	env := &CyExchangeEnvelope{
		EnvelopeID: envID,
		Sender: NodeCard{PublicKeyEd25519: hex.EncodeToString(pub)},
		RequiredState: "010",
		PayloadBytes:  []byte(`{"request":"MUTATE"}`), // Altered payload
		PayloadHash:   payloadHash,
		SignatureEd25519: hex.EncodeToString(sig),
	}

	if err := Validate(env); err == nil {
		t.Fatal("expected validation to fail on mutated hash")
	}
}
