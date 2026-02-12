package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	cpclient "github.com/codex-k8s/codex-k8s/services/jobs/agent-runner/internal/controlplane"
)

const (
	templatePathWork   = "docs/product/prompt-seeds/dev-work.md"
	templatePathReview = "docs/product/prompt-seeds/dev-review.md"

	promptTemplateKindWork   = "work"
	promptTemplateKindReview = "review"

	triggerKindDev       = "dev"
	triggerKindDevRevise = "dev_revise"
)

// ExitError allows caller to map runner failures to process exit code.
type ExitError struct {
	ExitCode int
	Err      error
}

func (e ExitError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("runner failed with exit code %d", e.ExitCode)
	}
	return e.Err.Error()
}

func (e ExitError) Unwrap() error {
	return e.Err
}

// Config defines runtime parameters for one agent-runner job.
type Config struct {
	RunID              string
	CorrelationID      string
	ProjectID          string
	RepositoryFullName string
	AgentKey           string
	IssueNumber        int64

	TriggerKind          string
	PromptTemplateKind   string
	PromptTemplateSource string
	PromptTemplateLocale string
	AgentModel           string
	AgentReasoningEffort string
	AgentBaseBranch      string
	AgentDisplayName     string

	ControlPlaneGRPCTarget string
	MCPBaseURL             string
	MCPBearerToken         string

	GitBotToken    string
	GitBotUsername string
	GitBotMail     string
	OpenAIAPIKey   string
}

// ControlPlaneCallbacks defines required control-plane callbacks for runner lifecycle.
type ControlPlaneCallbacks interface {
	UpsertAgentSession(ctx context.Context, params cpclient.AgentSessionUpsertParams) error
	GetLatestAgentSession(ctx context.Context, query cpclient.LatestAgentSessionQuery) (cpclient.AgentSessionSnapshot, bool, error)
	InsertRunFlowEvent(ctx context.Context, runID string, eventType floweventdomain.EventType, payload json.RawMessage) error
}

// Service runs one codex-driven development/revise cycle.
type Service struct {
	cfg    Config
	cp     ControlPlaneCallbacks
	logger *slog.Logger
}

// NewService creates runner service.
func NewService(cfg Config, cp ControlPlaneCallbacks, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{cfg: cfg, cp: cp, logger: logger}
}

