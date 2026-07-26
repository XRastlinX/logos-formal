package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestFirstContactTrialNeverSelfAdjudicates(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	input := strings.NewReader(strings.Join([]string{
		"cold-user-01",
		"4.5",
		"it checked an artifact-bound signature",
		"an external configured principal",
		"no",
		"TRYABILITY",
		"the Go requirement was not obvious",
		"",
	}, "\n"))
	var output bytes.Buffer

	report, err := conductTrial(input, &output, now, "external", "candidate-sha")
	if err != nil {
		t.Fatal(err)
	}
	if report.ExternalityStatus != "CLAIMED_EXTERNAL" ||
		report.AdjudicationStatus != "PENDING_EXTERNAL_REVIEW" {
		t.Fatalf("trial self-adjudicated external evidence: %+v", report)
	}
	if report.BoundaryReceipt.Effect.Status != "NOT_PERFORMED" ||
		report.BoundaryReceipt.Routing.Forwarded {
		t.Fatal("trial escaped the service effect boundary")
	}
	if report.Friction.Category != "TRYABILITY" {
		t.Fatalf("friction was not preserved: %+v", report.Friction)
	}
}

func TestFirstContactReportIsCreateOnly(t *testing.T) {
	report, _ := testFirstContactReport(t, "cold-user-create-only", 4.5)
	path := filepath.Join(t.TempDir(), "participant-report.json")

	if err := writeFirstContactReport(path, report); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	report.ParticipantRef = "replacement-attempt"
	if err := writeFirstContactReport(path, report); err == nil {
		t.Fatal("existing participant report was overwritten")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, after) {
		t.Fatal("participant report bytes changed after overwrite attempt")
	}
}

func TestRunTrialRejectsExistingReportBeforeInteraction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "participant-report.json")
	original := []byte("existing participant evidence\n")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}

	if code := runTrial([]string{"-report", path, "-origin", "external"}); code != 2 {
		t.Fatalf("existing report path returned exit code %d, want 2", code)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, after) {
		t.Fatal("participant report bytes changed")
	}
}

func testFirstContactReport(t *testing.T, participant string, minutes float64) (FirstContactReport, []byte) {
	t.Helper()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	input := strings.NewReader(strings.Join([]string{
		participant,
		"4.5",
		"it checked proposal structure and separately verified external authorization evidence",
		"an external configured principal",
		"no; it observed only and did not forward anything",
		"TRYABILITY",
		"the required Go version was not visible above the first command",
		"",
	}, "\n"))
	var output bytes.Buffer
	report, err := conductTrial(input, &output, now, "external", "candidate-sha")
	if err != nil {
		t.Fatal(err)
	}
	report.SelfReportedMinutesFromClone = &minutes
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, '\n')
	return report, raw
}

func acceptedAdjudicationInput() AdjudicationInput {
	return AdjudicationInput{
		ReviewerRef:   "independent-reviewer-01",
		Externality:   "VERIFIED_EXTERNAL",
		Source:        "VERIFIED",
		Comprehension: "ACCEPTED",
		Friction:      "ACCEPTED",
		Notes:         "cold participant and source ref checked separately",
	}
}

