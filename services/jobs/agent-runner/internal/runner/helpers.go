package runner

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"text/template"
	"time"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
)

//go:embed templates/*.tmpl
var runnerTemplates embed.FS

//go:embed promptseeds/*.md
var promptSeedsFS embed.FS

var (
	toolGapNotFoundQuotedPattern  = regexp.MustCompile(`['"]([a-zA-Z0-9._-]+)['"]:\s+executable file not found`)
	toolGapCommandNotFoundPattern = regexp.MustCompile(`(?m)(?:^|:\s)([a-zA-Z0-9._-]+):\s+command not found$`)
	toolGapMissingCommandPattern  = regexp.MustCompile(`(?m)missing (?:required )?command[:\s]+([a-zA-Z0-9._-]+)$`)
)

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

func loadProjectDocsForPrompt(repoDir string, roleKey string, triggerKind string, runtimeMode string) ([]promptProjectDocTemplateData, int, bool) {
	const maxPromptDocs = 16

	servicesPath := filepath.Join(strings.TrimSpace(repoDir), "services.yaml")
	raw, err := os.ReadFile(servicesPath)
	if err != nil {
		return nil, 0, false
	}

	env := resolvePromptDocsEnv(triggerKind, runtimeMode)
	loadResult, err := servicescfg.LoadFromYAML(raw, servicescfg.LoadOptions{Env: env})
	if err != nil || loadResult.Stack == nil {
		return nil, 0, false
	}
	if len(loadResult.Stack.Spec.ProjectDocs) == 0 {
		return nil, 0, false
	}

	normalizedRole := strings.ToLower(strings.TrimSpace(roleKey))
	items := make([]promptProjectDocTemplateData, 0, len(loadResult.Stack.Spec.ProjectDocs))
	for _, doc := range loadResult.Stack.Spec.ProjectDocs {
		path := strings.TrimSpace(doc.Path)
		if path == "" {
			continue
		}
		repository := strings.ToLower(strings.TrimSpace(doc.Repository))

		roles := make([]string, 0, len(doc.Roles))
		for _, rawRole := range doc.Roles {
			role := strings.ToLower(strings.TrimSpace(rawRole))
			if role == "" || slices.Contains(roles, role) {
				continue
			}
			roles = append(roles, role)
		}
		if len(roles) > 0 && normalizedRole != "" && !slices.Contains(roles, normalizedRole) {
			continue
		}

		items = append(items, promptProjectDocTemplateData{
			Repository:  repository,
			Path:        path,
			Description: strings.TrimSpace(doc.Description),
			Optional:    doc.Optional,
		})
	}
	if len(items) == 0 {
		return nil, 0, false
	}

	sort.SliceStable(items, func(i, j int) bool {
		left := items[i]
		right := items[j]
		leftPriority := promptDocsRepositoryPriority(left.Repository)
		rightPriority := promptDocsRepositoryPriority(right.Repository)
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		if left.Repository != right.Repository {
			return left.Repository < right.Repository
		}
		return left.Path < right.Path
	})

	deduped := make([]promptProjectDocTemplateData, 0, len(items))
	seenByPath := make(map[string]struct{}, len(items))
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.Path))
		if key == "" {
			continue
		}
		if _, exists := seenByPath[key]; exists {
			continue
		}
		seenByPath[key] = struct{}{}
		deduped = append(deduped, item)
	}
	items = deduped

	total := len(items)
	trimmed := false
	if total > maxPromptDocs {
		items = items[:maxPromptDocs]
		trimmed = true
	}
	return items, total, trimmed
}

func resolvePromptDocsEnv(triggerKind string, runtimeMode string) string {
	if strings.EqualFold(strings.TrimSpace(runtimeMode), runtimeModeFullEnv) {
		normalizedTrigger := webhookdomain.NormalizeTriggerKind(triggerKind)
		switch normalizedTrigger {
		case webhookdomain.TriggerKindDev,
			webhookdomain.TriggerKindDevRevise,
			webhookdomain.TriggerKindQA,
			webhookdomain.TriggerKindOps:
			return "ai"
		}
	}
	return "production"
}

