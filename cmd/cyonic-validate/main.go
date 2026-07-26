package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/XRastlinX/logos-formal/internal/cyonicvalidate"
)

type options struct {
	target       string
	manifestPath string
	format       string
}

func main() {
	os.Exit(run())
}

func run() int {
	options, err := parseArguments(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return cyonicvalidate.ExitInternalFailure
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "determine working directory:", err)
		return cyonicvalidate.ExitInternalFailure
	}
	repoRoot, err := cyonicvalidate.FindRepositoryRoot(workingDirectory)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return cyonicvalidate.ExitInvalidTarget
	}
	target, targetLabel, err := cyonicvalidate.ResolveBoundPath(repoRoot, options.target)
	if err != nil {
		result := failureResult(targetLabel, "", "INPUT_ERROR", "TARGET_INVALID")
		result.Target = options.target
		_ = cyonicvalidate.WriteResult(os.Stdout, result, options.format)
		fmt.Fprintln(os.Stderr, err)
		return cyonicvalidate.ExitInvalidTarget
	}

	manifestPath := options.manifestPath
	if manifestPath == "" {
		manifestPath = filepath.Join(repoRoot, "cyonic.validation.json")
	}
	manifestPath, err = cyonicvalidate.ResolveBoundFile(repoRoot, manifestPath)
	if err != nil {
		result := failureResult(targetLabel, "", "PROFILE_ERROR", "MANIFEST_UNAVAILABLE")
		_ = cyonicvalidate.WriteResult(os.Stdout, result, options.format)
		fmt.Fprintln(os.Stderr, "manifest path:", err)
		return cyonicvalidate.ExitInvalidManifest
	}
	manifest, err := cyonicvalidate.LoadManifest(manifestPath)
	if err != nil {
		result := failureResult(targetLabel, "", "PROFILE_ERROR", "MANIFEST_INVALID")
		_ = cyonicvalidate.WriteResult(os.Stdout, result, options.format)
		fmt.Fprintln(os.Stderr, err)
		return cyonicvalidate.ExitInvalidManifest
	}

	metadataScope := "UNDECLARED"
	manifestTarget, manifestLabel, resolveErr := cyonicvalidate.ResolveBoundPath(repoRoot, manifest.Target)
	if resolveErr != nil {
		result := failureResult(targetLabel, manifest.Profile, "PROFILE_ERROR", "MANIFEST_TARGET_INVALID")
		_ = cyonicvalidate.WriteResult(os.Stdout, result, options.format)
		fmt.Fprintln(os.Stderr, "manifest target:", resolveErr)
		return cyonicvalidate.ExitInvalidManifest
	}
	if manifestTarget == target && manifestLabel == targetLabel {
		metadataScope = "EXACT"
	}

	contextWithTimeout, cancel := context.WithTimeout(context.Background(), 9*time.Minute)
	defer cancel()
	result, exitCode := (cyonicvalidate.Validator{}).Validate(
		contextWithTimeout,
		repoRoot,
		target,
		targetLabel,
		manifest,
		metadataScope,
	)
	if err := cyonicvalidate.WriteResult(os.Stdout, result, options.format); err != nil {
		fmt.Fprintln(os.Stderr, "write validation result:", err)
		return cyonicvalidate.ExitInternalFailure
	}
	if contextWithTimeout.Err() != nil {
		fmt.Fprintln(os.Stderr, "validation deadline:", contextWithTimeout.Err())
		return cyonicvalidate.ExitInternalFailure
	}
	return exitCode
}

func parseArguments(arguments []string) (options, error) {
	parsed := options{target: ".", format: "text"}
	targetSet := false

	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		switch {
		case argument == "--format":
			index++
			if index >= len(arguments) {
				return options{}, fmt.Errorf("--format requires a value")
			}
			parsed.format = arguments[index]
		case strings.HasPrefix(argument, "--format="):
			parsed.format = strings.TrimPrefix(argument, "--format=")
		case argument == "--manifest":
			index++
			if index >= len(arguments) {
				return options{}, fmt.Errorf("--manifest requires a value")
			}
			parsed.manifestPath = arguments[index]
		case strings.HasPrefix(argument, "--manifest="):
			parsed.manifestPath = strings.TrimPrefix(argument, "--manifest=")
		case strings.HasPrefix(argument, "-"):
			return options{}, fmt.Errorf("unknown option %q", argument)
		default:
			if targetSet {
				return options{}, fmt.Errorf("only one validation target may be supplied")
			}
			parsed.target = argument
			targetSet = true
		}
	}
	if parsed.format != "text" && parsed.format != "json" {
		return options{}, fmt.Errorf("--format must be text or json")
	}
	return parsed, nil
}

func failureResult(target, profile, validationStatus, failureCode string) cyonicvalidate.Result {
	return cyonicvalidate.Result{
		Schema:                    cyonicvalidate.ResultSchema,
		Version:                   "1.0",
		Status:                    "NO_DECISION",
		ValidationStatus:          validationStatus,
		ValidatorVersion:          cyonicvalidate.ValidatorVersion,
		GoVersion:                 runtime.Version(),
		GOOS:                      runtime.GOOS,
		GOARCH:                    runtime.GOARCH,
		CubedBit:                  "010",
		ValidatorOperator:         "000",
		GovernanceAuthorityEffect: "NONE",
		ExecutionContainment:      "HOST",
		ValidationProfile:         profile,
		MetadataScope:             "UNDECLARED",
		Target:                    target,
		RouterDecision:            "NONE",
		FailureCode:               failureCode,
		Checks:                    cyonicvalidate.CheckSummary{Results: []cyonicvalidate.CheckResult{}},
	}
}
