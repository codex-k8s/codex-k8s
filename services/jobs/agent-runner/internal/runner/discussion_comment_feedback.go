package runner

import (
	"context"
	"fmt"
	"slices"
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

const (
	discussionCommentReactionObserved  = "eyes"
	discussionCommentReactionProcessed = "+1"
)

func deriveDiscussionIssueState(issue discussionIssueAPIResponse, comments []discussionIssueCommentResponse, botLogin string) discussionIssueState {
	state := discussionIssueState{
		State: strings.TrimSpace(issue.State),
	}

	normalizedBotLogin := normalizeDiscussionLogin(botLogin)
	lastAgentReplyIndex := -1
	for _, label := range issue.Labels {
		normalized := normalizeDiscussionLogin(label.Name)
		switch {
		case normalized == normalizeDiscussionLogin(webhookdomain.DefaultModeDiscussionLabel):
			state.HasDiscussionLabel = true
		case strings.HasPrefix(normalized, "run:"):
			state.HasRunLabel = true
		}
	}

	for idx, item := range comments {
		if isDiscussionHumanComment(item, normalizedBotLogin) && item.ID > state.MaxHumanCommentID {
			state.MaxHumanCommentID = item.ID
		}
		if isDiscussionAgentReply(item, normalizedBotLogin) {
			state.HasAgentReply = true
			lastAgentReplyIndex = idx
		}
	}

	pendingStart := 0
	if lastAgentReplyIndex >= 0 {
		pendingStart = lastAgentReplyIndex + 1
	}
	for _, item := range comments[pendingStart:] {
		if !isDiscussionHumanComment(item, normalizedBotLogin) {
			continue
		}
		state.PendingHumanComments = append(state.PendingHumanComments, discussionPendingHumanComment{ID: item.ID})
	}
	state.HasHumanAfterAgentReply = lastAgentReplyIndex >= 0 && len(state.PendingHumanComments) > 0

	return state
}

func isDiscussionHumanComment(item discussionIssueCommentResponse, botLogin string) bool {
	login := normalizeDiscussionLogin(item.User.Login)
	userType := strings.ToLower(strings.TrimSpace(item.User.Type))
	return login != botLogin && (userType == "" || userType == "user")
}

func isDiscussionAgentReply(item discussionIssueCommentResponse, botLogin string) bool {
	login := normalizeDiscussionLogin(item.User.Login)
	return login == botLogin && !strings.Contains(item.Body, runStatusCommentMarker)
}

func processedDiscussionCommentIDs(previous []discussionPendingHumanComment, current discussionIssueState) []int64 {
	if len(previous) == 0 {
		return nil
	}

	stillPending := make(map[int64]struct{}, len(current.PendingHumanComments))
	for _, item := range current.PendingHumanComments {
		stillPending[item.ID] = struct{}{}
	}

	processed := make([]int64, 0, len(previous))
	for _, item := range previous {
		if item.ID <= 0 {
			continue
		}
		if _, ok := stillPending[item.ID]; ok {
			continue
		}
		processed = append(processed, item.ID)
	}

	slices.Sort(processed)
	return processed
}

func (s *Service) ensureDiscussionCommentsObserved(ctx context.Context, comments []discussionPendingHumanComment) {
	s.ensureDiscussionIssueCommentReactions(ctx, comments, discussionCommentReactionObserved)
}

func (s *Service) ensureDiscussionCommentsProcessed(ctx context.Context, commentIDs []int64) {
	comments := make([]discussionPendingHumanComment, 0, len(commentIDs))
	for _, commentID := range commentIDs {
		comments = append(comments, discussionPendingHumanComment{ID: commentID})
	}
	s.ensureDiscussionIssueCommentReactions(ctx, comments, discussionCommentReactionProcessed)
}

func (s *Service) ensureDiscussionIssueCommentReactions(ctx context.Context, comments []discussionPendingHumanComment, content string) {
	for _, item := range comments {
		if item.ID <= 0 {
			continue
		}
		if err := s.createDiscussionIssueCommentReaction(ctx, item.ID, content); err != nil {
			s.logger.Warn("discussion comment reaction failed", "comment_id", item.ID, "content", content, "err", err)
		}
	}
}

func (s *Service) createDiscussionIssueCommentReaction(ctx context.Context, commentID int64, content string) error {
	repositoryFullName := strings.TrimSpace(s.cfg.RepositoryFullName)
	reactionContent := strings.TrimSpace(content)
	if repositoryFullName == "" {
		return fmt.Errorf("repository full name is required")
	}
	if commentID <= 0 {
		return fmt.Errorf("comment id must be positive")
	}
	if reactionContent == "" {
		return fmt.Errorf("reaction content is required")
	}

	_, err := runCommandCaptureCombinedOutput(
		ctx,
		"",
		"gh",
		"api",
		"--method",
		"POST",
		fmt.Sprintf("repos/%s/issues/comments/%d/reactions", repositoryFullName, commentID),
		"-H",
		"Accept: application/vnd.github+json",
		"-f",
		"content="+reactionContent,
	)
	if err != nil {
		return fmt.Errorf("create github issue comment reaction: %w", err)
	}
	return nil
}
