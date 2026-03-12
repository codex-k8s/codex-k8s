package worker

import (
	"context"
	"fmt"
	"strings"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/flowevent"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
)

func (s *Service) reclaimStaleRunning(ctx context.Context, limit int) ([]runqueuerepo.RunningRun, error) {
	activeWorkers, ok := s.activeWorkerSet(ctx)
	if !ok || limit <= 0 {
		return nil, nil
	}

	running, err := s.runs.ListRunning(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list running runs for stale lease reclaim: %w", err)
	}

	now := s.now().UTC()
	reclaimed := make([]runqueuerepo.RunningRun, 0)
	for _, run := range running {
		if !shouldAttemptStaleLeaseReclaim(run, s.cfg.WorkerID, activeWorkers, now) {
			continue
		}

		reclaimedRun, claimed, err := s.runs.ReclaimStaleRunning(ctx, runqueuerepo.ReclaimStaleRunningParams{
			RunID:              run.RunID,
			WorkerID:           s.cfg.WorkerID,
			PreviousLeaseOwner: run.LeaseOwner,
			LeaseTTL:           s.cfg.RunLeaseTTL,
		})
		if err != nil {
			return nil, fmt.Errorf("reclaim stale running run %s: %w", run.RunID, err)
		}
		if !claimed {
			continue
		}

		s.logger.Warn("previous lease owner missing", "run_id", run.RunID, "previous_lease_owner", run.LeaseOwner, "lease_until", formatLeaseTimestamp(run.LeaseUntil))
		s.logger.Warn("stale lease detected", "run_id", run.RunID, "previous_lease_owner", run.LeaseOwner, "lease_until", formatLeaseTimestamp(run.LeaseUntil))
		s.logger.Info("stale lease reclaimed", "run_id", run.RunID, "previous_lease_owner", run.LeaseOwner, "worker_id", s.cfg.WorkerID)

		if err := s.insertRunLeaseReclaimEvents(ctx, run, reclaimedRun); err != nil {
			return nil, err
		}

		reclaimed = append(reclaimed, reclaimedRun)
	}

	return reclaimed, nil
}

func (s *Service) activeWorkerSet(ctx context.Context) (map[string]struct{}, bool) {
	if s.workerPresence == nil {
		return nil, false
	}

	activeWorkerIDs, err := s.workerPresence.ListActiveWorkerIDs(ctx)
	if err != nil {
		s.logger.Warn("skip stale lease reclaim because listing active workers failed", "worker_id", s.cfg.WorkerID, "err", err)
		return nil, false
	}

	set := make(map[string]struct{}, len(activeWorkerIDs))
	for _, workerID := range activeWorkerIDs {
		workerID = strings.TrimSpace(workerID)
		if workerID == "" {
			continue
		}
		set[workerID] = struct{}{}
	}

	currentWorkerID := strings.TrimSpace(s.cfg.WorkerID)
	if currentWorkerID == "" {
		return nil, false
	}
	if _, ok := set[currentWorkerID]; !ok {
		s.logger.Warn("skip stale lease reclaim because current worker is missing from active worker set", "worker_id", currentWorkerID, "active_worker_count", len(set))
		return nil, false
	}

	return set, true
}

func shouldAttemptStaleLeaseReclaim(run runqueuerepo.RunningRun, workerID string, activeWorkers map[string]struct{}, now time.Time) bool {
	leaseOwner := strings.TrimSpace(run.LeaseOwner)
	if leaseOwner == "" || leaseOwner == strings.TrimSpace(workerID) {
		return false
	}
	if len(activeWorkers) == 0 {
		return false
	}
	if _, ok := activeWorkers[leaseOwner]; ok {
		return false
	}
	if run.LeaseUntil.IsZero() {
		return false
	}
	return run.LeaseUntil.After(now)
}

func (s *Service) insertRunLeaseReclaimEvents(ctx context.Context, previous runqueuerepo.RunningRun, current runqueuerepo.RunningRun) error {
	payload := encodeRunLeaseEventPayload(runLeaseEventPayload{
		RunID:              current.RunID,
		ProjectID:          current.ProjectID,
		PreviousLeaseOwner: previous.LeaseOwner,
		CurrentLeaseOwner:  s.cfg.WorkerID,
		PreviousLeaseUntil: formatLeaseTimestamp(previous.LeaseUntil),
	})

	for _, eventType := range []floweventdomain.EventType{
		floweventdomain.EventTypeRunLeaseOwnerMissing,
		floweventdomain.EventTypeRunLeaseStaleDetected,
		floweventdomain.EventTypeRunLeaseReclaimed,
	} {
		if err := s.insertEvent(ctx, floweventrepo.InsertParams{
			CorrelationID: current.CorrelationID,
			ActorType:     floweventdomain.ActorTypeSystem,
			ActorID:       floweventdomain.ActorID(s.cfg.WorkerID),
			EventType:     eventType,
			Payload:       payload,
			CreatedAt:     s.now().UTC(),
		}); err != nil {
			return fmt.Errorf("insert %s event: %w", eventType, err)
		}
	}

	return nil
}

func formatLeaseTimestamp(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}
	return ts.UTC().Format(time.RFC3339)
}
