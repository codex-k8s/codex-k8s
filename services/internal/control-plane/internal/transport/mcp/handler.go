package mcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/auth"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
)

const tokenInfoSessionKey = "codexk8s_session"

type domainService interface {
	VerifyRunToken(ctx context.Context, rawToken string) (mcpdomain.SessionContext, error)
	AllowedTools(ctx context.Context, session mcpdomain.SessionContext) ([]mcpdomain.ToolCapability, error)
	IsToolAllowed(ctx context.Context, session mcpdomain.SessionContext, toolName mcpdomain.ToolName) (bool, error)
	GitHubLabelsList(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.GitHubLabelsListInput) (mcpdomain.GitHubLabelsListResult, error)
	GitHubLabelsAdd(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.GitHubLabelsAddInput) (mcpdomain.GitHubLabelsMutationResult, error)
	GitHubLabelsRemove(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.GitHubLabelsRemoveInput) (mcpdomain.GitHubLabelsMutationResult, error)
	GitHubLabelsTransition(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.GitHubLabelsTransitionInput) (mcpdomain.GitHubLabelsMutationResult, error)
	RunStatusReport(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.RunStatusReportInput) (mcpdomain.RunStatusReportResult, error)
	MCPSecretSyncEnv(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.SecretSyncEnvInput) (mcpdomain.SecretSyncEnvResult, error)
	MCPDatabaseLifecycle(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.DatabaseLifecycleInput) (mcpdomain.DatabaseLifecycleResult, error)
	MCPOwnerFeedbackRequest(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.OwnerFeedbackRequestInput) (mcpdomain.OwnerFeedbackRequestResult, error)
	SelfImproveRunsList(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.SelfImproveRunsListInput) (mcpdomain.SelfImproveRunsListResult, error)
	SelfImproveRunLookup(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.SelfImproveRunLookupInput) (mcpdomain.SelfImproveRunLookupResult, error)
	SelfImproveSessionGet(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.SelfImproveSessionGetInput) (mcpdomain.SelfImproveSessionGetResult, error)
}

// NewHandler constructs authenticated MCP StreamableHTTP handler.
func NewHandler(service domainService, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	_ = logger
	if service == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "mcp service is not configured", http.StatusServiceUnavailable)
		})
	}

	server := sdkmcp.NewServer(&sdkmcp.Implementation{
		Name:    "codex-k8s-control-plane-mcp",
		Version: "v0.0.1",
	}, nil)

	registerToolAccessMiddleware(server, service)
	registerTools(server, service)

	streamHandler := sdkmcp.NewStreamableHTTPHandler(func(_ *http.Request) *sdkmcp.Server {
		return server
	}, nil)
	authMiddleware := auth.RequireBearerToken(verifyBearerToken(service), nil)
	return authMiddleware(streamHandler)
}

func verifyBearerToken(service domainService) auth.TokenVerifier {
	return func(ctx context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
		session, err := service.VerifyRunToken(ctx, strings.TrimSpace(token))
		if err != nil {
			return nil, errors.Join(auth.ErrInvalidToken, err)
		}
		return &auth.TokenInfo{
			UserID:     session.RunID,
			Expiration: session.ExpiresAt,
			Scopes:     []string{"codexk8s:mcp"},
			Extra: map[string]any{
				tokenInfoSessionKey: session,
			},
		}, nil
	}
}

func sessionFromTokenInfo(extra *sdkmcp.RequestExtra) (mcpdomain.SessionContext, error) {
	if extra == nil || extra.TokenInfo == nil {
		return mcpdomain.SessionContext{}, fmt.Errorf("missing token info in request")
	}
	if extra.TokenInfo.Extra == nil {
		return mcpdomain.SessionContext{}, fmt.Errorf("missing session in token info")
	}

	raw, ok := extra.TokenInfo.Extra[tokenInfoSessionKey]
	if !ok {
		return mcpdomain.SessionContext{}, fmt.Errorf("missing session in token info")
	}
	switch value := raw.(type) {
	case mcpdomain.SessionContext:
		return value, nil
	case *mcpdomain.SessionContext:
		if value == nil {
			return mcpdomain.SessionContext{}, fmt.Errorf("empty session in token info")
		}
		return *value, nil
	default:
		return mcpdomain.SessionContext{}, fmt.Errorf("unexpected session payload type")
	}
}
