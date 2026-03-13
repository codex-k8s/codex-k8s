package entity

import (
	"time"

	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// InteractionResponseRecord stores one typed user response extracted from callback evidence.
type InteractionResponseRecord struct {
	ID               int64
	InteractionID    string
	CallbackEventID  int64
	ResponseKind     enumtypes.InteractionResponseKind
	SelectedOptionID string
	FreeText         string
	ResponderRef     string
	Classification   enumtypes.InteractionCallbackRecordClassification
	IsEffective      bool
	RespondedAt      time.Time
}
