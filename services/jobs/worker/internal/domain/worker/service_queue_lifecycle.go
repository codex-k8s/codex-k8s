package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
)

// Tick executes one reconciliation iteration.
func (s *Service) Tick(ctx context.Context) error {
	if err := s.cleanupExpiredNamespaces(ctx); err != nil {
		return fmt.Errorf("cleanup expired namespaces: %w", err)
	}
	if err := s.reconcileRunning(ctx); err != nil {
		return fmt.Errorf("reconcile running runs: %w", err)
	}
	if err := s.launchPending(ctx); err != nil {
		return fmt.Errorf("launch pending runs: %w", err)
	}
	return nil
}

// reconcileRunning polls active runs and finalizes those with terminal Kubernetes job states.
func (s *Service) reconcileRunning(ctx context.Context) error {
	running, err := s.runs.ClaimRunning(ctx, runqueuerepo.ClaimRunningParams{
		WorkerID: s.cfg.WorkerID,
		LeaseTTL: s.cfg.RunLeaseTTL,
		Limit:    s.cfg.RunningCheckLimit,
	})
	if err != nil {
		return fmt.Errorf("claim running runs: %w", err)
	}

	for _, run := range running {
		s.keepRunSlotLeaseAlive(ctx, run)

		execution := resolveRunExecutionContext(run.RunID, run.ProjectID, run.RunPayload, s.cfg.RunNamespacePrefix)
		runtimePayload := parseRunRuntimePayload(run.RunPayload)
		deployOnlyRun := runtimePayload.Runtime != nil && runtimePayload.Runtime.DeployOnly
		aiRepairRun := isAIRepairRuntimePayload(runtimePayload)
		if aiRepairRun {
			execution.Namespace = s.resolveAIRepairNamespace(execution.Namespace)
		}

		if deployOnlyRun {
			prepareParams := buildPrepareRunEnvironmentParamsFromRunning(run, execution)
			prepared, ready, err := s.prepareRuntimeEnvironmentPoll(ctx, prepareParams)
			if err != nil {
				if errors.Is(err, errRuntimeDeployTaskCanceled) {
					if cancelErr := s.finishRuntimePrepareCanceledRun(ctx, run, execution, true); cancelErr != nil {
						return cancelErr
					}
					continue
				}
				s.logger.Error("prepare runtime environment for running deploy-only run failed", "run_id", run.RunID, "err", err)
				if finishErr := s.finishLaunchFailedRun(ctx, run, execution, err, runFailureReasonRuntimeDeployFailed); finishErr != nil {
					return finishErr
				}
				continue
			}
			if !ready {
				continue
			}

			if err := s.finishRun(ctx, finishRunParams{
				Run:                  run,
				Execution:            applyPreparedNamespace(execution, prepared.Namespace),
				Status:               rundomain.StatusSucceeded,
				EventType:            floweventdomain.EventTypeRunSucceeded,
				SkipNamespaceCleanup: true,
			}); err != nil {
				return err
			}
			continue
		}

		ref := s.launcher.JobRef(run.RunID, execution.Namespace)
		state, err := s.launcher.Status(ctx, ref)
		if err != nil {
			s.logger.Error("check run job status failed", "run_id", run.RunID, "job_name", ref.Name, "err", err)
			continue
		}

		if state == JobStateNotFound {
			// Full-env runs may be launched into persistent slot namespaces, while run payload keeps
			// the default namespace strategy (`codex-issue-*`). Resolve the actual namespace by label
			// to avoid failing runs with "job not found" after preparation succeeded.
			resolved, ok, err := s.launcher.FindRunJobRefByRunID(ctx, run.RunID)
			if err != nil {
				s.logger.Warn("resolve run job ref by run id failed", "run_id", run.RunID, "err", err)
			} else if ok {
				ref = resolved
				execution = applyPreparedNamespace(execution, resolved.Namespace)
				if execution.Namespace != "" {
					ref.Namespace = execution.Namespace
				}

				state, err = s.launcher.Status(ctx, ref)
				if err != nil {
					s.logger.Error("check run job status failed", "run_id", run.RunID, "job_name", ref.Name, "err", err)
					continue
				}
			}
		}

		switch state {
		case JobStateSucceeded:
			if err := s.finishRun(ctx, finishRunParams{
				Run:       run,
				Execution: execution,
				Status:    rundomain.StatusSucceeded,
				EventType: floweventdomain.EventTypeRunSucceeded,
				Ref:       ref,
			}); err != nil {
				return err
			}
		case JobStateFailed:
			if err := s.finishRun(ctx, finishRunParams{
				Run:       run,
				Execution: execution,
				Status:    rundomain.StatusFailed,
				EventType: floweventdomain.EventTypeRunFailed,
				Ref:       ref,
				Extra: runFinishedEventExtra{
					Reason: runFailureReasonKubernetesJobFailed,
				},
			}); err != nil {
				return err
			}
		case JobStateNotFound:
			recovered, err := s.tryRecoverMissingRunJob(ctx, run, execution)
			if err != nil {
				return err
			}
			if recovered {
				continue
			}

			// We mark runs as "running" when they are claimed, but full-env runs may spend
			// significant time in runtime preparation before the actual job exists.
			// With multiple worker replicas this prevents another worker from failing the run
			// while the claiming worker is still preparing the environment.
			if s.shouldIgnoreJobNotFound(run) {
				continue
			}
			if err := s.finishRun(ctx, finishRunParams{
				Run:       run,
				Execution: execution,
				Status:    rundomain.StatusFailed,
				EventType: floweventdomain.EventTypeRunFailedJobNotFound,
				Ref:       ref,
				Extra: runFinishedEventExtra{
					Reason: runFailureReasonKubernetesJobNotFound,
				},
			}); err != nil {
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

func (s *Service) shouldIgnoreJobNotFound(run runqueuerepo.RunningRun) bool {
	startedAt := run.StartedAt
	if startedAt.IsZero() {
		return true
	}

	grace := s.cfg.RuntimePrepareRetryTimeout
	if grace <= 0 {
		grace = 30 * time.Second
	}
	grace += 5 * time.Second

	now := s.now().UTC()
	if now.Before(startedAt) {
		return true
	}
	return now.Sub(startedAt) < grace
}
