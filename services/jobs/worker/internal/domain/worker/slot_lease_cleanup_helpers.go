package worker

import (
	"context"
	"fmt"
	"strings"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
)

const cleanupPeerCheckListLimit = 2000

// keepRunSlotLeaseAlive extends slot lease for active full-env runs to prevent accidental re-claim.
func (s *Service) keepRunSlotLeaseAlive(ctx context.Context, run runqueuerepo.RunningRun) {
	if run.SlotNo <= 0 {
		return
	}

	projectID := strings.TrimSpace(run.ProjectID)
	if projectID == "" {
		return
	}

	updated, err := s.runs.ExtendLease(ctx, runqueuerepo.ExtendLeaseParams{
		RunID:     run.RunID,
		ProjectID: projectID,
		LeaseTTL:  s.cfg.SlotLeaseTTL,
	})
	if err != nil {
		s.logger.Warn("extend slot lease failed", "run_id", run.RunID, "project_id", run.ProjectID, "slot_no", run.SlotNo, "err", err)
		return
	}
	if !updated {
		s.logger.Warn("extend slot lease skipped because slot lease is missing", "run_id", run.RunID, "project_id", run.ProjectID, "slot_no", run.SlotNo)
	}
}

// findRunningPeerOnSameSlot returns another running run id sharing the same project slot.
func (s *Service) findRunningPeerOnSameSlot(ctx context.Context, run runqueuerepo.RunningRun) (string, error) {
	if run.SlotNo <= 0 {
		return "", nil
	}

	projectID := strings.TrimSpace(run.ProjectID)
	if projectID == "" {
		return "", nil
	}

	limit := s.cfg.RunningCheckLimit
	if limit < cleanupPeerCheckListLimit {
		limit = cleanupPeerCheckListLimit
	}
	running, err := s.runs.ListRunning(ctx, limit)
	if err != nil {
		return "", fmt.Errorf("list running runs for slot peer check: %w", err)
	}

	for _, candidate := range running {
		if candidate.RunID == run.RunID {
			continue
		}
		if candidate.SlotNo <= 0 {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(candidate.ProjectID), projectID) {
			continue
		}
		if candidate.SlotNo != run.SlotNo {
			continue
		}
		return candidate.RunID, nil
	}

	return "", nil
}

func (s *Service) emitNamespaceCleanupSkipped(ctx context.Context, params finishRunParams, reason namespaceCleanupSkipReason, errText string) {
	if err := s.insertNamespaceLifecycleEvent(ctx, namespaceLifecycleEventParams{
		CorrelationID: params.Run.CorrelationID,
		EventType:     floweventdomain.EventTypeRunNamespaceCleanupSkipped,
		RunID:         params.Run.RunID,
		ProjectID:     params.Run.ProjectID,
		Execution:     params.Execution,
		Extra: namespaceLifecycleEventExtra{
			Error:          strings.TrimSpace(errText),
			Reason:         reason,
			CleanupCommand: buildNamespaceCleanupCommand(params.Execution.Namespace),
		},
	}); err != nil {
		s.logger.Error("insert run.namespace.cleanup_skipped event failed", "run_id", params.Run.RunID, "err", err)
	}
}
