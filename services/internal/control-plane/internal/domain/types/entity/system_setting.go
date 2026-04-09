package entity

import (
	"time"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

// SystemSettingRecord mirrors one persisted system_settings row.
type SystemSettingRecord struct {
	Key             enumtypes.SystemSettingKey
	ValueKind       enumtypes.SystemSettingValueKind
	BooleanValue    bool
	Source          enumtypes.SystemSettingSource
	Version         int64
	UpdatedAt       time.Time
	UpdatedByUserID string
	UpdatedByEmail  string
}

// SystemSetting is one staff-visible platform setting merged with catalog metadata.
type SystemSetting struct {
	Key                 enumtypes.SystemSettingKey
	Section             enumtypes.SystemSettingSection
	ValueKind           enumtypes.SystemSettingValueKind
	ReloadSemantics     enumtypes.SystemSettingReloadSemantics
	Visibility          enumtypes.SystemSettingVisibility
	BooleanValue        bool
	DefaultBooleanValue bool
	Source              enumtypes.SystemSettingSource
	Version             int64
	UpdatedAt           *time.Time
	UpdatedByUserID     string
	UpdatedByEmail      string
}
