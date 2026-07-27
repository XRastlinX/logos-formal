package placetime13d

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPublishedRegistryAndGeneratedContractsAgree(t *testing.T) {
	repoRoot := filepath.Join("..", "..")
	if err := VerifyRepositorySources(repoRoot); err != nil {
		t.Fatal(err)
	}
	registry, root, err := LoadRegistry(filepath.Join(repoRoot, "registry", "13D_coordinates.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if root != GeneratedRegistryRoot {
		t.Fatalf("registry root %q, generated %q", root, GeneratedRegistryRoot)
	}
	if len(registry.Dimensions) != 13 {
		t.Fatalf("dimension count %d, want 13", len(registry.Dimensions))
	}
}

func TestGeneratedEventSchemaRequiresAllCoordinatesAndGovernanceFence(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "registry", "event_envelope.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	required := stringsFromAny(t, schema["required"])
	for _, field := range []string{
		"artifactRoot", "parentEvents", "gitBinding", "coordinates", "requestedEffect",
		"authorityProof", "permitRef", "policyRoot", "decision", "authorityEffect",
		"validationReceipts",
	} {
		if !required[field] {
			t.Errorf("event schema does not require %q", field)
		}
	}
	properties := schema["properties"].(map[string]any)
	coordinates := properties["coordinates"].(map[string]any)
	coordinateRequired := stringsFromAny(t, coordinates["required"])
	if len(coordinateRequired) != 13 {
		t.Fatalf("event schema requires %d coordinates, want 13", len(coordinateRequired))
	}
	for _, dimension := range GeneratedDimensions {
		if !coordinateRequired[dimension.ID] {
			t.Errorf("event schema does not require %s", dimension.ID)
		}
	}
}

func stringsFromAny(t *testing.T, value any) map[string]bool {
	t.Helper()
	items, ok := value.([]any)
	if !ok {
		t.Fatalf("not a string array: %#v", value)
	}
	result := map[string]bool{}
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			t.Fatalf("not a string: %#v", item)
		}
		result[text] = true
	}
	return result
}
