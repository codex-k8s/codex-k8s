package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	cpclient "github.com/codex-k8s/codex-k8s/services/jobs/agent-runner/internal/controlplane"
)

// Run executes full runner flow and returns nil on successful completion.
func (s *Service) Run(ctx context.Context) (err error) {
	targetBranch := buildTargetBranch(s.cfg.RunTargetBranch, s.cfg.RunID, s.cfg.IssueNumber, s.cfg.TriggerKind, s.cfg.AgentBaseBranch)
	homeDir := strings.TrimSpace(os.Getenv("HOME"))
	if homeDir == "" {
		homeDir = "/root"
	}
	codexHomeDir := strings.TrimSpace(os.Getenv("CODEX_HOME"))
	if codexHomeDir == "" {
		codexHomeDir = filepath.Join(homeDir, ".codex")
	}
	runtimeMode := normalizeRuntimeMode(s.cfg.RuntimeMode)
	workspaceDir := runnerWorkspaceDir(runtimeMode, s.cfg.AgentKey, targetBranch)
	repoDir := runnerRepoDir(runtimeMode, s.cfg.AgentKey, targetBranch)

	state := codexState{
		homeDir:      homeDir,
		codexDir:     codexHomeDir,
		sessionsDir:  filepath.Join(codexHomeDir, "sessions"),
		workspaceDir: workspaceDir,
		repoDir:      repoDir,
	}

	if mkErr := os.MkdirAll(state.sessionsDir, 0o755); mkErr != nil {
		return fmt.Errorf("create sessions dir: %w", mkErr)
	}
	if mkErr := os.MkdirAll(state.workspaceDir, 0o755); mkErr != nil {
		return fmt.Errorf("create workspace dir: %w", mkErr)
	}
	gitAuthDir := filepath.Join(os.TempDir(), "codex-k8s-agent-runner-auth")
	if mkErr := os.MkdirAll(gitAuthDir, 0o755); mkErr != nil {
		return fmt.Errorf("create git auth dir: %w", mkErr)
	}
	if mkErr := os.MkdirAll(selfImproveSessionsDir, 0o755); mkErr != nil {
		return fmt.Errorf("create self-improve sessions dir: %w", mkErr)
	}
	cleanupGitAuthEnv, err := s.configureGitAuthEnvironment(gitAuthDir)
	if err != nil {
		return fmt.Errorf("configure git auth environment: %w", err)
	}
	defer cleanupGitAuthEnv()
	cleanupKubectlEnv, err := s.configureKubectlAccess(state.homeDir)
	if err != nil {
		return fmt.Errorf("configure kubectl access: %w", err)
	}
	defer cleanupKubectlEnv()
	triggerKind := normalizeTriggerKind(s.cfg.TriggerKind)
	templateKind := normalizeTemplateKind(s.cfg.PromptTemplateKind, triggerKind)
	sensitiveValues := s.sensitiveValues()

	runStartedAt := time.Now().UTC()
	result := runResult{
		targetBranch:     targetBranch,
		triggerKind:      triggerKind,
		templateKind:     templateKind,
		existingPRNumber: s.cfg.ExistingPRNumber,
	}
	writeScopePolicy := resolveRunWriteScopePolicy(result.triggerKind, s.cfg.AgentKey, s.cfg.DiscussionMode)
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
		persisted, persistErr := s.persistSessionSnapshot(ctx, &result, state, runStartedAt, runStatusFailed, &finishedAt)
		if persistErr != nil {
			s.logger.Warn("persist failed snapshot skipped", "err", persistErr)
		} else {
			_ = s.emitEvent(ctx, floweventdomain.EventTypeRunAgentSessionSaved, map[string]any{
				"status":            runStatusFailed,
				"snapshot_version":  persisted.SnapshotVersion,
				"snapshot_checksum": persisted.SnapshotChecksum,
			})
		}
	}()

	if err := s.emitEvent(ctx, floweventdomain.EventTypeRunAgentStarted, map[string]string{
		"branch":           targetBranch,
		"trigger_kind":     triggerKind,
		"runtime_mode":     runtimeMode,
		"model":            s.cfg.AgentModel,
		"reasoning_effort": s.cfg.AgentReasoningEffort,
		"agent_key":        s.cfg.AgentKey,
	}); err != nil {
		s.logger.Warn("emit run.agent.started failed", "err", err)
	}

	if s.cfg.DiscussionMode || webhookdomain.IsReviseTriggerKind(webhookdomain.NormalizeTriggerKind(triggerKind)) {
		restored, restoreErr := s.restoreLatestSession(ctx, result.targetBranch, state.sessionsDir)
		if restoreErr != nil {
			return ExitError{ExitCode: 5, Err: fmt.Errorf("restore latest session: %w", restoreErr)}
		}
		result.existingPRNumber = restored.existingPRNumber
		result.restoredSessionPath = restored.restoredSessionPath
		result.sessionID = restored.sessionID
		result.snapshotVersion = restored.snapshotVersion
		result.snapshotChecksum = restored.snapshotChecksum
		if restored.prNotFound {
			return s.failRevisePRNotFound(ctx, result, state, "pr_not_found")
		}
	}
	if writeScopePolicy.RequireExistingPR && result.existingPRNumber <= 0 {
		return fmt.Errorf("failed_precondition: reviewer run requires existing PR context")
	}

	if err := s.prepareRepository(ctx, result, state); err != nil {
		return err
	}
	baselineHead, err := gitCurrentHead(ctx, state.repoDir)
	if err != nil {
		return fmt.Errorf("resolve repository baseline head: %w", err)
	}

	if err := s.ensureCodexReady(ctx, state); err != nil {
		return err
	}
	if err := s.emitEvent(ctx, floweventdomain.EventTypeRunAgentReady, map[string]string{
		"branch":       targetBranch,
		"trigger_kind": triggerKind,
		"runtime_mode": runtimeMode,
		"agent_key":    s.cfg.AgentKey,
	}); err != nil {
		s.logger.Warn("emit run.agent.ready failed", "err", err)
	}
	if err := s.cp.UpsertRunStatusComment(ctx, cpclient.UpsertRunStatusCommentParams{
		RunID: s.cfg.RunID,
		Phase: "ready",
	}); err != nil {
		s.logger.Warn("upsert run status comment (ready) failed", "run_id", s.cfg.RunID, "err", err)
	}
	defer func() {
		if s.desiredCodexAuthMode() != codexAuthModeChatGPT {
			return
		}

		syncCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 20*time.Second)
		defer cancel()

		// Best-effort: Codex may update auth.json during execution (token refresh/rotation).
		if err := s.syncCodexAuthToControlPlane(syncCtx, state, "final"); err != nil {
			s.logger.Warn("final sync codex auth failed", "err", err)
		}
	}()

	outputSchemaFile := filepath.Join(os.TempDir(), "codex-output-schema.json")
	outputSchemaJSON, err := buildOutputSchemaJSON(outputSchemaParams{
		TriggerKind:  result.triggerKind,
		TemplateKind: result.templateKind,
		AgentKey:     s.cfg.AgentKey,
	})
	if err != nil {
		return fmt.Errorf("build output schema: %w", err)
	}
	if err := os.WriteFile(outputSchemaFile, outputSchemaJSON, 0o644); err != nil {
		return fmt.Errorf("write output schema: %w", err)
	}
	if s.cfg.DiscussionMode {
		if err := s.runDiscussionLoop(ctx, state, &result, runStartedAt, outputSchemaFile, sensitiveValues); err != nil {
			return err
		}
		finalized = true
		s.logger.Info("agent-runner completed", "branch", result.targetBranch, "discussion_mode", true)
		return nil
	}

	taskBody, err := s.renderTaskTemplate(result.templateKind, state.repoDir)
	if err != nil {
		return err
	}
	prompt, err := s.buildPrompt(taskBody, result, state.repoDir)
	if err != nil {
		return err
	}

	var codexOutput []byte
	if result.restoredSessionPath != "" {
		if err := s.emitEvent(ctx, floweventdomain.EventTypeRunAgentResumeUsed, map[string]string{"restored_session_path": result.restoredSessionPath}); err != nil {
			s.logger.Warn("emit run.agent.resume.used failed", "err", err)
		}
	}
	codexOutput, err = s.runCodexExecWithAuthRecovery(ctx, state, codexExecParams{
		RepoDir:          state.repoDir,
		Resume:           result.restoredSessionPath != "" || result.sessionID != "",
		ResumeSessionID:  result.sessionID,
		OutputSchemaFile: outputSchemaFile,
		Prompt:           prompt,
	})
	if err != nil {
		return fmt.Errorf("codex exec failed: %w", err)
	}
	result.codexExecOutput = redactSensitiveOutput(trimCapturedOutput(string(codexOutput), maxCapturedCommandOutput), sensitiveValues)

	report, _, err := parseCodexReportOutput(codexOutput)
	if err != nil {
		return err
	}
	report.ActionItems = normalizeStringList(report.ActionItems)
	report.EvidenceRefs = normalizeStringList(report.EvidenceRefs)
	report.ToolGaps = normalizeStringList(report.ToolGaps)
	result.report = report
	normalizedTriggerKind := webhookdomain.NormalizeTriggerKind(triggerKind)
	requiresPRFlow := !isAIRepairMainDirectTrigger(result.triggerKind) && !s.cfg.DiscussionMode
	if normalizedTriggerKind == webhookdomain.TriggerKindSelfImprove {
		if strings.TrimSpace(result.report.Diagnosis) == "" {
			return fmt.Errorf("invalid codex result: diagnosis is required for self_improve")
		}
		if len(result.report.ActionItems) == 0 {
			return fmt.Errorf("invalid codex result: action_items are required for self_improve")
		}
		if len(result.report.EvidenceRefs) == 0 {
			return fmt.Errorf("invalid codex result: evidence_refs are required for self_improve")
		}
		if err := s.emitEvent(ctx, floweventdomain.EventTypeRunSelfImproveDiagnosisReady, selfImproveDiagnosisReadyPayload{
			Diagnosis:    result.report.Diagnosis,
			ActionItems:  result.report.ActionItems,
			EvidenceRefs: result.report.EvidenceRefs,
			ToolGaps:     result.report.ToolGaps,
		}); err != nil {
			s.logger.Warn("emit run.self_improve.diagnosis_ready failed", "err", err)
		}
	}
	if requiresPRFlow {
		if report.PRNumber <= 0 {
			return fmt.Errorf("invalid codex result: pr_number is required")
		}
		if strings.TrimSpace(report.PRURL) == "" {
			return fmt.Errorf("invalid codex result: pr_url is required")
		}
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
	if err := enforceRunWriteScope(ctx, state.repoDir, baselineHead, result.triggerKind, s.cfg.AgentKey, result.existingPRNumber, s.cfg.DiscussionMode); err != nil {
		return err
	}

	if !s.cfg.DiscussionMode {
		gitPushOutput, pushErr := runCommandCaptureCombinedOutput(ctx, state.repoDir, "git", "push", "origin", result.targetBranch)
		result.gitPushOutput = redactSensitiveOutput(gitPushOutput, sensitiveValues)
		if pushErr != nil {
			return fmt.Errorf("git push failed: %w", pushErr)
		}
	}
	result.toolGaps = detectToolGaps(result.report, result.codexExecOutput, result.gitPushOutput)
	result.report.ToolGaps = result.toolGaps
	if len(result.toolGaps) > 0 {
		// This event is stored in flow_events and consumed by the self-improve loop:
		// km agent analyzes gaps, updates prompts/docs/toolchain (Dockerfile/bootstrap_tools),
		// and submits those updates as a separate PR for Owner review.
		if err := s.emitEvent(ctx, floweventdomain.EventTypeRunToolchainGapDetected, toolchainGapDetectedPayload{
			ToolGaps: result.toolGaps,
			Sources: []string{
				"codex_exec_output",
				"git_push_output",
				"report.tool_gaps",
			},
			SuggestedUpdatePaths: []string{
				"services/jobs/agent-runner/scripts/bootstrap_tools.sh",
				"services/jobs/agent-runner/Dockerfile",
			},
		}); err != nil {
			s.logger.Warn("emit run.toolchain.gap_detected failed", "err", err)
		}
	}

	finishedAt := time.Now().UTC()
	persisted, err := s.persistSessionSnapshot(ctx, &result, state, runStartedAt, runStatusSucceeded, &finishedAt)
	if err != nil {
		return err
	}
	if err := s.emitEvent(ctx, floweventdomain.EventTypeRunAgentSessionSaved, map[string]any{
		"status":            runStatusSucceeded,
		"snapshot_version":  persisted.SnapshotVersion,
		"snapshot_checksum": persisted.SnapshotChecksum,
	}); err != nil {
		s.logger.Warn("emit run.agent.session.saved failed", "err", err)
	}

	if requiresPRFlow {
		if webhookdomain.IsReviseTriggerKind(webhookdomain.NormalizeTriggerKind(triggerKind)) {
			if err := s.emitEvent(ctx, floweventdomain.EventTypeRunPRUpdated, map[string]any{"branch": result.targetBranch, "pr_url": result.prURL, "pr_number": result.prNumber}); err != nil {
				s.logger.Warn("emit run.pr.updated failed", "err", err)
			}
		} else {
			if err := s.emitEvent(ctx, floweventdomain.EventTypeRunPRCreated, map[string]any{"branch": result.targetBranch, "pr_url": result.prURL, "pr_number": result.prNumber}); err != nil {
				s.logger.Warn("emit run.pr.created failed", "err", err)
			}
		}
	} else {
		if err := s.emitEvent(ctx, floweventdomain.EventTypeRunAgentStatusReported, map[string]any{
			"branch":          result.targetBranch,
			"trigger":         normalizedTriggerKind,
			"main_direct":     isAIRepairMainDirectTrigger(result.triggerKind),
			"discussion_mode": s.cfg.DiscussionMode,
		}); err != nil {
			s.logger.Warn("emit run.agent.status_reported failed", "err", err)
		}
	}

	finalized = true
	s.logger.Info("agent-runner completed", "branch", result.targetBranch, "pr_number", result.prNumber)
	return nil
}

