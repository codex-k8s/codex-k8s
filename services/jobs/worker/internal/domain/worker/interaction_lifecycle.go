package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/flowevent"
)

const (
	interactionAttemptStatusAccepted  = "accepted"
	interactionAttemptStatusFailed    = "failed"
	interactionAttemptStatusExhausted = "exhausted"
)

type interactionDispatchAttemptedPayload struct {
	InteractionID string `json:"interaction_id"`
	AttemptNo     int    `json:"attempt_no"`
	DeliveryID    string `json:"delivery_id"`
	AdapterKind   string `json:"adapter_kind"`
	Status        string `json:"status"`
	ErrorCode     string `json:"error_code,omitempty"`
}

type interactionDispatchRetryScheduledPayload struct {
	InteractionID string `json:"interaction_id"`
	AttemptNo     int    `json:"attempt_no"`
	DeliveryID    string `json:"delivery_id"`
	AdapterKind   string `json:"adapter_kind"`
	NextRetryAt   string `json:"next_retry_at"`
	ErrorCode     string `json:"error_code,omitempty"`
}

type interactionDispatchFailurePayload struct {
	Accepted     bool   `json:"accepted"`
	Retryable    bool   `json:"retryable"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

func (s *Service) reconcileInteractions(ctx context.Context) error {
	if err := s.expireInteractions(ctx); err != nil {
		return fmt.Errorf("expire due interactions: %w", err)
	}
	if err := s.dispatchInteractions(ctx); err != nil {
		return fmt.Errorf("dispatch due interactions: %w", err)
	}
	return nil
}

func (s *Service) expireInteractions(ctx context.Context) error {
	for range s.cfg.InteractionExpiryLimit {
		result, err := s.interactions.ExpireNextInteraction(ctx)
		if err != nil {
			return err
		}
		if !result.Found {
			return nil
		}
		resumeScheduled := false
		if result.ResumeRequired {
			if err := s.scheduleInteractionResume(ctx, result.RunID, result.InteractionID, result.ResumeCorrelationID); err != nil {
				return err
			}
			resumeScheduled = true
		}
		s.logger.Info(
			"interaction expiry processed",
			"interaction_id",
			result.InteractionID,
			"interaction_state",
			result.InteractionState,
			"run_id",
			result.RunID,
			"resume_required",
			result.ResumeRequired,
			"resume_scheduled",
			resumeScheduled,
		)
	}
	return nil
}

func (s *Service) dispatchInteractions(ctx context.Context) error {
	for range s.cfg.InteractionDispatchLimit {
		claim, found, err := s.interactions.ClaimNextInteractionDispatch(ctx, s.cfg.InteractionPendingAttemptTimeout)
		if err != nil {
			return err
		}
		if !found {
			return nil
		}
		if err := s.dispatchInteraction(ctx, claim); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) dispatchInteraction(ctx context.Context, claim InteractionDispatchClaim) error {
	ack, dispatchErr := s.dispatcher.Dispatch(ctx, claim)
	completion := s.resolveInteractionDispatchCompletion(claim, ack, dispatchErr)

	result, err := s.interactions.CompleteInteractionDispatch(ctx, completion)
	if err != nil {
		return err
	}
	if err := s.insertInteractionDispatchAttemptedEvent(ctx, claim, completion); err != nil {
		return err
	}
	if err := s.insertInteractionRetryScheduledEvent(ctx, claim, completion); err != nil {
		return err
	}
	recordInteractionDispatchAttempt(completion.AdapterKind, completion.Status)
	if completion.Status == interactionAttemptStatusFailed && completion.NextRetryAt != nil {
		recordInteractionDispatchRetryScheduled(completion.AdapterKind, completion.LastErrorCode)
	}

	nextRetryAt := ""
	if completion.NextRetryAt != nil {
		nextRetryAt = completion.NextRetryAt.UTC().Format(time.RFC3339Nano)
	}
	s.logger.Info(
		"interaction dispatch completed",
		"interaction_id", claim.InteractionID,
		"run_id", claim.RunID,
		"attempt_no", claim.Attempt.AttemptNo,
		"delivery_id", claim.Attempt.DeliveryID,
		"adapter_kind", completion.AdapterKind,
		"status", completion.Status,
		"retryable", completion.Retryable,
		"error_code", completion.LastErrorCode,
		"next_retry_at", nextRetryAt,
		"interaction_state", result.InteractionState,
		"resume_required", result.ResumeRequired,
	)

	if dispatchErr != nil {
		s.logger.Warn("interaction dispatch failed", "interaction_id", claim.InteractionID, "delivery_id", claim.Attempt.DeliveryID, "err", dispatchErr)
	}
	if result.ResumeRequired {
		if err := s.scheduleInteractionResume(ctx, result.RunID, result.InteractionID, result.ResumeCorrelationID); err != nil {
			return err
		}
		s.logger.Info("interaction resume scheduled after terminal dispatch outcome", "interaction_id", result.InteractionID, "interaction_state", result.InteractionState, "run_id", result.RunID)
	}
	return nil
}

func (s *Service) resolveInteractionDispatchCompletion(claim InteractionDispatchClaim, ack InteractionDispatchAck, dispatchErr error) CompleteInteractionDispatchParams {
	now := s.now().UTC()
	params := CompleteInteractionDispatchParams{
		InteractionID:          claim.InteractionID,
		DeliveryID:             claim.Attempt.DeliveryID,
		AdapterKind:            resolveInteractionAdapterKind(claim, ack),
		RequestEnvelopeJSON:    claim.RequestEnvelopeJSON,
		AckPayloadJSON:         resolveInteractionAckPayload(ack, dispatchErr),
		AdapterDeliveryID:      ack.AdapterDeliveryID,
		ProviderMessageRefJSON: ack.ProviderMessageRefJSON,
		EditCapability:         strings.TrimSpace(ack.EditCapability),
		CallbackTokenKeyID:     strings.TrimSpace(ack.CallbackTokenKeyID),
		CallbackTokenExpiresAt: ack.CallbackTokenExpiresAt,
		FinishedAt:             now,
	}

	if dispatchErr == nil {
		params.Status = interactionAttemptStatusAccepted
		return params
	}

	params.LastErrorCode = resolveInteractionErrorCode(ack)
	if retryAt, ok := s.nextInteractionRetryAt(claim, ack, now); ok {
		params.Status = interactionAttemptStatusFailed
		params.Retryable = true
		params.NextRetryAt = retryAt
		return params
	}

	params.Status = interactionAttemptStatusExhausted
	return params
}

func (s *Service) nextInteractionRetryAt(claim InteractionDispatchClaim, ack InteractionDispatchAck, now time.Time) (*time.Time, bool) {
	if !ack.Retryable {
		return nil, false
	}
	if claim.Attempt.AttemptNo >= s.cfg.InteractionMaxAttempts {
		return nil, false
	}

	nextRetryAt := now.Add(s.interactionRetryBackoff(claim.Attempt.AttemptNo))
	if claim.ResponseDeadlineAt != nil {
		deadline := claim.ResponseDeadlineAt.UTC()
		if !deadline.IsZero() && !nextRetryAt.Before(deadline) {
			return nil, false
		}
	}
	result := nextRetryAt.UTC()
	return &result, true
}

func (s *Service) interactionRetryBackoff(attemptNo int) time.Duration {
	backoff := s.cfg.InteractionRetryBaseInterval
	if backoff <= 0 {
		return 0
	}
	for idx := 1; idx < attemptNo; idx++ {
		if backoff >= s.cfg.InteractionRetryMaxInterval/2 {
			return s.cfg.InteractionRetryMaxInterval
		}
		backoff *= 2
	}
	if backoff > s.cfg.InteractionRetryMaxInterval && s.cfg.InteractionRetryMaxInterval > 0 {
		return s.cfg.InteractionRetryMaxInterval
	}
	return backoff
}

func resolveInteractionAdapterKind(claim InteractionDispatchClaim, ack InteractionDispatchAck) string {
	switch {
	case ack.AdapterKind != "":
		return ack.AdapterKind
	case claim.Attempt.AdapterKind != "":
		return claim.Attempt.AdapterKind
	case claim.RecipientProvider != "":
		return claim.RecipientProvider
	default:
		return noopInteractionAdapterKind
	}
}

func resolveInteractionAckPayload(ack InteractionDispatchAck, dispatchErr error) json.RawMessage {
	if dispatchErr == nil {
		if len(ack.AckPayloadJSON) != 0 && json.Valid(ack.AckPayloadJSON) {
			return ack.AckPayloadJSON
		}
		return json.RawMessage(`{"accepted":true}`)
	}
	if len(ack.AckPayloadJSON) != 0 && json.Valid(ack.AckPayloadJSON) {
		return ack.AckPayloadJSON
	}
	return marshalJSONRaw(interactionDispatchFailurePayload{
		Accepted:     false,
		Retryable:    ack.Retryable,
		ErrorCode:    ack.ErrorCode,
		ErrorMessage: dispatchErr.Error(),
	})
}

func resolveInteractionErrorCode(ack InteractionDispatchAck) string {
	if ack.ErrorCode != "" {
		return ack.ErrorCode
	}
	return "dispatch_failed"
}

func (s *Service) insertInteractionDispatchAttemptedEvent(ctx context.Context, claim InteractionDispatchClaim, completion CompleteInteractionDispatchParams) error {
	return s.insertInteractionEvent(ctx, claim.CorrelationID, floweventdomain.EventTypeInteractionDispatchAttempted, interactionDispatchAttemptedPayload{
		InteractionID: claim.InteractionID,
		AttemptNo:     claim.Attempt.AttemptNo,
		DeliveryID:    claim.Attempt.DeliveryID,
		AdapterKind:   completion.AdapterKind,
		Status:        completion.Status,
		ErrorCode:     completion.LastErrorCode,
	})
}

func (s *Service) insertInteractionRetryScheduledEvent(ctx context.Context, claim InteractionDispatchClaim, completion CompleteInteractionDispatchParams) error {
	if completion.Status != interactionAttemptStatusFailed || completion.NextRetryAt == nil {
		return nil
	}
	return s.insertInteractionEvent(ctx, claim.CorrelationID, floweventdomain.EventTypeInteractionDispatchRetryScheduled, interactionDispatchRetryScheduledPayload{
		InteractionID: claim.InteractionID,
		AttemptNo:     claim.Attempt.AttemptNo,
		DeliveryID:    claim.Attempt.DeliveryID,
		AdapterKind:   completion.AdapterKind,
		NextRetryAt:   completion.NextRetryAt.UTC().Format(time.RFC3339Nano),
		ErrorCode:     completion.LastErrorCode,
	})
}

func (s *Service) insertInteractionEvent(ctx context.Context, correlationID string, eventType floweventdomain.EventType, payload any) error {
	if s.events == nil || correlationID == "" {
		return nil
	}
	return s.events.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: correlationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorIDWorker,
		EventType:     eventType,
		Payload:       marshalJSONRaw(payload),
		CreatedAt:     s.now().UTC(),
	})
}

func marshalJSONRaw(value any) json.RawMessage {
	raw, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return raw
}
