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

func TestExpireInteractionsPollsUntilQueueIsEmpty(t *testing.T) {
	t.Parallel()

	interactions := &fakeInteractionLifecycleClient{
		expireResults: []ExpireNextInteractionResult{
			{Found: true, InteractionID: "interaction-1", InteractionState: "expired", ResumeRequired: true},
		},
	}

	svc := NewService(Config{
		InteractionExpiryLimit: 3,
	}, Dependencies{
		Interactions: interactions,
		Logger:       slog.New(slog.NewJSONHandler(io.Discard, nil)),
	})

	if err := svc.expireInteractions(context.Background()); err != nil {
		t.Fatalf("expireInteractions returned error: %v", err)
	}

	if interactions.expireCalls != 2 {
		t.Fatalf("expire calls = %d, want 2 (one processed item + one empty poll)", interactions.expireCalls)
	}
}

type fakeInteractionLifecycleClient struct {
	claims        []InteractionDispatchClaim
	claimCalls    int
	completed     []CompleteInteractionDispatchParams
	expireCalls   int
	expireResults []ExpireNextInteractionResult
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