// Run executes full runner flow and returns nil only on successful PR update/create path.
func (s *Service) Run(ctx context.Context) (err error) {
	homeDir := strings.TrimSpace(os.Getenv("HOME"))
	if homeDir == "" {
		homeDir = "/root"
	}
	codexDir := filepath.Join(homeDir, ".codex")
	sessionsDir := filepath.Join(codexDir, "sessions")
	workspaceRoot := "/workspace"
	repoDir := filepath.Join(workspaceRoot, "repo")

	if mkErr := os.MkdirAll(sessionsDir, 0o755); mkErr != nil {
		return fmt.Errorf("create sessions dir: %w", mkErr)
	}
	if mkErr := os.MkdirAll(workspaceRoot, 0o755); mkErr != nil {
		return fmt.Errorf("create workspace dir: %w", mkErr)
	}

	targetBranch := buildTargetBranch(s.cfg.RunID, s.cfg.IssueNumber)
	triggerKind := normalizeTriggerKind(s.cfg.TriggerKind)
	templateKind := normalizeTemplateKind(s.cfg.PromptTemplateKind, triggerKind)

	runStartedAt := time.Now().UTC()

	result := runResult{
		targetBranch: targetBranch,
		triggerKind:  triggerKind,
		templateKind: templateKind,
	}
	finalized := false
	defer func() {
		if finalized || err == nil {
			return
		}
		var exitErr ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode == 42 {
			return
		}
		finishedAt := time.Now().UTC()
		if persistErr := s.persistSessionSnapshot(ctx, result, codexState{
			sessionsDir: sessionsDir,
		}, runStartedAt, runStatusFailed, &finishedAt); persistErr != nil {
			s.logger.Warn("persist failed snapshot skipped", "err", persistErr)
		}
		_ = s.emitEvent(ctx, floweventdomain.EventTypeRunAgentSessionSaved, map[string]string{"status": runStatusFailed})
	}()

	if err := s.emitEvent(ctx, floweventdomain.EventTypeRunAgentStarted, map[string]string{
		"branch":           targetBranch,
		"trigger_kind":     triggerKind,
		"model":            s.cfg.AgentModel,
		"reasoning_effort": s.cfg.AgentReasoningEffort,
		"agent_key":        s.cfg.AgentKey,
	}); err != nil {
		s.logger.Warn("emit run.agent.started failed", "err", err)
	}

	state := codexState{
		homeDir:      homeDir,
		codexDir:     codexDir,
		sessionsDir:  sessionsDir,
		workspaceDir: workspaceRoot,
		repoDir:      repoDir,
	}

	if triggerKind == triggerKindDevRevise {
		restored, restoreErr := s.restoreLatestSession(ctx, result.targetBranch, state.sessionsDir)
		if restoreErr != nil {
			return ExitError{ExitCode: 5, Err: fmt.Errorf("restore latest session: %w", restoreErr)}
		}
		result.existingPRNumber = restored.existingPRNumber
		result.restoredSessionPath = restored.restoredSessionPath
		if restored.prNotFound {
			return s.failRevisePRNotFound(ctx, result, state, "pr_not_found")
		}
	}

	if err := s.prepareRepository(ctx, result, state); err != nil {
		return err
	}

	if err := runCommandWithInput(ctx, []byte(s.cfg.OpenAIAPIKey), io.Discard, io.Discard, "codex", "login", "--with-api-key"); err != nil {
		return fmt.Errorf("codex login failed: %w", err)
	}

	if err := s.writeCodexConfig(state.codexDir); err != nil {
		return err
	}

	taskBody, err := s.loadTaskTemplate(state.repoDir, result.templateKind)
	if err != nil {
		return err
	}
	prompt, err := s.buildPrompt(taskBody, result)
	if err != nil {
		return err
	}

	outputSchemaFile := filepath.Join(os.TempDir(), "codex-output-schema.json")
	if err := os.WriteFile(outputSchemaFile, []byte(outputSchemaJSON), 0o644); err != nil {
		return fmt.Errorf("write output schema: %w", err)
	}

	lastMessageFile := filepath.Join(os.TempDir(), "codex-last-message.json")
	if result.restoredSessionPath != "" {
		if err := s.emitEvent(ctx, floweventdomain.EventTypeRunAgentResumeUsed, map[string]string{"restored_session_path": result.restoredSessionPath}); err != nil {
			s.logger.Warn("emit run.agent.resume.used failed", "err", err)
		}
		err = runCommandLogged(ctx,
			"codex",
			"exec",
			"resume",
			"--last",
			"--cd", state.repoDir,
			"--output-schema", outputSchemaFile,
			"--last-message-file", lastMessageFile,
			prompt,
		)
	} else {
		err = runCommandLogged(ctx,
			"codex",
			"exec",
			"--cd", state.repoDir,
			"--output-schema", outputSchemaFile,
			"--last-message-file", lastMessageFile,
			prompt,
		)
	}
	if err != nil {
		return fmt.Errorf("codex exec failed: %w", err)
	}

	reportBytes, reportErr := os.ReadFile(lastMessageFile)
	if reportErr != nil {
		return fmt.Errorf("read codex result: %w", reportErr)
	}
	result.reportJSON = json.RawMessage(reportBytes)

	var report codexReport
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		return fmt.Errorf("decode codex result: %w", err)
	}
	if report.PRNumber <= 0 {
		return fmt.Errorf("invalid codex result: pr_number is required")
	}
	if strings.TrimSpace(report.PRURL) == "" {
		return fmt.Errorf("invalid codex result: pr_url is required")
	}
	result.prNumber = report.PRNumber
	result.prURL = strings.TrimSpace(report.PRURL)
	result.sessionID = strings.TrimSpace(report.SessionID)

	if strings.TrimSpace(report.Branch) != "" && strings.TrimSpace(report.Branch) != result.targetBranch {
		s.logger.Warn("codex reported different branch; forcing target branch", "reported_branch", report.Branch, "target_branch", result.targetBranch)
	}

	result.sessionFilePath = latestSessionFile(state.sessionsDir)
	if result.sessionID == "" && result.sessionFilePath != "" {
		result.sessionID = extractSessionIDFromFile(result.sessionFilePath)
	}

	_ = runCommandQuiet(ctx, state.repoDir, "git", "push", "origin", result.targetBranch)

	finishedAt := time.Now().UTC()
	if err := s.persistSessionSnapshot(ctx, result, state, runStartedAt, runStatusSucceeded, &finishedAt); err != nil {
		return err
	}
	if err := s.emitEvent(ctx, floweventdomain.EventTypeRunAgentSessionSaved, map[string]string{"status": runStatusSucceeded}); err != nil {
		s.logger.Warn("emit run.agent.session.saved failed", "err", err)
	}

	if triggerKind == triggerKindDevRevise {
		if err := s.emitEvent(ctx, floweventdomain.EventTypeRunPRUpdated, map[string]any{"branch": result.targetBranch, "pr_url": result.prURL, "pr_number": result.prNumber}); err != nil {
			s.logger.Warn("emit run.pr.updated failed", "err", err)
		}
	} else {
		if err := s.emitEvent(ctx, floweventdomain.EventTypeRunPRCreated, map[string]any{"branch": result.targetBranch, "pr_url": result.prURL, "pr_number": result.prNumber}); err != nil {
			s.logger.Warn("emit run.pr.created failed", "err", err)
		}
	}

	finalized = true
	s.logger.Info("agent-runner completed", "branch", result.targetBranch, "pr_number", result.prNumber)
	return nil
}

