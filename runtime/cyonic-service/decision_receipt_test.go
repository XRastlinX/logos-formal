package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testDigest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func signDecisionReceiptForTest(
	t *testing.T,
	receipt *DecisionReceipt,
	privateKey ed25519.PrivateKey,
) {
	t.Helper()
	signingBytes, err := decisionReceiptSigningBytes(*receipt)
	if err != nil {
		t.Fatal(err)
	}
	receipt.Proof.Signature = base64.StdEncoding.EncodeToString(
		ed25519.Sign(privateKey, signingBytes),
	)
}

func testSignedDecisionReceipt(
	t *testing.T,
) (DecisionReceipt, ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return testDecisionReceiptForKeys(t, publicKey, privateKey), publicKey, privateKey
}

func testDecisionReceiptForKeys(
	t *testing.T,
	publicKey ed25519.PublicKey,
	privateKey ed25519.PrivateKey,
) DecisionReceipt {
	t.Helper()
	receipt := DecisionReceipt{
		Core: testDecisionReceiptCore(),
		Proof: DecisionReceiptProof{
			Algorithm: decisionReceiptSigningAlgorithm,
			KeyID:     keyIDForPublicKey(publicKey),
		},
	}
	signDecisionReceiptForTest(t, &receipt, privateKey)
	return receipt
}

