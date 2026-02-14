package runner

import (
	"encoding/json"
	"testing"
)

type outputSchemaEnvelope struct {
	Properties map[string]json.RawMessage `json:"properties"`
	Required   []string                   `json:"required"`
}

func TestOutputSchemaRequiresAllDeclaredProperties(t *testing.T) {
	t.Parallel()

	var schema outputSchemaEnvelope
	if err := json.Unmarshal([]byte(outputSchemaJSON), &schema); err != nil {
		t.Fatalf("unmarshal outputSchemaJSON: %v", err)
	}

	if len(schema.Properties) == 0 {
		t.Fatal("expected non-empty schema properties")
	}

	requiredSet := make(map[string]struct{}, len(schema.Required))
	for _, name := range schema.Required {
		requiredSet[name] = struct{}{}
	}

	for name := range schema.Properties {
		if _, ok := requiredSet[name]; !ok {
			t.Fatalf("property %q is missing in required list", name)
		}
	}
}
