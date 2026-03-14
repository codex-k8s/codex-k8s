package runner

import (
	"fmt"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/mcp/userinteraction"
)

const (
	interactionRequestStatusAnswered          = "answered"
	interactionRequestStatusExpired           = "expired"
	interactionRequestStatusDeliveryExhausted = "delivery_exhausted"
	interactionRequestStatusCancelled         = "cancelled"

	interactionResponseKindOption   = "option"
	interactionResponseKindFreeText = "free_text"
	interactionResponseKindNone     = "none"
)

type interactionResumePayload struct {
	InteractionID    string `json:"interaction_id"`
	ToolName         string `json:"tool_name"`
	RequestStatus    string `json:"request_status"`
	ResponseKind     string `json:"response_kind"`
	SelectedOptionID string `json:"selected_option_id,omitempty"`
	FreeText         string `json:"free_text,omitempty"`
	ResolvedAt       string `json:"resolved_at"`
	ResolutionReason string `json:"resolution_reason"`
}

func buildInteractionResumePromptBlock(locale string, rawPayload string, resume bool) (string, error) {
	return buildDeterministicResumePromptBlock(
		locale,
		rawPayload,
		resume,
		"interaction resume payload requires restored codex session",
		parseInteractionResumePayload,
		"Детерминированный resume context:",
		"Ниже machine-readable terminal outcome предыдущего user interaction. Используйте этот JSON как authoritative source для продолжения сессии, не делайте повторный lookup у adapter и не задавайте пользователю тот же вопрос повторно.",
		"Deterministic resume context:",
		"Below is the machine-readable terminal outcome of the previous user interaction. Treat this JSON as the authoritative source for continuation, do not re-query the adapter, and do not ask the same question again.",
	)
}

func parseInteractionResumePayload(rawPayload string) (interactionResumePayload, error) {
	return parseDeterministicResumePayload(
		rawPayload,
		userinteraction.ResumePayloadMaxBytes,
		"interaction resume payload",
		normalizeInteractionResumePayload,
		validateInteractionResumePayload,
	)
}

func normalizeInteractionResumePayload(payload interactionResumePayload) interactionResumePayload {
	trimStringFields(
		&payload.InteractionID,
		&payload.ToolName,
		&payload.RequestStatus,
		&payload.ResponseKind,
		&payload.SelectedOptionID,
		&payload.FreeText,
		&payload.ResolvedAt,
		&payload.ResolutionReason,
	)
	return payload
}

func validateInteractionResumePayload(payload interactionResumePayload) error {
	if payload.InteractionID == "" {
		return fmt.Errorf("interaction resume payload: interaction_id is required")
	}
	if payload.ToolName == "" {
		return fmt.Errorf("interaction resume payload: tool_name is required")
	}
	if payload.RequestStatus == "" {
		return fmt.Errorf("interaction resume payload: request_status is required")
	}
	switch payload.RequestStatus {
	case interactionRequestStatusAnswered,
		interactionRequestStatusExpired,
		interactionRequestStatusDeliveryExhausted,
		interactionRequestStatusCancelled:
	default:
		return fmt.Errorf("interaction resume payload: request_status %q is not supported", payload.RequestStatus)
	}

	switch payload.ResponseKind {
	case interactionResponseKindOption:
		if payload.SelectedOptionID == "" {
			return fmt.Errorf("interaction resume payload: selected_option_id is required for option responses")
		}
	case interactionResponseKindFreeText:
		if payload.FreeText == "" {
			return fmt.Errorf("interaction resume payload: free_text is required for free_text responses")
		}
	case interactionResponseKindNone:
	default:
		return fmt.Errorf("interaction resume payload: response_kind %q is not supported", payload.ResponseKind)
	}

	if payload.ResolvedAt == "" {
		return fmt.Errorf("interaction resume payload: resolved_at is required")
	}
	if _, err := time.Parse(time.RFC3339Nano, payload.ResolvedAt); err != nil {
		return fmt.Errorf("interaction resume payload: resolved_at must be RFC3339: %w", err)
	}
	if payload.ResolutionReason == "" {
		return fmt.Errorf("interaction resume payload: resolution_reason is required")
	}
	return nil
}
