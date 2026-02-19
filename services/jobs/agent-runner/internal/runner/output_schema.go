package runner

import (
	"encoding/json"
	"fmt"
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

type outputSchemaProfile string

const (
	outputSchemaProfilePRFlow      outputSchemaProfile = "pr_flow"
	outputSchemaProfileSelfImprove outputSchemaProfile = "self_improve"
	outputSchemaTypeObject         string              = "object"
	outputSchemaTypeArray          string              = "array"
	outputSchemaTypeString         string              = "string"
	outputSchemaTypeInteger        string              = "integer"
	outputSchemaMinPRNumber        int                 = 1
	outputSchemaMinPRURLLength     int                 = 1
)

const (
	outputFieldSummary         = "summary"
	outputFieldBranch          = "branch"
	outputFieldPRNumber        = "pr_number"
	outputFieldPRURL           = "pr_url"
	outputFieldSessionID       = "session_id"
	outputFieldModel           = "model"
	outputFieldReasoningEffort = "reasoning_effort"
	outputFieldDiagnosis       = "diagnosis"
	outputFieldActionItems     = "action_items"
	outputFieldEvidenceRefs    = "evidence_refs"
	outputFieldToolGaps        = "tool_gaps"
)

type outputSchemaParams struct {
	TriggerKind  string
	TemplateKind string
	AgentKey     string
}

type outputSchemaDocument struct {
	Type                 string                          `json:"type"`
	Properties           map[string]outputSchemaProperty `json:"properties"`
	Required             []string                        `json:"required"`
	AdditionalProperties bool                            `json:"additionalProperties"`
}

type outputSchemaProperty struct {
	Type      string            `json:"type"`
	Minimum   *int              `json:"minimum,omitempty"`
	MinLength *int              `json:"minLength,omitempty"`
	Items     *outputSchemaItem `json:"items,omitempty"`
}

type outputSchemaItem struct {
	Type string `json:"type"`
}

func buildOutputSchemaJSON(params outputSchemaParams) ([]byte, error) {
	profile := resolveOutputSchemaProfile(params)
	schema := newBaseOutputSchema()
	if profile == outputSchemaProfileSelfImprove {
		addSelfImproveOutputFields(schema)
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("marshal output schema: %w", err)
	}
	return raw, nil
}

func resolveOutputSchemaProfile(params outputSchemaParams) outputSchemaProfile {
	normalizedTrigger := webhookdomain.NormalizeTriggerKind(params.TriggerKind)
	normalizedTemplate := normalizeTemplateKind(params.TemplateKind, string(normalizedTrigger))
	normalizedAgentKey := strings.TrimSpace(params.AgentKey)

	if normalizedTrigger == webhookdomain.TriggerKindSelfImprove {
		return outputSchemaProfileSelfImprove
	}

	if normalizedTemplate == promptTemplateKindRevise && normalizedAgentKey != "" {
		return outputSchemaProfilePRFlow
	}

	return outputSchemaProfilePRFlow
}

func newBaseOutputSchema() *outputSchemaDocument {
	schema := &outputSchemaDocument{
		Type:                 outputSchemaTypeObject,
		Properties:           make(map[string]outputSchemaProperty),
		Required:             make([]string, 0, 8),
		AdditionalProperties: false,
	}

	addOutputField(schema, outputFieldSummary, outputSchemaProperty{Type: outputSchemaTypeString})
	addOutputField(schema, outputFieldBranch, outputSchemaProperty{Type: outputSchemaTypeString})
	addOutputField(schema, outputFieldPRNumber, outputSchemaProperty{Type: outputSchemaTypeInteger, Minimum: intPtr(outputSchemaMinPRNumber)})
	addOutputField(schema, outputFieldPRURL, outputSchemaProperty{Type: outputSchemaTypeString, MinLength: intPtr(outputSchemaMinPRURLLength)})
	addOutputField(schema, outputFieldSessionID, outputSchemaProperty{Type: outputSchemaTypeString})
	addOutputField(schema, outputFieldModel, outputSchemaProperty{Type: outputSchemaTypeString})
	addOutputField(schema, outputFieldReasoningEffort, outputSchemaProperty{Type: outputSchemaTypeString})

	return schema
}

func addSelfImproveOutputFields(schema *outputSchemaDocument) {
	addOutputField(schema, outputFieldDiagnosis, outputSchemaProperty{Type: outputSchemaTypeString})
	addOutputField(schema, outputFieldActionItems, outputSchemaProperty{
		Type:  outputSchemaTypeArray,
		Items: &outputSchemaItem{Type: outputSchemaTypeString},
	})
	addOutputField(schema, outputFieldEvidenceRefs, outputSchemaProperty{
		Type:  outputSchemaTypeArray,
		Items: &outputSchemaItem{Type: outputSchemaTypeString},
	})
	addOutputField(schema, outputFieldToolGaps, outputSchemaProperty{
		Type:  outputSchemaTypeArray,
		Items: &outputSchemaItem{Type: outputSchemaTypeString},
	})
}

func addOutputField(schema *outputSchemaDocument, fieldName string, property outputSchemaProperty) {
	schema.Properties[fieldName] = property
	schema.Required = append(schema.Required, fieldName)
}

func intPtr(value int) *int {
	return &value
}
