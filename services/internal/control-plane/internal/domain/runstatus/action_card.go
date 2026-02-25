package runstatus

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

type launchProfile string

const (
	launchProfileQuickFix   launchProfile = "quick-fix"
	launchProfileFeature    launchProfile = "feature"
	launchProfileNewService launchProfile = "new-service"
)

const (
	guardrailNotePrecheckRequired    = "precheck_required"
	guardrailNoteAmbiguousStageLabel = "ambiguous_stage_labels"
	guardrailNoteStagePathUnresolved = "stage_path_unresolved"
	guardrailNoteNeedInputOnly       = "need_input_only"
)

const (
	needInputLabel = "need:input"
)

type stageDescriptor struct {
	Stage       string
	RunLabel    string
	ReviseLabel string
}

type nextStepActionCard struct {
	LaunchProfile  string
	StagePath      string
	PrimaryAction  string
	FallbackAction string
	GuardrailNote  string
	Blocked        bool
}

type runPayloadRawEnvelope struct {
	RawPayload json.RawMessage `json:"raw_payload"`
}

type runRawLabelsPayload struct {
	Issue       *runRawLabelsScope `json:"issue"`
	PullRequest *runRawLabelsScope `json:"pull_request"`
}

type runRawLabelsScope struct {
	Labels []runRawLabel `json:"labels"`
}

type runRawLabel struct {
	Name string `json:"name"`
}

func stageDescriptorByName(stage string) (stageDescriptor, bool) {
	switch strings.TrimSpace(stage) {
	case "intake":
		return stageDescriptor{
			Stage:       "intake",
			RunLabel:    webhookdomain.DefaultRunIntakeLabel,
			ReviseLabel: webhookdomain.DefaultRunIntakeReviseLabel,
		}, true
	case "vision":
		return stageDescriptor{
			Stage:       "vision",
			RunLabel:    webhookdomain.DefaultRunVisionLabel,
			ReviseLabel: webhookdomain.DefaultRunVisionReviseLabel,
		}, true
	case "prd":
		return stageDescriptor{
			Stage:       "prd",
			RunLabel:    webhookdomain.DefaultRunPRDLabel,
			ReviseLabel: webhookdomain.DefaultRunPRDReviseLabel,
		}, true
	case "arch":
		return stageDescriptor{
			Stage:       "arch",
			RunLabel:    webhookdomain.DefaultRunArchLabel,
			ReviseLabel: webhookdomain.DefaultRunArchReviseLabel,
		}, true
	case "design":
		return stageDescriptor{
			Stage:       "design",
			RunLabel:    webhookdomain.DefaultRunDesignLabel,
			ReviseLabel: webhookdomain.DefaultRunDesignReviseLabel,
		}, true
	case "plan":
		return stageDescriptor{
			Stage:       "plan",
			RunLabel:    webhookdomain.DefaultRunPlanLabel,
			ReviseLabel: webhookdomain.DefaultRunPlanReviseLabel,
		}, true
	case "dev":
		return stageDescriptor{
			Stage:       "dev",
			RunLabel:    webhookdomain.DefaultRunDevLabel,
			ReviseLabel: webhookdomain.DefaultRunDevReviseLabel,
		}, true
	case "qa":
		return stageDescriptor{
			Stage:    "qa",
			RunLabel: webhookdomain.DefaultRunQALabel,
		}, true
	case "release":
		return stageDescriptor{
			Stage:    "release",
			RunLabel: webhookdomain.DefaultRunReleaseLabel,
		}, true
	case "postdeploy":
		return stageDescriptor{
			Stage:    "postdeploy",
			RunLabel: webhookdomain.DefaultRunPostDeployLabel,
		}, true
	case "ops":
		return stageDescriptor{
			Stage:    "ops",
			RunLabel: webhookdomain.DefaultRunOpsLabel,
		}, true
	default:
		return stageDescriptor{}, false
	}
}

