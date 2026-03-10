package runstatus

import (
	"net/url"
	"slices"
	"strconv"
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
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

type stageDescriptor struct {
	Stage       string
	RunLabel    string
	ReviseLabel string
}

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
	TargetLabel    string
	DisplayVariant string
}

type specialStageMatrix struct {
	Actions          []specialMatrixAction
	AllowDocAudit    bool
	AllowRethink     bool
	AllowSelfImprove bool
}

func stageDescriptorByName(stage string) (stageDescriptor, bool) {
	switch strings.TrimSpace(stage) {
	case "intake":
		return stageDescriptor{Stage: "intake", RunLabel: webhookdomain.DefaultRunIntakeLabel, ReviseLabel: webhookdomain.DefaultRunIntakeReviseLabel}, true
	case "vision":
		return stageDescriptor{Stage: "vision", RunLabel: webhookdomain.DefaultRunVisionLabel, ReviseLabel: webhookdomain.DefaultRunVisionReviseLabel}, true
	case "prd":
		return stageDescriptor{Stage: "prd", RunLabel: webhookdomain.DefaultRunPRDLabel, ReviseLabel: webhookdomain.DefaultRunPRDReviseLabel}, true
	case "arch":
		return stageDescriptor{Stage: "arch", RunLabel: webhookdomain.DefaultRunArchLabel, ReviseLabel: webhookdomain.DefaultRunArchReviseLabel}, true
	case "design":
		return stageDescriptor{Stage: "design", RunLabel: webhookdomain.DefaultRunDesignLabel, ReviseLabel: webhookdomain.DefaultRunDesignReviseLabel}, true
	case "plan":
		return stageDescriptor{Stage: "plan", RunLabel: webhookdomain.DefaultRunPlanLabel, ReviseLabel: webhookdomain.DefaultRunPlanReviseLabel}, true
	case "dev":
		return stageDescriptor{Stage: "dev", RunLabel: webhookdomain.DefaultRunDevLabel, ReviseLabel: webhookdomain.DefaultRunDevReviseLabel}, true
	case "doc-audit":
		return stageDescriptor{Stage: "doc-audit", RunLabel: webhookdomain.DefaultRunDocAuditLabel, ReviseLabel: webhookdomain.DefaultRunDocAuditReviseLabel}, true
	case "qa":
		return stageDescriptor{Stage: "qa", RunLabel: webhookdomain.DefaultRunQALabel, ReviseLabel: webhookdomain.DefaultRunQAReviseLabel}, true
	case "release":
		return stageDescriptor{Stage: "release", RunLabel: webhookdomain.DefaultRunReleaseLabel, ReviseLabel: webhookdomain.DefaultRunReleaseReviseLabel}, true
	case "postdeploy":
		return stageDescriptor{Stage: "postdeploy", RunLabel: webhookdomain.DefaultRunPostDeployLabel, ReviseLabel: webhookdomain.DefaultRunPostDeployReviseLabel}, true
	case "ops":
		return stageDescriptor{Stage: "ops", RunLabel: webhookdomain.DefaultRunOpsLabel, ReviseLabel: webhookdomain.DefaultRunOpsReviseLabel}, true
	case "self-improve":
		return stageDescriptor{Stage: "self-improve", RunLabel: webhookdomain.DefaultRunSelfImproveLabel, ReviseLabel: webhookdomain.DefaultRunSelfImproveReviseLabel}, true
	case "rethink":
		return stageDescriptor{Stage: "rethink", RunLabel: webhookdomain.DefaultRunRethinkLabel}, true
	default:
		return stageDescriptor{}, false
	}
}

func stageDescriptorFromTriggerKind(triggerKind string) (stageDescriptor, bool) {
	switch normalizeTriggerKind(triggerKind) {
	case string(webhookdomain.TriggerKindIntake), string(webhookdomain.TriggerKindIntakeRevise):
		return stageDescriptorByName("intake")
	case string(webhookdomain.TriggerKindVision), string(webhookdomain.TriggerKindVisionRevise):
		return stageDescriptorByName("vision")
	case string(webhookdomain.TriggerKindPRD), string(webhookdomain.TriggerKindPRDRevise):
		return stageDescriptorByName("prd")
	case string(webhookdomain.TriggerKindArch), string(webhookdomain.TriggerKindArchRevise):
		return stageDescriptorByName("arch")
	case string(webhookdomain.TriggerKindDesign), string(webhookdomain.TriggerKindDesignRevise):
		return stageDescriptorByName("design")
	case string(webhookdomain.TriggerKindPlan), string(webhookdomain.TriggerKindPlanRevise):
		return stageDescriptorByName("plan")
	case string(webhookdomain.TriggerKindDev), string(webhookdomain.TriggerKindDevRevise):
		return stageDescriptorByName("dev")
	case string(webhookdomain.TriggerKindDocAudit):
		return stageDescriptorByName("doc-audit")
	case string(webhookdomain.TriggerKindQA):
		return stageDescriptorByName("qa")
	case string(webhookdomain.TriggerKindRelease):
		return stageDescriptorByName("release")
	case string(webhookdomain.TriggerKindPostDeploy):
		return stageDescriptorByName("postdeploy")
	case string(webhookdomain.TriggerKindOps):
		return stageDescriptorByName("ops")
	case string(webhookdomain.TriggerKindSelfImprove):
		return stageDescriptorByName("self-improve")
	case string(webhookdomain.TriggerKindRethink):
		return stageDescriptorByName("rethink")
	default:
		return stageDescriptor{}, false
	}
}

