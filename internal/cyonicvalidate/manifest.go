package cyonicvalidate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxManifestBytes = 1 << 20

func LoadManifest(path string) (Manifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("open validation manifest: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxManifestBytes+1))
	if err != nil {
		return Manifest{}, fmt.Errorf("read validation manifest: %w", err)
	}
	if len(data) > maxManifestBytes {
		return Manifest{}, fmt.Errorf("validation manifest exceeds %d bytes", maxManifestBytes)
	}
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return Manifest{}, err
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
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

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanJSONValue(decoder); err != nil {
		return fmt.Errorf("inspect validation manifest keys: %w", err)
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("validation manifest contains more than one JSON value")
		}
		return fmt.Errorf("inspect validation manifest trailer: %w", err)
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}

	switch delimiter {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("object key is not a string")
			}
			if seen[key] {
				return fmt.Errorf("duplicate object key %q", key)
			}
			seen[key] = true
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil {
			return err
		}
		if closing != json.Delim('}') {
			return errors.New("object does not close with }")
		}
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil {
			return err
		}
		if closing != json.Delim(']') {
			return errors.New("array does not close with ]")
		}
	default:
		return fmt.Errorf("unexpected delimiter %q", delimiter)
	}
	return nil
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
	if manifest.Profile == PublicProfile {
		for _, path := range publicProfileRequiredPaths {
			if !seenPaths[filepath.Clean(path)] {
				return fmt.Errorf("public profile is missing required control path %q", path)
			}
		}
	}
	return nil
}

func ManifestRoot(manifest Manifest) string {
	requiredPaths := append([]string(nil), manifest.RequiredPaths...)
	checks := append([]string(nil), manifest.Checks...)
	sort.Strings(requiredPaths)
	sort.Strings(checks)

	digest := sha256.New()
	writeHashField(digest, "logos-formal.validation-profile.v1")
	for _, value := range []string{
		manifest.Schema,
		manifest.Version,
		manifest.Profile,
		filepath.ToSlash(filepath.Clean(manifest.Target)),
	} {
		writeHashField(digest, value)
	}
	for _, path := range requiredPaths {
		writeHashField(digest, path)
	}
	for _, check := range checks {
		writeHashField(digest, check)
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil))
}
