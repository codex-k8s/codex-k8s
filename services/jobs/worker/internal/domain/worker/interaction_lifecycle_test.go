package worker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
)

func TestDispatchInteractionsSchedulesRetryableFailure(t *testing.T) {
	t.Parallel()

	events := &fakeFlowEvents{}
	runs := &fakeRunQueue{}
	interactions := &fakeInteractionLifecycleClient{
		claims: []InteractionDispatchClaim{
			{
				CorrelationID:      "corr-1",
				InteractionID:      "interaction-1",
				InteractionKind:    "decision_request",
				ResponseDeadlineAt: timePtr(time.Date(2026, 3, 13, 12, 5, 0, 0, time.UTC)),
				Attempt: InteractionDispatchAttempt{
					ID:         11,
					AttemptNo:  1,
					DeliveryID: "delivery-1",
				},
				RequestEnvelopeJSON: []byte(`{"delivery_id":"delivery-1"}`),
			},
		},
	}
	dispatcher := fakeInteractionDispatcher{
		ack: InteractionDispatchAck{
			AdapterKind: "noop",
			Retryable:   true,
			ErrorCode:   "transport_unavailable",
		},
		err: errors.New("temporary transport outage"),
	}

	svc := NewService(Config{
		InteractionDispatchLimit:         2,
		InteractionRetryBaseInterval:     30 * time.Second,
		InteractionRetryMaxInterval:      5 * time.Minute,
		InteractionMaxAttempts:           3,
		InteractionPendingAttemptTimeout: time.Minute,
	}, Dependencies{
		Runs:                  runs,
		Events:                events,
		Interactions:          interactions,
		InteractionDispatcher: dispatcher,
		Logger:                slog.New(slog.NewJSONHandler(io.Discard, nil)),
	})
	svc.now = func() time.Time { return time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC) }

	if err := svc.dispatchInteractions(context.Background()); err != nil {
		t.Fatalf("dispatchInteractions returned error: %v", err)
	}

	if interactions.claimCalls != 2 {
		t.Fatalf("claim calls = %d, want 2 (one claim + one empty poll)", interactions.claimCalls)
	}
	if len(interactions.completed) != 1 {
		t.Fatalf("completed attempts = %d, want 1", len(interactions.completed))
	}
	completed := interactions.completed[0]
	if completed.Status != interactionAttemptStatusFailed {
		t.Fatalf("status = %q, want %q", completed.Status, interactionAttemptStatusFailed)
	}
	if completed.NextRetryAt == nil {
		t.Fatal("expected next_retry_at for retryable failure")
	}
	if got, want := completed.NextRetryAt.UTC().Format(time.RFC3339), "2026-03-13T12:00:30Z"; got != want {
		t.Fatalf("next_retry_at = %q, want %q", got, want)
	}
	if len(events.inserted) != 2 {
		t.Fatalf("events inserted = %d, want 2", len(events.inserted))
	}
	if events.inserted[0].EventType != floweventdomain.EventTypeInteractionDispatchAttempted {
		t.Fatalf("first event = %q, want %q", events.inserted[0].EventType, floweventdomain.EventTypeInteractionDispatchAttempted)
	}
	if events.inserted[1].EventType != floweventdomain.EventTypeInteractionDispatchRetryScheduled {
		t.Fatalf("second event = %q, want %q", events.inserted[1].EventType, floweventdomain.EventTypeInteractionDispatchRetryScheduled)
	}
}

