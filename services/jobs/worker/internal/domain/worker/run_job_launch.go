package worker

import (
	"context"
	"fmt"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/flowevent"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
	valuetypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/value"
)

func (s *Service) launchPreparedFullEnvRunJob(ctx context.Context, run runqueuerepo.RunningRun, execution valuetypes.RunExecutionContext, agentCtx runAgentContext, lease namespaceLeaseSpec) error {
	namespaceSpec := NamespaceSpec{
		RunID:         run.RunID,
		ProjectID:     run.ProjectID,
		IssueNumber:   lease.IssueNumber,
		AgentKey:      lease.AgentKey,
		CorrelationID: run.CorrelationID,
		RuntimeMode:   execution.RuntimeMode,
		Namespace:     execution.Namespace,
	}

	if execution.RuntimeMode == agentdomain.RuntimeModeFullEnv {
		ttl := lease.TTL
		if ttl <= 0 {
			ttl = s.cfg.DefaultNamespaceTTL
		}
		namespaceSpec.LeaseTTL = ttl
		namespaceSpec.LeaseExpiresAt = s.now().UTC().Add(ttl)

		ensureResult, err := s.launcher.EnsureNamespace(ctx, namespaceSpec)
		if err != nil {
			s.logger.Error(
				"prepare run namespace failed",
				"run_id", run.RunID,
				"namespace", execution.Namespace,
				"runtime_mode", execution.RuntimeMode,
				"err", err,
			)
			if finishErr := s.finishLaunchFailedRun(ctx, run, execution, err, runFailureReasonNamespacePrepareFailed); finishErr != nil {
				return fmt.Errorf("mark run failed after namespace prepare error: %w", finishErr)
			}
			return nil
		}
		leaseExpiresAt := ensureResult.LeaseExpiresAt
		if leaseExpiresAt.IsZero() {
			leaseExpiresAt = namespaceSpec.LeaseExpiresAt
		}

		if err := s.insertNamespaceLifecycleEvent(ctx, namespaceLifecycleEventParams{
			CorrelationID: run.CorrelationID,
			EventType:     floweventdomain.EventTypeRunNamespacePrepared,
			RunID:         run.RunID,
			ProjectID:     run.ProjectID,
			Execution:     execution,
		}); err != nil {
			return fmt.Errorf("insert run.namespace.prepared event: %w", err)
		}

		namespaceTTLEventType := floweventdomain.EventTypeRunNamespaceTTLScheduled
		if ensureResult.Reused {
			namespaceTTLEventType = floweventdomain.EventTypeRunNamespaceTTLExtended
		}
		if err := s.insertNamespaceLifecycleEvent(ctx, namespaceLifecycleEventParams{
			CorrelationID: run.CorrelationID,
			EventType:     namespaceTTLEventType,
			RunID:         run.RunID,
			ProjectID:     run.ProjectID,
			Execution:     execution,
			Extra: namespaceLifecycleEventExtra{
				NamespaceLeaseTTL:       ttl,
				NamespaceLeaseExpiresAt: leaseExpiresAt,
				NamespaceReused:         ensureResult.Reused,
			},
		}); err != nil {
			return fmt.Errorf("insert %s event: %w", namespaceTTLEventType, err)
		}
	}

	issuedMCPToken, err := s.mcpTokens.IssueRunMCPToken(ctx, IssueMCPTokenParams{
		RunID:       run.RunID,
		Namespace:   execution.Namespace,
		RuntimeMode: execution.RuntimeMode,
	})
	if err != nil {
		s.logger.Error("issue run mcp token failed", "run_id", run.RunID, "err", err)
		if finishErr := s.finishLaunchFailedRun(ctx, run, execution, err, runFailureReasonMCPTokenIssueFailed); finishErr != nil {
			return fmt.Errorf("mark run failed after mcp token issue error: %w", finishErr)
		}
		return nil
	}

	if _, err := s.runStatus.UpsertRunStatusComment(ctx, RunStatusCommentParams{
		RunID:           run.RunID,
		Phase:           RunStatusPhaseCreated,
		RuntimeMode:     string(execution.RuntimeMode),
		Namespace:       execution.Namespace,
		TriggerKind:     agentCtx.TriggerKind,
		PromptLocale:    agentCtx.PromptTemplateLocale,
		Model:           agentCtx.Model,
		ReasoningEffort: agentCtx.ReasoningEffort,
		RunStatus:       string(rundomain.StatusRunning),
	}); err != nil {
		s.logger.Warn("upsert run status comment (created) failed", "run_id", run.RunID, "err", err)
	}

	if err := s.insertEvent(ctx, floweventrepo.InsertParams{
		CorrelationID: run.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorID(s.cfg.WorkerID),
		EventType:     floweventdomain.EventTypeRunProfileResolved,
		Payload: encodeRunProfileResolvedEventPayload(runProfileResolvedEventPayload{
			RunID:              run.RunID,
			ProjectID:          run.ProjectID,
			RepositoryFullName: agentCtx.RepositoryFullName,
			IssueNumber:        agentCtx.IssueNumber,
			PullRequestNumber:  agentCtx.ExistingPRNumber,
			TriggerKind:        agentCtx.TriggerKind,
			Model:              agentCtx.Model,
			ModelSource:        agentCtx.ModelSource,
			ReasoningEffort:    agentCtx.ReasoningEffort,
			ReasoningSource:    agentCtx.ReasoningSource,
		}),
		CreatedAt: s.now().UTC(),
	}); err != nil {
		return fmt.Errorf("insert run.profile.resolved event: %w", err)
	}

	ref, err := s.launcher.Launch(ctx, JobSpec{
		RunID:                  run.RunID,
		CorrelationID:          run.CorrelationID,
		ProjectID:              run.ProjectID,
		SlotNo:                 run.SlotNo,
		RuntimeMode:            execution.RuntimeMode,
		Namespace:              execution.Namespace,
		ControlPlaneGRPCTarget: s.cfg.ControlPlaneGRPCTarget,
		MCPBaseURL:             s.cfg.ControlPlaneMCPBaseURL,
		MCPBearerToken:         issuedMCPToken.Token,
		RepositoryFullName:     agentCtx.RepositoryFullName,
		IssueNumber:            agentCtx.IssueNumber,
		TriggerKind:            agentCtx.TriggerKind,
		TriggerLabel:           agentCtx.TriggerLabel,
		TargetBranch:           agentCtx.TargetBranch,
		ExistingPRNumber:       agentCtx.ExistingPRNumber,
		AgentKey:               agentCtx.AgentKey,
		AgentModel:             agentCtx.Model,
		AgentReasoningEffort:   agentCtx.ReasoningEffort,
		PromptTemplateKind:     agentCtx.PromptTemplateKind,
		PromptTemplateSource:   agentCtx.PromptTemplateSource,
		PromptTemplateLocale:   agentCtx.PromptTemplateLocale,
		StateInReviewLabel:     s.cfg.StateInReviewLabel,
		BaseBranch:             s.cfg.AgentBaseBranch,
		OpenAIAPIKey:           s.cfg.OpenAIAPIKey,
		Context7APIKey:         s.cfg.Context7APIKey,
		GitBotToken:            s.cfg.GitBotToken,
		AgentDisplayName:       agentCtx.AgentDisplayName,
		GitBotUsername:         s.cfg.GitBotUsername,
		GitBotMail:             s.cfg.GitBotMail,
	})
	if err != nil {
		s.logger.Error("launch run job failed", "run_id", run.RunID, "err", err)
		if finishErr := s.finishRun(ctx, finishRunParams{
			Run:       run,
			Execution: execution,
			Status:    rundomain.StatusFailed,
			EventType: floweventdomain.EventTypeRunFailedLaunchError,
			Ref:       ref,
			Extra: runFinishedEventExtra{
				Error: err.Error(),
			},
		}); finishErr != nil {
			return fmt.Errorf("mark run failed after launch error: %w", finishErr)
		}
		return nil
	}

	if err := s.insertEvent(ctx, floweventrepo.InsertParams{
		CorrelationID: run.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorID(s.cfg.WorkerID),
		EventType:     floweventdomain.EventTypeRunStarted,
		Payload: encodeRunStartedEventPayload(runStartedEventPayload{
			RunID:                run.RunID,
			ProjectID:            run.ProjectID,
			SlotNo:               run.SlotNo,
			JobName:              ref.Name,
			JobNamespace:         ref.Namespace,
			RuntimeMode:          execution.RuntimeMode,
			RepositoryFullName:   agentCtx.RepositoryFullName,
			AgentKey:             agentCtx.AgentKey,
			IssueNumber:          agentCtx.IssueNumber,
			TriggerKind:          agentCtx.TriggerKind,
			TriggerLabel:         agentCtx.TriggerLabel,
			Model:                agentCtx.Model,
			ModelSource:          agentCtx.ModelSource,
			ReasoningEffort:      agentCtx.ReasoningEffort,
			ReasoningSource:      agentCtx.ReasoningSource,
			PromptTemplateKind:   agentCtx.PromptTemplateKind,
			PromptTemplateSource: agentCtx.PromptTemplateSource,
			PromptTemplateLocale: agentCtx.PromptTemplateLocale,
			BaseBranch:           s.cfg.AgentBaseBranch,
		}),
		CreatedAt: s.now().UTC(),
	}); err != nil {
		return fmt.Errorf("insert run.started event: %w", err)
	}

	if _, err := s.runStatus.UpsertRunStatusComment(ctx, RunStatusCommentParams{
		RunID:           run.RunID,
		Phase:           RunStatusPhaseStarted,
		JobName:         ref.Name,
		JobNamespace:    ref.Namespace,
		RuntimeMode:     string(execution.RuntimeMode),
		Namespace:       execution.Namespace,
		TriggerKind:     agentCtx.TriggerKind,
		PromptLocale:    agentCtx.PromptTemplateLocale,
		Model:           agentCtx.Model,
		ReasoningEffort: agentCtx.ReasoningEffort,
		RunStatus:       string(rundomain.StatusRunning),
	}); err != nil {
		s.logger.Warn("upsert run status comment (started) failed", "run_id", run.RunID, "err", err)
	}

	return nil
}
