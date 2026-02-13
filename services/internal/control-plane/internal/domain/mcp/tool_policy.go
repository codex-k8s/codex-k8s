package mcp

import "sort"

// DefaultToolCatalog returns deterministic MCP tool catalog for prompt/context and policy checks.
func DefaultToolCatalog() []ToolCapability {
	items := []ToolCapability{
		{Name: ToolPromptContextGet, Description: "Get deterministic run prompt context", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolGitHubLabelsList, Description: "List labels on GitHub issue", Category: ToolCategoryRead, Approval: ToolApprovalNone},
		{Name: ToolGitHubLabelsAdd, Description: "Add labels to GitHub issue or pull request", Category: ToolCategoryWrite, Approval: ToolApprovalNone},
		{Name: ToolGitHubLabelsRemove, Description: "Remove labels from GitHub issue or pull request", Category: ToolCategoryWrite, Approval: ToolApprovalNone},
		{Name: ToolGitHubLabelsTransition, Description: "Replace labels atomically on GitHub issue or pull request", Category: ToolCategoryWrite, Approval: ToolApprovalNone},
		{Name: ToolMCPSecretSyncEnv, Description: "Sync one secret value between GitHub repository and Kubernetes namespace", Category: ToolCategoryWrite, Approval: ToolApprovalOwner},
		{Name: ToolMCPDatabaseLifecycle, Description: "Create, drop or describe one environment database", Category: ToolCategoryWrite, Approval: ToolApprovalOwner},
		{Name: ToolMCPOwnerFeedbackRequest, Description: "Request owner feedback with predefined options", Category: ToolCategoryWrite, Approval: ToolApprovalOwner},
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
