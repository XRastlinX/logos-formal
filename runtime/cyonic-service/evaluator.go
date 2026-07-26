package main

import (
	"crypto/ed25519"
	"fmt"
	"strings"
	"time"
)

type EvaluationConfig struct {
	TrustedIssuer string
	PublicKey     ed25519.PublicKey
	Now           func() time.Time
	OriginClaim   string
}

func baseReceipt(request BoundaryRequest, config EvaluationConfig) BoundaryReceipt {
	now := time.Now
	if config.Now != nil {
		now = config.Now
	}

	originClaim := strings.ToUpper(strings.TrimSpace(config.OriginClaim))
	if originClaim == "" {
		originClaim = "INTERNAL"
	}

	externalityStatus := "INTERNAL"
	externalityReason := "The caller was declared internal."
	if originClaim == "EXTERNAL" {
		externalityStatus = "CLAIMED_EXTERNAL"
		externalityReason = "Externality is caller/operator-declared and requires independent adjudication before it counts as external evidence."
	} else if originClaim != "INTERNAL" {
		externalityStatus = "UNDETERMINED"
		externalityReason = "The service cannot independently determine the caller's relationship to the project."
	}

	return BoundaryReceipt{
		Schema:          receiptSchema,
		RequestID:       request.RequestID,
		ObservedAt:      now().UTC().Format(time.RFC3339Nano),
		GovernanceState: "010",
		AuthorityEffect: "NONE",
		Routing: RoutingResult{
			Operation:       request.Operation,
			RequiredProfile: "010",
			Decision:        "REJECT",
			Forwarded:       false,
			Detail:          "Only read-only permit-evidence verification is routable.",
		},
		Interpretation: LayerResult{
			Status:          "NOT_VALIDATED",
			AuthorityEffect: "NONE",
			Detail:          "No interpretation-layer validation has completed.",
		},
		Authorization: AuthorizationResult{
			Status:                 "NOT_VERIFIED",
			ServiceAuthorityEffect: "NONE",
			Detail:                 "The service has not verified external authorization evidence.",
		},
		Effect: LayerResult{
			Status:          "NOT_PERFORMED",
			AuthorityEffect: "NONE",
			Detail:          "This service has no actuator or Apply capability.",
		},
		Verdict: "REJECTED",
		Friction: Friction{
			Category: "NONE",
			Code:     "NONE",
			Message:  "No friction classified.",
		},
		Externality: Externality{
			Claim:  originClaim,
			Status: externalityStatus,
			Reason: externalityReason,
		},
		Limits: []string{
			"Verifies one Ed25519 permit against one configured public key.",
			"Does not issue, consume, revoke, or persist permits.",
			"Does not perform the requested effect.",
			"Does not prove semantic truth, safety, ownership, or legal authority.",
			"Does not independently prove that a caller is external.",
		},
	}
}

func reject(receipt BoundaryReceipt, exitCode int, category, code, message string) (BoundaryReceipt, int) {
	receipt.Verdict = "REJECTED"
	receipt.Friction = Friction{
		Category: category,
		Code:     code,
		Message:  message,
	}
	return receipt, exitCode
}

