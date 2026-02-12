package worker

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

const (
	triggerKindDev       = "dev"
	triggerKindDevRevise = "dev_revise"

	promptTemplateKindWork   = "work"
	promptTemplateKindReview = "review"
	promptTemplateSourceSeed = "repo_seed"

	modelSourceDefault    = "agent_default"
	modelSourceIssueLabel = "issue_label"
	modelSourceFallback   = "auth_file_fallback"

	modelGPT53Codex     = "gpt-5.3-codex"
	modelGPT52Codex     = "gpt-5.2-codex"
	modelGPT52          = "gpt-5.2"
	modelGPT51CodexMax  = "gpt-5.1-codex-max"
	modelGPT51CodexMini = "gpt-5.1-codex-mini"
)

type runAgentContext struct {
	RepositoryFullName   string
	AgentKey             string
	IssueNumber          int64
	TriggerKind          string
	TriggerLabel         string
	AgentDisplayName     string
	PromptTemplateKind   string
	PromptTemplateSource string
	PromptTemplateLocale string
	Model                string
	ModelSource          string
	ReasoningEffort      string
	ReasoningSource      string
}

type runAgentPayload struct {
	Repository *runAgentRepository `json:"repository"`
	Issue      *runAgentIssue      `json:"issue"`
	Trigger    *runAgentTrigger    `json:"trigger"`
	Agent      *runAgentDescriptor `json:"agent"`
	RawPayload json.RawMessage     `json:"raw_payload"`
}

type runAgentRepository struct {
	FullName string `json:"full_name"`
}

type runAgentIssue struct {
	Number int64 `json:"number"`
}

type runAgentTrigger struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

type runAgentDescriptor struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

func resolveRunAgentContext(runPayload json.RawMessage, defaults runAgentDefaults) (runAgentContext, error) {
	payload := parseRunAgentPayload(runPayload)

	ctx := runAgentContext{
		RepositoryFullName: strings.TrimSpace(payload.repositoryFullName),
		AgentKey:           strings.TrimSpace(payload.agentKey),
		IssueNumber:        payload.issueNumber,
		TriggerKind:        normalizeTriggerKind(payload.triggerKind),
		TriggerLabel:       strings.TrimSpace(payload.triggerLabel),
		AgentDisplayName:   strings.TrimSpace(payload.agentDisplayName),
		PromptTemplateKind: promptTemplateKindWork,
		Model:              defaults.DefaultModel,
		ModelSource:        modelSourceDefault,
		ReasoningEffort:    defaults.DefaultReasoningEffort,
		ReasoningSource:    modelSourceDefault,
		PromptTemplateLocale: func() string {
			locale := strings.TrimSpace(defaults.DefaultLocale)
			if locale == "" {
				return "ru"
			}
			return locale
		}(),
		PromptTemplateSource: promptTemplateSourceSeed,
	}
	if ctx.TriggerKind == triggerKindDevRevise {
		ctx.PromptTemplateKind = promptTemplateKindReview
	}
	if ctx.AgentKey == "" {
		return runAgentContext{}, fmt.Errorf("failed_precondition: run payload missing agent.key")
	}
	if ctx.AgentDisplayName == "" {
		return runAgentContext{}, fmt.Errorf("failed_precondition: run payload missing agent.name")
	}

	labels := payload.issueLabels
	model, modelSource, err := resolveModelFromLabels(labels, defaults.DefaultModel)
	if err != nil {
		return runAgentContext{}, err
	}
	reasoning, reasoningSource, err := resolveReasoningFromLabels(labels, defaults.DefaultReasoningEffort)
	if err != nil {
		return runAgentContext{}, err
	}
	ctx.Model = model
	ctx.ModelSource = modelSource
	if !defaults.AllowGPT53 && strings.EqualFold(ctx.Model, modelGPT53Codex) {
		ctx.Model = modelGPT52Codex
		ctx.ModelSource = modelSourceFallback
	}
	ctx.ReasoningEffort = reasoning
	ctx.ReasoningSource = reasoningSource

	return ctx, nil
}

type runAgentDefaults struct {
	DefaultModel           string
	DefaultReasoningEffort string
	DefaultLocale          string
	AllowGPT53             bool
}

type parsedRunAgentPayload struct {
	repositoryFullName string
	agentKey           string
	issueNumber        int64
	triggerKind        string
	triggerLabel       string
	agentDisplayName   string
	issueLabels        []string
}

func parseRunAgentPayload(raw json.RawMessage) parsedRunAgentPayload {
	var payload runAgentPayload
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &payload)
	}

	out := parsedRunAgentPayload{
		issueLabels: make([]string, 0, 4),
	}
	if payload.Repository != nil {
		out.repositoryFullName = strings.TrimSpace(payload.Repository.FullName)
	}
	if payload.Issue != nil && payload.Issue.Number > 0 {
		out.issueNumber = payload.Issue.Number
	}
	if payload.Trigger != nil {
		out.triggerKind = strings.TrimSpace(payload.Trigger.Kind)
		out.triggerLabel = strings.TrimSpace(payload.Trigger.Label)
	}
	if payload.Agent != nil {
		out.agentKey = strings.TrimSpace(payload.Agent.Key)
		out.agentDisplayName = strings.TrimSpace(payload.Agent.Name)
	}
	out.issueLabels = extractIssueLabels(payload.RawPayload)
	return out
}

func normalizeTriggerKind(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case triggerKindDevRevise:
		return triggerKindDevRevise
	default:
		return triggerKindDev
	}
}

func resolveModelFromLabels(labels []string, defaultModel string) (model string, source string, err error) {
	modelByLabel := map[string]string{
		"ai-model-gpt-5.3-codex":      modelGPT53Codex,
		"ai-model-gpt-5.2-codex":      modelGPT52Codex,
		"ai-model-gpt-5.2":            modelGPT52,
		"ai-model-gpt-5.1-codex-max":  modelGPT51CodexMax,
		"ai-model-gpt-5.1-codex-mini": modelGPT51CodexMini,
	}
	return resolveSingleLabelValue(labels, defaultModel, modelByLabel, "ai-model")
}

func resolveReasoningFromLabels(labels []string, defaultReasoning string) (reasoning string, source string, err error) {
	reasoningByLabel := map[string]string{
		"ai-reasoning-low":        "low",
		"ai-reasoning-medium":     "medium",
		"ai-reasoning-high":       "high",
		"ai-reasoning-extra-high": "high",
	}
	return resolveSingleLabelValue(labels, defaultReasoning, reasoningByLabel, "ai-reasoning")
}

func resolveSingleLabelValue(labels []string, defaultValue string, known map[string]string, labelKind string) (value string, source string, err error) {
	matches := collectResolvedLabelValues(labels, known)
	if len(matches) > 1 {
		return "", "", fmt.Errorf("failed_precondition: multiple %s labels found: %s", labelKind, strings.Join(matches, ", "))
	}
	if len(matches) == 1 {
		return known[matches[0]], modelSourceIssueLabel, nil
	}
	return defaultValue, modelSourceDefault, nil
}

func collectResolvedLabelValues(labels []string, known map[string]string) []string {
	found := make([]string, 0, 1)
	for _, rawLabel := range labels {
		normalized := normalizeLabelToken(rawLabel)
		if normalized == "" {
			continue
		}
		if _, ok := known[normalized]; !ok {
			continue
		}
		if !slices.Contains(found, normalized) {
			found = append(found, normalized)
		}
	}
	return found
}
