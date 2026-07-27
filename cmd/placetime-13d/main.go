package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/XRastlinX/logos-formal/internal/placetime13d"
)

const (
	exitOK       = 0
	exitRejected = 10
	exitUsage    = 64
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(arguments []string) int {
	if len(arguments) == 0 {
		printUsage()
		return exitUsage
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	switch arguments[0] {
	case "verify-repository":
		repo, err := repositoryArgument(arguments[1:])
		if err != nil {
			return usageError(err)
		}
		if err := verifyRepository(ctx, repo); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitRejected
		}
		fmt.Println("PLACETIME_13D_REPOSITORY_VALID")
		fmt.Println("status: PROPOSED")
		fmt.Println("authorityEffect: NONE")
		return exitOK
	case "validate-event":
		if len(arguments) != 2 {
			return usageError(errors.New("validate-event requires an event file"))
		}
		event, err := placetime13d.LoadEvent(arguments[1])
		if err == nil {
			err = placetime13d.ValidateEvent(event)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitRejected
		}
		fmt.Println("PLACETIME_13D_EVENT_VALID")
		fmt.Println("authorityEffect:", event.AuthorityEffect)
		return exitOK
	case "observe-index":
		repo, err := repositoryArgument(arguments[1:])
		if err != nil {
			return usageError(err)
		}
		profile, err := loadProfile(repo)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitRejected
		}
		binding, err := (placetime13d.GitObserver{}).ObserveIndex(ctx, repo, profile.GitBinding.ArtifactExcludes)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitRejected
		}
		return writeJSON(binding)
	case "bind-index":
		if len(arguments) != 3 {
			return usageError(errors.New("bind-index requires REPOSITORY EVENT_FILE"))
		}
		event, err := bindIndex(ctx, arguments[1], arguments[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitRejected
		}
		return writeJSON(event)
	case "verify-index":
		if len(arguments) != 3 {
			return usageError(errors.New("verify-index requires REPOSITORY EVENT_FILE"))
		}
		if err := verifyIndex(ctx, arguments[1], arguments[2]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitRejected
		}
		fmt.Println("PLACETIME_13D_INDEX_VALID")
		fmt.Println("authorityEffect: NONE")
		return exitOK
	case "verify-message":
		if len(arguments) != 3 {
			return usageError(errors.New("verify-message requires EVENT_FILE COMMIT_MESSAGE_FILE"))
		}
		if err := verifyMessage(arguments[1], arguments[2]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitRejected
		}
		fmt.Println("PLACETIME_13D_MESSAGE_VALID")
		return exitOK
	case "witness-git":
		if len(arguments) != 3 {
			return usageError(errors.New("witness-git requires REPOSITORY EVENT_FILE"))
		}
		if err := witnessGit(ctx, arguments[1], arguments[2]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitRejected
		}
		return exitOK
	case "metrics":
		if len(arguments) != 3 {
			return usageError(errors.New("metrics requires EVENT_FILE METRICS_PROFILE"))
		}
		event, err := placetime13d.LoadEvent(arguments[1])
		if err == nil {
			err = placetime13d.ValidateEvent(event)
		}
		var profile placetime13d.MetricsProfile
		if err == nil {
			profile, err = placetime13d.LoadMetricsProfile(arguments[2])
		}
		var receipt placetime13d.MetricReceipt
		if err == nil {
			receipt, err = placetime13d.ComputeMetrics(event, profile)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitRejected
		}
		return writeJSON(receipt)
	default:
		return usageError(fmt.Errorf("unknown command %q", arguments[0]))
	}
}

func bindIndex(ctx context.Context, repo, eventPath string) (placetime13d.EventEnvelope, error) {
	event, err := placetime13d.LoadEvent(eventPath)
	if err != nil {
		return placetime13d.EventEnvelope{}, err
	}
	profile, err := loadProfile(repo)
	if err != nil {
		return placetime13d.EventEnvelope{}, err
	}
	binding, err := (placetime13d.GitObserver{}).ObserveIndex(ctx, repo, profile.GitBinding.ArtifactExcludes)
	if err != nil {
		return placetime13d.EventEnvelope{}, err
	}
	event.BindingPhase = "PROPOSAL"
	event.ArtifactRoot = binding.ArtifactRoot
	event.GitBinding = binding
	if err := placetime13d.ValidateEvent(event); err != nil {
		return placetime13d.EventEnvelope{}, err
	}
	return event, nil
}

func verifyIndex(ctx context.Context, repo, eventPath string) error {
	event, err := placetime13d.LoadEvent(eventPath)
	if err != nil {
		return err
	}
	if err := placetime13d.ValidateEvent(event); err != nil {
		return err
	}
	bound, err := bindIndex(ctx, repo, eventPath)
	if err != nil {
		return err
	}
	if event.ArtifactRoot != bound.ArtifactRoot ||
		event.GitBinding.ArtifactRoot != bound.GitBinding.ArtifactRoot {
		return fmt.Errorf("tracked proposal artifact root is stale; run placetime-13d bind-index and update .meta/event.yaml")
	}
	if !equalStrings(event.GitBinding.ParentOIDs, bound.GitBinding.ParentOIDs) {
		return fmt.Errorf("tracked proposal parent OIDs are stale")
	}
	return placetime13d.VerifyParentEvents(ctx, repo, event)
}

