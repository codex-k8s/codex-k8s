package query

import (
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// InteractionCallbackApplyResult describes aggregate mutation after callback classification.
type InteractionCallbackApplyResult struct {
	Interaction         entitytypes.InteractionRequest
	CallbackEvent       entitytypes.InteractionCallbackEvent
	ResponseRecord      *entitytypes.InteractionResponseRecord
	Classification      enumtypes.InteractionCallbackResultClassification
	EffectiveResponseID int64
	ResumeRequired      bool
}
