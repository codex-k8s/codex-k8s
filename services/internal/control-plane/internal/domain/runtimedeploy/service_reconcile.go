package runtimedeploy

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

// ReconcileNext claims one pending deploy task and applies desired state.
func (s *Service) ReconcileNext(ctx context.Context, leaseOwner string, leaseTTL time.Duration) (bool, error) {
	leaseOwner = strings.TrimSpace(leaseOwner)
	if leaseOwner == "" {
		return false, fmt.Errorf("runtime deploy reconciler lease owner is required")
	}
	if leaseTTL < time.Second {
		leaseTTL = time.Second
	}

	leaseTTLString := fmt.Sprintf("%d seconds", int64(leaseTTL.Seconds()))
	task, ok, err := s.tasks.ClaimNext(ctx, runtimedeploytaskrepo.ClaimParams{
		LeaseOwner: leaseOwner,
		LeaseTTL:   leaseTTLString,
	})
	if err != nil {
		return false, fmt.Errorf("claim runtime deploy task: %w", err)
	}
	if !ok {
		return false, nil
	}
	s.appendTaskLogBestEffort(ctx, task.RunID, "reconcile", "info", "Task claimed by reconciler "+leaseOwner)

	renewCtx, cancelRenew := context.WithCancel(ctx)
	renewDone := make(chan struct{})
	go func() {
		defer close(renewDone)

		// Keep the lease short for fast recovery when the reconciler dies during self-deploy,
		// but renew it while we are actively processing the task.
		interval := leaseTTL / 2
		if interval > 30*time.Second {
			interval = 30 * time.Second
		}
		if interval < time.Second {
			interval = time.Second
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-renewCtx.Done():
				return
			case <-ticker.C:
				updated, err := s.tasks.RenewLease(renewCtx, runtimedeploytaskrepo.RenewLeaseParams{
					RunID:      task.RunID,
					LeaseOwner: leaseOwner,
					LeaseTTL:   leaseTTLString,
				})
				if err != nil {
					s.logger.Error("renew runtime deploy task lease failed", "run_id", task.RunID, "lease_owner", leaseOwner, "err", err)
					continue
				}
				if !updated {
					s.logger.Warn("runtime deploy task lease lost while renewing", "run_id", task.RunID, "lease_owner", leaseOwner)
					cancelRenew()
					return
				}
			}
		}
	}()

	result, runErr := s.applyDesiredState(renewCtx, PrepareParams{
		RunID:              task.RunID,
		RuntimeMode:        task.RuntimeMode,
		Namespace:          task.Namespace,
		TargetEnv:          task.TargetEnv,
		SlotNo:             task.SlotNo,
		RepositoryFullName: task.RepositoryFullName,
		ServicesYAMLPath:   task.ServicesYAMLPath,
		BuildRef:           task.BuildRef,
		DeployOnly:         task.DeployOnly,
	})
	cancelRenew()
	<-renewDone
	if runErr != nil {
		s.appendTaskLogBestEffort(ctx, task.RunID, "reconcile", "error", "Task failed: "+runErr.Error())
		if errors.Is(runErr, context.Canceled) || errors.Is(runErr, context.DeadlineExceeded) {
			if ctx.Err() != nil {
				return true, runErr
			}
			if s.isTaskCanceled(ctx, task.RunID) {
				s.appendTaskLogBestEffort(ctx, task.RunID, "reconcile", "info", "Task canceled because newer deploy superseded current one")
				return true, nil
			}
		}
		lastError := strings.TrimSpace(runErr.Error())
		if len(lastError) > 4000 {
			lastError = lastError[:4000]
		}
		updated, markErr := s.tasks.MarkFailed(ctx, runtimedeploytaskrepo.MarkFailedParams{
			RunID:      task.RunID,
			LeaseOwner: leaseOwner,
			LastError:  lastError,
		})
		if markErr != nil {
			return true, fmt.Errorf("mark runtime deploy task %s as failed: %w", task.RunID, markErr)
		}
		if !updated {
			if s.isTaskCanceled(ctx, task.RunID) {
				s.appendTaskLogBestEffort(ctx, task.RunID, "reconcile", "info", "Task canceled before failed mark commit")
				return true, nil
			}
			return true, fmt.Errorf("mark runtime deploy task %s as failed: lease lost", task.RunID)
		}
		return true, nil
	}

	updated, err := s.tasks.MarkSucceeded(ctx, runtimedeploytaskrepo.MarkSucceededParams{
		RunID:           task.RunID,
		LeaseOwner:      leaseOwner,
		ResultNamespace: result.Namespace,
		ResultTargetEnv: result.TargetEnv,
	})
	if err != nil {
		return true, fmt.Errorf("mark runtime deploy task %s as succeeded: %w", task.RunID, err)
	}
	if !updated {
		if s.isTaskCanceled(ctx, task.RunID) {
			s.appendTaskLogBestEffort(ctx, task.RunID, "reconcile", "info", "Task result ignored because task was canceled")
			return true, nil
		}
		return true, fmt.Errorf("mark runtime deploy task %s as succeeded: lease lost", task.RunID)
	}
	s.appendTaskLogBestEffort(ctx, task.RunID, "reconcile", "info", "Task succeeded for namespace "+result.Namespace+" env "+result.TargetEnv)
	return true, nil
}

