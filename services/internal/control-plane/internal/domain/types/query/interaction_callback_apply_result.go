package query

import (
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// InteractionCallbackApplyResult describes aggregate mutation after callback classification.
type InteractionCallbackApplyResult struct {
	Interaction         entitytypes.InteractionRequest
	Binding             *entitytypes.InteractionChannelBinding
	CallbackEvent       entitytypes.InteractionCallbackEvent
	ResponseRecord      *entitytypes.InteractionResponseRecord
	Accepted            bool
	Classification      enumtypes.InteractionCallbackResultClassification
	EffectiveResponseID int64
	ContinuationAction  enumtypes.InteractionContinuationAction
	OperatorSignalCode  enumtypes.InteractionOperatorSignalCode
	ResumeRequired      bool
}
