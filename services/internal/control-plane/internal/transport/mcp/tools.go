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
	addTool(server, mcpdomain.ToolKubernetesPodLogsGet, "Read pod logs in run namespace", service.KubernetesPodLogsGet)
	addTool(server, mcpdomain.ToolKubernetesPodExec, "Exec command in pod", service.KubernetesPodExec)
	addTool(server, mcpdomain.ToolKubernetesManifestApply, "Apply manifest (approval required)", service.KubernetesManifestApply)
	addTool(server, mcpdomain.ToolKubernetesManifestDelete, "Delete manifest (approval required)", service.KubernetesManifestDelete)
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
