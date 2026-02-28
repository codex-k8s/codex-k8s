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

//go:embed templates/*.tmpl
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

func promptCommunicationLanguage(value string) string {
	if normalizePromptLocale(value) == promptLocaleRU {
		return "русский"
	}
	return "English"
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
	if normalizePromptTemplateKind(templateKind) == promptTemplateKindRevise {
		templateName = templateNamePromptRevise
	}
	return renderTemplate(templateName, templateData)
}

func (s *Service) writeCodexConfig(codexDir string, model string, reasoningEffort string) error {
	hasContext7 := strings.TrimSpace(os.Getenv(envContext7APIKey)) != ""
	content, err := renderTemplate(templateNameCodexConfig, codexConfigTemplateData{
		Model:           strings.TrimSpace(model),
		ReasoningEffort: strings.TrimSpace(reasoningEffort),
		MCPBaseURL:      s.cfg.MCPBaseURL,
		HasContext7:     hasContext7,
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
	isReviseTrigger := webhookdomain.IsReviseTriggerKind(webhookdomain.NormalizeTriggerKind(result.triggerKind))
	roleDisplayName, roleCapabilities := resolvePromptRoleProfile(s.cfg.AgentKey)
	projectDocs, docsTotal, docsTrimmed := loadProjectDocsForPrompt(repoDir, s.cfg.AgentKey, result.triggerKind, runtimeMode)
	roleDocTemplates, roleTemplatesTotal, roleTemplatesTrimmed := loadRoleDocTemplatesForPrompt(repoDir, s.cfg.AgentKey, result.triggerKind, runtimeMode)
	return renderTemplate(templateNamePromptEnvelope, promptEnvelopeTemplateData{
		RepositoryFullName:           s.cfg.RepositoryFullName,
		RunID:                        s.cfg.RunID,
		IssueNumber:                  s.cfg.IssueNumber,
		AgentKey:                     s.cfg.AgentKey,
		RoleDisplayName:              roleDisplayName,
		RoleCapabilities:             roleCapabilities,
		RuntimeMode:                  runtimeMode,
		IsFullEnv:                    runtimeMode == runtimeModeFullEnv,
		TargetBranch:                 result.targetBranch,
		BaseBranch:                   s.cfg.AgentBaseBranch,
		TriggerKind:                  result.triggerKind,
		IsAIRepairMainDirect:         isAIRepairMainDirectTrigger(result.triggerKind),
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
		CommunicationLanguage:        promptCommunicationLanguage(s.cfg.PromptTemplateLocale),
		ProjectDocs:                  projectDocs,
		ProjectDocsTotal:             docsTotal,
		ProjectDocsTrimmed:           docsTrimmed,
		RoleDocTemplates:             roleDocTemplates,
		RoleDocTemplatesTotal:        roleTemplatesTotal,
		RoleDocTemplatesTrimmed:      roleTemplatesTrimmed,
		TaskBody:                     taskBody,
	})
}

func resolvePromptRoleProfile(agentKey string) (string, []string) {
	key := strings.ToLower(strings.TrimSpace(agentKey))
	profiles := map[string]struct {
		name         string
		capabilities []string
	}{
		"dev": {
			name: "Developer",
			capabilities: []string{
				"Реализация изменений в коде и миграциях",
				"Запуск тестов и исправление регрессий",
				"Обновление документации при изменении поведения",
			},
		},
		"pm": {
			name: "Product Manager",
			capabilities: []string{
				"Уточнение продуктовых требований и критериев готовности",
				"Декомпозиция работ на реализуемые инкременты",
				"Контроль трассируемости изменений по этапам",
			},
		},
		"sa": {
			name: "Solution Architect",
			capabilities: []string{
				"Проектирование сервисных границ и контрактов",
				"Анализ архитектурных рисков и компромиссов",
				"Контроль соответствия кодовой базы архитектурным стандартам",
			},
		},
		"em": {
			name: "Engineering Manager",
			capabilities: []string{
				"Планирование исполнения и синхронизация командных задач",
				"Контроль quality-gates и критериев завершения",
				"Управление handover между ролями",
			},
		},
		"reviewer": {
			name: "Reviewer",
			capabilities: []string{
				"Поиск багов, рисков и регрессий в PR",
				"Проверка полноты тестового покрытия",
				"Проверка консистентности кода и документации",
			},
		},
		"qa": {
			name: "QA",
			capabilities: []string{
				"Проектирование тест-кейсов и edge-case сценариев",
				"Воспроизведение дефектов и верификация исправлений",
				"Регрессионные проверки критических пользовательских потоков",
			},
		},
		"sre": {
			name: "SRE",
			capabilities: []string{
				"Диагностика runtime/deploy проблем в Kubernetes",
				"Оценка надежности и эксплуатационных рисков",
				"Стабилизация и hardening инфраструктурных сценариев",
			},
		},
		"km": {
			name: "Knowledge Manager",
			capabilities: []string{
				"Поддержка актуальности docset и эксплуатационной документации",
				"Эволюция prompt templates и операционных инструкций",
				"Сбор evidence для self-improve цикла",
			},
		},
	}

	if profile, ok := profiles[key]; ok {
		return profile.name, profile.capabilities
	}
	if key == "" {
		return "Developer", profiles["dev"].capabilities
	}
	return strings.ToUpper(key), []string{"Следовать контракту задачи и проектным стандартам"}
}
