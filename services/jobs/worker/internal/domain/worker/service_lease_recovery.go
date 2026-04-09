package worker

import (
	"context"
	"fmt"
	"time"

	floweventdomain "github.com/codex-k8s/kodex/libs/go/domain/flowevent"
	floweventrepo "github.com/codex-k8s/kodex/services/jobs/worker/internal/domain/repository/flowevent"
	runqueuerepo "github.com/codex-k8s/kodex/services/jobs/worker/internal/domain/repository/runqueue"
)

func (s *Service) releaseStaleRunningLeases(ctx context.Context) error {
	params := runqueuerepo.ReleaseStaleLeasesParams{
		Limit: s.cfg.StaleLeaseSweepLimit,
	}

	activeWorkerIDs, err := s.launcher.ListWorkerPodNames(ctx, s.cfg.WorkerPodNamespace)
	if err != nil {
		s.logger.Warn("list active worker pods failed; falling back to lease ttl for missing owners", "namespace", s.cfg.WorkerPodNamespace, "err", err)
	} else {
		params.ReleaseMissingOwners = true
		params.ActiveWorkerIDs = activeWorkerIDs
	}

	released, err := s.runs.ReleaseStaleLeases(ctx, params)
	if err != nil {
		return fmt.Errorf("release stale running leases: %w", err)
	}

	for _, item := range released {
		if err := s.insertWorkerHeartbeatMissedEvent(ctx, item); err != nil {
			return err
		}
		if err := s.insertRunLeaseStaleEvent(ctx, item, floweventdomain.EventTypeRunLeaseDetectedStale); err != nil {
			return err
		}
		if err := s.insertRunLeaseStaleEvent(ctx, item, floweventdomain.EventTypeRunLeaseReleased); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) insertWorkerHeartbeatMissedEvent(ctx context.Context, item runqueuerepo.ReleasedStaleLease) error {
	return s.insertEvent(ctx, floweventrepo.InsertParams{
		CorrelationID: item.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorID(s.cfg.WorkerID),
		EventType:     floweventdomain.EventTypeWorkerInstanceHeartbeatMissed,
		Payload: encodeWorkerHeartbeatMissedEventPayload(workerHeartbeatMissedEventPayload{
			RunID:              item.RunID,
			ProjectID:          item.ProjectID,
			WorkerID:           item.PreviousLeaseOwner,
			WorkerStatus:       item.WorkerStatus,
			WorkerHeartbeatAt:  formatEventTime(item.WorkerHeartbeatAt),
			WorkerExpiresAt:    formatEventTime(item.WorkerExpiresAt),
			PreviousLeaseUntil: formatEventTime(item.PreviousLeaseUntil),
		}),
		CreatedAt: s.now().UTC(),
	})
}

func (s *Service) insertRunLeaseStaleEvent(ctx context.Context, item runqueuerepo.ReleasedStaleLease, eventType floweventdomain.EventType) error {
	return s.insertEvent(ctx, floweventrepo.InsertParams{
		CorrelationID: item.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorID(s.cfg.WorkerID),
		EventType:     eventType,
		Payload: encodeRunLeaseStaleEventPayload(runLeaseStaleEventPayload{
			RunID:              item.RunID,
			ProjectID:          item.ProjectID,
			PreviousLeaseOwner: item.PreviousLeaseOwner,
			PreviousLeaseUntil: formatEventTime(item.PreviousLeaseUntil),
			WorkerStatus:       item.WorkerStatus,
			WorkerHeartbeatAt:  formatEventTime(item.WorkerHeartbeatAt),
			WorkerExpiresAt:    formatEventTime(item.WorkerExpiresAt),
		}),
		CreatedAt: s.now().UTC(),
	})
}

func (s *Service) insertRunLeaseRecoveredEvent(ctx context.Context, run runqueuerepo.RunningRun) error {
	runtimePayload := parseRunRuntimePayload(run.RunPayload)
	execution := resolveRunExecutionContext(run.RunID, run.ProjectID, run.RunPayload, s.cfg.RunNamespacePrefix)
	triggerKind := ""
	if runtimePayload.Trigger != nil {
		triggerKind = string(runtimePayload.Trigger.Kind)
	}

	return s.insertEvent(ctx, floweventrepo.InsertParams{
		CorrelationID: run.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorID(s.cfg.WorkerID),
		EventType:     floweventdomain.EventTypeRunReclaimedAfterStaleLease,
		Payload: encodeRunLeaseRecoveredEventPayload(runLeaseRecoveredEventPayload{
			RunID:          run.RunID,
			ProjectID:      run.ProjectID,
			WorkerID:       s.cfg.WorkerID,
			RuntimeMode:    string(execution.RuntimeMode),
			Namespace:      execution.Namespace,
			TriggerKind:    triggerKind,
			DiscussionMode: runtimePayload.DiscussionMode,
		}),
		CreatedAt: s.now().UTC(),
	})
}

func formatEventTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
