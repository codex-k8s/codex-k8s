package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/auth"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
)

const (
	tokenInfoSessionKey   = "codexk8s_session"
	promptContextResource = "codex://prompt/context"
)

type domainService interface {
	VerifyRunToken(ctx context.Context, rawToken string) (mcpdomain.SessionContext, error)
	PromptContext(ctx context.Context, session mcpdomain.SessionContext) (mcpdomain.PromptContextResult, error)
	GitHubIssueGet(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.GitHubIssueGetInput) (mcpdomain.GitHubIssueGetResult, error)
	GitHubPullRequestGet(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.GitHubPullRequestGetInput) (mcpdomain.GitHubPullRequestGetResult, error)
	GitHubIssueCommentsList(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.GitHubIssueCommentsListInput) (mcpdomain.GitHubIssueCommentsListResult, error)
	GitHubLabelsList(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.GitHubLabelsListInput) (mcpdomain.GitHubLabelsListResult, error)
	GitHubBranchesList(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.GitHubBranchesListInput) (mcpdomain.GitHubBranchesListResult, error)
	GitHubBranchEnsure(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.GitHubBranchEnsureInput) (mcpdomain.GitHubBranchEnsureResult, error)
	GitHubPullRequestUpsert(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.GitHubPullRequestUpsertInput) (mcpdomain.GitHubPullRequestUpsertResult, error)
	GitHubIssueCommentCreate(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.GitHubIssueCommentCreateInput) (mcpdomain.GitHubIssueCommentCreateResult, error)
	GitHubLabelsAdd(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.GitHubLabelsAddInput) (mcpdomain.GitHubLabelsMutationResult, error)
	GitHubLabelsRemove(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.GitHubLabelsRemoveInput) (mcpdomain.GitHubLabelsMutationResult, error)
	KubernetesPodsList(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.KubernetesPodsListInput) (mcpdomain.KubernetesPodsListResult, error)
	KubernetesEventsList(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.KubernetesEventsListInput) (mcpdomain.KubernetesEventsListResult, error)
	KubernetesPodLogsGet(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.KubernetesPodLogsGetInput) (mcpdomain.KubernetesPodLogsGetResult, error)
	KubernetesPodExec(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.KubernetesPodExecInput) (mcpdomain.KubernetesPodExecToolResult, error)
	KubernetesManifestApply(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.KubernetesManifestApplyInput) (mcpdomain.ApprovalRequiredResult, error)
	KubernetesManifestDelete(ctx context.Context, session mcpdomain.SessionContext, input mcpdomain.KubernetesManifestDeleteInput) (mcpdomain.ApprovalRequiredResult, error)
}

// NewHandler constructs authenticated MCP StreamableHTTP handler.
func NewHandler(service domainService, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	if service == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "mcp service is not configured", http.StatusServiceUnavailable)
		})
	}

	server := sdkmcp.NewServer(&sdkmcp.Implementation{
		Name:    "codex-k8s-control-plane-mcp",
		Version: "s2-day3.5",
	}, nil)

	registerTools(server, service)
	registerResources(server, service, logger)

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

func registerResources(server *sdkmcp.Server, service domainService, logger *slog.Logger) {
	server.AddResource(&sdkmcp.Resource{
		Name:        "prompt_context",
		Description: "Effective prompt runtime context for authenticated run",
		MIMEType:    "application/json",
		URI:         promptContextResource,
	}, func(ctx context.Context, req *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
		if req == nil || req.Params == nil {
			return nil, fmt.Errorf("read resource request is required")
		}
		if req.Params.URI != promptContextResource {
			return nil, sdkmcp.ResourceNotFoundError(req.Params.URI)
		}

		session, err := sessionFromTokenInfo(req.Extra)
		if err != nil {
			return nil, err
		}
		result, err := service.PromptContext(ctx, session)
		if err != nil {
			return nil, err
		}

		raw, err := json.Marshal(result.Context)
		if err != nil {
			logger.Error("marshal prompt context resource failed", "err", err)
			return nil, fmt.Errorf("marshal prompt context: %w", err)
		}

		return &sdkmcp.ReadResourceResult{
			Contents: []*sdkmcp.ResourceContents{{
				URI:      promptContextResource,
				MIMEType: "application/json",
				Text:     string(raw),
			}},
		}, nil
	})
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