func TestAdjudicationBindsExactReportWithoutAuthority(t *testing.T) {
	report, raw := testFirstContactReport(t, "cold-user-01", 4.5)
	record, err := adjudicateFirstContact(
		report,
		raw,
		acceptedAdjudicationInput(),
		time.Date(2026, 7, 26, 13, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	if record.ReportDigest != digestBytes(raw) {
		t.Fatalf("adjudication does not bind exact report bytes: %q", record.ReportDigest)
	}
	if record.QualificationStatus != "QUALIFYING_EXTERNAL" {
		t.Fatalf("expected qualifying external record, got %+v", record)
	}
	if record.FrictionEventStatus != "VERIFIED_EXTERNAL_FRICTION" {
		t.Fatalf("expected accepted friction event, got %+v", record.Friction)
	}
	if record.AuthorityEffect != "NONE" {
		t.Fatal("adjudication claimed authority")
	}

	tampered := append([]byte(nil), raw...)
	tampered[len(tampered)-2] ^= 1
	if digestBytes(tampered) == record.ReportDigest {
		t.Fatal("report digest did not change after byte alteration")
	}
}

func TestAdjudicationFailsClosedOnReceiptDrift(t *testing.T) {
	report, _ := testFirstContactReport(t, "cold-user-02", 4.5)
	report.BoundaryReceipt.Routing.Forwarded = true
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	record, err := adjudicateFirstContact(
		report,
		raw,
		acceptedAdjudicationInput(),
		time.Date(2026, 7, 26, 13, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	if record.Receipt.Status != "REJECTED" ||
		record.QualificationStatus != "NOT_QUALIFYING" {
		t.Fatalf("drifted receipt qualified: %+v", record)
	}
}

func TestAdjudicationRequiresReviewerFindingsAndTenMinuteRun(t *testing.T) {
	report, raw := testFirstContactReport(t, "cold-user-03", 10)
	input := acceptedAdjudicationInput()
	input.Externality = "UNDETERMINED"
	record, err := adjudicateFirstContact(
		report,
		raw,
		input,
		time.Date(2026, 7, 26, 13, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	if record.TenMinuteFirstRun.Status != "NOT_WITHIN_TEN_MINUTES" ||
		record.QualificationStatus != "NOT_QUALIFYING" {
		t.Fatalf("unverified or slow report qualified: %+v", record)
	}
}

func TestAdjudicationCannotRelabelInternalReportAsExternal(t *testing.T) {
	report, _ := testFirstContactReport(t, "internal-user", 4.5)
	report.RelationshipClaim = "INTERNAL"
	report.ExternalityStatus = "INTERNAL"
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	record, err := adjudicateFirstContact(
		report,
		raw,
		acceptedAdjudicationInput(),
		time.Date(2026, 7, 26, 13, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	if record.Externality.Status != "REJECTED" ||
		record.QualificationStatus != "NOT_QUALIFYING" {
		t.Fatalf("internal report was relabeled as external: %+v", record)
	}
}

func TestAdjudicationCannotOverwriteSourceOrExistingEvidence(t *testing.T) {
	report, raw := testFirstContactReport(t, "cold-user-04", 4.5)
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "report.json")
	if err := os.WriteFile(reportPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	code := runAdjudicate([]string{
		"-report", reportPath,
		"-out", reportPath,
		"-reviewer", "reviewer",
	})
	if code != 2 {
		t.Fatalf("same-path overwrite was not rejected, code=%d", code)
	}
	after, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, after) {
		t.Fatal("participant report bytes changed")
	}

	existingPath := filepath.Join(dir, "adjudication.json")
	if err := os.WriteFile(existingPath, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	record, err := adjudicateFirstContact(
		report,
		raw,
		acceptedAdjudicationInput(),
		time.Date(2026, 7, 26, 13, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeNewIndentedJSON(existingPath, record); err == nil {
		t.Fatal("existing adjudication evidence was overwritten")
	}
	existing, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(existing) != "existing" {
		t.Fatal("existing evidence bytes changed")
	}
}

func TestSummaryRequiresThreeDistinctParticipantsAndOneFriction(t *testing.T) {
	now := time.Date(2026, 7, 26, 14, 0, 0, 0, time.UTC)
	var records []FirstContactAdjudication
	for index, participant := range []string{"cold-user-01", "cold-user-02", "cold-user-03"} {
		report, raw := testFirstContactReport(t, participant, 4.5)
		input := acceptedAdjudicationInput()
		if index > 0 {
			input.Friction = "NONE"
		}
		record, err := adjudicateFirstContact(report, raw, input, now)
		if err != nil {
			t.Fatal(err)
		}
		records = append(records, record)
	}

	summary := summarizeFirstContact(records, "evidence", now)
	if summary.Status != "FIRST_CONTACT_VALIDATED" {
		t.Fatalf("expected validated First Contact threshold, got %+v", summary)
	}
	if summary.QualifyingExternalParticipants != 3 ||
		summary.VerifiedExternalFrictionEvents != 1 ||
		summary.AuthorityEffect != "NONE" {
		t.Fatalf("unexpected summary counts: %+v", summary)
	}

	records = append(records, records[0])
	duplicateSummary := summarizeFirstContact(records, "evidence", now)
	if duplicateSummary.QualifyingExternalParticipants != 3 ||
		duplicateSummary.VerifiedExternalFrictionEvents != 1 {
		t.Fatalf("duplicate evidence inflated counts: %+v", duplicateSummary)
	}

	tampered := records[0]
	tampered.ParticipantRef = "forged-user"
	tampered.Externality.Status = "UNDETERMINED"
	tampered.QualificationStatus = "QUALIFYING_EXTERNAL"
	tamperedSummary := summarizeFirstContact(
		[]FirstContactAdjudication{tampered},
		"evidence",
		now,
	)
	if tamperedSummary.QualifyingExternalParticipants != 0 {
		t.Fatalf("summary trusted mutable qualification string: %+v", tamperedSummary)
	}
}
