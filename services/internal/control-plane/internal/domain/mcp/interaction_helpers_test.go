package mcp

import (
	"testing"
	"time"

	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

func TestNormalizeUserNotifyInputRejectsActionURLWithoutLabel(t *testing.T) {
	t.Parallel()

	_, err := normalizeUserNotifyInput(UserNotifyInput{
		NotificationKind: UserNotificationKindStatusUpdate,
		Summary:          "Need your attention",
		ActionURL:        "https://example.com/runs/123",
	})
	if err == nil {
		t.Fatal("expected validation error for missing action_label")
	}
}

func TestNormalizeUserDecisionRequestInputRejectsDuplicateOptionIDs(t *testing.T) {
	t.Parallel()

	_, err := normalizeUserDecisionRequestInput(UserDecisionRequestInput{
		Question:           "Pick a path",
		ResponseTTLSeconds: 300,
		Options: []UserDecisionOption{
			{OptionID: "approve", Label: "Approve"},
			{OptionID: "approve", Label: "Approve again"},
		},
	})
	if err == nil {
		t.Fatal("expected validation error for duplicate option ids")
	}
}

func TestBuildInteractionResumePayloadResolvedOptionResponse(t *testing.T) {
	t.Parallel()

	request := entitytypes.InteractionRequest{
		ID:              "interaction-1",
		InteractionKind: enumtypes.InteractionKindDecisionRequest,
		State:           enumtypes.InteractionStateResolved,
		UpdatedAt:       time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC),
	}
	response := &entitytypes.InteractionResponseRecord{
		ResponseKind:     enumtypes.InteractionResponseKindOption,
		SelectedOptionID: "approve",
		RespondedAt:      time.Date(2026, time.March, 13, 12, 1, 0, 0, time.UTC),
	}

	payload := buildInteractionResumePayload(request, response)
	if payload == nil {
		t.Fatal("expected resume payload")
	}
	if payload.RequestStatus != enumtypes.InteractionRequestStatusAnswered {
		t.Fatalf("request status = %q, want %q", payload.RequestStatus, enumtypes.InteractionRequestStatusAnswered)
	}
	if payload.ResponseKind != enumtypes.InteractionResponseKindOption {
		t.Fatalf("response kind = %q, want %q", payload.ResponseKind, enumtypes.InteractionResponseKindOption)
	}
	if payload.SelectedOptionID != "approve" {
		t.Fatalf("selected option id = %q, want approve", payload.SelectedOptionID)
	}
}

func TestBuildInteractionResumePayloadExpired(t *testing.T) {
	t.Parallel()

	request := entitytypes.InteractionRequest{
		ID:              "interaction-2",
		InteractionKind: enumtypes.InteractionKindDecisionRequest,
		State:           enumtypes.InteractionStateExpired,
		UpdatedAt:       time.Date(2026, time.March, 13, 12, 5, 0, 0, time.UTC),
	}

	payload := buildInteractionResumePayload(request, nil)
	if payload == nil {
		t.Fatal("expected resume payload")
	}
	if payload.RequestStatus != enumtypes.InteractionRequestStatusExpired {
		t.Fatalf("request status = %q, want %q", payload.RequestStatus, enumtypes.InteractionRequestStatusExpired)
	}
	if payload.ResponseKind != enumtypes.InteractionResponseKindNone {
		t.Fatalf("response kind = %q, want %q", payload.ResponseKind, enumtypes.InteractionResponseKindNone)
	}
}
