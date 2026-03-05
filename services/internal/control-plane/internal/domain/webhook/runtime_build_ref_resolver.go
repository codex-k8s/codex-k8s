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
	SHA string `json:"sha"`
}

type rawRunPayloadBuildRef struct {
	PullRequest *normalizedRunPayloadPullRequest `json:"pull_request"`
}

func (s *Service) resolveRuntimeBuildRefForIssueTrigger(ctx context.Context, projectID string, envelope githubWebhookEnvelope, defaultRef string, runtimeMode agentdomain.RuntimeMode) string {
	resolved := strings.TrimSpace(defaultRef)
	if !strings.EqualFold(strings.TrimSpace(string(runtimeMode)), string(agentdomain.RuntimeModeFullEnv)) {
		return resolved
	}

	normalizedProjectID := strings.TrimSpace(projectID)
	repositoryFullName := strings.TrimSpace(envelope.Repository.FullName)
	issueNumber := envelope.Issue.Number
	if s.agentRuns != nil && normalizedProjectID != "" && repositoryFullName != "" && issueNumber > 0 {
		items, err := s.agentRuns.SearchRecentByProjectIssueOrPullRequest(ctx, normalizedProjectID, repositoryFullName, issueNumber, 0, 50)
		if err == nil {
			for _, item := range items {
				runID := strings.TrimSpace(item.RunID)
				if runID == "" {
					continue
				}
				runItem, found, runErr := s.agentRuns.GetByID(ctx, runID)
				if runErr != nil || !found {
					continue
				}
				if ref := extractPullRequestHeadBuildRefFromNormalizedRunPayload(runItem.RunPayload); ref != "" {
					return ref
				}
			}
		}
	}
	if ref := strings.TrimSpace(resolved); ref != "" {
		if resolvedSHA := s.resolveRepositoryRefToSHA(ctx, repositoryFullName, ref); resolvedSHA != "" {
			return resolvedSHA
		}
	}
	return resolved
}

func extractPullRequestHeadBuildRefFromNormalizedRunPayload(runPayload json.RawMessage) string {
	if len(runPayload) == 0 {
		return ""
	}

	var normalized normalizedRunPayloadBuildRef
	if err := json.Unmarshal(runPayload, &normalized); err != nil {
		return ""
	}

	if normalized.PullRequest != nil {
		if sha := strings.TrimSpace(normalized.PullRequest.Head.SHA); sha != "" {
			return sha
		}
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
	if sha := strings.TrimSpace(raw.PullRequest.Head.SHA); sha != "" {
		return sha
	}
	return strings.TrimSpace(raw.PullRequest.Head.Ref)
}

func (s *Service) resolveRepositoryRefToSHA(ctx context.Context, repositoryFullName string, ref string) string {
	repositoryFullName = strings.TrimSpace(repositoryFullName)
	ref = strings.TrimSpace(ref)
	if repositoryFullName == "" || ref == "" {
		return ""
	}
	if s.githubMgmt == nil || strings.TrimSpace(s.githubToken) == "" {
		return ""
	}

	owner, repo, ok := strings.Cut(repositoryFullName, "/")
	if !ok {
		return ""
	}
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	if owner == "" || repo == "" {
		return ""
	}

	sha, err := s.githubMgmt.ResolveRefToCommitSHA(ctx, strings.TrimSpace(s.githubToken), owner, repo, ref)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(sha)
}
