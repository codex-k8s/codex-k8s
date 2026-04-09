package casters

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/codex-k8s/kodex/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/kodex/proto/gen/go/kodex/controlplane/v1"
	"github.com/codex-k8s/kodex/services/external/api-gateway/internal/transport/http/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	interactionCallbackSchemaVersion = "telegram-interaction-v1"

	interactionCallbackKindDeliveryReceipt  = "delivery_receipt"
	interactionCallbackKindOptionSelected   = "option_selected"
	interactionCallbackKindFreeTextReceived = "free_text_received"
	interactionCallbackKindTransportFailure = "transport_failure"

	interactionDeliveryStatusAccepted  = "accepted"
	interactionDeliveryStatusDelivered = "delivered"
	interactionDeliveryStatusFailed    = "failed"
)

func InteractionCallbackRequest(item models.InteractionCallbackEnvelope) (*controlplanev1.SubmitInteractionCallbackRequest, error) {
	normalized, occurredAt, err := normalizeInteractionCallbackEnvelope(item)
	if err != nil {
		return nil, err
	}

	rawPayloadJSON, err := json.Marshal(normalized)
	if err != nil {
		return nil, err
	}

	providerMessageRefJSON, err := marshalInteractionProviderMessageRef(normalized.ProviderMessageRef)
	if err != nil {
		return nil, err
	}

	req := &controlplanev1.SubmitInteractionCallbackRequest{
		InteractionId:           normalized.InteractionID,
		DeliveryId:              optionalString(normalized.DeliveryID),
		AdapterEventId:          normalized.AdapterEventID,
		CallbackKind:            normalized.CallbackKind,
		OccurredAt:              timestamppb.New(occurredAt),
		CallbackHandle:          optionalString(normalized.CallbackHandle),
		FreeText:                optionalString(normalized.FreeText),
		ResponderRef:            optionalString(normalized.ResponderRef),
		ProviderMessageRefJson:  providerMessageRefJSON,
		ProviderUpdateId:        optionalString(normalized.ProviderUpdateID),
		ProviderCallbackQueryId: optionalString(normalized.ProviderCallbackQueryID),
		DeliveryStatus:          optionalString(normalized.DeliveryStatus),
		RawPayloadJson:          rawPayloadJSON,
	}
	if normalized.Error != nil {
		req.TransportErrorCode = optionalString(normalized.Error.Code)
		req.TransportRetryable = normalized.Error.Retryable
	}

	return req, nil
}

func InteractionCallbackOutcome(item *controlplanev1.SubmitInteractionCallbackResponse) models.InteractionCallbackOutcome {
	if item == nil {
		return models.InteractionCallbackOutcome{}
	}

	return models.InteractionCallbackOutcome{
		Accepted:           item.GetAccepted(),
		Classification:     strings.TrimSpace(item.GetClassification()),
		InteractionState:   strings.TrimSpace(item.GetInteractionState()),
		ResumeRequired:     item.GetResumeRequired(),
		ContinuationAction: strings.TrimSpace(item.GetContinuationAction()),
	}
}

