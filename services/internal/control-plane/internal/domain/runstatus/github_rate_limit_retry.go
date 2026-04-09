package runstatus

import (
	"context"
	"fmt"
	"strings"

	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

// RetryGitHubRateLimitComment replays one delayed run-status comment update after rate-limit recovery.
func (s *Service) RetryGitHubRateLimitComment(ctx context.Context, payload valuetypes.GitHubRateLimitRunStatusCommentRetryPayload) error {
	if s == nil {
		return fmt.Errorf("run status service is not configured")
	}

	phase := Phase(strings.TrimSpace(payload.Phase))
	if phase == "" {
		return fmt.Errorf("github rate-limit run-status retry payload.phase is required")
	}

	_, err := s.UpsertRunStatusComment(ctx, UpsertCommentParams{
		RunID:                    strings.TrimSpace(payload.RunID),
		Phase:                    phase,
		JobName:                  strings.TrimSpace(payload.JobName),
		JobNamespace:             strings.TrimSpace(payload.JobNamespace),
		RuntimeMode:              strings.TrimSpace(payload.RuntimeMode),
		Namespace:                strings.TrimSpace(payload.Namespace),
		TriggerKind:              strings.TrimSpace(payload.TriggerKind),
		PromptLocale:             strings.TrimSpace(payload.PromptLocale),
		Model:                    strings.TrimSpace(payload.Model),
		ReasoningEffort:          strings.TrimSpace(payload.ReasoningEffort),
		RunStatus:                strings.TrimSpace(payload.RunStatus),
		CodexAuthVerificationURL: strings.TrimSpace(payload.CodexAuthVerificationURL),
		CodexAuthUserCode:        strings.TrimSpace(payload.CodexAuthUserCode),
		Deleted:                  payload.Deleted,
		AlreadyDeleted:           payload.AlreadyDeleted,
	})
	if err != nil {
		return fmt.Errorf("retry github rate-limit run status comment: %w", err)
	}
	return nil
}