func testDecisionReceiptCore() DecisionReceiptCore {
	permitDigest := testDigest("permit")
	return DecisionReceiptCore{
		Schema:           decisionReceiptSchema,
		Version:          decisionReceiptVersion,
		Canonicalization: decisionReceiptCanonicalization,
		ObservedAt:       time.Date(2026, 7, 27, 20, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
		Proposal: DecisionProposalBinding{
			RequestDigest:  testDigest("request"),
			ArtifactDigest: testDigest("artifact"),
			Action:         "repository.propose",
			Target:         "refs/heads/cyonic/a4-boundary-hardening",
			PayloadDigest:  testDigest("payload"),
		},
		AuthorizationBasis: DecisionAuthorizationBasis{
			Kind:               "EXTERNAL_PERMIT_EVIDENCE",
			PermitDigest:       permitDigest,
			VerificationStatus: "VERIFIED",
			Issuer:             "test-principal",
			KeyID:              testDigest("permit-public-key"),
		},
		Decision: DecisionBinding{
			Operation:  "VERIFY_PERMIT_EVIDENCE",
			Outcome:    "OBSERVE_ONLY",
			ReasonCode: "NONE",
			Forwarded:  false,
		},
		Effect: DecisionEffect{
			Status:          "NOT_PERFORMED",
			AuthorityEffect: "NONE",
		},
		Governance: DecisionGovernance{
			State:           "010",
			AuthorityEffect: "NONE",
		},
		PolicyRoot:   testDigest("policy"),
		OperatorRoot: testDigest("observer-operator"),
		RuntimeRoot:  testDigest("runtime"),
		EvidenceIDs:  []string{permitDigest},
	}
}

func TestVerifyDecisionReceiptAcceptsExactSignedPayload(t *testing.T) {
	receipt, publicKey, _ := testSignedDecisionReceipt(t)
	digest, err := verifyDecisionReceipt(receipt, publicKey)
	if err != nil {
		t.Fatal(err)
	}
	expectedDigest, err := decisionReceiptDigest(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if digest != expectedDigest || !validSHA256Digest(digest) {
		t.Fatalf("unexpected decision receipt digest %q", digest)
	}
}

func TestDecisionReceiptCanonicalCoreUsesRFC8785Ordering(t *testing.T) {
	receipt, _, _ := testSignedDecisionReceipt(t)
	canonical, err := decisionReceiptCanonicalCore(receipt.Core)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(canonical, []byte(`{"authorizationBasis":`)) {
		t.Fatalf("core is not in RFC 8785 property order: %s", canonical)
	}
	if bytes.Contains(canonical, []byte(`"proof"`)) ||
		bytes.Contains(canonical, []byte(`"signature"`)) {
		t.Fatalf("detached proof leaked into signed core: %s", canonical)
	}
}

func TestDecisionReceiptGoldenVector(t *testing.T) {
	publicKeyBytes, err := base64.StdEncoding.Strict().DecodeString(
		"A6EHv/POEL4dcN0Y50vAmWfk1jCbpQ1fHdyGZBJVMbg=",
	)
	if err != nil {
		t.Fatal(err)
	}
	publicKey := ed25519.PublicKey(publicKeyBytes)
	receipt := DecisionReceipt{
		Core: testDecisionReceiptCore(),
		Proof: DecisionReceiptProof{
			Algorithm: decisionReceiptSigningAlgorithm,
			KeyID:     "sha256:56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c",
			Signature: "Ga9SvVuR3q/5HOeMczRGtHbi4/gMYPoMH3HX0RjIdMg/uqb68FDlyrDlr7sXY9EYZowxklTBa+Y/nVJfPOULDQ==",
		},
	}
	digest, err := verifyDecisionReceipt(receipt, publicKey)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := encodePublicKey(publicKey),
		"A6EHv/POEL4dcN0Y50vAmWfk1jCbpQ1fHdyGZBJVMbg="; got != want {
		t.Fatalf("public key = %q, want %q", got, want)
	}
	if got, want := receipt.Proof.KeyID,
		"sha256:56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c"; got != want {
		t.Fatalf("key ID = %q, want %q", got, want)
	}
	if got, want := digest,
		"sha256:21c171a69028ed80d8d6c169ed25163d840b39ac8940a405d4b5b9b41b8b5dfa"; got != want {
		t.Fatalf("receipt digest = %q, want %q", got, want)
	}
}

func TestVerifyDecisionReceiptRejectsTamperedProposal(t *testing.T) {
	receipt, publicKey, _ := testSignedDecisionReceipt(t)
	receipt.Core.Proposal.Target = "refs/heads/main"
	if _, err := verifyDecisionReceipt(receipt, publicKey); err == nil {
		t.Fatal("tampered proposal retained a valid decision receipt")
	}
}

func TestVerifyDecisionReceiptRejectsEffectOrForwardingClaims(t *testing.T) {
	for _, mutate := range []func(*DecisionReceipt){
		func(receipt *DecisionReceipt) { receipt.Core.Decision.Forwarded = true },
		func(receipt *DecisionReceipt) { receipt.Core.Effect.Status = "PERFORMED" },
		func(receipt *DecisionReceipt) { receipt.Core.Governance.State = "111" },
	} {
		receipt, publicKey, privateKey := testSignedDecisionReceipt(t)
		mutate(&receipt)
		signDecisionReceiptForTest(t, &receipt, privateKey)
		if _, err := verifyDecisionReceipt(receipt, publicKey); err == nil {
			t.Fatal("observer receipt accepted an effectful or authority-bearing claim")
		}
	}
}

func TestVerifyDecisionReceiptRejectsUnboundOrUnorderedEvidence(t *testing.T) {
	receipt, publicKey, privateKey := testSignedDecisionReceipt(t)
	receipt.Core.EvidenceIDs = []string{
		testDigest("z-evidence"),
		testDigest("a-evidence"),
	}
	signDecisionReceiptForTest(t, &receipt, privateKey)
	if _, err := verifyDecisionReceipt(receipt, publicKey); err == nil {
		t.Fatal("receipt accepted unordered and permit-unbound evidence identifiers")
	}
}

func TestReadDecisionReceiptRejectsDuplicateObjectKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "duplicate.json")
	raw := []byte(`{"core":{"schema":"first","schema":"second"},"proof":{}}`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readDecisionReceipt(path); err == nil {
		t.Fatal("receipt parser accepted duplicate object keys")
	}
}

func TestDecisionReceiptFileVerificationFailsClosed(t *testing.T) {
	receipt, publicKey, _ := testSignedDecisionReceipt(t)
	dir := t.TempDir()
	receiptPath := filepath.Join(dir, "receipt.json")
	keyPath := filepath.Join(dir, "receipt-public.key")

	raw, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(receiptPath, append(raw, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, []byte(encodePublicKey(publicKey)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, code := evaluateDecisionReceiptFile(receiptPath, keyPath, receipt.Proof.KeyID)
	if code != 0 ||
		result.CryptographicStatus != "VERIFIED" ||
		result.ArtifactStatus != "PROPOSED" ||
		result.RouterDecision != "OBSERVE_ONLY" ||
		result.AuthorityEffect != "NONE" ||
		result.Effect != "NOT_PERFORMED" ||
		result.Forwarded {
		t.Fatalf("unexpected receipt verification result: code=%d result=%+v", code, result)
	}

	result, code = evaluateDecisionReceiptFile(receiptPath, keyPath, testDigest("wrong-signer"))
	if code != 1 ||
		result.CryptographicStatus != "REJECTED" ||
		result.ReasonCode != "UNEXPECTED_SIGNER" ||
		result.SignerKeyID != "" ||
		result.RecordedOutcome != "" ||
		result.RecordedAuthorizationStatus != "" ||
		result.AuthorityEffect != "NONE" {
		t.Fatalf("unexpected signer did not fail closed: code=%d result=%+v", code, result)
	}
}

func TestVerifyDecisionReceiptRejectsMalformedNotVerifiedIdentity(t *testing.T) {
	receipt, publicKey, privateKey := testSignedDecisionReceipt(t)
	receipt.Core.AuthorizationBasis = DecisionAuthorizationBasis{
		Kind:               "EXTERNAL_PERMIT_EVIDENCE",
		VerificationStatus: "NOT_VERIFIED",
		Issuer:             "claimed-issuer",
		KeyID:              "not-a-digest",
	}
	receipt.Core.EvidenceIDs = nil
	signDecisionReceiptForTest(t, &receipt, privateKey)

	if _, err := verifyDecisionReceipt(receipt, publicKey); err == nil {
		t.Fatal("runtime verifier accepted a NOT_VERIFIED key ID rejected by the schema")
	}
}
