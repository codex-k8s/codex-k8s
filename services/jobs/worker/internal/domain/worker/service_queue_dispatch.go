package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
)

// launchPending claims pending runs, prepares runtime namespace (for full-env), and launches run workloads.
func (s *Service) launchPending(ctx context.Context) error {
	for range s.cfg.ClaimLimit {
		claimed, ok, err := s.runs.ClaimNextPending(ctx, runqueuerepo.ClaimParams{
			WorkerID:                   s.cfg.WorkerID,
			SlotsPerProject:            s.cfg.SlotsPerProject,
			LeaseTTL:                   s.cfg.SlotLeaseTTL,
			RunLeaseTTL:                s.cfg.RunLeaseTTL,
			ProjectLearningModeDefault: s.cfg.ProjectLearningModeDefault,
		})
		if err != nil {
			return fmt.Errorf("claim pending run: %w", err)
		}
		if !ok {
			return nil
		}

		execution := resolveRunExecutionContext(claimed.RunID, claimed.ProjectID, claimed.RunPayload, s.cfg.RunNamespacePrefix)
		runPayload := parseRunRuntimePayload(claimed.RunPayload)
		runningRun := runningRunFromClaimed(claimed)
		prepareParams := buildPrepareRunEnvironmentParams(claimed, execution)
		deployOnlyRun := prepareParams.DeployOnly
		aiRepairRun := isAIRepairRuntimePayload(runPayload)
		if aiRepairRun {
			execution.Namespace = s.resolveAIRepairNamespace(execution.Namespace)
		}

		leaseCtx := resolveNamespaceLeaseContext(claimed.RunPayload)
		leaseTTL := s.cfg.DefaultNamespaceTTL
		triggerKind := ""
		var agentCtx runAgentContext

		if deployOnlyRun {
			if runPayload.Trigger != nil {
				triggerKind = string(runPayload.Trigger.Kind)
			}
		} else {
			agentCtx, err = resolveRunAgentContext(claimed.RunPayload, runAgentDefaults{
				DefaultModel:           s.cfg.AgentDefaultModel,
				DefaultReasoningEffort: s.cfg.AgentDefaultReasoningEffort,
				DefaultLocale:          s.cfg.AgentDefaultLocale,
				AllowGPT53:             true,
				LabelCatalog:           s.labels,
			})
			if err != nil {
				s.logger.Error("resolve run agent context failed", "run_id", claimed.RunID, "err", err)
				if finishErr := s.failRunAfterAgentContextResolve(ctx, runningRun, execution, err); finishErr != nil {
					return finishErr
				}
				continue
			}
			triggerKind = agentCtx.TriggerKind
			if leaseCtx.AgentKey == "" {
				leaseCtx.AgentKey = strings.ToLower(strings.TrimSpace(agentCtx.AgentKey))
			}
			if leaseCtx.IssueNumber <= 0 {
				leaseCtx.IssueNumber = agentCtx.IssueNumber
			}
			if !leaseCtx.IsRevise {
				leaseCtx.IsRevise = resolvePromptTemplateKindForTrigger(agentCtx.TriggerKind) == promptTemplateKindRevise
			}
			leaseTTL = s.resolveNamespaceTTL(leaseCtx.AgentKey)

			if execution.RuntimeMode == agentdomain.RuntimeModeFullEnv &&
				leaseCtx.IsRevise &&
				prepareParams.Namespace == "" &&
				leaseCtx.IssueNumber > 0 &&
				leaseCtx.AgentKey != "" {
				reusableNamespace, found, reuseErr := s.launcher.FindReusableNamespace(ctx, NamespaceReuseLookup{
					ProjectID:   runningRun.ProjectID,
					IssueNumber: leaseCtx.IssueNumber,
					AgentKey:    leaseCtx.AgentKey,
					Now:         s.now().UTC(),
				})
				if reuseErr != nil {
					s.logger.Warn(
						"resolve reusable namespace for revise run failed",
						"run_id", runningRun.RunID,
						"project_id", runningRun.ProjectID,
						"issue_number", leaseCtx.IssueNumber,
						"agent_key", leaseCtx.AgentKey,
						"err", reuseErr,
					)
				} else if found {
					prepareParams.Namespace = reusableNamespace.Namespace
					execution.Namespace = reusableNamespace.Namespace
				}
			}
		}

		if aiRepairRun {
			if err := s.launchPreparedRunWorkload(ctx, runningRun, execution, agentCtx, namespaceLeaseSpec{}, runLaunchOptions{
				ServiceAccountName: s.cfg.AIRepairServiceAccount,
			}); err != nil {
				return err
			}
			continue
		}

		if execution.RuntimeMode != agentdomain.RuntimeModeFullEnv && !deployOnlyRun {
			if err := s.launchPreparedRunWorkload(ctx, runningRun, execution, agentCtx, namespaceLeaseSpec{}, runLaunchOptions{}); err != nil {
				return err
			}
			continue
		}

		if _, err := s.runStatus.UpsertRunStatusComment(ctx, RunStatusCommentParams{
			RunID:       runningRun.RunID,
			Phase:       RunStatusPhasePreparingRuntime,
			RuntimeMode: string(execution.RuntimeMode),
			Namespace:   execution.Namespace,
			TriggerKind: triggerKind,
			RunStatus:   string(rundomain.StatusRunning),
		}); err != nil {
			s.logger.Warn("upsert run status comment (preparing runtime) failed", "run_id", runningRun.RunID, "err", err)
		}

		prepared, ready, err := s.prepareRuntimeEnvironmentPoll(ctx, prepareParams)
		if err != nil {
			if errors.Is(err, errRuntimeDeployTaskCanceled) {
				if cancelErr := s.finishRuntimePrepareCanceledRun(ctx, runningRun, execution, deployOnlyRun); cancelErr != nil {
					return cancelErr
				}
				continue
			}
			s.logger.Error("prepare runtime environment failed", "run_id", claimed.RunID, "err", err)
			if finishErr := s.finishLaunchFailedRun(ctx, runningRun, execution, err, runFailureReasonRuntimeDeployFailed); finishErr != nil {
				return fmt.Errorf("mark run failed after runtime deploy error: %w", finishErr)
			}
			continue
		}
		if !ready {
			continue
		}

		launchExecution := applyPreparedNamespace(execution, prepared.Namespace)
		if deployOnlyRun {
			if err := s.finishRun(ctx, finishRunParams{
				Run:                  runningRun,
				Execution:            launchExecution,
				Status:               rundomain.StatusSucceeded,
				EventType:            floweventdomain.EventTypeRunSucceeded,
				SkipNamespaceCleanup: true,
			}); err != nil {
				return fmt.Errorf("finish deploy-only run: %w", err)
			}
			continue
		}

		if err := s.launchPreparedRunWorkload(ctx, runningRun, launchExecution, agentCtx, namespaceLeaseSpec{
			AgentKey:    leaseCtx.AgentKey,
			IssueNumber: leaseCtx.IssueNumber,
			TTL:         leaseTTL,
		}, runLaunchOptions{}); err != nil {
			return err
		}
	}

	return nil
}
