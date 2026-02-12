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
		{Name: ToolKubernetesPodLogsGet, Description: "Read pod logs in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolKubernetesPodExec, Description: "Exec command inside pod in run namespace", Category: ToolCategoryRead, Approval: ToolApprovalNone},
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
