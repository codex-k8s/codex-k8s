package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

const runStatusCommentMarker = "<!-- codex-k8s:run-status "

type discussionIssueAPIResponse struct {
	State  string                     `json:"state"`
	Labels []discussionIssueLabelItem `json:"labels"`
}

type discussionIssueLabelItem struct {
	Name string `json:"name"`
}

type discussionIssueCommentResponse struct {
	ID        int64                      `json:"id"`
	Body      string                     `json:"body"`
	CreatedAt string                     `json:"created_at"`
	User      discussionIssueCommentUser `json:"user"`
}

type discussionIssueCommentUser struct {
	Login string `json:"login"`
	Type  string `json:"type"`
}

func (s *Service) runDiscussionLoop(ctx context.Context, state codexState, result *runResult, runStartedAt time.Time, outputSchemaFile string, sensitiveValues []string) error {
	pollInterval := s.cfg.DiscussionPollInterval
	if pollInterval <= 0 {
		pollInterval = 15 * time.Second
	}

	var lastProcessedHumanCommentID int64
	for {
		issueState, err := s.loadDiscussionIssueState(ctx)
		if err != nil {
			return err
		}
		if shouldStopDiscussionLoop(issueState) {
			finishedAt := time.Now().UTC()
			if err := s.persistSessionSnapshot(ctx, *result, state, runStartedAt, runStatusSucceeded, &finishedAt); err != nil {
				return err
			}
			if err := s.emitEvent(ctx, floweventdomain.EventTypeRunAgentStatusReported, map[string]any{
				"branch":                result.targetBranch,
				"trigger":               webhookdomain.NormalizeTriggerKind(result.triggerKind),
				"discussion_mode":       true,
				"stop_reason":           resolveDiscussionStopReason(issueState),
				"last_human_comment_id": lastProcessedHumanCommentID,
			}); err != nil {
				s.logger.Warn("emit run.agent.status_reported failed", "err", err)
			}
			return nil
		}
		if !shouldRunDiscussionCycle(issueState, lastProcessedHumanCommentID) {
			if err := waitForDiscussionPoll(ctx, pollInterval); err != nil {
				return err
			}
			continue
		}

		if err := s.prepareRepository(ctx, *result, state); err != nil {
			return err
		}
		prompt, err := s.renderDiscussionPrompt(*result, state.repoDir)
		if err != nil {
			return err
		}

		resume := result.restoredSessionPath != "" || result.sessionFilePath != "" || result.sessionID != ""
		if resume {
			if err := s.emitEvent(ctx, floweventdomain.EventTypeRunAgentResumeUsed, map[string]string{"restored_session_path": strings.TrimSpace(result.restoredSessionPath)}); err != nil {
				s.logger.Warn("emit run.agent.resume.used failed", "err", err)
			}
		}

		codexOutput, err := s.runCodexExecWithAuthRecovery(ctx, state, resume, outputSchemaFile, prompt)
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
		result.sessionID = strings.TrimSpace(report.SessionID)
		result.sessionFilePath = latestSessionFile(state.sessionsDir)
		if result.sessionID == "" && result.sessionFilePath != "" {
			result.sessionID = extractSessionIDFromFile(result.sessionFilePath)
		}
		result.toolGaps = detectToolGaps(result.report, result.codexExecOutput, "")
		result.report.ToolGaps = result.toolGaps
		if len(result.toolGaps) > 0 {
			if err := s.emitEvent(ctx, floweventdomain.EventTypeRunToolchainGapDetected, toolchainGapDetectedPayload{
				ToolGaps: result.toolGaps,
				Sources: []string{
					"codex_exec_output",
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
		if err := s.persistSessionSnapshot(ctx, *result, state, runStartedAt, runStatusSucceeded, &finishedAt); err != nil {
			return err
		}
		if err := s.emitEvent(ctx, floweventdomain.EventTypeRunAgentSessionSaved, map[string]string{"status": runStatusSucceeded}); err != nil {
			s.logger.Warn("emit run.agent.session.saved failed", "err", err)
		}
		if err := s.emitEvent(ctx, floweventdomain.EventTypeRunAgentStatusReported, map[string]any{
			"branch":                result.targetBranch,
			"trigger":               webhookdomain.NormalizeTriggerKind(result.triggerKind),
			"discussion_mode":       true,
			"last_human_comment_id": issueState.MaxHumanCommentID,
		}); err != nil {
			s.logger.Warn("emit run.agent.status_reported failed", "err", err)
		}

		lastProcessedHumanCommentID = issueState.MaxHumanCommentID
		result.restoredSessionPath = result.sessionFilePath
		if err := waitForDiscussionPoll(ctx, pollInterval); err != nil {
			return err
		}
	}
}

func (s *Service) loadDiscussionIssueState(ctx context.Context) (discussionIssueState, error) {
	issuePath := fmt.Sprintf("repos/%s/issues/%d", strings.TrimSpace(s.cfg.RepositoryFullName), s.cfg.IssueNumber)
	issueOutput, err := runCommandCaptureCombinedOutput(ctx, "", "gh", "api", issuePath)
	if err != nil {
		return discussionIssueState{}, fmt.Errorf("load discussion issue state: %w", err)
	}

	var issue discussionIssueAPIResponse
	if err := json.Unmarshal([]byte(issueOutput), &issue); err != nil {
		return discussionIssueState{}, fmt.Errorf("decode discussion issue state: %w", err)
	}

	commentsPath := fmt.Sprintf("repos/%s/issues/%d/comments?per_page=100", strings.TrimSpace(s.cfg.RepositoryFullName), s.cfg.IssueNumber)
	commentsOutput, err := runCommandCaptureCombinedOutput(ctx, "", "gh", "api", commentsPath)
	if err != nil {
		return discussionIssueState{}, fmt.Errorf("load discussion issue comments: %w", err)
	}

	comments := make([]discussionIssueCommentResponse, 0, 16)
	if strings.TrimSpace(commentsOutput) != "" {
		if err := json.Unmarshal([]byte(commentsOutput), &comments); err != nil {
			return discussionIssueState{}, fmt.Errorf("decode discussion issue comments: %w", err)
		}
	}
	sort.Slice(comments, func(i, j int) bool {
		return strings.TrimSpace(comments[i].CreatedAt) < strings.TrimSpace(comments[j].CreatedAt)
	})

	state := discussionIssueState{
		State: strings.TrimSpace(issue.State),
	}
	for _, label := range issue.Labels {
		normalized := normalizeDiscussionLogin(label.Name)
		switch {
		case normalized == normalizeDiscussionLogin(webhookdomain.DefaultModeDiscussionLabel):
			state.HasDiscussionLabel = true
		case strings.HasPrefix(normalized, "run:"):
			state.HasRunLabel = true
		}
	}

	botLogin := normalizeDiscussionLogin(s.cfg.GitBotUsername)
	seenAgentReply := false
	for _, item := range comments {
		login := normalizeDiscussionLogin(item.User.Login)
		userType := strings.ToLower(strings.TrimSpace(item.User.Type))
		if login != botLogin && (userType == "" || userType == "user") && item.ID > state.MaxHumanCommentID {
			state.MaxHumanCommentID = item.ID
		}
		if login == botLogin && !strings.Contains(item.Body, runStatusCommentMarker) {
			state.HasAgentReply = true
			state.HasHumanAfterAgentReply = false
			seenAgentReply = true
			continue
		}
		if seenAgentReply && login != botLogin && (userType == "" || userType == "user") {
			state.HasHumanAfterAgentReply = true
		}
	}

	return state, nil
}

func shouldRunDiscussionCycle(state discussionIssueState, lastProcessedHumanCommentID int64) bool {
	if !state.HasAgentReply {
		return true
	}
	if state.HasHumanAfterAgentReply {
		return true
	}
	return state.MaxHumanCommentID > lastProcessedHumanCommentID
}

func shouldStopDiscussionLoop(state discussionIssueState) bool {
	if !strings.EqualFold(strings.TrimSpace(state.State), "open") {
		return true
	}
	if !state.HasDiscussionLabel {
		return true
	}
	return state.HasRunLabel
}

func resolveDiscussionStopReason(state discussionIssueState) string {
	switch {
	case !strings.EqualFold(strings.TrimSpace(state.State), "open"):
		return "issue_closed"
	case !state.HasDiscussionLabel:
		return "discussion_label_removed"
	case state.HasRunLabel:
		return "run_label_detected"
	default:
		return "stopped"
	}
}

func waitForDiscussionPoll(ctx context.Context, pollInterval time.Duration) error {
	timer := time.NewTimer(pollInterval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func normalizeDiscussionLogin(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func (s *Service) renderDiscussionPrompt(result runResult, repoDir string) (string, error) {
	taskBody, err := s.renderTaskTemplate(result.templateKind, repoDir)
	if err != nil {
		return "", err
	}
	return s.buildPrompt(taskBody, result, repoDir)
}
