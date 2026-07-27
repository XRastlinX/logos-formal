package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	firstContactReportSchema       = "urn:cyonic:first-contact-report:v1"
	firstContactAdjudicationSchema = "urn:cyonic:first-contact-adjudication:v1"
	firstContactSummarySchema      = "urn:cyonic:first-contact-summary:v1"
)

type ReviewFinding struct {
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type FirstContactAdjudication struct {
	Schema              string        `json:"schema"`
	AdjudicationID      string        `json:"adjudicationId"`
	AdjudicatedAt       string        `json:"adjudicatedAt"`
	ReportDigest        string        `json:"reportDigest"`
	TrialID             string        `json:"trialId"`
	SourceRef           string        `json:"sourceRef"`
	ParticipantRef      string        `json:"participantRef"`
	ReviewerRef         string        `json:"reviewerRef"`
	Externality         ReviewFinding `json:"externality"`
	Source              ReviewFinding `json:"source"`
	Receipt             ReviewFinding `json:"receipt"`
	Comprehension       ReviewFinding `json:"comprehension"`
	TenMinuteFirstRun   ReviewFinding `json:"tenMinuteFirstRun"`
	Friction            ReviewFinding `json:"friction"`
	QualificationStatus string        `json:"qualificationStatus"`
	FrictionEventStatus string        `json:"frictionEventStatus"`
	AuthorityEffect     string        `json:"authorityEffect"`
	Notes               string        `json:"notes,omitempty"`
	Limits              []string      `json:"limits"`
}

type FirstContactSummary struct {
	Schema                         string   `json:"schema"`
	GeneratedAt                    string   `json:"generatedAt"`
	InputDirectory                 string   `json:"inputDirectory"`
	AdjudicationFilesRead          int      `json:"adjudicationFilesRead"`
	UniqueReports                  int      `json:"uniqueReports"`
	QualifyingExternalParticipants int      `json:"qualifyingExternalParticipants"`
	VerifiedExternalFrictionEvents int      `json:"verifiedExternalFrictionEvents"`
	ConflictingRecords             int      `json:"conflictingRecords"`
	Status                         string   `json:"status"`
	AuthorityEffect                string   `json:"authorityEffect"`
	QualifyingTrialIDs             []string `json:"qualifyingTrialIds"`
	FrictionTrialIDs               []string `json:"frictionTrialIds"`
	Limits                         []string `json:"limits"`
}

type AdjudicationInput struct {
	ReviewerRef   string
	Externality   string
	Source        string
	Comprehension string
	Friction      string
	Notes         string
}

func decodeStrictJSON[T any](raw []byte) (T, error) {
	var value T
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, err
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return value, err
	}
	return value, nil
}

func readFirstContactReport(path string) (FirstContactReport, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return FirstContactReport{}, nil, fmt.Errorf("read report: %w", err)
	}
	report, err := decodeStrictJSON[FirstContactReport](raw)
	if err != nil {
		return FirstContactReport{}, nil, fmt.Errorf("decode report: %w", err)
	}
	if report.Schema != firstContactReportSchema {
		return FirstContactReport{}, nil, fmt.Errorf("unexpected report schema %q", report.Schema)
	}
	return report, raw, nil
}

func digestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func normalizedReviewStatus(value string, allowed ...string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	for _, candidate := range allowed {
		if normalized == candidate {
			return normalized, nil
		}
	}
	return "", fmt.Errorf("status %q must be one of %s", value, strings.Join(allowed, ", "))
}

