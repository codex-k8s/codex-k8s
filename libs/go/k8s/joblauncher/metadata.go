package joblauncher

import "strings"

const (
	metadataLabelManagedBy                = "codex-k8s.dev/managed-by"
	metadataLabelNamespacePurpose         = "codex-k8s.dev/namespace-purpose"
	metadataLabelRuntimeMode              = "codex-k8s.dev/runtime-mode"
	metadataLabelProjectID                = "codex-k8s.dev/project-id"
	metadataLabelRunID                    = "codex-k8s.dev/run-id"
	metadataAnnotationCorrelationID       = "codex-k8s.dev/correlation-id"
	legacyMetadataLabelManagedBy          = "codexk8s.io/managed-by"
	legacyMetadataLabelNamespacePurpose   = "codexk8s.io/namespace-purpose"
	legacyMetadataLabelRuntimeMode        = "codexk8s.io/runtime-mode"
	legacyMetadataLabelProjectID          = "codexk8s.io/project-id"
	legacyMetadataLabelRunID              = "codexk8s.io/run-id"
	legacyMetadataAnnotationCorrelationID = "codexk8s.io/correlation-id"
)

// firstNonEmptyValue returns the first non-empty value found by keys lookup.
func firstNonEmptyValue(values map[string]string, keys ...string) string {
	for _, key := range keys {
		if strings.TrimSpace(values[key]) != "" {
			return values[key]
		}
	}
	return ""
}
