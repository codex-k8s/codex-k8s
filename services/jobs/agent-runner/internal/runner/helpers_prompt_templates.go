package runner

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

//go:embed templates/*.tmpl templates/prompt_blocks/*.tmpl
var runnerTemplates embed.FS

//go:embed promptseeds/*.md
var promptSeedsFS embed.FS

func renderTemplate(templateName string, data any) (string, error) {
	tplBytes, err := runnerTemplates.ReadFile(templateName)
	if err != nil {
		return "", fmt.Errorf("read embedded template %s: %w", templateName, err)
	}

	tpl, err := template.New(filepath.Base(templateName)).Option("missingkey=error").Parse(string(tplBytes))
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", templateName, err)
	}

	var out strings.Builder
	if err := tpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("render template %s: %w", templateName, err)
	}
	return out.String(), nil
}

func normalizePromptLocale(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch {
	case strings.HasPrefix(normalized, promptLocaleRU):
		return promptLocaleRU
	case strings.HasPrefix(normalized, promptLocaleEN):
		return promptLocaleEN
	default:
		return promptLocaleEN
	}
}

func (s *Service) renderTaskTemplate(templateKind string, repoDir string) (string, error) {
	templateData := promptTaskTemplateData{
		BaseBranch:   s.cfg.AgentBaseBranch,
		PromptLocale: normalizePromptLocale(s.cfg.PromptTemplateLocale),
	}
	_ = repoDir

	for _, candidate := range promptSeedCandidates(s.cfg.AgentKey, s.cfg.TriggerKind, templateKind, s.cfg.PromptTemplateLocale) {
		seedPath := filepath.Join(promptSeedsDirRelativePath, candidate)
		seedBytes, err := promptSeedsFS.ReadFile(seedPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", fmt.Errorf("read prompt seed %s: %w", seedPath, err)
		}
		seedTemplate, err := template.New(candidate).Option("missingkey=error").Parse(string(seedBytes))
		if err != nil {
			return "", fmt.Errorf("parse prompt seed %s: %w", seedPath, err)
		}
		var out strings.Builder
		if err := seedTemplate.Execute(&out, templateData); err != nil {
			return "", fmt.Errorf("render prompt seed %s: %w", seedPath, err)
		}
		return out.String(), nil
	}

	templateName := templateNamePromptWork
	switch normalizePromptTemplateKind(templateKind) {
	case promptTemplateKindRevise:
		templateName = templateNamePromptRevise
	case promptTemplateKindDiscussion:
		templateName = templateNamePromptDiscussion
	}
	return renderTemplate(templateName, templateData)
}

func (s *Service) writeCodexConfig(codexDir string, model string, reasoningEffort string) error {
	context7APIKey := strings.TrimSpace(os.Getenv(envContext7APIKey))
	hasContext7 := context7APIKey != ""
	content, err := renderTemplate(templateNameCodexConfig, codexConfigTemplateData{
		Model:           strings.TrimSpace(model),
		ReasoningEffort: strings.TrimSpace(reasoningEffort),
		MCPBaseURL:      s.cfg.MCPBaseURL,
		HasContext7:     hasContext7,
		Context7APIKey:  context7APIKey,
	})
	if err != nil {
		return err
	}

	configPath := filepath.Join(codexDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write codex config: %w", err)
	}
	return nil
}

