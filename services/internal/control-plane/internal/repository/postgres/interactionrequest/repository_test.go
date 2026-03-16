package interactionrequest

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/mcp/userinteraction"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func TestClassifyDecisionResponsePayloadAcceptsKnownOption(t *testing.T) {
	t.Parallel()

	request := entitytypes.InteractionRequest{
		ID:                 "interaction-1",
		RequestPayloadJSON: json.RawMessage(`{"options":[{"option_id":"approve"},{"option_id":"reject"}]}`),
	}
	handle := &entitytypes.InteractionCallbackHandle{
		InteractionID:    "interaction-1",
		ChannelBindingID: 11,
		HandleKind:       enumtypes.InteractionCallbackHandleKindOption,
		OptionID:         "approve",
	}
	decision, ok := classifyDecisionResponsePayload(request, handle, querytypes.InteractionCallbackApplyParams{
		CallbackKind: enumtypes.InteractionCallbackKindOptionSelected,
		OccurredAt:   time.Date(2026, time.March, 13, 16, 5, 0, 0, time.UTC),
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

func TestCreateDeliveryAttemptSQLHasMatchingTargetColumnsAndValues(t *testing.T) {
	t.Parallel()

	re := regexp.MustCompile(`(?is)INSERT\s+INTO\s+interaction_delivery_attempts\s*\((.*?)\)\s*VALUES\s*\((.*?)\)`)
	matches := re.FindStringSubmatch(queryCreateDeliveryAttempt)
	if len(matches) != 3 {
		t.Fatalf("expected INSERT ... VALUES structure in create_delivery_attempt.sql")
	}

	columnCount := countCommaSeparatedSQLItems(matches[1])
	valueCount := countCommaSeparatedSQLItems(matches[2])
	if columnCount != valueCount {
		t.Fatalf("target columns count = %d, values count = %d", columnCount, valueCount)
	}
}

func countCommaSeparatedSQLItems(source string) int {
	items := strings.Split(source, ",")
	count := 0
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		count++
	}
	return count
}

func TestClassifyDecisionResponsePayloadRejectsUnknownOption(t *testing.T) {
	t.Parallel()

	request := entitytypes.InteractionRequest{
		ID:                 "interaction-1",
		RequestPayloadJSON: json.RawMessage(`{"options":[{"option_id":"approve"}]}`),
	}
	handle := &entitytypes.InteractionCallbackHandle{
		InteractionID:    "interaction-1",
		ChannelBindingID: 11,
		HandleKind:       enumtypes.InteractionCallbackHandleKindOption,
	}
	_, ok := classifyDecisionResponsePayload(request, handle, querytypes.InteractionCallbackApplyParams{
		CallbackKind: enumtypes.InteractionCallbackKindOptionSelected,
		OccurredAt:   time.Date(2026, time.March, 13, 16, 5, 0, 0, time.UTC),
	})
	if ok {
		t.Fatal("expected payload validation failure for unknown option")
	}
}

func TestClassifyDecisionResponsePayloadAcceptsFreeTextWhenEnabled(t *testing.T) {
	t.Parallel()

	request := entitytypes.InteractionRequest{
		ID:                 "interaction-1",
		RequestPayloadJSON: json.RawMessage(`{"allow_free_text":true,"options":[{"option_id":"approve"}]}`),
	}
	handle := &entitytypes.InteractionCallbackHandle{
		InteractionID:    "interaction-1",
		ChannelBindingID: 11,
		HandleKind:       enumtypes.InteractionCallbackHandleKindFreeTextSession,
	}
	decision, ok := classifyDecisionResponsePayload(request, handle, querytypes.InteractionCallbackApplyParams{
		CallbackKind: enumtypes.InteractionCallbackKindFreeTextReceived,
		FreeText:     "ship it",
		OccurredAt:   time.Date(2026, time.March, 13, 16, 5, 0, 0, time.UTC),
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

	request := entitytypes.InteractionRequest{
		ID:                 "interaction-1",
		RequestPayloadJSON: json.RawMessage(`{"allow_free_text":false,"options":[{"option_id":"approve"}]}`),
	}
	handle := &entitytypes.InteractionCallbackHandle{
		InteractionID:    "interaction-1",
		ChannelBindingID: 11,
		HandleKind:       enumtypes.InteractionCallbackHandleKindFreeTextSession,
	}
	_, ok := classifyDecisionResponsePayload(request, handle, querytypes.InteractionCallbackApplyParams{
		CallbackKind: enumtypes.InteractionCallbackKindFreeTextReceived,
		FreeText:     "ship it",
		OccurredAt:   time.Date(2026, time.March, 13, 16, 5, 0, 0, time.UTC),
	})
	if ok {
		t.Fatal("expected payload validation failure when free text is disabled")
	}
}

func TestClassifyDecisionResponsePayloadRejectsOversizedFreeText(t *testing.T) {
	t.Parallel()

	request := entitytypes.InteractionRequest{
		ID:                 "interaction-1",
		RequestPayloadJSON: json.RawMessage(`{"allow_free_text":true,"options":[{"option_id":"approve"}]}`),
	}
	handle := &entitytypes.InteractionCallbackHandle{
		InteractionID:    "interaction-1",
		ChannelBindingID: 11,
		HandleKind:       enumtypes.InteractionCallbackHandleKindFreeTextSession,
	}
	_, ok := classifyDecisionResponsePayload(request, handle, querytypes.InteractionCallbackApplyParams{
		CallbackKind: enumtypes.InteractionCallbackKindFreeTextReceived,
		FreeText:     strings.Repeat("a", userinteraction.DecisionResponseFreeTextMaxBytes+1),
		OccurredAt:   time.Date(2026, time.March, 13, 16, 5, 0, 0, time.UTC),
	})
	if ok {
		t.Fatal("expected payload validation failure for oversized free text")
	}
}

func TestClassifyDecisionResponsePayloadRejectsOversizedOption(t *testing.T) {
	t.Parallel()

	oversizedOptionID := strings.Repeat("a", userinteraction.ResumePayloadMaxBytes)
	request := entitytypes.InteractionRequest{
		ID:                 "interaction-1",
		RequestPayloadJSON: json.RawMessage(`{"options":[{"option_id":"` + oversizedOptionID + `"}]}`),
	}
	handle := &entitytypes.InteractionCallbackHandle{
		InteractionID:    "interaction-1",
		ChannelBindingID: 11,
		HandleKind:       enumtypes.InteractionCallbackHandleKindOption,
		OptionID:         oversizedOptionID,
	}
	_, ok := classifyDecisionResponsePayload(request, handle, querytypes.InteractionCallbackApplyParams{
		CallbackKind: enumtypes.InteractionCallbackKindOptionSelected,
		OccurredAt:   time.Date(2026, time.March, 13, 16, 5, 0, 0, time.UTC),
	})
	if ok {
		t.Fatal("expected payload validation failure for oversized option response")
	}
}

func TestClassifyCallbackMarksExpiredPastDeadline(t *testing.T) {
	t.Parallel()

	deadline := time.Date(2026, time.March, 13, 11, 59, 0, 0, time.UTC)
	now := time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC)
	request := entitytypes.InteractionRequest{
		ID:                 "interaction-1",
		InteractionKind:    enumtypes.InteractionKindDecisionRequest,
		State:              enumtypes.InteractionStateOpen,
		ResolutionKind:     enumtypes.InteractionResolutionKindNone,
		RequestPayloadJSON: json.RawMessage(`{"options":[{"option_id":"approve"}]}`),
		ResponseDeadlineAt: &deadline,
	}
	binding := &entitytypes.InteractionChannelBinding{ID: 11}
	handle := &entitytypes.InteractionCallbackHandle{
		InteractionID:      "interaction-1",
		ChannelBindingID:   11,
		HandleKind:         enumtypes.InteractionCallbackHandleKindOption,
		OptionID:           "approve",
		ResponseDeadlineAt: deadline,
		GraceExpiresAt:     deadline.Add(24 * time.Hour),
	}

	decision := classifyCallback(request, binding, handle, []byte("hash"), querytypes.InteractionCallbackApplyParams{
		CallbackKind:   enumtypes.InteractionCallbackKindOptionSelected,
		CallbackHandle: "raw-handle",
		OccurredAt:     now,
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
		ID:                 "interaction-1",
		InteractionKind:    enumtypes.InteractionKindDecisionRequest,
		State:              enumtypes.InteractionStateResolved,
		ResolutionKind:     enumtypes.InteractionResolutionKindOptionSelected,
		RequestPayloadJSON: json.RawMessage(`{"options":[{"option_id":"approve"}]}`),
	}
	binding := &entitytypes.InteractionChannelBinding{ID: 11}
	handle := &entitytypes.InteractionCallbackHandle{
		InteractionID:      "interaction-1",
		ChannelBindingID:   11,
		HandleKind:         enumtypes.InteractionCallbackHandleKindOption,
		OptionID:           "approve",
		ResponseDeadlineAt: now.Add(time.Minute),
		GraceExpiresAt:     now.Add(24 * time.Hour),
	}

	decision := classifyCallback(request, binding, handle, []byte("hash"), querytypes.InteractionCallbackApplyParams{
		CallbackKind:   enumtypes.InteractionCallbackKindOptionSelected,
		CallbackHandle: "raw-handle",
		OccurredAt:     now,
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

func TestResolveClaimedDeliveryAttemptFallsBackToFollowUpWithoutProviderMessageRef(t *testing.T) {
	t.Parallel()

	role, reason := resolveClaimedDeliveryAttempt(&entitytypes.InteractionChannelBinding{
		ContinuationState: enumtypes.InteractionContinuationStateReadyForEdit,
		EditCapability:    enumtypes.InteractionEditCapabilityEditable,
	}, entitytypes.InteractionDeliveryAttempt{}, false)

	if got, want := role, enumtypes.InteractionDeliveryRoleFollowUpNotify; got != want {
		t.Fatalf("delivery role = %q, want %q", got, want)
	}
	if got, want := reason, "applied_response"; got != want {
		t.Fatalf("continuation reason = %q, want %q", got, want)
	}
}

func TestClassifyContinuationDispatchCompletionSchedulesFollowUpAfterEditFailure(t *testing.T) {
	t.Parallel()

	decision := classifyContinuationDispatchCompletion(entitytypes.InteractionChannelBinding{
		ContinuationState: enumtypes.InteractionContinuationStateReadyForEdit,
	}, entitytypes.InteractionDeliveryAttempt{
		DeliveryRole: enumtypes.InteractionDeliveryRoleMessageEdit,
		Status:       enumtypes.InteractionDeliveryAttemptStatusExhausted,
	})

	if !decision.updateBinding {
		t.Fatal("expected binding projection update")
	}
	if got, want := decision.continuationState, enumtypes.InteractionContinuationStateFollowUpRequired; got != want {
		t.Fatalf("continuation_state = %q, want %q", got, want)
	}
	if decision.updateRequestProjection {
		t.Fatal("did not expect request projection update while scheduling follow-up")
	}
}

func TestClassifyContinuationDispatchCompletionMarksManualFallbackAfterFollowUpFailure(t *testing.T) {
	t.Parallel()

	decision := classifyContinuationDispatchCompletion(entitytypes.InteractionChannelBinding{
		ContinuationState: enumtypes.InteractionContinuationStateFollowUpRequired,
	}, entitytypes.InteractionDeliveryAttempt{
		DeliveryRole: enumtypes.InteractionDeliveryRoleFollowUpNotify,
		Status:       enumtypes.InteractionDeliveryAttemptStatusExhausted,
	})

	if got, want := decision.continuationState, enumtypes.InteractionContinuationStateManualFallbackRequired; got != want {
		t.Fatalf("continuation_state = %q, want %q", got, want)
	}
	if !decision.updateRequestProjection {
		t.Fatal("expected request projection update for follow-up failure")
	}
	if got, want := decision.operatorState, enumtypes.InteractionOperatorStateManualFallbackRequired; got != want {
		t.Fatalf("operator_state = %q, want %q", got, want)
	}
	if got, want := decision.operatorSignalCode, enumtypes.InteractionOperatorSignalCodeFollowUpFailed; got != want {
		t.Fatalf("operator_signal_code = %q, want %q", got, want)
	}
}
