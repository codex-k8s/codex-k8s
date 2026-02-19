package runner

import (
	"fmt"
	"slices"
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

const (
	promptSeedStageDev = "dev"
)

func normalizePromptTemplateKind(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case promptTemplateKindRevise, promptTemplateKindReviewOld:
		return promptTemplateKindRevise
	default:
		return promptTemplateKindWork
	}
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

func promptSeedCandidates(agentKey string, triggerKind string, templateKind string, locale string) []string {
	stage := promptSeedStageByTriggerKind(triggerKind)
	kind := normalizePromptTemplateKind(templateKind)
	normalizedLocale := normalizePromptLocale(locale)
	normalizedRole := strings.ToLower(strings.TrimSpace(agentKey))
	kinds := []string{kind}
	if kind == promptTemplateKindRevise {
		// Backward compatibility: allow legacy `*-review*.md` seeds as fallback
		// while runtime canonical kind is `revise`.
		kinds = append(kinds, promptTemplateKindReviewOld)
	}

	candidates := make([]string, 0, 24)
	for _, currentKind := range kinds {
		if normalizedRole != "" {
			candidates = append(candidates,
				fmt.Sprintf("%s-%s-%s_%s.md", stage, normalizedRole, currentKind, normalizedLocale),
				fmt.Sprintf("%s-%s-%s.md", stage, normalizedRole, currentKind),
				fmt.Sprintf("role-%s-%s_%s.md", normalizedRole, currentKind, normalizedLocale),
				fmt.Sprintf("role-%s-%s.md", normalizedRole, currentKind),
			)
		}

		candidates = append(candidates,
			fmt.Sprintf("%s-%s_%s.md", stage, currentKind, normalizedLocale),
			fmt.Sprintf("%s-%s.md", stage, currentKind),
		)
		if stage != promptSeedStageDev {
			candidates = append(candidates,
				fmt.Sprintf("%s-%s_%s.md", promptSeedStageDev, currentKind, normalizedLocale),
				fmt.Sprintf("%s-%s.md", promptSeedStageDev, currentKind),
			)
		}
		candidates = append(candidates,
			fmt.Sprintf("default-%s_%s.md", currentKind, normalizedLocale),
			fmt.Sprintf("default-%s.md", currentKind),
		)
	}

	return slices.Compact(candidates)
}
