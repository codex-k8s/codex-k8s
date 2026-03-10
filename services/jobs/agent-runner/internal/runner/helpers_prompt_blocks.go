package runner

import (
	"fmt"
	"strings"
)

type promptBlockTemplateData struct {
	AgentKey    string
	IssueNumber int64
	IssueURL    string
	StageName   string
}

func renderPromptRoleProfileBlock(agentKey string, locale string) (string, error) {
	return renderTemplate(promptRoleProfileTemplateName(locale), promptBlockTemplateData{
		AgentKey: normalizePromptBlockAgentKey(agentKey),
	})
}

func renderPromptArtifactContractBlocks(repoFullName string, issueNumber int64, agentKey string, triggerKind string, templateKind string, locale string) (string, string, error) {
	data := promptBlockTemplateData{
		AgentKey:    normalizePromptBlockAgentKey(agentKey),
		IssueNumber: issueNumber,
		IssueURL:    buildPromptIssueURL(repoFullName, issueNumber),
		StageName:   promptSeedStageByTriggerKind(triggerKind),
	}

	issueBlock, err := renderTemplate(promptIssueContractTemplateName(locale), data)
	if err != nil {
		return "", "", err
	}
	prBlock, err := renderTemplate(promptPRContractTemplateName(triggerKind, templateKind, locale), data)
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(issueBlock), strings.TrimSpace(prBlock), nil
}

func promptRoleProfileTemplateName(locale string) string {
	return fmt.Sprintf("templates/prompt_blocks/role_profile_%s.tmpl", normalizePromptLocale(locale))
}

func promptIssueContractTemplateName(locale string) string {
	return fmt.Sprintf("templates/prompt_blocks/issue_contract_%s.tmpl", normalizePromptLocale(locale))
}

func promptPRContractTemplateName(triggerKind string, templateKind string, locale string) string {
	if isAIRepairMainDirectTrigger(triggerKind) {
		return fmt.Sprintf("templates/prompt_blocks/pr_contract_main_direct_%s.tmpl", normalizePromptLocale(locale))
	}
	switch normalizePromptTemplateKind(templateKind) {
	case promptTemplateKindDiscussion:
		return fmt.Sprintf("templates/prompt_blocks/pr_contract_discussion_%s.tmpl", normalizePromptLocale(locale))
	case promptTemplateKindRevise:
		return fmt.Sprintf("templates/prompt_blocks/pr_contract_revise_%s.tmpl", normalizePromptLocale(locale))
	default:
		return fmt.Sprintf("templates/prompt_blocks/pr_contract_work_%s.tmpl", normalizePromptLocale(locale))
	}
}

func normalizePromptBlockAgentKey(agentKey string) string {
	switch strings.ToLower(strings.TrimSpace(agentKey)) {
	case "dev", "pm", "sa", "em", "reviewer", "qa", "sre", "km":
		return strings.ToLower(strings.TrimSpace(agentKey))
	default:
		return "default"
	}
}

func buildPromptIssueURL(repoFullName string, issueNumber int64) string {
	trimmedRepo := strings.TrimSpace(repoFullName)
	if trimmedRepo == "" || issueNumber <= 0 {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/issues/%d", trimmedRepo, issueNumber)
}