func (s *Service) runCodexExecWithAuthRecovery(ctx context.Context, state codexState, params codexExecParams) ([]byte, error) {
	output, stderr, err := runCodexExec(ctx, params)
	if err == nil {
		return output, nil
	}
	if !isCodexAuthenticationError(err.Error(), string(output), stderr) {
		return nil, err
	}

	s.logger.Warn("codex exec returned auth error; retrying after device auth", "run_id", s.cfg.RunID)
	if authErr := s.ensureCodexAuthenticated(ctx, state); authErr != nil {
		return nil, fmt.Errorf("recover codex auth: %w", authErr)
	}

	output, _, err = runCodexExec(ctx, params)
	if err != nil {
		return nil, err
	}
	return output, nil
}

type codexExecParams struct {
	RepoDir          string
	Resume           bool
	ResumeSessionID  string
	OutputSchemaFile string
	Prompt           string
}

func runCodexExec(ctx context.Context, params codexExecParams) ([]byte, string, error) {
	outputFile, err := os.CreateTemp("", "codex-output-*.json")
	if err != nil {
		return nil, "", fmt.Errorf("create codex output file: %w", err)
	}
	outputFilePath := outputFile.Name()
	_ = outputFile.Close()
	defer func() {
		_ = os.Remove(outputFilePath)
	}()

	args := make([]string, 0, 12)
	if params.Resume {
		args = append(args, "exec", "resume", "--output-last-message", outputFilePath)
		if sessionID := strings.TrimSpace(params.ResumeSessionID); sessionID != "" {
			args = append(args, sessionID)
		} else {
			args = append(args, "--last")
		}
		args = append(args, params.Prompt)
	} else {
		args = append(
			args,
			"exec",
			"--cd", params.RepoDir,
			"--output-schema", params.OutputSchemaFile,
			"--output-last-message", outputFilePath,
			params.Prompt,
		)
	}

	stdout, stderr, err := runCommandCaptureOutputWithStderr(ctx, params.RepoDir, "codex", args...)
	if captured := readCommandOutputFileOrNil(outputFilePath); len(captured) > 0 {
		return captured, stderr, err
	}
	return stdout, stderr, err
}