func promptDocsRepositoryPriority(repository string) int {
	repository = strings.ToLower(strings.TrimSpace(repository))
	switch {
	case repository == "":
		return 2
	case strings.Contains(repository, "policy"), strings.Contains(repository, "docs"):
		return 0
	case strings.Contains(repository, "orchestrator"):
		return 1
	default:
		return 2
	}
}

func normalizeTriggerKind(value string) string {
	return string(webhookdomain.NormalizeTriggerKind(value))
}

func normalizeRuntimeMode(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), runtimeModeFullEnv) {
		return runtimeModeFullEnv
	}
	return runtimeModeCodeOnly
}

func normalizeTemplateKind(value string, triggerKind string) string {
	normalizedTrigger := webhookdomain.NormalizeTriggerKind(triggerKind)
	if webhookdomain.IsReviseTriggerKind(normalizedTrigger) {
		return promptTemplateKindRevise
	}
	if strings.EqualFold(strings.TrimSpace(value), promptTemplateKindRevise) {
		return promptTemplateKindRevise
	}
	return promptTemplateKindWork
}

func buildTargetBranch(explicitBranch string, runID string, issueNumber int64, triggerKind string, baseBranch string) string {
	trimmedExplicit := strings.TrimSpace(explicitBranch)
	if trimmedExplicit != "" {
		return trimmedExplicit
	}
	if isAIRepairMainDirectTrigger(triggerKind) {
		base := strings.TrimSpace(baseBranch)
		if base != "" {
			return base
		}
		return "main"
	}
	if issueNumber > 0 {
		return fmt.Sprintf("codex/issue-%d", issueNumber)
	}
	trimmedRunID := strings.TrimSpace(runID)
	if len(trimmedRunID) > 12 {
		trimmedRunID = trimmedRunID[:12]
	}
	return "codex/run-" + trimmedRunID
}

func isAIRepairMainDirectTrigger(triggerKind string) bool {
	return webhookdomain.NormalizeTriggerKind(strings.TrimSpace(triggerKind)) == webhookdomain.TriggerKindAIRepair
}

func optionalIssueNumber(value int64) *int {
	if value <= 0 {
		return nil
	}
	intValue := int(value)
	return &intValue
}

func optionalInt(value int) *int {
	if value <= 0 {
		return nil
	}
	intValue := value
	return &intValue
}

func readJSONFileOrNil(path string) json.RawMessage {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	if !json.Valid(bytes) {
		return nil
	}
	return json.RawMessage(bytes)
}

func latestSessionFile(sessionsDir string) string {
	type candidate struct {
		path string
		mod  time.Time
	}
	files := make([]candidate, 0, 4)

	_ = filepath.WalkDir(sessionsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(d.Name())) != ".json" {
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil {
			return nil
		}
		files = append(files, candidate{path: path, mod: info.ModTime()})
		return nil
	})
	if len(files) == 0 {
		return ""
	}
	sort.Slice(files, func(i, j int) bool { return files[i].mod.After(files[j].mod) })
	return files[0].path
}

func extractSessionIDFromFile(path string) string {
	type codexSessionIdentity struct {
		SessionID      string `json:"session_id"`
		ID             string `json:"id"`
		ConversationID string `json:"conversation_id"`
		ThreadID       string `json:"thread_id"`
	}

	bytes, err := os.ReadFile(path)
	if err != nil || !json.Valid(bytes) {
		return ""
	}

	var payload codexSessionIdentity
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return ""
	}

	for _, value := range []string{payload.SessionID, payload.ID, payload.ConversationID, payload.ThreadID} {
		stringValue := strings.TrimSpace(value)
		if stringValue != "" {
			return stringValue
		}
	}
	return ""
}

