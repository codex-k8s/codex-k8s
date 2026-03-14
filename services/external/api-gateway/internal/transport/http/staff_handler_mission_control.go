package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/generated"
)

func (h *staffHandler) GetMissionControlDashboard(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolveMissionControlDashboardArg, func(principal *controlplanev1.Principal, arg missionControlDashboardArg) error {
		snapshot, resumeToken, err := h.fetchMissionControlDashboardSnapshot(c.Request().Context(), principal, arg)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.MissionControlDashboardSnapshot(snapshot, resumeToken))
	})
}

func (h *staffHandler) GetMissionControlEntity(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolveMissionControlEntityArg, func(principal *controlplanev1.Principal, arg missionControlEntityArg) error {
		item, err := h.getMissionControlEntityCall(c.Request().Context(), principal, arg)
		if err != nil {
			return err
		}
		response, err := casters.MissionControlEntityDetails(item)
		if err != nil {
			return fmt.Errorf("cast mission control entity details: %w", err)
		}
		return c.JSON(http.StatusOK, response)
	})
}

func (h *staffHandler) ListMissionControlTimeline(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolveMissionControlTimelineArg, func(principal *controlplanev1.Principal, arg missionControlTimelineArg) error {
		resp, err := h.listMissionControlTimelineCall(c.Request().Context(), principal, arg)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.MissionControlTimelineItems(resp.GetItems(), resp.NextCursor))
	})
}

func (h *staffHandler) SubmitMissionControlCommand(c *echo.Context) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		var body generated.MissionControlCommandRequest
		if err := bindBody(c, &body); err != nil {
			return err
		}

		idempotencyKey := strings.TrimSpace(c.Request().Header.Get("Idempotency-Key"))
		if idempotencyKey == "" {
			return errs.Validation{Field: "Idempotency-Key", Msg: "is required"}
		}

		correlationID := missionControlCorrelationID(c)
		setMissionControlCorrelationHeader(c, correlationID)

		req, err := casters.MissionControlCommandRequest(body, correlationID, time.Now().UTC())
		if err != nil {
			return errs.Validation{Field: "body", Msg: err.Error()}
		}
		req.Principal = principal

		businessIntentKey := strings.TrimSpace(req.GetBusinessIntentKey())
		if businessIntentKey == "" {
			req.BusinessIntentKey = idempotencyKey
			businessIntentKey = idempotencyKey
		}
		if businessIntentKey != idempotencyKey {
			return errs.Validation{Field: "Idempotency-Key", Msg: "must match business_intent_key"}
		}

		item, err := h.cp.Service().SubmitMissionControlCommand(c.Request().Context(), req)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.MissionControlCommandState(item))
	})
}

func (h *staffHandler) GetMissionControlCommand(c *echo.Context) error {
	return getByPathResp(c, "command_id", h.getMissionControlCommandCall, casters.MissionControlCommandState)
}

func (h *staffHandler) fetchMissionControlDashboardSnapshot(
	ctx context.Context,
	principal *controlplanev1.Principal,
	arg missionControlDashboardArg,
) (*controlplanev1.MissionControlDashboardSnapshot, string, error) {
	resp, err := h.getMissionControlDashboardCall(ctx, principal, arg)
	if err != nil {
		return nil, "", err
	}
	if resp == nil || resp.GetSnapshot() == nil {
		return nil, "", fmt.Errorf("mission control snapshot is missing")
	}

	resumeToken, err := encodeMissionControlResumeToken(newMissionControlResumeTokenPayload(arg, resp.GetSnapshot().GetSnapshotId()))
	if err != nil {
		return nil, "", fmt.Errorf("encode mission control resume token: %w", err)
	}

	return resp.GetSnapshot(), resumeToken, nil
}
