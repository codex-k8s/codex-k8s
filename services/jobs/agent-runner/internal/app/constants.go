package app

import webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"

const (
	promptTemplateKindWork   = "work"
	promptTemplateKindRevise = "revise"
	promptTemplateSourceSeed = "repo_seed"

	modelGPT53Codex      = "gpt-5.3-codex"
	modelGPT53CodexSpark = "gpt-5.3-codex-spark"
	modelGPT52Codex      = "gpt-5.2-codex"

	runtimeModeFullEnv  = "full-env"
	runtimeModeCodeOnly = "code-only"

	stateInReviewLabelDefault = "state:in-review"
)

func normalizeTriggerKind(value string) string {
	return string(webhookdomain.NormalizeTriggerKind(value))
}

func isReviseTriggerKind(value string) bool {
	return webhookdomain.IsReviseTriggerKind(webhookdomain.NormalizeTriggerKind(value))
}
