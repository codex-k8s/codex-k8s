package systemsetting

import (
	"context"

	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
)

type (
	SystemSettingRecord = entitytypes.SystemSettingRecord
	BooleanWriteParams  = querytypes.SystemSettingBooleanWriteParams
)

// Repository stores typed platform settings in control-plane-owned PostgreSQL tables.
type Repository interface {
	List(ctx context.Context) ([]SystemSettingRecord, error)
	UpsertBoolean(ctx context.Context, params BooleanWriteParams) (SystemSettingRecord, error)
}