func buildNextStepActions(publicBaseURL string, runCtx runContext, state commentState) []nextStepCommentAction {
	descriptor, ok := stageDescriptorFromTriggerKind(state.TriggerKind)
	if !ok {
		return nil
	}

	if matrix, ok := specialStageMatrixForStage(descriptor.Stage); ok {
		return buildSpecialMatrixActions(
			publicBaseURL,
			state.RepositoryFullName,
			state.IssueNumber,
			pullRequestNumberFromRunContext(runCtx),
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

	for _, item := range groupProfileFlowActions(descriptor.Stage) {
		displayVariant := displayVariantForProfiles(item.Profiles)
		if displayVariant == "" {
			continue
		}
		actions = append(actions, nextStepCommentAction{
			ActionKind:     querytypes.NextStepActionKindIssueStageTransition,
			DisplayVariant: displayVariant,
			TargetLabel:    item.TargetLabel,
			URL:            buildNextStepActionURL(publicBaseURL, state.RepositoryFullName, state.IssueNumber, pullRequestNumberFromRunContext(runCtx), querytypes.NextStepActionKindIssueStageTransition, item.TargetLabel, displayVariant),
		})
	}
	if descriptor.Stage == "design" {
		actions = append(actions, nextStepCommentAction{
			ActionKind:     querytypes.NextStepActionKindIssueStageTransition,
			DisplayVariant: nextStepDisplayGoToDev,
			TargetLabel:    webhookdomain.DefaultRunDevLabel,
			URL:            buildNextStepActionURL(publicBaseURL, state.RepositoryFullName, state.IssueNumber, pullRequestNumberFromRunContext(runCtx), querytypes.NextStepActionKindIssueStageTransition, webhookdomain.DefaultRunDevLabel, nextStepDisplayGoToDev),
		})
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
		actions = append(actions, nextStepCommentAction{
			ActionKind:     querytypes.NextStepActionKindIssueStageTransition,
			DisplayVariant: nextStepDisplayRethink,
			TargetLabel:    webhookdomain.DefaultRunRethinkLabel,
			URL:            buildNextStepActionURL(publicBaseURL, state.RepositoryFullName, state.IssueNumber, pullRequestNumberFromRunContext(runCtx), querytypes.NextStepActionKindIssueStageTransition, webhookdomain.DefaultRunRethinkLabel, nextStepDisplayRethink),
		})
	}
	if allowDocAuditAction(descriptor.Stage) {
		actions = append(actions, nextStepCommentAction{
			ActionKind:     querytypes.NextStepActionKindIssueStageTransition,
			DisplayVariant: nextStepDisplayDocAudit,
			TargetLabel:    webhookdomain.DefaultRunDocAuditLabel,
			URL:            buildNextStepActionURL(publicBaseURL, state.RepositoryFullName, state.IssueNumber, pullRequestNumberFromRunContext(runCtx), querytypes.NextStepActionKindIssueStageTransition, webhookdomain.DefaultRunDocAuditLabel, nextStepDisplayDocAudit),
		})
	}
	if allowSelfImproveAction(descriptor.Stage) {
		actions = append(actions, nextStepCommentAction{
			ActionKind:     querytypes.NextStepActionKindIssueStageTransition,
			DisplayVariant: nextStepDisplaySelfImprove,
			TargetLabel:    webhookdomain.DefaultRunSelfImproveLabel,
			URL:            buildNextStepActionURL(publicBaseURL, state.RepositoryFullName, state.IssueNumber, pullRequestNumberFromRunContext(runCtx), querytypes.NextStepActionKindIssueStageTransition, webhookdomain.DefaultRunSelfImproveLabel, nextStepDisplaySelfImprove),
		})
	}

	return actions
}

func specialStageMatrixForStage(stage string) (specialStageMatrix, bool) {
	switch stage {
	case "doc-audit":
		return specialStageMatrix{
			Actions: []specialMatrixAction{
				{TargetLabel: webhookdomain.DefaultRunPlanLabel, DisplayVariant: nextStepDisplayPreparePlan},
				{TargetLabel: webhookdomain.DefaultRunDevLabel, DisplayVariant: nextStepDisplayGoToDev},
				{TargetLabel: webhookdomain.DefaultRunQALabel, DisplayVariant: nextStepDisplayGoToQA},
			},
			AllowDocAudit: true,
		}, true
	case "self-improve":
		return specialStageMatrix{
			Actions: []specialMatrixAction{
				{TargetLabel: webhookdomain.DefaultRunQALabel, DisplayVariant: nextStepDisplayGoToQA},
			},
			AllowSelfImprove: true,
		}, true
	case "rethink":
		return specialStageMatrix{
			Actions: []specialMatrixAction{
				{TargetLabel: webhookdomain.DefaultRunIntakeLabel, DisplayVariant: nextStepDisplayRestartFull},
				{TargetLabel: webhookdomain.DefaultRunPRDLabel, DisplayVariant: nextStepDisplayRestartShortened},
				{TargetLabel: webhookdomain.DefaultRunPlanLabel, DisplayVariant: nextStepDisplayRestartVeryShort},
			},
		}, true
	default:
		return specialStageMatrix{}, false
	}
}

func buildSpecialMatrixActions(publicBaseURL string, repositoryFullName string, issueNumber int, pullRequestNumber int, descriptor stageDescriptor, targets []specialMatrixAction, includeRethink bool, includeDocAudit bool, includeReviewer bool) []nextStepCommentAction {
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
		if strings.TrimSpace(target.TargetLabel) == "" || strings.TrimSpace(target.DisplayVariant) == "" {
			continue
		}
		actions = append(actions, nextStepCommentAction{
			ActionKind:     querytypes.NextStepActionKindIssueStageTransition,
			DisplayVariant: target.DisplayVariant,
			TargetLabel:    target.TargetLabel,
			URL:            buildNextStepActionURL(publicBaseURL, repositoryFullName, issueNumber, pullRequestNumber, querytypes.NextStepActionKindIssueStageTransition, target.TargetLabel, target.DisplayVariant),
		})
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
		actions = append(actions, nextStepCommentAction{
			ActionKind:     querytypes.NextStepActionKindIssueStageTransition,
			DisplayVariant: nextStepDisplayRethink,
			TargetLabel:    webhookdomain.DefaultRunRethinkLabel,
			URL:            buildNextStepActionURL(publicBaseURL, repositoryFullName, issueNumber, pullRequestNumber, querytypes.NextStepActionKindIssueStageTransition, webhookdomain.DefaultRunRethinkLabel, nextStepDisplayRethink),
		})
	}
	if includeDocAudit {
		actions = append(actions, nextStepCommentAction{
			ActionKind:     querytypes.NextStepActionKindIssueStageTransition,
			DisplayVariant: nextStepDisplayDocAudit,
			TargetLabel:    webhookdomain.DefaultRunDocAuditLabel,
			URL:            buildNextStepActionURL(publicBaseURL, repositoryFullName, issueNumber, pullRequestNumber, querytypes.NextStepActionKindIssueStageTransition, webhookdomain.DefaultRunDocAuditLabel, nextStepDisplayDocAudit),
		})
	}
	return actions
}

func pullRequestNumberFromRunContext(runCtx runContext) int {
	if runCtx.payload.PullRequest == nil || runCtx.payload.PullRequest.Number <= 0 {
		return 0
	}
	return int(runCtx.payload.PullRequest.Number)
}

func groupProfileFlowActions(currentStage string) []groupedFlowAction {
	targetToProfiles := make(map[string][]launchProfile, 3)
	for _, profile := range []launchProfile{launchProfileNewService, launchProfileFeature, launchProfileQuickFix} {
		targetStage, ok := resolveProfileTargetStage(currentStage, profile)
		if !ok {
			continue
		}
		targetDescriptor, ok := stageDescriptorByName(targetStage)
		if !ok {
			continue
		}
		targetToProfiles[targetDescriptor.RunLabel] = append(targetToProfiles[targetDescriptor.RunLabel], profile)
	}
	labels := make([]string, 0, len(targetToProfiles))
	for targetLabel := range targetToProfiles {
		labels = append(labels, targetLabel)
	}
	slices.SortFunc(labels, func(left string, right string) int {
		return compareFlowTargets(left, right, targetToProfiles)
	})
	out := make([]groupedFlowAction, 0, len(labels))
	for _, targetLabel := range labels {
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
	currentIndex := slices.Index(fullStagePath(), currentStage)
	if currentIndex < 0 {
		return "", false
	}
	profilePath := profileStagePath(profile)
	for _, candidate := range fullStagePath()[currentIndex+1:] {
		if slices.Contains(profilePath, candidate) {
			return candidate, true
		}
	}
	return "", false
}

func fullStagePath() []string {
	return []string{"intake", "vision", "prd", "arch", "design", "plan", "dev", "qa", "release", "postdeploy", "ops"}
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