func evaluate(request BoundaryRequest, config EvaluationConfig) (BoundaryReceipt, int) {
	receipt := baseReceipt(request, config)

	if request.Schema != requestSchema {
		return reject(receipt, 2, "TRYABILITY", "UNSUPPORTED_REQUEST_SCHEMA",
			fmt.Sprintf("request schema must be %q", requestSchema))
	}
	if strings.TrimSpace(request.RequestID) == "" {
		return reject(receipt, 2, "TRYABILITY", "MISSING_REQUEST_ID", "requestId is required")
	}
	if request.Operation != "VERIFY_PERMIT_EVIDENCE" {
		return reject(receipt, 1, "SCOPE_MISMATCH", "EFFECT_ROUTE_FORBIDDEN",
			"this service routes only VERIFY_PERMIT_EVIDENCE and has no Apply dispatcher")
	}
	if strings.TrimSpace(request.Artifact.Proposer) == "" ||
		strings.TrimSpace(request.Artifact.Action) == "" ||
		strings.TrimSpace(request.Artifact.Target) == "" {
		return reject(receipt, 2, "TRYABILITY", "MALFORMED_ARTIFACT",
			"artifact proposer, action, and target are required")
	}
	if !validSHA256Digest(request.Artifact.PayloadDigest) {
		return reject(receipt, 2, "EVIDENCE", "INVALID_PAYLOAD_DIGEST",
			"artifact payloadDigest must be sha256:<64 lowercase or uppercase hexadecimal characters>")
	}

	receipt.ArtifactDigest = hashArtifact(request.Artifact)
	receipt.Interpretation = LayerResult{
		Status:          "STRUCTURALLY_VALIDATED",
		AuthorityEffect: "NONE",
		Detail:          "Artifact fields and payload digest syntax are valid; semantic truth and safety were not evaluated.",
	}

	permit := request.Permit
	if permit.Schema != permitSchema {
		return reject(receipt, 2, "TRYABILITY", "UNSUPPORTED_PERMIT_SCHEMA",
			fmt.Sprintf("permit schema must be %q", permitSchema))
	}
	if len(config.PublicKey) != ed25519.PublicKeySize || strings.TrimSpace(config.TrustedIssuer) == "" {
		return reject(receipt, 2, "GOVERNANCE", "TRUST_CONFIGURATION_MISSING",
			"a trusted issuer and Ed25519 public key must be configured outside the request")
	}
	if permit.Issuer != config.TrustedIssuer {
		return reject(receipt, 1, "GOVERNANCE", "UNTRUSTED_ISSUER",
			"permit issuer does not match the configured trust anchor")
	}
	if permit.ArtifactDigest != receipt.ArtifactDigest {
		return reject(receipt, 1, "EVIDENCE", "ARTIFACT_DIGEST_MISMATCH",
			"permit is not bound to the exact artifact")
	}
	if permit.Action != request.Artifact.Action || permit.Target != request.Artifact.Target {
		return reject(receipt, 1, "GOVERNANCE", "PERMIT_SCOPE_MISMATCH",
			"permit action and target must exactly match the artifact")
	}
	if strings.TrimSpace(permit.Nonce) == "" {
		return reject(receipt, 2, "GOVERNANCE", "MISSING_NONCE", "permit nonce is required")
	}
	expiresAt, err := time.Parse(time.RFC3339, permit.ExpiresAt)
	if err != nil {
		return reject(receipt, 2, "TRYABILITY", "INVALID_EXPIRY",
			"permit expiresAt must be an RFC3339 timestamp")
	}
	now := time.Now
	if config.Now != nil {
		now = config.Now
	}
	if !now().UTC().Before(expiresAt.UTC()) {
		return reject(receipt, 1, "GOVERNANCE", "PERMIT_EXPIRED", "permit has expired")
	}
	if err := verifyPermitSignature(permit, config.PublicKey); err != nil {
		return reject(receipt, 1, "GOVERNANCE", "INVALID_SIGNATURE", err.Error())
	}

	receipt.Authorization = AuthorizationResult{
		Status:                 "EXTERNAL_PERMIT_VERIFIED",
		Issuer:                 permit.Issuer,
		Scope:                  permit.Action + "@" + permit.Target,
		ServiceAuthorityEffect: "NONE",
		Detail:                 "A signature from the configured external trust anchor matches the exact permit bytes.",
	}
	receipt.Routing = RoutingResult{
		Operation:       request.Operation,
		RequiredProfile: "010",
		Decision:        "OBSERVE_ONLY",
		Forwarded:       false,
		Detail:          "Permit evidence was verified, but the proposed action was not forwarded to any runtime.",
	}
	receipt.Verdict = "PERMIT_EVIDENCE_VALID"
	receipt.Friction = Friction{
		Category: "NONE",
		Code:     "NONE",
		Message:  "No boundary rejection occurred.",
	}
	return receipt, 0
}
