package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func usage() {
	fmt.Fprintln(os.Stderr, `Cyonic Boundary Service

Read-only verification of an artifact-bound external permit.
The service never issues a permit and never performs an effect.

Usage:
  cyonic-service demo
  cyonic-service serve-demo [-listen 127.0.0.1:8787] [-instance-id <id>]
  cyonic-service serve -trust-key <public.key> -issuer <issuer> [-listen 127.0.0.1:8787] [-instance-id <id>]
  cyonic-service probe-http [-base-url http://127.0.0.1:8787] [-source-ref <ref>] [-expected-instance <id>]
  cyonic-service smoke-http
  cyonic-service trial    -report <trial-report.json> [-origin external]
  cyonic-service adjudicate -report <trial-report.json> -out <adjudication.json> -reviewer <ref> [review options]
  cyonic-service summarize-trials -dir <adjudication-dir> -out <summary.json>
  cyonic-service fixture  -dir <directory>
  cyonic-service evaluate -request <request.json> -trust-key <public.key> -issuer <issuer> [options]
  cyonic-service receipt verify -receipt <decision-receipt.json> -trust-key <receipt-public.key> [options]

Evaluate options:
  -origin internal|external     Caller/operator claim; never self-verifies externality
  -evidence-log <path>         Append the returned receipt as one JSONL record

Exit codes:
  0  requested verification valid; no effect performed
  1  typed boundary rejection
  2  malformed input or service configuration`)
}

func readRequest(path string) (BoundaryRequest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return BoundaryRequest{}, fmt.Errorf("read request: %w", err)
	}
	var request BoundaryRequest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return BoundaryRequest{}, fmt.Errorf("decode request: %w", err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return BoundaryRequest{}, err
	}
	return request, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("request contains multiple JSON values")
		}
		return fmt.Errorf("decode trailing request data: %w", err)
	}
	return nil
}

func printReceipt(receipt BoundaryReceipt) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(receipt)
}

func appendEvidence(path string, receipt BoundaryReceipt) error {
	if path == "" {
		return nil
	}
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create evidence-log directory: %w", err)
		}
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open evidence log: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(receipt); err != nil {
		return fmt.Errorf("append evidence receipt: %w", err)
	}
	return nil
}

func runDemo() int {
	now := time.Now().UTC()
	fixture, err := makeFixture(now)
	if err != nil {
		fmt.Fprintln(os.Stderr, "demo:", err)
		return 2
	}
	receipt, code := evaluate(fixture.Request, EvaluationConfig{
		TrustedIssuer: "example-principal",
		PublicKey:     fixture.PublicKey,
		Now:           func() time.Time { return now },
		OriginClaim:   "internal",
	})
	fmt.Fprintln(os.Stderr, "Demo boundary result: interpretation is structurally validated, external permit evidence is verified, and effect remains NOT_PERFORMED.")
	if err := printReceipt(receipt); err != nil {
		fmt.Fprintln(os.Stderr, "demo output:", err)
		return 2
	}
	return code
}

func runFixture(args []string) int {
	flags := flag.NewFlagSet("fixture", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	dir := flags.String("dir", ".cyonic-fixture", "output directory")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	fixture, err := makeFixture(time.Now().UTC())
	if err != nil {
		fmt.Fprintln(os.Stderr, "fixture:", err)
		return 2
	}
	if err := writeFixture(*dir, fixture); err != nil {
		fmt.Fprintln(os.Stderr, "fixture:", err)
		return 2
	}
	fmt.Printf("Wrote %s and %s\n",
		filepath.Join(*dir, "trusted-public.key"),
		filepath.Join(*dir, "request.json"))
	fmt.Println("The private key was discarded. This fixture is for boundary evaluation only.")
	return 0
}

func runEvaluate(args []string) int {
	flags := flag.NewFlagSet("evaluate", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	requestPath := flags.String("request", "", "request JSON path")
	trustKeyPath := flags.String("trust-key", "", "trusted Ed25519 public key path")
	issuer := flags.String("issuer", "", "trusted issuer identifier")
	origin := flags.String("origin", "internal", "internal or external caller/operator claim")
	evidenceLog := flags.String("evidence-log", "", "optional JSONL evidence log")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *requestPath == "" || *trustKeyPath == "" || *issuer == "" {
		fmt.Fprintln(os.Stderr, "evaluate: -request, -trust-key, and -issuer are required")
		return 2
	}

	request, err := readRequest(*requestPath)
	if err != nil {
		receipt, _ := reject(baseReceipt(BoundaryRequest{}, EvaluationConfig{OriginClaim: *origin}), 2,
			"TRYABILITY", "MALFORMED_REQUEST", err.Error())
		_ = printReceipt(receipt)
		_ = appendEvidence(*evidenceLog, receipt)
		return 2
	}
	publicKey, err := loadPublicKey(*trustKeyPath)
	if err != nil {
		receipt, _ := reject(baseReceipt(request, EvaluationConfig{OriginClaim: *origin}), 2,
			"GOVERNANCE", "INVALID_TRUST_KEY", err.Error())
		_ = printReceipt(receipt)
		_ = appendEvidence(*evidenceLog, receipt)
		return 2
	}

	receipt, code := evaluate(request, EvaluationConfig{
		TrustedIssuer: *issuer,
		PublicKey:     publicKey,
		OriginClaim:   *origin,
	})
	if err := printReceipt(receipt); err != nil {
		fmt.Fprintln(os.Stderr, "evaluate output:", err)
		return 2
	}
	if err := appendEvidence(*evidenceLog, receipt); err != nil {
		fmt.Fprintln(os.Stderr, "evaluate evidence:", err)
		return 2
	}
	return code
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var code int
	switch os.Args[1] {
	case "demo":
		code = runDemo()
	case "serve-demo":
		code = runServeDemo(os.Args[2:])
	case "serve":
		code = runServe(os.Args[2:])
	case "probe-http":
		code = runProbeHTTP(os.Args[2:])
	case "smoke-http":
		code = runSmokeHTTP(os.Args[2:])
	case "trial":
		code = runTrial(os.Args[2:])
	case "adjudicate":
		code = runAdjudicate(os.Args[2:])
	case "summarize-trials":
		code = runSummarizeTrials(os.Args[2:])
	case "fixture":
		code = runFixture(os.Args[2:])
	case "evaluate":
		code = runEvaluate(os.Args[2:])
	case "receipt":
		code = runReceipt(os.Args[2:])
	case "help", "-h", "--help":
		usage()
		code = 0
	default:
		usage()
		fmt.Fprintf(os.Stderr, "\nunknown command %q\n", os.Args[1])
		code = 2
	}
	os.Exit(code)
}
