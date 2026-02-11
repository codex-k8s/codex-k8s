package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/flowevent"
	learningfeedbackrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/learningfeedback"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
)

// Config defines worker run-loop behavior.
type Config struct {
	// WorkerID uniquely identifies current worker instance.
	WorkerID string
	// ClaimLimit limits number of pending runs claimed per tick.
	ClaimLimit int
	// RunningCheckLimit limits running runs reconciled per tick.
	RunningCheckLimit int
	// SlotsPerProject defines slot pool size per project scope.
	SlotsPerProject int
	// SlotLeaseTTL defines maximum duration of slot ownership.
	SlotLeaseTTL time.Duration

	// ProjectLearningModeDefault is applied when the worker auto-creates projects from webhook payloads.
	ProjectLearningModeDefault bool
	// RunNamespacePrefix defines prefix for full-env run namespaces.
	RunNamespacePrefix string
	// CleanupFullEnvNamespace enables namespace cleanup after run completion.
	CleanupFullEnvNamespace bool
}

// Service orchestrates pending runs to Kubernetes Jobs and final statuses.
type Service struct {
	cfg      Config
	runs     runqueuerepo.Repository
	events   floweventrepo.Repository
	feedback learningfeedbackrepo.Repository
	launcher Launcher
	logger   *slog.Logger
	now      func() time.Time
}

// NewService creates worker orchestrator instance.
func NewService(
	cfg Config,
	runs runqueuerepo.Repository,
	events floweventrepo.Repository,
	feedback learningfeedbackrepo.Repository,
	launcher Launcher,
	logger *slog.Logger,
) *Service {
	if cfg.ClaimLimit <= 0 {
		cfg.ClaimLimit = 1
	}
	if cfg.RunningCheckLimit <= 0 {
		cfg.RunningCheckLimit = 100
	}
	if cfg.SlotsPerProject <= 0 {
		cfg.SlotsPerProject = 1
	}
	if cfg.SlotLeaseTTL <= 0 {
		cfg.SlotLeaseTTL = 5 * time.Minute
	}
	if cfg.WorkerID == "" {
		cfg.WorkerID = "worker"
	}
	if cfg.RunNamespacePrefix == "" {
		cfg.RunNamespacePrefix = defaultRunNamespacePrefix
	}

	return &Service{
		cfg:      cfg,
		runs:     runs,
		events:   events,
		feedback: feedback,
		launcher: launcher,
		logger:   logger,
		now:      time.Now,
	}
}

// Tick executes one reconciliation iteration.
func (s *Service) Tick(ctx context.Context) error {
	if err := s.reconcileRunning(ctx); err != nil {
		return fmt.Errorf("reconcile running runs: %w", err)
	}
	if err := s.launchPending(ctx); err != nil {
		return fmt.Errorf("launch pending runs: %w", err)
	}
	return nil
}

func (s *Service) reconcileRunning(ctx context.Context) error {
	running, err := s.runs.ListRunning(ctx, s.cfg.RunningCheckLimit)
	if err != nil {
		return fmt.Errorf("list running runs: %w", err)
	}

	for _, run := range running {
		execution := resolveRunExecutionContext(run.RunID, run.ProjectID, run.RunPayload, s.cfg.RunNamespacePrefix)
		ref := s.launcher.JobRef(run.RunID, execution.Namespace)
		state, err := s.launcher.Status(ctx, ref)
		if err != nil {
			s.logger.Error("check run job status failed", "run_id", run.RunID, "job_name", ref.Name, "err", err)
			continue
		}

		switch state {
		case JobStateSucceeded:
			if err := s.finishRun(ctx, run, execution, rundomain.StatusSucceeded, floweventdomain.EventTypeRunSucceeded, ref, nil); err != nil {
				return err
			}
		case JobStateFailed:
			failure := map[string]any{"reason": "kubernetes job failed"}
			if err := s.finishRun(ctx, run, execution, rundomain.StatusFailed, floweventdomain.EventTypeRunFailed, ref, failure); err != nil {
				return err
			}
		case JobStateNotFound:
			failure := map[string]any{"reason": "kubernetes job not found"}
			if err := s.finishRun(ctx, run, execution, rundomain.StatusFailed, floweventdomain.EventTypeRunFailedJobNotFound, ref, failure); err != nil {
				return err
			}
		case JobStatePending, JobStateRunning:
			continue
		default:
			s.logger.Warn("unknown job state", "run_id", run.RunID, "state", state)
		}
	}

	return nil
}