func (s *Service) failRevisePRNotFound(ctx context.Context, result runResult, state codexState, reason string) error {
	if err := s.emitEvent(ctx, floweventdomain.EventTypeRunRevisePRNotFound, map[string]string{"branch": result.targetBranch, "reason": reason}); err != nil {
		s.logger.Warn("emit run.revise.pr_not_found failed", "err", err)
	}
	finishedAt := time.Now().UTC()
	if err := s.persistSessionSnapshot(ctx, result, state, time.Now().UTC(), runStatusFailedPrecondition, &finishedAt); err != nil {
		s.logger.Warn("persist failed_precondition snapshot failed", "err", err)
	}
	_ = s.emitEvent(ctx, floweventdomain.EventTypeRunAgentSessionSaved, map[string]string{"status": runStatusFailedPrecondition})
	return ExitError{ExitCode: 42, Err: errors.New("revise precondition failed: pull request not found")}
}

func (s *Service) restoreLatestSession(ctx context.Context, branch string, sessionsDir string) (restoredSession, error) {
	snapshot, found, err := s.cp.GetLatestAgentSession(ctx, cpclient.LatestAgentSessionQuery{
		RepositoryFullName: s.cfg.RepositoryFullName,
		BranchName:         branch,
		AgentKey:           s.cfg.AgentKey,
	})
	if err != nil {
		return restoredSession{}, err
	}
	if !found {
		return restoredSession{}, nil
	}

	if snapshot.PRNumber <= 0 {
		return restoredSession{prNotFound: true}, nil
	}

	result := restoredSession{existingPRNumber: snapshot.PRNumber}
	if len(snapshot.CodexSessionJSON) > 0 {
		restoredPath := filepath.Join(sessionsDir, fmt.Sprintf("restored-%s.json", s.cfg.RunID))
		if writeErr := os.WriteFile(restoredPath, snapshot.CodexSessionJSON, 0o600); writeErr != nil {
			return restoredSession{}, fmt.Errorf("restore codex session file: %w", writeErr)
		}
		result.restoredSessionPath = restoredPath
		if err := s.emitEvent(ctx, floweventdomain.EventTypeRunAgentSessionRestored, map[string]string{"restored_session_path": restoredPath}); err != nil {
			s.logger.Warn("emit run.agent.session.restored failed", "err", err)
		}
	}
	return result, nil
}

