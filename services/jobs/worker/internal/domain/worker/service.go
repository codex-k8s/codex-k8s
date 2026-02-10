package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

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
		ref := s.launcher.JobRef(run.RunID)
		state, err := s.launcher.Status(ctx, ref)
		if err != nil {
			s.logger.Error("check run job status failed", "run_id", run.RunID, "job_name", ref.Name, "err", err)
			continue
		}

		switch state {
		case JobStateSucceeded:
			if err := s.finishRun(ctx, run, "succeeded", "run.succeeded", ref, nil); err != nil {
				return err
			}
		case JobStateFailed:
			failure := map[string]any{"reason": "kubernetes job failed"}
			if err := s.finishRun(ctx, run, "failed", "run.failed", ref, failure); err != nil {
				return err
			}
		case JobStateNotFound:
			failure := map[string]any{"reason": "kubernetes job not found"}
			if err := s.finishRun(ctx, run, "failed", "run.failed.job_not_found", ref, failure); err != nil {
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

		ref, err := s.launcher.Launch(ctx, JobSpec{
			RunID:         claimed.RunID,
			CorrelationID: claimed.CorrelationID,
			ProjectID:     claimed.ProjectID,
			SlotNo:        claimed.SlotNo,
		})
		if err != nil {
			s.logger.Error("launch run job failed", "run_id", claimed.RunID, "err", err)
			if finishErr := s.finishRun(ctx, runqueuerepo.RunningRun{
				RunID:         claimed.RunID,
				CorrelationID: claimed.CorrelationID,
				ProjectID:     claimed.ProjectID,
				LearningMode:  claimed.LearningMode,
			}, "failed", "run.failed.launch_error", ref, map[string]any{"error": err.Error()}); finishErr != nil {
				return fmt.Errorf("mark run failed after launch error: %w", finishErr)
			}
			continue
		}

		if err := s.insertEvent(ctx, floweventrepo.InsertParams{
			CorrelationID: claimed.CorrelationID,
			ActorType:     "system",
			ActorID:       s.cfg.WorkerID,
			EventType:     "run.started",
			Payload:       mustJSON(map[string]any{"run_id": claimed.RunID, "project_id": claimed.ProjectID, "slot_no": claimed.SlotNo, "job_name": ref.Name, "job_namespace": ref.Namespace}),
			CreatedAt:     s.now().UTC(),
		}); err != nil {
			return fmt.Errorf("insert run.started event: %w", err)
		}
	}

	return nil
}

func (s *Service) finishRun(
	ctx context.Context,
	run runqueuerepo.RunningRun,
	status string,
	eventType string,
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
		"status":        status,
		"job_name":      ref.Name,
		"job_namespace": ref.Namespace,
	}
	for k, v := range extra {
		payload[k] = v
	}

	if err := s.insertEvent(ctx, floweventrepo.InsertParams{
		CorrelationID: run.CorrelationID,
		ActorType:     "system",
		ActorID:       s.cfg.WorkerID,
		EventType:     eventType,
		Payload:       mustJSON(payload),
		CreatedAt:     finishedAt,
	}); err != nil {
		return fmt.Errorf("insert finish event: %w", err)
	}

	if run.LearningMode && s.feedback != nil {
		explanation := fmt.Sprintf(
			"Learning mode is enabled for this run.\n\n"+
				"Why this is executed as a Kubernetes Job: it provides isolation, reproducibility and clear lifecycle states.\n"+
				"Why we use DB-backed slots: it prevents concurrent workers from overloading a project and makes multi-pod behavior deterministic.\n"+
				"Tradeoffs: Jobs are heavier than in-process execution; DB locking requires careful indexing and timeouts.\n\n"+
				"Result: status=%s, job=%s/%s.",
			status,
			ref.Namespace,
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

	return nil
}

func (s *Service) insertEvent(ctx context.Context, params floweventrepo.InsertParams) error {
	if err := s.events.Insert(ctx, params); err != nil {
		return fmt.Errorf("insert flow event %s for correlation %s: %w", params.EventType, params.CorrelationID, err)
	}
	return nil
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{"error":"payload_marshal_failed"}`)
	}
	return b
}