func normalizeInteractionCallbackEnvelope(item models.InteractionCallbackEnvelope) (models.InteractionCallbackEnvelope, time.Time, error) {
	item.SchemaVersion = strings.TrimSpace(item.SchemaVersion)
	if item.SchemaVersion != interactionCallbackSchemaVersion {
		return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "schema_version", Msg: "must be telegram-interaction-v1"}
	}

	item.InteractionID = strings.TrimSpace(item.InteractionID)
	if item.InteractionID == "" {
		return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "interaction_id", Msg: "is required"}
	}

	item.DeliveryID = strings.TrimSpace(item.DeliveryID)
	item.AdapterEventID = strings.TrimSpace(item.AdapterEventID)
	if item.AdapterEventID == "" {
		return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "adapter_event_id", Msg: "is required"}
	}

	item.CallbackKind = strings.ToLower(strings.TrimSpace(item.CallbackKind))
	switch item.CallbackKind {
	case interactionCallbackKindDeliveryReceipt,
		interactionCallbackKindOptionSelected,
		interactionCallbackKindFreeTextReceived,
		interactionCallbackKindTransportFailure:
	default:
		return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "callback_kind", Msg: "must be delivery_receipt, option_selected, free_text_received, or transport_failure"}
	}

	item.OccurredAt = strings.TrimSpace(item.OccurredAt)
	occurredAt, err := time.Parse(time.RFC3339Nano, item.OccurredAt)
	if err != nil {
		return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "occurred_at", Msg: "must be a valid RFC3339 timestamp"}
	}
	occurredAt = occurredAt.UTC()

	item.CallbackHandle = strings.TrimSpace(item.CallbackHandle)
	item.FreeText = strings.TrimSpace(item.FreeText)
	item.ResponderRef = strings.TrimSpace(item.ResponderRef)
	item.ProviderUpdateID = strings.TrimSpace(item.ProviderUpdateID)
	item.ProviderCallbackQueryID = strings.TrimSpace(item.ProviderCallbackQueryID)
	item.DeliveryStatus = strings.ToLower(strings.TrimSpace(item.DeliveryStatus))

	normalizedProviderMessageRef, err := normalizeInteractionProviderMessageRef(item.ProviderMessageRef)
	if err != nil {
		return models.InteractionCallbackEnvelope{}, time.Time{}, err
	}
	item.ProviderMessageRef = normalizedProviderMessageRef

	if item.Error != nil {
		item.Error.Code = strings.TrimSpace(item.Error.Code)
		item.Error.Message = strings.TrimSpace(item.Error.Message)
	}

	switch item.CallbackKind {
	case interactionCallbackKindDeliveryReceipt:
		if item.CallbackHandle != "" {
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "callback_handle", Msg: "must be omitted for delivery_receipt"}
		}
		if item.FreeText != "" {
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "free_text", Msg: "must be omitted for delivery_receipt"}
		}
		if item.Error != nil {
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "error", Msg: "must be omitted for delivery_receipt"}
		}
		switch item.DeliveryStatus {
		case interactionDeliveryStatusAccepted, interactionDeliveryStatusDelivered, interactionDeliveryStatusFailed:
		case "":
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "delivery_status", Msg: "is required for delivery_receipt"}
		default:
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "delivery_status", Msg: "must be accepted, delivered, or failed"}
		}
	case interactionCallbackKindOptionSelected:
		if err := validateResolvedInteractionCallback(item, interactionCallbackKindOptionSelected, false); err != nil {
			return models.InteractionCallbackEnvelope{}, time.Time{}, err
		}
	case interactionCallbackKindFreeTextReceived:
		if err := validateResolvedInteractionCallback(item, interactionCallbackKindFreeTextReceived, true); err != nil {
			return models.InteractionCallbackEnvelope{}, time.Time{}, err
		}
	case interactionCallbackKindTransportFailure:
		if item.CallbackHandle != "" {
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "callback_handle", Msg: "must be omitted for transport_failure"}
		}
		if item.FreeText != "" {
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "free_text", Msg: "must be omitted for transport_failure"}
		}
		if item.DeliveryStatus != "" {
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "delivery_status", Msg: "must be omitted for transport_failure"}
		}
		if item.Error == nil {
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "error", Msg: "is required for transport_failure"}
		}
		if item.Error.Code == "" {
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "error.code", Msg: "is required for transport_failure"}
		}
	}

	item.OccurredAt = occurredAt.Format(time.RFC3339Nano)
	return item, occurredAt, nil
}

func normalizeInteractionProviderMessageRef(item *models.InteractionProviderMessageRef) (*models.InteractionProviderMessageRef, error) {
	if item == nil {
		return nil, nil
	}

	normalized := *item
	normalized.ChatRef = strings.TrimSpace(normalized.ChatRef)
	normalized.MessageID = strings.TrimSpace(normalized.MessageID)
	normalized.InlineMessageID = strings.TrimSpace(normalized.InlineMessageID)
	normalized.SentAt = strings.TrimSpace(normalized.SentAt)

	if normalized.SentAt != "" {
		sentAt, err := time.Parse(time.RFC3339Nano, normalized.SentAt)
		if err != nil {
			return nil, errs.Validation{Field: "provider_message_ref.sent_at", Msg: "must be a valid RFC3339 timestamp"}
		}
		normalized.SentAt = sentAt.UTC().Format(time.RFC3339Nano)
	}
	if normalized.ChatRef == "" && normalized.MessageID == "" && normalized.InlineMessageID == "" && normalized.SentAt == "" {
		return nil, nil
	}
	return &normalized, nil
}

func validateResolvedInteractionCallback(item models.InteractionCallbackEnvelope, callbackKind string, requireFreeText bool) error {
	if item.CallbackHandle == "" {
		return errs.Validation{Field: "callback_handle", Msg: "is required for " + callbackKind}
	}
	if requireFreeText {
		if item.FreeText == "" {
			return errs.Validation{Field: "free_text", Msg: "is required for " + callbackKind}
		}
	} else if item.FreeText != "" {
		return errs.Validation{Field: "free_text", Msg: "must be omitted for " + callbackKind}
	}
	if item.DeliveryStatus != "" {
		return errs.Validation{Field: "delivery_status", Msg: "must be omitted for " + callbackKind}
	}
	if item.Error != nil {
		return errs.Validation{Field: "error", Msg: "must be omitted for " + callbackKind}
	}

	return nil
}

func marshalInteractionProviderMessageRef(item *models.InteractionProviderMessageRef) ([]byte, error) {
	if item == nil {
		return nil, nil
	}
	return json.Marshal(item)
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
