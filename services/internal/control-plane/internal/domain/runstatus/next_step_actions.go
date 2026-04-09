package runstatus

import (
	"net/url"
	"slices"
	"strconv"
	"strings"

	nextstepdomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/nextstep"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
)

type launchProfile string

const (
	launchProfileQuickFix   launchProfile = "quick-fix"
	launchProfileFeature    launchProfile = "feature"
	launchProfileNewService launchProfile = "new-service"
)

const (
	nextStepDisplayRevise                   = "revise"
	nextStepDisplayFullFlow                 = "full_flow"
	nextStepDisplayShortenedFlow            = "shortened_flow"
	nextStepDisplayVeryShortFlow            = "very_short_flow"
	nextStepDisplayFullOrShortenedFlow      = "full_or_shortened_flow"
	nextStepDisplayFullOrVeryShortFlow      = "full_or_very_short_flow"
	nextStepDisplayShortenedOrVeryShortFlow = "shortened_or_very_short_flow"
	nextStepDisplayAllFlows                 = "all_flows"
	nextStepDisplayReviewer                 = "reviewer"
	nextStepDisplayRethink                  = "rethink"
	nextStepDisplayDocAudit                 = "doc_audit"
	nextStepDisplaySelfImprove              = "self_improve"
	nextStepDisplayPreparePlan              = "prepare_plan"
	nextStepDisplayGoToDev                  = "go_to_dev"
	nextStepDisplayGoToQA                   = "go_to_qa"
	nextStepDisplayRestartFull              = "restart_full"
	nextStepDisplayRestartShortened         = "restart_shortened"
	nextStepDisplayRestartVeryShort         = "restart_very_short"
)

type nextStepCommentAction struct {
	ActionKind     string
	DisplayVariant string
	TargetLabel    string
	URL            string
}

type groupedFlowAction struct {
	TargetLabel string
	Profiles    []launchProfile
}

type specialMatrixAction struct {
	TargetStage    string
	DisplayVariant string
}

type specialStageMatrix struct {
	Actions          []specialMatrixAction
	AllowDocAudit    bool
	AllowRethink     bool
	AllowSelfImprove bool
}

func stageDescriptorByName(labels nextstepdomain.Labels, stage string) (nextstepdomain.StageDescriptor, bool) {
	return labels.DescriptorByStage(stage)
}

func stageDescriptorFromTriggerKind(labels nextstepdomain.Labels, triggerKind string) (nextstepdomain.StageDescriptor, bool) {
	return labels.DescriptorByTriggerKind(normalizeTriggerKind(triggerKind))
}

