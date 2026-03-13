package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ClaimNextInteractionDispatch(ctx context.Context, req *controlplanev1.ClaimNextInteractionDispatchRequest) (*controlplanev1.ClaimNextInteractionDispatchResponse, error) {
	if s.mcp == nil {
		return nil, status.Error(codes.FailedPrecondition, "mcp service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	item, found, err := s.mcp.ClaimNextInteractionDispatch(ctx, mcpdomain.ClaimNextInteractionDispatchParams{
		PendingAttemptTimeout: time.Duration(req.GetPendingAttemptTimeoutSeconds()) * time.Second,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	if !found {
		return &controlplanev1.ClaimNextInteractionDispatchResponse{Found: false}, nil
	}

	resp := &controlplanev1.ClaimNextInteractionDispatchResponse{
		Found:               true,
		CorrelationId:       strings.TrimSpace(item.CorrelationID),
		InteractionId:       strings.TrimSpace(item.Interaction.ID),
		RunId:               strings.TrimSpace(item.Interaction.RunID),
		InteractionKind:     strings.TrimSpace(string(item.Interaction.InteractionKind)),
		RecipientProvider:   strings.TrimSpace(item.Interaction.RecipientProvider),
		RecipientRef:        strings.TrimSpace(item.Interaction.RecipientRef),
		AttemptId:           item.Attempt.ID,
		AttemptNo:           int32(item.Attempt.AttemptNo),
		DeliveryId:          strings.TrimSpace(item.Attempt.DeliveryID),
		AdapterKind:         strings.TrimSpace(item.Attempt.AdapterKind),
		RequestEnvelopeJson: item.RequestEnvelopeJSON,
	}
	if item.Interaction.ResponseDeadlineAt != nil {
		resp.ResponseDeadlineAt = timestamppb.New(item.Interaction.ResponseDeadlineAt.UTC())
	}
	return resp, nil
}

func (s *Server) CompleteInteractionDispatch(ctx context.Context, req *controlplanev1.CompleteInteractionDispatchRequest) (*controlplanev1.CompleteInteractionDispatchResponse, error) {
	if s.mcp == nil {
		return nil, status.Error(codes.FailedPrecondition, "mcp service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if strings.TrimSpace(req.GetInteractionId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "interaction_id is required")
	}
	if strings.TrimSpace(req.GetDeliveryId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "delivery_id is required")
	}

	attemptStatus, err := parseInteractionAttemptStatus(req.GetStatus())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.mcp.CompleteInteractionDispatch(ctx, mcpdomain.CompleteInteractionDispatchParams{
		InteractionID:       strings.TrimSpace(req.GetInteractionId()),
		DeliveryID:          strings.TrimSpace(req.GetDeliveryId()),
		AdapterKind:         strings.TrimSpace(req.GetAdapterKind()),
		Status:              attemptStatus,
		RequestEnvelopeJSON: req.GetRequestEnvelopeJson(),
		AckPayloadJSON:      req.GetAckPayloadJson(),
		AdapterDeliveryID:   strings.TrimSpace(req.GetAdapterDeliveryId()),
		Retryable:           req.GetRetryable(),
		NextRetryAt:         optionalTime(req.GetNextRetryAt()),
		LastErrorCode:       strings.TrimSpace(req.GetLastErrorCode()),
		FinishedAt:          tsToTime(req.GetFinishedAt()),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &controlplanev1.CompleteInteractionDispatchResponse{
		InteractionId:       result.InteractionID,
		RunId:               result.RunID,
		InteractionState:    string(result.InteractionState),
		ResumeRequired:      result.ResumeRequired,
		ResumeCorrelationId: result.ResumeCorrelationID,
	}, nil
}

func (s *Server) ExpireNextInteraction(ctx context.Context, req *controlplanev1.ExpireNextInteractionRequest) (*controlplanev1.ExpireNextInteractionResponse, error) {
	if s.mcp == nil {
		return nil, status.Error(codes.FailedPrecondition, "mcp service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	result, found, err := s.mcp.ExpireNextDueInteraction(ctx)
	if err != nil {
		return nil, toStatus(err)
	}
	return &controlplanev1.ExpireNextInteractionResponse{
		Found:               found,
		InteractionId:       result.InteractionID,
		RunId:               result.RunID,
		InteractionState:    string(result.InteractionState),
		ResumeRequired:      result.ResumeRequired,
		ResumeCorrelationId: result.ResumeCorrelationID,
	}, nil
}

func parseInteractionAttemptStatus(value string) (enumtypes.InteractionDeliveryAttemptStatus, error) {
	statusValue := enumtypes.InteractionDeliveryAttemptStatus(strings.ToLower(strings.TrimSpace(value)))
	switch statusValue {
	case enumtypes.InteractionDeliveryAttemptStatusAccepted,
		enumtypes.InteractionDeliveryAttemptStatusFailed,
		enumtypes.InteractionDeliveryAttemptStatusExhausted:
		return statusValue, nil
	default:
		return "", fmt.Errorf("status must be accepted|failed|exhausted")
	}
}