func (s *Service) launchPending(ctx context.Context) error {
	for range s.cfg.ClaimLimit {
		claimed, ok, err := s.runs.ClaimNextPending(ctx, runqueuerepo.ClaimParams{
			WorkerID:                   s.cfg.WorkerID,
			SlotsPerProject:            s.cfg.SlotsPerProject,
			LeaseTTL:                   s.cfg.SlotLeaseTTL,
			ProjectLearningModeDefault: s.cfg.ProjectLearningModeDefault,
		})
		if err != nil {
			return fmt.Errorf("claim pending run: %w", err)
		}
		if !ok {
			return nil
		}

		execution := resolveRunExecutionContext(claimed.RunID, claimed.ProjectID, claimed.RunPayload, s.cfg.RunNamespacePrefix)
		namespaceSpec := NamespaceSpec{
			RunID:         claimed.RunID,
			ProjectID:     claimed.ProjectID,
			CorrelationID: claimed.CorrelationID,
			RuntimeMode:   execution.RuntimeMode,
			Namespace:     execution.Namespace,
		}
		if execution.RuntimeMode == agentdomain.RuntimeModeFullEnv {
			if err := s.launcher.EnsureNamespace(ctx, namespaceSpec); err != nil {
				s.logger.Error(
					"prepare run namespace failed",
					"run_id", claimed.RunID,
					"namespace", execution.Namespace,
					"runtime_mode", execution.RuntimeMode,
					"err", err,
				)
				if finishErr := s.finishRun(ctx, runqueuerepo.RunningRun{
					RunID:         claimed.RunID,
					CorrelationID: claimed.CorrelationID,
					ProjectID:     claimed.ProjectID,
					LearningMode:  claimed.LearningMode,
					RunPayload:    claimed.RunPayload,
				}, execution, rundomain.StatusFailed, floweventdomain.EventTypeRunFailedLaunchError, JobRef{}, map[string]any{
					"error":        err.Error(),
					"reason":       "namespace_prepare_failed",
					"runtime_mode": execution.RuntimeMode,
					"namespace":    execution.Namespace,
				}); finishErr != nil {
					return fmt.Errorf("mark run failed after namespace prepare error: %w", finishErr)
				}
				continue
			}

			if err := s.insertNamespaceLifecycleEvent(
				ctx,
				claimed.CorrelationID,
				floweventdomain.EventTypeRunNamespacePrepared,
				claimed.RunID,
				claimed.ProjectID,
				execution,
				nil,
			); err != nil {
				return fmt.Errorf("insert run.namespace.prepared event: %w", err)
			}
		}

		ref, err := s.launcher.Launch(ctx, JobSpec{
			RunID:         claimed.RunID,
			CorrelationID: claimed.CorrelationID,
			ProjectID:     claimed.ProjectID,
			SlotNo:        claimed.SlotNo,
			RuntimeMode:   execution.RuntimeMode,
			Namespace:     execution.Namespace,
		})
		if err != nil {
			s.logger.Error("launch run job failed", "run_id", claimed.RunID, "err", err)
			if finishErr := s.finishRun(ctx, runqueuerepo.RunningRun{
				RunID:         claimed.RunID,
				CorrelationID: claimed.CorrelationID,
				ProjectID:     claimed.ProjectID,
				LearningMode:  claimed.LearningMode,
				RunPayload:    claimed.RunPayload,
			}, execution, rundomain.StatusFailed, floweventdomain.EventTypeRunFailedLaunchError, ref, map[string]any{
				"error":        err.Error(),
				"runtime_mode": execution.RuntimeMode,
				"namespace":    execution.Namespace,
			}); finishErr != nil {
				return fmt.Errorf("mark run failed after launch error: %w", finishErr)
			}
			continue
		}

		if err := s.insertEvent(ctx, floweventrepo.InsertParams{
			CorrelationID: claimed.CorrelationID,
			ActorType:     floweventdomain.ActorTypeSystem,
			ActorID:       floweventdomain.ActorID(s.cfg.WorkerID),
			EventType:     floweventdomain.EventTypeRunStarted,
			Payload: mustJSON(map[string]any{
				"run_id":        claimed.RunID,
				"project_id":    claimed.ProjectID,
				"slot_no":       claimed.SlotNo,
				"job_name":      ref.Name,
				"job_namespace": ref.Namespace,
				"runtime_mode":  execution.RuntimeMode,
			}),
			CreatedAt: s.now().UTC(),
		}); err != nil {
			return fmt.Errorf("insert run.started event: %w", err)
		}
	}

	return nil
}

