package webhook

import (
	"context"
	"encoding/json"
	"strings"

	agentdomain "github.com/codex-k8s/kodex/libs/go/domain/agent"
	webhookdomain "github.com/codex-k8s/kodex/libs/go/domain/webhook"
	agentrunrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/agentrun"
	runstatusdomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/runstatus"
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
	Runtime     *rawRunPayloadRuntimeBuildRef    `json:"runtime"`
}

type rawRunPayloadRuntimeBuildRef struct {
	Namespace     string `json:"namespace"`
	BuildRef      string `json:"build_ref"`
	AccessProfile string `json:"access_profile"`
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
					if resolvedSHA := s.resolveRepositoryRefToSHA(ctx, repositoryFullName, ref); resolvedSHA != "" {
						return resolvedSHA
					}
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
		if raw.Runtime != nil {
			if buildRef := strings.TrimSpace(raw.Runtime.BuildRef); buildRef != "" {
				return buildRef
			}
		}
		return ""
	}
	if sha := strings.TrimSpace(raw.PullRequest.Head.SHA); sha != "" {
		return sha
	}
	return strings.TrimSpace(raw.PullRequest.Head.Ref)
}

func extractRuntimeNamespaceFromNormalizedRunPayload(runPayload json.RawMessage) string {
	if len(runPayload) == 0 {
		return ""
	}

	var normalized struct {
		Runtime *rawRunPayloadRuntimeBuildRef `json:"runtime"`
	}
	if err := json.Unmarshal(runPayload, &normalized); err != nil {
		return ""
	}
	if normalized.Runtime == nil {
		return ""
	}
	return strings.TrimSpace(normalized.Runtime.Namespace)
}

type issueTriggerRuntimeProfile struct {
	TargetEnv       string
	Namespace       string
	BuildRef        string
	AccessProfile   agentdomain.RuntimeAccessProfile
	WarningReason   string
	SuggestedLabels []string
}

type candidateRuntimeIdentity struct {
	Namespace string
	BuildRef  string
}

func (s *Service) resolveIssueTriggerRuntimeProfile(
	ctx context.Context,
	projectID string,
	envelope githubWebhookEnvelope,
	trigger issueRunTrigger,
	defaultRef string,
	runtimeMode agentdomain.RuntimeMode,
) issueTriggerRuntimeProfile {
	result := issueTriggerRuntimeProfile{
		BuildRef:      strings.TrimSpace(defaultRef),
		AccessProfile: agentdomain.RuntimeAccessProfileCandidate,
	}
	if !strings.EqualFold(strings.TrimSpace(string(runtimeMode)), string(agentdomain.RuntimeModeFullEnv)) {
		return result
	}

	switch webhookdomain.NormalizeTriggerKind(string(trigger.Kind)) {
	case webhookdomain.TriggerKindDev, webhookdomain.TriggerKindDevRevise:
		result.TargetEnv = "ai"
		if identity, ok := s.resolveIssueTriggerCandidateRuntime(ctx, projectID, envelope); ok {
			if identity.Namespace != "" {
				result.Namespace = identity.Namespace
			}
			if identity.BuildRef != "" {
				result.BuildRef = identity.BuildRef
			}
		}
		return result
	case webhookdomain.TriggerKindQA,
		webhookdomain.TriggerKindQARevise,
		webhookdomain.TriggerKindRelease,
		webhookdomain.TriggerKindReleaseRevise:
		result.TargetEnv = "ai"
		identity, ok := s.resolveIssueTriggerCandidateRuntime(ctx, projectID, envelope)
		if !ok || strings.TrimSpace(identity.Namespace) == "" || strings.TrimSpace(identity.BuildRef) == "" {
			result.WarningReason = string(runstatusdomain.TriggerWarningReasonIssueTriggerCandidateNotFound)
			result.SuggestedLabels = []string{
				webhookdomain.DefaultRunDevLabel,
				webhookdomain.DefaultRunDevReviseLabel,
			}
			return result
		}
		result.Namespace = identity.Namespace
		result.BuildRef = identity.BuildRef
		return result
	case webhookdomain.TriggerKindPostDeploy,
		webhookdomain.TriggerKindPostDeployRevise,
		webhookdomain.TriggerKindOps,
		webhookdomain.TriggerKindOpsRevise:
		result.TargetEnv = "production"
		result.Namespace = strings.TrimSpace(s.platformNamespace)
		result.AccessProfile = agentdomain.RuntimeAccessProfileProductionReadOnly
		if resolvedSHA := s.resolveRepositoryRefToSHA(ctx, strings.TrimSpace(envelope.Repository.FullName), strings.TrimSpace(defaultRef)); resolvedSHA != "" {
			result.BuildRef = resolvedSHA
		}
		return result
	default:
		return result
	}
}

