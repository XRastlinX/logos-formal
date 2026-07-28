// Status: PROPOSED
// authority_effect: NONE
package envelope

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	jsoncanonicalizer "github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
)

// Canonicalize applies RFC 8785 JSON Canonicalization Scheme bytes. Invalid
// JSON is rejected; there is no permissive or simulated fallback.
func Canonicalize(payload []byte) ([]byte, error) {
	canonical, err := jsoncanonicalizer.Transform(payload)
	if err != nil {
		return nil, fmt.Errorf("RFC 8785 canonicalization: %w", err)
	}
	return canonical, nil
}

func HashPayload(payload []byte) (string, error) {
	canonical, err := Canonicalize(payload)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(canonical)
	return hex.EncodeToString(hash[:]), nil
}
