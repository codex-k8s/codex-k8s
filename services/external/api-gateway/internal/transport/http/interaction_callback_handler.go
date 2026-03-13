package http

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/controlplane"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

type interactionCallbackHandler struct {
	cp *controlplane.Client
}

func newInteractionCallbackHandler(cp *controlplane.Client) *interactionCallbackHandler {
	return &interactionCallbackHandler{cp: cp}
}

func (h *interactionCallbackHandler) Callback(c *echo.Context) error {
	callbackToken := strings.TrimSpace(resolveMCPCallbackToken(
		c.Request().Header.Get(headerMCPCallbackToken),
		c.Request().Header.Get(echo.HeaderAuthorization),
	))
	if callbackToken == "" {
		return errs.Unauthorized{Msg: "missing mcp callback token"}
	}
	if h.cp == nil {
		return errs.Unauthorized{Msg: "mcp callback service is unavailable"}
	}

	var req models.InteractionCallbackEnvelope
	if err := bindBody(c, &req); err != nil {
		return err
	}

	grpcReq, err := casters.InteractionCallbackRequest(req)
	if err != nil {
		return err
	}

	result, err := h.cp.SubmitInteractionCallback(c.Request().Context(), callbackToken, grpcReq)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, casters.InteractionCallbackOutcome(result))
}
