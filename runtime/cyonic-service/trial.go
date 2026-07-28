package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type TrialArticulation struct {
	ValidatedWhat       string `json:"validatedWhat"`
	AuthorizationOrigin string `json:"authorizationOrigin"`
	EffectPerformed     string `json:"effectPerformed"`
}

type TrialFriction struct {
	Category string `json:"category"`
	Detail   string `json:"detail"`
}

type FirstContactReport struct {
	Schema                       string            `json:"schema"`
	TrialID                      string            `json:"trialId"`
	RecordedAt                   string            `json:"recordedAt"`
	SourceRef                    string            `json:"sourceRef"`
	ParticipantRef               string            `json:"participantRef"`
	RelationshipClaim            string            `json:"relationshipClaim"`
	ExternalityStatus            string            `json:"externalityStatus"`
	CommandElapsedSeconds        float64           `json:"commandElapsedSeconds"`
	SelfReportedMinutesFromClone *float64          `json:"selfReportedMinutesFromClone,omitempty"`
	BoundaryReceipt              BoundaryReceipt   `json:"boundaryReceipt"`
	Articulation                 TrialArticulation `json:"articulation"`
	Friction                     TrialFriction     `json:"friction"`
	AdjudicationStatus           string            `json:"adjudicationStatus"`
	AuthorityEffect              string            `json:"authorityEffect"`
	Limits                       []string          `json:"limits"`
}

func prompt(reader *bufio.Reader, writer io.Writer, label string) (string, error) {
	fmt.Fprint(writer, label)
	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(strings.TrimPrefix(value, "\ufeff")), nil
}

func normalizeTrialCategory(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "NONE", "LEGIBILITY", "TRYABILITY", "INTEGRATION", "EVIDENCE",
		"GOVERNANCE", "PRODUCT", "SCOPE_MISMATCH":
		return strings.ToUpper(strings.TrimSpace(value))
	default:
		return "UNCLASSIFIED"
	}
}

func detectSourceRef() string {
	command := exec.Command("git", "rev-parse", "HEAD")
	output, err := command.Output()
	if err != nil {
		return "UNDECLARED"
	}
	value := strings.TrimSpace(string(output))
	if value == "" {
		return "UNDECLARED"
	}
	status := exec.Command("git", "status", "--porcelain")
	if dirty, statusErr := status.Output(); statusErr == nil && len(dirty) != 0 {
		return value + "+dirty"
	}
	return value
}