func (s *Service) failRevisePRNotFound(ctx context.Context, result runResult, state codexState, reason string) error {
	if err := s.emitEvent(ctx, floweventdomain.EventTypeRunRevisePRNotFound, map[string]string{"branch": result.targetBranch, "reason": reason}); err != nil {
		s.logger.Warn("emit run.revise.pr_not_found failed", "err", err)
	}
	finishedAt := time.Now().UTC()
	persisted, err := s.persistSessionSnapshot(ctx, &result, state, time.Now().UTC(), runStatusFailedPrecondition, &finishedAt)
	if err != nil {
		s.logger.Warn("persist failed_precondition snapshot failed", "err", err)
	}
	if err == nil {
		_ = s.emitEvent(ctx, floweventdomain.EventTypeRunAgentSessionSaved, map[string]any{
			"status":            runStatusFailedPrecondition,
			"snapshot_version":  persisted.SnapshotVersion,
			"snapshot_checksum": persisted.SnapshotChecksum,
		})
	}
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
		if s.cfg.DiscussionMode {
			result := restoredSession{}
			if snapshot.RunID == s.cfg.RunID {
				result.snapshotVersion = snapshot.SnapshotVersion
				result.snapshotChecksum = snapshot.SnapshotChecksum
			}
			if sessionID := strings.TrimSpace(snapshot.SessionID); sessionID != "" {
				result.sessionID = sessionID
			}
			if len(snapshot.CodexSessionJSON) > 0 {
				restoredPath, err := s.restoreSessionSnapshot(ctx, sessionsDir, snapshot.CodexSessionJSON)
				if err != nil {
					return restoredSession{}, err
				}
				result.restoredSessionPath = restoredPath
				if result.sessionID == "" {
					result.sessionID = extractSessionIDFromFile(restoredPath)
				}
			}
			return result, nil
		}
		if s.cfg.ExistingPRNumber > 0 {
			result := restoredSession{existingPRNumber: s.cfg.ExistingPRNumber}
			if snapshot.RunID == s.cfg.RunID {
				result.snapshotVersion = snapshot.SnapshotVersion
				result.snapshotChecksum = snapshot.SnapshotChecksum
			}
			return result, nil
		}
		return restoredSession{prNotFound: true}, nil
	}

	result := restoredSession{
		existingPRNumber: snapshot.PRNumber,
		sessionID:        strings.TrimSpace(snapshot.SessionID),
	}
	if snapshot.RunID == s.cfg.RunID {
		result.snapshotVersion = snapshot.SnapshotVersion
		result.snapshotChecksum = snapshot.SnapshotChecksum
	}
	if len(snapshot.CodexSessionJSON) > 0 {
		restoredPath, err := s.restoreSessionSnapshot(ctx, sessionsDir, snapshot.CodexSessionJSON)
		if err != nil {
			return restoredSession{}, err
		}
		result.restoredSessionPath = restoredPath
		if result.sessionID == "" {
			result.sessionID = extractSessionIDFromFile(restoredPath)
		}
	}
	return result, nil
}

