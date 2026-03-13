package grpc

import (
	"context"
	"testing"
	"time"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestSubmitInteractionCallback_RejectsTokenSubjectMismatch(t *testing.T) {
	t.Parallel()

	srv := &Server{
		mcp: fakeMCPRunTokenService{
			verifyInteractionCallbackToken: func(ctx context.Context, rawToken string, interactionID string) (mcpdomain.SessionContext, error) {
				return mcpdomain.SessionContext{}, status.Error(codes.Unauthenticated, "subject mismatch")
			},
		},
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer token-1"))
	_, err := srv.SubmitInteractionCallback(ctx, &controlplanev1.SubmitInteractionCallbackRequest{
		InteractionId:  "interaction-1",
		OccurredAt:     timestamppb.New(time.Now().UTC()),
		CallbackKind:   string(enumtypes.InteractionCallbackKindDecisionResponse),
		AdapterEventId: "event-1",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status, got %T", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %s", st.Code())
	}
	if st.Message() != "invalid interaction callback token" {
		t.Fatalf("unexpected message: %q", st.Message())
	}
}

func TestSubmitInteractionCallback_MapsAcceptedClassificationToApplied(t *testing.T) {
	t.Parallel()

	occurredAt := time.Date(2026, time.March, 13, 15, 4, 5, 0, time.UTC)
	var gotParams mcpdomain.SubmitInteractionCallbackParams

	srv := &Server{
		mcp: fakeMCPRunTokenService{
			verifyInteractionCallbackToken: func(ctx context.Context, rawToken string, interactionID string) (mcpdomain.SessionContext, error) {
				if rawToken != "token-1" {
					t.Fatalf("unexpected token %q", rawToken)
				}
				if interactionID != "interaction-1" {
					t.Fatalf("unexpected interactionID %q", interactionID)
				}
				return mcpdomain.SessionContext{
					RunID: "run-1",
				}, nil
			},
			submitInteractionCallback: func(ctx context.Context, params mcpdomain.SubmitInteractionCallbackParams) (mcpdomain.SubmitInteractionCallbackResult, error) {
				gotParams = params
				return mcpdomain.SubmitInteractionCallbackResult{
					Accepted:            true,
					Classification:      enumtypes.InteractionCallbackResultClassificationAccepted,
					InteractionState:    "resolved",
					ResumeRequired:      true,
					EffectiveResponseID: 55,
				}, nil
			},
		},
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer token-1"))
	resp, err := srv.SubmitInteractionCallback(ctx, &controlplanev1.SubmitInteractionCallbackRequest{
		InteractionId:    "interaction-1",
		DeliveryId:       stringPtr("delivery-1"),
		AdapterEventId:   "event-1",
		CallbackKind:     string(enumtypes.InteractionCallbackKindDecisionResponse),
		OccurredAt:       timestamppb.New(occurredAt),
		ResponseKind:     stringPtr("option"),
		SelectedOptionId: stringPtr("approve"),
		ResponderRef:     stringPtr("user-42"),
		RawPayloadJson:   []byte(`{"interaction_id":"interaction-1"}`),
	})
	if err != nil {
		t.Fatalf("SubmitInteractionCallback returned error: %v", err)
	}

	if resp.GetClassification() != "applied" {
		t.Fatalf("classification = %q, want applied", resp.GetClassification())
	}
	if !resp.GetAccepted() {
		t.Fatal("accepted = false, want true")
	}
	if !resp.GetResumeRequired() {
		t.Fatal("resume_required = false, want true")
	}
	if resp.GetEffectiveResponseId() != 55 {
		t.Fatalf("effective_response_id = %d, want 55", resp.GetEffectiveResponseId())
	}

	if gotParams.InteractionID != "interaction-1" {
		t.Fatalf("interaction_id = %q, want interaction-1", gotParams.InteractionID)
	}
	if gotParams.DeliveryID != "delivery-1" {
		t.Fatalf("delivery_id = %q, want delivery-1", gotParams.DeliveryID)
	}
	if gotParams.AdapterEventID != "event-1" {
		t.Fatalf("adapter_event_id = %q, want event-1", gotParams.AdapterEventID)
	}
	if gotParams.ResponseKind != enumtypes.InteractionResponseKindOption {
		t.Fatalf("response_kind = %q, want option", gotParams.ResponseKind)
	}
	if gotParams.SelectedOptionID != "approve" {
		t.Fatalf("selected_option_id = %q, want approve", gotParams.SelectedOptionID)
	}
	if gotParams.ResponderRef != "user-42" {
		t.Fatalf("responder_ref = %q, want user-42", gotParams.ResponderRef)
	}
	if !gotParams.OccurredAt.Equal(occurredAt) {
		t.Fatalf("occurred_at = %s, want %s", gotParams.OccurredAt, occurredAt)
	}
}

type fakeMCPRunTokenService struct {
	verifyRunToken                 func(ctx context.Context, rawToken string) (mcpdomain.SessionContext, error)
	verifyInteractionCallbackToken func(ctx context.Context, rawToken string, interactionID string) (mcpdomain.SessionContext, error)
	submitInteractionCallback      func(ctx context.Context, params mcpdomain.SubmitInteractionCallbackParams) (mcpdomain.SubmitInteractionCallbackResult, error)
}

func (f fakeMCPRunTokenService) IssueRunToken(ctx context.Context, params mcpdomain.IssueRunTokenParams) (mcpdomain.IssuedToken, error) {
	return mcpdomain.IssuedToken{}, nil
}

func (f fakeMCPRunTokenService) VerifyRunToken(ctx context.Context, rawToken string) (mcpdomain.SessionContext, error) {
	if f.verifyRunToken != nil {
		return f.verifyRunToken(ctx, rawToken)
	}
	return mcpdomain.SessionContext{}, nil
}

func (f fakeMCPRunTokenService) VerifyInteractionCallbackToken(ctx context.Context, rawToken string, interactionID string) (mcpdomain.SessionContext, error) {
	if f.verifyInteractionCallbackToken != nil {
		return f.verifyInteractionCallbackToken(ctx, rawToken, interactionID)
	}
	return mcpdomain.SessionContext{}, nil
}

func (f fakeMCPRunTokenService) ListPendingApprovals(ctx context.Context, limit int) ([]mcpdomain.ApprovalListItem, error) {
	return nil, nil
}

func (f fakeMCPRunTokenService) ResolveApproval(ctx context.Context, params mcpdomain.ResolveApprovalParams) (mcpdomain.ResolveApprovalResult, error) {
	return mcpdomain.ResolveApprovalResult{}, nil
}

func (f fakeMCPRunTokenService) ClaimNextInteractionDispatch(ctx context.Context, params mcpdomain.ClaimNextInteractionDispatchParams) (mcpdomain.InteractionDispatchClaim, bool, error) {
	return mcpdomain.InteractionDispatchClaim{}, false, nil
}

func (f fakeMCPRunTokenService) CompleteInteractionDispatch(ctx context.Context, params mcpdomain.CompleteInteractionDispatchParams) (mcpdomain.CompleteInteractionDispatchResult, error) {
	return mcpdomain.CompleteInteractionDispatchResult{}, nil
}

func (f fakeMCPRunTokenService) ExpireNextDueInteraction(ctx context.Context) (mcpdomain.ExpireNextInteractionResult, bool, error) {
	return mcpdomain.ExpireNextInteractionResult{}, false, nil
}

func (f fakeMCPRunTokenService) SubmitInteractionCallback(ctx context.Context, params mcpdomain.SubmitInteractionCallbackParams) (mcpdomain.SubmitInteractionCallbackResult, error) {
	if f.submitInteractionCallback != nil {
		return f.submitInteractionCallback(ctx, params)
	}
	return mcpdomain.SubmitInteractionCallbackResult{}, nil
}

func stringPtr(value string) *string {
	return &value
}