func inspectTrialReceipt(report FirstContactReport) ReviewFinding {
	receipt := report.BoundaryReceipt
	if report.AuthorityEffect != "NONE" {
		return ReviewFinding{Status: "REJECTED", Detail: "report authorityEffect is not NONE"}
	}
	if report.AdjudicationStatus != "PENDING_EXTERNAL_REVIEW" {
		return ReviewFinding{Status: "REJECTED", Detail: "original report is not pending independent review"}
	}
	if receipt.Schema != receiptSchema ||
		receipt.GovernanceState != "010" ||
		receipt.AuthorityEffect != "NONE" ||
		receipt.Routing.Decision != "OBSERVE_ONLY" ||
		receipt.Routing.Forwarded ||
		receipt.Effect.Status != "NOT_PERFORMED" ||
		receipt.Effect.AuthorityEffect != "NONE" ||
		receipt.Authorization.ServiceAuthorityEffect != "NONE" {
		return ReviewFinding{
			Status: "REJECTED",
			Detail: "embedded receipt violates the required 010, observe-only, no-effect shape",
		}
	}
	return ReviewFinding{
		Status: "STRUCTURE_CONSISTENT",
		Detail: "embedded receipt retains 010, OBSERVE_ONLY, forwarded=false, NOT_PERFORMED, and authorityEffect=NONE",
	}
}

func inspectTenMinuteRun(report FirstContactReport) ReviewFinding {
	if report.SelfReportedMinutesFromClone == nil {
		return ReviewFinding{
			Status: "UNDETERMINED",
			Detail: "participant did not provide a parseable non-negative setup time",
		}
	}
	minutes := *report.SelfReportedMinutesFromClone
	if minutes < 10 {
		return ReviewFinding{
			Status: "WITHIN_TEN_MINUTES",
			Detail: fmt.Sprintf("participant self-reported %.2f minutes from clone/setup", minutes),
		}
	}
	return ReviewFinding{
		Status: "NOT_WITHIN_TEN_MINUTES",
		Detail: fmt.Sprintf("participant self-reported %.2f minutes from clone/setup", minutes),
	}
}

func concreteFriction(report FirstContactReport) bool {
	category := strings.ToUpper(strings.TrimSpace(report.Friction.Category))
	detail := strings.TrimSpace(report.Friction.Detail)
	return category != "" &&
		category != "NONE" &&
		category != "UNCLASSIFIED" &&
		detail != "" &&
		!strings.EqualFold(detail, "none")
}

