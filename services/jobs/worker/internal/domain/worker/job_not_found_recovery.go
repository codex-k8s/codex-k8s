package worker

import (
	"context"
	"errors"
	"strings"

	agentdomain "github.com/codex-k8s/kodex/libs/go/domain/agent"
	runqueuerepo "github.com/codex-k8s/kodex/services/jobs/worker/internal/domain/repository/runqueue"
	valuetypes "github.com/codex-k8s/kodex/services/jobs/worker/internal/domain/types/value"
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
	runtimePayload := parseRunRuntimePayload(run.RunPayload)
	runtimeAccessProfile := resolveRuntimeAccessProfile(runtimePayload)
	if runtimeAccessProfile == agentdomain.RuntimeAccessProfileProductionReadOnly {
		return s.recoverMissingProductionReadOnlyRunJob(ctx, run, execution, runtimeAccessProfile)
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
		if finishErr := s.failRunAfterAgentContextResolve(ctx, run, execution, err); finishErr != nil {
			return true, finishErr
		}
		return true, nil
	}

	leaseCtx := resolveNamespaceLeaseContext(run.RunPayload)
	if leaseCtx.AgentKey == "" {
		leaseCtx.AgentKey = strings.ToLower(strings.TrimSpace(agentCtx.AgentKey))
	}
	if leaseCtx.IssueNumber <= 0 {
		leaseCtx.IssueNumber = agentCtx.IssueNumber
	}
	reuseResolution, err := s.resolveRuntimeReuseForRevise(ctx, run, execution, prepareParams, leaseCtx, agentCtx.TriggerKind)
	if err != nil {
		return true, err
	}
	execution = reuseResolution.execution
	prepareParams = reuseResolution.prepareParams
	if reuseResolution.reusable {
		leaseTTL := s.resolveNamespaceTTL(leaseCtx.AgentKey)
		s.logger.Info("recovering run without job via runtime reuse fast-path", "run_id", run.RunID, "namespace", execution.Namespace)
		if err := s.launchPreparedRunWorkload(ctx, run, execution, agentCtx, namespaceLeaseSpec{
			AgentKey:    leaseCtx.AgentKey,
			IssueNumber: leaseCtx.IssueNumber,
			TTL:         leaseTTL,
		}, runLaunchOptions{}); err != nil {
			return true, err
		}
		return true, nil
	}

	prepared, ready, err := s.prepareRuntimeEnvironmentPoll(ctx, prepareParams)
	if err != nil {
		if errors.Is(err, errRuntimeDeployTaskCanceled) {
			if cancelErr := s.finishRuntimePrepareCanceledRun(ctx, run, execution, false); cancelErr != nil {
				return true, cancelErr
			}
			return true, nil
		}
		s.logger.Error("prepare runtime environment for running run failed", "run_id", run.RunID, "err", err)
		if finishErr := s.finishLaunchFailedRun(ctx, run, execution, err, runFailureReasonRuntimeDeployFailed); finishErr != nil {
			return true, finishErr
		}
		return true, nil
	}
	if !ready {
		// Runtime deploy is still preparing (or transiently unavailable). Keep run in
		// running state and retry on next tick without flipping to failed.
		return true, nil
	}

	launchExecution := applyPreparedNamespace(execution, prepared.Namespace)
	if launchExecution.Namespace == "" {
		// No resolved runtime namespace yet: the run is still preparing.
		return false, nil
	}
	leaseTTL := s.resolveNamespaceTTL(leaseCtx.AgentKey)

	s.logger.Info("recovering run without job by launching into prepared namespace", "run_id", run.RunID, "namespace", launchExecution.Namespace)
	if err := s.launchPreparedRunWorkload(ctx, run, launchExecution, agentCtx, namespaceLeaseSpec{
		AgentKey:    leaseCtx.AgentKey,
		IssueNumber: leaseCtx.IssueNumber,
		TTL:         leaseTTL,
	}, runLaunchOptions{}); err != nil {
		return true, err
	}

	return true, nil
}

func (s *Service) recoverMissingProductionReadOnlyRunJob(
	ctx context.Context,
	run runqueuerepo.RunningRun,
	execution valuetypes.RunExecutionContext,
	runtimeAccessProfile agentdomain.RuntimeAccessProfile,
) (bool, error) {
	launchExecution := execution
	launchExecution.Namespace = s.resolveProductionReadonlyNamespace(launchExecution.Namespace)
	if launchExecution.Namespace == "" {
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
		if finishErr := s.failRunAfterAgentContextResolve(ctx, run, launchExecution, err); finishErr != nil {
			return true, finishErr
		}
		return true, nil
	}

	s.logger.Info(
		"recovering production-readonly run without job by relaunching in existing namespace",
		"run_id", run.RunID,
		"namespace", launchExecution.Namespace,
	)
	if err := s.launchPreparedRunWorkload(ctx, run, launchExecution, agentCtx, namespaceLeaseSpec{}, runLaunchOptions{
		SkipNamespacePreparation: true,
		RuntimeAccessProfile:     runtimeAccessProfile,
	}); err != nil {
		return true, err
	}

	return true, nil
}
