package runner

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	gh "github.com/google/go-github/v82/github"
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
			persisted, err := s.persistSessionSnapshot(ctx, result, state, runStartedAt, runStatusSucceeded, &finishedAt)
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
		pendingHumanComments := append([]discussionPendingHumanComment(nil), issueState.PendingHumanComments...)
		s.ensureDiscussionCommentsObserved(ctx, pendingHumanComments)

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

		codexOutput, err := s.runCodexExecWithAuthRecovery(ctx, state, result, runStartedAt, codexExecParams{
			RepoDir:          state.repoDir,
			Resume:           resume,
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
		persisted, err := s.persistSessionSnapshot(ctx, result, state, runStartedAt, runStatusSucceeded, &finishedAt)
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
		if err := s.emitEvent(ctx, floweventdomain.EventTypeRunAgentStatusReported, map[string]any{
			"branch":                result.targetBranch,
			"trigger":               webhookdomain.NormalizeTriggerKind(result.triggerKind),
			"discussion_mode":       true,
			"last_human_comment_id": issueState.MaxHumanCommentID,
		}); err != nil {
			s.logger.Warn("emit run.agent.status_reported failed", "err", err)
		}
		updatedIssueState, err := s.loadDiscussionIssueState(ctx)
		if err != nil {
			s.logger.Warn("load discussion issue state for processed reactions failed", "err", err)
		} else {
			s.ensureDiscussionCommentsProcessed(ctx, processedDiscussionCommentIDs(pendingHumanComments, updatedIssueState))
		}

		lastProcessedHumanCommentID = issueState.MaxHumanCommentID
		result.restoredSessionPath = result.sessionFilePath
		if err := waitForDiscussionPoll(ctx, pollInterval); err != nil {
			return err
		}
	}
}

func (s *Service) loadDiscussionIssueState(ctx context.Context) (discussionIssueState, error) {
	client, repositoryOwner, repositoryName, err := s.newDiscussionGitHubClient()
	if err != nil {
		return discussionIssueState{}, fmt.Errorf("build github client: %w", err)
	}

	issueItem, _, err := client.Issues.Get(ctx, repositoryOwner, repositoryName, int(s.cfg.IssueNumber))
	if err != nil {
		return discussionIssueState{}, fmt.Errorf("load discussion issue state: %w", err)
	}

	sortByCreated := "created"
	sortDirectionAsc := "asc"
	commentItems, _, err := client.Issues.ListComments(ctx, repositoryOwner, repositoryName, int(s.cfg.IssueNumber), &gh.IssueListCommentsOptions{
		Sort:        &sortByCreated,
		Direction:   &sortDirectionAsc,
		ListOptions: gh.ListOptions{PerPage: 100},
	})
	if err != nil {
		return discussionIssueState{}, fmt.Errorf("load discussion issue comments: %w", err)
	}

	issue := discussionIssueAPIResponse{
		State: issueItem.GetState(),
	}
	if len(issueItem.Labels) > 0 {
		issue.Labels = make([]discussionIssueLabelItem, 0, len(issueItem.Labels))
		for _, label := range issueItem.Labels {
			issue.Labels = append(issue.Labels, discussionIssueLabelItem{Name: label.GetName()})
		}
	}

	comments := make([]discussionIssueCommentResponse, 0, len(commentItems))
	for _, item := range commentItems {
		createdAt := ""
		if item != nil && item.CreatedAt != nil {
			createdAt = item.CreatedAt.Time.UTC().Format(time.RFC3339Nano)
		}
		comments = append(comments, discussionIssueCommentResponse{
			ID:        item.GetID(),
			Body:      item.GetBody(),
			CreatedAt: createdAt,
			User: discussionIssueCommentUser{
				Login: item.GetUser().GetLogin(),
				Type:  item.GetUser().GetType(),
			},
		})
	}
	sort.Slice(comments, func(i, j int) bool {
		return strings.TrimSpace(comments[i].CreatedAt) < strings.TrimSpace(comments[j].CreatedAt)
	})

	return deriveDiscussionIssueState(issue, comments, s.cfg.GitBotUsername), nil
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
