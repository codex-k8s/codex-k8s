package interactionrequest

import (
	domainrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/interactionrequest"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	"github.com/codex-k8s/kodex/services/internal/control-plane/internal/repository/postgres/interactionrequest/dbmodel"
)

func requestFromDBModel(row dbmodel.RequestRow) domainrepo.Request {
	item := domainrepo.Request{
		ID:                    row.ID,
		ProjectID:             row.ProjectID,
		RunID:                 row.RunID,
		InteractionKind:       enumtypes.InteractionKind(row.InteractionKind),
		ChannelFamily:         enumtypes.InteractionChannelFamily(row.ChannelFamily),
		State:                 enumtypes.InteractionState(row.State),
		ResolutionKind:        enumtypes.InteractionResolutionKind(row.ResolutionKind),
		RecipientProvider:     row.RecipientProvider,
		RecipientRef:          row.RecipientRef,
		RequestPayloadJSON:    row.RequestPayloadJSON,
		ContextLinksJSON:      row.ContextLinksJSON,
		OperatorState:         enumtypes.InteractionOperatorState(row.OperatorState),
		LastDeliveryAttemptNo: int(row.LastDeliveryAttemptNo),
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
	if row.ResponseDeadlineAt.Valid {
		value := row.ResponseDeadlineAt.Time
		item.ResponseDeadlineAt = &value
	}
	if item.ChannelFamily == "" {
		item.ChannelFamily = enumtypes.InteractionChannelFamilyPlatformOnly
	}
	if row.EffectiveResponseID.Valid {
		item.EffectiveResponseID = row.EffectiveResponseID.Int64
	}
	if row.ActiveChannelBindingID.Valid {
		item.ActiveChannelBindingID = row.ActiveChannelBindingID.Int64
	}
	if item.OperatorState == "" {
		item.OperatorState = enumtypes.InteractionOperatorStateNominal
	}
	if row.OperatorSignalCode.Valid {
		item.OperatorSignalCode = enumtypes.InteractionOperatorSignalCode(row.OperatorSignalCode.String)
	}
	if row.OperatorSignalAt.Valid {
		value := row.OperatorSignalAt.Time
		item.OperatorSignalAt = &value
	}
	return item
}

func deliveryAttemptFromDBModel(row dbmodel.DeliveryAttemptRow) domainrepo.DeliveryAttempt {
	item := domainrepo.DeliveryAttempt{
		ID:                     row.ID,
		InteractionID:          row.InteractionID,
		AttemptNo:              int(row.AttemptNo),
		DeliveryID:             row.DeliveryID,
		AdapterKind:            row.AdapterKind,
		DeliveryRole:           enumtypes.InteractionDeliveryRole(row.DeliveryRole),
		Status:                 enumtypes.InteractionDeliveryAttemptStatus(row.Status),
		RequestEnvelopeJSON:    row.RequestEnvelopeJSON,
		AckPayloadJSON:         row.AckPayloadJSON,
		ProviderMessageRefJSON: row.ProviderMessageRefJSON,
		Retryable:              row.Retryable,
		StartedAt:              row.StartedAt,
	}
	if row.ChannelBindingID.Valid {
		item.ChannelBindingID = row.ChannelBindingID.Int64
	}
	if item.DeliveryRole == "" {
		item.DeliveryRole = enumtypes.InteractionDeliveryRolePrimaryDispatch
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
	if row.ContinuationReason.Valid {
		item.ContinuationReason = row.ContinuationReason.String
	}
	if row.FinishedAt.Valid {
		value := row.FinishedAt.Time
		item.FinishedAt = &value
	}
	return item
}

func callbackEventFromDBModel(row dbmodel.CallbackEventRow) domainrepo.CallbackEvent {
	item := domainrepo.CallbackEvent{
		ID:                     row.ID,
		InteractionID:          row.InteractionID,
		AdapterEventID:         row.AdapterEventID,
		CallbackKind:           enumtypes.InteractionCallbackKind(row.CallbackKind),
		Classification:         enumtypes.InteractionCallbackRecordClassification(row.Classification),
		CallbackHandleHash:     row.CallbackHandleHash,
		NormalizedPayloadJSON:  row.NormalizedPayloadJSON,
		RawPayloadJSON:         row.RawPayloadJSON,
		ProviderMessageRefJSON: row.ProviderMessageRefJSON,
		ReceivedAt:             row.ReceivedAt,
	}
	if row.ChannelBindingID.Valid {
		item.ChannelBindingID = row.ChannelBindingID.Int64
	}
	if row.DeliveryID.Valid {
		item.DeliveryID = row.DeliveryID.String
	}
	if row.ProviderUpdateID.Valid {
		item.ProviderUpdateID = row.ProviderUpdateID.String
	}
	if row.ProviderCallbackQueryID.Valid {
		item.ProviderCallbackQueryID = row.ProviderCallbackQueryID.String
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
		ChannelBindingID: row.ChannelBindingID,
		CallbackEventID:  row.CallbackEventID,
		HandleKind:       enumtypes.InteractionCallbackHandleKind(row.HandleKind),
		ResponseKind:     enumtypes.InteractionResponseKind(row.ResponseKind),
		SelectedOptionID: row.SelectedOptionID,
		FreeText:         row.FreeText,
		ResponderRef:     row.ResponderRef,
		Classification:   enumtypes.InteractionCallbackRecordClassification(row.Classification),
		IsEffective:      row.IsEffective,
		RespondedAt:      row.RespondedAt,
	}
}

func channelBindingFromDBModel(row dbmodel.ChannelBindingRow) domainrepo.ChannelBinding {
	item := domainrepo.ChannelBinding{
		ID:                     row.ID,
		InteractionID:          row.InteractionID,
		AdapterKind:            row.AdapterKind,
		RecipientRef:           row.RecipientRef,
		ProviderMessageRefJSON: row.ProviderMessageRefJSON,
		EditCapability:         enumtypes.InteractionEditCapability(row.EditCapability),
		ContinuationState:      enumtypes.InteractionContinuationState(row.ContinuationState),
		CreatedAt:              row.CreatedAt,
		UpdatedAt:              row.UpdatedAt,
	}
	if row.ProviderChatRef.Valid {
		item.ProviderChatRef = row.ProviderChatRef.String
	}
	if item.EditCapability == "" {
		item.EditCapability = enumtypes.InteractionEditCapabilityUnknown
	}
	if item.ContinuationState == "" {
		item.ContinuationState = enumtypes.InteractionContinuationStatePendingPrimaryDelivery
	}
	if row.CallbackTokenKeyID.Valid {
		item.CallbackTokenKeyID = row.CallbackTokenKeyID.String
	}
	if row.CallbackTokenExpiresAt.Valid {
		value := row.CallbackTokenExpiresAt.Time
		item.CallbackTokenExpiresAt = &value
	}
	if row.LastOperatorSignalCode.Valid {
		item.LastOperatorSignalCode = enumtypes.InteractionOperatorSignalCode(row.LastOperatorSignalCode.String)
	}
	if row.LastOperatorSignalAt.Valid {
		value := row.LastOperatorSignalAt.Time
		item.LastOperatorSignalAt = &value
	}
	return item
}

func callbackHandleFromDBModel(row dbmodel.CallbackHandleRow) domainrepo.CallbackHandle {
	item := domainrepo.CallbackHandle{
		ID:                 row.ID,
		InteractionID:      row.InteractionID,
		ChannelBindingID:   row.ChannelBindingID,
		HandleHash:         row.HandleHash,
		HandleKind:         enumtypes.InteractionCallbackHandleKind(row.HandleKind),
		State:              enumtypes.InteractionCallbackHandleState(row.State),
		ResponseDeadlineAt: row.ResponseDeadlineAt,
		GraceExpiresAt:     row.GraceExpiresAt,
		CreatedAt:          row.CreatedAt,
	}
	if row.OptionID.Valid {
		item.OptionID = row.OptionID.String
	}
	if item.State == "" {
		item.State = enumtypes.InteractionCallbackHandleStateOpen
	}
	if row.UsedCallbackEventID.Valid {
		item.UsedCallbackEventID = row.UsedCallbackEventID.Int64
	}
	if row.UsedAt.Valid {
		value := row.UsedAt.Time
		item.UsedAt = &value
	}
	return item
}