func stageDescriptorFromTriggerKind(triggerKind string) (stageDescriptor, bool) {
	switch normalizeTriggerKind(triggerKind) {
	case "intake", "intake_revise":
		return stageDescriptorByName("intake")
	case "vision", "vision_revise":
		return stageDescriptorByName("vision")
	case "prd", "prd_revise":
		return stageDescriptorByName("prd")
	case "arch", "arch_revise":
		return stageDescriptorByName("arch")
	case "design", "design_revise":
		return stageDescriptorByName("design")
	case "plan", "plan_revise":
		return stageDescriptorByName("plan")
	case "dev", "dev_revise":
		return stageDescriptorByName("dev")
	case "qa":
		return stageDescriptorByName("qa")
	case "release":
		return stageDescriptorByName("release")
	case "postdeploy":
		return stageDescriptorByName("postdeploy")
	case "ops":
		return stageDescriptorByName("ops")
	default:
		return stageDescriptor{}, false
	}
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

func profileStagePathString(profile launchProfile) string {
	path := profileStagePath(profile)
	if len(path) == 0 {
		return ""
	}
	return strings.Join(path, " -> ")
}

func launchProfileByStage(stage string) launchProfile {
	switch strings.TrimSpace(stage) {
	case "vision", "arch":
		return launchProfileNewService
	case "prd", "design":
		return launchProfileFeature
	default:
		return launchProfileQuickFix
	}
}

func launchProfileRank(profile launchProfile) int {
	switch profile {
	case launchProfileNewService:
		return 3
	case launchProfileFeature:
		return 2
	case launchProfileQuickFix:
		return 1
	default:
		return 0
	}
}

func maxLaunchProfile(left launchProfile, right launchProfile) launchProfile {
	if launchProfileRank(right) > launchProfileRank(left) {
		return right
	}
	return left
}

func resolveRiskEscalationProfile(labels []string) launchProfile {
	result := launchProfile("")
	for _, raw := range labels {
		label := strings.ToLower(strings.TrimSpace(raw))
		if label == "" {
			continue
		}
		if strings.Contains(label, "new-service") || strings.Contains(label, "new_service") ||
			strings.Contains(label, "architecture-boundary") || strings.Contains(label, "nfr-impact") {
			return launchProfileNewService
		}
		if strings.Contains(label, "cross-service") || strings.Contains(label, "cross_service") ||
			strings.Contains(label, "new-integration") || strings.Contains(label, "integration") ||
			strings.Contains(label, "migration") || strings.Contains(label, "rbac") || strings.Contains(label, "policy") {
			result = maxLaunchProfile(result, launchProfileFeature)
		}
	}
	return result
}

func resolveLaunchProfileForStage(stage string, labels []string) launchProfile {
	base := launchProfileByStage(stage)
	return maxLaunchProfile(base, resolveRiskEscalationProfile(labels))
}

func resolveNextStageForProfile(profile launchProfile, currentStage string) (string, bool) {
	path := profileStagePath(profile)
	normalizedCurrentStage := strings.TrimSpace(currentStage)
	for index, stage := range path {
		if stage != normalizedCurrentStage {
			continue
		}
		if index+1 >= len(path) {
			return "", false
		}
		return path[index+1], true
	}
	return "", false
}

func extractThreadLabelsFromRunPayload(runPayload json.RawMessage, targetKind commentTargetKind) []string {
	if len(runPayload) == 0 {
		return nil
	}
	var envelope runPayloadRawEnvelope
	if err := json.Unmarshal(runPayload, &envelope); err != nil {
		return nil
	}
	if len(envelope.RawPayload) == 0 {
		return nil
	}
	var payload runRawLabelsPayload
	if err := json.Unmarshal(envelope.RawPayload, &payload); err != nil {
		return nil
	}

	issueLabels := extractLabelNames(payload.Issue)
	pullRequestLabels := extractLabelNames(payload.PullRequest)

	if targetKind == commentTargetKindPullRequest && len(pullRequestLabels) > 0 {
		return normalizeLabelNames(pullRequestLabels)
	}
	if len(issueLabels) > 0 {
		return normalizeLabelNames(issueLabels)
	}
	return normalizeLabelNames(pullRequestLabels)
}

func extractLabelNames(scope *runRawLabelsScope) []string {
	if scope == nil || len(scope.Labels) == 0 {
		return nil
	}
	out := make([]string, 0, len(scope.Labels))
	for _, item := range scope.Labels {
		label := strings.TrimSpace(item.Name)
		if label == "" {
			continue
		}
		out = append(out, label)
	}
	return out
}

func normalizeLabelNames(labels []string) []string {
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

func collectStageLabels(labels []string) []string {
	out := make([]string, 0, 2)
	for _, label := range labels {
		descriptor, ok := stageDescriptorByRunLabel(label)
		if !ok {
			continue
		}
		if !slices.Contains(out, descriptor.RunLabel) && label == descriptor.RunLabel {
			out = append(out, descriptor.RunLabel)
		}
		if descriptor.ReviseLabel != "" && label == descriptor.ReviseLabel && !slices.Contains(out, descriptor.ReviseLabel) {
			out = append(out, descriptor.ReviseLabel)
		}
	}
	slices.Sort(out)
	return out
}

func stageDescriptorByRunLabel(label string) (stageDescriptor, bool) {
	candidates := []string{"intake", "vision", "prd", "arch", "design", "plan", "dev", "qa", "release", "postdeploy", "ops"}
	for _, stage := range candidates {
		descriptor, ok := stageDescriptorByName(stage)
		if !ok {
			continue
		}
		if label == descriptor.RunLabel || (descriptor.ReviseLabel != "" && label == descriptor.ReviseLabel) {
			return descriptor, true
		}
	}
	return stageDescriptor{}, false
}

func buildFallbackTransitionCommand(issueNumber int, currentRunLabel string, currentReviseLabel string, nextRunLabel string) string {
	if issueNumber <= 0 || strings.TrimSpace(nextRunLabel) == "" {
		return ""
	}

	preCheck := fmt.Sprintf("gh issue view %d --json labels --jq '.labels[].name'", issueNumber)

	parts := []string{fmt.Sprintf("gh issue edit %d", issueNumber)}
	normalizedCurrentRunLabel := strings.TrimSpace(currentRunLabel)
	normalizedCurrentReviseLabel := strings.TrimSpace(currentReviseLabel)
	normalizedNextRunLabel := strings.TrimSpace(nextRunLabel)

	if normalizedCurrentRunLabel != "" && normalizedCurrentRunLabel != normalizedNextRunLabel {
		parts = append(parts, fmt.Sprintf(`--remove-label "%s"`, normalizedCurrentRunLabel))
	}
	if normalizedCurrentReviseLabel != "" &&
		normalizedCurrentReviseLabel != normalizedCurrentRunLabel &&
		normalizedCurrentReviseLabel != normalizedNextRunLabel {
		parts = append(parts, fmt.Sprintf(`--remove-label "%s"`, normalizedCurrentReviseLabel))
	}
	parts = append(parts, fmt.Sprintf(`--add-label "%s"`, normalizedNextRunLabel))

	return preCheck + "\n" + strings.Join(parts, " ")
}

func buildNeedInputCommand(targetKind commentTargetKind, targetNumber int, issueNumber int) string {
	if issueNumber > 0 {
		return fmt.Sprintf(`gh issue edit %d --add-label "%s"`, issueNumber, needInputLabel)
	}
	if targetKind == commentTargetKindPullRequest && targetNumber > 0 {
		return fmt.Sprintf(`gh pr edit %d --add-label "%s"`, targetNumber, needInputLabel)
	}
	return ""
}

func isBlockedActionCard(guardrailNote string) bool {
	normalizedGuardrail := strings.TrimSpace(guardrailNote)
	if normalizedGuardrail == "" {
		return false
	}
	return normalizedGuardrail != guardrailNotePrecheckRequired
}
