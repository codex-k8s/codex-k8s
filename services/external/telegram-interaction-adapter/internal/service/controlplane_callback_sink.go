package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	controlplaneclient "github.com/codex-k8s/codex-k8s/services/external/telegram-interaction-adapter/internal/controlplane"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ControlPlaneCallbackSink submits normalized Telegram callbacks over internal gRPC.
type ControlPlaneCallbackSink struct {
	client *controlplaneclient.Client
}

// NewControlPlaneCallbackSink builds a gRPC-backed callback sink.
func NewControlPlaneCallbackSink(client *controlplaneclient.Client) *ControlPlaneCallbackSink {
	return &ControlPlaneCallbackSink{client: client}
}

// Submit forwards one normalized callback to control-plane and returns the typed outcome.
func (s *ControlPlaneCallbackSink) Submit(ctx context.Context, envelope CallbackEnvelope) (CallbackOutcome, error) {
	if s == nil || s.client == nil {
		return CallbackOutcome{}, fmt.Errorf("control-plane callback client is not configured")
	}

	occurredAt, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(envelope.OccurredAt))
	if err != nil {
		return CallbackOutcome{}, fmt.Errorf("parse callback occurred_at: %w", err)
	}

	var providerMessageRefJSON []byte
	if envelope.ProviderMessageRef != nil {
		providerMessageRefJSON, err = json.Marshal(envelope.ProviderMessageRef)
		if err != nil {
			return CallbackOutcome{}, fmt.Errorf("marshal provider_message_ref: %w", err)
		}
	}

	rawPayload, err := json.Marshal(envelope)
	if err != nil {
		return CallbackOutcome{}, fmt.Errorf("marshal adapter callback payload: %w", err)
	}

	resp, err := s.client.SubmitAdapterInteractionCallback(ctx, &controlplanev1.SubmitInteractionCallbackRequest{
		InteractionId:           strings.TrimSpace(envelope.InteractionID),
		DeliveryId:              optionalStringValue(envelope.DeliveryID),
		AdapterEventId:          strings.TrimSpace(envelope.AdapterEventID),
		CallbackKind:            strings.TrimSpace(envelope.CallbackKind),
		OccurredAt:              timestamppb.New(occurredAt.UTC()),
		CallbackHandle:          optionalStringValue(envelope.CallbackHandle),
		FreeText:                optionalStringValue(envelope.FreeText),
		ResponderRef:            optionalStringValue(envelope.ResponderRef),
		ProviderMessageRefJson:  providerMessageRefJSON,
		ProviderUpdateId:        optionalStringValue(envelope.ProviderUpdateID),
		ProviderCallbackQueryId: optionalStringValue(envelope.ProviderCallbackQueryID),
		RawPayloadJson:          rawPayload,
	})
	if err != nil {
		return CallbackOutcome{}, err
	}

	return CallbackOutcome{
		Accepted:           resp.GetAccepted(),
		Classification:     strings.TrimSpace(resp.GetClassification()),
		InteractionState:   strings.TrimSpace(resp.GetInteractionState()),
		ResumeRequired:     resp.GetResumeRequired(),
		ContinuationAction: strings.TrimSpace(resp.GetContinuationAction()),
	}, nil
}

func optionalStringValue(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
