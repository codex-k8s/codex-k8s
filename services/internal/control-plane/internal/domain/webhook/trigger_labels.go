package webhook

import (
	"slices"
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

func (labels TriggerLabels) withDefaults() TriggerLabels {
	defaults := defaultTriggerLabels()
	if strings.TrimSpace(labels.RunIntake) == "" {
		labels.RunIntake = defaults.RunIntake
	}
	if strings.TrimSpace(labels.RunIntakeRevise) == "" {
		labels.RunIntakeRevise = defaults.RunIntakeRevise
	}
	if strings.TrimSpace(labels.RunVision) == "" {
		labels.RunVision = defaults.RunVision
	}
	if strings.TrimSpace(labels.RunVisionRevise) == "" {
		labels.RunVisionRevise = defaults.RunVisionRevise
	}
	if strings.TrimSpace(labels.RunPRD) == "" {
		labels.RunPRD = defaults.RunPRD
	}
	if strings.TrimSpace(labels.RunPRDRevise) == "" {
		labels.RunPRDRevise = defaults.RunPRDRevise
	}
	if strings.TrimSpace(labels.RunArch) == "" {
		labels.RunArch = defaults.RunArch
	}
	if strings.TrimSpace(labels.RunArchRevise) == "" {
		labels.RunArchRevise = defaults.RunArchRevise
	}
	if strings.TrimSpace(labels.RunDesign) == "" {
		labels.RunDesign = defaults.RunDesign
	}
	if strings.TrimSpace(labels.RunDesignRevise) == "" {
		labels.RunDesignRevise = defaults.RunDesignRevise
	}
	if strings.TrimSpace(labels.RunPlan) == "" {
		labels.RunPlan = defaults.RunPlan
	}
	if strings.TrimSpace(labels.RunPlanRevise) == "" {
		labels.RunPlanRevise = defaults.RunPlanRevise
	}
	if strings.TrimSpace(labels.RunDev) == "" {
		labels.RunDev = defaults.RunDev
	}
	if strings.TrimSpace(labels.RunDevRevise) == "" {
		labels.RunDevRevise = defaults.RunDevRevise
	}
	if strings.TrimSpace(labels.RunDocAudit) == "" {
		labels.RunDocAudit = defaults.RunDocAudit
	}
	if strings.TrimSpace(labels.RunAIRepair) == "" {
		labels.RunAIRepair = defaults.RunAIRepair
	}
	if strings.TrimSpace(labels.RunQA) == "" {
		labels.RunQA = defaults.RunQA
	}
	if strings.TrimSpace(labels.RunRelease) == "" {
		labels.RunRelease = defaults.RunRelease
	}
	if strings.TrimSpace(labels.RunPostDeploy) == "" {
		labels.RunPostDeploy = defaults.RunPostDeploy
	}
	if strings.TrimSpace(labels.RunOps) == "" {
		labels.RunOps = defaults.RunOps
	}
	if strings.TrimSpace(labels.RunSelfImprove) == "" {
		labels.RunSelfImprove = defaults.RunSelfImprove
	}
	if strings.TrimSpace(labels.RunRethink) == "" {
		labels.RunRethink = defaults.RunRethink
	}
	if strings.TrimSpace(labels.NeedReviewer) == "" {
		labels.NeedReviewer = defaults.NeedReviewer
	}
	return labels
}

func (labels TriggerLabels) labelToKind() map[string]webhookdomain.TriggerKind {
	normalized := labels.withDefaults()
	return map[string]webhookdomain.TriggerKind{
		normalizeLabelToken(normalized.RunIntake):       webhookdomain.TriggerKindIntake,
		normalizeLabelToken(normalized.RunIntakeRevise): webhookdomain.TriggerKindIntakeRevise,
		normalizeLabelToken(normalized.RunVision):       webhookdomain.TriggerKindVision,
		normalizeLabelToken(normalized.RunVisionRevise): webhookdomain.TriggerKindVisionRevise,
		normalizeLabelToken(normalized.RunPRD):          webhookdomain.TriggerKindPRD,
		normalizeLabelToken(normalized.RunPRDRevise):    webhookdomain.TriggerKindPRDRevise,
		normalizeLabelToken(normalized.RunArch):         webhookdomain.TriggerKindArch,
		normalizeLabelToken(normalized.RunArchRevise):   webhookdomain.TriggerKindArchRevise,
		normalizeLabelToken(normalized.RunDesign):       webhookdomain.TriggerKindDesign,
		normalizeLabelToken(normalized.RunDesignRevise): webhookdomain.TriggerKindDesignRevise,
		normalizeLabelToken(normalized.RunPlan):         webhookdomain.TriggerKindPlan,
		normalizeLabelToken(normalized.RunPlanRevise):   webhookdomain.TriggerKindPlanRevise,
		normalizeLabelToken(normalized.RunDev):          webhookdomain.TriggerKindDev,
		normalizeLabelToken(normalized.RunDevRevise):    webhookdomain.TriggerKindDevRevise,
		normalizeLabelToken(normalized.RunDocAudit):     webhookdomain.TriggerKindDocAudit,
		normalizeLabelToken(normalized.RunAIRepair):     webhookdomain.TriggerKindAIRepair,
		normalizeLabelToken(normalized.RunQA):           webhookdomain.TriggerKindQA,
		normalizeLabelToken(normalized.RunRelease):      webhookdomain.TriggerKindRelease,
		normalizeLabelToken(normalized.RunPostDeploy):   webhookdomain.TriggerKindPostDeploy,
		normalizeLabelToken(normalized.RunOps):          webhookdomain.TriggerKindOps,
		normalizeLabelToken(normalized.RunSelfImprove):  webhookdomain.TriggerKindSelfImprove,
		normalizeLabelToken(normalized.RunRethink):      webhookdomain.TriggerKindRethink,
	}
}

func (labels TriggerLabels) resolveKind(label string) (webhookdomain.TriggerKind, bool) {
	kind, ok := labels.labelToKind()[normalizeLabelToken(label)]
	if !ok {
		return "", false
	}
	return kind, true
}

func (labels TriggerLabels) collectIssueTriggerLabels(issueLabels []githubLabelRecord) []string {
	catalog := labels.labelToKind()
	resolved := make([]string, 0, len(issueLabels))
	for _, item := range issueLabels {
		resolvedLabel := normalizeLabelToken(item.Name)
		if resolvedLabel == "" {
			continue
		}
		if _, ok := catalog[resolvedLabel]; !ok {
			continue
		}
		if !slices.Contains(resolved, resolvedLabel) {
			resolved = append(resolved, resolvedLabel)
		}
	}
	slices.Sort(resolved)
	return resolved
}

func (labels TriggerLabels) isNeedReviewerLabel(label string) bool {
	reviewerLabel := normalizeLabelToken(labels.withDefaults().NeedReviewer)
	if reviewerLabel == "" {
		return false
	}
	return normalizeLabelToken(label) == reviewerLabel
}

func normalizeLabelToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
