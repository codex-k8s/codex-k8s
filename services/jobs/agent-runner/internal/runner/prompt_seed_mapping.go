package runner

import (
	"fmt"
	"slices"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

const (
	promptSeedStageDev = "dev"
)

func normalizePromptTemplateKind(value string) string {
	if value == promptTemplateKindReview {
		return promptTemplateKindReview
	}
	return promptTemplateKindWork
}

func promptSeedStageByTriggerKind(triggerKind string) string {
	switch webhookdomain.NormalizeTriggerKind(triggerKind) {
	case webhookdomain.TriggerKindIntake, webhookdomain.TriggerKindIntakeRevise:
		return "intake"
	case webhookdomain.TriggerKindVision, webhookdomain.TriggerKindVisionRevise:
		return "vision"
	case webhookdomain.TriggerKindPRD, webhookdomain.TriggerKindPRDRevise:
		return "prd"
	case webhookdomain.TriggerKindArch, webhookdomain.TriggerKindArchRevise:
		return "arch"
	case webhookdomain.TriggerKindDesign, webhookdomain.TriggerKindDesignRevise:
		return "design"
	case webhookdomain.TriggerKindPlan, webhookdomain.TriggerKindPlanRevise:
		return "plan"
	case webhookdomain.TriggerKindDocAudit:
		return "doc-audit"
	case webhookdomain.TriggerKindQA:
		return "qa"
	case webhookdomain.TriggerKindRelease:
		return "release"
	case webhookdomain.TriggerKindPostDeploy:
		return "postdeploy"
	case webhookdomain.TriggerKindOps:
		return "ops"
	case webhookdomain.TriggerKindSelfImprove:
		return "self-improve"
	case webhookdomain.TriggerKindRethink:
		return "rethink"
	default:
		return promptSeedStageDev
	}
}

func promptSeedCandidates(triggerKind string, templateKind string, locale string) []string {
	stage := promptSeedStageByTriggerKind(triggerKind)
	kind := normalizePromptTemplateKind(templateKind)
	normalizedLocale := normalizePromptLocale(locale)

	candidates := []string{
		fmt.Sprintf("%s-%s_%s.md", stage, kind, normalizedLocale),
		fmt.Sprintf("%s-%s.md", stage, kind),
	}
	if stage != promptSeedStageDev {
		candidates = append(candidates,
			fmt.Sprintf("%s-%s_%s.md", promptSeedStageDev, kind, normalizedLocale),
			fmt.Sprintf("%s-%s.md", promptSeedStageDev, kind),
		)
	}

	return slices.Compact(candidates)
}
