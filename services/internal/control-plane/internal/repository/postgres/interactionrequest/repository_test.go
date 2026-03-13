package interactionrequest

import (
	"encoding/json"
	"testing"
	"time"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func TestClassifyDecisionResponsePayloadAcceptsKnownOption(t *testing.T) {
	t.Parallel()

	requestPayload := json.RawMessage(`{"options":[{"option_id":"approve"},{"option_id":"reject"}]}`)
	decision, ok := classifyDecisionResponsePayload(requestPayload, querytypes.InteractionCallbackApplyParams{
		ResponseKind:     enumtypes.InteractionResponseKindOption,
		SelectedOptionID: "approve",
	})
	if !ok {
		t.Fatal("expected payload validation success")
	}
	if decision.responseKind != enumtypes.InteractionResponseKindOption {
		t.Fatalf("response kind = %q, want %q", decision.responseKind, enumtypes.InteractionResponseKindOption)
	}
	if decision.selectedOptionID != "approve" {
		t.Fatalf("selected option id = %q, want approve", decision.selectedOptionID)
	}
}

func TestClassifyDecisionResponsePayloadRejectsUnknownOption(t *testing.T) {
	t.Parallel()

	requestPayload := json.RawMessage(`{"options":[{"option_id":"approve"}]}`)
	_, ok := classifyDecisionResponsePayload(requestPayload, querytypes.InteractionCallbackApplyParams{
		ResponseKind:     enumtypes.InteractionResponseKindOption,
		SelectedOptionID: "reject",
	})
	if ok {
		t.Fatal("expected payload validation failure for unknown option")
	}
}

func TestClassifyDecisionResponsePayloadAcceptsFreeTextWhenEnabled(t *testing.T) {
	t.Parallel()

	requestPayload := json.RawMessage(`{"allow_free_text":true,"options":[{"option_id":"approve"}]}`)
	decision, ok := classifyDecisionResponsePayload(requestPayload, querytypes.InteractionCallbackApplyParams{
		ResponseKind: enumtypes.InteractionResponseKindFreeText,
		FreeText:     "ship it",
	})
	if !ok {
		t.Fatal("expected payload validation success for allowed free text")
	}
	if decision.responseKind != enumtypes.InteractionResponseKindFreeText {
		t.Fatalf("response kind = %q, want %q", decision.responseKind, enumtypes.InteractionResponseKindFreeText)
	}
	if decision.freeText != "ship it" {
		t.Fatalf("free text = %q, want %q", decision.freeText, "ship it")
	}
}

func TestClassifyDecisionResponsePayloadRejectsFreeTextWhenDisabled(t *testing.T) {
	t.Parallel()

	requestPayload := json.RawMessage(`{"allow_free_text":false,"options":[{"option_id":"approve"}]}`)
	_, ok := classifyDecisionResponsePayload(requestPayload, querytypes.InteractionCallbackApplyParams{
		ResponseKind: enumtypes.InteractionResponseKindFreeText,
		FreeText:     "ship it",
	})
	if ok {
		t.Fatal("expected payload validation failure when free text is disabled")
	}
}

func TestClassifyCallbackMarksExpiredPastDeadline(t *testing.T) {
	t.Parallel()

	deadline := time.Date(2026, time.March, 13, 11, 59, 0, 0, time.UTC)
	now := time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC)
	request := entitytypes.InteractionRequest{
		InteractionKind:    enumtypes.InteractionKindDecisionRequest,
		State:              enumtypes.InteractionStateOpen,
		ResolutionKind:     enumtypes.InteractionResolutionKindNone,
		RequestPayloadJSON: json.RawMessage(`{"options":[{"option_id":"approve"}]}`),
		ResponseDeadlineAt: &deadline,
	}

	decision := classifyCallback(request, querytypes.InteractionCallbackApplyParams{
		CallbackKind:     enumtypes.InteractionCallbackKindDecisionResponse,
		ResponseKind:     enumtypes.InteractionResponseKindOption,
		SelectedOptionID: "approve",
	}, now)

	if decision.resultClassification != enumtypes.InteractionCallbackResultClassificationExpired {
		t.Fatalf("classification = %q, want %q", decision.resultClassification, enumtypes.InteractionCallbackResultClassificationExpired)
	}
	if decision.nextState != enumtypes.InteractionStateExpired {
		t.Fatalf("next state = %q, want %q", decision.nextState, enumtypes.InteractionStateExpired)
	}
	if !decision.resumeRequired {
		t.Fatal("expected resumeRequired for expired callback")
	}
}

