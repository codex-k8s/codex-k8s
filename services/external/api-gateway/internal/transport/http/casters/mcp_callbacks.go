package casters

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	interactionCallbackSchemaVersion = "v1"

	interactionCallbackKindDeliveryReceipt  = "delivery_receipt"
	interactionCallbackKindDecisionResponse = "decision_response"

	interactionResponseKindOption   = "option"
	interactionResponseKindFreeText = "free_text"

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

	req := &controlplanev1.SubmitInteractionCallbackRequest{
		InteractionId:  normalized.InteractionID,
		DeliveryId:     optionalString(normalized.DeliveryID),
		AdapterEventId: normalized.AdapterEventID,
		CallbackKind:   normalized.CallbackKind,
		OccurredAt:     timestamppb.New(occurredAt),
		DeliveryStatus: optionalString(normalized.DeliveryStatus),
		RawPayloadJson: rawPayloadJSON,
	}
	if normalized.Response != nil {
		req.ResponseKind = optionalString(normalized.Response.ResponseKind)
		req.SelectedOptionId = optionalString(normalized.Response.SelectedOptionID)
		req.FreeText = optionalString(normalized.Response.FreeText)
		req.ResponderRef = optionalString(normalized.Response.ResponderRef)
	}

	return req, nil
}

func InteractionCallbackOutcome(item *controlplanev1.SubmitInteractionCallbackResponse) models.InteractionCallbackOutcome {
	if item == nil {
		return models.InteractionCallbackOutcome{}
	}

	return models.InteractionCallbackOutcome{
		Accepted:         item.GetAccepted(),
		Classification:   strings.TrimSpace(item.GetClassification()),
		InteractionState: strings.TrimSpace(item.GetInteractionState()),
		ResumeRequired:   item.GetResumeRequired(),
	}
}

func normalizeInteractionCallbackEnvelope(item models.InteractionCallbackEnvelope) (models.InteractionCallbackEnvelope, time.Time, error) {
	item.SchemaVersion = strings.TrimSpace(item.SchemaVersion)
	if item.SchemaVersion != interactionCallbackSchemaVersion {
		return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "schema_version", Msg: "must be v1"}
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
	case interactionCallbackKindDeliveryReceipt, interactionCallbackKindDecisionResponse:
	default:
		return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "callback_kind", Msg: "must be delivery_receipt or decision_response"}
	}

	item.OccurredAt = strings.TrimSpace(item.OccurredAt)
	occurredAt, err := time.Parse(time.RFC3339Nano, item.OccurredAt)
	if err != nil {
		return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "occurred_at", Msg: "must be a valid RFC3339 timestamp"}
	}
	occurredAt = occurredAt.UTC()

	item.AdapterDeliveryID = strings.TrimSpace(item.AdapterDeliveryID)
	item.DeliveryStatus = strings.ToLower(strings.TrimSpace(item.DeliveryStatus))

	switch item.CallbackKind {
	case interactionCallbackKindDeliveryReceipt:
		if item.Response != nil {
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "response", Msg: "must be omitted for delivery_receipt"}
		}
		switch item.DeliveryStatus {
		case interactionDeliveryStatusAccepted, interactionDeliveryStatusDelivered, interactionDeliveryStatusFailed:
		case "":
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "delivery_status", Msg: "is required for delivery_receipt"}
		default:
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "delivery_status", Msg: "must be accepted, delivered, or failed"}
		}
	case interactionCallbackKindDecisionResponse:
		if item.DeliveryStatus != "" {
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "delivery_status", Msg: "must be omitted for decision_response"}
		}
		if item.Response == nil {
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "response", Msg: "is required for decision_response"}
		}
		normalizedResponse := *item.Response
		normalizedResponse.ResponseKind = strings.ToLower(strings.TrimSpace(normalizedResponse.ResponseKind))
		normalizedResponse.SelectedOptionID = strings.TrimSpace(normalizedResponse.SelectedOptionID)
		normalizedResponse.FreeText = strings.TrimSpace(normalizedResponse.FreeText)
		normalizedResponse.ResponderRef = strings.TrimSpace(normalizedResponse.ResponderRef)

		switch normalizedResponse.ResponseKind {
		case interactionResponseKindOption:
			if normalizedResponse.SelectedOptionID == "" {
				return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "response.selected_option_id", Msg: "is required for option response"}
			}
		case interactionResponseKindFreeText:
			if normalizedResponse.FreeText == "" {
				return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "response.free_text", Msg: "is required for free_text response"}
			}
		default:
			return models.InteractionCallbackEnvelope{}, time.Time{}, errs.Validation{Field: "response.response_kind", Msg: "must be option or free_text"}
		}

		item.Response = &normalizedResponse
	}

	if item.Error != nil {
		normalizedError := *item.Error
		normalizedError.Code = strings.TrimSpace(normalizedError.Code)
		normalizedError.Message = strings.TrimSpace(normalizedError.Message)
		item.Error = &normalizedError
	}

	item.OccurredAt = occurredAt.Format(time.RFC3339Nano)
	return item, occurredAt, nil
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
