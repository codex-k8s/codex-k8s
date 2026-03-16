package casters

import (
	"github.com/codex-k8s/codex-k8s/libs/go/cast"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

func SystemSetting(item *controlplanev1.SystemSetting) models.SystemSetting {
	if item == nil {
		return models.SystemSetting{}
	}
	return models.SystemSetting{
		Key:                 item.GetKey(),
		Section:             item.GetSection(),
		ValueKind:           item.GetValueKind(),
		ReloadSemantics:     item.GetReloadSemantics(),
		Visibility:          item.GetVisibility(),
		BooleanValue:        item.GetBooleanValue(),
		DefaultBooleanValue: item.GetDefaultBooleanValue(),
		Source:              item.GetSource(),
		Version:             item.GetVersion(),
		UpdatedAt:           cast.OptionalTimestampRFC3339Nano(item.GetUpdatedAt()),
		UpdatedByUserID:     cast.OptionalTrimmedString(item.UpdatedByUserId),
		UpdatedByEmail:      cast.OptionalTrimmedString(item.UpdatedByEmail),
	}
}

func SystemSettings(items []*controlplanev1.SystemSetting) []models.SystemSetting {
	out := make([]models.SystemSetting, 0, len(items))
	for _, item := range items {
		out = append(out, SystemSetting(item))
	}
	return out
}
