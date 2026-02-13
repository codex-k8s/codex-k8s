package worker

const (
	labelKindAIModel     = "ai-model"
	labelKindAIReasoning = "ai-reasoning"

	reasoningEffortLow       = "low"
	reasoningEffortMedium    = "medium"
	reasoningEffortHigh      = "high"
	reasoningEffortExtraHigh = "extra-high"

	defaultAIModelGPT53CodexLabel      = "ai-model-gpt-5.3-codex"
	defaultAIModelGPT53CodexSparkLabel = "ai-model-gpt-5.3-codex-spark"
	defaultAIModelGPT52CodexLabel      = "ai-model-gpt-5.2-codex"
	defaultAIModelGPT52Label           = "ai-model-gpt-5.2"
	defaultAIModelGPT51CodexMaxLabel   = "ai-model-gpt-5.1-codex-max"
	defaultAIModelGPT51CodexMiniLabel  = "ai-model-gpt-5.1-codex-mini"

	defaultAIReasoningLowLabel       = "ai-reasoning-low"
	defaultAIReasoningMediumLabel    = "ai-reasoning-medium"
	defaultAIReasoningHighLabel      = "ai-reasoning-high"
	defaultAIReasoningExtraHighLabel = "ai-reasoning-extra-high"
)

type runAgentLabelCatalog struct {
	AIModelGPT53CodexLabel      string
	AIModelGPT53CodexSparkLabel string
	AIModelGPT52CodexLabel      string
	AIModelGPT52Label           string
	AIModelGPT51CodexMaxLabel   string
	AIModelGPT51CodexMiniLabel  string

	AIReasoningLowLabel       string
	AIReasoningMediumLabel    string
	AIReasoningHighLabel      string
	AIReasoningExtraHighLabel string
}

func defaultRunAgentLabelCatalog() runAgentLabelCatalog {
	return runAgentLabelCatalog{
		AIModelGPT53CodexLabel:      defaultAIModelGPT53CodexLabel,
		AIModelGPT53CodexSparkLabel: defaultAIModelGPT53CodexSparkLabel,
		AIModelGPT52CodexLabel:      defaultAIModelGPT52CodexLabel,
		AIModelGPT52Label:           defaultAIModelGPT52Label,
		AIModelGPT51CodexMaxLabel:   defaultAIModelGPT51CodexMaxLabel,
		AIModelGPT51CodexMiniLabel:  defaultAIModelGPT51CodexMiniLabel,
		AIReasoningLowLabel:         defaultAIReasoningLowLabel,
		AIReasoningMediumLabel:      defaultAIReasoningMediumLabel,
		AIReasoningHighLabel:        defaultAIReasoningHighLabel,
		AIReasoningExtraHighLabel:   defaultAIReasoningExtraHighLabel,
	}
}

func normalizeRunAgentLabelCatalog(catalog runAgentLabelCatalog) runAgentLabelCatalog {
	def := defaultRunAgentLabelCatalog()
	catalog.AIModelGPT53CodexLabel = normalizeLabelCatalogValue(catalog.AIModelGPT53CodexLabel, def.AIModelGPT53CodexLabel)
	catalog.AIModelGPT53CodexSparkLabel = normalizeLabelCatalogValue(catalog.AIModelGPT53CodexSparkLabel, def.AIModelGPT53CodexSparkLabel)
	catalog.AIModelGPT52CodexLabel = normalizeLabelCatalogValue(catalog.AIModelGPT52CodexLabel, def.AIModelGPT52CodexLabel)
	catalog.AIModelGPT52Label = normalizeLabelCatalogValue(catalog.AIModelGPT52Label, def.AIModelGPT52Label)
	catalog.AIModelGPT51CodexMaxLabel = normalizeLabelCatalogValue(catalog.AIModelGPT51CodexMaxLabel, def.AIModelGPT51CodexMaxLabel)
	catalog.AIModelGPT51CodexMiniLabel = normalizeLabelCatalogValue(catalog.AIModelGPT51CodexMiniLabel, def.AIModelGPT51CodexMiniLabel)
	catalog.AIReasoningLowLabel = normalizeLabelCatalogValue(catalog.AIReasoningLowLabel, def.AIReasoningLowLabel)
	catalog.AIReasoningMediumLabel = normalizeLabelCatalogValue(catalog.AIReasoningMediumLabel, def.AIReasoningMediumLabel)
	catalog.AIReasoningHighLabel = normalizeLabelCatalogValue(catalog.AIReasoningHighLabel, def.AIReasoningHighLabel)
	catalog.AIReasoningExtraHighLabel = normalizeLabelCatalogValue(catalog.AIReasoningExtraHighLabel, def.AIReasoningExtraHighLabel)
	return catalog
}

func normalizeLabelCatalogValue(value string, fallback string) string {
	normalized := normalizeLabelToken(value)
	if normalized == "" {
		return fallback
	}
	return normalized
}

func runAgentLabelCatalogFromConfig(cfg Config) runAgentLabelCatalog {
	return normalizeRunAgentLabelCatalog(runAgentLabelCatalog{
		AIModelGPT53CodexLabel:      cfg.AIModelGPT53CodexLabel,
		AIModelGPT53CodexSparkLabel: cfg.AIModelGPT53CodexSparkLabel,
		AIModelGPT52CodexLabel:      cfg.AIModelGPT52CodexLabel,
		AIModelGPT52Label:           cfg.AIModelGPT52Label,
		AIModelGPT51CodexMaxLabel:   cfg.AIModelGPT51CodexMaxLabel,
		AIModelGPT51CodexMiniLabel:  cfg.AIModelGPT51CodexMiniLabel,
		AIReasoningLowLabel:         cfg.AIReasoningLowLabel,
		AIReasoningMediumLabel:      cfg.AIReasoningMediumLabel,
		AIReasoningHighLabel:        cfg.AIReasoningHighLabel,
		AIReasoningExtraHighLabel:   cfg.AIReasoningExtraHighLabel,
	})
}

func (c runAgentLabelCatalog) modelByLabel() map[string]string {
	return map[string]string{
		c.AIModelGPT53CodexLabel:      modelGPT53Codex,
		c.AIModelGPT53CodexSparkLabel: modelGPT53CodexSpark,
		c.AIModelGPT52CodexLabel:      modelGPT52Codex,
		c.AIModelGPT52Label:           modelGPT52,
		c.AIModelGPT51CodexMaxLabel:   modelGPT51CodexMax,
		c.AIModelGPT51CodexMiniLabel:  modelGPT51CodexMini,
	}
}

func (c runAgentLabelCatalog) reasoningByLabel() map[string]string {
	return map[string]string{
		c.AIReasoningLowLabel:       reasoningEffortLow,
		c.AIReasoningMediumLabel:    reasoningEffortMedium,
		c.AIReasoningHighLabel:      reasoningEffortHigh,
		c.AIReasoningExtraHighLabel: reasoningEffortExtraHigh,
	}
}