func TestDispatchInteractionsExhaustsWhenRetryBudgetIsUsed(t *testing.T) {
	t.Parallel()

	events := &fakeFlowEvents{}
	runs := &fakeRunQueue{}
	interactions := &fakeInteractionLifecycleClient{
		claims: []InteractionDispatchClaim{
			{
				CorrelationID:      "corr-2",
				InteractionID:      "interaction-2",
				InteractionKind:    "decision_request",
				ResponseDeadlineAt: timePtr(time.Date(2026, 3, 13, 12, 5, 0, 0, time.UTC)),
				Attempt: InteractionDispatchAttempt{
					ID:         12,
					AttemptNo:  3,
					DeliveryID: "delivery-2",
				},
				RequestEnvelopeJSON: []byte(`{"delivery_id":"delivery-2"}`),
			},
		},
	}
	dispatcher := fakeInteractionDispatcher{
		ack: InteractionDispatchAck{
			AdapterKind: "noop",
			Retryable:   true,
			ErrorCode:   "transport_unavailable",
		},
		err: errors.New("temporary transport outage"),
	}

	svc := NewService(Config{
		InteractionDispatchLimit:         1,
		InteractionRetryBaseInterval:     30 * time.Second,
		InteractionRetryMaxInterval:      5 * time.Minute,
		InteractionMaxAttempts:           3,
		InteractionPendingAttemptTimeout: time.Minute,
	}, Dependencies{
		Runs:                  runs,
		Events:                events,
		Interactions:          interactions,
		InteractionDispatcher: dispatcher,
		Logger:                slog.New(slog.NewJSONHandler(io.Discard, nil)),
	})
	svc.now = func() time.Time { return time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC) }

	if err := svc.dispatchInteractions(context.Background()); err != nil {
		t.Fatalf("dispatchInteractions returned error: %v", err)
	}

	if len(interactions.completed) != 1 {
		t.Fatalf("completed attempts = %d, want 1", len(interactions.completed))
	}
	if got, want := interactions.completed[0].Status, interactionAttemptStatusExhausted; got != want {
		t.Fatalf("status = %q, want %q", got, want)
	}
	if interactions.completed[0].NextRetryAt != nil {
		t.Fatal("did not expect next_retry_at for exhausted attempt")
	}
	if len(events.inserted) != 1 {
		t.Fatalf("events inserted = %d, want 1", len(events.inserted))
	}
}

func TestDispatchInteractionsSchedulesResumeAfterTerminalOutcome(t *testing.T) {
	t.Parallel()

	events := &fakeFlowEvents{}
	runs := &fakeRunQueue{}
	interactions := &fakeInteractionLifecycleClient{
		claims: []InteractionDispatchClaim{
			{
				CorrelationID:   "corr-3",
				InteractionID:   "interaction-3",
				RunID:           "run-3",
				InteractionKind: "decision_request",
				Attempt: InteractionDispatchAttempt{
					ID:         13,
					AttemptNo:  2,
					DeliveryID: "delivery-3",
				},
				RequestEnvelopeJSON: []byte(`{"delivery_id":"delivery-3"}`),
			},
		},
		completeResult: CompleteInteractionDispatchResult{
			InteractionID:       "interaction-3",
			RunID:               "run-3",
			InteractionState:    "delivery_exhausted",
			ResumeRequired:      true,
			ResumeCorrelationID: "interaction-resume:interaction-3",
		},
	}
	dispatcher := fakeInteractionDispatcher{
		ack: InteractionDispatchAck{
			AdapterKind: "noop",
			ErrorCode:   "transport_unavailable",
		},
		err: errors.New("final transport failure"),
	}

	svc := NewService(Config{
		InteractionDispatchLimit:         1,
		InteractionRetryBaseInterval:     30 * time.Second,
		InteractionRetryMaxInterval:      5 * time.Minute,
		InteractionMaxAttempts:           1,
		InteractionPendingAttemptTimeout: time.Minute,
	}, Dependencies{
		Runs:                  runs,
		Events:                events,
		Interactions:          interactions,
		InteractionDispatcher: dispatcher,
		Logger:                slog.New(slog.NewJSONHandler(io.Discard, nil)),
	})
	svc.now = func() time.Time { return time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC) }

	if err := svc.dispatchInteractions(context.Background()); err != nil {
		t.Fatalf("dispatchInteractions returned error: %v", err)
	}

	if len(runs.resumePending) != 1 {
		t.Fatalf("resume pending calls = %d, want 1", len(runs.resumePending))
	}
	if got, want := runs.resumePending[0].SourceRunID, "run-3"; got != want {
		t.Fatalf("source run id = %q, want %q", got, want)
	}
	if got, want := runs.resumePending[0].CorrelationID, "interaction-resume:interaction-3"; got != want {
		t.Fatalf("correlation id = %q, want %q", got, want)
	}
}

