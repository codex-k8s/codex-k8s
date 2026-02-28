package worker

import (
	"context"
	"fmt"
	"strings"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/flowevent"
	learningfeedbackrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/learningfeedback"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
	valuetypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/value"
)

// finishRun persists terminal run state, emits flow events, and finalizes runtime namespace lifecycle.
func (s *Service) finishRun(ctx context.Context, params finishRunParams) error {
	finishedAt := s.now().UTC()
	updated, err := s.runs.FinishRun(ctx, runqueuerepo.FinishParams{
		RunID:      params.Run.RunID,
		ProjectID:  params.Run.ProjectID,
		LeaseOwner: s.cfg.WorkerID,
		Status:     params.Status,
		FinishedAt: finishedAt,
	})
	if err != nil {
		return fmt.Errorf("finish run %s as %s: %w", params.Run.RunID, params.Status, err)
	}
	if !updated {
		return nil
	}

	payload := runFinishedEventPayload{
		RunID:        params.Run.RunID,
		ProjectID:    params.Run.ProjectID,
		Status:       params.Status,
		JobName:      params.Ref.Name,
		JobNamespace: params.Ref.Namespace,
		RuntimeMode:  params.Execution.RuntimeMode,
		Namespace:    params.Execution.Namespace,
		Error:        params.Extra.Error,
		Reason:       params.Extra.Reason,
	}

	if err := s.insertEvent(ctx, floweventrepo.InsertParams{
		CorrelationID: params.Run.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorID(s.cfg.WorkerID),
		EventType:     params.EventType,
		Payload:       encodeRunFinishedEventPayload(payload),
		CreatedAt:     finishedAt,
	}); err != nil {
		return fmt.Errorf("insert finish event: %w", err)
	}

	if _, err := s.runStatus.UpsertRunStatusComment(ctx, RunStatusCommentParams{
		RunID:        params.Run.RunID,
		Phase:        RunStatusPhaseFinished,
		JobName:      params.Ref.Name,
		JobNamespace: params.Ref.Namespace,
		RuntimeMode:  string(params.Execution.RuntimeMode),
		Namespace:    params.Execution.Namespace,
		RunStatus:    string(params.Status),
	}); err != nil {
		s.logger.Warn("upsert run status comment (finished) failed", "run_id", params.Run.RunID, "err", err)
	}

	if params.Run.LearningMode && s.feedback != nil {
		namespace := params.Ref.Namespace
		if params.Execution.Namespace != "" {
			namespace = params.Execution.Namespace
		}
		explanation := fmt.Sprintf(
			"Learning mode is enabled for this run.\n\n"+
				"Why this is executed as a Kubernetes Job: it provides isolation, reproducibility and clear lifecycle states.\n"+
				"Why we use DB-backed slots: it prevents concurrent workers from overloading a project and makes multi-pod behavior deterministic.\n"+
				"Tradeoffs: Jobs are heavier than in-process execution; DB locking requires careful indexing and timeouts.\n\n"+
				"Result: status=%s, job=%s/%s.",
			params.Status,
			namespace,
			params.Ref.Name,
		)
		if err := s.feedback.Insert(ctx, learningfeedbackrepo.InsertParams{
			RunID:       params.Run.RunID,
			Kind:        learningfeedbackrepo.KindInline,
			Explanation: explanation,
		}); err != nil {
			s.logger.Error("insert learning feedback failed", "run_id", params.Run.RunID, "err", err)
		}
	}

	if params.Execution.RuntimeMode == agentdomain.RuntimeModeFullEnv &&
		params.Execution.Namespace != "" &&
		!params.SkipNamespaceCleanup {
		s.upsertNamespaceStatusComment(ctx, params, false, "upsert run status comment (namespace retained by ttl policy) failed")
	}

	return nil
}

func (s *Service) finishLaunchFailedRun(ctx context.Context, run runqueuerepo.RunningRun, execution valuetypes.RunExecutionContext, failure error, reason runFailureReason) error {
	return s.finishRun(ctx, finishRunParams{
		Run:       run,
		Execution: execution,
		Status:    rundomain.StatusFailed,
		EventType: floweventdomain.EventTypeRunFailedLaunchError,
		Extra: runFinishedEventExtra{
			Error:  failure.Error(),
			Reason: reason,
		},
	})
}

func (s *Service) upsertNamespaceStatusComment(ctx context.Context, params finishRunParams, deleted bool, warnMessage string) {
	if _, err := s.runStatus.UpsertRunStatusComment(ctx, RunStatusCommentParams{
		RunID:        params.Run.RunID,
		Phase:        RunStatusPhaseNamespaceDeleted,
		JobName:      params.Ref.Name,
		JobNamespace: params.Ref.Namespace,
		RuntimeMode:  string(params.Execution.RuntimeMode),
		Namespace:    params.Execution.Namespace,
		RunStatus:    string(params.Status),
		Deleted:      deleted,
	}); err != nil {
		s.logger.Warn(warnMessage, "run_id", params.Run.RunID, "err", err)
	}
}

// insertEvent persists one flow event with contextual error wrapping.
func (s *Service) insertEvent(ctx context.Context, params floweventrepo.InsertParams) error {
	if err := s.events.Insert(ctx, params); err != nil {
		return fmt.Errorf("insert flow event %s for correlation %s: %w", params.EventType, params.CorrelationID, err)
	}
	return nil
}

// insertNamespaceLifecycleEvent records namespace lifecycle transitions in flow_events.
func (s *Service) insertNamespaceLifecycleEvent(ctx context.Context, params namespaceLifecycleEventParams) error {
	return s.insertEvent(ctx, floweventrepo.InsertParams{
		CorrelationID: params.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorID(s.cfg.WorkerID),
		EventType:     params.EventType,
		Payload: encodeNamespaceLifecycleEventPayload(namespaceLifecycleEventPayload{
			RunID:          params.RunID,
			ProjectID:      params.ProjectID,
			RuntimeMode:    params.Execution.RuntimeMode,
			Namespace:      params.Execution.Namespace,
			Error:          params.Extra.Error,
			Reason:         params.Extra.Reason,
			CleanupCommand: params.Extra.CleanupCommand,
			NamespaceLeaseTTL: func() string {
				if params.Extra.NamespaceLeaseTTL <= 0 {
					return ""
				}
				return params.Extra.NamespaceLeaseTTL.String()
			}(),
			NamespaceLeaseExpiresAt: func() string {
				if params.Extra.NamespaceLeaseExpiresAt.IsZero() {
					return ""
				}
				return params.Extra.NamespaceLeaseExpiresAt.UTC().Format(time.RFC3339)
			}(),
			NamespaceReused: params.Extra.NamespaceReused,
		}),
		CreatedAt: s.now().UTC(),
	})
}

// runningRunFromClaimed reuses claimed fields for failure finalization paths before the next reconcile tick.
func runningRunFromClaimed(claimed runqueuerepo.ClaimedRun) runqueuerepo.RunningRun {
	return runqueuerepo.RunningRun{
		RunID:         claimed.RunID,
		CorrelationID: claimed.CorrelationID,
		ProjectID:     claimed.ProjectID,
		SlotID:        claimed.SlotID,
		SlotNo:        claimed.SlotNo,
		LearningMode:  claimed.LearningMode,
		RunPayload:    claimed.RunPayload,
	}
}

func isFailedPreconditionError(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(err.Error()), "failed_precondition:")
}
