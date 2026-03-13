package worker

import (
	"context"
	"strings"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/flowevent"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
	valuetypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/value"
)

const (
	runtimeReuseReasonFingerprintMatch     = "fingerprint_match"
	runtimeReuseReasonEvaluationFailed     = "reuse_evaluation_failed"
	runtimeReuseReasonNamespaceNotManaged  = "namespace_not_managed"
	runtimeReuseReasonNamespaceTerminating = "namespace_terminating"
	runtimeReuseReasonNamespaceMismatch    = "namespace_mismatch"
)

type runtimeReuseResolution struct {
	execution     valuetypes.RunExecutionContext
	prepareParams PrepareRunEnvironmentParams
	reusable      bool
}

func (s *Service) resolveRuntimeReuseForRevise(
	ctx context.Context,
	run runqueuerepo.RunningRun,
	execution valuetypes.RunExecutionContext,
	prepareParams PrepareRunEnvironmentParams,
	leaseCtx namespaceLeaseContext,
	triggerKind string,
) (runtimeReuseResolution, error) {
	result := runtimeReuseResolution{
		execution:     execution,
		prepareParams: prepareParams,
	}
	if execution.RuntimeMode != agentdomain.RuntimeModeFullEnv {
		return result, nil
	}
	if prepareParams.DeployOnly || prepareParams.Namespace != "" {
		return result, nil
	}
	if !leaseCtx.IsRevise || leaseCtx.IssueNumber <= 0 || strings.TrimSpace(leaseCtx.AgentKey) == "" {
		return result, nil
	}

	reusableNamespace, found, err := s.launcher.FindReusableNamespace(ctx, NamespaceReuseLookup{
		ProjectID:   run.ProjectID,
		IssueNumber: leaseCtx.IssueNumber,
		AgentKey:    leaseCtx.AgentKey,
		Now:         s.now().UTC(),
	})
	if err != nil {
		s.logger.Warn(
			"resolve reusable namespace for revise run failed",
			"run_id", run.RunID,
			"project_id", run.ProjectID,
			"issue_number", leaseCtx.IssueNumber,
			"agent_key", leaseCtx.AgentKey,
			"err", err,
		)
		return result, nil
	}
	if !found {
		return result, nil
	}

	namespace := sanitizeDNSLabelValue(reusableNamespace.Namespace)
	if namespace == "" {
		return result, nil
	}

	evaluated, err := s.deployer.EvaluateRuntimeReuse(ctx, EvaluateRuntimeReuseParams{
		RunID:              run.RunID,
		ProjectID:          run.ProjectID,
		IssueNumber:        leaseCtx.IssueNumber,
		AgentKey:           leaseCtx.AgentKey,
		RuntimeMode:        string(execution.RuntimeMode),
		Namespace:          namespace,
		TargetEnv:          prepareParams.TargetEnv,
		SlotNo:             prepareParams.SlotNo,
		RepositoryFullName: prepareParams.RepositoryFullName,
		ServicesYAMLPath:   prepareParams.ServicesYAMLPath,
		BuildRef:           prepareParams.BuildRef,
		DeployOnly:         prepareParams.DeployOnly,
	})
	if err != nil {
		s.logger.Warn(
			"evaluate runtime reuse failed, fallback to runtime redeploy",
			"run_id", run.RunID,
			"namespace", namespace,
			"err", err,
		)
		if eventErr := s.insertRuntimeReuseEvent(ctx, run, floweventdomain.EventTypeRunNamespaceReuseFallback, runtimeReuseEventPayload{
			RunID:       run.RunID,
			ProjectID:   run.ProjectID,
			Namespace:   namespace,
			IssueNumber: leaseCtx.IssueNumber,
			AgentKey:    leaseCtx.AgentKey,
			TriggerKind: triggerKind,
			Reason:      runtimeReuseReasonEvaluationFailed,
		}); eventErr != nil {
			return result, eventErr
		}
		result.prepareParams.Namespace = namespace
		result.execution.Namespace = namespace
		return result, nil
	}

	eventNamespace := namespace
	if evaluatedNamespace := sanitizeDNSLabelValue(evaluated.Namespace); evaluatedNamespace != "" {
		eventNamespace = evaluatedNamespace
	}
	if targetEnv := strings.TrimSpace(evaluated.TargetEnv); targetEnv != "" {
		result.prepareParams.TargetEnv = targetEnv
	}

	eventPayload := runtimeReuseEventPayload{
		RunID:             run.RunID,
		ProjectID:         run.ProjectID,
		Namespace:         eventNamespace,
		IssueNumber:       leaseCtx.IssueNumber,
		AgentKey:          leaseCtx.AgentKey,
		TriggerKind:       triggerKind,
		Reason:            strings.TrimSpace(evaluated.Reason),
		EffectiveBuildRef: strings.TrimSpace(evaluated.EffectiveBuildRef),
		FingerprintHash:   strings.TrimSpace(evaluated.FingerprintHash),
	}
	if evaluated.Reusable {
		result.prepareParams.Namespace = eventNamespace
		result.execution.Namespace = eventNamespace
		if eventPayload.Reason == "" {
			eventPayload.Reason = runtimeReuseReasonFingerprintMatch
		}
		if err := s.insertRuntimeReuseEvent(ctx, run, floweventdomain.EventTypeRunNamespaceReuseFastPath, eventPayload); err != nil {
			return result, err
		}
		result.reusable = true
		return result, nil
	}
	if shouldResetNamespaceForReuseFallback(eventPayload.Reason) {
		result.prepareParams.Namespace = ""
		result.execution.Namespace = ""
	} else {
		result.prepareParams.Namespace = eventNamespace
		result.execution.Namespace = eventNamespace
	}

	if err := s.insertRuntimeReuseEvent(ctx, run, floweventdomain.EventTypeRunNamespaceReuseFallback, eventPayload); err != nil {
		return result, err
	}
	return result, nil
}

func shouldResetNamespaceForReuseFallback(reason string) bool {
	switch strings.TrimSpace(reason) {
	case runtimeReuseReasonNamespaceNotManaged, runtimeReuseReasonNamespaceTerminating, runtimeReuseReasonNamespaceMismatch:
		return true
	default:
		return false
	}
}

func (s *Service) insertRuntimeReuseEvent(
	ctx context.Context,
	run runqueuerepo.RunningRun,
	eventType floweventdomain.EventType,
	payload runtimeReuseEventPayload,
) error {
	return s.insertEvent(ctx, floweventrepo.InsertParams{
		CorrelationID: run.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorID(s.cfg.WorkerID),
		EventType:     eventType,
		Payload:       encodeRuntimeReuseEventPayload(payload),
		CreatedAt:     s.now().UTC(),
	})
}
