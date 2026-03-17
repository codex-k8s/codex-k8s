package nextstep

import (
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

type StageDescriptor struct {
	Stage       string
	RunLabel    string
	ReviseLabel string
}

type Config struct {
	RunIntake            string
	RunIntakeRevise      string
	RunVision            string
	RunVisionRevise      string
	RunPRD               string
	RunPRDRevise         string
	RunArch              string
	RunArchRevise        string
	RunDesign            string
	RunDesignRevise      string
	RunPlan              string
	RunPlanRevise        string
	RunDev               string
	RunDevRevise         string
	RunDocAudit          string
	RunDocAuditRevise    string
	RunQA                string
	RunQARevise          string
	RunRelease           string
	RunReleaseRevise     string
	RunPostDeploy        string
	RunPostDeployRevise  string
	RunOps               string
	RunOpsRevise         string
	RunSelfImprove       string
	RunSelfImproveRevise string
	RunRethink           string
}

type Labels struct {
	descriptors            map[string]StageDescriptor
	knownStageLabels       map[string]struct{}
	knownPullRequestLabels map[string]struct{}
}

func NewLabels(cfg Config) Labels {
	labels := Labels{
		descriptors:            make(map[string]StageDescriptor, 13),
		knownStageLabels:       make(map[string]struct{}, 25),
		knownPullRequestLabels: make(map[string]struct{}, 1),
	}
	labels.knownPullRequestLabels[normalizeToken(webhookdomain.DefaultNeedReviewerLabel)] = struct{}{}

	for _, descriptor := range []StageDescriptor{
		{Stage: "intake", RunLabel: valueOrDefault(cfg.RunIntake, webhookdomain.DefaultRunIntakeLabel), ReviseLabel: valueOrDefault(cfg.RunIntakeRevise, webhookdomain.DefaultRunIntakeReviseLabel)},
		{Stage: "vision", RunLabel: valueOrDefault(cfg.RunVision, webhookdomain.DefaultRunVisionLabel), ReviseLabel: valueOrDefault(cfg.RunVisionRevise, webhookdomain.DefaultRunVisionReviseLabel)},
		{Stage: "prd", RunLabel: valueOrDefault(cfg.RunPRD, webhookdomain.DefaultRunPRDLabel), ReviseLabel: valueOrDefault(cfg.RunPRDRevise, webhookdomain.DefaultRunPRDReviseLabel)},
		{Stage: "arch", RunLabel: valueOrDefault(cfg.RunArch, webhookdomain.DefaultRunArchLabel), ReviseLabel: valueOrDefault(cfg.RunArchRevise, webhookdomain.DefaultRunArchReviseLabel)},
		{Stage: "design", RunLabel: valueOrDefault(cfg.RunDesign, webhookdomain.DefaultRunDesignLabel), ReviseLabel: valueOrDefault(cfg.RunDesignRevise, webhookdomain.DefaultRunDesignReviseLabel)},
		{Stage: "plan", RunLabel: valueOrDefault(cfg.RunPlan, webhookdomain.DefaultRunPlanLabel), ReviseLabel: valueOrDefault(cfg.RunPlanRevise, webhookdomain.DefaultRunPlanReviseLabel)},
		{Stage: "dev", RunLabel: valueOrDefault(cfg.RunDev, webhookdomain.DefaultRunDevLabel), ReviseLabel: valueOrDefault(cfg.RunDevRevise, webhookdomain.DefaultRunDevReviseLabel)},
		{Stage: "doc-audit", RunLabel: valueOrDefault(cfg.RunDocAudit, webhookdomain.DefaultRunDocAuditLabel), ReviseLabel: valueOrDefault(cfg.RunDocAuditRevise, webhookdomain.DefaultRunDocAuditReviseLabel)},
		{Stage: "qa", RunLabel: valueOrDefault(cfg.RunQA, webhookdomain.DefaultRunQALabel), ReviseLabel: valueOrDefault(cfg.RunQARevise, webhookdomain.DefaultRunQAReviseLabel)},
		{Stage: "release", RunLabel: valueOrDefault(cfg.RunRelease, webhookdomain.DefaultRunReleaseLabel), ReviseLabel: valueOrDefault(cfg.RunReleaseRevise, webhookdomain.DefaultRunReleaseReviseLabel)},
		{Stage: "postdeploy", RunLabel: valueOrDefault(cfg.RunPostDeploy, webhookdomain.DefaultRunPostDeployLabel), ReviseLabel: valueOrDefault(cfg.RunPostDeployRevise, webhookdomain.DefaultRunPostDeployReviseLabel)},
		{Stage: "ops", RunLabel: valueOrDefault(cfg.RunOps, webhookdomain.DefaultRunOpsLabel), ReviseLabel: valueOrDefault(cfg.RunOpsRevise, webhookdomain.DefaultRunOpsReviseLabel)},
		{Stage: "self-improve", RunLabel: valueOrDefault(cfg.RunSelfImprove, webhookdomain.DefaultRunSelfImproveLabel), ReviseLabel: valueOrDefault(cfg.RunSelfImproveRevise, webhookdomain.DefaultRunSelfImproveReviseLabel)},
		{Stage: "rethink", RunLabel: valueOrDefault(cfg.RunRethink, webhookdomain.DefaultRunRethinkLabel)},
	} {
		labels.put(descriptor)
	}

	return labels
}

func DefaultLabels() Labels {
	return NewLabels(Config{})
}

func (l Labels) DescriptorByStage(stage string) (StageDescriptor, bool) {
	labels := l.withDefaults()
	descriptor, ok := labels.descriptors[normalizeStage(stage)]
	return descriptor, ok
}

func (l Labels) DescriptorByTriggerKind(triggerKind string) (StageDescriptor, bool) {
	switch normalizeToken(triggerKind) {
	case string(webhookdomain.TriggerKindIntake), string(webhookdomain.TriggerKindIntakeRevise):
		return l.DescriptorByStage("intake")
	case string(webhookdomain.TriggerKindVision), string(webhookdomain.TriggerKindVisionRevise):
		return l.DescriptorByStage("vision")
	case string(webhookdomain.TriggerKindPRD), string(webhookdomain.TriggerKindPRDRevise):
		return l.DescriptorByStage("prd")
	case string(webhookdomain.TriggerKindArch), string(webhookdomain.TriggerKindArchRevise):
		return l.DescriptorByStage("arch")
	case string(webhookdomain.TriggerKindDesign), string(webhookdomain.TriggerKindDesignRevise):
		return l.DescriptorByStage("design")
	case string(webhookdomain.TriggerKindPlan), string(webhookdomain.TriggerKindPlanRevise):
		return l.DescriptorByStage("plan")
	case string(webhookdomain.TriggerKindDev), string(webhookdomain.TriggerKindDevRevise):
		return l.DescriptorByStage("dev")
	case string(webhookdomain.TriggerKindDocAudit), string(webhookdomain.TriggerKindDocAuditRevise):
		return l.DescriptorByStage("doc-audit")
	case string(webhookdomain.TriggerKindQA), string(webhookdomain.TriggerKindQARevise):
		return l.DescriptorByStage("qa")
	case string(webhookdomain.TriggerKindRelease), string(webhookdomain.TriggerKindReleaseRevise):
		return l.DescriptorByStage("release")
	case string(webhookdomain.TriggerKindPostDeploy), string(webhookdomain.TriggerKindPostDeployRevise):
		return l.DescriptorByStage("postdeploy")
	case string(webhookdomain.TriggerKindOps), string(webhookdomain.TriggerKindOpsRevise):
		return l.DescriptorByStage("ops")
	case string(webhookdomain.TriggerKindSelfImprove), string(webhookdomain.TriggerKindSelfImproveRevise):
		return l.DescriptorByStage("self-improve")
	case string(webhookdomain.TriggerKindRethink):
		return l.DescriptorByStage("rethink")
	default:
		return StageDescriptor{}, false
	}
}

func (l Labels) IsKnownStageLabel(label string) bool {
	labels := l.withDefaults()
	_, ok := labels.knownStageLabels[normalizeToken(label)]
	return ok
}

func (l Labels) IsKnownPullRequestLabel(label string) bool {
	labels := l.withDefaults()
	_, ok := labels.knownPullRequestLabels[normalizeToken(label)]
	return ok
}

func (l Labels) IsKnownLabelForThreadKind(threadKind string, label string) bool {
	switch normalizeToken(threadKind) {
	case "issue":
		return l.IsKnownStageLabel(label)
	case "pull_request", "pr":
		return l.IsKnownPullRequestLabel(label)
	default:
		return false
	}
}

func (l Labels) put(descriptor StageDescriptor) {
	stage := normalizeStage(descriptor.Stage)
	runLabel := normalizeToken(descriptor.RunLabel)
	reviseLabel := normalizeToken(descriptor.ReviseLabel)
	normalized := StageDescriptor{
		Stage:       stage,
		RunLabel:    runLabel,
		ReviseLabel: reviseLabel,
	}
	l.descriptors[stage] = normalized
	if runLabel != "" {
		l.knownStageLabels[runLabel] = struct{}{}
	}
	if reviseLabel != "" {
		l.knownStageLabels[reviseLabel] = struct{}{}
	}
}

func (l Labels) withDefaults() Labels {
	if len(l.descriptors) == 0 || len(l.knownStageLabels) == 0 || len(l.knownPullRequestLabels) == 0 {
		return DefaultLabels()
	}
	return l
}

func normalizeStage(value string) string {
	return normalizeToken(value)
}

func normalizeToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func valueOrDefault(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
