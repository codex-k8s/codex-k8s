package mcp

import (
	"context"
	"fmt"
	"strings"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
)

func (s *Service) KubernetesPodsList(ctx context.Context, session SessionContext, input KubernetesPodsListInput) (KubernetesPodsListResult, error) {
	return newKubernetesPodsListResult(
		kubernetesList(ctx, s, session, ToolKubernetesPodsList, input.Limit, "kubernetes pods list", s.kubernetes.ListPods),
	)
}

func (s *Service) KubernetesEventsList(ctx context.Context, session SessionContext, input KubernetesEventsListInput) (KubernetesEventsListResult, error) {
	events, err := kubernetesList(ctx, s, session, ToolKubernetesEventsList, input.Limit, "kubernetes events list", s.kubernetes.ListEvents)
	if err != nil {
		return KubernetesEventsListResult{}, err
	}

	return KubernetesEventsListResult{Status: ToolExecutionStatusOK, Events: events}, nil
}

func (s *Service) KubernetesResourcesList(ctx context.Context, session SessionContext, input KubernetesResourceListInput) (KubernetesResourceListResult, error) {
	kind := KubernetesResourceKind(strings.TrimSpace(string(input.Kind)))
	if kind == "" {
		return KubernetesResourceListResult{}, fmt.Errorf("resource kind is required")
	}

	items, err := kubernetesResourcesList(ctx, s, session, input.Kind, input.Limit, "kubernetes resources list")
	if err != nil {
		return KubernetesResourceListResult{}, err
	}
	return KubernetesResourceListResult{
		Status: ToolExecutionStatusOK,
		Items:  items,
	}, nil
}

func newKubernetesPodsListResult(pods []KubernetesPod, err error) (KubernetesPodsListResult, error) {
	if err != nil {
		return KubernetesPodsListResult{}, err
	}

	return KubernetesPodsListResult{Status: ToolExecutionStatusOK, Pods: pods}, nil
}

func (s *Service) KubernetesPodLogsGet(ctx context.Context, session SessionContext, input KubernetesPodLogsGetInput) (KubernetesPodLogsGetResult, error) {
	tool, err := s.toolCapability(ToolKubernetesPodLogsGet)
	if err != nil {
		return KubernetesPodLogsGetResult{}, err
	}
	if err := requireRuntimeNamespace(session); err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return KubernetesPodLogsGetResult{}, err
	}
	podName := strings.TrimSpace(input.Pod)
	if podName == "" {
		err := fmt.Errorf("pod is required")
		s.auditToolFailed(ctx, session, tool, err)
		return KubernetesPodLogsGetResult{}, err
	}

	s.auditToolCalled(ctx, session, tool)
	logs, err := s.kubernetes.GetPodLogs(ctx, session.Namespace, podName, strings.TrimSpace(input.Container), clampTail(input.TailLines))
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return KubernetesPodLogsGetResult{}, fmt.Errorf("kubernetes pod logs get: %w", err)
	}
	s.auditToolSucceeded(ctx, session, tool)
	return KubernetesPodLogsGetResult{
		Status: ToolExecutionStatusOK,
		Logs:   logs,
	}, nil
}

func (s *Service) KubernetesPodExec(ctx context.Context, session SessionContext, input KubernetesPodExecInput) (KubernetesPodExecToolResult, error) {
	tool, err := s.toolCapability(ToolKubernetesPodExec)
	if err != nil {
		return KubernetesPodExecToolResult{}, err
	}
	if err := requireRuntimeNamespace(session); err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return KubernetesPodExecToolResult{}, err
	}
	podName := strings.TrimSpace(input.Pod)
	if podName == "" {
		err := fmt.Errorf("pod is required")
		s.auditToolFailed(ctx, session, tool, err)
		return KubernetesPodExecToolResult{}, err
	}
	if len(input.Command) == 0 {
		err := fmt.Errorf("command is required")
		s.auditToolFailed(ctx, session, tool, err)
		return KubernetesPodExecToolResult{}, err
	}

	s.auditToolCalled(ctx, session, tool)
	execResult, err := s.kubernetes.ExecPod(ctx, session.Namespace, podName, strings.TrimSpace(input.Container), input.Command)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return KubernetesPodExecToolResult{}, fmt.Errorf("kubernetes pod exec: %w", err)
	}
	s.auditToolSucceeded(ctx, session, tool)
	return KubernetesPodExecToolResult{
		Status: ToolExecutionStatusOK,
		Exec:   execResult,
	}, nil
}

func (s *Service) KubernetesPodPortForward(ctx context.Context, session SessionContext, _ KubernetesPodPortForwardInput) (KubernetesPodPortForwardResult, error) {
	result, err := s.kubernetesApprovalOnly(ctx, session, ToolKubernetesPodPortForward, "approval is required by policy before pod port-forward")
	if err != nil {
		return KubernetesPodPortForwardResult{}, err
	}
	return KubernetesPodPortForwardResult{
		Status:  result.Status,
		Message: result.Message,
	}, nil
}

func (s *Service) KubernetesManifestApply(ctx context.Context, session SessionContext, _ KubernetesManifestApplyInput) (ApprovalRequiredResult, error) {
	return s.kubernetesApprovalOnly(ctx, session, ToolKubernetesManifestApply, "approval is required by policy before manifest apply")
}

