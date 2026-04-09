package http

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/kodex/libs/go/errs"
	"github.com/codex-k8s/kodex/services/external/api-gateway/internal/controlplane"
	"github.com/codex-k8s/kodex/services/external/api-gateway/internal/transport/http/casters"
	"github.com/codex-k8s/kodex/services/external/api-gateway/internal/transport/http/models"
)

type interactionCallbackHandler struct {
	cp     *controlplane.Client
	logger *slog.Logger
}

func newInteractionCallbackHandler(cp *controlplane.Client) *interactionCallbackHandler {
	return &interactionCallbackHandler{
		cp:     cp,
		logger: slog.Default(),
	}
}

func (h *interactionCallbackHandler) Callback(c *echo.Context) error {
	startedAt := time.Now().UTC()
	callbackToken := strings.TrimSpace(resolveMCPCallbackToken(
		c.Request().Header.Get(headerMCPCallbackToken),
		c.Request().Header.Get(echo.HeaderAuthorization),
	))
	if callbackToken == "" {
		recordInteractionCallbackMetrics(interactionCallbackMetricUnknown, interactionCallbackMetricError, startedAt)
		return errs.Unauthorized{Msg: "missing mcp callback token"}
	}
	if h.cp == nil {
		recordInteractionCallbackMetrics(interactionCallbackMetricUnknown, interactionCallbackMetricError, startedAt)
		return errs.Unauthorized{Msg: "mcp callback service is unavailable"}
	}

	var req models.InteractionCallbackEnvelope
	if err := bindBody(c, &req); err != nil {
		recordInteractionCallbackMetrics(interactionCallbackMetricUnknown, interactionCallbackMetricError, startedAt)
		return err
	}
	callbackKind := normalizeInteractionCallbackMetricLabel(req.CallbackKind)

	grpcReq, err := casters.InteractionCallbackRequest(req)
	if err != nil {
		recordInteractionCallbackMetrics(callbackKind, interactionCallbackMetricError, startedAt)
		return err
	}

	result, err := h.cp.SubmitInteractionCallback(c.Request().Context(), callbackToken, grpcReq)
	if err != nil {
		recordInteractionCallbackMetrics(callbackKind, interactionCallbackMetricError, startedAt)
		return err
	}

	response := casters.InteractionCallbackOutcome(result)
	classification := normalizeInteractionCallbackMetricLabel(response.Classification)
	recordInteractionCallbackMetrics(callbackKind, classification, startedAt)
	h.logger.Info(
		"interaction callback handled",
		"interaction_id", strings.TrimSpace(req.InteractionID),
		"delivery_id", strings.TrimSpace(req.DeliveryID),
		"adapter_event_id", strings.TrimSpace(req.AdapterEventID),
		"callback_kind", callbackKind,
		"classification", classification,
		"interaction_state", strings.TrimSpace(response.InteractionState),
		"resume_required", response.ResumeRequired,
	)

	return c.JSON(http.StatusOK, response)
}
