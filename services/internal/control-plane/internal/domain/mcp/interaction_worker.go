package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

const (
	interactionDeliveryEnvelopeSchemaVersion = "v1"
	interactionCallbackPath                  = "/api/v1/mcp/interactions/callback"
	interactionDeliveryLocaleDefault         = "ru"
)

type interactionDeliveryEnvelope struct {
	SchemaVersion     string                    `json:"schema_version"`
	DeliveryID        string                    `json:"delivery_id"`
	InteractionID     string                    `json:"interaction_id"`
	InteractionKind   enumtypes.InteractionKind `json:"interaction_kind"`
	RecipientProvider string                    `json:"recipient_provider"`
	RecipientRef      string                    `json:"recipient_ref"`
	Locale            string                    `json:"locale,omitempty"`
	ContextLinks      interactionContextLinks   `json:"context_links"`
	Content           json.RawMessage           `json:"content"`
	CallbackURL       string                    `json:"callback_url"`
	CallbackBearer    string                    `json:"callback_bearer_token"`
	ExpiresAt         string                    `json:"expires_at,omitempty"`
}

// ClaimNextInteractionDispatch reserves or reclaims one due dispatch attempt for worker delivery.
func (s *Service) ClaimNextInteractionDispatch(ctx context.Context, params ClaimNextInteractionDispatchParams) (InteractionDispatchClaim, bool, error) {
	if s.interactions == nil {
		return InteractionDispatchClaim{}, false, fmt.Errorf("interaction repository is not configured")
	}

	item, found, err := s.interactions.ClaimNextDispatch(ctx, querytypes.InteractionDispatchClaimParams{
		Now:                   s.now().UTC(),
		PendingAttemptTimeout: params.PendingAttemptTimeout,
	})
	if err != nil {
		return InteractionDispatchClaim{}, false, err
	}
	if !found {
		return InteractionDispatchClaim{}, false, nil
	}

	run, found, err := s.runs.GetByID(ctx, item.Interaction.RunID)
	if err != nil {
		return InteractionDispatchClaim{}, false, fmt.Errorf("load run for interaction dispatch claim: %w", err)
	}
	if !found {
		return InteractionDispatchClaim{}, false, fmt.Errorf("run not found for interaction dispatch claim")
	}

	envelopeJSON, err := s.buildInteractionDeliveryEnvelope(run, item.Interaction, item.Attempt)
	if err != nil {
		return InteractionDispatchClaim{}, false, err
	}

	return InteractionDispatchClaim{
		CorrelationID:       run.CorrelationID,
		Interaction:         item.Interaction,
		Attempt:             item.Attempt,
		RequestEnvelopeJSON: envelopeJSON,
	}, true, nil
}

// CompleteInteractionDispatch stores one dispatch outcome and schedules resume when dispatch becomes terminal.
func (s *Service) CompleteInteractionDispatch(ctx context.Context, params CompleteInteractionDispatchParams) (CompleteInteractionDispatchResult, error) {
	if s.interactions == nil {
		return CompleteInteractionDispatchResult{}, fmt.Errorf("interaction repository is not configured")
	}

	result, err := s.interactions.CompleteDispatch(ctx, querytypes.InteractionDispatchCompleteParams{
		InteractionID:       strings.TrimSpace(params.InteractionID),
		DeliveryID:          strings.TrimSpace(params.DeliveryID),
		AdapterKind:         strings.TrimSpace(params.AdapterKind),
		Status:              params.Status,
		RequestEnvelopeJSON: params.RequestEnvelopeJSON,
		AckPayloadJSON:      params.AckPayloadJSON,
		AdapterDeliveryID:   strings.TrimSpace(params.AdapterDeliveryID),
		Retryable:           params.Retryable,
		NextRetryAt:         params.NextRetryAt,
		LastErrorCode:       strings.TrimSpace(params.LastErrorCode),
		FinishedAt:          params.FinishedAt,
	})
	if err != nil {
		return CompleteInteractionDispatchResult{}, err
	}

	resumeRequired, err := s.finalizeWorkerTerminalInteraction(ctx, result.Interaction, result.ResumeRequired)
	if err != nil {
		return CompleteInteractionDispatchResult{}, err
	}

	return CompleteInteractionDispatchResult{
		InteractionID:       result.Interaction.ID,
		RunID:               result.Interaction.RunID,
		InteractionState:    result.Interaction.State,
		ResumeRequired:      resumeRequired,
		ResumeCorrelationID: resolveInteractionResumeCorrelationID(result.Interaction, resumeRequired),
	}, nil
}

