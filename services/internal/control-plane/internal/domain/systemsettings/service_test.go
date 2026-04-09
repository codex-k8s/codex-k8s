package systemsettings

import (
	"context"
	"errors"
	"testing"

	"github.com/codex-k8s/kodex/libs/go/errs"
	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

func TestServiceUnknownKey_ReturnsNotFoundAcrossOperations(t *testing.T) {
	t.Parallel()

	svc, err := NewService(systemSettingsRepositoryStub{})
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	checkNotFound := func(t *testing.T, err error) {
		t.Helper()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var notFound errs.NotFound
		if !errors.As(err, &notFound) {
			t.Fatalf("expected errs.NotFound, got %T", err)
		}
	}

	t.Run("get", func(t *testing.T) {
		_, err := svc.Get("missing_setting")
		checkNotFound(t, err)
	})

	t.Run("update", func(t *testing.T) {
		_, err := svc.UpdateBoolean(context.Background(), querytypes.SystemSettingBooleanWriteParams{
			Key:          enumtypes.SystemSettingKey("missing_setting"),
			BooleanValue: true,
		})
		checkNotFound(t, err)
	})

	t.Run("reset", func(t *testing.T) {
		_, err := svc.Reset(context.Background(), "missing_setting", "user-1", "user@example.com")
		checkNotFound(t, err)
	})
}

func TestServiceStaffVisibility_FiltersInternalOnlySettings(t *testing.T) {
	t.Parallel()

	staffKey := enumtypes.SystemSettingKey("staff_visible_setting")
	internalKey := enumtypes.SystemSettingKey("internal_only_setting")

	svc, err := newService(systemSettingsRepositoryStub{}, map[enumtypes.SystemSettingKey]valuetypes.SystemSettingCatalogEntry{
		staffKey: {
			Key:                 staffKey,
			Section:             enumtypes.SystemSettingSectionGitHubRateLimit,
			ValueKind:           enumtypes.SystemSettingValueKindBoolean,
			ReloadSemantics:     enumtypes.SystemSettingReloadSemanticsHotReload,
			Visibility:          enumtypes.SystemSettingVisibilityStaffVisible,
			DefaultBooleanValue: false,
		},
		internalKey: {
			Key:                 internalKey,
			Section:             enumtypes.SystemSettingSectionGitHubRateLimit,
			ValueKind:           enumtypes.SystemSettingValueKindBoolean,
			ReloadSemantics:     enumtypes.SystemSettingReloadSemanticsHotReload,
			Visibility:          enumtypes.SystemSettingVisibilityInternalOnly,
			DefaultBooleanValue: false,
		},
	})
	if err != nil {
		t.Fatalf("newService returned error: %v", err)
	}
	svc.replaceCache([]entitytypes.SystemSettingRecord{
		{Key: staffKey, ValueKind: enumtypes.SystemSettingValueKindBoolean, BooleanValue: true, Source: enumtypes.SystemSettingSourceStaff, Version: 2},
		{Key: internalKey, ValueKind: enumtypes.SystemSettingValueKindBoolean, BooleanValue: true, Source: enumtypes.SystemSettingSourceStaff, Version: 3},
	})

	items := svc.List()
	if len(items) != 1 {
		t.Fatalf("List returned %d items, want 1", len(items))
	}
	if got, want := items[0].Key, staffKey; got != want {
		t.Fatalf("List returned key %q, want %q", got, want)
	}

	_, err = svc.Get(string(internalKey))
	if err == nil {
		t.Fatal("expected Get to hide internal-only setting")
	}
	var notFound errs.NotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected errs.NotFound for internal-only Get, got %T", err)
	}
}

type systemSettingsRepositoryStub struct{}

func (systemSettingsRepositoryStub) List(context.Context) ([]entitytypes.SystemSettingRecord, error) {
	return nil, nil
}

func (systemSettingsRepositoryStub) UpsertBoolean(context.Context, querytypes.SystemSettingBooleanWriteParams) (entitytypes.SystemSettingRecord, error) {
	return entitytypes.SystemSettingRecord{}, nil
}
