package http

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/controlplane"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

const (
	headerMCPCallbackToken = "X-Codex-MCP-Token"
	authPrefixBearer       = "Bearer "
	defaultMCPCallbackUser = "mcp-callback"

	approvalDecisionApproved = "approved"
	approvalDecisionDenied   = "denied"
	approvalDecisionExpired  = "expired"
	approvalDecisionFailed   = "failed"
	approvalDecisionApplied  = "applied"
)

type mcpCallbackHandler struct {
	cp            *controlplane.Client
	callbackToken string
}

func newMCPCallbackHandler(cfg ServerConfig, cp *controlplane.Client) *mcpCallbackHandler {
	return &mcpCallbackHandler{
		cp:            cp,
		callbackToken: strings.TrimSpace(cfg.MCPCallbackToken),
	}
}

func (h *mcpCallbackHandler) CallbackApprover(c *echo.Context) error {
	return h.handleDecisionCallback(c)
}

func (h *mcpCallbackHandler) CallbackExecutor(c *echo.Context) error {
	return h.handleDecisionCallback(c)
}

func (h *mcpCallbackHandler) handleDecisionCallback(c *echo.Context) error {
	if strings.TrimSpace(h.callbackToken) == "" {
		return errs.Unauthorized{Msg: "mcp callback is not configured"}
	}
	if strings.TrimSpace(resolveMCPCallbackToken(c.Request().Header.Get(headerMCPCallbackToken), c.Request().Header.Get(echo.HeaderAuthorization))) != h.callbackToken {
		return errs.Unauthorized{Msg: "invalid mcp callback token"}
	}
	if h.cp == nil || h.cp.Service() == nil {
		return errs.Unauthorized{Msg: "mcp callback service is unavailable"}
	}

	var req models.MCPApprovalCallbackRequest
	if err := bindBody(c, &req); err != nil {
		return err
	}
	if req.ApprovalRequestID <= 0 {
		return errs.Validation{Field: "approval_request_id", Msg: "must be a positive int64"}
	}

	decision := strings.ToLower(strings.TrimSpace(req.Decision))
	if decision == "" {
		return errs.Validation{Field: "decision", Msg: "is required"}
	}
	if !isMCPDecisionAllowed(decision) {
		return errs.Validation{Field: "decision", Msg: "must be one of: approved, denied, expired, failed, applied"}
	}

	actorID := strings.TrimSpace(req.ActorID)
	if actorID == "" {
		actorID = defaultMCPCallbackUser
	}

	principal := &controlplanev1.Principal{
		UserId:          defaultMCPCallbackUser + ":" + actorID,
		Email:           "",
		GithubLogin:     actorID,
		IsPlatformAdmin: true,
		IsPlatformOwner: false,
	}
	result, err := h.cp.Service().ResolveApprovalDecision(c.Request().Context(), &controlplanev1.ResolveApprovalDecisionRequest{
		Principal:         principal,
		ApprovalRequestId: req.ApprovalRequestID,
		Decision:          decision,
		Reason:            optionalStringPtr(req.Reason),
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, casters.ResolveApprovalDecision(result))
}

func resolveMCPCallbackToken(callbackHeader string, authorizationHeader string) string {
	token := strings.TrimSpace(callbackHeader)
	if token != "" {
		return token
	}
	authorization := strings.TrimSpace(authorizationHeader)
	if strings.HasPrefix(strings.ToLower(authorization), strings.ToLower(authPrefixBearer)) {
		return strings.TrimSpace(authorization[len(authPrefixBearer):])
	}
	return ""
}

func isMCPDecisionAllowed(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case approvalDecisionApproved,
		approvalDecisionDenied,
		approvalDecisionExpired,
		approvalDecisionFailed,
		approvalDecisionApplied:
		return true
	default:
		return false
	}
}