func adjudicateFirstContact(
	report FirstContactReport,
	rawReport []byte,
	input AdjudicationInput,
	now time.Time,
) (FirstContactAdjudication, error) {
	if strings.TrimSpace(input.ReviewerRef) == "" {
		return FirstContactAdjudication{}, errors.New("reviewer reference is required")
	}

	externality, err := normalizedReviewStatus(
		input.Externality,
		"VERIFIED_EXTERNAL", "NOT_EXTERNAL", "UNDETERMINED",
	)
	if err != nil {
		return FirstContactAdjudication{}, err
	}
	source, err := normalizedReviewStatus(
		input.Source,
		"VERIFIED", "REJECTED", "UNDETERMINED",
	)
	if err != nil {
		return FirstContactAdjudication{}, err
	}
	comprehension, err := normalizedReviewStatus(
		input.Comprehension,
		"ACCEPTED", "REJECTED", "UNDETERMINED",
	)
	if err != nil {
		return FirstContactAdjudication{}, err
	}
	friction, err := normalizedReviewStatus(
		input.Friction,
		"ACCEPTED", "REJECTED", "NONE", "UNDETERMINED",
	)
	if err != nil {
		return FirstContactAdjudication{}, err
	}

	receiptFinding := inspectTrialReceipt(report)
	timeFinding := inspectTenMinuteRun(report)
	externalityFinding := ReviewFinding{
		Status: externality,
		Detail: "reviewer attestation; the service does not independently discover participant identity",
	}
	if externality == "VERIFIED_EXTERNAL" &&
		(strings.ToUpper(strings.TrimSpace(report.RelationshipClaim)) != "EXTERNAL" ||
			report.ExternalityStatus != "CLAIMED_EXTERNAL") {
		externalityFinding = ReviewFinding{
			Status: "REJECTED",
			Detail: "reviewer finding conflicts with the participant report's recorded relationship claim",
		}
	}
	sourceFinding := ReviewFinding{
		Status: source,
		Detail: "reviewer attestation that the declared source ref exists and contains the trial implementation",
	}
	if source == "VERIFIED" && strings.TrimSpace(report.SourceRef) == "UNDECLARED" {
		sourceFinding = ReviewFinding{
			Status: "REJECTED",
			Detail: "source cannot be verified because the participant report declares no source ref",
		}
	}
	frictionFinding := ReviewFinding{
		Status: friction,
		Detail: "reviewer classification; see the immutable source report for the participant's description",
	}

	qualification := "NOT_QUALIFYING"
	if strings.TrimSpace(report.TrialID) != "" &&
		strings.TrimSpace(report.ParticipantRef) != "" &&
		externalityFinding.Status == "VERIFIED_EXTERNAL" &&
		sourceFinding.Status == "VERIFIED" &&
		receiptFinding.Status == "STRUCTURE_CONSISTENT" &&
		comprehension == "ACCEPTED" &&
		timeFinding.Status == "WITHIN_TEN_MINUTES" {
		qualification = "QUALIFYING_EXTERNAL"
	}

	frictionEvent := "NOT_VERIFIED"
	if qualification == "QUALIFYING_EXTERNAL" &&
		friction == "ACCEPTED" &&
		concreteFriction(report) {
		frictionEvent = "VERIFIED_EXTERNAL_FRICTION"
	}

	reportDigest := digestBytes(rawReport)
	adjudicationKey := digestBytes([]byte(
		reportDigest + "\x00" + strings.TrimSpace(input.ReviewerRef),
	))
	return FirstContactAdjudication{
		Schema:         firstContactAdjudicationSchema,
		AdjudicationID: "adjudication-" + strings.TrimPrefix(adjudicationKey, "sha256:")[:16],
		AdjudicatedAt:  now.UTC().Format(time.RFC3339Nano),
		ReportDigest:   reportDigest,
		TrialID:        report.TrialID,
		SourceRef:      report.SourceRef,
		ParticipantRef: report.ParticipantRef,
		ReviewerRef:    strings.TrimSpace(input.ReviewerRef),
		Externality:    externalityFinding,
		Source:         sourceFinding,
		Receipt:        receiptFinding,
		Comprehension: ReviewFinding{
			Status: comprehension,
			Detail: "reviewer judgment of whether the participant distinguished interpretation, external authorization evidence, and absence of effect",
		},
		TenMinuteFirstRun:   timeFinding,
		Friction:            frictionFinding,
		QualificationStatus: qualification,
		FrictionEventStatus: frictionEvent,
		AuthorityEffect:     "NONE",
		Notes:               strings.TrimSpace(input.Notes),
		Limits: []string{
			"The report digest binds this record to exact participant-report bytes.",
			"Reviewer identity, independence, externality, source verification, comprehension, and friction acceptance are attestations, not self-proving facts.",
			"Receipt inspection checks declared structure; it does not re-verify the discarded trial fixture private key or prove semantic truth.",
			"This record cannot issue a Permit, perform an effect, elevate canon, or establish Phase 3 adoption.",
		},
	}, nil
}

func writeNewIndentedJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(raw)
	return err
}

