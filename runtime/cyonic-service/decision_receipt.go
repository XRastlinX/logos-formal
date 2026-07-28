package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gowebpki/jcs"
)

const (
	decisionReceiptSchema           = "urn:cyonic:decision-receipt:v1"
	decisionReceiptVersion          = "1.0"
	decisionReceiptSigningDomain    = "cyonic-decision-receipt-core.v1"
	decisionReceiptSigningAlgorithm = "Ed25519"
	decisionReceiptCanonicalization = "RFC8785"

	decisionReceiptVerificationSchema = "urn:cyonic:decision-receipt-verification:v1"
)

// DecisionReceipt is an immutable observer record. The proof is detached from
// the signed core so that no receipt identifier or signature is recursively
// included in the material it identifies.
type DecisionReceipt struct {
	Core  DecisionReceiptCore  `json:"core"`
	Proof DecisionReceiptProof `json:"proof"`
}

type DecisionReceiptCore struct {
	Schema                string                     `json:"schema"`
	Version               string                     `json:"version"`
	Canonicalization      string                     `json:"canonicalization"`
	ObservedAt            string                     `json:"observedAt"`
	Proposal              DecisionProposalBinding    `json:"proposal"`
	AuthorizationBasis    DecisionAuthorizationBasis `json:"authorizationBasis"`
	Decision              DecisionBinding            `json:"decision"`
	Effect                DecisionEffect             `json:"effect"`
	Governance            DecisionGovernance         `json:"governance"`
	PolicyRoot            string                     `json:"policyRoot"`
	OperatorRoot          string                     `json:"operatorRoot"`
	RuntimeRoot           string                     `json:"runtimeRoot"`
	EvidenceIDs           []string                   `json:"evidenceIds"`
	PreviousReceiptDigest string                     `json:"previousReceiptDigest,omitempty"`
}

type DecisionProposalBinding struct {
	RequestDigest  string `json:"requestDigest"`
	ArtifactDigest string `json:"artifactDigest"`
	Action         string `json:"action"`
	Target         string `json:"target"`
	PayloadDigest  string `json:"payloadDigest"`
}

type DecisionAuthorizationBasis struct {
	Kind               string `json:"kind"`
	PermitDigest       string `json:"permitDigest,omitempty"`
	VerificationStatus string `json:"verificationStatus"`
	Issuer             string `json:"issuer,omitempty"`
	KeyID              string `json:"keyId,omitempty"`
}

type DecisionBinding struct {
	Operation  string `json:"operation"`
	Outcome    string `json:"outcome"`
	ReasonCode string `json:"reasonCode"`
	Forwarded  bool   `json:"forwarded"`
}

type DecisionEffect struct {
	Status          string `json:"status"`
	AuthorityEffect string `json:"authorityEffect"`
}

type DecisionGovernance struct {
	State           string `json:"state"`
	AuthorityEffect string `json:"authorityEffect"`
}

type DecisionReceiptProof struct {
	Algorithm string `json:"algorithm"`
	KeyID     string `json:"keyId"`
	Signature string `json:"signature"`
}

type DecisionReceiptVerification struct {
	Schema                      string `json:"schema"`
	CryptographicStatus         string `json:"cryptographicStatus"`
	ArtifactStatus              string `json:"artifactStatus"`
	RouterDecision              string `json:"routerDecision"`
	ReceiptDigest               string `json:"receiptDigest,omitempty"`
	SignerKeyID                 string `json:"signerKeyId,omitempty"`
	RecordedOutcome             string `json:"recordedOutcome,omitempty"`
	RecordedAuthorizationStatus string `json:"recordedAuthorizationStatus,omitempty"`
	ReasonCode                  string `json:"reasonCode"`
	Detail                      string `json:"detail"`
	GovernanceState             string `json:"governanceState"`
	AuthorityEffect             string `json:"authorityEffect"`
	Effect                      string `json:"effect"`
	Forwarded                   bool   `json:"forwarded"`
}

func decisionReceiptCanonicalCore(core DecisionReceiptCore) ([]byte, error) {
	raw, err := json.Marshal(core)
	if err != nil {
		return nil, fmt.Errorf("encode decision receipt core: %w", err)
	}
	canonical, err := jcs.Transform(raw)
	if err != nil {
		return nil, fmt.Errorf("canonicalize decision receipt core with RFC 8785: %w", err)
	}
	return canonical, nil
}

func decisionReceiptSigningBytes(receipt DecisionReceipt) ([]byte, error) {
	canonical, err := decisionReceiptCanonicalCore(receipt.Core)
	if err != nil {
		return nil, err
	}
	prefix := []byte(decisionReceiptSigningDomain + "\x00")
	return append(prefix, canonical...), nil
}

