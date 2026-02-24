package worker

import (
	"context"
	"errors"
	"strings"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
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
		if finishErr := s.failRunAfterAgentContextResolve(ctx, run, launchExecution, err); finishErr != nil {
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