func samePath(left, right string) bool {
	leftAbs, leftErr := filepath.Abs(left)
	rightAbs, rightErr := filepath.Abs(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	return strings.EqualFold(filepath.Clean(leftAbs), filepath.Clean(rightAbs))
}

func runAdjudicate(args []string) int {
	flags := flag.NewFlagSet("adjudicate", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	reportPath := flags.String("report", "", "unaltered first-contact report JSON")
	outputPath := flags.String("out", "", "additive adjudication JSON output")
	reviewer := flags.String("reviewer", "", "pseudonymous independent reviewer reference")
	externality := flags.String("externality", "UNDETERMINED", "VERIFIED_EXTERNAL, NOT_EXTERNAL, or UNDETERMINED")
	source := flags.String("source", "UNDETERMINED", "VERIFIED, REJECTED, or UNDETERMINED")
	comprehension := flags.String("comprehension", "UNDETERMINED", "ACCEPTED, REJECTED, or UNDETERMINED")
	friction := flags.String("friction", "UNDETERMINED", "ACCEPTED, REJECTED, NONE, or UNDETERMINED")
	notes := flags.String("notes", "", "optional reviewer notes")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *reportPath == "" || *outputPath == "" || *reviewer == "" {
		fmt.Fprintln(os.Stderr, "adjudicate: -report, -out, and -reviewer are required")
		return 2
	}
	if samePath(*reportPath, *outputPath) {
		fmt.Fprintln(os.Stderr, "adjudicate: -out must not overwrite the participant report")
		return 2
	}

	report, raw, err := readFirstContactReport(*reportPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "adjudicate:", err)
		return 2
	}
	record, err := adjudicateFirstContact(report, raw, AdjudicationInput{
		ReviewerRef:   *reviewer,
		Externality:   *externality,
		Source:        *source,
		Comprehension: *comprehension,
		Friction:      *friction,
		Notes:         *notes,
	}, time.Now().UTC())
	if err != nil {
		fmt.Fprintln(os.Stderr, "adjudicate:", err)
		return 2
	}
	if err := writeNewIndentedJSON(*outputPath, record); err != nil {
		fmt.Fprintln(os.Stderr, "adjudicate write:", err)
		return 2
	}

	fmt.Printf("Wrote additive adjudication: %s\n", *outputPath)
	fmt.Printf("Qualification: %s\n", record.QualificationStatus)
	fmt.Printf("External friction: %s\n", record.FrictionEventStatus)
	fmt.Println("Authority effect: NONE")
	return 0
}

func readAdjudications(dir string) ([]FirstContactAdjudication, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var records []FirstContactAdjudication
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		var probe struct {
			Schema string `json:"schema"`
		}
		if err := json.Unmarshal(raw, &probe); err != nil || probe.Schema != firstContactAdjudicationSchema {
			continue
		}
		record, err := decodeStrictJSON[FirstContactAdjudication](raw)
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", path, err)
		}
		if record.AuthorityEffect != "NONE" {
			return nil, fmt.Errorf("%s claims authority effect %q", path, record.AuthorityEffect)
		}
		records = append(records, record)
	}
	return records, nil
}

