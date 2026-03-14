package grpc

import (
	"context"
	"fmt"
	"strings"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) SubmitInteractionCallback(ctx context.Context, req *controlplanev1.SubmitInteractionCallbackRequest) (*controlplanev1.SubmitInteractionCallbackResponse, error) {
	if s.mcp == nil {
		return nil, status.Error(codes.FailedPrecondition, "mcp service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	interactionID := strings.TrimSpace(req.GetInteractionId())
	if interactionID == "" {
		return nil, status.Error(codes.InvalidArgument, "interaction_id is required")
	}
	if req.GetOccurredAt() == nil {
		return nil, status.Error(codes.InvalidArgument, "occurred_at is required")
	}
	if _, err := s.authenticateInteractionCallbackToken(ctx, interactionID); err != nil {
		return nil, err
	}

	callbackKind, err := parseInteractionCallbackKind(req.GetCallbackKind())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.mcp.SubmitInteractionCallback(ctx, mcpdomain.SubmitInteractionCallbackParams{
		InteractionID:           interactionID,
		DeliveryID:              strings.TrimSpace(req.GetDeliveryId()),
		AdapterEventID:          strings.TrimSpace(req.GetAdapterEventId()),
		CallbackKind:            callbackKind,
		OccurredAt:              tsToTime(req.GetOccurredAt()),
		CallbackHandle:          strings.TrimSpace(req.GetCallbackHandle()),
		FreeText:                strings.TrimSpace(req.GetFreeText()),
		ResponderRef:            strings.TrimSpace(req.GetResponderRef()),
		ProviderMessageRefJSON:  req.GetProviderMessageRefJson(),
		ProviderUpdateID:        strings.TrimSpace(req.GetProviderUpdateId()),
		ProviderCallbackQueryID: strings.TrimSpace(req.GetProviderCallbackQueryId()),
		DeliveryStatus:          strings.TrimSpace(req.GetDeliveryStatus()),
		TransportErrorCode:      strings.TrimSpace(req.GetTransportErrorCode()),
		TransportRetryable:      req.GetTransportRetryable(),
		RawPayloadJSON:          req.GetRawPayloadJson(),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &controlplanev1.SubmitInteractionCallbackResponse{
		Accepted:            result.Accepted,
		Classification:      transportInteractionCallbackClassification(result.Classification),
		InteractionState:    strings.TrimSpace(result.InteractionState),
		ResumeRequired:      result.ResumeRequired,
		ContinuationAction:  strings.TrimSpace(string(result.ContinuationAction)),
		EffectiveResponseId: result.EffectiveResponseID,
	}, nil
}

func parseInteractionCallbackKind(value string) (enumtypes.InteractionCallbackKind, error) {
	callbackKind := enumtypes.InteractionCallbackKind(strings.ToLower(strings.TrimSpace(value)))
	switch callbackKind {
	case enumtypes.InteractionCallbackKindDeliveryReceipt,
		enumtypes.InteractionCallbackKindOptionSelected,
		enumtypes.InteractionCallbackKindFreeTextReceived,
		enumtypes.InteractionCallbackKindTransportFailure:
		return callbackKind, nil
	default:
		return "", fmt.Errorf("callback_kind must be delivery_receipt|option_selected|free_text_received|transport_failure")
	}
}

func transportInteractionCallbackClassification(value enumtypes.InteractionCallbackResultClassification) string {
	if value == enumtypes.InteractionCallbackResultClassificationAccepted {
		return string(enumtypes.InteractionCallbackRecordClassificationApplied)
	}
	return strings.TrimSpace(string(value))
}
