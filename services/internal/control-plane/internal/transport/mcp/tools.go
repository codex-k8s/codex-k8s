package mcp

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
)

func registerTools(server *sdkmcp.Server, service domainService) {
	addTool(server, mcpdomain.ToolGitHubLabelsList, "List issue or pull request labels", service.GitHubLabelsList)
	addTool(server, mcpdomain.ToolGitHubLabelsAdd, "Add labels to issue or pull request", service.GitHubLabelsAdd)
	addTool(server, mcpdomain.ToolGitHubLabelsRemove, "Remove labels from issue or pull request", service.GitHubLabelsRemove)
	addTool(server, mcpdomain.ToolGitHubLabelsTransition, "Transition labels (remove + add) on issue or pull request", service.GitHubLabelsTransition)
	addToolWithInputSchema(
		server,
		mcpdomain.ToolRunStatusReport,
		"Report current run status in user locale (status is required and limited to 100 characters)",
		buildRunStatusReportInputSchema(),
		service.RunStatusReport,
	)
	addTool(server, mcpdomain.ToolMCPSecretSyncEnv, "Sync one secret to GitHub repository and Kubernetes namespace", service.MCPSecretSyncEnv)
	addTool(server, mcpdomain.ToolMCPDatabaseLifecycle, "Create, drop or describe one environment database", service.MCPDatabaseLifecycle)
	addTool(server, mcpdomain.ToolMCPOwnerFeedbackRequest, "Request owner feedback with predefined options", service.MCPOwnerFeedbackRequest)
	addTool(server, mcpdomain.ToolSelfImproveRunsList, "List project runs for self-improve diagnostics", service.SelfImproveRunsList)
	addTool(server, mcpdomain.ToolSelfImproveRunLookup, "Find project runs by issue/pr references for self-improve diagnostics", service.SelfImproveRunLookup)
	addTool(server, mcpdomain.ToolSelfImproveSessionGet, "Get codex-cli session JSON for one run with /tmp path metadata", service.SelfImproveSessionGet)
}

func addTool[In any, Out any](server *sdkmcp.Server, name mcpdomain.ToolName, description string, run func(context.Context, mcpdomain.SessionContext, In) (Out, error)) {
	addToolWithInputSchema(server, name, description, nil, run)
}

func addToolWithInputSchema[In any, Out any](
	server *sdkmcp.Server,
	name mcpdomain.ToolName,
	description string,
	inputSchema *jsonschema.Schema,
	run func(context.Context, mcpdomain.SessionContext, In) (Out, error),
) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        string(name),
		Description: description,
		InputSchema: inputSchema,
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

func buildRunStatusReportInputSchema() *jsonschema.Schema {
	schema, err := jsonschema.For[mcpdomain.RunStatusReportInput](nil)
	if err != nil || schema == nil {
		return nil
	}

	if schema.Properties == nil {
		schema.Properties = make(map[string]*jsonschema.Schema, 1)
	}
	statusSchema, ok := schema.Properties["status"]
	if !ok || statusSchema == nil {
		statusSchema = &jsonschema.Schema{Type: "string"}
		schema.Properties["status"] = statusSchema
	}

	statusSchema.Description = "Short status text about what the agent is doing now (1..100 characters)."
	statusSchema.MinLength = jsonschema.Ptr(1)
	statusSchema.MaxLength = jsonschema.Ptr(100)

	return schema
}