func (s *Service) restoreSessionSnapshot(ctx context.Context, sessionsDir string, sessionJSON []byte) (string, error) {
	restoredPath := filepath.Join(sessionsDir, fmt.Sprintf("restored-%s.json", s.cfg.RunID))
	if writeErr := os.WriteFile(restoredPath, sessionJSON, 0o600); writeErr != nil {
		return "", fmt.Errorf("restore codex session file: %w", writeErr)
	}
	if err := s.emitEvent(ctx, floweventdomain.EventTypeRunAgentSessionRestored, map[string]string{"restored_session_path": restoredPath}); err != nil {
		s.logger.Warn("emit run.agent.session.restored failed", "err", err)
	}
	return restoredPath, nil
}

func (s *Service) prepareRepository(ctx context.Context, result runResult, state codexState) error {
	repoURL := fmt.Sprintf("https://github.com/%s.git", s.cfg.RepositoryFullName)
	if usesPreparedFullEnvRepository(s.cfg.RuntimeMode) {
		if err := ensureExistingRepoDirCheckout(ctx, state.repoDir, repoURL); err != nil {
			return err
		}
		_ = runCommandQuiet(ctx, state.repoDir, "git", "config", "user.name", s.cfg.AgentDisplayName)
		_ = runCommandQuiet(ctx, state.repoDir, "git", "config", "user.email", s.cfg.GitBotMail)
		if err := ensurePreparedFullEnvBranch(ctx, state.repoDir, result.targetBranch); err != nil {
			return err
		}
		return nil
	}

	if err := ensureRepoDirCheckout(ctx, state.repoDir, repoURL); err != nil {
		return err
	}

	_ = runCommandQuiet(ctx, state.repoDir, "git", "config", "user.name", s.cfg.AgentDisplayName)
	_ = runCommandQuiet(ctx, state.repoDir, "git", "config", "user.email", s.cfg.GitBotMail)
	if err := runCommandQuiet(ctx, state.repoDir, "git", "fetch", "--prune", "--tags", "origin"); err != nil {
		return fmt.Errorf("git fetch failed")
	}

	branchExists := runCommandQuiet(ctx, state.repoDir, "git", "ls-remote", "--exit-code", "--heads", "origin", result.targetBranch) == nil
	if branchExists {
		if err := runCommandQuiet(ctx, state.repoDir, "git", "checkout", "-B", result.targetBranch, "origin/"+result.targetBranch); err != nil {
			return fmt.Errorf("checkout existing branch failed")
		}
	} else {
		if webhookdomain.IsReviseTriggerKind(webhookdomain.NormalizeTriggerKind(result.triggerKind)) && !s.cfg.DiscussionMode {
			return s.failRevisePRNotFound(ctx, result, state, "branch_not_found")
		}
		if err := runCommandQuiet(ctx, state.repoDir, "git", "checkout", "-B", result.targetBranch, "origin/"+s.cfg.AgentBaseBranch); err != nil {
			return fmt.Errorf("checkout base branch failed")
		}
	}

	if err := runCommandQuiet(ctx, state.repoDir, "git", "reset", "--hard"); err != nil {
		return fmt.Errorf("git reset failed")
	}
	if err := runCommandQuiet(ctx, state.repoDir, "git", gitCleanArgs(s.cfg.RuntimeMode)...); err != nil {
		return fmt.Errorf("git clean failed")
	}

	return nil
}

