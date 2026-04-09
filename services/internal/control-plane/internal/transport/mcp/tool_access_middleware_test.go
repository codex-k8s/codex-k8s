package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	mcpdomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/mcp"
)

func TestHandleToolsListAccessFiltersToAllowedCatalog(t *testing.T) {
	t.Parallel()

	req := &sdkmcp.ServerRequest[*sdkmcp.ListToolsParams]{
		Params: &sdkmcp.ListToolsParams{},
		Extra: &sdkmcp.RequestExtra{
			TokenInfo: &auth.TokenInfo{
				Extra: map[string]any{
					tokenInfoSessionKey: mcpdomain.SessionContext{RunID: "run-1"},
				},
			},
		},
	}
	service := toolAccessMiddlewareTestService{
		allowedTools: []mcpdomain.ToolCapability{
			{Name: mcpdomain.ToolGitHubLabelsList},
			{Name: mcpdomain.ToolMCPUserNotify},
			{Name: mcpdomain.ToolMCPUserDecisionRequest},
		},
	}

	result, err := handleToolsListAccess(context.Background(), mcpMethodToolsList, req, service, func(context.Context, string, sdkmcp.Request) (sdkmcp.Result, error) {
		return &sdkmcp.ListToolsResult{
			Tools: []*sdkmcp.Tool{
				{Name: string(mcpdomain.ToolGitHubLabelsList)},
				{Name: string(mcpdomain.ToolMCPUserNotify)},
				{Name: string(mcpdomain.ToolMCPUserDecisionRequest)},
				{Name: "not.allowed"},
			},
		}, nil
	})
	if err != nil {
		t.Fatalf("handleToolsListAccess() error = %v", err)
	}

	toolsResult, ok := result.(*sdkmcp.ListToolsResult)
	if !ok {
		t.Fatalf("result type = %T, want *sdkmcp.ListToolsResult", result)
	}
	if len(toolsResult.Tools) != 3 {
		t.Fatalf("tools/list filtered count = %d, want 3", len(toolsResult.Tools))
	}
	if toolsResult.Tools[1].Name != string(mcpdomain.ToolMCPUserNotify) {
		t.Fatalf("filtered tool[1] = %q, want %q", toolsResult.Tools[1].Name, mcpdomain.ToolMCPUserNotify)
	}
	if toolsResult.Tools[2].Name != string(mcpdomain.ToolMCPUserDecisionRequest) {
		t.Fatalf("filtered tool[2] = %q, want %q", toolsResult.Tools[2].Name, mcpdomain.ToolMCPUserDecisionRequest)
	}
}

type toolAccessMiddlewareTestService struct {
	allowedTools []mcpdomain.ToolCapability
}

func (s toolAccessMiddlewareTestService) VerifyRunToken(context.Context, string) (mcpdomain.SessionContext, error) {
	return mcpdomain.SessionContext{}, nil
}

func (s toolAccessMiddlewareTestService) AllowedTools(context.Context, mcpdomain.SessionContext) ([]mcpdomain.ToolCapability, error) {
	return s.allowedTools, nil
}

func (s toolAccessMiddlewareTestService) IsToolAllowed(_ context.Context, _ mcpdomain.SessionContext, toolName mcpdomain.ToolName) (bool, error) {
	for _, tool := range s.allowedTools {
		if tool.Name == toolName {
			return true, nil
		}
	}
	return false, nil
}

func (toolAccessMiddlewareTestService) GitHubLabelsList(context.Context, mcpdomain.SessionContext, mcpdomain.GitHubLabelsListInput) (mcpdomain.GitHubLabelsListResult, error) {
	return mcpdomain.GitHubLabelsListResult{}, nil
}

func (toolAccessMiddlewareTestService) GitHubLabelsAdd(context.Context, mcpdomain.SessionContext, mcpdomain.GitHubLabelsAddInput) (mcpdomain.GitHubLabelsMutationResult, error) {
	return mcpdomain.GitHubLabelsMutationResult{}, nil
}

func (toolAccessMiddlewareTestService) GitHubLabelsRemove(context.Context, mcpdomain.SessionContext, mcpdomain.GitHubLabelsRemoveInput) (mcpdomain.GitHubLabelsMutationResult, error) {
	return mcpdomain.GitHubLabelsMutationResult{}, nil
}

func (toolAccessMiddlewareTestService) GitHubLabelsTransition(context.Context, mcpdomain.SessionContext, mcpdomain.GitHubLabelsTransitionInput) (mcpdomain.GitHubLabelsMutationResult, error) {
	return mcpdomain.GitHubLabelsMutationResult{}, nil
}

func (toolAccessMiddlewareTestService) RunStatusReport(context.Context, mcpdomain.SessionContext, mcpdomain.RunStatusReportInput) (mcpdomain.RunStatusReportResult, error) {
	return mcpdomain.RunStatusReportResult{}, nil
}

func (toolAccessMiddlewareTestService) MCPSecretSyncEnv(context.Context, mcpdomain.SessionContext, mcpdomain.SecretSyncEnvInput) (mcpdomain.SecretSyncEnvResult, error) {
	return mcpdomain.SecretSyncEnvResult{}, nil
}

func (toolAccessMiddlewareTestService) MCPDatabaseLifecycle(context.Context, mcpdomain.SessionContext, mcpdomain.DatabaseLifecycleInput) (mcpdomain.DatabaseLifecycleResult, error) {
	return mcpdomain.DatabaseLifecycleResult{}, nil
}

func (toolAccessMiddlewareTestService) MCPOwnerFeedbackRequest(context.Context, mcpdomain.SessionContext, mcpdomain.OwnerFeedbackRequestInput) (mcpdomain.OwnerFeedbackRequestResult, error) {
	return mcpdomain.OwnerFeedbackRequestResult{}, nil
}

func (toolAccessMiddlewareTestService) MCPUserNotify(context.Context, mcpdomain.SessionContext, mcpdomain.UserNotifyInput) (mcpdomain.UserNotifyResult, error) {
	return mcpdomain.UserNotifyResult{}, nil
}

func (toolAccessMiddlewareTestService) MCPUserDecisionRequest(context.Context, mcpdomain.SessionContext, mcpdomain.UserDecisionRequestInput) (mcpdomain.UserDecisionRequestResult, error) {
	return mcpdomain.UserDecisionRequestResult{}, nil
}

func (toolAccessMiddlewareTestService) SelfImproveRunsList(context.Context, mcpdomain.SessionContext, mcpdomain.SelfImproveRunsListInput) (mcpdomain.SelfImproveRunsListResult, error) {
	return mcpdomain.SelfImproveRunsListResult{}, nil
}

func (toolAccessMiddlewareTestService) SelfImproveRunLookup(context.Context, mcpdomain.SessionContext, mcpdomain.SelfImproveRunLookupInput) (mcpdomain.SelfImproveRunLookupResult, error) {
	return mcpdomain.SelfImproveRunLookupResult{}, nil
}

func (toolAccessMiddlewareTestService) SelfImproveSessionGet(context.Context, mcpdomain.SessionContext, mcpdomain.SelfImproveSessionGetInput) (mcpdomain.SelfImproveSessionGetResult, error) {
	return mcpdomain.SelfImproveSessionGetResult{}, nil
}

func TestSessionFromTokenInfoRejectsMissingSession(t *testing.T) {
	t.Parallel()

	_, err := sessionFromTokenInfo(&sdkmcp.RequestExtra{TokenInfo: &auth.TokenInfo{
		Expiration: time.Now(),
	}})
	if err == nil {
		t.Fatal("expected error for missing session")
	}
}
