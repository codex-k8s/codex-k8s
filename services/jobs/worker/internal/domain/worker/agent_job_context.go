package worker

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

const (
	promptTemplateKindWork   = "work"
	promptTemplateKindRevise = "revise"
	promptTemplateSourceSeed = "repo_seed"

	modelSourceDefault          = "agent_default"
	modelSourceIssueLabel       = "issue_label"
	modelSourcePullRequestLabel = "pull_request_label"
	modelSourceLastRunContext   = "last_run_context"
	modelSourceFallback         = "auth_file_fallback"

	modelGPT53Codex      = "gpt-5.3-codex"
	modelGPT53CodexSpark = "gpt-5.3-codex-spark"
	modelGPT52Codex      = "gpt-5.2-codex"
	modelGPT52           = "gpt-5.2"
	modelGPT51CodexMax   = "gpt-5.1-codex-max"
	modelGPT51CodexMini  = "gpt-5.1-codex-mini"
)

type runAgentTarget struct {
	RepositoryFullName string
	AgentKey           string
	IssueNumber        int64
	TargetBranch       string
	ExistingPRNumber   int
	TriggerKind        string
	TriggerLabel       string
	AgentDisplayName   string
}

type runAgentPromptContext struct {
	PromptTemplateKind   string
	PromptTemplateSource string
	PromptTemplateLocale string
}

type runAgentModelContext struct {
	Model           string
	ModelSource     string
	ReasoningEffort string
	ReasoningSource string
}

type runAgentContext struct {
	runAgentTarget
	runAgentPromptContext
	runAgentModelContext
}

type runAgentPayload struct {
	Repository   *runAgentRepository   `json:"repository"`
	Issue        *runAgentIssue        `json:"issue"`
	Trigger      *runAgentTrigger      `json:"trigger"`
	Agent        *runAgentDescriptor   `json:"agent"`
	ProfileHints *runAgentProfileHints `json:"profile_hints"`
	RawPayload   json.RawMessage       `json:"raw_payload"`
}

type runAgentRepository struct {
	FullName string `json:"full_name"`
}

type runAgentIssue struct {
	Number int64 `json:"number"`
}

type runAgentTrigger struct {
	Source string `json:"source"`
	Kind   string `json:"kind"`
	Label  string `json:"label"`
}

type runAgentProfileHints struct {
	LastRunIssueLabels       []string `json:"last_run_issue_labels"`
	LastRunPullRequestLabels []string `json:"last_run_pull_request_labels"`
}

type runAgentDescriptor struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

func resolveRunAgentContext(runPayload json.RawMessage, defaults runAgentDefaults) (runAgentContext, error) {
	payload := parseRunAgentPayload(runPayload)

	ctx := runAgentContext{
		runAgentTarget: runAgentTarget{
			RepositoryFullName: strings.TrimSpace(payload.repositoryFullName),
			AgentKey:           strings.TrimSpace(payload.agentKey),
			IssueNumber:        payload.issueNumber,
			TargetBranch:       strings.TrimSpace(payload.targetBranch),
			ExistingPRNumber:   payload.existingPRNumber,
			TriggerKind:        normalizeTriggerKind(payload.triggerKind),
			TriggerLabel:       strings.TrimSpace(payload.triggerLabel),
			AgentDisplayName:   strings.TrimSpace(payload.agentDisplayName),
		},
		runAgentPromptContext: runAgentPromptContext{
			PromptTemplateKind:   promptTemplateKindWork,
			PromptTemplateSource: promptTemplateSourceSeed,
			PromptTemplateLocale: func() string {
				locale := strings.TrimSpace(defaults.DefaultLocale)
				if locale == "" {
					return "ru"
				}
				return locale
			}(),
		},
		runAgentModelContext: runAgentModelContext{
			Model:           defaults.DefaultModel,
			ModelSource:     modelSourceDefault,
			ReasoningEffort: defaults.DefaultReasoningEffort,
			ReasoningSource: modelSourceDefault,
		},
	}
	if resolvePromptTemplateKindForTrigger(ctx.TriggerKind) == promptTemplateKindRevise {
		ctx.PromptTemplateKind = promptTemplateKindRevise
	}
	if ctx.AgentKey == "" {
		return runAgentContext{}, fmt.Errorf("failed_precondition: run payload missing agent.key")
	}
	if ctx.AgentDisplayName == "" {
		return runAgentContext{}, fmt.Errorf("failed_precondition: run payload missing agent.name")
	}

	labelCatalog := normalizeRunAgentLabelCatalog(defaults.LabelCatalog)
	reviewDrivenRevise := strings.EqualFold(strings.TrimSpace(payload.triggerSource), webhookdomain.TriggerSourcePullRequestReview) &&
		resolvePromptTemplateKindForTrigger(ctx.TriggerKind) == promptTemplateKindRevise
	model, modelSource, err := resolveModelFromLabelsWithPriorityAndCatalog(
		payload.issueLabels,
		payload.pullRequestLabels,
		payload.historyIssueLabels,
		payload.historyPRLabels,
		defaults.DefaultModel,
		labelCatalog,
		reviewDrivenRevise,
	)
	if err != nil {
		return runAgentContext{}, err
	}
	reasoning, reasoningSource, err := resolveReasoningFromLabelsWithPriorityAndCatalog(
		payload.issueLabels,
		payload.pullRequestLabels,
		payload.historyIssueLabels,
		payload.historyPRLabels,
		defaults.DefaultReasoningEffort,
		labelCatalog,
		reviewDrivenRevise,
	)
	if err != nil {
		return runAgentContext{}, err
	}
	ctx.Model = model
	ctx.ModelSource = modelSource
	if !defaults.AllowGPT53 && isGPT53Model(ctx.Model) {
		ctx.Model = modelGPT52Codex
		ctx.ModelSource = modelSourceFallback
	}
	ctx.ReasoningEffort = reasoning
	ctx.ReasoningSource = reasoningSource

	return ctx, nil
}

