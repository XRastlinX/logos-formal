package cyonicvalidate

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type fakeRunner struct {
	lookPathError error
	run           func(directory, name string, arguments ...string) (string, error)
}

func (runner fakeRunner) LookPath(string) (string, error) {
	if runner.lookPathError != nil {
		return "", runner.lookPathError
	}
	return "go", nil
}

func (runner fakeRunner) Run(
	_ context.Context,
	directory string,
	name string,
	arguments ...string,
) (string, error) {
	if runner.run == nil {
		return "", nil
	}
	return runner.run(directory, name, arguments...)
}

func TestValidateSuccess(t *testing.T) {
	root := newTestRepository(t)
	manifest := testManifest()

	result, exitCode := (Validator{Runner: fakeRunner{}}).Validate(
		context.Background(),
		root,
		root,
		".",
		manifest,
		"EXACT",
	)

	if exitCode != ExitValidated {
		t.Fatalf("exit code = %d, want %d; result = %#v", exitCode, ExitValidated, result)
	}
	if result.Status != "VALIDATED" ||
		result.ValidationStatus != "COMPLETE" ||
		result.RouterDecision != "OBSERVE_ONLY" {
		t.Fatalf("unexpected successful result: %#v", result)
	}
	if result.GovernanceAuthorityEffect != "NONE" ||
		result.ExecutionContainment != "HOST" ||
		result.CubedBit != "010" ||
		result.ValidatorOperator != "000" {
		t.Fatalf("governance boundary changed: %#v", result)
	}
	if result.Checks.Passed != 5 || result.Checks.Total != 5 {
		t.Fatalf("checks = %d/%d, want 5/5", result.Checks.Passed, result.Checks.Total)
	}
	if !strings.HasPrefix(result.TargetRoot, "sha256:") ||
		!strings.HasPrefix(result.ProfileRoot, "sha256:") ||
		!strings.HasPrefix(result.ValidatorRuntimeRoot, "sha256:") ||
		!strings.HasPrefix(result.DecisionRoot, "sha256:") {
		t.Fatalf("missing content roots: %#v", result)
	}
}

func TestValidateRejectsMutation(t *testing.T) {
	root := newTestRepository(t)
	mutated := false
	runner := fakeRunner{
		run: func(directory, _ string, _ ...string) (string, error) {
			if !mutated {
				mutated = true
				if err := os.WriteFile(filepath.Join(directory, "mutation.txt"), []byte("changed"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			return "", nil
		},
	}

	result, exitCode := (Validator{Runner: runner}).Validate(
		context.Background(),
		root,
		root,
		".",
		testManifest(),
		"EXACT",
	)

	if exitCode != ExitRejected {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitRejected)
	}
	if result.RejectedInvariant != CheckTargetImmutability ||
		result.ValidationStatus != "COMPLETE" ||
		result.RouterDecision != "REJECT" ||
		result.GovernanceAuthorityEffect != "NONE" {
		t.Fatalf("mutation was not rejected correctly: %#v", result)
	}
}

func TestValidateRejectsMissingRequiredControlFile(t *testing.T) {
	root := newTestRepository(t)
	manifest := testManifest()
	manifest.RequiredPaths = append(manifest.RequiredPaths, "AGENTS.md")

	result, exitCode := (Validator{Runner: fakeRunner{}}).Validate(
		context.Background(),
		root,
		root,
		".",
		manifest,
		"EXACT",
	)

	if exitCode != ExitRejected {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitRejected)
	}
	if result.RejectedInvariant != CheckRequiredPaths ||
		result.RouterDecision != "REJECT" ||
		result.GovernanceAuthorityEffect != "NONE" {
		t.Fatalf("missing control file did not fail closed: %#v", result)
	}
}

func TestValidateReportsMissingGoDependency(t *testing.T) {
	root := newTestRepository(t)
	result, exitCode := (Validator{
		Runner: fakeRunner{lookPathError: errors.New("not found")},
	}).Validate(context.Background(), root, root, ".", testManifest(), "EXACT")

	if exitCode != ExitMissingDependency {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitMissingDependency)
	}
	if result.ValidationStatus != "ENVIRONMENT_ERROR" ||
		result.Status != "NO_DECISION" ||
		result.RouterDecision != "NONE" ||
		result.FailureCode != "GO_TOOLCHAIN_UNAVAILABLE" ||
		result.GovernanceAuthorityEffect != "NONE" {
		t.Fatalf("dependency failure escaped governance boundary: %#v", result)
	}
}

func TestResolveBoundPathRejectsTraversal(t *testing.T) {
	root := newTestRepository(t)
	outside := t.TempDir()

	if _, _, err := ResolveBoundPath(root, outside); err == nil {
		t.Fatal("expected path outside repository to be rejected")
	}
}

func TestHashTreeRejectsEscapingSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("ordinary Windows CI users may not have symlink creation privilege")
	}
	root := newTestRepository(t)
	outsideFile := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outsideFile, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsideFile, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}

	if _, err := HashTree(root, root); err == nil || !strings.Contains(err.Error(), "forbidden") {
		t.Fatalf("expected symlink rejection, got %v", err)
	}
}

func TestHashTreeRejectsInternalSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("ordinary Windows CI users may not have symlink creation privilege")
	}
	root := newTestRepository(t)
	if err := os.Symlink("README.md", filepath.Join(root, "alias")); err != nil {
		t.Fatal(err)
	}
	if _, err := HashTree(root, root); err == nil || !strings.Contains(err.Error(), "forbidden") {
		t.Fatalf("expected all v0.1 symlinks to be rejected, got %v", err)
	}
}

func TestLoadManifestRejectsUnknownFieldsAndMissingChecks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.json")
	data := `{
	  "schema":"urn:cyonic:validation-manifest:v1",
	  "version":"1.0",
	  "profile":"test",
	  "target":".",
	  "requiredPaths":[],
	  "checks":[],
	  "command":"rm -rf ."
	}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(path); err == nil {
		t.Fatal("expected manifest with arbitrary command field to be rejected")
	}
}

func TestLoadManifestAcceptsOnlyApprovedCheckRegistry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.json")
	data := `{
	  "schema":"urn:cyonic:validation-manifest:v1",
	  "version":"1.0",
	  "profile":"test",
	  "target":".",
	  "requiredPaths":[],
	  "checks":[
	    "CV-TARGET-BOUNDARY",
	    "CV-REQUIRED-PATHS",
	    "CV-GO-VET",
	    "CV-GO-TEST-FRESH",
	    "CV-TARGET-IMMUTABILITY"
	  ]
	}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(path); err != nil {
		t.Fatalf("approved manifest was rejected: %v", err)
	}
}

func TestLoadPublicProfileRejectsRemovedControlReference(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.json")
	data := `{
	  "schema":"urn:cyonic:validation-manifest:v1",
	  "version":"1.0",
	  "profile":"logos-formal-node-setup-010",
	  "target":".",
	  "requiredPaths":[],
	  "checks":[
	    "CV-TARGET-BOUNDARY",
	    "CV-REQUIRED-PATHS",
	    "CV-GO-VET",
	    "CV-GO-TEST-FRESH",
	    "CV-TARGET-IMMUTABILITY"
	  ]
	}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(path); err == nil ||
		!strings.Contains(err.Error(), `public profile is missing required control path "AGENTS.md"`) {
		t.Fatalf("expected public control-floor rejection, got %v", err)
	}
}

func TestLoadManifestRejectsDuplicateKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.json")
	data := `{
	  "schema":"urn:cyonic:validation-manifest:v1",
	  "version":"1.0",
	  "profile":"first",
	  "profile":"second",
	  "target":".",
	  "requiredPaths":[],
	  "checks":[
	    "CV-TARGET-BOUNDARY",
	    "CV-REQUIRED-PATHS",
	    "CV-GO-VET",
	    "CV-GO-TEST-FRESH",
	    "CV-TARGET-IMMUTABILITY"
	  ]
	}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(path); err == nil || !strings.Contains(err.Error(), `duplicate object key "profile"`) {
		t.Fatalf("expected duplicate key rejection, got %v", err)
	}
}

func TestHashTreeIsStableAndExcludesGitMetadata(t *testing.T) {
	root := newTestRepository(t)
	first, err := HashTree(root, root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "volatile"), []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	second, err := HashTree(root, root)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf(".git metadata changed tree root: %s != %s", first, second)
	}
}

func TestHashTreeRejectsCaseCollisions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the Windows filesystem cannot create this portable-collision fixture")
	}
	root := newTestRepository(t)
	if err := os.WriteFile(filepath.Join(root, "Case.txt"), []byte("one"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "case.txt"), []byte("two"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := HashTree(root, root); err == nil || !strings.Contains(err.Error(), "case-colliding") {
		t.Fatalf("expected case-collision rejection, got %v", err)
	}
}

func TestResultJSONOmitsProposedEpistemicAndPermitFields(t *testing.T) {
	result := baseResult()
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, forbidden := range []string{"generationalDepth", "effectiveUncertainty", "permitId"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("normative result contains proposed or irrelevant field %q: %s", forbidden, text)
		}
	}
}

func TestEnvironmentClearsAmbientGOFLAGS(t *testing.T) {
	t.Setenv("GOFLAGS", "-run=Never")
	found := false
	for _, variable := range environmentWithoutGOFLAGS() {
		if variable == "GOFLAGS=" {
			found = true
		}
		if strings.HasPrefix(variable, "GOFLAGS=") && variable != "GOFLAGS=" {
			t.Fatalf("ambient GOFLAGS leaked into check environment: %q", variable)
		}
	}
	if !found {
		t.Fatal("explicit empty GOFLAGS was not installed")
	}
}

func newTestRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, directory := range []string{".git", "docs", "runtime"} {
		if err := os.MkdirAll(filepath.Join(root, directory), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for path, contents := range map[string]string{
		"go.mod":          "module example.invalid/test\n\ngo 1.24\n",
		"README.md":       "test\n",
		"docs/test.md":    "docs\n",
		"runtime/main.go": "package runtime\n",
	} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(path)), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func testManifest() Manifest {
	return Manifest{
		Schema:        ManifestSchema,
		Version:       "1.0",
		Profile:       "test-010",
		Target:        ".",
		RequiredPaths: []string{"README.md", "go.mod", "docs", "runtime"},
		Checks:        append([]string(nil), requiredCheckOrder...),
	}
}
