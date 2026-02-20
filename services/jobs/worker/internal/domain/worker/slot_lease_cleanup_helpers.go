package worker

import (
	"context"
	"strings"

	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
)

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