func (s *Service) KubernetesManifestDelete(ctx context.Context, session SessionContext, _ KubernetesManifestDeleteInput) (ApprovalRequiredResult, error) {
	return s.kubernetesApprovalOnly(ctx, session, ToolKubernetesManifestDelete, "approval is required by policy before manifest delete")
}

func kubernetesList[T any](
	ctx context.Context,
	svc *Service,
	session SessionContext,
	toolName ToolName,
	limit int,
	errorPrefix string,
	listFn func(context.Context, string, int) ([]T, error),
) ([]T, error) {
	tool, err := svc.toolCapability(toolName)
	if err != nil {
		return nil, err
	}
	if err := requireRuntimeNamespace(session); err != nil {
		svc.auditToolFailed(ctx, session, tool, err)
		return nil, err
	}

	svc.auditToolCalled(ctx, session, tool)
	items, err := listFn(ctx, session.Namespace, clampLimit(limit, defaultK8sLimit, maxK8sLimit))
	if err != nil {
		svc.auditToolFailed(ctx, session, tool, err)
		return nil, fmt.Errorf("%s: %w", errorPrefix, err)
	}
	svc.auditToolSucceeded(ctx, session, tool)
	return items, nil
}

func kubernetesResourcesList(
	ctx context.Context,
	svc *Service,
	session SessionContext,
	kind KubernetesResourceKind,
	limit int,
	errorPrefix string,
) ([]KubernetesResourceRef, error) {
	toolName := toolByResourceKind(kind)
	if toolName == "" {
		return nil, fmt.Errorf("unsupported resource kind %q", kind)
	}
	tool, err := svc.toolCapability(toolName)
	if err != nil {
		return nil, err
	}
	if err := requireRuntimeNamespace(session); err != nil {
		svc.auditToolFailed(ctx, session, tool, err)
		return nil, err
	}

	svc.auditToolCalled(ctx, session, tool)
	items, err := svc.kubernetes.ListResources(ctx, session.Namespace, kind, clampLimit(limit, defaultK8sLimit, maxK8sLimit))
	if err != nil {
		svc.auditToolFailed(ctx, session, tool, err)
		return nil, fmt.Errorf("%s: %w", errorPrefix, err)
	}
	svc.auditToolSucceeded(ctx, session, tool)
	return items, nil
}

func toolByResourceKind(kind KubernetesResourceKind) ToolName {
	switch kind {
	case KubernetesResourceKindDeployment:
		return ToolKubernetesDeploymentsList
	case KubernetesResourceKindDaemonSet:
		return ToolKubernetesDaemonSetsList
	case KubernetesResourceKindStatefulSet:
		return ToolKubernetesStatefulSetsList
	case KubernetesResourceKindReplicaSet:
		return ToolKubernetesReplicaSetsList
	case KubernetesResourceKindReplicationController:
		return ToolKubernetesReplicationControllersList
	case KubernetesResourceKindJob:
		return ToolKubernetesJobsList
	case KubernetesResourceKindCronJob:
		return ToolKubernetesCronJobsList
	case KubernetesResourceKindConfigMap:
		return ToolKubernetesConfigMapsList
	case KubernetesResourceKindSecret:
		return ToolKubernetesSecretsList
	case KubernetesResourceKindResourceQuota:
		return ToolKubernetesResourceQuotasList
	case KubernetesResourceKindHPA:
		return ToolKubernetesHorizontalPodAutoscalersList
	case KubernetesResourceKindService:
		return ToolKubernetesServicesList
	case KubernetesResourceKindEndpoints:
		return ToolKubernetesEndpointsList
	case KubernetesResourceKindIngress:
		return ToolKubernetesIngressesList
	case KubernetesResourceKindIngressClass:
		return ToolKubernetesIngressClassesList
	case KubernetesResourceKindNetworkPolicy:
		return ToolKubernetesNetworkPoliciesList
	case KubernetesResourceKindPVC:
		return ToolKubernetesPersistentVolumeClaimsList
	case KubernetesResourceKindPV:
		return ToolKubernetesPersistentVolumesList
	case KubernetesResourceKindStorageClass:
		return ToolKubernetesStorageClassesList
	default:
		return ""
	}
}

func (s *Service) kubernetesApprovalOnly(ctx context.Context, session SessionContext, toolName ToolName, message string) (ApprovalRequiredResult, error) {
	tool, err := s.toolCapability(toolName)
	if err != nil {
		return ApprovalRequiredResult{}, err
	}
	if err := requireRuntimeNamespace(session); err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return ApprovalRequiredResult{}, err
	}

	s.auditToolCalled(ctx, session, tool)
	s.auditToolApprovalPending(ctx, session, tool, message)
	return ApprovalRequiredResult{
		Status:  ToolExecutionStatusApprovalRequired,
		Tool:    tool.Name,
		Message: message,
	}, nil
}

func requireRuntimeNamespace(session SessionContext) error {
	if session.RuntimeMode != agentdomain.RuntimeModeFullEnv {
		return fmt.Errorf("runtime_mode %q does not allow Kubernetes tools", session.RuntimeMode)
	}
	if strings.TrimSpace(session.Namespace) == "" {
		return fmt.Errorf("namespace is required for Kubernetes tools")
	}
	return nil
}

func clampTail(value int64) int64 {
	if value <= 0 {
		return defaultTailLines
	}
	if value > maxTailLines {
		return maxTailLines
	}
	return value
}
