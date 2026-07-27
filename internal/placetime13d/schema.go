package placetime13d

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func ValidateJSONFile(schemaPath, instancePath string) error {
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}
	instanceData, err := os.ReadFile(instancePath)
	if err != nil {
		return err
	}
	schemaDocument, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaData))
	if err != nil {
		return fmt.Errorf("parse JSON Schema %s: %w", schemaPath, err)
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(instanceData))
	if err != nil {
		return fmt.Errorf("parse JSON instance %s: %w", instancePath, err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.AssertFormat()
	const resource = "urn:logos-formal:local-schema"
	if err := compiler.AddResource(resource, schemaDocument); err != nil {
		return fmt.Errorf("load JSON Schema %s: %w", schemaPath, err)
	}
	compiled, err := compiler.Compile(resource)
	if err != nil {
		return fmt.Errorf("compile JSON Schema %s: %w", schemaPath, err)
	}
	if err := compiled.Validate(instance); err != nil {
		return fmt.Errorf("%s violates %s: %w", instancePath, schemaPath, err)
	}
	return nil
}

func ValidateSchemaFile(schemaPath string) error {
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("parse JSON Schema %s: %w", schemaPath, err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	const resource = "urn:logos-formal:schema-under-test"
	if err := compiler.AddResource(resource, document); err != nil {
		return err
	}
	if _, err := compiler.Compile(resource); err != nil {
		return fmt.Errorf("compile JSON Schema %s: %w", schemaPath, err)
	}
	return nil
}

func findEventSchema(eventPath string) (string, error) {
	absolute, err := filepath.Abs(eventPath)
	if err != nil {
		return "", err
	}
	directory := filepath.Dir(absolute)
	for {
		candidate := filepath.Join(directory, "registry", "event_envelope.schema.json")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("cannot locate registry/event_envelope.schema.json above %s", eventPath)
		}
		directory = parent
	}
}