func TestClassifyCallbackMarksStaleForResolvedInteraction(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC)
	request := entitytypes.InteractionRequest{
		InteractionKind:    enumtypes.InteractionKindDecisionRequest,
		State:              enumtypes.InteractionStateResolved,
		ResolutionKind:     enumtypes.InteractionResolutionKindOptionSelected,
		RequestPayloadJSON: json.RawMessage(`{"options":[{"option_id":"approve"}]}`),
	}

	decision := classifyCallback(request, querytypes.InteractionCallbackApplyParams{
		CallbackKind:     enumtypes.InteractionCallbackKindDecisionResponse,
		ResponseKind:     enumtypes.InteractionResponseKindOption,
		SelectedOptionID: "approve",
	}, now)

	if decision.resultClassification != enumtypes.InteractionCallbackResultClassificationStale {
		t.Fatalf("classification = %q, want %q", decision.resultClassification, enumtypes.InteractionCallbackResultClassificationStale)
	}
	if decision.stateChanged {
		t.Fatal("expected stale callback to leave state unchanged")
	}
}

func TestClassifyDispatchCompletionAcceptsDecisionRequest(t *testing.T) {
	t.Parallel()

	request := entitytypes.InteractionRequest{
		InteractionKind: enumtypes.InteractionKindDecisionRequest,
		State:           enumtypes.InteractionStatePendingDispatch,
		ResolutionKind:  enumtypes.InteractionResolutionKindNone,
	}

	decision := classifyDispatchCompletion(request, enumtypes.InteractionDeliveryAttemptStatusAccepted)

	if decision.nextState != enumtypes.InteractionStateOpen {
		t.Fatalf("next state = %q, want %q", decision.nextState, enumtypes.InteractionStateOpen)
	}
	if !decision.stateChanged {
		t.Fatal("expected accepted decision request to mutate aggregate state")
	}
	if decision.resumeRequired {
		t.Fatal("did not expect resume to be required after accepted dispatch")
	}
}

func TestClassifyDispatchCompletionExhaustsDecisionRequest(t *testing.T) {
	t.Parallel()

	request := entitytypes.InteractionRequest{
		InteractionKind: enumtypes.InteractionKindDecisionRequest,
		State:           enumtypes.InteractionStatePendingDispatch,
		ResolutionKind:  enumtypes.InteractionResolutionKindNone,
	}

	decision := classifyDispatchCompletion(request, enumtypes.InteractionDeliveryAttemptStatusExhausted)

	if decision.nextState != enumtypes.InteractionStateDeliveryExhausted {
		t.Fatalf("next state = %q, want %q", decision.nextState, enumtypes.InteractionStateDeliveryExhausted)
	}
	if !decision.resumeRequired {
		t.Fatal("expected delivery exhausted decision request to require resume")
	}
}

func TestClassifyExpiryKeepsPendingDispatchAsDeliveryExhausted(t *testing.T) {
	t.Parallel()

	request := entitytypes.InteractionRequest{
		InteractionKind: enumtypes.InteractionKindDecisionRequest,
		State:           enumtypes.InteractionStatePendingDispatch,
		ResolutionKind:  enumtypes.InteractionResolutionKindNone,
	}

	decision := classifyExpiry(request)

	if decision.nextState != enumtypes.InteractionStateDeliveryExhausted {
		t.Fatalf("next state = %q, want %q", decision.nextState, enumtypes.InteractionStateDeliveryExhausted)
	}
	if !decision.resumeRequired {
		t.Fatal("expected pending dispatch expiry to require resume")
	}
}

func TestClassifyExpiryMarksOpenDecisionExpired(t *testing.T) {
	t.Parallel()

	request := entitytypes.InteractionRequest{
		InteractionKind: enumtypes.InteractionKindDecisionRequest,
		State:           enumtypes.InteractionStateOpen,
		ResolutionKind:  enumtypes.InteractionResolutionKindNone,
	}

	decision := classifyExpiry(request)

	if decision.nextState != enumtypes.InteractionStateExpired {
		t.Fatalf("next state = %q, want %q", decision.nextState, enumtypes.InteractionStateExpired)
	}
	if !decision.stateChanged {
		t.Fatal("expected open decision interaction expiry to mutate state")
	}
}
