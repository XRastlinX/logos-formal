package placetime13d

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Registry struct {
	Schema          string              `json:"schema"`
	Version         string              `json:"version"`
	Status          string              `json:"status"`
	AuthorityEffect string              `json:"authorityEffect"`
	Dimensions      []RegistryDimension `json:"dimensions"`
	Definitions     map[string]any      `json:"$defs"`
}

type RegistryDimension struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	CoordinateClass string          `json:"coordinateClass"`
	Definition      string          `json:"definition"`
	NullMeaning     string          `json:"nullMeaning"`
	ValueSchema     json.RawMessage `json:"valueSchema"`
	DerivedMetrics  []string        `json:"derivedMetrics"`
}

type RepositoryProfile struct {
	Schema            string `json:"schema"`
	Version           string `json:"version"`
	Profile           string `json:"profile"`
	Status            string `json:"status"`
	AuthorityEffect   string `json:"authorityEffect"`
	RegistryPath      string `json:"registryPath"`
	EventSchemaPath   string `json:"eventSchemaPath"`
	ModuleContextPath string `json:"moduleContextPath"`
	GeneratedGoPath   string `json:"generatedGoPath"`
	GitBinding        struct {
		AllowedObjectFormats        []string `json:"allowedObjectFormats"`
		RequireCommitOID            bool     `json:"requireCommitOid"`
		RequireTreeOID              bool     `json:"requireTreeOid"`
		RequireParentOIDs           bool     `json:"requireParentOids"`
		RequireExternalArtifactRoot bool     `json:"requireExternalArtifactRoot"`
		ArtifactRootAlgorithm       string   `json:"artifactRootAlgorithm"`
		ArtifactExcludes            []string `json:"artifactExcludes"`
	} `json:"gitBinding"`
	GovernanceFence struct {
		CoordinatesAuthorize       bool     `json:"coordinatesAuthorize"`
		CertifierReferenceIsPermit bool     `json:"certifierReferenceIsPermit"`
		LineageCommitmentIsPermit  bool     `json:"lineageCommitmentIsPermit"`
		EntityBindingIsPermit      bool     `json:"entityBindingIsPermit"`
		RequiredEnvelopeFields     []string `json:"requiredEnvelopeFields"`
	} `json:"governanceFence"`
}

func LoadRegistry(path string) (Registry, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Registry{}, "", err
	}
	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return Registry{}, "", fmt.Errorf("parse registry: %w", err)
	}
	sum := sha256.Sum256(data)
	root := "sha256:" + hex.EncodeToString(sum[:])
	if err := ValidateRegistry(registry, root); err != nil {
		return Registry{}, "", err
	}
	return registry, root, nil
}

func ValidateRegistry(registry Registry, root string) error {
	if registry.Schema != RegistrySchema || registry.Version != "0.1.0" {
		return fmt.Errorf("unsupported registry schema or version")
	}
	if registry.Status != "PROPOSED" || registry.AuthorityEffect != "NONE" {
		return fmt.Errorf("registry must remain PROPOSED with authorityEffect NONE")
	}
	if len(registry.Dimensions) != 13 {
		return fmt.Errorf("registry contains %d dimensions, want 13", len(registry.Dimensions))
	}
	if root != GeneratedRegistryRoot {
		return fmt.Errorf("generated registry root %s is stale; source root is %s", GeneratedRegistryRoot, root)
	}
	for index, dimension := range registry.Dimensions {
		expectedID := fmt.Sprintf("d%02d", index+1)
		generated := GeneratedDimensions[index]
		if dimension.ID != expectedID {
			return fmt.Errorf("dimension %d id is %q, want %q", index, dimension.ID, expectedID)
		}
		if dimension.ID != generated.ID || dimension.Name != generated.Name || dimension.CoordinateClass != generated.CoordinateClass {
			return fmt.Errorf("%s does not match the generated Go registry", expectedID)
		}
	}
	return nil
}

func LoadRepositoryProfile(path string) (RepositoryProfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RepositoryProfile{}, err
	}
	var profile RepositoryProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return RepositoryProfile{}, fmt.Errorf("parse repository profile: %w", err)
	}
	if profile.Schema != ProfileSchema || profile.Version != "0.1.0" {
		return RepositoryProfile{}, fmt.Errorf("unsupported repository profile")
	}
	if profile.Status != "PROPOSED" || profile.AuthorityEffect != "NONE" {
		return RepositoryProfile{}, fmt.Errorf("repository profile must remain PROPOSED with authorityEffect NONE")
	}
	if profile.GovernanceFence.CoordinatesAuthorize ||
		profile.GovernanceFence.CertifierReferenceIsPermit ||
		profile.GovernanceFence.LineageCommitmentIsPermit ||
		profile.GovernanceFence.EntityBindingIsPermit {
		return RepositoryProfile{}, fmt.Errorf("repository governance fence permits coordinate authority leakage")
	}
	return profile, nil
}

func VerifyRepositorySources(repoRoot string) error {
	profile, err := LoadRepositoryProfile(filepath.Join(repoRoot, "registry", "repository_profile.json"))
	if err != nil {
		return err
	}
	_, root, err := LoadRegistry(filepath.Join(repoRoot, filepath.FromSlash(profile.RegistryPath)))
	if err != nil {
		return err
	}
	schemaData, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(profile.EventSchemaPath)))
	if err != nil {
		return err
	}
	var schema struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		return fmt.Errorf("parse generated event schema: %w", err)
	}
	var registryRootProperty struct {
		Const string `json:"const"`
	}
	if err := json.Unmarshal(schema.Properties["registryRoot"], &registryRootProperty); err != nil {
		return fmt.Errorf("parse event schema registryRoot: %w", err)
	}
	if registryRootProperty.Const != root {
		return fmt.Errorf("generated event schema registry root %q, want %q", registryRootProperty.Const, root)
	}
	return VerifyModuleContexts(repoRoot, profile.ModuleContextPath)
}