func (s *Service) resolveIssueTriggerCandidateRuntime(ctx context.Context, projectID string, envelope githubWebhookEnvelope) (candidateRuntimeIdentity, bool) {
	normalizedProjectID := strings.TrimSpace(projectID)
	repositoryFullName := strings.TrimSpace(envelope.Repository.FullName)
	issueNumber := envelope.Issue.Number
	if s.agentRuns == nil || normalizedProjectID == "" || repositoryFullName == "" || issueNumber <= 0 {
		return candidateRuntimeIdentity{}, false
	}

	items, err := s.agentRuns.SearchRecentByProjectIssueOrPullRequest(ctx, normalizedProjectID, repositoryFullName, issueNumber, 0, 50)
	if err != nil {
		return candidateRuntimeIdentity{}, false
	}

	owner, repo, ok := strings.Cut(repositoryFullName, "/")
	if !ok {
		return candidateRuntimeIdentity{}, false
	}

	for _, item := range items {
		identity, found := s.resolveCandidateRuntimeFromRunLookup(ctx, owner, repo, item)
		if found {
			return identity, true
		}
	}
	return candidateRuntimeIdentity{}, false
}

func (s *Service) resolveCandidateRuntimeFromRunLookup(ctx context.Context, owner string, repo string, item agentrunrepo.RunLookupItem) (candidateRuntimeIdentity, bool) {
	runID := strings.TrimSpace(item.RunID)
	if runID == "" {
		return candidateRuntimeIdentity{}, false
	}

	identity := candidateRuntimeIdentity{}
	if s.runStatus != nil {
		runtimeState, err := s.runStatus.GetRunRuntimeState(ctx, runID)
		if err == nil {
			identity.Namespace = strings.TrimSpace(runtimeState.Namespace)
		}
	}

	runItem, found, runErr := s.agentRuns.GetByID(ctx, runID)
	if runErr != nil || !found {
		return candidateRuntimeIdentity{}, false
	}
	if identity.Namespace == "" {
		identity.Namespace = extractRuntimeNamespaceFromNormalizedRunPayload(runItem.RunPayload)
	}
	if buildRef := extractPullRequestHeadBuildRefFromNormalizedRunPayload(runItem.RunPayload); buildRef != "" {
		identity.BuildRef = buildRef
	}

	prNumber := int(item.PullRequestNumber)
	if prNumber <= 0 {
		return candidateRuntimeIdentity{}, identity.Namespace != "" && identity.BuildRef != ""
	}
	prHead, ok := s.resolvePullRequestHead(ctx, strings.TrimSpace(owner), strings.TrimSpace(repo), prNumber)
	if !ok {
		return candidateRuntimeIdentity{}, false
	}
	if prHead.State != "" && !strings.EqualFold(prHead.State, "open") {
		return candidateRuntimeIdentity{}, false
	}
	if strings.TrimSpace(prHead.HeadSHA) != "" {
		identity.BuildRef = strings.TrimSpace(prHead.HeadSHA)
	} else if strings.TrimSpace(prHead.HeadRef) != "" {
		identity.BuildRef = strings.TrimSpace(prHead.HeadRef)
	} else {
		return candidateRuntimeIdentity{}, false
	}

	return identity, identity.Namespace != "" && identity.BuildRef != ""
}

func (s *Service) resolvePullRequestHead(ctx context.Context, owner string, repo string, number int) (GitHubPullRequestHeadDetails, bool) {
	if number <= 0 || s.githubMgmt == nil || strings.TrimSpace(s.githubToken) == "" {
		return GitHubPullRequestHeadDetails{}, false
	}
	result, err := s.githubMgmt.GetPullRequestHead(ctx, strings.TrimSpace(s.githubToken), owner, repo, number)
	if err != nil {
		return GitHubPullRequestHeadDetails{}, false
	}
	return result, true
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
