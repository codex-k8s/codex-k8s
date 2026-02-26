package webhook

import (
	"context"
	"encoding/json"
	"strings"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
)

type normalizedRunPayloadBuildRef struct {
	PullRequest *normalizedRunPayloadPullRequest `json:"pull_request"`
	RawPayload  json.RawMessage                  `json:"raw_payload"`
}

type normalizedRunPayloadPullRequest struct {
	Head normalizedRunPayloadPullRequestRef `json:"head"`
}

type normalizedRunPayloadPullRequestRef struct {
	Ref string `json:"ref"`
}

type rawRunPayloadBuildRef struct {
	PullRequest *normalizedRunPayloadPullRequest `json:"pull_request"`
}

func (s *Service) resolveRuntimeBuildRefForIssueTrigger(ctx context.Context, projectID string, envelope githubWebhookEnvelope, defaultRef string, runtimeMode agentdomain.RuntimeMode) string {
	resolved := strings.TrimSpace(defaultRef)
	if !strings.EqualFold(strings.TrimSpace(string(runtimeMode)), string(agentdomain.RuntimeModeFullEnv)) {
		return resolved
	}
	if s.agentRuns == nil {
		return resolved
	}

	normalizedProjectID := strings.TrimSpace(projectID)
	repositoryFullName := strings.TrimSpace(envelope.Repository.FullName)
	issueNumber := envelope.Issue.Number
	if normalizedProjectID == "" || repositoryFullName == "" || issueNumber <= 0 {
		return resolved
	}

	items, err := s.agentRuns.SearchRecentByProjectIssueOrPullRequest(ctx, normalizedProjectID, repositoryFullName, issueNumber, 0, 50)
	if err != nil {
		return resolved
	}
	for _, item := range items {
		runID := strings.TrimSpace(item.RunID)
		if runID == "" {
			continue
		}
		runItem, found, runErr := s.agentRuns.GetByID(ctx, runID)
		if runErr != nil || !found {
			continue
		}
		if ref := extractPullRequestHeadRefFromNormalizedRunPayload(runItem.RunPayload); ref != "" {
			return ref
		}
	}

	return resolved
}

func extractPullRequestHeadRefFromNormalizedRunPayload(runPayload json.RawMessage) string {
	if len(runPayload) == 0 {
		return ""
	}

	var normalized normalizedRunPayloadBuildRef
	if err := json.Unmarshal(runPayload, &normalized); err != nil {
		return ""
	}

	if normalized.PullRequest != nil {
		if ref := strings.TrimSpace(normalized.PullRequest.Head.Ref); ref != "" {
			return ref
		}
	}

	if len(normalized.RawPayload) == 0 {
		return ""
	}

	var raw rawRunPayloadBuildRef
	if err := json.Unmarshal(normalized.RawPayload, &raw); err != nil {
		return ""
	}
	if raw.PullRequest == nil {
		return ""
	}
	return strings.TrimSpace(raw.PullRequest.Head.Ref)
}