func resolvePromptTemplateKindForTrigger(triggerKind string) string {
	normalized := webhookdomain.NormalizeTriggerKind(triggerKind)
	if webhookdomain.IsReviseTriggerKind(normalized) {
		return promptTemplateKindRevise
	}
	return promptTemplateKindWork
}

type runAgentDefaults struct {
	DefaultModel           string
	DefaultReasoningEffort string
	DefaultLocale          string
	AllowGPT53             bool
	LabelCatalog           runAgentLabelCatalog
}

type parsedRunAgentPayload struct {
	repositoryFullName string
	agentKey           string
	issueNumber        int64
	targetBranch       string
	existingPRNumber   int
	triggerSource      string
	triggerKind        string
	triggerLabel       string
	agentDisplayName   string
	issueLabels        []string
	pullRequestLabels  []string
	historyIssueLabels []string
	historyPRLabels    []string
}

func parseRunAgentPayload(raw json.RawMessage) parsedRunAgentPayload {
	var payload runAgentPayload
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &payload)
	}

	out := parsedRunAgentPayload{
		issueLabels:        make([]string, 0, 4),
		pullRequestLabels:  make([]string, 0, 4),
		historyIssueLabels: make([]string, 0, 4),
		historyPRLabels:    make([]string, 0, 4),
	}
	if payload.Repository != nil {
		out.repositoryFullName = strings.TrimSpace(payload.Repository.FullName)
	}
	if payload.Issue != nil && payload.Issue.Number > 0 {
		out.issueNumber = payload.Issue.Number
	}
	prNumber, targetBranch := extractPullRequestHints(payload.RawPayload)
	if out.issueNumber <= 0 && prNumber > 0 {
		out.issueNumber = prNumber
	}
	if prNumber > 0 {
		out.existingPRNumber = int(prNumber)
	}
	out.targetBranch = strings.TrimSpace(targetBranch)
	if payload.Trigger != nil {
		out.triggerSource = strings.TrimSpace(payload.Trigger.Source)
		out.triggerKind = strings.TrimSpace(payload.Trigger.Kind)
		out.triggerLabel = strings.TrimSpace(payload.Trigger.Label)
	}
	if payload.Agent != nil {
		out.agentKey = strings.TrimSpace(payload.Agent.Key)
		out.agentDisplayName = strings.TrimSpace(payload.Agent.Name)
	}
	if payload.ProfileHints != nil {
		for _, raw := range payload.ProfileHints.LastRunIssueLabels {
			label := strings.TrimSpace(raw)
			if label == "" {
				continue
			}
			out.historyIssueLabels = append(out.historyIssueLabels, label)
		}
		for _, raw := range payload.ProfileHints.LastRunPullRequestLabels {
			label := strings.TrimSpace(raw)
			if label == "" {
				continue
			}
			out.historyPRLabels = append(out.historyPRLabels, label)
		}
	}
	out.issueLabels, out.pullRequestLabels = extractIssueAndPullRequestLabels(payload.RawPayload)
	return out
}

type pullRequestHintsPayload struct {
	PullRequest *pullRequestHintsItem `json:"pull_request"`
}

type pullRequestHintsItem struct {
	Number int64                 `json:"number"`
	Head   *pullRequestHintsHead `json:"head"`
}

type pullRequestHintsHead struct {
	Ref string `json:"ref"`
}

func extractPullRequestHints(raw json.RawMessage) (number int64, targetBranch string) {
	if len(raw) == 0 {
		return 0, ""
	}

	var payload pullRequestHintsPayload
	if err := json.Unmarshal(raw, &payload); err != nil || payload.PullRequest == nil {
		return 0, ""
	}

	number = payload.PullRequest.Number
	if payload.PullRequest.Head != nil {
		targetBranch = strings.TrimSpace(payload.PullRequest.Head.Ref)
	}
	return number, targetBranch
}