func buildNextStepActions(publicBaseURL string, labels nextstepdomain.Labels, runCtx runContext, state commentState) []nextStepCommentAction {
	if isDiscussionTriggerKind(state.TriggerKind, state.TriggerLabel, state.DiscussionMode) {
		return nil
	}
	descriptor, ok := stageDescriptorFromTriggerKind(labels, state.TriggerKind)
	if !ok {
		return nil
	}

	if matrix, ok := specialStageMatrixForStage(descriptor.Stage); ok {
		return buildSpecialMatrixActions(
			publicBaseURL,
			state.RepositoryFullName,
			state.IssueNumber,
			pullRequestNumberFromRunContext(runCtx),
			labels,
			descriptor,
			matrix.Actions,
			matrix.AllowDocAudit,
			matrix.AllowRethink,
			matrix.AllowSelfImprove,
		)
	}

	actions := make([]nextStepCommentAction, 0, 8)
	if descriptor.ReviseLabel != "" {
		actions = append(actions, nextStepCommentAction{
			ActionKind:     querytypes.NextStepActionKindIssueStageTransition,
			DisplayVariant: nextStepDisplayRevise,
			TargetLabel:    descriptor.ReviseLabel,
			URL:            buildNextStepActionURL(publicBaseURL, state.RepositoryFullName, state.IssueNumber, pullRequestNumberFromRunContext(runCtx), querytypes.NextStepActionKindIssueStageTransition, descriptor.ReviseLabel, nextStepDisplayRevise),
		})
	}

	for _, item := range groupProfileFlowActions(descriptor.Stage, labels) {
		displayVariant := displayVariantForProfiles(item.Profiles)
		if displayVariant == "" {
			continue
		}
		actions = appendIssueStageTransitionAction(actions, publicBaseURL, state.RepositoryFullName, state.IssueNumber, pullRequestNumberFromRunContext(runCtx), item.TargetLabel, displayVariant)
	}
	if descriptor.Stage == "design" {
		updatedActions, ok := appendStageTransitionAction(actions, publicBaseURL, state.RepositoryFullName, state.IssueNumber, pullRequestNumberFromRunContext(runCtx), labels, "dev", nextStepDisplayGoToDev)
		if !ok {
			return actions
		}
		actions = updatedActions
	}

	if pullRequestNumber := pullRequestNumberFromRunContext(runCtx); pullRequestNumber > 0 {
		actions = append(actions, nextStepCommentAction{
			ActionKind:     querytypes.NextStepActionKindPullRequestLabelAdd,
			DisplayVariant: nextStepDisplayReviewer,
			TargetLabel:    nextStepReviewerLabel,
			URL:            buildNextStepActionURL(publicBaseURL, state.RepositoryFullName, state.IssueNumber, pullRequestNumber, querytypes.NextStepActionKindPullRequestLabelAdd, nextStepReviewerLabel, nextStepDisplayReviewer),
		})
	}
	if allowRethinkAction(descriptor.Stage) {
		updatedActions, ok := appendStageTransitionAction(actions, publicBaseURL, state.RepositoryFullName, state.IssueNumber, pullRequestNumberFromRunContext(runCtx), labels, "rethink", nextStepDisplayRethink)
		if !ok {
			return actions
		}
		actions = updatedActions
	}
	if allowDocAuditAction(descriptor.Stage) {
		updatedActions, ok := appendStageTransitionAction(actions, publicBaseURL, state.RepositoryFullName, state.IssueNumber, pullRequestNumberFromRunContext(runCtx), labels, "doc-audit", nextStepDisplayDocAudit)
		if !ok {
			return actions
		}
		actions = updatedActions
	}
	if allowSelfImproveAction(descriptor.Stage) {
		updatedActions, ok := appendStageTransitionAction(actions, publicBaseURL, state.RepositoryFullName, state.IssueNumber, pullRequestNumberFromRunContext(runCtx), labels, "self-improve", nextStepDisplaySelfImprove)
		if !ok {
			return actions
		}
		actions = updatedActions
	}

	return actions
}

func specialStageMatrixForStage(stage string) (specialStageMatrix, bool) {
	switch stage {
	case "doc-audit":
		return specialStageMatrix{
			Actions: []specialMatrixAction{
				{TargetStage: "plan", DisplayVariant: nextStepDisplayPreparePlan},
				{TargetStage: "dev", DisplayVariant: nextStepDisplayGoToDev},
				{TargetStage: "qa", DisplayVariant: nextStepDisplayGoToQA},
			},
			AllowDocAudit: true,
		}, true
	case "self-improve":
		return specialStageMatrix{
			Actions: []specialMatrixAction{
				{TargetStage: "qa", DisplayVariant: nextStepDisplayGoToQA},
			},
			AllowSelfImprove: true,
		}, true
	case "rethink":
		return specialStageMatrix{
			Actions: []specialMatrixAction{
				{TargetStage: "intake", DisplayVariant: nextStepDisplayRestartFull},
				{TargetStage: "prd", DisplayVariant: nextStepDisplayRestartShortened},
				{TargetStage: "plan", DisplayVariant: nextStepDisplayRestartVeryShort},
			},
		}, true
	default:
		return specialStageMatrix{}, false
	}
}

