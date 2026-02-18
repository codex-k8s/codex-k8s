package servicescfg

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
)

var (
	//go:embed schema/services.schema.json
	embeddedServicesSchema []byte

	servicesSchemaOnce sync.Once
	servicesSchema     *jsonschema.Resolved
	servicesSchemaErr  error
)

func validateRenderedSchema(raw []byte) error {
	if len(bytes.TrimSpace(raw)) == 0 {
		return fmt.Errorf("services schema validation failed: rendered yaml is empty")
	}

	doc, err := decodeYAMLToMap(raw)
	if err != nil {
		return fmt.Errorf("services schema validation failed: decode yaml: %w", err)
	}

	resolved, err := resolvedServicesSchema()
	if err != nil {
		return fmt.Errorf("services schema validation failed: load schema: %w", err)
	}
	if err := resolved.Validate(doc); err != nil {
		return fmt.Errorf("services schema validation failed: %w", err)
	}
	return nil
}

func resolvedServicesSchema() (*jsonschema.Resolved, error) {
	servicesSchemaOnce.Do(func() {
		var schema jsonschema.Schema
		if err := json.Unmarshal(embeddedServicesSchema, &schema); err != nil {
			servicesSchemaErr = fmt.Errorf("decode services schema json: %w", err)
			return
		}
		resolved, err := schema.Resolve(nil)
		if err != nil {
			servicesSchemaErr = fmt.Errorf("resolve services schema references: %w", err)
			return
		}
		servicesSchema = resolved
	})
	if servicesSchemaErr != nil {
		return nil, servicesSchemaErr
	}
	if servicesSchema == nil {
		return nil, fmt.Errorf("services schema is not initialized")
	}
	return servicesSchema, nil
}
