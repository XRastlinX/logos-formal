package cyonicvalidate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPublishedResultSchemaMatchesRuntimeContract(t *testing.T) {
	schemaPath := filepath.Join("..", "..", "schemas", "cyonic-validation-result.schema.json")
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("result schema is not valid JSON: %v", err)
	}

	required := stringSet(t, schema["required"])
	for _, field := range []string{
		"schema",
		"version",
		"status",
		"validationStatus",
		"validatorVersion",
		"goVersion",
		"goos",
		"goarch",
		"cubedBit",
		"validatorOperator",
		"governanceAuthorityEffect",
		"executionContainment",
		"validationProfile",
		"metadataScope",
		"target",
		"checks",
		"routerDecision",
	} {
		if !required[field] {
			t.Errorf("published result schema does not require %q", field)
		}
	}

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("result schema properties are missing")
	}
	assertConst(t, properties, "cubedBit", "010")
	assertConst(t, properties, "validatorOperator", "000")
	assertConst(t, properties, "governanceAuthorityEffect", "NONE")
	assertConst(t, properties, "executionContainment", "HOST")
}

func TestPublishedManifestSchemaMatchesCompiledRegistry(t *testing.T) {
	schemaPath := filepath.Join("..", "..", "schemas", "cyonic-validation.schema.json")
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("manifest schema is not valid JSON: %v", err)
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("manifest schema properties are missing")
	}
	checks, ok := properties["checks"].(map[string]any)
	if !ok {
		t.Fatal("manifest checks schema is missing")
	}
	items, ok := checks["items"].(map[string]any)
	if !ok {
		t.Fatal("manifest check item schema is missing")
	}
	published := stringSet(t, items["enum"])
	if len(published) != len(requiredCheckOrder) {
		t.Fatalf("published check registry has %d values, want %d", len(published), len(requiredCheckOrder))
	}
	for _, check := range requiredCheckOrder {
		if !published[check] {
			t.Errorf("published check registry is missing %q", check)
		}
	}
}

func stringSet(t *testing.T, value any) map[string]bool {
	t.Helper()
	array, ok := value.([]any)
	if !ok {
		t.Fatalf("value is not a JSON array: %#v", value)
	}
	result := make(map[string]bool, len(array))
	for _, item := range array {
		text, ok := item.(string)
		if !ok {
			t.Fatalf("array item is not a string: %#v", item)
		}
		result[text] = true
	}
	return result
}

func assertConst(t *testing.T, properties map[string]any, field, expected string) {
	t.Helper()
	property, ok := properties[field].(map[string]any)
	if !ok {
		t.Fatalf("schema property %q is missing", field)
	}
	if property["const"] != expected {
		t.Fatalf("schema property %q const = %#v, want %q", field, property["const"], expected)
	}
}
