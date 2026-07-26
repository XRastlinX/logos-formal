// Status: PROPOSED
// authority_effect: NONE
package permit

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestVerifier_ExpiredPermit(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	pubHex := hex.EncodeToString(pub)

	permit := &SovereignApplyPermit{
		PermitID:     "P-1",
		TargetDID:    "did:target",
		ArtifactHash: "hash123",
		Nonce:        "nonce123",
		ValidFromUTC: 100,
		ExpiresAtUTC: 200,
	}

	msgHex := fmt.Sprintf("%x%x%x%x", []byte(permit.PermitID), []byte(permit.TargetDID), []byte(permit.ArtifactHash), []byte(permit.Nonce))
	msgBytes, _ := hex.DecodeString(msgHex)
	sig := ed25519.Sign(priv, msgBytes)
	permit.SignatureEd25519 = hex.EncodeToString(sig)

	// Test with current time AFTER expiration
	currentTime := int64(300)
	err := Verify(permit, currentTime, pubHex)
	if err == nil {
		t.Fatal("expected permit to be rejected due to expiration")
	}
}
