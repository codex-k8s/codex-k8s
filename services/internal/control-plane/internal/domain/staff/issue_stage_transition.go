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

var knownStageTransitionLabels = map[string]struct{}{
	webhookdomain.DefaultRunIntakeLabel:       {},
	webhookdomain.DefaultRunIntakeReviseLabel: {},
	webhookdomain.DefaultRunVisionLabel:       {},
	webhookdomain.DefaultRunVisionReviseLabel: {},
	webhookdomain.DefaultRunPRDLabel:          {},
	webhookdomain.DefaultRunPRDReviseLabel:    {},
	webhookdomain.DefaultRunArchLabel:         {},
	webhookdomain.DefaultRunArchReviseLabel:   {},
	webhookdomain.DefaultRunDesignLabel:       {},
	webhookdomain.DefaultRunDesignReviseLabel: {},
	webhookdomain.DefaultRunPlanLabel:         {},
	webhookdomain.DefaultRunPlanReviseLabel:   {},
	webhookdomain.DefaultRunDevLabel:          {},
	webhookdomain.DefaultRunDevReviseLabel:    {},
	webhookdomain.DefaultRunDocAuditLabel:     {},
	webhookdomain.DefaultRunQALabel:           {},
	webhookdomain.DefaultRunReleaseLabel:      {},
	webhookdomain.DefaultRunPostDeployLabel:   {},
	webhookdomain.DefaultRunOpsLabel:          {},
	webhookdomain.DefaultRunSelfImproveLabel:  {},
	webhookdomain.DefaultRunRethinkLabel:      {},
}

// TransitionIssueStageLabel applies one stage transition on GitHub issue labels.
func (s *Service) TransitionIssueStageLabel(ctx context.Context, principal Principal, params querytypes.IssueStageLabelTransitionParams) (querytypes.IssueStageLabelTransitionResult, error) {
	if !principal.IsPlatformAdmin {
		return querytypes.IssueStageLabelTransitionResult{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.githubMgmt == nil {
		return querytypes.IssueStageLabelTransitionResult{}, fmt.Errorf("failed_precondition: github management client is not configured")
	}

	repositoryFullName := strings.TrimSpace(params.RepositoryFullName)
	if repositoryFullName == "" {
		return querytypes.IssueStageLabelTransitionResult{}, errs.Validation{Field: "repository_full_name", Msg: "is required"}
	}
	if params.IssueNumber <= 0 {
		return querytypes.IssueStageLabelTransitionResult{}, errs.Validation{Field: "issue_number", Msg: "must be positive"}
	}

	targetLabel := strings.ToLower(strings.TrimSpace(params.TargetLabel))
	if targetLabel == "" {
		return querytypes.IssueStageLabelTransitionResult{}, errs.Validation{Field: "target_label", Msg: "is required"}
	}
	if _, ok := knownStageTransitionLabels[targetLabel]; !ok {
		return querytypes.IssueStageLabelTransitionResult{}, errs.Validation{Field: "target_label", Msg: "must be a known run:* label"}
	}

	owner, repo, err := parseGitHubFullName(repositoryFullName)
	if err != nil {
		return querytypes.IssueStageLabelTransitionResult{}, errs.Validation{Field: "repository_full_name", Msg: err.Error()}
	}

	binding, ok, err := s.repos.FindByProviderOwnerName(ctx, string(repoprovider.ProviderGitHub), owner, repo)
	if err != nil {
		return querytypes.IssueStageLabelTransitionResult{}, err
	}
	if !ok {
		return querytypes.IssueStageLabelTransitionResult{}, errs.Validation{Field: "repository_full_name", Msg: "repository is not bound to any project"}
	}

	_, botToken, _, _, err := s.resolveEffectiveGitHubTokens(ctx, binding.ProjectID, binding.RepositoryID)
	if err != nil {
		return querytypes.IssueStageLabelTransitionResult{}, err
	}

	existingLabels, err := s.githubMgmt.ListIssueLabels(ctx, botToken, owner, repo, params.IssueNumber)
	if err != nil {
		return querytypes.IssueStageLabelTransitionResult{}, fmt.Errorf("list issue labels: %w", err)
	}
	labelsToRemove := collectRunLabelsToRemove(existingLabels, targetLabel)

	removed := make([]string, 0, len(labelsToRemove))
	for _, label := range labelsToRemove {
		if err := s.githubMgmt.RemoveIssueLabel(ctx, botToken, owner, repo, params.IssueNumber, label); err != nil {
			return querytypes.IssueStageLabelTransitionResult{}, fmt.Errorf("remove issue label %q: %w", label, err)
		}
		removed = append(removed, label)
	}

	added := make([]string, 0, 1)
	if !slices.Contains(existingLabels, targetLabel) {
		if _, err := s.githubMgmt.AddIssueLabels(ctx, botToken, owner, repo, params.IssueNumber, []string{targetLabel}); err != nil {
			return querytypes.IssueStageLabelTransitionResult{}, fmt.Errorf("add issue label %q: %w", targetLabel, err)
		}
		added = append(added, targetLabel)
	}

	finalLabels, err := s.githubMgmt.ListIssueLabels(ctx, botToken, owner, repo, params.IssueNumber)
	if err != nil {
		return querytypes.IssueStageLabelTransitionResult{}, fmt.Errorf("list issue labels after transition: %w", err)
	}

	return querytypes.IssueStageLabelTransitionResult{
		RepositoryFullName: owner + "/" + repo,
		IssueNumber:        params.IssueNumber,
		IssueURL:           fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, repo, params.IssueNumber),
		RemovedLabels:      removed,
		AddedLabels:        added,
		FinalLabels:        finalLabels,
	}, nil
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