func (s *Service) finishRun(
	ctx context.Context,
	run runqueuerepo.RunningRun,
	execution runExecutionContext,
	status rundomain.Status,
	eventType floweventdomain.EventType,
	ref JobRef,
	extra map[string]any,
) error {
	finishedAt := s.now().UTC()
	updated, err := s.runs.FinishRun(ctx, runqueuerepo.FinishParams{
		RunID:      run.RunID,
		ProjectID:  run.ProjectID,
		Status:     status,
		FinishedAt: finishedAt,
	})
	if err != nil {
		return fmt.Errorf("finish run %s as %s: %w", run.RunID, status, err)
	}
	if !updated {
		return nil
	}

	payload := map[string]any{
		"run_id":        run.RunID,
		"project_id":    run.ProjectID,
		"status":        string(status),
		"job_name":      ref.Name,
		"job_namespace": ref.Namespace,
		"runtime_mode":  execution.RuntimeMode,
	}
	if execution.Namespace != "" {
		payload["namespace"] = execution.Namespace
	}
	for k, v := range extra {
		payload[k] = v
	}

	if err := s.insertEvent(ctx, floweventrepo.InsertParams{
		CorrelationID: run.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorID(s.cfg.WorkerID),
		EventType:     eventType,
		Payload:       mustJSON(payload),
		CreatedAt:     finishedAt,
	}); err != nil {
		return fmt.Errorf("insert finish event: %w", err)
	}

	if run.LearningMode && s.feedback != nil {
		namespace := ref.Namespace
		if execution.Namespace != "" {
			namespace = execution.Namespace
		}
		explanation := fmt.Sprintf(
			"Learning mode is enabled for this run.\n\n"+
				"Why this is executed as a Kubernetes Job: it provides isolation, reproducibility and clear lifecycle states.\n"+
				"Why we use DB-backed slots: it prevents concurrent workers from overloading a project and makes multi-pod behavior deterministic.\n"+
				"Tradeoffs: Jobs are heavier than in-process execution; DB locking requires careful indexing and timeouts.\n\n"+
				"Result: status=%s, job=%s/%s.",
			status,
			namespace,
			ref.Name,
		)
		if err := s.feedback.Insert(ctx, learningfeedbackrepo.InsertParams{
			RunID:       run.RunID,
			Kind:        "inline",
			Explanation: explanation,
		}); err != nil {
			s.logger.Error("insert learning feedback failed", "run_id", run.RunID, "err", err)
		}
	}

	if execution.RuntimeMode == agentdomain.RuntimeModeFullEnv && execution.Namespace != "" && s.cfg.CleanupFullEnvNamespace {
		cleanupSpec := NamespaceSpec{
			RunID:         run.RunID,
			ProjectID:     run.ProjectID,
			CorrelationID: run.CorrelationID,
			RuntimeMode:   execution.RuntimeMode,
			Namespace:     execution.Namespace,
		}
		cleanupErr := s.launcher.CleanupNamespace(ctx, cleanupSpec)
		if cleanupErr != nil {
			s.logger.Error(
				"cleanup run namespace failed",
				"run_id", run.RunID,
				"namespace", execution.Namespace,
				"err", cleanupErr,
			)
			if err := s.insertNamespaceLifecycleEvent(
				ctx,
				run.CorrelationID,
				floweventdomain.EventTypeRunNamespaceCleanupFailed,
				run.RunID,
				run.ProjectID,
				execution,
				map[string]any{"error": cleanupErr.Error()},
			); err != nil {
				s.logger.Error("insert run.namespace.cleanup_failed event failed", "run_id", run.RunID, "err", err)
			}
		} else {
			if err := s.insertNamespaceLifecycleEvent(
				ctx,
				run.CorrelationID,
				floweventdomain.EventTypeRunNamespaceCleaned,
				run.RunID,
				run.ProjectID,
				execution,
				nil,
			); err != nil {
				s.logger.Error("insert run.namespace.cleaned event failed", "run_id", run.RunID, "err", err)
			}
		}
	}

	return nil
}

func (s *Service) insertEvent(ctx context.Context, params floweventrepo.InsertParams) error {
	if err := s.events.Insert(ctx, params); err != nil {
		return fmt.Errorf("insert flow event %s for correlation %s: %w", params.EventType, params.CorrelationID, err)
	}
	return nil
}

func (s *Service) insertNamespaceLifecycleEvent(
	ctx context.Context,
	correlationID string,
	eventType floweventdomain.EventType,
	runID string,
	projectID string,
	execution runExecutionContext,
	extra map[string]any,
) error {
	return s.insertEvent(ctx, floweventrepo.InsertParams{
		CorrelationID: correlationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorID(s.cfg.WorkerID),
		EventType:     eventType,
		Payload:       mustJSON(buildNamespaceLifecyclePayload(runID, projectID, execution, extra)),
		CreatedAt:     s.now().UTC(),
	})
}

func buildNamespaceLifecyclePayload(runID string, projectID string, execution runExecutionContext, extra map[string]any) map[string]any {
	payload := map[string]any{
		"run_id":       runID,
		"project_id":   projectID,
		"runtime_mode": execution.RuntimeMode,
		"namespace":    execution.Namespace,
	}
	for key, value := range extra {
		payload[key] = value
	}
	return payload
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{"error":"payload_marshal_failed"}`)
	}
	return b
}