func TestExpireInteractionsPollsUntilQueueIsEmpty(t *testing.T) {
	t.Parallel()

	runs := &fakeRunQueue{}
	interactions := &fakeInteractionLifecycleClient{
		expireResults: []ExpireNextInteractionResult{
			{
				Found:               true,
				InteractionID:       "interaction-1",
				RunID:               "run-1",
				InteractionState:    "expired",
				ResumeRequired:      true,
				ResumeCorrelationID: "interaction-resume:interaction-1",
			},
		},
	}

	svc := NewService(Config{
		InteractionExpiryLimit: 3,
	}, Dependencies{
		Runs:         runs,
		Interactions: interactions,
		Logger:       slog.New(slog.NewJSONHandler(io.Discard, nil)),
	})

	if err := svc.expireInteractions(context.Background()); err != nil {
		t.Fatalf("expireInteractions returned error: %v", err)
	}

	if interactions.expireCalls != 2 {
		t.Fatalf("expire calls = %d, want 2 (one processed item + one empty poll)", interactions.expireCalls)
	}
	if len(runs.resumePending) != 1 {
		t.Fatalf("resume pending calls = %d, want 1", len(runs.resumePending))
	}
	if got, want := runs.resumePending[0].SourceRunID, "run-1"; got != want {
		t.Fatalf("source run id = %q, want %q", got, want)
	}
	if got, want := runs.resumePending[0].CorrelationID, "interaction-resume:interaction-1"; got != want {
		t.Fatalf("correlation id = %q, want %q", got, want)
	}
}

type fakeInteractionLifecycleClient struct {
	claims         []InteractionDispatchClaim
	claimCalls     int
	completed      []CompleteInteractionDispatchParams
	completeResult CompleteInteractionDispatchResult
	expireCalls    int
	expireResults  []ExpireNextInteractionResult
}

func (f *fakeInteractionLifecycleClient) ClaimNextInteractionDispatch(_ context.Context, _ time.Duration) (InteractionDispatchClaim, bool, error) {
	f.claimCalls++
	if f.claimCalls > len(f.claims) {
		return InteractionDispatchClaim{}, false, nil
	}
	return f.claims[f.claimCalls-1], true, nil
}

func (f *fakeInteractionLifecycleClient) CompleteInteractionDispatch(_ context.Context, params CompleteInteractionDispatchParams) (CompleteInteractionDispatchResult, error) {
	f.completed = append(f.completed, params)
	if f.completeResult.InteractionID != "" || f.completeResult.RunID != "" || f.completeResult.ResumeRequired || f.completeResult.ResumeCorrelationID != "" {
		return f.completeResult, nil
	}
	return CompleteInteractionDispatchResult{
		InteractionID:    params.InteractionID,
		InteractionState: params.Status,
	}, nil
}

func (f *fakeInteractionLifecycleClient) ExpireNextInteraction(_ context.Context) (ExpireNextInteractionResult, error) {
	f.expireCalls++
	if f.expireCalls > len(f.expireResults) {
		return ExpireNextInteractionResult{}, nil
	}
	return f.expireResults[f.expireCalls-1], nil
}

type fakeInteractionDispatcher struct {
	ack InteractionDispatchAck
	err error
}

func (f fakeInteractionDispatcher) Dispatch(context.Context, InteractionDispatchClaim) (InteractionDispatchAck, error) {
	return f.ack, f.err
}