// ExpireNextDueInteraction marks one deadline-expired decision interaction terminal and schedules resume when needed.
func (s *Service) ExpireNextDueInteraction(ctx context.Context) (ExpireNextInteractionResult, bool, error) {
	if s.interactions == nil {
		return ExpireNextInteractionResult{}, false, fmt.Errorf("interaction repository is not configured")
	}

	result, found, err := s.interactions.ExpireNextDue(ctx, querytypes.InteractionExpireDueParams{Now: s.now().UTC()})
	if err != nil {
		return ExpireNextInteractionResult{}, false, err
	}
	if !found {
		return ExpireNextInteractionResult{}, false, nil
	}

	resumeRequired, err := s.finalizeWorkerTerminalInteraction(ctx, result.Interaction, result.ResumeRequired)
	if err != nil {
		return ExpireNextInteractionResult{}, false, err
	}

	return ExpireNextInteractionResult{
		InteractionID:       result.Interaction.ID,
		RunID:               result.Interaction.RunID,
		InteractionState:    result.Interaction.State,
		ResumeRequired:      resumeRequired,
		ResumeCorrelationID: resolveInteractionResumeCorrelationID(result.Interaction, resumeRequired),
	}, true, nil
}

func (s *Service) buildInteractionDeliveryEnvelope(run entitytypes.AgentRun, request entitytypes.InteractionRequest, attempt entitytypes.InteractionDeliveryAttempt) (json.RawMessage, error) {
	var contextLinks interactionContextLinks
	if len(request.ContextLinksJSON) > 0 {
		if err := json.Unmarshal(request.ContextLinksJSON, &contextLinks); err != nil {
			return nil, fmt.Errorf("unmarshal interaction context links: %w", err)
		}
	}

	callbackToken, err := s.issueInteractionCallbackToken(run, request)
	if err != nil {
		return nil, fmt.Errorf("issue interaction callback token: %w", err)
	}

	envelope := interactionDeliveryEnvelope{
		SchemaVersion:     interactionDeliveryEnvelopeSchemaVersion,
		DeliveryID:        attempt.DeliveryID,
		InteractionID:     request.ID,
		InteractionKind:   request.InteractionKind,
		RecipientProvider: request.RecipientProvider,
		RecipientRef:      request.RecipientRef,
		Locale:            interactionDeliveryLocaleDefault,
		ContextLinks:      contextLinks,
		Content:           jsonOrEmptyRawMessage(request.RequestPayloadJSON),
		CallbackURL:       strings.TrimRight(strings.TrimSpace(s.cfg.PublicBaseURL), "/") + interactionCallbackPath,
		CallbackBearer:    callbackToken,
	}
	if request.ResponseDeadlineAt != nil {
		envelope.ExpiresAt = request.ResponseDeadlineAt.UTC().Format(time.RFC3339Nano)
	}
	return marshalRawJSON(envelope), nil
}

func (s *Service) finalizeWorkerTerminalInteraction(ctx context.Context, interaction entitytypes.InteractionRequest, requireCurrentWait bool) (bool, error) {
	return s.finalizeInteractionResume(ctx, interaction, nil, requireCurrentWait)
}

func jsonOrEmptyRawMessage(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || !json.Valid(raw) {
		return json.RawMessage(`{}`)
	}
	return raw
}

func resolveInteractionResumeCorrelationID(interaction entitytypes.InteractionRequest, resumeRequired bool) string {
	if !resumeRequired {
		return ""
	}
	return buildInteractionResumeCorrelationID(interaction.ID)
}