func summarizeFirstContact(records []FirstContactAdjudication, dir string, now time.Time) FirstContactSummary {
	uniqueReports := make(map[string]struct{})
	qualifyingParticipants := make(map[string]string)
	frictionReports := make(map[string]string)
	reportIdentity := make(map[string]string)
	reportOutcome := make(map[string]string)
	conflictingReports := make(map[string]struct{})

	for _, record := range records {
		if validSHA256Digest(record.ReportDigest) {
			uniqueReports[record.ReportDigest] = struct{}{}
			identity := record.ParticipantRef + "\x00" + record.TrialID
			if previous, exists := reportIdentity[record.ReportDigest]; exists && previous != identity {
				conflictingReports[record.ReportDigest] = struct{}{}
			} else {
				reportIdentity[record.ReportDigest] = identity
			}
			outcome := fmt.Sprintf(
				"%t\x00%t",
				adjudicationQualifies(record),
				adjudicationHasVerifiedFriction(record),
			)
			if previous, exists := reportOutcome[record.ReportDigest]; exists && previous != outcome {
				conflictingReports[record.ReportDigest] = struct{}{}
			} else {
				reportOutcome[record.ReportDigest] = outcome
			}
		}
	}

	for _, record := range records {
		if _, conflict := conflictingReports[record.ReportDigest]; conflict {
			continue
		}
		if adjudicationQualifies(record) {
			qualifyingParticipants[record.ParticipantRef] = record.TrialID
		}
		if adjudicationHasVerifiedFriction(record) {
			frictionReports[record.ReportDigest] = record.TrialID
		}
	}

	qualifyingIDs := make([]string, 0, len(qualifyingParticipants))
	for _, trialID := range qualifyingParticipants {
		qualifyingIDs = append(qualifyingIDs, trialID)
	}
	frictionIDs := make([]string, 0, len(frictionReports))
	for _, trialID := range frictionReports {
		frictionIDs = append(frictionIDs, trialID)
	}
	sort.Strings(qualifyingIDs)
	sort.Strings(frictionIDs)

	status := "EXTERNAL_EVALUATION_NOT_ESTABLISHED"
	if len(qualifyingParticipants) > 0 {
		status = "EXTERNAL_FEEDBACK_RECORDED"
	}
	return FirstContactSummary{
		Schema:                         firstContactSummarySchema,
		GeneratedAt:                    now.UTC().Format(time.RFC3339Nano),
		InputDirectory:                 dir,
		AdjudicationFilesRead:          len(records),
		UniqueReports:                  len(uniqueReports),
		QualifyingExternalParticipants: len(qualifyingParticipants),
		VerifiedExternalFrictionEvents: len(frictionReports),
		ConflictingRecords:             len(conflictingReports),
		Status:                         status,
		AuthorityEffect:                "NONE",
		QualifyingTrialIDs:             qualifyingIDs,
		FrictionTrialIDs:               frictionIDs,
		Limits: []string{
			"Counts depend on reviewer attestations contained in additive adjudication records.",
			"Distinct qualifying participants are deduplicated by participantRef; distinct friction events are deduplicated by reportDigest.",
			"Qualification is recomputed from findings; conflicting identities for one reportDigest are excluded.",
			"External feedback is optional and is not a prerequisite for engineering work, merge, publication, or authority.",
			"Recorded feedback does not establish adoption, certification, canon elevation, or semantic truth.",
		},
	}
}

func adjudicationQualifies(record FirstContactAdjudication) bool {
	return record.Schema == firstContactAdjudicationSchema &&
		record.AuthorityEffect == "NONE" &&
		validSHA256Digest(record.ReportDigest) &&
		strings.TrimSpace(record.TrialID) != "" &&
		strings.TrimSpace(record.ParticipantRef) != "" &&
		strings.TrimSpace(record.ReviewerRef) != "" &&
		record.Externality.Status == "VERIFIED_EXTERNAL" &&
		record.Source.Status == "VERIFIED" &&
		record.Receipt.Status == "STRUCTURE_CONSISTENT" &&
		record.Comprehension.Status == "ACCEPTED" &&
		record.TenMinuteFirstRun.Status == "WITHIN_TEN_MINUTES"
}

func adjudicationHasVerifiedFriction(record FirstContactAdjudication) bool {
	return adjudicationQualifies(record) &&
		record.Friction.Status == "ACCEPTED" &&
		record.FrictionEventStatus == "VERIFIED_EXTERNAL_FRICTION"
}

func runSummarizeTrials(args []string) int {
	flags := flag.NewFlagSet("summarize-trials", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	dir := flags.String("dir", "", "directory containing additive adjudication JSON files")
	outputPath := flags.String("out", "", "summary JSON output")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *dir == "" || *outputPath == "" {
		fmt.Fprintln(os.Stderr, "summarize-trials: -dir and -out are required")
		return 2
	}

	records, err := readAdjudications(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "summarize-trials:", err)
		return 2
	}
	summary := summarizeFirstContact(records, *dir, time.Now().UTC())
	if err := writeNewIndentedJSON(*outputPath, summary); err != nil {
		fmt.Fprintln(os.Stderr, "summarize-trials write:", err)
		return 2
	}

	fmt.Printf("Wrote First Contact summary: %s\n", *outputPath)
	fmt.Printf("Qualifying external participants recorded: %d\n", summary.QualifyingExternalParticipants)
	fmt.Printf("Verified external friction events recorded: %d\n", summary.VerifiedExternalFrictionEvents)
	fmt.Printf("Status: %s\n", summary.Status)
	fmt.Println("Authority effect: NONE")
	return 0
}