func buildSpecialMatrixActions(publicBaseURL string, repositoryFullName string, issueNumber int, pullRequestNumber int, labels nextstepdomain.Labels, descriptor nextstepdomain.StageDescriptor, targets []specialMatrixAction, includeRethink bool, includeDocAudit bool, includeReviewer bool) []nextStepCommentAction {
	actions := make([]nextStepCommentAction, 0, len(targets)+4)
	if descriptor.ReviseLabel != "" {
		actions = append(actions, nextStepCommentAction{
			ActionKind:     querytypes.NextStepActionKindIssueStageTransition,
			DisplayVariant: nextStepDisplayRevise,
			TargetLabel:    descriptor.ReviseLabel,
			URL:            buildNextStepActionURL(publicBaseURL, repositoryFullName, issueNumber, pullRequestNumber, querytypes.NextStepActionKindIssueStageTransition, descriptor.ReviseLabel, nextStepDisplayRevise),
		})
	}
	for _, target := range targets {
		if strings.TrimSpace(target.TargetStage) == "" || strings.TrimSpace(target.DisplayVariant) == "" {
			continue
		}
		updatedActions, ok := appendStageTransitionAction(actions, publicBaseURL, repositoryFullName, issueNumber, pullRequestNumber, labels, target.TargetStage, target.DisplayVariant)
		if !ok {
			continue
		}
		actions = updatedActions
	}
	if includeReviewer && pullRequestNumber > 0 {
		actions = append(actions, nextStepCommentAction{
			ActionKind:     querytypes.NextStepActionKindPullRequestLabelAdd,
			DisplayVariant: nextStepDisplayReviewer,
			TargetLabel:    nextStepReviewerLabel,
			URL:            buildNextStepActionURL(publicBaseURL, repositoryFullName, issueNumber, pullRequestNumber, querytypes.NextStepActionKindPullRequestLabelAdd, nextStepReviewerLabel, nextStepDisplayReviewer),
		})
	}
	if includeRethink {
		updatedActions, ok := appendStageTransitionAction(actions, publicBaseURL, repositoryFullName, issueNumber, pullRequestNumber, labels, "rethink", nextStepDisplayRethink)
		if !ok {
			return actions
		}
		actions = updatedActions
	}
	if includeDocAudit {
		updatedActions, ok := appendStageTransitionAction(actions, publicBaseURL, repositoryFullName, issueNumber, pullRequestNumber, labels, "doc-audit", nextStepDisplayDocAudit)
		if !ok {
			return actions
		}
		actions = updatedActions
	}
	return actions
}

func appendStageTransitionAction(actions []nextStepCommentAction, publicBaseURL string, repositoryFullName string, issueNumber int, pullRequestNumber int, labels nextstepdomain.Labels, stage string, displayVariant string) ([]nextStepCommentAction, bool) {
	descriptor, ok := stageDescriptorByName(labels, stage)
	if !ok || strings.TrimSpace(descriptor.RunLabel) == "" {
		return actions, false
	}
	return appendIssueStageTransitionAction(actions, publicBaseURL, repositoryFullName, issueNumber, pullRequestNumber, descriptor.RunLabel, displayVariant), true
}

func appendIssueStageTransitionAction(actions []nextStepCommentAction, publicBaseURL string, repositoryFullName string, issueNumber int, pullRequestNumber int, targetLabel string, displayVariant string) []nextStepCommentAction {
	return append(actions, nextStepCommentAction{
		ActionKind:     querytypes.NextStepActionKindIssueStageTransition,
		DisplayVariant: displayVariant,
		TargetLabel:    targetLabel,
		URL:            buildNextStepActionURL(publicBaseURL, repositoryFullName, issueNumber, pullRequestNumber, querytypes.NextStepActionKindIssueStageTransition, targetLabel, displayVariant),
	})
}

func pullRequestNumberFromRunContext(runCtx runContext) int {
	if runCtx.payload.PullRequest == nil || runCtx.payload.PullRequest.Number <= 0 {
		return 0
	}
	return int(runCtx.payload.PullRequest.Number)
}

func groupProfileFlowActions(currentStage string, labels nextstepdomain.Labels) []groupedFlowAction {
	targetToProfiles := make(map[string][]launchProfile, 3)
	for _, profile := range []launchProfile{launchProfileNewService, launchProfileFeature, launchProfileQuickFix} {
		targetStage, ok := resolveProfileTargetStage(currentStage, profile)
		if !ok {
			continue
		}
		targetDescriptor, ok := stageDescriptorByName(labels, targetStage)
		if !ok {
			continue
		}
		targetToProfiles[targetDescriptor.RunLabel] = append(targetToProfiles[targetDescriptor.RunLabel], profile)
	}
	targetLabels := make([]string, 0, len(targetToProfiles))
	for targetLabel := range targetToProfiles {
		targetLabels = append(targetLabels, targetLabel)
	}
	slices.SortFunc(targetLabels, func(left string, right string) int {
		return compareFlowTargets(left, right, targetToProfiles)
	})
	out := make([]groupedFlowAction, 0, len(targetLabels))
	for _, targetLabel := range targetLabels {
		profiles := targetToProfiles[targetLabel]
		slices.SortFunc(profiles, compareLaunchProfile)
		out = append(out, groupedFlowAction{TargetLabel: targetLabel, Profiles: profiles})
	}
	return out
}