func ensureExistingRepoDirCheckout(ctx context.Context, repoDir string, repoURL string) error {
	repoDir = strings.TrimSpace(repoDir)
	repoURL = strings.TrimSpace(repoURL)
	if repoDir == "" {
		return fmt.Errorf("repo dir is required")
	}
	if repoURL == "" {
		return fmt.Errorf("repo url is required")
	}
	if runCommandQuiet(ctx, repoDir, "git", "rev-parse", "--is-inside-work-tree") != nil {
		return fmt.Errorf("full-env repository is missing prepared git work tree: %s", repoDir)
	}
	if err := runCommandQuiet(ctx, repoDir, "git", "remote", "set-url", "origin", repoURL); err == nil {
		return nil
	}
	if err := runCommandQuiet(ctx, repoDir, "git", "remote", "add", "origin", repoURL); err == nil {
		return nil
	}
	if err := runCommandQuiet(ctx, repoDir, "git", "remote", "remove", "origin"); err != nil {
		return fmt.Errorf("remove stale git origin failed")
	}
	if err := runCommandQuiet(ctx, repoDir, "git", "remote", "add", "origin", repoURL); err != nil {
		return fmt.Errorf("reconfigure git origin failed")
	}
	return nil
}

func ensurePreparedFullEnvBranch(ctx context.Context, repoDir string, targetBranch string) error {
	targetBranch = strings.TrimSpace(targetBranch)
	if targetBranch == "" {
		return fmt.Errorf("target branch is required")
	}

	currentBranch, err := currentLocalBranch(ctx, repoDir)
	if err == nil {
		if currentBranch == targetBranch {
			return nil
		}
		return fmt.Errorf("full-env repository branch mismatch: got %s want %s", currentBranch, targetBranch)
	}

	if !isDetachedHeadErr(err) {
		return fmt.Errorf("resolve full-env repository branch: %w", err)
	}
	if runCommandQuiet(ctx, repoDir, "git", "checkout", "-B", targetBranch) != nil {
		return fmt.Errorf("attach full-env repository branch failed")
	}
	return nil
}

