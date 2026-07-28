// Status: PROPOSED
// authority_effect: NONE
// Package crypto provides pure mathematical verification routines.
package crypto

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
)

// VerifyEd25519Signature performs a pure mathematical verification of a signature.
// This function strictly performs no network lookups or authority assertions.
func VerifyEd25519Signature(publicKeyHex, messageHex, signatureHex string) (bool, error) {
	pubKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return false, fmt.Errorf("invalid public key hex: %w", err)
	}
	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return false, fmt.Errorf("public key must be %d bytes", ed25519.PublicKeySize)
	}

	msgBytes, err := hex.DecodeString(messageHex)
	if err != nil {
		return false, fmt.Errorf("invalid message hex: %w", err)
	}

	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false, fmt.Errorf("invalid signature hex: %w", err)
	}

	return ed25519.Verify(ed25519.PublicKey(pubKeyBytes), msgBytes, sigBytes), nil
}