func decisionReceiptDigest(receipt DecisionReceipt) (string, error) {
	signingBytes, err := decisionReceiptSigningBytes(receipt)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(signingBytes)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func keyIDForPublicKey(publicKey ed25519.PublicKey) string {
	digest := sha256.Sum256(publicKey)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func readDecisionReceipt(path string) (DecisionReceipt, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return DecisionReceipt{}, fmt.Errorf("read decision receipt: %w", err)
	}
	if !utf8.Valid(raw) {
		return DecisionReceipt{}, errors.New("decision receipt must be valid UTF-8")
	}
	if err := rejectDuplicateJSONKeys(raw); err != nil {
		return DecisionReceipt{}, fmt.Errorf("decode decision receipt: %w", err)
	}

	var receipt DecisionReceipt
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return DecisionReceipt{}, fmt.Errorf("decode decision receipt: %w", err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return DecisionReceipt{}, err
	}
	return receipt, nil
}

func rejectDuplicateJSONKeys(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	return ensureJSONEOF(decoder)
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}

	switch delimiter {
	case '{':
		seen := map[string]struct{}{}
		for decoder.More() {
			nameToken, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := nameToken.(string)
			if !ok {
				return errors.New("JSON object member name is not a string")
			}
			if _, exists := seen[name]; exists {
				return fmt.Errorf("duplicate JSON object member %q", name)
			}
			seen[name] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim('}') {
			return errors.New("JSON object is not properly terminated")
		}
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim(']') {
			return errors.New("JSON array is not properly terminated")
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
	return nil
}

func validateDecisionReceipt(receipt DecisionReceipt) error {
	core := receipt.Core
	if core.Schema != decisionReceiptSchema {
		return fmt.Errorf("unsupported decision receipt schema %q", core.Schema)
	}
	if core.Version != decisionReceiptVersion {
		return fmt.Errorf("unsupported decision receipt version %q", core.Version)
	}
	if core.Canonicalization != decisionReceiptCanonicalization {
		return fmt.Errorf("unsupported decision receipt canonicalization %q",
			core.Canonicalization)
	}
	if _, err := time.Parse(time.RFC3339Nano, core.ObservedAt); err != nil {
		return fmt.Errorf("observedAt must be RFC3339: %w", err)
	}

	digests := []struct {
		name  string
		value string
	}{
		{"proposal.requestDigest", core.Proposal.RequestDigest},
		{"proposal.artifactDigest", core.Proposal.ArtifactDigest},
		{"proposal.payloadDigest", core.Proposal.PayloadDigest},
		{"policyRoot", core.PolicyRoot},
		{"operatorRoot", core.OperatorRoot},
		{"runtimeRoot", core.RuntimeRoot},
	}
	for _, digest := range digests {
		if !validSHA256Digest(digest.value) {
			return fmt.Errorf("%s must be a SHA-256 digest", digest.name)
		}
	}
	if strings.TrimSpace(core.Proposal.Action) == "" ||
		strings.TrimSpace(core.Proposal.Target) == "" {
		return errors.New("proposal action and target are required")
	}

	if !slices.IsSorted(core.EvidenceIDs) {
		return errors.New("evidenceIds must be sorted lexicographically")
	}
	for i, evidenceID := range core.EvidenceIDs {
		if !validSHA256Digest(evidenceID) {
			return fmt.Errorf("evidenceIds[%d] must be a SHA-256 digest", i)
		}
		if i > 0 && evidenceID == core.EvidenceIDs[i-1] {
			return errors.New("evidenceIds must not contain duplicates")
		}
	}

	switch core.AuthorizationBasis.VerificationStatus {
	case "NOT_PRESENT":
		if core.AuthorizationBasis.Kind != "NONE" ||
			core.AuthorizationBasis.PermitDigest != "" ||
			core.AuthorizationBasis.Issuer != "" ||
			core.AuthorizationBasis.KeyID != "" {
			return errors.New("NOT_PRESENT authorization basis must not carry permit identity")
		}
	case "NOT_VERIFIED":
		if core.AuthorizationBasis.Kind != "EXTERNAL_PERMIT_EVIDENCE" {
			return errors.New("NOT_VERIFIED authorization basis must identify external permit evidence")
		}
		if core.AuthorizationBasis.PermitDigest != "" &&
			!validSHA256Digest(core.AuthorizationBasis.PermitDigest) {
			return errors.New("authorizationBasis.permitDigest must be empty or a SHA-256 digest")
		}
		if core.AuthorizationBasis.Issuer != "" &&
			strings.TrimSpace(core.AuthorizationBasis.Issuer) == "" {
			return errors.New("authorizationBasis.issuer must be empty or non-blank")
		}
		if core.AuthorizationBasis.KeyID != "" &&
			!validSHA256Digest(core.AuthorizationBasis.KeyID) {
			return errors.New("authorizationBasis.keyId must be empty or a SHA-256 digest")
		}
	case "VERIFIED":
		if core.AuthorizationBasis.Kind != "EXTERNAL_PERMIT_EVIDENCE" ||
			!validSHA256Digest(core.AuthorizationBasis.PermitDigest) ||
			strings.TrimSpace(core.AuthorizationBasis.Issuer) == "" ||
			!validSHA256Digest(core.AuthorizationBasis.KeyID) {
			return errors.New("VERIFIED authorization basis requires permit digest, issuer, and key ID")
		}
	default:
		return fmt.Errorf("unsupported authorization verification status %q",
			core.AuthorizationBasis.VerificationStatus)
	}
	if core.AuthorizationBasis.PermitDigest != "" &&
		!slices.Contains(core.EvidenceIDs, core.AuthorizationBasis.PermitDigest) {
		return errors.New("authorizationBasis.permitDigest must be present in evidenceIds")
	}

	if core.Decision.Operation != "VERIFY_PERMIT_EVIDENCE" {
		return fmt.Errorf("unsupported decision operation %q", core.Decision.Operation)
	}
	switch core.Decision.Outcome {
	case "OBSERVE_ONLY", "REJECT", "NO_DECISION":
	default:
		return fmt.Errorf("unsupported decision outcome %q", core.Decision.Outcome)
	}
	if strings.TrimSpace(core.Decision.ReasonCode) == "" {
		return errors.New("decision reasonCode is required")
	}
	if core.Decision.Forwarded {
		return errors.New("observer decision receipt cannot record forwarded=true")
	}
	if core.Effect.Status != "NOT_PERFORMED" ||
		core.Effect.AuthorityEffect != "NONE" {
		return errors.New("observer decision receipt must remain NOT_PERFORMED / NONE")
	}
	if core.Governance.State != "010" ||
		core.Governance.AuthorityEffect != "NONE" {
		return errors.New("observer decision receipt must remain 010 / NONE")
	}
	if core.PreviousReceiptDigest != "" &&
		!validSHA256Digest(core.PreviousReceiptDigest) {
		return errors.New("previousReceiptDigest must be empty or a SHA-256 digest")
	}
	if receipt.Proof.Algorithm != decisionReceiptSigningAlgorithm ||
		!validSHA256Digest(receipt.Proof.KeyID) {
		return errors.New("unsupported or incomplete decision receipt proof profile")
	}
	return nil
}

func verifyDecisionReceipt(
	receipt DecisionReceipt,
	publicKey ed25519.PublicKey,
) (string, error) {
	if err := validateDecisionReceipt(receipt); err != nil {
		return "", err
	}
	if len(publicKey) != ed25519.PublicKeySize {
		return "", fmt.Errorf("receipt trust key has %d bytes, expected %d",
			len(publicKey), ed25519.PublicKeySize)
	}
	expectedKeyID := keyIDForPublicKey(publicKey)
	if receipt.Proof.KeyID != expectedKeyID {
		return "", fmt.Errorf("receipt signer key ID %q does not match trusted key %q",
			receipt.Proof.KeyID, expectedKeyID)
	}
	signature, err := base64.StdEncoding.Strict().DecodeString(receipt.Proof.Signature)
	if err != nil {
		return "", fmt.Errorf("decode receipt signature: %w", err)
	}
	if len(signature) != ed25519.SignatureSize {
		return "", fmt.Errorf("receipt signature has %d bytes, expected %d",
			len(signature), ed25519.SignatureSize)
	}
	signingBytes, err := decisionReceiptSigningBytes(receipt)
	if err != nil {
		return "", err
	}
	if !ed25519.Verify(publicKey, signingBytes, signature) {
		return "", errors.New("decision receipt Ed25519 signature verification failed")
	}
	return decisionReceiptDigest(receipt)
}

func baseDecisionReceiptVerification() DecisionReceiptVerification {
	return DecisionReceiptVerification{
		Schema:              decisionReceiptVerificationSchema,
		CryptographicStatus: "NO_DECISION",
		ArtifactStatus:      "UNCLASSIFIED",
		RouterDecision:      "NONE",
		ReasonCode:          "NOT_EVALUATED",
		Detail:              "No decision receipt verification has completed.",
		GovernanceState:     "010",
		AuthorityEffect:     "NONE",
		Effect:              "NOT_PERFORMED",
		Forwarded:           false,
	}
}
