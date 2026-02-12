package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	cpclient "github.com/codex-k8s/codex-k8s/services/jobs/agent-runner/internal/controlplane"
)

// Run executes full runner flow and returns nil only on successful PR update/create path.
func (s *Service) Run(ctx context.Context) (err error) {
	homeDir := strings.TrimSpace(os.Getenv("HOME"))
	if homeDir == "" {
		homeDir = "/root"
	}

	state := codexState{
		homeDir:      homeDir,
		codexDir:     filepath.Join(homeDir, ".codex"),
		sessionsDir:  filepath.Join(homeDir, ".codex", "sessions"),
		workspaceDir: "/workspace",
		repoDir:      filepath.Join("/workspace", "repo"),
	}

	if mkErr := os.MkdirAll(state.sessionsDir, 0o755); mkErr != nil {
		return fmt.Errorf("create sessions dir: %w", mkErr)
	}
	if mkErr := os.MkdirAll(state.workspaceDir, 0o755); mkErr != nil {
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
		if persistErr := s.persistSessionSnapshot(ctx, result, state, runStartedAt, runStatusFailed, &finishedAt); persistErr != nil {
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

	taskBody, err := s.renderTaskTemplate(result.templateKind)
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
		err = runCommandLogged(
			ctx,
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
		err = runCommandLogged(
			ctx,
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