func runCommandQuiet(ctx context.Context, dir string, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	if strings.TrimSpace(dir) != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

func runCommandWithInput(ctx context.Context, input []byte, stdout io.Writer, stderr io.Writer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func runCommandCaptureOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdoutBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return stdoutBuffer.Bytes(), nil
}

func runCommandCaptureCombinedOutput(ctx context.Context, dir string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if strings.TrimSpace(dir) != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	return trimCapturedOutput(string(output), maxCapturedCommandOutput), err
}

func parseCodexReportOutput(output []byte) (codexReport, json.RawMessage, error) {
	trimmedOutput := strings.TrimSpace(string(output))
	if trimmedOutput == "" {
		return codexReport{}, nil, fmt.Errorf("empty codex output")
	}

	tryDecode := func(raw []byte) (codexReport, bool) {
		if !json.Valid(raw) {
			return codexReport{}, false
		}
		var report codexReport
		if err := json.Unmarshal(raw, &report); err != nil {
			return codexReport{}, false
		}
		return report, true
	}

	if report, ok := tryDecode([]byte(trimmedOutput)); ok {
		return report, json.RawMessage(trimmedOutput), nil
	}

	lines := strings.Split(trimmedOutput, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if report, ok := tryDecode([]byte(line)); ok {
			return report, json.RawMessage(line), nil
		}
	}

	return codexReport{}, nil, fmt.Errorf("failed to parse codex structured output")
}

func trimCapturedOutput(raw string, maxBytes int) string {
	trimmed := strings.TrimSpace(raw)
	if maxBytes <= 0 || len(trimmed) <= maxBytes {
		return trimmed
	}
	if maxBytes < len("...(truncated)") {
		return trimmed[:maxBytes]
	}
	cutoff := maxBytes - len("...(truncated)")
	return trimmed[:cutoff] + "...(truncated)"
}

func buildSessionLogJSON(result runResult, status string) json.RawMessage {
	report := result.report
	report.ActionItems = normalizeStringList(report.ActionItems)
	report.EvidenceRefs = normalizeStringList(report.EvidenceRefs)
	report.ToolGaps = normalizeStringList(result.toolGaps)
	payload := sessionLogSnapshot{
		Version: sessionLogVersionV1,
		Status:  strings.TrimSpace(status),
		Report:  report,
		Runtime: sessionRuntimeLogFields{
			TargetBranch:     strings.TrimSpace(result.targetBranch),
			CodexExecOutput:  strings.TrimSpace(result.codexExecOutput),
			GitPushOutput:    strings.TrimSpace(result.gitPushOutput),
			ExistingPRNumber: result.existingPRNumber,
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return json.RawMessage(raw)
}

func normalizeStringList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item == "" {
			continue
		}
		lower := strings.ToLower(item)
		if _, exists := seen[lower]; exists {
			continue
		}
		seen[lower] = struct{}{}
		normalized = append(normalized, item)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func detectToolGaps(report codexReport, outputs ...string) []string {
	candidates := make([]string, 0, len(report.ToolGaps)+4)
	candidates = append(candidates, report.ToolGaps...)
	for _, output := range outputs {
		trimmed := strings.TrimSpace(output)
		if trimmed == "" {
			continue
		}
		candidates = append(candidates, extractToolGapCandidates(trimmed)...)
	}
	return normalizeStringList(candidates)
}

func extractToolGapCandidates(output string) []string {
	candidates := make([]string, 0, 4)

	for _, match := range toolGapNotFoundQuotedPattern.FindAllStringSubmatch(output, -1) {
		if len(match) >= 2 {
			candidates = append(candidates, strings.TrimSpace(match[1]))
		}
	}
	for _, match := range toolGapCommandNotFoundPattern.FindAllStringSubmatch(output, -1) {
		if len(match) >= 2 {
			candidates = append(candidates, strings.TrimSpace(match[1]))
		}
	}
	for _, match := range toolGapMissingCommandPattern.FindAllStringSubmatch(strings.ToLower(output), -1) {
		if len(match) >= 2 {
			candidates = append(candidates, strings.TrimSpace(match[1]))
		}
	}

	return candidates
}