func (s *Service) prepareRepository(ctx context.Context, result runResult, state codexState) error {
	if err := os.RemoveAll(state.repoDir); err != nil {
		return fmt.Errorf("cleanup repo dir: %w", err)
	}

	repoURL := fmt.Sprintf("https://%s:%s@github.com/%s.git", s.cfg.GitBotUsername, s.cfg.GitBotToken, s.cfg.RepositoryFullName)
	if err := runCommandQuiet(ctx, "", "git", "clone", repoURL, state.repoDir); err != nil {
		return fmt.Errorf("git clone failed")
	}

	_ = runCommandQuiet(ctx, state.repoDir, "git", "config", "user.name", s.cfg.AgentDisplayName)
	_ = runCommandQuiet(ctx, state.repoDir, "git", "config", "user.email", s.cfg.GitBotMail)
	_ = runCommandQuiet(ctx, state.repoDir, "git", "fetch", "origin", s.cfg.AgentBaseBranch, "--depth=50")

	branchExists := runCommandQuiet(ctx, state.repoDir, "git", "ls-remote", "--exit-code", "--heads", "origin", result.targetBranch) == nil
	if branchExists {
		if err := runCommandQuiet(ctx, state.repoDir, "git", "checkout", "-B", result.targetBranch, "origin/"+result.targetBranch); err != nil {
			return fmt.Errorf("checkout existing branch failed")
		}
		return nil
	}

	if result.triggerKind == triggerKindDevRevise {
		return s.failRevisePRNotFound(ctx, result, state, "branch_not_found")
	}

	if err := runCommandQuiet(ctx, state.repoDir, "git", "checkout", "-B", result.targetBranch, "origin/"+s.cfg.AgentBaseBranch); err != nil {
		return fmt.Errorf("checkout base branch failed")
	}
	return nil
}

func (s *Service) writeCodexConfig(codexDir string) error {
	configPath := filepath.Join(codexDir, "config.toml")
	content := fmt.Sprintf(`model = %q
model_reasoning_effort = %q
approval_policy = "never"
sandbox_mode = "danger-full-access"
model_verbosity = "low"
web_search_request = true

[history]
persistence = "save-all"

[mcp_servers.codex_k8s]
url = %q
bearer_token_env_var = "CODEXK8S_MCP_BEARER_TOKEN"
tool_timeout_sec = 180
`, s.cfg.AgentModel, s.cfg.AgentReasoningEffort, s.cfg.MCPBaseURL)
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write codex config: %w", err)
	}
	return nil
}

func (s *Service) loadTaskTemplate(repoDir string, templateKind string) (string, error) {
	templatePath := templatePathWork
	if templateKind == promptTemplateKindReview {
		templatePath = templatePathReview
	}
	absolutePath := filepath.Join(repoDir, templatePath)
	data, err := os.ReadFile(absolutePath)
	if err != nil {
		return "", fmt.Errorf("read template %s: %w", templatePath, err)
	}
	return string(data), nil
}

func (s *Service) buildPrompt(taskBody string, result runResult) (string, error) {
	var builder strings.Builder
	builder.WriteString("You are system agent dev for codex-k8s.\n")
	builder.WriteString("Repository: " + s.cfg.RepositoryFullName + "\n")
	builder.WriteString("Run ID: " + s.cfg.RunID + "\n")
	builder.WriteString(fmt.Sprintf("Issue number: %d\n", s.cfg.IssueNumber))
	builder.WriteString("Agent key: " + s.cfg.AgentKey + "\n")
	builder.WriteString("Target branch: " + result.targetBranch + "\n")
	builder.WriteString("Base branch: " + s.cfg.AgentBaseBranch + "\n")
	builder.WriteString("Trigger kind: " + result.triggerKind + "\n")
	if result.triggerKind == triggerKindDevRevise && result.existingPRNumber > 0 {
		builder.WriteString(fmt.Sprintf("Existing PR number: %d\n", result.existingPRNumber))
	}
	builder.WriteString("\nMandatory rules:\n")
	builder.WriteString("- Use MCP tools for issue/pr/comment/label operations.\n")
	builder.WriteString("- Git token may be used only for git fetch/commit/push transport.\n")
	builder.WriteString("- For dev run create or update PR to " + s.cfg.AgentBaseBranch + "; for revise update existing PR only.\n")
	builder.WriteString("- Keep branch " + result.targetBranch + ".\n")
	builder.WriteString("- Return ONLY JSON object matching output schema.\n")
	builder.WriteString("\nTask body:\n")
	builder.WriteString(taskBody)
	return builder.String(), nil
}

