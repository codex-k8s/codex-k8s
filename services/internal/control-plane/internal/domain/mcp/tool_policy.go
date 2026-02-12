package mcp

import "sort"

// DefaultToolCatalog returns deterministic MCP tool catalog for prompt/context and policy checks.
func DefaultToolCatalog() []ToolCapability {
	items := []ToolCapability{
		{Name: ToolPromptContextGet, Description: "Build effective prompt runtime context", Category: ToolCategoryRead, Approval: ToolApprovalNone},

		{Name: ToolGitHubIssueGet, Description: "Read GitHub issue details", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolGitHubPullRequestGet, Description: "Read GitHub pull request details", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolGitHubIssueComments, Description: "List GitHub issue comments", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolGitHubLabelsList, Description: "List labels on GitHub issue", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolGitHubBranchesList, Description: "List repository branches", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolGitHubBranchEnsure, Description: "Create or fast-forward GitHub branch", Category: ToolCategoryWrite, Approval: ToolApprovalNone},
		{Name: ToolGitHubPullRequestUpsert, Description: "Create or update pull request", Category: ToolCategoryWrite, Approval: ToolApprovalNone},
		{Name: ToolGitHubIssueCommentCreate, Description: "Create GitHub issue/PR comment", Category: ToolCategoryWrite, Approval: ToolApprovalNone},
		{Name: ToolGitHubLabelsAdd, Description: "Add labels to GitHub issue", Category: ToolCategoryWrite, Approval: ToolApprovalRequired},
		{Name: ToolGitHubLabelsRemove, Description: "Remove labels from GitHub issue", Category: ToolCategoryWrite, Approval: ToolApprovalRequired},

		{Name: ToolKubernetesPodsList, Description: "List Kubernetes pods in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesEventsList, Description: "List Kubernetes events in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesDeploymentsList, Description: "List Kubernetes deployments in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesDaemonSetsList, Description: "List Kubernetes daemonsets in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesStatefulSetsList, Description: "List Kubernetes statefulsets in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesReplicaSetsList, Description: "List Kubernetes replicasets in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesReplicationControllersList, Description: "List Kubernetes replicationcontrollers in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesJobsList, Description: "List Kubernetes jobs in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesCronJobsList, Description: "List Kubernetes cronjobs in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesConfigMapsList, Description: "List Kubernetes configmaps in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesSecretsList, Description: "List Kubernetes secrets in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesResourceQuotasList, Description: "List Kubernetes resourcequotas in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesHorizontalPodAutoscalersList, Description: "List Kubernetes HPAs in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesServicesList, Description: "List Kubernetes services in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesEndpointsList, Description: "List Kubernetes endpoints in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesIngressesList, Description: "List Kubernetes ingresses in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesIngressClassesList, Description: "List Kubernetes ingressclasses", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesNetworkPoliciesList, Description: "List Kubernetes networkpolicies in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesPersistentVolumeClaimsList, Description: "List Kubernetes PVCs in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesPersistentVolumesList, Description: "List Kubernetes PVs", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesStorageClassesList, Description: "List Kubernetes storageclasses", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesPodLogsGet, Description: "Read pod logs in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesPodExec, Description: "Exec command inside pod in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesPodPortForward, Description: "Port-forward pod in run namespace", Category: ToolCategoryWrite, Approval: ToolApprovalRequired},
		{Name: ToolKubernetesManifestApply, Description: "Apply Kubernetes manifest in run namespace", Category: ToolCategoryWrite, Approval: ToolApprovalRequired},
		{Name: ToolKubernetesManifestDelete, Description: "Delete Kubernetes object in run namespace", Category: ToolCategoryWrite, Approval: ToolApprovalRequired},
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return items
}

func toolCapabilityByName(catalog []ToolCapability, name ToolName) (ToolCapability, bool) {
	for _, tool := range catalog {
		if tool.Name == name {
			return tool, true
		}
	}
	return ToolCapability{}, false
}
