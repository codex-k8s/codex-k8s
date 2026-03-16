package staff

import (
	"context"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func (s *Service) ListSystemSettings(_ context.Context, principal Principal) ([]entitytypes.SystemSetting, error) {
	if !principal.IsPlatformAdmin {
		return nil, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.systemSettings == nil {
		return nil, errs.FailedPrecondition{Msg: "system settings service is not configured"}
	}
	return s.systemSettings.List(), nil
}

func (s *Service) GetSystemSetting(_ context.Context, principal Principal, key string) (entitytypes.SystemSetting, error) {
	if !principal.IsPlatformAdmin {
		return entitytypes.SystemSetting{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.systemSettings == nil {
		return entitytypes.SystemSetting{}, errs.FailedPrecondition{Msg: "system settings service is not configured"}
	}

	return s.systemSettings.Get(key)
}

func (s *Service) UpdateSystemSettingBoolean(ctx context.Context, principal Principal, key string, value bool) (entitytypes.SystemSetting, error) {
	if !principal.IsPlatformAdmin {
		return entitytypes.SystemSetting{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.systemSettings == nil {
		return entitytypes.SystemSetting{}, errs.FailedPrecondition{Msg: "system settings service is not configured"}
	}

	return s.systemSettings.UpdateBoolean(ctx, querytypes.SystemSettingBooleanWriteParams{
		Key:          enumtypes.SystemSettingKey(key),
		BooleanValue: value,
		Source:       enumtypes.SystemSettingSourceStaff,
		ChangeKind:   enumtypes.SystemSettingChangeKindUpdated,
		ActorUserID:  principal.UserID,
		ActorEmail:   principal.Email,
	})
}

func (s *Service) ResetSystemSetting(ctx context.Context, principal Principal, key string) (entitytypes.SystemSetting, error) {
	if !principal.IsPlatformAdmin {
		return entitytypes.SystemSetting{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.systemSettings == nil {
		return entitytypes.SystemSetting{}, errs.FailedPrecondition{Msg: "system settings service is not configured"}
	}
	return s.systemSettings.Reset(ctx, key, principal.UserID, principal.Email)
}
