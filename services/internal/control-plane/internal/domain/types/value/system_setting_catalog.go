package value

import enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"

// SystemSettingCatalogEntry is one typed runtime setting definition.
type SystemSettingCatalogEntry struct {
	Key                 enumtypes.SystemSettingKey
	Section             enumtypes.SystemSettingSection
	ValueKind           enumtypes.SystemSettingValueKind
	ReloadSemantics     enumtypes.SystemSettingReloadSemantics
	Visibility          enumtypes.SystemSettingVisibility
	DefaultBooleanValue bool
}