func normalizeTriggerKind(value string) string {
	return string(webhookdomain.NormalizeTriggerKind(value))
}

func resolveModelFromLabels(labels []string, defaultModel string) (model string, source string, err error) {
	return resolveModelFromLabelsAndCatalog(labels, defaultModel, defaultRunAgentLabelCatalog())
}

func resolveModelFromLabelsAndCatalog(labels []string, defaultModel string, labelCatalog runAgentLabelCatalog) (model string, source string, err error) {
	return resolveSingleLabelValue(labels, defaultModel, labelCatalog.modelByLabel(), labelKindAIModel)
}

func resolveModelFromLabelsWithPriorityAndCatalog(
	issueLabels []string,
	pullRequestLabels []string,
	historyIssueLabels []string,
	historyPullRequestLabels []string,
	defaultModel string,
	labelCatalog runAgentLabelCatalog,
	reviewDrivenRevise bool,
) (model string, source string, err error) {
	return resolveSingleLabelValueWithPriorityAndHistory(
		issueLabels,
		pullRequestLabels,
		historyIssueLabels,
		historyPullRequestLabels,
		defaultModel,
		labelCatalog.modelByLabel(),
		labelKindAIModel,
		reviewDrivenRevise,
	)
}

func resolveReasoningFromLabelsWithPriorityAndCatalog(
	issueLabels []string,
	pullRequestLabels []string,
	historyIssueLabels []string,
	historyPullRequestLabels []string,
	defaultReasoning string,
	labelCatalog runAgentLabelCatalog,
	reviewDrivenRevise bool,
) (reasoning string, source string, err error) {
	return resolveSingleLabelValueWithPriorityAndHistory(
		issueLabels,
		pullRequestLabels,
		historyIssueLabels,
		historyPullRequestLabels,
		defaultReasoning,
		labelCatalog.reasoningByLabel(),
		labelKindAIReasoning,
		reviewDrivenRevise,
	)
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

func resolveSingleLabelValueWithPriorityAndHistory(
	issueLabels []string,
	pullRequestLabels []string,
	historyIssueLabels []string,
	historyPullRequestLabels []string,
	defaultValue string,
	known map[string]string,
	labelKind string,
	reviewDrivenRevise bool,
) (value string, source string, err error) {
	primaryLabels := pullRequestLabels
	primaryScope := "pull_request"
	primarySource := modelSourcePullRequestLabel
	secondaryLabels := issueLabels
	secondaryScope := "issue"
	secondarySource := modelSourceIssueLabel
	if reviewDrivenRevise {
		primaryLabels = issueLabels
		primaryScope = "issue"
		primarySource = modelSourceIssueLabel
		secondaryLabels = pullRequestLabels
		secondaryScope = "pull_request"
		secondarySource = modelSourcePullRequestLabel
	}

	primaryMatches := collectResolvedLabelValues(primaryLabels, known)
	if len(primaryMatches) > 1 {
		return "", "", fmt.Errorf("failed_precondition: multiple %s labels found on %s: %s", labelKind, primaryScope, strings.Join(primaryMatches, ", "))
	}
	if len(primaryMatches) == 1 {
		return known[primaryMatches[0]], primarySource, nil
	}

	secondaryMatches := collectResolvedLabelValues(secondaryLabels, known)
	if len(secondaryMatches) > 1 {
		return "", "", fmt.Errorf("failed_precondition: multiple %s labels found on %s: %s", labelKind, secondaryScope, strings.Join(secondaryMatches, ", "))
	}
	if len(secondaryMatches) == 1 {
		return known[secondaryMatches[0]], secondarySource, nil
	}

	historyIssueMatches := collectResolvedLabelValues(historyIssueLabels, known)
	if len(historyIssueMatches) > 1 {
		return "", "", fmt.Errorf("failed_precondition: multiple %s labels found in last_run_context(issue): %s", labelKind, strings.Join(historyIssueMatches, ", "))
	}
	if len(historyIssueMatches) == 1 {
		return known[historyIssueMatches[0]], modelSourceLastRunContext, nil
	}

	historyPRMatches := collectResolvedLabelValues(historyPullRequestLabels, known)
	if len(historyPRMatches) > 1 {
		return "", "", fmt.Errorf("failed_precondition: multiple %s labels found in last_run_context(pull_request): %s", labelKind, strings.Join(historyPRMatches, ", "))
	}
	if len(historyPRMatches) == 1 {
		return known[historyPRMatches[0]], modelSourceLastRunContext, nil
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

func isGPT53Model(model string) bool {
	normalizedModel := strings.TrimSpace(model)
	return strings.EqualFold(normalizedModel, modelGPT53Codex) || strings.EqualFold(normalizedModel, modelGPT53CodexSpark)
}
