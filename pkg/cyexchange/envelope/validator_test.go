// Status: PROPOSED
// authority_effect: NONE
package envelope

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"strings"
	"testing"
)

func signedTestEnvelope(t *testing.T) *CyExchangeEnvelope {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"request":"observe"}`)
	payloadHash, err := HashPayload(payload)
	if err != nil {
		t.Fatal(err)
	}
	env := &CyExchangeEnvelope{
		EnvelopeID: "ENV-1234",
		Sender: NodeCard{
			DID:              "did:key:sender",
			PublicKeyEd25519: hex.EncodeToString(pub),
			CapabilityScope:  "010-observer",
		},
		Receiver:         NodeCard{DID: "did:key:receiver"},
		RequiredState:    "010",
		PayloadSchemaURI: "urn:cyexchange:test",
		PayloadBytes:     payload,
		PayloadHash:      payloadHash,
	}
	signingBytes, err := SigningBytes(env)
	if err != nil {
		t.Fatal(err)
	}
	env.SignatureEd25519 = hex.EncodeToString(ed25519.Sign(priv, signingBytes))
	return env
}

func TestValidatorValid010(t *testing.T) {
	if err := Validate(signedTestEnvelope(t)); err != nil {
		t.Fatalf("expected valid envelope, got error: %v", err)
	}
}

func TestValidatorRejectsMutatedHash(t *testing.T) {
	env := signedTestEnvelope(t)
	env.PayloadBytes = []byte(`{"request":"MUTATE"}`)
	if err := Validate(env); err == nil {
		t.Fatal("expected validation to fail on mutated hash")
	}
}

func TestValidatorRejects101(t *testing.T) {
	env := signedTestEnvelope(t)
	env.RequiredState = "101"
	if err := Validate(env); err == nil {
		t.Fatal("expected 101 envelope to be rejected")
	}
}

func TestDecodeStrictRejectsDuplicateAndUnknownFields(t *testing.T) {
	tests := map[string]string{
		"duplicate": `{"envelope_id":"a","envelope_id":"b"}`,
		"unknown":   `{"unknown":true}`,
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeStrict(strings.NewReader(raw), MaxEnvelopeBytes); err == nil {
				t.Fatal("expected strict decoding failure")
			}
		})
	}
}

func TestDecodeStrictRejectsOversize(t *testing.T) {
	if _, err := DecodeStrict(bytes.NewReader(bytes.Repeat([]byte("x"), 9)), 8); err == nil {
		t.Fatal("expected byte-limit failure")
	}
}

func TestCanonicalizeRejectsDuplicatePayloadKeys(t *testing.T) {
	if _, err := Canonicalize([]byte(`{"claim":1,"claim":2}`)); err == nil {
		t.Fatal("expected duplicate payload key rejection")
	}
}