func verifyMessage(eventPath, messagePath string) error {
	event, err := placetime13d.LoadEvent(eventPath)
	if err != nil {
		return err
	}
	message, err := os.ReadFile(messagePath)
	if err != nil {
		return err
	}
	trailer := "Placetime-Event-ID: " + event.EventID
	if !strings.Contains(string(message), trailer) {
		return fmt.Errorf("commit message must contain trailer %q", trailer)
	}
	return nil
}

func verifyRepository(ctx context.Context, repo string) error {
	repo, err := filepath.Abs(repo)
	if err != nil {
		return err
	}
	if err := placetime13d.VerifyRepositorySources(repo); err != nil {
		return err
	}
	eventPath := filepath.Join(repo, ".meta", "event.yaml")
	event, err := placetime13d.LoadEvent(eventPath)
	if err != nil {
		return fmt.Errorf("load tracked event envelope: %w", err)
	}
	if err := placetime13d.ValidateEvent(event); err != nil {
		return err
	}
	profile, err := loadProfile(repo)
	if err != nil {
		return err
	}
	actual, err := (placetime13d.GitObserver{}).Observe(ctx, repo, "HEAD", profile.GitBinding.ArtifactExcludes)
	if err != nil {
		return err
	}
	if event.ArtifactRoot != actual.ArtifactRoot {
		return fmt.Errorf("event artifact root %s does not match HEAD artifact root %s", event.ArtifactRoot.Value, actual.ArtifactRoot.Value)
	}
	if event.GitBinding.ArtifactRoot != actual.ArtifactRoot {
		return fmt.Errorf("event Git binding artifact root does not match HEAD")
	}
	if !equalStrings(event.GitBinding.ParentOIDs, actual.ParentOIDs) {
		return fmt.Errorf("event parent OIDs do not match HEAD parents")
	}
	if err := placetime13d.VerifyParentEvents(ctx, repo, event); err != nil {
		return err
	}
	message, err := gitCommitMessage(ctx, repo)
	if err != nil {
		return err
	}
	if !strings.Contains(message, "Placetime-Event-ID: "+event.EventID) {
		return fmt.Errorf("HEAD commit message does not contain Placetime-Event-ID trailer for %s", event.EventID)
	}
	return nil
}

func witnessGit(ctx context.Context, repo, eventPath string) error {
	event, err := placetime13d.LoadEvent(eventPath)
	if err != nil {
		return err
	}
	if err := placetime13d.ValidateEvent(event); err != nil {
		return err
	}
	profile, err := loadProfile(repo)
	if err != nil {
		return err
	}
	actual, err := (placetime13d.GitObserver{}).Observe(ctx, repo, "HEAD", profile.GitBinding.ArtifactExcludes)
	if err != nil {
		return err
	}
	if event.ArtifactRoot != actual.ArtifactRoot {
		return fmt.Errorf("proposal artifact root does not match HEAD")
	}
	event.BindingPhase = "WITNESS"
	event.GitBinding = actual
	if err := placetime13d.ValidateEvent(event); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(event)
}

func loadProfile(repo string) (placetime13d.RepositoryProfile, error) {
	return placetime13d.LoadRepositoryProfile(filepath.Join(repo, "registry", "repository_profile.json"))
}

func repositoryArgument(arguments []string) (string, error) {
	if len(arguments) == 0 {
		return ".", nil
	}
	if len(arguments) != 2 || arguments[0] != "--repo" {
		return "", errors.New("expected optional --repo PATH")
	}
	return arguments[1], nil
}

func gitCommitMessage(ctx context.Context, repo string) (string, error) {
	command := exec.CommandContext(ctx, "git", "show", "-s", "--format=%B", "HEAD")
	command.Dir = repo
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("read HEAD commit message: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func writeJSON(value any) int {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitRejected
	}
	return exitOK
}

func usageError(err error) int {
	fmt.Fprintln(os.Stderr, err)
	printUsage()
	return exitUsage
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  placetime-13d verify-repository [--repo PATH]")
	fmt.Fprintln(os.Stderr, "  placetime-13d validate-event EVENT_FILE")
	fmt.Fprintln(os.Stderr, "  placetime-13d observe-index [--repo PATH]")
	fmt.Fprintln(os.Stderr, "  placetime-13d bind-index REPOSITORY EVENT_FILE")
	fmt.Fprintln(os.Stderr, "  placetime-13d verify-index REPOSITORY EVENT_FILE")
	fmt.Fprintln(os.Stderr, "  placetime-13d verify-message EVENT_FILE COMMIT_MESSAGE_FILE")
	fmt.Fprintln(os.Stderr, "  placetime-13d witness-git REPOSITORY EVENT_FILE")
	fmt.Fprintln(os.Stderr, "  placetime-13d metrics EVENT_FILE METRICS_PROFILE")
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