func ensureRepoDirCheckout(ctx context.Context, repoDir string, repoURL string) error {
	repoDir = strings.TrimSpace(repoDir)
	repoURL = strings.TrimSpace(repoURL)
	if repoDir == "" {
		return fmt.Errorf("repo dir is required")
	}
	if repoURL == "" {
		return fmt.Errorf("repo url is required")
	}

	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		return fmt.Errorf("create repo dir: %w", err)
	}

	if runCommandQuiet(ctx, repoDir, "git", "rev-parse", "--is-inside-work-tree") != nil {
		if err := cleanupDirectoryContents(repoDir); err != nil {
			return fmt.Errorf("cleanup repo dir contents: %w", err)
		}
		if err := runCommandQuiet(ctx, repoDir, "git", "init"); err != nil {
			return fmt.Errorf("git init failed")
		}
		if err := runCommandQuiet(ctx, repoDir, "git", "remote", "add", "origin", repoURL); err != nil {
			return fmt.Errorf("configure git origin failed")
		}
		return nil
	}

	if err := runCommandQuiet(ctx, repoDir, "git", "remote", "set-url", "origin", repoURL); err == nil {
		return nil
	}
	if err := runCommandQuiet(ctx, repoDir, "git", "remote", "add", "origin", repoURL); err == nil {
		return nil
	}
	if err := runCommandQuiet(ctx, repoDir, "git", "remote", "remove", "origin"); err != nil {
		return fmt.Errorf("remove stale git origin failed")
	}
	if err := runCommandQuiet(ctx, repoDir, "git", "remote", "add", "origin", repoURL); err != nil {
		return fmt.Errorf("reconfigure git origin failed")
	}
	return nil
}

