package app

import "strings"

const (
	triggerKindDev       = "dev"
	triggerKindDevRevise = "dev_revise"

	promptTemplateKindWork   = "work"
	promptTemplateKindReview = "review"
	promptTemplateSourceSeed = "repo_seed"

	modelGPT53Codex      = "gpt-5.3-codex"
	modelGPT53CodexSpark = "gpt-5.3-codex-spark"
	modelGPT52Codex      = "gpt-5.2-codex"

	runtimeModeFullEnv  = "full-env"
	runtimeModeCodeOnly = "code-only"

	runDevLabelDefault        = "run:dev"
	runDevReviseLabelDefault  = "run:dev:revise"
	stateInReviewLabelDefault = "state:in-review"
)

func normalizeTriggerKind(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), triggerKindDevRevise) {
		return triggerKindDevRevise
	}
	return triggerKindDev
}
