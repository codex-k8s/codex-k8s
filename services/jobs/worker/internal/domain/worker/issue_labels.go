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
	Issue       *githubIssueLabelsIssue       `json:"issue"`
	PullRequest *githubIssueLabelsPullRequest `json:"pull_request"`
}

// githubIssueLabelsIssue keeps issue labels list.
type githubIssueLabelsIssue struct {
	Labels []githubIssueLabelsLabel `json:"labels"`
}

// githubIssueLabelsPullRequest keeps pull request labels list.
type githubIssueLabelsPullRequest struct {
	Labels []githubIssueLabelsLabel `json:"labels"`
}

// githubIssueLabelsLabel keeps a single issue label name.
type githubIssueLabelsLabel struct {
	Name string `json:"name"`
}

// extractIssueLabels returns raw issue label names from GitHub event payload.
func extractIssueLabels(raw json.RawMessage) []string {
	issueLabels, pullRequestLabels := extractIssueAndPullRequestLabels(raw)
	if issueLabels == nil && pullRequestLabels == nil {
		return nil
	}
	labels := make([]string, 0, len(issueLabels)+len(pullRequestLabels))
	labels = append(labels, issueLabels...)
	labels = append(labels, pullRequestLabels...)
	return labels
}

// extractIssueAndPullRequestLabels returns label names split by GitHub payload scope.
func extractIssueAndPullRequestLabels(raw json.RawMessage) (issueLabels []string, pullRequestLabels []string) {
	if len(raw) == 0 {
		return nil, nil
	}
	var event githubIssueLabelsEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, nil
	}
	if event.Issue == nil && event.PullRequest == nil {
		return nil, nil
	}

	issueLabels = make([]string, 0, 8)
	pullRequestLabels = make([]string, 0, 8)
	if event.Issue != nil {
		issueLabels = appendRawLabelNames(issueLabels, event.Issue.Labels)
	}
	if event.PullRequest != nil {
		pullRequestLabels = appendRawLabelNames(pullRequestLabels, event.PullRequest.Labels)
	}
	return issueLabels, pullRequestLabels
}

func appendRawLabelNames(labels []string, source []githubIssueLabelsLabel) []string {
	for _, label := range source {
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
