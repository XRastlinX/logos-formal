// Status: PROPOSED
// authority_effect: NONE
// Package envelope handles the wrapping and validation of CyExchange payloads.
package envelope

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Canonicalize applies RFC 8785 (JCS) canonicalization to a payload.
// For the PROPOSED stub, this simulates canonicalization by standardizing JSON.
func Canonicalize(payload []byte) ([]byte, error) {
	var generic map[string]interface{}
	if err := json.Unmarshal(payload, &generic); err != nil {
		return nil, fmt.Errorf("invalid payload format, must be JSON for canonicalization: %w", err)
	}
	
	// Simulated canonicalization (standard Go json.Marshal sorts map keys).
	// In ACTIVE_CANON, this must strictly conform to RFC 8785.
	return json.Marshal(generic)
}

// HashPayload generates the SHA-256 hash of the canonicalized payload.
func HashPayload(payload []byte) (string, error) {
	canonical, err := Canonicalize(payload)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(canonical)
	return hex.EncodeToString(hash[:]), nil
}