func (s *Service) isTaskCanceled(ctx context.Context, runID string) bool {
	task, ok, err := s.tasks.GetByRunID(ctx, runID)
	if err != nil {
		s.logger.Warn("load runtime deploy task status for cancellation check failed", "run_id", runID, "err", err)
		return false
	}
	if !ok {
		return false
	}
	return task.Status == entitytypes.RuntimeDeployTaskStatusCanceled
}

// applyDesiredState builds images and applies infrastructure/services for one runtime target namespace.
func (s *Service) applyDesiredState(ctx context.Context, params PrepareParams) (PrepareResult, error) {
	zero := PrepareResult{}
	runID := strings.TrimSpace(params.RunID)
	if runID == "" {
		return zero, fmt.Errorf("run_id is required")
	}
	s.appendTaskLogBestEffort(ctx, runID, "prepare", "info", "Start runtime deploy applyDesiredState")

	targetEnv := strings.TrimSpace(params.TargetEnv)
	if targetEnv == "" {
		targetEnv = "ai"
	}
	targetNamespace := strings.TrimSpace(params.Namespace)
	templateVars := s.buildTemplateVars(params, targetNamespace)
	repositoryRoot, err := s.resolveRunRepositoryRoot(ctx, params, templateVars, runID)
	if err != nil {
		s.appendTaskLogBestEffort(ctx, runID, "repo-sync", "error", "Resolve repository snapshot failed: "+err.Error())
		return zero, fmt.Errorf("resolve repository snapshot: %w", err)
	}
	servicesConfigPath := s.resolveServicesConfigPath(repositoryRoot, params.ServicesYAMLPath)
	loaded, err := servicescfg.Load(servicesConfigPath, servicescfg.LoadOptions{
		Env:       targetEnv,
		Namespace: targetNamespace,
		Slot:      params.SlotNo,
		Vars:      templateVars,
	})
	if err != nil {
		s.appendTaskLogBestEffort(ctx, runID, "prepare", "error", "Load services config failed: "+err.Error())
		return zero, fmt.Errorf("load services config: %w", err)
	}

	if strings.TrimSpace(targetNamespace) == "" {
		targetNamespace = strings.TrimSpace(loaded.Context.Namespace)
	}
	if targetNamespace == "" {
		return zero, fmt.Errorf("resolved target namespace is empty")
	}
	if effectiveEnv := strings.TrimSpace(loaded.Context.Env); effectiveEnv != "" {
		targetEnv = effectiveEnv
		params.TargetEnv = targetEnv
	}

	// Template vars are used to render Kubernetes manifests. Some variables depend on
	// the final namespace and must be (re)computed after services.yaml resolved it.
	templateVars = s.buildTemplateVars(params, targetNamespace)
	applyStackImageVars(templateVars, loaded.Stack)

	templateVars["CODEXK8S_PRODUCTION_NAMESPACE"] = targetNamespace
	templateVars["CODEXK8S_WORKER_K8S_NAMESPACE"] = targetNamespace
	templateVars["CODEXK8S_REPOSITORY_ROOT"] = repositoryRoot
	if repoName := strings.TrimSpace(params.RepositoryFullName); repoName != "" {
		templateVars["CODEXK8S_GITHUB_REPO"] = repoName
	}
	if strings.TrimSpace(templateVars["CODEXK8S_WORKER_JOB_IMAGE"]) == "" {
		if value := strings.TrimSpace(templateVars["CODEXK8S_AGENT_RUNNER_IMAGE"]); value != "" {
			templateVars["CODEXK8S_WORKER_JOB_IMAGE"] = value
		}
	}

	// Allow services.yaml to override public host resolution (full-env domainTemplate).
	if loaded.Stack != nil {
		if envCfg, err := servicescfg.ResolveEnvironment(loaded.Stack, targetEnv); err == nil {
			if host := strings.TrimSpace(envCfg.DomainTemplate); host != "" {
				templateVars["CODEXK8S_PUBLIC_DOMAIN"] = host
				if strings.EqualFold(targetEnv, "ai") || strings.TrimSpace(templateVars["CODEXK8S_PUBLIC_BASE_URL"]) == "" {
					templateVars["CODEXK8S_PUBLIC_BASE_URL"] = "https://" + host
				}
			}
		}
	}

	if strings.EqualFold(strings.TrimSpace(loaded.Stack.Spec.Project), "codex-k8s") {
		s.appendTaskLogBestEffort(ctx, runID, "prerequisites", "info", "Ensuring codex-k8s prerequisites")
		if err := s.ensureCodexK8sPrerequisites(ctx, repositoryRoot, targetNamespace, templateVars, loaded.Stack, runID); err != nil {
			s.appendTaskLogBestEffort(ctx, runID, "prerequisites", "error", "Ensure prerequisites failed: "+err.Error())
			return zero, fmt.Errorf("ensure codex-k8s prerequisites: %w", err)
		}
	}

	issuerBefore := strings.TrimSpace(templateVars["CODEXK8S_CERT_ISSUER_ENABLED"])
	if err := s.prepareTLS(ctx, repositoryRoot, targetEnv, targetNamespace, templateVars, runID); err != nil {
		s.appendTaskLogBestEffort(ctx, runID, "tls", "error", "Prepare TLS failed: "+err.Error())
		return zero, fmt.Errorf("prepare tls: %w", err)
	}
	if strings.TrimSpace(templateVars["CODEXK8S_CERT_ISSUER_ENABLED"]) != issuerBefore {
		reloaded, err := servicescfg.Load(servicesConfigPath, servicescfg.LoadOptions{
			Env:       targetEnv,
			Namespace: targetNamespace,
			Slot:      params.SlotNo,
			Vars:      templateVars,
		})
		if err != nil {
			s.appendTaskLogBestEffort(ctx, runID, "prepare", "error", "Reload services config after TLS update failed: "+err.Error())
			return zero, fmt.Errorf("reload services config after tls update: %w", err)
		}
		loaded = reloaded
	}

	if _, err := s.applyInfrastructure(ctx, repositoryRoot, loaded.Stack, targetNamespace, templateVars, runID); err != nil {
		s.appendTaskLogBestEffort(ctx, runID, "infrastructure", "error", "Apply infrastructure failed: "+err.Error())
		return zero, fmt.Errorf("apply infrastructure: %w", err)
	}
	if err := s.buildImages(ctx, repositoryRoot, params, loaded.Stack, targetNamespace, templateVars); err != nil {
		s.appendTaskLogBestEffort(ctx, runID, "build", "error", "Build images failed: "+err.Error())
		return zero, fmt.Errorf("build images: %w", err)
	}
	appliedInfra, err := s.applyInfrastructure(ctx, repositoryRoot, loaded.Stack, targetNamespace, templateVars, runID)
	if err != nil {
		s.appendTaskLogBestEffort(ctx, runID, "infrastructure", "error", "Re-apply infrastructure failed: "+err.Error())
		return zero, fmt.Errorf("re-apply infrastructure: %w", err)
	}
	if err := s.applyServices(ctx, repositoryRoot, loaded.Stack, targetNamespace, templateVars, appliedInfra, runID); err != nil {
		s.appendTaskLogBestEffort(ctx, runID, "services", "error", "Apply services failed: "+err.Error())
		return zero, fmt.Errorf("apply services: %w", err)
	}

	if err := s.finalizeTLS(ctx, targetEnv, targetNamespace, templateVars, runID); err != nil {
		s.appendTaskLogBestEffort(ctx, runID, "tls", "error", "Finalize TLS failed: "+err.Error())
		return zero, fmt.Errorf("finalize tls: %w", err)
	}
	s.appendTaskLogBestEffort(ctx, runID, "prepare", "info", "Runtime deploy finished successfully")
	return PrepareResult{
		Namespace: targetNamespace,
		TargetEnv: targetEnv,
	}, nil
}
