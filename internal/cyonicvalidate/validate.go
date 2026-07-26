package cyonicvalidate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
)

type CommandRunner interface {
	LookPath(name string) (string, error)
	Run(ctx context.Context, directory, name string, arguments ...string) (string, error)
}

type ExecRunner struct{}

func (ExecRunner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func (ExecRunner) Run(ctx context.Context, directory, name string, arguments ...string) (string, error) {
	command := exec.CommandContext(ctx, name, arguments...)
	command.Dir = directory
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	err := command.Run()
	return output.String(), err
}

type Validator struct {
	Runner CommandRunner
}

func (validator Validator) Validate(
	ctx context.Context,
	repoRoot string,
	target string,
	targetLabel string,
	manifest Manifest,
	metadataScope string,
) (Result, int) {
	result := baseResult()
	result.ValidationProfile = manifest.Profile
	result.MetadataScope = metadataScope
	result.Target = targetLabel
	result.Checks.Total = len(requiredCheckOrder)
	result.Checks.Results = make([]CheckResult, 0, len(requiredCheckOrder))

	beforeRoot, err := HashTree(repoRoot, target)
	if err != nil {
		result.RejectedInvariant = CheckTargetBoundary
		result.Checks.Results = append(result.Checks.Results, failed(CheckTargetBoundary, err.Error()))
		result.DecisionRoot = decisionRoot(result)
		return result, ExitInvalidTarget
	}
	result.TargetRoot = beforeRoot
	result.Checks.Results = append(result.Checks.Results, passed(CheckTargetBoundary, "target resolves inside repository boundary"))

	runner := validator.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	if _, err := runner.LookPath("go"); err != nil {
		result.RejectedInvariant = CheckGoVet
		result.Checks.Results = append(result.Checks.Results, failed(CheckGoVet, "Go toolchain not found"))
		countPassed(&result)
		result.DecisionRoot = decisionRoot(result)
		return result, ExitMissingDependency
	}

	if err := ValidateRequiredPaths(repoRoot, manifest.RequiredPaths); err != nil {
		result.Checks.Results = append(result.Checks.Results, failed(CheckRequiredPaths, err.Error()))
		setFirstRejected(&result, CheckRequiredPaths)
	} else {
		result.Checks.Results = append(result.Checks.Results, passed(CheckRequiredPaths, "all declared structural paths exist inside repository"))
	}

	if output, err := runner.Run(ctx, repoRoot, "go", "vet", "./..."); err != nil {
		result.Checks.Results = append(result.Checks.Results, failed(CheckGoVet, commandFailure(output, err)))
		setFirstRejected(&result, CheckGoVet)
	} else {
		result.Checks.Results = append(result.Checks.Results, passed(CheckGoVet, "go vet ./... passed"))
	}

	if output, err := runner.Run(ctx, repoRoot, "go", "test", "-count=1", "./..."); err != nil {
		result.Checks.Results = append(result.Checks.Results, failed(CheckGoTestFresh, commandFailure(output, err)))
		setFirstRejected(&result, CheckGoTestFresh)
	} else {
		result.Checks.Results = append(result.Checks.Results, passed(CheckGoTestFresh, "go test -count=1 ./... passed"))
	}

	afterRoot, err := HashTree(repoRoot, target)
	if err != nil {
		result.Checks.Results = append(result.Checks.Results, failed(CheckTargetImmutability, err.Error()))
		setFirstRejected(&result, CheckTargetImmutability)
	} else if afterRoot != beforeRoot {
		result.Checks.Results = append(
			result.Checks.Results,
			failed(CheckTargetImmutability, fmt.Sprintf("target root changed from %s to %s", beforeRoot, afterRoot)),
		)
		setFirstRejected(&result, CheckTargetImmutability)
	} else {
		result.Checks.Results = append(result.Checks.Results, passed(CheckTargetImmutability, "target root remained byte-stable during validation"))
	}

	countPassed(&result)
	if result.Checks.Passed == result.Checks.Total {
		result.Status = "VALIDATED"
		result.RouterDecision = "OBSERVE_ONLY"
		result.RejectedInvariant = ""
	}
	result.DecisionRoot = decisionRoot(result)
	if result.Status == "VALIDATED" {
		return result, ExitValidated
	}
	return result, ExitRejected
}

func countPassed(result *Result) {
	result.Checks.Passed = 0
	for _, check := range result.Checks.Results {
		if check.Status == "PASSED" {
			result.Checks.Passed++
		}
	}
}

func passed(id, detail string) CheckResult {
	return CheckResult{ID: id, Status: "PASSED", Detail: detail}
}

func failed(id, detail string) CheckResult {
	return CheckResult{ID: id, Status: "FAILED", Detail: detail}
}

func setFirstRejected(result *Result, invariant string) {
	if result.RejectedInvariant == "" {
		result.RejectedInvariant = invariant
	}
}

func commandFailure(output string, err error) string {
	output = strings.TrimSpace(output)
	const maxDetail = 4096
	if len(output) > maxDetail {
		output = output[:maxDetail] + "...[truncated]"
	}
	if output == "" {
		return err.Error()
	}
	return output
}

func decisionRoot(result Result) string {
	hasher := sha256.New()
	writeHashField(hasher, "logos-formal.validation-decision.v1")
	for _, value := range []string{
		result.Schema,
		result.Version,
		result.ValidatorVersion,
		result.CubedBit,
		result.ValidatorOperator,
		result.AuthorityEffect,
		result.ValidationProfile,
		result.MetadataScope,
		result.Target,
		result.TargetRoot,
		result.Status,
		result.RouterDecision,
		result.RejectedInvariant,
	} {
		writeHashField(hasher, value)
	}
	for _, check := range result.Checks.Results {
		writeHashField(hasher, check.ID)
		writeHashField(hasher, check.Status)
	}
	return "sha256:" + hex.EncodeToString(hasher.Sum(nil))
}
