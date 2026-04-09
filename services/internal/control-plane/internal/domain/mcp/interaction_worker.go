package mcp

import (
	"context"
	"fmt"
	"strings"

	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
)

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

	envelopeJSON, err := s.buildInteractionDeliveryEnvelope(ctx, interactionDeliveryEnvelopeParams{
		Run:     run,
		Request: item.Interaction,
		Attempt: item.Attempt,
		Binding: item.Binding,
	})
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
		InteractionID:          strings.TrimSpace(params.InteractionID),
		DeliveryID:             strings.TrimSpace(params.DeliveryID),
		AdapterKind:            strings.TrimSpace(params.AdapterKind),
		Status:                 params.Status,
		RequestEnvelopeJSON:    params.RequestEnvelopeJSON,
		AckPayloadJSON:         params.AckPayloadJSON,
		AdapterDeliveryID:      strings.TrimSpace(params.AdapterDeliveryID),
		ProviderMessageRefJSON: params.ProviderMessageRefJSON,
		EditCapability:         params.EditCapability,
		Retryable:              params.Retryable,
		NextRetryAt:            params.NextRetryAt,
		LastErrorCode:          strings.TrimSpace(params.LastErrorCode),
		CallbackTokenKeyID:     strings.TrimSpace(params.CallbackTokenKeyID),
		CallbackTokenExpiresAt: params.CallbackTokenExpiresAt,
		FinishedAt:             params.FinishedAt,
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

func (s *Service) finalizeWorkerTerminalInteraction(ctx context.Context, interaction entitytypes.InteractionRequest, requireCurrentWait bool) (bool, error) {
	return s.finalizeInteractionResume(ctx, interaction, nil, requireCurrentWait)
}

func resolveInteractionResumeCorrelationID(interaction entitytypes.InteractionRequest, resumeRequired bool) string {
	if !resumeRequired {
		return ""
	}
	return buildInteractionResumeCorrelationID(interaction.ID)
}
