package webhook

import (
	"context"
	"strings"

	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

func (s *Service) resolveDiscussionTrigger(issueLabels []githubLabelRecord, fallbackKind webhookdomain.TriggerKind, source string) (issueRunTrigger, triggerConflictResult) {
	conflictingLabels := s.triggerLabels.collectIssueTriggerLabels(issueLabels)
	if len(conflictingLabels) > 1 {
		return issueRunTrigger{}, triggerConflictResult{
			ConflictingLabels: conflictingLabels,
			IgnoreReason:      "discussion_mode_trigger_label_conflict",
			SuggestedLabels:   conflictingLabels,
		}
	}

	kind := fallbackKind
	if len(conflictingLabels) == 1 {
		resolvedKind, ok := s.triggerLabels.resolveKind(conflictingLabels[0])
		if !ok {
			return issueRunTrigger{}, triggerConflictResult{}
		}
		kind = resolvedKind
	}
	if kind == "" {
		kind = webhookdomain.TriggerKindDev
	}

	return issueRunTrigger{
			Source:         strings.TrimSpace(source),
			Label:          strings.TrimSpace(s.triggerLabels.withDefaults().ModeDiscussion),
			Kind:           kind,
			DiscussionMode: true,
		}, triggerConflictResult{
			ConflictingLabels: conflictingLabels,
		}
}

func (s *Service) hasActiveDiscussionRun(ctx context.Context, projectID string, repositoryFullName string, issueNumber int64) (bool, error) {
	if s.agentRuns == nil || strings.TrimSpace(projectID) == "" || strings.TrimSpace(repositoryFullName) == "" || issueNumber <= 0 {
		return false, nil
	}

	items, err := s.agentRuns.SearchRecentByProjectIssueOrPullRequest(ctx, projectID, repositoryFullName, issueNumber, 0, 20)
	if err != nil {
		return false, err
	}

	discussionLabel := normalizeLabelToken(s.triggerLabels.withDefaults().ModeDiscussion)
	for _, item := range items {
		if normalizeLabelToken(item.TriggerLabel) != discussionLabel {
			continue
		}
		switch rundomain.Status(strings.TrimSpace(item.Status)) {
		case rundomain.StatusPending, rundomain.StatusRunning:
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) isDiscussionCommentSenderAllowed(sender githubActorRecord) bool {
	senderType := normalizeLabelToken(sender.Type)
	if senderType != "" && senderType != gitHubSenderTypeUser {
		return false
	}
	login := normalizeLabelToken(sender.Login)
	if login == "" {
		return false
	}
	if s.gitBotUsername != "" && login == s.gitBotUsername {
		return false
	}
	return true
}
