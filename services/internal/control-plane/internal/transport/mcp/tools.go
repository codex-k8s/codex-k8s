package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
)

func registerTools(server *sdkmcp.Server, service domainService) {
	addTool(server, mcpdomain.ToolGitHubLabelsList, "List issue or pull request labels", service.GitHubLabelsList)
	addTool(server, mcpdomain.ToolGitHubLabelsAdd, "Add labels to issue or pull request", service.GitHubLabelsAdd)
	addTool(server, mcpdomain.ToolGitHubLabelsRemove, "Remove labels from issue or pull request", service.GitHubLabelsRemove)
	addTool(server, mcpdomain.ToolGitHubLabelsTransition, "Transition labels (remove + add) on issue or pull request", service.GitHubLabelsTransition)
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
