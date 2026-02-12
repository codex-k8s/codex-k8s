package app

import "strings"

const (
	triggerKindDev       = "dev"
	triggerKindDevRevise = "dev_revise"

	promptTemplateKindWork   = "work"
	promptTemplateKindReview = "review"
	promptTemplateSourceSeed = "repo_seed"
)

func normalizeTriggerKind(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), triggerKindDevRevise) {
		return triggerKindDevRevise
	}
	return triggerKindDev
}
