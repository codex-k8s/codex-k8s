package query

import enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"

// SystemSettingBooleanWriteParams carries one typed boolean setting mutation.
type SystemSettingBooleanWriteParams struct {
	Key          enumtypes.SystemSettingKey
	BooleanValue bool
	Source       enumtypes.SystemSettingSource
	ChangeKind   enumtypes.SystemSettingChangeKind
	ActorUserID  string
	ActorEmail   string
}
