package http

import (
	"context"
	"net/http"

	sharedsystemsettings "github.com/codex-k8s/codex-k8s/libs/go/systemsettings"
	"github.com/labstack/echo/v5"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

func (h *staffHandler) ListSystemSettings(c *echo.Context) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		resp, err := h.listSystemSettingsCall(c.Request().Context(), principal)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, models.ItemsResponse[models.SystemSetting]{Items: casters.SystemSettings(resp.GetItems())})
	})
}

func (h *staffHandler) GetSystemSetting(c *echo.Context) error {
	return h.respondSystemSettingByKey(c, h.getSystemSettingCall)
}

func (h *staffHandler) UpdateSystemSettingBoolean(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolveSystemSettingUpdateArg, func(principal *controlplanev1.Principal, arg systemSettingUpdateArg) error {
		item, err := h.updateSystemSettingBooleanCall(c.Request().Context(), principal, arg)
		return h.writeSystemSettingResponse(c, item, err)
	})
}

func (h *staffHandler) ResetSystemSetting(c *echo.Context) error {
	return h.respondSystemSettingByKey(c, h.resetSystemSettingCall)
}

func (h *staffHandler) respondSystemSettingByKey(c *echo.Context, call systemSettingCall) error {
	return withPrincipalAndResolved(c, resolvePathUnescaped("setting_key"), func(principal *controlplanev1.Principal, key string) error {
		item, err := call(c.Request().Context(), principal, key)
		return h.writeSystemSettingResponse(c, item, err)
	})
}

func (h *staffHandler) writeSystemSettingResponse(c *echo.Context, item *controlplanev1.SystemSetting, err error) error {
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, casters.SystemSetting(item))
}

func (h *staffHandler) SystemSettingsRealtime(c *echo.Context) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		return streamRealtimeSnapshots(
			c,
			func(ctx context.Context) ([]models.SystemSetting, error) {
				resp, err := h.listSystemSettingsCall(ctx, principal)
				if err != nil {
					return nil, err
				}
				return filterRealtimeSystemSettings(casters.SystemSettings(resp.GetItems())), nil
			},
			func(items []models.SystemSetting) any {
				return models.SystemSettingsRealtimeMessage{
					Type:   models.ListRealtimeMessageTypeSnapshot,
					Items:  items,
					SentAt: realtimeSentAt(),
				}
			},
			func(err error) any {
				return models.SystemSettingsRealtimeMessage{
					Type:    models.ListRealtimeMessageTypeError,
					Message: realtimeErrorMessagePtr(err),
					SentAt:  realtimeSentAt(),
				}
			},
		)
	})
}

func resolveSystemSettingUpdateArg(c *echo.Context) (systemSettingUpdateArg, error) {
	key, err := resolvePathUnescaped("setting_key")(c)
	if err != nil {
		return systemSettingUpdateArg{}, err
	}

	var body models.UpdateSystemSettingBooleanRequest
	if err := bindBody(c, &body); err != nil {
		return systemSettingUpdateArg{}, err
	}
	return systemSettingUpdateArg{key: key, body: body}, nil
}

type systemSettingCall func(context.Context, *controlplanev1.Principal, string) (*controlplanev1.SystemSetting, error)

func filterRealtimeSystemSettings(items []models.SystemSetting) []models.SystemSetting {
	out := make([]models.SystemSetting, 0, len(items))
	for _, item := range items {
		if !sharedsystemsettings.IsVisibleOnSurface(item.Visibility, sharedsystemsettings.ExposureSurfaceRealtime) {
			continue
		}
		out = append(out, item)
	}
	return out
}
