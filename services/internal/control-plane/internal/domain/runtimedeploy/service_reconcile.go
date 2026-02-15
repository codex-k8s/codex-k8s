package runtimedeploy

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
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

	task, ok, err := s.tasks.ClaimNext(ctx, runtimedeploytaskrepo.ClaimParams{
		LeaseOwner: leaseOwner,
		LeaseTTL:   fmt.Sprintf("%d seconds", int64(leaseTTL.Seconds())),
	})
	if err != nil {
		return false, fmt.Errorf("claim runtime deploy task: %w", err)
	}
	if !ok {
		return false, nil
	}

	result, runErr := s.applyDesiredState(ctx, PrepareParams{
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
	if runErr != nil {
		if errors.Is(runErr, context.Canceled) || errors.Is(runErr, context.DeadlineExceeded) {
			if ctx.Err() != nil {
				return true, runErr
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
		return true, fmt.Errorf("mark runtime deploy task %s as succeeded: lease lost", task.RunID)
	}
	return true, nil
}

// applyDesiredState builds images and applies infrastructure/services for one runtime target namespace.
func (s *Service) applyDesiredState(ctx context.Context, params PrepareParams) (PrepareResult, error) {
	zero := PrepareResult{}
	runID := strings.TrimSpace(params.RunID)
	if runID == "" {
		return zero, fmt.Errorf("run_id is required")
	}

	targetEnv := strings.TrimSpace(params.TargetEnv)
	if targetEnv == "" {
		targetEnv = "ai"
	}
	targetNamespace := strings.TrimSpace(params.Namespace)
	templateVars := s.buildTemplateVars(params, targetNamespace)
	servicesConfigPath := s.resolveServicesConfigPath(params.ServicesYAMLPath)
	loaded, err := servicescfg.Load(servicesConfigPath, servicescfg.LoadOptions{
		Env:       targetEnv,
		Namespace: targetNamespace,
		Slot:      params.SlotNo,
		Vars:      templateVars,
	})
	if err != nil {
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

	templateVars["CODEXK8S_STAGING_NAMESPACE"] = targetNamespace
	templateVars["CODEXK8S_WORKER_K8S_NAMESPACE"] = targetNamespace
	templateVars["CODEXK8S_GITHUB_REPO"] = strings.TrimSpace(params.RepositoryFullName)
	if strings.TrimSpace(templateVars["CODEXK8S_WORKER_JOB_IMAGE"]) == "" {
		if value := strings.TrimSpace(templateVars["CODEXK8S_AGENT_RUNNER_IMAGE"]); value != "" {
			templateVars["CODEXK8S_WORKER_JOB_IMAGE"] = value
		}
	}

	if strings.EqualFold(strings.TrimSpace(loaded.Stack.Spec.Project), "codex-k8s") {
		if err := s.ensureCodexK8sPrerequisites(ctx, targetNamespace, templateVars); err != nil {
			return zero, fmt.Errorf("ensure codex-k8s prerequisites: %w", err)
		}
	}

	if err := s.buildImages(ctx, params, loaded.Stack, targetNamespace, templateVars); err != nil {
		return zero, fmt.Errorf("build images: %w", err)
	}

	appliedInfra, err := s.applyInfrastructure(ctx, loaded.Stack, targetNamespace, templateVars)
	if err != nil {
		return zero, fmt.Errorf("apply infrastructure: %w", err)
	}
	if err := s.applyServices(ctx, loaded.Stack, targetNamespace, templateVars, appliedInfra); err != nil {
		return zero, fmt.Errorf("apply services: %w", err)
	}
	return PrepareResult{
		Namespace: targetNamespace,
		TargetEnv: targetEnv,
	}, nil
}
