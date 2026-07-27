package placetime13d

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitObserverBindsObjectsAndExcludesMetadata(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.name", "Placetime Test")
	runGit(t, repo, "config", "user.email", "placetime@example.invalid")
	if err := os.WriteFile(filepath.Join(repo, "artifact.txt"), []byte("stable artifact\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, ".meta"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".meta", "event.yaml"), []byte("first\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "first")

	observer := GitObserver{}
	first, err := observer.Observe(context.Background(), repo, "HEAD", []string{".meta/"})
	if err != nil {
		t.Fatal(err)
	}
	if first.CommitOID == nil || first.TreeOID == nil {
		t.Fatal("witness binding is missing Git object IDs")
	}
	if err := os.WriteFile(filepath.Join(repo, ".meta", "event.yaml"), []byte("second\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "metadata only")
	second, err := observer.Observe(context.Background(), repo, "HEAD", []string{".meta/"})
	if err != nil {
		t.Fatal(err)
	}
	if first.ArtifactRoot != second.ArtifactRoot {
		t.Fatalf("excluded metadata changed artifact root: %v != %v", first.ArtifactRoot, second.ArtifactRoot)
	}
	if *first.CommitOID == *second.CommitOID || *first.TreeOID == *second.TreeOID {
		t.Fatal("Git object IDs did not change after metadata commit")
	}
}

func TestParentEventsFollowGitParentEnvelopes(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.name", "Placetime Test")
	runGit(t, repo, "config", "user.email", "placetime@example.invalid")
	if err := os.MkdirAll(filepath.Join(repo, "meta"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(repo, "meta", "event.yaml"),
		[]byte("{\"eventId\":\"evt-parent\"}\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "parent")
	parentOID := strings.TrimSpace(runGitOutput(t, repo, "rev-parse", "HEAD"))

	event := EventEnvelope{
		ParentEvents: []string{"evt-parent"},
		GitBinding: GitBinding{
			ParentOIDs: []string{parentOID},
		},
	}
	if err := VerifyParentEvents(context.Background(), repo, event); err != nil {
		t.Fatal(err)
	}
	event.ParentEvents[0] = "evt-wrong"
	if err := VerifyParentEvents(context.Background(), repo, event); err == nil {
		t.Fatal("expected mismatched parent event to be rejected")
	}
}

func runGit(t *testing.T, directory string, arguments ...string) {
	t.Helper()
	command := exec.Command("git", arguments...)
	command.Dir = directory
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", arguments, err, output)
	}
}

func runGitOutput(t *testing.T, directory string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", arguments...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", arguments, err, output)
	}
	return string(output)
}
