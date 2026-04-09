package githubratelimit

import (
	"context"
	"fmt"
	"time"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
)

func (s *Service) appendResumeScheduledEvidence(
	ctx context.Context,
	wait Wait,
	signalID string,
	signalOrigin enumtypes.GitHubRateLimitSignalOrigin,
	observedAt time.Time,
	nextStepKind enumtypes.GitHubRateLimitNextStepKind,
) error {
	if _, err := s.waits.AppendEvidence(ctx, querytypes.GitHubRateLimitWaitEvidenceCreateParams{
		WaitID:       wait.ID,
		EventKind:    enumtypes.GitHubRateLimitEvidenceEventResumeScheduled,
		SignalID:     signalID,
		SignalOrigin: signalOrigin,
		PayloadJSON: marshalJSONPayload(waitResumeScheduledEvidencePayload{
			WaitID:          wait.ID,
			ResumeNotBefore: wait.ResumeNotBefore,
			AttemptsUsed:    wait.AutoResumeAttemptsUsed,
			MaxAttempts:     wait.MaxAutoResumeAttempts,
			NextStepKind:    nextStepKind,
			SignalOrigin:    signalOrigin,
		}),
		ObservedAt: observedAt,
	}); err != nil {
		return fmt.Errorf("append github rate-limit resume schedule evidence: %w", err)
	}
	return nil
}
