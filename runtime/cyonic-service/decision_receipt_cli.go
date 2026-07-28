package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

func evaluateDecisionReceiptFile(
	receiptPath string,
	trustKeyPath string,
	expectedSigner string,
) (DecisionReceiptVerification, int) {
	result := baseDecisionReceiptVerification()

	receipt, err := readDecisionReceipt(receiptPath)
	if err != nil {
		result.ReasonCode = "MALFORMED_RECEIPT"
		result.Detail = err.Error()
		return result, 2
	}
	result.ArtifactStatus = "PROPOSED"

	publicKey, err := loadPublicKey(trustKeyPath)
	if err != nil {
		result.ReasonCode = "INVALID_TRUST_KEY"
		result.Detail = err.Error()
		return result, 2
	}
	if expectedSigner = strings.TrimSpace(expectedSigner); expectedSigner != "" &&
		receipt.Proof.KeyID != expectedSigner {
		result.CryptographicStatus = "REJECTED"
		result.RouterDecision = "REJECT"
		result.ReasonCode = "UNEXPECTED_SIGNER"
		result.Detail = fmt.Sprintf("receipt signer %q does not match expected signer %q",
			receipt.Proof.KeyID, expectedSigner)
		return result, 1
	}

	digest, err := verifyDecisionReceipt(receipt, publicKey)
	if err != nil {
		result.CryptographicStatus = "REJECTED"
		result.RouterDecision = "REJECT"
		result.ReasonCode = "RECEIPT_VERIFICATION_FAILED"
		result.Detail = err.Error()
		return result, 1
	}

	result.RecordedOutcome = receipt.Core.Decision.Outcome
	result.RecordedAuthorizationStatus = receipt.Core.AuthorizationBasis.VerificationStatus
	result.SignerKeyID = receipt.Proof.KeyID
	result.CryptographicStatus = "VERIFIED"
	result.RouterDecision = "OBSERVE_ONLY"
	result.ReceiptDigest = digest
	result.ReasonCode = "NONE"
	result.Detail = "The RFC 8785 canonical receipt core is bound to the trusted Ed25519 signer profile; this verifies provenance, not truth, authorization, or an effect."
	return result, 0
}

func printDecisionReceiptVerification(result DecisionReceiptVerification) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func runReceipt(args []string) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprintln(os.Stderr, `Usage:
  cyonic-service receipt verify -receipt <decision-receipt.json> -trust-key <public.key> [-expected-signer sha256:...]

The command verifies an observer decision receipt. It does not issue a Permit,
authorize an effect, or invoke a tool.`)
		if len(args) == 0 {
			return 2
		}
		return 0
	}
	if args[0] != "verify" {
		fmt.Fprintf(os.Stderr, "receipt: unknown subcommand %q\n", args[0])
		return 2
	}

	flags := flag.NewFlagSet("receipt verify", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	receiptPath := flags.String("receipt", "", "decision receipt JSON path")
	trustKeyPath := flags.String("trust-key", "", "trusted Ed25519 receipt-signer public key path")
	expectedSigner := flags.String("expected-signer", "", "optional expected sha256 key ID")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if *receiptPath == "" || *trustKeyPath == "" {
		fmt.Fprintln(os.Stderr, "receipt verify: -receipt and -trust-key are required")
		return 2
	}

	result, code := evaluateDecisionReceiptFile(*receiptPath, *trustKeyPath, *expectedSigner)
	if err := printDecisionReceiptVerification(result); err != nil {
		fmt.Fprintln(os.Stderr, "receipt verify output:", err)
		return 2
	}
	return code
}
