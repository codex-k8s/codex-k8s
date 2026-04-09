package worker

import (
	"context"
	"fmt"
	"strings"

	runqueuerepo "github.com/codex-k8s/kodex/services/jobs/worker/internal/domain/repository/runqueue"
)

func (s *Service) scheduleInteractionResume(ctx context.Context, runID string, interactionID string, resumeCorrelationID string) error {
	if s.runs == nil {
		return fmt.Errorf("run queue repository is not configured")
	}

	sourceRunID := strings.TrimSpace(runID)
	if sourceRunID == "" {
		return fmt.Errorf("interaction resume scheduling requires run_id for interaction %s", strings.TrimSpace(interactionID))
	}
	correlationID := strings.TrimSpace(resumeCorrelationID)
	if correlationID == "" {
		return fmt.Errorf("interaction resume scheduling requires resume_correlation_id for interaction %s", strings.TrimSpace(interactionID))
	}

	_, err := s.runs.CreatePendingResumeIfAbsent(ctx, runqueuerepo.CreatePendingResumeParams{
		SourceRunID:   sourceRunID,
		CorrelationID: correlationID,
	})
	if err != nil {
		return fmt.Errorf("schedule interaction resume for %s: %w", strings.TrimSpace(interactionID), err)
	}
	return nil
}
