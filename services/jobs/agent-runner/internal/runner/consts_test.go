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

	testCases := []struct {
		name      string
		params    outputSchemaParams
		extraMust []string
		extraMiss []string
	}{
		{
			name: "default profile for dev run",
			params: outputSchemaParams{
				TriggerKind:  "dev",
				TemplateKind: promptTemplateKindWork,
				AgentKey:     "dev",
			},
			extraMiss: []string{
				outputFieldDiagnosis,
				outputFieldActionItems,
				outputFieldEvidenceRefs,
				outputFieldToolGaps,
			},
		},
		{
			name: "self-improve profile includes diagnosis fields",
			params: outputSchemaParams{
				TriggerKind:  "self_improve",
				TemplateKind: promptTemplateKindWork,
				AgentKey:     "km",
			},
			extraMust: []string{
				outputFieldDiagnosis,
				outputFieldActionItems,
				outputFieldEvidenceRefs,
				outputFieldToolGaps,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rawSchema, err := buildOutputSchemaJSON(tc.params)
			if err != nil {
				t.Fatalf("buildOutputSchemaJSON() error = %v", err)
			}

			var schema outputSchemaEnvelope
			if err := json.Unmarshal(rawSchema, &schema); err != nil {
				t.Fatalf("unmarshal output schema: %v", err)
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

			for _, field := range tc.extraMust {
				if _, ok := schema.Properties[field]; !ok {
					t.Fatalf("expected property %q to be present", field)
				}
			}

			for _, field := range tc.extraMiss {
				if _, ok := schema.Properties[field]; ok {
					t.Fatalf("expected property %q to be absent", field)
				}
			}
		})
	}
}
