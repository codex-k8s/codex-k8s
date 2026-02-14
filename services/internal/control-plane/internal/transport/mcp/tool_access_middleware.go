package mcp

import (
	"context"
	"fmt"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
)

const (
	mcpMethodToolsList = "tools/list"
	mcpMethodToolsCall = "tools/call"
)

func registerToolAccessMiddleware(server *sdkmcp.Server, service domainService) {
	server.AddReceivingMiddleware(func(next sdkmcp.MethodHandler) sdkmcp.MethodHandler {
		return func(ctx context.Context, method string, req sdkmcp.Request) (sdkmcp.Result, error) {
			switch method {
			case mcpMethodToolsList:
				return handleToolsListAccess(ctx, method, req, service, next)
			case mcpMethodToolsCall:
				return handleToolCallAccess(ctx, method, req, service, next)
			default:
				return next(ctx, method, req)
			}
		}
	})
}

func handleToolsListAccess(ctx context.Context, method string, req sdkmcp.Request, service domainService, next sdkmcp.MethodHandler) (sdkmcp.Result, error) {
	session, err := sessionFromRequest(req)
	if err != nil {
		return nil, err
	}
	allowedTools, err := service.AllowedTools(ctx, session)
	if err != nil {
		return nil, err
	}
	allowedToolNames := allowedToolNameSet(allowedTools)

	result, err := next(ctx, method, req)
	if err != nil {
		return nil, err
	}

	toolsResult, ok := result.(*sdkmcp.ListToolsResult)
	if !ok {
		return result, nil
	}

	filteredTools := make([]*sdkmcp.Tool, 0, len(toolsResult.Tools))
	for _, tool := range toolsResult.Tools {
		if tool == nil {
			continue
		}
		if _, ok := allowedToolNames[mcpdomain.ToolName(tool.Name)]; ok {
			filteredTools = append(filteredTools, tool)
		}
	}

	return &sdkmcp.ListToolsResult{
		Meta:       toolsResult.Meta,
		NextCursor: toolsResult.NextCursor,
		Tools:      filteredTools,
	}, nil
}

func handleToolCallAccess(ctx context.Context, method string, req sdkmcp.Request, service domainService, next sdkmcp.MethodHandler) (sdkmcp.Result, error) {
	session, err := sessionFromRequest(req)
	if err != nil {
		return nil, err
	}

	toolName, err := toolNameFromRequest(req)
	if err != nil {
		return nil, err
	}
	allowed, err := service.IsToolAllowed(ctx, session, toolName)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("tool %q is not available for current run profile", toolName)
	}

	return next(ctx, method, req)
}

func sessionFromRequest(req sdkmcp.Request) (mcpdomain.SessionContext, error) {
	return sessionFromTokenInfo(req.GetExtra())
}

func toolNameFromRequest(req sdkmcp.Request) (mcpdomain.ToolName, error) {
	params, ok := req.GetParams().(*sdkmcp.CallToolParamsRaw)
	if !ok || params == nil {
		return "", fmt.Errorf("invalid call_tool request params")
	}
	toolName := mcpdomain.ToolName(strings.TrimSpace(params.Name))
	if toolName == "" {
		return "", fmt.Errorf("tool name is required")
	}
	return toolName, nil
}

func allowedToolNameSet(tools []mcpdomain.ToolCapability) map[mcpdomain.ToolName]struct{} {
	out := make(map[mcpdomain.ToolName]struct{}, len(tools))
	for _, tool := range tools {
		out[tool.Name] = struct{}{}
	}
	return out
}