func conductTrial(
	input io.Reader,
	output io.Writer,
	now time.Time,
	origin string,
	sourceRef string,
) (FirstContactReport, error) {
	started := time.Now()
	fixture, err := makeFixture(now)
	if err != nil {
		return FirstContactReport{}, err
	}
	receipt, code := evaluate(fixture.Request, EvaluationConfig{
		TrustedIssuer: "example-principal",
		PublicKey:     fixture.PublicKey,
		Now:           func() time.Time { return now },
		OriginClaim:   origin,
	})
	if code != 0 {
		return FirstContactReport{}, fmt.Errorf("internal trial fixture rejected with exit code %d", code)
	}

	fmt.Fprintln(output, "Cyonic Boundary result")
	fmt.Fprintf(output, "  interpretation: %s\n", receipt.Interpretation.Status)
	fmt.Fprintf(output, "  authorization: %s\n", receipt.Authorization.Status)
	fmt.Fprintf(output, "  routing: %s (forwarded=%t)\n", receipt.Routing.Decision, receipt.Routing.Forwarded)
	fmt.Fprintf(output, "  effect: %s\n", receipt.Effect.Status)
	fmt.Fprintf(output, "  authority effect: %s\n\n", receipt.AuthorityEffect)
	fmt.Fprintln(output, "Please answer from what you just observed. There are no required keywords.")

	reader := bufio.NewReader(input)
	participantRef, err := prompt(reader, output, "Pseudonymous participant reference: ")
	if err != nil {
		return FirstContactReport{}, err
	}
	minutesRaw, err := prompt(reader, output, "Approximate minutes since you began cloning/setup: ")
	if err != nil {
		return FirstContactReport{}, err
	}
	validatedWhat, err := prompt(reader, output, "In your own words, what did the service validate? ")
	if err != nil {
		return FirstContactReport{}, err
	}
	authorizationOrigin, err := prompt(reader, output, "Where did the authorization originate? ")
	if err != nil {
		return FirstContactReport{}, err
	}
	effectPerformed, err := prompt(reader, output, "Did the service perform or forward the proposed effect? ")
	if err != nil {
		return FirstContactReport{}, err
	}
	frictionCategory, err := prompt(reader, output,
		"Friction category [NONE, LEGIBILITY, TRYABILITY, INTEGRATION, EVIDENCE, GOVERNANCE, PRODUCT, SCOPE_MISMATCH]: ")
	if err != nil {
		return FirstContactReport{}, err
	}
	frictionDetail, err := prompt(reader, output, "Briefly describe the friction or write none: ")
	if err != nil {
		return FirstContactReport{}, err
	}

	var minutes *float64
	if parsed, parseErr := strconv.ParseFloat(minutesRaw, 64); parseErr == nil && parsed >= 0 {
		minutes = &parsed
	}

	relationshipClaim := strings.ToUpper(strings.TrimSpace(origin))
	externalityStatus := "INTERNAL"
	if relationshipClaim == "EXTERNAL" {
		externalityStatus = "CLAIMED_EXTERNAL"
	} else if relationshipClaim != "INTERNAL" {
		externalityStatus = "UNDETERMINED"
	}

	trialID := "trial-" + now.UTC().Format("20060102T150405.000000000Z")
	return FirstContactReport{
		Schema:                       firstContactReportSchema,
		TrialID:                      trialID,
		RecordedAt:                   now.UTC().Format(time.RFC3339Nano),
		SourceRef:                    sourceRef,
		ParticipantRef:               participantRef,
		RelationshipClaim:            relationshipClaim,
		ExternalityStatus:            externalityStatus,
		CommandElapsedSeconds:        time.Since(started).Seconds(),
		SelfReportedMinutesFromClone: minutes,
		BoundaryReceipt:              receipt,
		Articulation: TrialArticulation{
			ValidatedWhat:       validatedWhat,
			AuthorizationOrigin: authorizationOrigin,
			EffectPerformed:     effectPerformed,
		},
		Friction: TrialFriction{
			Category: normalizeTrialCategory(frictionCategory),
			Detail:   frictionDetail,
		},
		AdjudicationStatus: "PENDING_EXTERNAL_REVIEW",
		AuthorityEffect:    "NONE",
		Limits: []string{
			"Participant relationship and setup time are self-reported.",
			"The service does not score the participant's explanation.",
			"Only an independent reviewer may validate externality, comprehension, or friction.",
			"This report is evidence input, not adoption, certification, or authority.",
		},
	}, nil
}

func runTrial(args []string) int {
	flags := flag.NewFlagSet("trial", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	reportPath := flags.String("report", "", "output JSON report path")
	origin := flags.String("origin", "external", "external or internal relationship claim")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *reportPath == "" {
		fmt.Fprintln(os.Stderr, "trial: -report is required")
		return 2
	}
	if err := requireNewReportPath(*reportPath); err != nil {
		fmt.Fprintln(os.Stderr, "trial:", err)
		return 2
	}

	report, err := conductTrial(os.Stdin, os.Stdout, time.Now().UTC(), *origin, detectSourceRef())
	if err != nil {
		fmt.Fprintln(os.Stderr, "trial:", err)
		return 2
	}
	if err := writeFirstContactReport(*reportPath, report); err != nil {
		fmt.Fprintln(os.Stderr, "trial write:", err)
		return 2
	}

	fmt.Fprintf(os.Stdout, "\nWrote trial report: %s\n", *reportPath)
	fmt.Fprintln(os.Stdout, "Status remains PENDING_EXTERNAL_REVIEW.")
	return 0
}

func writeFirstContactReport(path string, report FirstContactReport) error {
	return writeNewIndentedJSON(path, report)
}

func requireNewReportPath(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return fmt.Errorf("report path already exists; choose a new filename: %s", path)
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("inspect report path: %w", err)
	}
	return nil
}
