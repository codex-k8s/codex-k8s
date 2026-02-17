package worker

import (
	"context"
	"fmt"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
	valuetypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/value"
)

// tryRecoverMissingRunJob attempts to resume runs that are stuck in "running" state
// without a Kubernetes Job (for example when the worker crashed/errored during runtime preparation).
//
// Returns true when the run was recovered (job launched) or finalized (marked failed).
func (s *Service) tryRecoverMissingRunJob(ctx context.Context, run runqueuerepo.RunningRun, execution valuetypes.RunExecutionContext) (bool, error) {
	if execution.RuntimeMode != agentdomain.RuntimeModeFullEnv {
		return false, nil
	}

	prepareParams := buildPrepareRunEnvironmentParamsFromRunning(run, execution)
	if prepareParams.DeployOnly {
		return false, nil
	}

	// PrepareRunEnvironment is a blocking RPC (it waits for reconciler completion). Use a short timeout
	// to keep reconcile loop responsive while still allowing progress on stuck runs.
	pollTimeout := s.cfg.RuntimePrepareRetryInterval
	if pollTimeout <= 0 {
		pollTimeout = 3 * time.Second
	}
	if pollTimeout < 2*time.Second {
		pollTimeout = 2 * time.Second
	}

	pollCtx, cancel := context.WithTimeout(ctx, pollTimeout)
	prepared, err := s.deployer.PrepareRunEnvironment(pollCtx, prepareParams)
	cancel()
	if err != nil {
		if isRetryableRuntimeDeployError(err) {
			return false, nil
		}
		s.logger.Error("prepare runtime environment for running run failed", "run_id", run.RunID, "err", err)
		if finishErr := s.finishLaunchFailedRun(ctx, run, execution, err, runFailureReasonRuntimeDeployFailed); finishErr != nil {
			return true, finishErr
		}
		return true, nil
	}

	launchExecution := execution
	if resolvedNamespace := sanitizeDNSLabelValue(prepared.Namespace); resolvedNamespace != "" {
		launchExecution.Namespace = resolvedNamespace
	} else {
		// No resolved runtime namespace yet: the run is still preparing.
		return false, nil
	}

	agentCtx, err := resolveRunAgentContext(run.RunPayload, runAgentDefaults{
		DefaultModel:           s.cfg.AgentDefaultModel,
		DefaultReasoningEffort: s.cfg.AgentDefaultReasoningEffort,
		DefaultLocale:          s.cfg.AgentDefaultLocale,
		AllowGPT53:             true,
		LabelCatalog:           s.labels,
	})
	if err != nil {
		s.logger.Error("resolve run agent context failed", "run_id", run.RunID, "err", err)
		eventType := floweventdomain.EventTypeRunFailedLaunchError
		reason := runFailureReasonAgentContextResolve
		if isFailedPreconditionError(err) {
			eventType = floweventdomain.EventTypeRunFailedPrecondition
			reason = runFailureReasonPreconditionFailed
		}
		if finishErr := s.finishRun(ctx, finishRunParams{
			Run:       run,
			Execution: launchExecution,
			Status:    rundomain.StatusFailed,
			EventType: eventType,
			Ref:       s.launcher.JobRef(run.RunID, launchExecution.Namespace),
			Extra: runFinishedEventExtra{
				Error:  err.Error(),
				Reason: reason,
			},
		}); finishErr != nil {
			return true, fmt.Errorf("mark run failed after context resolve error: %w", finishErr)
		}
		return true, nil
	}

	s.logger.Info("recovering run without job by launching into prepared namespace", "run_id", run.RunID, "namespace", launchExecution.Namespace)
	if err := s.launchPreparedFullEnvRunJob(ctx, run, launchExecution, agentCtx); err != nil {
		return true, err
	}

	return true, nil
}
