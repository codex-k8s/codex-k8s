package worker

import (
	"encoding/json"
	"strings"
)

const (
	squareBracketOpen  = "["
	squareBracketClose = "]"
)

// runRawPayloadEnvelope keeps only the raw GitHub payload from agent run payload.
type runRawPayloadEnvelope struct {
	RawPayload json.RawMessage `json:"raw_payload"`
}

// githubIssueLabelsEvent keeps only issue labels used for runtime policy decisions.
type githubIssueLabelsEvent struct {
	Issue *githubIssueLabelsIssue `json:"issue"`
}

// githubIssueLabelsIssue keeps issue labels list.
type githubIssueLabelsIssue struct {
	Labels []githubIssueLabelsLabel `json:"labels"`
}

// githubIssueLabelsLabel keeps a single issue label name.
type githubIssueLabelsLabel struct {
	Name string `json:"name"`
}

// extractIssueLabels returns raw issue label names from GitHub event payload.
func extractIssueLabels(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var event githubIssueLabelsEvent
	if err := json.Unmarshal(raw, &event); err != nil || event.Issue == nil {
		return nil
	}

	labels := make([]string, 0, len(event.Issue.Labels))
	for _, label := range event.Issue.Labels {
		name := strings.TrimSpace(label.Name)
		if name == "" {
			continue
		}
		labels = append(labels, name)
	}
	return labels
}

// extractIssueLabelsFromRunPayload returns issue labels from normalized run payload.
func extractIssueLabelsFromRunPayload(runPayload json.RawMessage) []string {
	if len(runPayload) == 0 {
		return nil
	}
	var envelope runRawPayloadEnvelope
	if err := json.Unmarshal(runPayload, &envelope); err != nil {
		return nil
	}
	return extractIssueLabels(envelope.RawPayload)
}

// hasIssueLabelInRunPayload reports whether run payload issue labels include the target label.
func hasIssueLabelInRunPayload(runPayload json.RawMessage, label string) bool {
	normalizedTarget := normalizeLabelToken(label)
	if normalizedTarget == "" {
		return false
	}

	labels := extractIssueLabelsFromRunPayload(runPayload)
	for _, rawLabel := range labels {
		if normalizeLabelToken(rawLabel) == normalizedTarget {
			return true
		}
	}
	return false
}

// normalizeLabelToken normalizes bracketed and plain label values for policy comparisons.
func normalizeLabelToken(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimPrefix(trimmed, squareBracketOpen)
	trimmed = strings.TrimSuffix(trimmed, squareBracketClose)
	return strings.ToLower(strings.TrimSpace(trimmed))
}
