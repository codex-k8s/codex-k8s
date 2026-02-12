package app

import "strings"

const (
	triggerKindDev       = "dev"
	triggerKindDevRevise = "dev_revise"

	promptTemplateKindWork   = "work"
	promptTemplateKindReview = "review"
	promptTemplateSourceSeed = "repo_seed"

	modelGPT53Codex = "gpt-5.3-codex"
	modelGPT52Codex = "gpt-5.2-codex"
)

func normalizeTriggerKind(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), triggerKindDevRevise) {
		return triggerKindDevRevise
	}
	return triggerKindDev
}
