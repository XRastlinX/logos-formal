package cyonicvalidate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func LoadManifest(path string) (Manifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("open validation manifest: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode validation manifest: %w", err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return Manifest{}, err
	}
	if err := validateManifest(manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("validation manifest contains more than one JSON value")
	}
	return fmt.Errorf("decode trailing validation manifest data: %w", err)
}

func validateManifest(manifest Manifest) error {
	if manifest.Schema != ManifestSchema {
		return fmt.Errorf("unsupported manifest schema %q", manifest.Schema)
	}
	if manifest.Version != "1.0" {
		return fmt.Errorf("unsupported manifest version %q", manifest.Version)
	}
	if manifest.Profile == "" {
		return errors.New("manifest profile is required")
	}
	if manifest.Target == "" {
		return errors.New("manifest target is required")
	}

	seenChecks := make(map[string]bool, len(manifest.Checks))
	for _, check := range manifest.Checks {
		if seenChecks[check] {
			return fmt.Errorf("duplicate check identifier %q", check)
		}
		seenChecks[check] = true
	}
	if len(seenChecks) != len(requiredCheckOrder) {
		return fmt.Errorf("manifest must declare exactly %d approved checks", len(requiredCheckOrder))
	}
	for _, check := range requiredCheckOrder {
		if !seenChecks[check] {
			return fmt.Errorf("manifest is missing approved check %q", check)
		}
	}

	seenPaths := make(map[string]bool, len(manifest.RequiredPaths))
	for _, path := range manifest.RequiredPaths {
		if path == "" || filepath.IsAbs(path) {
			return fmt.Errorf("required path %q must be a non-empty relative path", path)
		}
		if strings.Contains(path, `\`) {
			return fmt.Errorf("required path %q must use portable forward slashes", path)
		}
		clean := filepath.Clean(path)
		if clean == "." || clean == ".." || startsWithParent(clean) {
			return fmt.Errorf("required path %q escapes or aliases the repository root", path)
		}
		if seenPaths[clean] {
			return fmt.Errorf("duplicate required path %q", path)
		}
		seenPaths[clean] = true
	}
	return nil
}
