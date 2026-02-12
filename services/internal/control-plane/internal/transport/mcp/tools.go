package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
)

func registerTools(server *sdkmcp.Server, service domainService) {
	addTool(server, mcpdomain.ToolPromptContextGet, "Build effective prompt runtime context", func(ctx context.Context, session mcpdomain.SessionContext, _ struct{}) (mcpdomain.PromptContextResult, error) {
		return service.PromptContext(ctx, session)
	})
	addTool(server, mcpdomain.ToolGitHubIssueGet, "Read GitHub issue details", service.GitHubIssueGet)
	addTool(server, mcpdomain.ToolGitHubPullRequestGet, "Read GitHub pull request details", service.GitHubPullRequestGet)
	addTool(server, mcpdomain.ToolGitHubIssueComments, "List issue comments", service.GitHubIssueCommentsList)
	addTool(server, mcpdomain.ToolGitHubLabelsList, "List issue labels", service.GitHubLabelsList)
	addTool(server, mcpdomain.ToolGitHubBranchesList, "List repository branches", service.GitHubBranchesList)
	addTool(server, mcpdomain.ToolGitHubBranchEnsure, "Create or sync branch", service.GitHubBranchEnsure)
	addTool(server, mcpdomain.ToolGitHubPullRequestUpsert, "Create or update pull request", service.GitHubPullRequestUpsert)
	addTool(server, mcpdomain.ToolGitHubIssueCommentCreate, "Create issue or pull request comment", service.GitHubIssueCommentCreate)
	addTool(server, mcpdomain.ToolGitHubLabelsAdd, "Add labels to issue", service.GitHubLabelsAdd)
	addTool(server, mcpdomain.ToolGitHubLabelsRemove, "Remove labels from issue", service.GitHubLabelsRemove)

	addTool(server, mcpdomain.ToolKubernetesPodsList, "List pods in run namespace", service.KubernetesPodsList)
	addTool(server, mcpdomain.ToolKubernetesEventsList, "List events in run namespace", service.KubernetesEventsList)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesDeploymentsList, "List deployments in run namespace", mcpdomain.KubernetesResourceKindDeployment)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesDaemonSetsList, "List daemonsets in run namespace", mcpdomain.KubernetesResourceKindDaemonSet)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesStatefulSetsList, "List statefulsets in run namespace", mcpdomain.KubernetesResourceKindStatefulSet)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesReplicaSetsList, "List replicasets in run namespace", mcpdomain.KubernetesResourceKindReplicaSet)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesReplicationControllersList, "List replicationcontrollers in run namespace", mcpdomain.KubernetesResourceKindReplicationController)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesJobsList, "List jobs in run namespace", mcpdomain.KubernetesResourceKindJob)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesCronJobsList, "List cronjobs in run namespace", mcpdomain.KubernetesResourceKindCronJob)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesConfigMapsList, "List configmaps in run namespace", mcpdomain.KubernetesResourceKindConfigMap)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesSecretsList, "List secrets in run namespace", mcpdomain.KubernetesResourceKindSecret)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesResourceQuotasList, "List resourcequotas in run namespace", mcpdomain.KubernetesResourceKindResourceQuota)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesHorizontalPodAutoscalersList, "List HPAs in run namespace", mcpdomain.KubernetesResourceKindHPA)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesServicesList, "List services in run namespace", mcpdomain.KubernetesResourceKindService)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesEndpointsList, "List endpoints in run namespace", mcpdomain.KubernetesResourceKindEndpoints)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesIngressesList, "List ingresses in run namespace", mcpdomain.KubernetesResourceKindIngress)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesIngressClassesList, "List ingress classes", mcpdomain.KubernetesResourceKindIngressClass)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesNetworkPoliciesList, "List networkpolicies in run namespace", mcpdomain.KubernetesResourceKindNetworkPolicy)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesPersistentVolumeClaimsList, "List PVCs in run namespace", mcpdomain.KubernetesResourceKindPVC)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesPersistentVolumesList, "List persistent volumes", mcpdomain.KubernetesResourceKindPV)
	registerKubernetesResourceListTool(server, service, mcpdomain.ToolKubernetesStorageClassesList, "List storage classes", mcpdomain.KubernetesResourceKindStorageClass)
	addTool(server, mcpdomain.ToolKubernetesPodLogsGet, "Read pod logs in run namespace", service.KubernetesPodLogsGet)
	addTool(server, mcpdomain.ToolKubernetesPodExec, "Exec command in pod", service.KubernetesPodExec)
	addTool(server, mcpdomain.ToolKubernetesPodPortForward, "Port-forward pod (approval required)", service.KubernetesPodPortForward)
	addTool(server, mcpdomain.ToolKubernetesManifestApply, "Apply manifest (approval required)", service.KubernetesManifestApply)
	addTool(server, mcpdomain.ToolKubernetesManifestDelete, "Delete manifest (approval required)", service.KubernetesManifestDelete)
}

func registerKubernetesResourceListTool(
	server *sdkmcp.Server,
	service domainService,
	tool mcpdomain.ToolName,
	description string,
	kind mcpdomain.KubernetesResourceKind,
) {
	addTool(server, tool, description, func(
		ctx context.Context,
		session mcpdomain.SessionContext,
		input mcpdomain.KubernetesResourceListInput,
	) (mcpdomain.KubernetesResourceListResult, error) {
		input.Kind = kind
		return service.KubernetesResourcesList(ctx, session, input)
	})
}

func addTool[In any, Out any](server *sdkmcp.Server, name mcpdomain.ToolName, description string, run func(context.Context, mcpdomain.SessionContext, In) (Out, error)) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        string(name),
		Description: description,
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, input In) (*sdkmcp.CallToolResult, Out, error) {
		var zero Out

		session, err := sessionFromTokenInfo(req.Extra)
		if err != nil {
			return nil, zero, err
		}
		output, err := run(ctx, session, input)
		if err != nil {
			return nil, zero, err
		}
		return nil, output, nil
	})
}
