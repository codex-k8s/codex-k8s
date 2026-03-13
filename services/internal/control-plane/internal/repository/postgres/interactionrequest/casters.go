package interactionrequest

import (
	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/interactionrequest"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/interactionrequest/dbmodel"
)

func requestFromDBModel(row dbmodel.RequestRow) domainrepo.Request {
	item := domainrepo.Request{
		ID:                    row.ID,
		ProjectID:             row.ProjectID,
		RunID:                 row.RunID,
		InteractionKind:       enumtypes.InteractionKind(row.InteractionKind),
		State:                 enumtypes.InteractionState(row.State),
		ResolutionKind:        enumtypes.InteractionResolutionKind(row.ResolutionKind),
		RecipientProvider:     row.RecipientProvider,
		RecipientRef:          row.RecipientRef,
		RequestPayloadJSON:    row.RequestPayloadJSON,
		ContextLinksJSON:      row.ContextLinksJSON,
		LastDeliveryAttemptNo: int(row.LastDeliveryAttemptNo),
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
	if row.ResponseDeadlineAt.Valid {
		value := row.ResponseDeadlineAt.Time
		item.ResponseDeadlineAt = &value
	}
	if row.EffectiveResponseID.Valid {
		item.EffectiveResponseID = row.EffectiveResponseID.Int64
	}
	return item
}

func deliveryAttemptFromDBModel(row dbmodel.DeliveryAttemptRow) domainrepo.DeliveryAttempt {
	item := domainrepo.DeliveryAttempt{
		ID:                  row.ID,
		InteractionID:       row.InteractionID,
		AttemptNo:           int(row.AttemptNo),
		DeliveryID:          row.DeliveryID,
		AdapterKind:         row.AdapterKind,
		Status:              enumtypes.InteractionDeliveryAttemptStatus(row.Status),
		RequestEnvelopeJSON: row.RequestEnvelopeJSON,
		AckPayloadJSON:      row.AckPayloadJSON,
		Retryable:           row.Retryable,
		StartedAt:           row.StartedAt,
	}
	if row.AdapterDeliveryID.Valid {
		item.AdapterDeliveryID = row.AdapterDeliveryID.String
	}
	if row.NextRetryAt.Valid {
		value := row.NextRetryAt.Time
		item.NextRetryAt = &value
	}
	if row.LastErrorCode.Valid {
		item.LastErrorCode = row.LastErrorCode.String
	}
	if row.FinishedAt.Valid {
		value := row.FinishedAt.Time
		item.FinishedAt = &value
	}
	return item
}

func callbackEventFromDBModel(row dbmodel.CallbackEventRow) domainrepo.CallbackEvent {
	item := domainrepo.CallbackEvent{
		ID:                    row.ID,
		InteractionID:         row.InteractionID,
		AdapterEventID:        row.AdapterEventID,
		CallbackKind:          enumtypes.InteractionCallbackKind(row.CallbackKind),
		Classification:        enumtypes.InteractionCallbackRecordClassification(row.Classification),
		NormalizedPayloadJSON: row.NormalizedPayloadJSON,
		RawPayloadJSON:        row.RawPayloadJSON,
		ReceivedAt:            row.ReceivedAt,
	}
	if row.DeliveryID.Valid {
		item.DeliveryID = row.DeliveryID.String
	}
	if row.ProcessedAt.Valid {
		value := row.ProcessedAt.Time
		item.ProcessedAt = &value
	}
	return item
}

func responseRecordFromDBModel(row dbmodel.ResponseRecordRow) domainrepo.ResponseRecord {
	return domainrepo.ResponseRecord{
		ID:               row.ID,
		InteractionID:    row.InteractionID,
		CallbackEventID:  row.CallbackEventID,
		ResponseKind:     enumtypes.InteractionResponseKind(row.ResponseKind),
		SelectedOptionID: row.SelectedOptionID,
		FreeText:         row.FreeText,
		ResponderRef:     row.ResponderRef,
		Classification:   enumtypes.InteractionCallbackRecordClassification(row.Classification),
		IsEffective:      row.IsEffective,
		RespondedAt:      row.RespondedAt,
	}
}