func compareFlowTargets(left string, right string, targetToProfiles map[string][]launchProfile) int {
	leftProfiles := targetToProfiles[left]
	rightProfiles := targetToProfiles[right]
	if diff := compareLaunchProfile(leftProfiles[0], rightProfiles[0]); diff != 0 {
		return diff
	}
	return strings.Compare(left, right)
}

func compareLaunchProfile(left launchProfile, right launchProfile) int {
	return launchProfileOrder(left) - launchProfileOrder(right)
}

func launchProfileOrder(profile launchProfile) int {
	switch profile {
	case launchProfileNewService:
		return 0
	case launchProfileFeature:
		return 1
	case launchProfileQuickFix:
		return 2
	default:
		return 3
	}
}

func resolveProfileTargetStage(currentStage string, profile launchProfile) (string, bool) {
	stagePath := nextstepdomain.CanonicalMainStagePath()
	currentIndex := slices.Index(stagePath, currentStage)
	if currentIndex < 0 {
		return "", false
	}
	profilePath := profileStagePath(profile)
	for _, candidate := range stagePath[currentIndex+1:] {
		if slices.Contains(profilePath, candidate) {
			return candidate, true
		}
	}
	return "", false
}

func profileStagePath(profile launchProfile) []string {
	switch profile {
	case launchProfileNewService:
		return []string{"intake", "vision", "prd", "arch", "design", "plan", "dev", "qa", "release", "postdeploy", "ops"}
	case launchProfileFeature:
		return []string{"intake", "prd", "design", "plan", "dev", "qa", "release", "postdeploy", "ops"}
	default:
		return []string{"intake", "plan", "dev", "qa", "release", "postdeploy", "ops"}
	}
}

func displayVariantForProfiles(profiles []launchProfile) string {
	normalized := append([]launchProfile(nil), profiles...)
	slices.SortFunc(normalized, compareLaunchProfile)
	switch {
	case slices.Equal(normalized, []launchProfile{launchProfileNewService}):
		return nextStepDisplayFullFlow
	case slices.Equal(normalized, []launchProfile{launchProfileFeature}):
		return nextStepDisplayShortenedFlow
	case slices.Equal(normalized, []launchProfile{launchProfileQuickFix}):
		return nextStepDisplayVeryShortFlow
	case slices.Equal(normalized, []launchProfile{launchProfileNewService, launchProfileFeature}):
		return nextStepDisplayFullOrShortenedFlow
	case slices.Equal(normalized, []launchProfile{launchProfileNewService, launchProfileQuickFix}):
		return nextStepDisplayFullOrVeryShortFlow
	case slices.Equal(normalized, []launchProfile{launchProfileFeature, launchProfileQuickFix}):
		return nextStepDisplayShortenedOrVeryShortFlow
	case slices.Equal(normalized, []launchProfile{launchProfileNewService, launchProfileFeature, launchProfileQuickFix}):
		return nextStepDisplayAllFlows
	default:
		return ""
	}
}

func allowRethinkAction(stage string) bool {
	return slices.Contains([]string{"vision", "prd", "arch", "design", "plan", "dev", "qa", "release", "postdeploy", "ops"}, stage)
}

func allowDocAuditAction(stage string) bool {
	return slices.Contains([]string{"dev", "qa", "release", "postdeploy", "ops"}, stage)
}

func allowSelfImproveAction(stage string) bool {
	return slices.Contains([]string{"postdeploy", "ops"}, stage)
}

func buildNextStepActionURL(publicBaseURL string, repositoryFullName string, issueNumber int, pullRequestNumber int, actionKind string, targetLabel string, displayVariant string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	repositoryFullName = strings.TrimSpace(repositoryFullName)
	actionKind = strings.TrimSpace(actionKind)
	targetLabel = strings.TrimSpace(targetLabel)
	displayVariant = strings.TrimSpace(displayVariant)
	if baseURL == "" || repositoryFullName == "" || issueNumber <= 0 || actionKind == "" || targetLabel == "" || displayVariant == "" {
		return ""
	}

	values := url.Values{}
	values.Set("modal", "next-step")
	values.Set("repository_full_name", repositoryFullName)
	values.Set("issue_number", strconv.Itoa(issueNumber))
	if pullRequestNumber > 0 {
		values.Set("pull_request_number", strconv.Itoa(pullRequestNumber))
	}
	values.Set("action_kind", actionKind)
	values.Set("target_label", targetLabel)
	values.Set("display_variant", displayVariant)
	return baseURL + "/?" + values.Encode()
}
