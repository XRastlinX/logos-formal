package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func testRequest(t *testing.T, now time.Time) (BoundaryRequest, EvaluationConfig, ed25519.PrivateKey) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	payload := sha256.Sum256([]byte("test payload"))
	artifact := Artifact{
		Proposer:      "external-test-client",
		Action:        "demo.echo",
		Target:        "test-target",
		PayloadDigest: "sha256:" + hex.EncodeToString(payload[:]),
	}
	permit := Permit{
		Schema:         permitSchema,
		ArtifactDigest: hashArtifact(artifact),
		Action:         artifact.Action,
		Target:         artifact.Target,
		Issuer:         "test-principal",
		ExpiresAt:      now.Add(time.Minute).Format(time.RFC3339),
		Nonce:          "test-nonce",
	}
	signPermit(&permit, privateKey)
	return BoundaryRequest{
			Schema:    requestSchema,
			RequestID: "test-request",
			Operation: "VERIFY_PERMIT_EVIDENCE",
			Artifact:  artifact,
			Permit:    permit,
		}, EvaluationConfig{
			TrustedIssuer: "test-principal",
			PublicKey:     publicKey,
			Now:           func() time.Time { return now },
			OriginClaim:   "external",
		}, privateKey
}

func TestValidPermitEvidenceNeverPerformsEffect(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	request, config, _ := testRequest(t, now)

	receipt, code := evaluate(request, config)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %+v", code, receipt.Friction)
	}
	if receipt.Verdict != "PERMIT_EVIDENCE_VALID" {
		t.Fatalf("unexpected verdict %q", receipt.Verdict)
	}
	if receipt.Authorization.Status != "EXTERNAL_PERMIT_VERIFIED" {
		t.Fatalf("unexpected authorization status %q", receipt.Authorization.Status)
	}
	if receipt.Effect.Status != "NOT_PERFORMED" {
		t.Fatalf("service performed or implied an effect: %+v", receipt.Effect)
	}
	if receipt.Routing.Decision != "OBSERVE_ONLY" || receipt.Routing.Forwarded {
		t.Fatalf("service forwarded an effectful route: %+v", receipt.Routing)
	}
	if receipt.AuthorityEffect != "NONE" || receipt.Authorization.ServiceAuthorityEffect != "NONE" {
		t.Fatal("service claimed authority")
	}
	if receipt.Externality.Status != "CLAIMED_EXTERNAL" {
		t.Fatalf("external call must remain unverified, got %q", receipt.Externality.Status)
	}
}

func TestEffectfulRouteIsAlwaysRejected(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	request, config, _ := testRequest(t, now)
	request.Operation = "APPLY"

	receipt, code := evaluate(request, config)
	if code != 1 || receipt.Friction.Code != "EFFECT_ROUTE_FORBIDDEN" {
		t.Fatalf("expected effect-route rejection, code=%d receipt=%+v", code, receipt)
	}
	if receipt.Routing.Forwarded || receipt.Effect.Status != "NOT_PERFORMED" {
		t.Fatal("effectful route escaped fail-closed boundary")
	}
}

func TestTamperedArtifactFailsClosed(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	request, config, _ := testRequest(t, now)
	request.Artifact.Target = "different-target"

	receipt, code := evaluate(request, config)
	if code != 1 || receipt.Friction.Code != "ARTIFACT_DIGEST_MISMATCH" {
		t.Fatalf("expected artifact mismatch rejection, code=%d receipt=%+v", code, receipt)
	}
	if receipt.Effect.Status != "NOT_PERFORMED" {
		t.Fatal("rejected request changed effect status")
	}
}

func TestPermitScopeMismatchFailsClosed(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	request, config, privateKey := testRequest(t, now)
	request.Permit.Target = "other-target"
	signPermit(&request.Permit, privateKey)

	receipt, code := evaluate(request, config)
	if code != 1 || receipt.Friction.Code != "PERMIT_SCOPE_MISMATCH" {
		t.Fatalf("expected scope rejection, code=%d receipt=%+v", code, receipt)
	}
}

func TestExpiredPermitFailsClosed(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	request, config, privateKey := testRequest(t, now)
	request.Permit.ExpiresAt = now.Format(time.RFC3339)
	signPermit(&request.Permit, privateKey)

	receipt, code := evaluate(request, config)
	if code != 1 || receipt.Friction.Code != "PERMIT_EXPIRED" {
		t.Fatalf("expected expiry rejection, code=%d receipt=%+v", code, receipt)
	}
}

func TestInvalidSignatureFailsClosed(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	request, config, _ := testRequest(t, now)
	request.Permit.Signature = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="

	receipt, code := evaluate(request, config)
	if code != 1 || receipt.Friction.Code != "INVALID_SIGNATURE" {
		t.Fatalf("expected signature rejection, code=%d receipt=%+v", code, receipt)
	}
}

func TestUntrustedIssuerFailsClosed(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	request, config, privateKey := testRequest(t, now)
	request.Permit.Issuer = "unknown-principal"
	signPermit(&request.Permit, privateKey)

	receipt, code := evaluate(request, config)
	if code != 1 || receipt.Friction.Code != "UNTRUSTED_ISSUER" {
		t.Fatalf("expected issuer rejection, code=%d receipt=%+v", code, receipt)
	}
}

func TestMalformedInputIsTryabilityFriction(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	request, config, _ := testRequest(t, now)
	request.Artifact.PayloadDigest = "not-a-digest"

	receipt, code := evaluate(request, config)
	if code != 2 || receipt.Friction.Category != "EVIDENCE" ||
		receipt.Friction.Code != "INVALID_PAYLOAD_DIGEST" {
		t.Fatalf("expected typed malformed digest, code=%d receipt=%+v", code, receipt)
	}
}

func TestArtifactDigestDeterministic(t *testing.T) {
	artifact := Artifact{
		Proposer:      "p",
		Action:        "a",
		Target:        "t",
		PayloadDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	first := hashArtifact(artifact)
	second := hashArtifact(artifact)
	if first != second || !validSHA256Digest(first) {
		t.Fatalf("artifact digest is not deterministic SHA-256: %q vs %q", first, second)
	}
}
