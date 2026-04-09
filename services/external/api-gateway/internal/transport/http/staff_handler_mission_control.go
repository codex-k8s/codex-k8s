package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/kodex/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/kodex/proto/gen/go/kodex/controlplane/v1"
	"github.com/codex-k8s/kodex/services/external/api-gateway/internal/transport/http/casters"
	"github.com/codex-k8s/kodex/services/external/api-gateway/internal/transport/http/generated"
)

func (h *staffHandler) GetMissionControlWorkspace(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolveMissionControlWorkspaceArg, func(principal *controlplanev1.Principal, arg missionControlWorkspaceArg) error {
		snapshot, resumeToken, err := h.fetchMissionControlWorkspaceSnapshot(c.Request().Context(), principal, arg)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.MissionControlWorkspaceSnapshot(snapshot, resumeToken))
	})
}

func (h *staffHandler) GetMissionControlNode(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolveMissionControlNodeArg, func(principal *controlplanev1.Principal, arg missionControlNodeArg) error {
		item, err := h.getMissionControlNodeCall(c.Request().Context(), principal, arg)
		if err != nil {
			return err
		}
		response, err := casters.MissionControlNodeDetails(item)
		if err != nil {
			return fmt.Errorf("cast mission control node details: %w", err)
		}
		return c.JSON(http.StatusOK, response)
	})
}

func (h *staffHandler) ListMissionControlNodeActivity(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolveMissionControlActivityArg, func(principal *controlplanev1.Principal, arg missionControlActivityArg) error {
		resp, err := h.listMissionControlNodeActivityCall(c.Request().Context(), principal, arg)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.MissionControlNodeActivityItems(resp.GetItems(), resp.NextCursor))
	})
}

func (h *staffHandler) PreviewMissionControlLaunch(c *echo.Context) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		var body generated.PreviewMissionControlLaunchJSONRequestBody
		if err := bindBody(c, &body); err != nil {
			return err
		}

		item, err := h.previewMissionControlLaunchCall(c.Request().Context(), principal, body)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.MissionControlLaunchPreview(item))
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

func (h *staffHandler) fetchMissionControlWorkspaceSnapshot(
	ctx context.Context,
	principal *controlplanev1.Principal,
	arg missionControlWorkspaceArg,
) (*controlplanev1.MissionControlWorkspaceSnapshot, string, error) {
	resp, err := h.getMissionControlWorkspaceCall(ctx, principal, arg)
	if err != nil {
		return nil, "", err
	}
	if resp == nil || resp.GetSnapshot() == nil {
		return nil, "", fmt.Errorf("mission control workspace snapshot is missing")
	}

	resumeToken, err := encodeMissionControlResumeToken(newMissionControlResumeTokenPayload(arg, resp.GetSnapshot().GetSnapshotId()))
	if err != nil {
		return nil, "", fmt.Errorf("encode mission control resume token: %w", err)
	}

	return resp.GetSnapshot(), resumeToken, nil
}