func (s *Service) persistSessionSnapshot(ctx context.Context, result runResult, state codexState, startedAt time.Time, status string, finishedAt *time.Time) error {
	issueNumber := optionalIssueNumber(s.cfg.IssueNumber)
	prNumber := optionalInt(result.prNumber)

	reportJSON := json.RawMessage(`{}`)
	if len(result.reportJSON) > 0 && json.Valid(result.reportJSON) {
		reportJSON = result.reportJSON
	}

	codexSessionJSON := readJSONFileOrNil(result.sessionFilePath)

	params := cpclient.AgentSessionUpsertParams{
		Identity: cpclient.SessionIdentity{
			RunID:              s.cfg.RunID,
			CorrelationID:      s.cfg.CorrelationID,
			ProjectID:          s.cfg.ProjectID,
			RepositoryFullName: s.cfg.RepositoryFullName,
			AgentKey:           s.cfg.AgentKey,
			IssueNumber:        issueNumber,
			BranchName:         result.targetBranch,
			PRNumber:           prNumber,
			PRURL:              result.prURL,
		},
		Template: cpclient.SessionTemplateContext{
			TriggerKind:     result.triggerKind,
			TemplateKind:    result.templateKind,
			TemplateSource:  s.cfg.PromptTemplateSource,
			TemplateLocale:  s.cfg.PromptTemplateLocale,
			Model:           s.cfg.AgentModel,
			ReasoningEffort: s.cfg.AgentReasoningEffort,
		},
		Runtime: cpclient.SessionRuntimeState{
			Status:           status,
			SessionID:        result.sessionID,
			SessionJSON:      reportJSON,
			CodexSessionPath: result.sessionFilePath,
			CodexSessionJSON: codexSessionJSON,
			StartedAt:        startedAt.UTC(),
			FinishedAt:       finishedAt,
		},
	}
	if err := s.cp.UpsertAgentSession(ctx, params); err != nil {
		return fmt.Errorf("persist session snapshot: %w", err)
	}
	return nil
}

func (s *Service) emitEvent(ctx context.Context, eventType floweventdomain.EventType, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s payload: %w", eventType, err)
	}
	if err := s.cp.InsertRunFlowEvent(ctx, s.cfg.RunID, eventType, json.RawMessage(raw)); err != nil {
		return err
	}
	return nil
}

func normalizeTriggerKind(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), triggerKindDevRevise) {
		return triggerKindDevRevise
	}
	return triggerKindDev
}

func normalizeTemplateKind(value string, triggerKind string) string {
	if triggerKind == triggerKindDevRevise {
		return promptTemplateKindReview
	}
	if strings.EqualFold(strings.TrimSpace(value), promptTemplateKindReview) {
		return promptTemplateKindReview
	}
	return promptTemplateKindWork
}

func buildTargetBranch(runID string, issueNumber int64) string {
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
	bytes, err := os.ReadFile(path)
	if err != nil || !json.Valid(bytes) {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return ""
	}
	for _, key := range []string{"session_id", "id", "conversation_id", "thread_id"} {
		value, ok := payload[key]
		if !ok {
			continue
		}
		stringValue := strings.TrimSpace(fmt.Sprint(value))
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
	cmd.Stdin = strings.NewReader(string(input))
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func runCommandLogged(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type runResult struct {
	targetBranch        string
	triggerKind         string
	templateKind        string
	restoredSessionPath string
	sessionFilePath     string
	sessionID           string
	existingPRNumber    int
	prNumber            int
	prURL               string
	reportJSON          json.RawMessage
}

type restoredSession struct {
	restoredSessionPath string
	existingPRNumber    int
	prNotFound          bool
}

type codexState struct {
	homeDir      string
	codexDir     string
	sessionsDir  string
	workspaceDir string
	repoDir      string
}

type codexReport struct {
	Summary         string `json:"summary"`
	Branch          string `json:"branch"`
	PRNumber        int    `json:"pr_number"`
	PRURL           string `json:"pr_url"`
	SessionID       string `json:"session_id"`
	Model           string `json:"model"`
	ReasoningEffort string `json:"reasoning_effort"`
}

const (
	runStatusSucceeded          = "succeeded"
	runStatusFailed             = "failed"
	runStatusFailedPrecondition = "failed_precondition"
)

const outputSchemaJSON = `{
  "type": "object",
  "properties": {
    "summary": { "type": "string" },
    "branch": { "type": "string" },
    "pr_number": { "type": "integer", "minimum": 1 },
    "pr_url": { "type": "string", "minLength": 1 },
    "session_id": { "type": "string" },
    "model": { "type": "string" },
    "reasoning_effort": { "type": "string" }
  },
  "required": ["summary", "branch", "pr_number", "pr_url"],
  "additionalProperties": true
}`
