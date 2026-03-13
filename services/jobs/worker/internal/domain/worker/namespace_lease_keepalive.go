package worker

import (
	"context"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
)

// keepRunNamespaceLeaseAlive refreshes managed full-env namespace retention for active runs,
// including long waits that outlive the initial namespace TTL.
func (s *Service) keepRunNamespaceLeaseAlive(ctx context.Context, run runqueuerepo.RunningRun) {
	execution := resolveRunExecutionContext(run.RunID, run.ProjectID, run.RunPayload, s.cfg.RunNamespacePrefix)
	if execution.RuntimeMode != agentdomain.RuntimeModeFullEnv {
		return
	}

	runtimePayload := parseRunRuntimePayload(run.RunPayload)
	if isAIRepairRuntimePayload(runtimePayload) {
		return
	}

	runtimeAccessProfile := resolveRuntimeAccessProfile(runtimePayload)
	if runtimeAccessProfile == agentdomain.RuntimeAccessProfileProductionReadOnly {
		return
	}

	leaseCtx := resolveNamespaceLeaseContext(run.RunPayload)
	leaseTTL := s.resolveNamespaceTTL(leaseCtx.AgentKey)
	_, err := s.launcher.EnsureNamespace(ctx, NamespaceSpec{
		RunID:          run.RunID,
		ProjectID:      run.ProjectID,
		IssueNumber:    leaseCtx.IssueNumber,
		AgentKey:       leaseCtx.AgentKey,
		CorrelationID:  run.CorrelationID,
		RuntimeMode:    execution.RuntimeMode,
		Namespace:      execution.Namespace,
		AccessProfile:  runtimeAccessProfile,
		LeaseTTL:       leaseTTL,
		LeaseExpiresAt: s.now().UTC().Add(leaseTTL),
	})
	if err != nil {
		s.logger.Warn("extend namespace lease failed", "run_id", run.RunID, "namespace", execution.Namespace, "err", err)
	}
}