func currentLocalBranch(ctx context.Context, repoDir string) (string, error) {
	output, err := runCommandCaptureCombinedOutput(ctx, repoDir, "git", "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func isDetachedHeadErr(err error) bool {
	if err == nil {
		return false
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode() == 1
	}
	return false
}

func cleanupDirectoryContents(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(path, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) persistSessionSnapshot(ctx context.Context, result *runResult, state codexState, startedAt time.Time, status string, finishedAt *time.Time) (cpclient.AgentSessionUpsertResult, error) {
	if result == nil {
		return cpclient.AgentSessionUpsertResult{}, fmt.Errorf("run result is required")
	}
	issueNumber := optionalIssueNumber(s.cfg.IssueNumber)
	prNumber := optionalInt(result.prNumber)

	codexSessionJSON := readJSONFileOrNil(result.sessionFilePath)
	sessionJSON := buildSessionLogJSON(*result, status)
	snapshotChecksum, err := agentdomain.ComputeSessionSnapshotChecksum(sessionJSON, codexSessionJSON)
	if err != nil {
		return cpclient.AgentSessionUpsertResult{}, fmt.Errorf("compute session snapshot checksum: %w", err)
	}

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
			SessionJSON:      sessionJSON,
			CodexSessionPath: result.sessionFilePath,
			CodexSessionJSON: codexSessionJSON,
			SnapshotVersion:  result.snapshotVersion,
			SnapshotChecksum: snapshotChecksum,
			StartedAt:        startedAt.UTC(),
			FinishedAt:       finishedAt,
		},
	}
	persisted, err := s.cp.UpsertAgentSession(ctx, params)
	if err != nil {
		return cpclient.AgentSessionUpsertResult{}, fmt.Errorf("persist session snapshot: %w", err)
	}
	result.snapshotVersion = persisted.SnapshotVersion
	result.snapshotChecksum = persisted.SnapshotChecksum
	return persisted, nil
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
