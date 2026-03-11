package staff

import (
	"context"
	"fmt"
	"slices"
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	repoprovider "github.com/codex-k8s/codex-k8s/libs/go/repo/provider"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

var knownNextStepPRLabels = map[string]struct{}{
	webhookdomain.DefaultNeedReviewerLabel: {},
}

// PreviewNextStepAction returns label diff preview without mutating GitHub state.
func (s *Service) PreviewNextStepAction(ctx context.Context, principal Principal, params querytypes.NextStepActionParams) (querytypes.NextStepActionResult, error) {
	return s.resolveNextStepAction(ctx, principal, params, false)
}

// ExecuteNextStepAction applies one next-step action on GitHub labels.
func (s *Service) ExecuteNextStepAction(ctx context.Context, principal Principal, params querytypes.NextStepActionParams) (querytypes.NextStepActionResult, error) {
	return s.resolveNextStepAction(ctx, principal, params, true)
}

func (s *Service) resolveNextStepAction(ctx context.Context, principal Principal, params querytypes.NextStepActionParams, apply bool) (querytypes.NextStepActionResult, error) {
	if !principal.IsPlatformAdmin {
		return querytypes.NextStepActionResult{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.githubMgmt == nil {
		return querytypes.NextStepActionResult{}, fmt.Errorf("failed_precondition: github management client is not configured")
	}

	repositoryFullName := strings.TrimSpace(params.RepositoryFullName)
	if repositoryFullName == "" {
		return querytypes.NextStepActionResult{}, errs.Validation{Field: "repository_full_name", Msg: "is required"}
	}
	actionKind := strings.ToLower(strings.TrimSpace(params.ActionKind))
	if actionKind == "" {
		return querytypes.NextStepActionResult{}, errs.Validation{Field: "action_kind", Msg: "is required"}
	}
	targetLabel := strings.ToLower(strings.TrimSpace(params.TargetLabel))
	if targetLabel == "" {
		return querytypes.NextStepActionResult{}, errs.Validation{Field: "target_label", Msg: "is required"}
	}

	owner, repo, err := parseGitHubFullName(repositoryFullName)
	if err != nil {
		return querytypes.NextStepActionResult{}, errs.Validation{Field: "repository_full_name", Msg: err.Error()}
	}

	binding, ok, err := s.repos.FindByProviderOwnerName(ctx, string(repoprovider.ProviderGitHub), owner, repo)
	if err != nil {
		return querytypes.NextStepActionResult{}, err
	}
	if !ok {
		return querytypes.NextStepActionResult{}, errs.Validation{Field: "repository_full_name", Msg: "repository is not bound to any project"}
	}

	_, botToken, _, _, err := s.resolveEffectiveGitHubTokens(ctx, binding.ProjectID, binding.RepositoryID)
	if err != nil {
		return querytypes.NextStepActionResult{}, err
	}

	switch actionKind {
	case querytypes.NextStepActionKindIssueStageTransition:
		return s.previewOrExecuteIssueStageTransition(ctx, botToken, owner, repo, params.IssueNumber, targetLabel, apply)
	case querytypes.NextStepActionKindPullRequestLabelAdd:
		return s.previewOrExecutePullRequestLabelAdd(ctx, botToken, owner, repo, params.PullRequestNumber, targetLabel, apply)
	default:
		return querytypes.NextStepActionResult{}, errs.Validation{Field: "action_kind", Msg: "must be a known next-step action"}
	}
}

func (s *Service) previewOrExecuteIssueStageTransition(ctx context.Context, botToken string, owner string, repo string, issueNumber int, targetLabel string, apply bool) (querytypes.NextStepActionResult, error) {
	if issueNumber <= 0 {
		return querytypes.NextStepActionResult{}, errs.Validation{Field: "issue_number", Msg: "must be positive"}
	}
	if !s.cfg.NextStepLabels.IsKnownStageLabel(targetLabel) {
		return querytypes.NextStepActionResult{}, errs.Validation{Field: "target_label", Msg: "must be a known run:* label"}
	}

	existingLabels, err := s.githubMgmt.ListIssueLabels(ctx, botToken, owner, repo, issueNumber)
	if err != nil {
		return querytypes.NextStepActionResult{}, fmt.Errorf("list issue labels: %w", err)
	}
	normalizedExisting := normalizeManagedLabels(existingLabels)
	removed := collectRunLabelsToRemove(normalizedExisting, targetLabel)
	added, finalLabels := previewIssueStageTransitionLabels(normalizedExisting, removed, targetLabel)
	if apply {
		for _, label := range removed {
			if err := s.githubMgmt.RemoveIssueLabel(ctx, botToken, owner, repo, issueNumber, label); err != nil {
				return querytypes.NextStepActionResult{}, fmt.Errorf("remove issue label %q: %w", label, err)
			}
		}
		if len(added) > 0 {
			if _, err := s.githubMgmt.AddIssueLabels(ctx, botToken, owner, repo, issueNumber, added); err != nil {
				return querytypes.NextStepActionResult{}, fmt.Errorf("add issue label %q: %w", targetLabel, err)
			}
		}
		finalLabels, err = s.githubMgmt.ListIssueLabels(ctx, botToken, owner, repo, issueNumber)
		if err != nil {
			return querytypes.NextStepActionResult{}, fmt.Errorf("list issue labels after transition: %w", err)
		}
		finalLabels = normalizeManagedLabels(finalLabels)
	}

	return querytypes.NextStepActionResult{
		RepositoryFullName: owner + "/" + repo,
		ThreadKind:         querytypes.NextStepThreadKindIssue,
		ThreadNumber:       issueNumber,
		ThreadURL:          fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, repo, issueNumber),
		RemovedLabels:      removed,
		AddedLabels:        added,
		FinalLabels:        finalLabels,
	}, nil
}

func (s *Service) previewOrExecutePullRequestLabelAdd(ctx context.Context, botToken string, owner string, repo string, pullRequestNumber int, targetLabel string, apply bool) (querytypes.NextStepActionResult, error) {
	if pullRequestNumber <= 0 {
		return querytypes.NextStepActionResult{}, errs.Validation{Field: "pull_request_number", Msg: "must be positive"}
	}
	if _, ok := knownNextStepPRLabels[targetLabel]; !ok {
		return querytypes.NextStepActionResult{}, errs.Validation{Field: "target_label", Msg: "must be a known pull-request label"}
	}

	existingLabels, err := s.githubMgmt.ListIssueLabels(ctx, botToken, owner, repo, pullRequestNumber)
	if err != nil {
		return querytypes.NextStepActionResult{}, fmt.Errorf("list pull request labels: %w", err)
	}
	normalizedExisting := normalizeManagedLabels(existingLabels)
	added, finalLabels := previewLabelAdd(normalizedExisting, targetLabel)
	if apply && len(added) > 0 {
		if _, err := s.githubMgmt.AddIssueLabels(ctx, botToken, owner, repo, pullRequestNumber, added); err != nil {
			return querytypes.NextStepActionResult{}, fmt.Errorf("add pull request label %q: %w", targetLabel, err)
		}
		finalLabels, err = s.githubMgmt.ListIssueLabels(ctx, botToken, owner, repo, pullRequestNumber)
		if err != nil {
			return querytypes.NextStepActionResult{}, fmt.Errorf("list pull request labels after add: %w", err)
		}
		finalLabels = normalizeManagedLabels(finalLabels)
	}

	return querytypes.NextStepActionResult{
		RepositoryFullName: owner + "/" + repo,
		ThreadKind:         querytypes.NextStepThreadKindPullRequest,
		ThreadNumber:       pullRequestNumber,
		ThreadURL:          fmt.Sprintf("https://github.com/%s/%s/pull/%d", owner, repo, pullRequestNumber),
		RemovedLabels:      []string{},
		AddedLabels:        added,
		FinalLabels:        finalLabels,
	}, nil
}

func previewIssueStageTransitionLabels(existingLabels []string, removedLabels []string, targetLabel string) (addedLabels []string, finalLabels []string) {
	finalLabels = make([]string, 0, len(existingLabels)+1)
	for _, label := range existingLabels {
		if slices.Contains(removedLabels, label) {
			continue
		}
		if !slices.Contains(finalLabels, label) {
			finalLabels = append(finalLabels, label)
		}
	}
	if !slices.Contains(finalLabels, targetLabel) {
		addedLabels = []string{targetLabel}
		finalLabels = append(finalLabels, targetLabel)
	}
	slices.Sort(finalLabels)
	return addedLabels, finalLabels
}

func previewLabelAdd(existingLabels []string, targetLabel string) (addedLabels []string, finalLabels []string) {
	finalLabels = append([]string(nil), existingLabels...)
	if !slices.Contains(finalLabels, targetLabel) {
		addedLabels = []string{targetLabel}
		finalLabels = append(finalLabels, targetLabel)
		slices.Sort(finalLabels)
	}
	if len(finalLabels) == 0 {
		finalLabels = []string{}
	}
	return addedLabels, finalLabels
}

func collectRunLabelsToRemove(existing []string, targetLabel string) []string {
	out := make([]string, 0, len(existing))
	for _, label := range existing {
		if !strings.HasPrefix(label, "run:") {
			continue
		}
		if label == targetLabel {
			continue
		}
		out = append(out, label)
	}
	slices.Sort(out)
	return out
}

func normalizeManagedLabels(labels []string) []string {
	out := make([]string, 0, len(labels))
	for _, raw := range labels {
		label := strings.ToLower(strings.TrimSpace(raw))
		if label == "" || slices.Contains(out, label) {
			continue
		}
		out = append(out, label)
	}
	slices.Sort(out)
	return out
}
