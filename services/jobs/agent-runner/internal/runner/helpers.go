package runner

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

//go:embed templates/*.tmpl
var runnerTemplates embed.FS

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

func (s *Service) renderTaskTemplate(templateKind string) (string, error) {
	templateName := templateNamePromptWork
	if templateKind == promptTemplateKindReview {
		templateName = templateNamePromptReview
	}

	return renderTemplate(templateName, promptTaskTemplateData{
		BaseBranch:   s.cfg.AgentBaseBranch,
		PromptLocale: normalizePromptLocale(s.cfg.PromptTemplateLocale),
	})
}

func (s *Service) writeCodexConfig(codexDir string) error {
	hasContext7 := strings.TrimSpace(os.Getenv(envContext7APIKey)) != ""
	content, err := renderTemplate(templateNameCodexConfig, codexConfigTemplateData{
		Model:           s.cfg.AgentModel,
		ReasoningEffort: s.cfg.AgentReasoningEffort,
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

func (s *Service) writeCodexAuthFile(codexDir string) (bool, error) {
	authContent := strings.TrimSpace(s.cfg.OpenAIAuthFile)
	if authContent == "" {
		return false, nil
	}
	if !json.Valid([]byte(authContent)) {
		return false, fmt.Errorf("CODEXK8S_OPENAI_AUTH_FILE must be valid JSON")
	}

	authPath := filepath.Join(codexDir, "auth.json")
	if err := os.WriteFile(authPath, []byte(authContent), 0o600); err != nil {
		return false, fmt.Errorf("write codex auth file: %w", err)
	}
	return true, nil
}

func (s *Service) buildPrompt(taskBody string, result runResult) (string, error) {
	hasContext7 := strings.TrimSpace(os.Getenv(envContext7APIKey)) != ""
	runtimeMode := normalizeRuntimeMode(s.cfg.RuntimeMode)
	isReviseTrigger := webhookdomain.IsReviseTriggerKind(webhookdomain.NormalizeTriggerKind(result.triggerKind))
	return renderTemplate(templateNamePromptEnvelope, promptEnvelopeTemplateData{
		RepositoryFullName: s.cfg.RepositoryFullName,
		RunID:              s.cfg.RunID,
		IssueNumber:        s.cfg.IssueNumber,
		AgentKey:           s.cfg.AgentKey,
		RuntimeMode:        runtimeMode,
		IsFullEnv:          runtimeMode == runtimeModeFullEnv,
		TargetBranch:       result.targetBranch,
		BaseBranch:         s.cfg.AgentBaseBranch,
		TriggerKind:        result.triggerKind,
		IsReviseTrigger:    isReviseTrigger,
		HasExistingPR:      isReviseTrigger && result.existingPRNumber > 0,
		ExistingPRNumber:   result.existingPRNumber,
		TriggerLabel:       strings.TrimSpace(s.cfg.TriggerLabel),
		StateInReviewLabel: strings.TrimSpace(s.cfg.StateInReviewLabel),
		HasContext7:        hasContext7,
		PromptLocale:       normalizePromptLocale(s.cfg.PromptTemplateLocale),
		TaskBody:           taskBody,
	})
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
	if webhookdomain.IsReviseTriggerKind(webhookdomain.NormalizeTriggerKind(triggerKind)) {
		return promptTemplateKindReview
	}
	if strings.EqualFold(strings.TrimSpace(value), promptTemplateKindReview) {
		return promptTemplateKindReview
	}
	return promptTemplateKindWork
}

func buildTargetBranch(explicitBranch string, runID string, issueNumber int64) string {
	trimmedExplicit := strings.TrimSpace(explicitBranch)
	if trimmedExplicit != "" {
		return trimmedExplicit
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
	payload := sessionLogSnapshot{
		Version: sessionLogVersionV1,
		Status:  strings.TrimSpace(status),
		Report:  result.report,
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
