package placetime13d

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const moduleContextsSchema = "urn:logos-formal:placetime-13d-module-contexts:v0.1"

type ModuleContexts struct {
	Schema          string          `json:"schema"`
	Version         string          `json:"version"`
	Status          string          `json:"status"`
	AuthorityEffect string          `json:"authorityEffect"`
	Modules         []ModuleContext `json:"modules"`
}

type ModuleContext struct {
	Path                 string `json:"path"`
	SemanticClass        string `json:"semanticClass"`
	LifecycleState       string `json:"lifecycleState"`
	InstrumentationClass string `json:"instrumentationClass"`
	EntityBinding        struct {
		Namespace  string `json:"namespace"`
		Identifier string `json:"identifier"`
	} `json:"entityBinding"`
}

func VerifyModuleContexts(repoRoot, manifestPath string) error {
	absoluteRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return err
	}
	repoRoot = absoluteRoot
	data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(manifestPath)))
	if err != nil {
		return err
	}
	var manifest ModuleContexts
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("parse module contexts: %w", err)
	}
	if manifest.Schema != moduleContextsSchema || manifest.Version != "0.1.0" ||
		manifest.Status != "PROPOSED" || manifest.AuthorityEffect != "NONE" {
		return fmt.Errorf("module contexts must remain PROPOSED with authorityEffect NONE")
	}

	declared := map[string]bool{}
	for _, module := range manifest.Modules {
		if module.Path == "" || module.SemanticClass == "" ||
			module.InstrumentationClass != "GO_PACKAGE_STATIC" ||
			module.EntityBinding.Namespace != "go-import-path" ||
			module.EntityBinding.Identifier == "" {
			return fmt.Errorf("module context %q is incomplete", module.Path)
		}
		switch module.LifecycleState {
		case "PROPOSED", "IMPLEMENTED", "EXPERIMENTAL", "DEPRECATED":
		default:
			return fmt.Errorf("module context %q has invalid lifecycle state", module.Path)
		}
		if declared[module.Path] {
			return fmt.Errorf("duplicate module context %q", module.Path)
		}
		declared[module.Path] = true
	}

	command := exec.Command("go", "list", "-f", "{{.Dir}}|{{.ImportPath}}", "./...")
	command.Dir = repoRoot
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("list Go packages: %w: %s", err, strings.TrimSpace(string(output)))
	}
	var missing []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, "|", 2)
		if len(fields) != 2 {
			return fmt.Errorf("invalid go list output %q", line)
		}
		relative, err := filepath.Rel(repoRoot, fields[0])
		if err != nil {
			return err
		}
		modulePath := filepath.ToSlash(relative)
		if !declared[modulePath] {
			missing = append(missing, modulePath)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("Go packages missing module contexts: %s", strings.Join(missing, ", "))
	}
	return nil
}