func (s *Service) buildPrompt(taskBody string, result runResult, repoDir string) (string, error) {
	hasContext7 := strings.TrimSpace(os.Getenv(envContext7APIKey)) != ""
	runtimeMode := normalizeRuntimeMode(s.cfg.RuntimeMode)
	runtimeTargetEnv := normalizePromptRuntimeTargetEnv(s.cfg.RuntimeTargetEnv)
	runtimeBuildRef := strings.TrimSpace(s.cfg.RuntimeBuildRef)
	runtimeAccessProfile := normalizePromptRuntimeAccessProfile(s.cfg.RuntimeAccessProfile)
	isDiscussionMode := s.cfg.DiscussionMode
	resumePromptBlock, err := buildInteractionResumePromptBlock(
		s.cfg.PromptTemplateLocale,
		s.cfg.InteractionResumePayload,
		result.restoredSessionPath != "" || result.sessionID != "",
	)
	if err != nil {
		return "", fmt.Errorf("build interaction resume prompt block: %w", err)
	}
	isReviseTrigger := !isDiscussionMode && webhookdomain.IsReviseTriggerKind(webhookdomain.NormalizeTriggerKind(result.triggerKind))
	roleProfileBlock, err := renderPromptRoleProfileBlock(s.cfg.AgentKey, s.cfg.PromptTemplateLocale)
	if err != nil {
		return "", fmt.Errorf("render prompt role profile: %w", err)
	}
	issueContractBlock, prContractBlock, err := renderPromptArtifactContractBlocks(s.cfg.RepositoryFullName, s.cfg.IssueNumber, s.cfg.AgentKey, result.triggerKind, result.templateKind, s.cfg.PromptTemplateLocale)
	if err != nil {
		return "", fmt.Errorf("render prompt artifact contracts: %w", err)
	}
	projectDocs, docsTotal, docsTrimmed := loadProjectDocsForPrompt(repoDir, s.cfg.AgentKey, result.triggerKind, runtimeMode, runtimeTargetEnv)
	roleDocTemplates, roleTemplatesTotal, roleTemplatesTrimmed := loadRoleDocTemplatesForPrompt(repoDir, s.cfg.AgentKey, result.triggerKind, runtimeMode, runtimeTargetEnv)
	prompt, err := renderTemplate(templateNamePromptEnvelope, promptEnvelopeTemplateData{
		RepositoryFullName:           s.cfg.RepositoryFullName,
		RunID:                        s.cfg.RunID,
		IssueNumber:                  s.cfg.IssueNumber,
		AgentKey:                     s.cfg.AgentKey,
		RuntimeMode:                  runtimeMode,
		RuntimeTargetEnv:             runtimeTargetEnv,
		RuntimeBuildRef:              runtimeBuildRef,
		RuntimeAccessProfile:         runtimeAccessProfile,
		RuntimeRestrictions:          promptRuntimeRestrictions(runtimeAccessProfile, normalizePromptLocale(s.cfg.PromptTemplateLocale)),
		IsFullEnv:                    runtimeMode == runtimeModeFullEnv,
		TargetBranch:                 result.targetBranch,
		BaseBranch:                   s.cfg.AgentBaseBranch,
		TriggerKind:                  result.triggerKind,
		IsAIRepairMainDirect:         isAIRepairMainDirectTrigger(result.triggerKind),
		IsDiscussionMode:             isDiscussionMode,
		IsReviseTrigger:              isReviseTrigger,
		IsMarkdownDocsOnlyScope:      isMarkdownOnlyScope(result.triggerKind, s.cfg.AgentKey),
		IsReviewerCommentOnlyScope:   isReviewerCommentOnlyScope(result.triggerKind, s.cfg.AgentKey),
		IsSelfImproveRestrictedScope: isSelfImproveRestrictedScope(result.triggerKind, s.cfg.AgentKey),
		HasExistingPR:                isReviseTrigger && result.existingPRNumber > 0,
		ExistingPRNumber:             result.existingPRNumber,
		TriggerLabel:                 strings.TrimSpace(s.cfg.TriggerLabel),
		StateInReviewLabel:           strings.TrimSpace(s.cfg.StateInReviewLabel),
		HasContext7:                  hasContext7,
		PromptLocale:                 normalizePromptLocale(s.cfg.PromptTemplateLocale),
		RoleProfileBlock:             roleProfileBlock,
		IssueContractBlock:           issueContractBlock,
		PRContractBlock:              prContractBlock,
		ProjectDocs:                  projectDocs,
		ProjectDocsTotal:             docsTotal,
		ProjectDocsTrimmed:           docsTrimmed,
		RoleDocTemplates:             roleDocTemplates,
		RoleDocTemplatesTotal:        roleTemplatesTotal,
		RoleDocTemplatesTrimmed:      roleTemplatesTrimmed,
		TaskBody:                     taskBody,
	})
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(resumePromptBlock) == "" {
		return prompt, nil
	}
	return resumePromptBlock + "\n\n" + prompt, nil
}

func normalizePromptRuntimeTargetEnv(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "ai", "production":
		return normalized
	default:
		return ""
	}
}

func normalizePromptRuntimeAccessProfile(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "candidate", "production-readonly":
		return normalized
	default:
		return ""
	}
}

func promptRuntimeRestrictions(accessProfile string, locale string) []string {
	switch normalizePromptRuntimeAccessProfile(accessProfile) {
	case "production-readonly":
		if locale == promptLocaleRU {
			return []string{
				"доступ только на чтение: `get/list/watch` + `pods/log` + events/status;",
				"`pods/exec`, `pods/port-forward`, mutations и доступ к `secrets` запрещены;",
				"работайте против production namespace и не пытайтесь менять runtime-ресурсы.",
			}
		}
		return []string{
			"read-only only: `get/list/watch` + `pods/log` + events/status;",
			"`pods/exec`, `pods/port-forward`, mutations, and `secrets` access are forbidden;",
			"operate against the production namespace and do not mutate runtime resources.",
		}
	case "candidate":
		if locale == promptLocaleRU {
			return []string{
				"работайте в candidate environment этой Issue/PR и сохраняйте continuity до merge;",
				"если нужно обновить candidate, используйте тот же namespace/build ref, который пришёл в runtime context.",
			}
		}
		return []string{
			"operate in the candidate environment for this issue/PR and preserve continuity until merge;",
			"when candidate refresh is needed, keep using the namespace/build ref from runtime context.",
		}
	default:
		return nil
	}
}
