package mcp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	interactionrequestrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/interactionrequest"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

const (
	interactionDeliveryEnvelopeSchemaVersion          = "telegram-interaction-v1"
	interactionCallbackPath                           = "/api/v1/mcp/interactions/callback"
	interactionDeliveryLocaleDefault                  = "ru"
	interactionContinuationPreferredModeEditInPlace   = "edit_in_place_first"
	interactionCallbackHandlePrefix                   = "tgh1_"
	interactionFreeTextReplyInstruction               = "Если ни один вариант не подходит, ответьте сообщением в этот чат."
	interactionCallbackHandleRandomBytes        int   = 20
)

type interactionDeliveryEnvelope struct {
	SchemaVersion      string                            `json:"schema_version"`
	DeliveryID         string                            `json:"delivery_id"`
	InteractionID      string                            `json:"interaction_id"`
	InteractionKind    enumtypes.InteractionKind         `json:"interaction_kind"`
	RecipientProvider  string                            `json:"recipient_provider"`
	RecipientRef       string                            `json:"recipient_ref"`
	Locale             string                            `json:"locale,omitempty"`
	ContextLinks       interactionContextLinks           `json:"context_links"`
	Content            json.RawMessage                   `json:"content"`
	CallbackEndpoint   *interactionCallbackEndpoint      `json:"callback_endpoint,omitempty"`
	ContinuationPolicy interactionContinuationPolicy     `json:"continuation_policy"`
	DeliveryDeadlineAt string                            `json:"delivery_deadline_at,omitempty"`
}

type interactionCallbackEndpoint struct {
	URL            string                      `json:"url"`
	BearerToken    string                      `json:"bearer_token"`
	TokenExpiresAt string                      `json:"token_expires_at"`
	Handles        []interactionCallbackHandle `json:"handles"`
}

type interactionCallbackHandle struct {
	Handle      string                                      `json:"handle"`
	HandleKind  enumtypes.InteractionCallbackHandleKind     `json:"handle_kind"`
	ButtonLabel string                                      `json:"button_label,omitempty"`
	OptionID    string                                      `json:"option_id,omitempty"`
	ExpiresAt   string                                      `json:"expires_at"`
}

type interactionContinuationPolicy struct {
	PreferredMode                 string `json:"preferred_mode"`
	DisableKeyboardOnResolution   bool   `json:"disable_keyboard_on_resolution"`
	SendFollowUpOnEditFailure     bool   `json:"send_follow_up_on_edit_failure"`
	ManualFallbackOnFollowUpFailure bool `json:"manual_fallback_on_follow_up_failure"`
}

type interactionNotifyContent struct {
	NotificationKind UserNotificationKind `json:"notification_kind"`
	Summary          string               `json:"summary"`
	DetailsMarkdown  string               `json:"details_markdown,omitempty"`
	ActionLabel      string               `json:"action_label,omitempty"`
	ActionURL        string               `json:"action_url,omitempty"`
}

type interactionDecisionContent struct {
	Question            string                      `json:"question"`
	DetailsMarkdown     string                      `json:"details_markdown,omitempty"`
	Options             []interactionDecisionOption `json:"options"`
	AllowFreeText       bool                        `json:"allow_free_text,omitempty"`
	FreeTextPlaceholder string                      `json:"free_text_placeholder,omitempty"`
	ExpiresAt           string                      `json:"expires_at"`
	ReplyInstruction    string                      `json:"reply_instruction,omitempty"`
}

type interactionDecisionOption struct {
	OptionID       string `json:"option_id"`
	Label          string `json:"label"`
	Description    string `json:"description,omitempty"`
	CallbackHandle string `json:"callback_handle"`
}

func (s *Service) buildInteractionDeliveryEnvelope(
	ctx context.Context,
	run entitytypes.AgentRun,
	request entitytypes.InteractionRequest,
	attempt entitytypes.InteractionDeliveryAttempt,
) (json.RawMessage, error) {
	var contextLinks interactionContextLinks
	if len(request.ContextLinksJSON) > 0 {
		if err := json.Unmarshal(request.ContextLinksJSON, &contextLinks); err != nil {
			return nil, fmt.Errorf("unmarshal interaction context links: %w", err)
		}
	}

	callbackToken, err := s.issueInteractionCallbackToken(run, request, attempt.DeliveryID)
	if err != nil {
		return nil, fmt.Errorf("issue interaction callback token: %w", err)
	}
	callbackTokenExpiresAt := callbackToken.ExpiresAt
	binding, err := s.interactions.EnsureChannelBinding(ctx, interactionrequestrepo.EnsureChannelBindingParams{
		InteractionID:          request.ID,
		AdapterKind:            interactionRecipientProviderTelegram,
		RecipientRef:           request.RecipientRef,
		CallbackTokenKeyID:     callbackToken.KeyID,
		CallbackTokenExpiresAt: &callbackTokenExpiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("ensure interaction channel binding: %w", err)
	}

	callbackEndpoint := &interactionCallbackEndpoint{
		URL:            strings.TrimRight(strings.TrimSpace(s.cfg.PublicBaseURL), "/") + interactionCallbackPath,
		BearerToken:    callbackToken.Token,
		TokenExpiresAt: callbackToken.ExpiresAt.UTC().Format(time.RFC3339Nano),
		Handles:        []interactionCallbackHandle{},
	}

	content, err := s.buildInteractionDeliveryContent(ctx, request, binding, callbackEndpoint)
	if err != nil {
		return nil, err
	}

	envelope := interactionDeliveryEnvelope{
		SchemaVersion:      interactionDeliveryEnvelopeSchemaVersion,
		DeliveryID:         attempt.DeliveryID,
		InteractionID:      request.ID,
		InteractionKind:    request.InteractionKind,
		RecipientProvider:  request.RecipientProvider,
		RecipientRef:       request.RecipientRef,
		Locale:             interactionDeliveryLocaleDefault,
		ContextLinks:       contextLinks,
		Content:            content,
		CallbackEndpoint:   callbackEndpoint,
		ContinuationPolicy: defaultInteractionContinuationPolicy(),
	}
	if request.ResponseDeadlineAt != nil {
		envelope.DeliveryDeadlineAt = request.ResponseDeadlineAt.UTC().Format(time.RFC3339Nano)
	}
	return marshalRawJSON(envelope), nil
}

func (s *Service) buildInteractionDeliveryContent(
	ctx context.Context,
	request entitytypes.InteractionRequest,
	binding interactionrequestrepo.ChannelBinding,
	callbackEndpoint *interactionCallbackEndpoint,
) (json.RawMessage, error) {
	switch request.InteractionKind {
	case enumtypes.InteractionKindNotify:
		var payload UserNotifyInput
		if err := json.Unmarshal(request.RequestPayloadJSON, &payload); err != nil {
			return nil, fmt.Errorf("unmarshal notify interaction payload: %w", err)
		}
		return marshalRawJSON(interactionNotifyContent{
			NotificationKind: payload.NotificationKind,
			Summary:          strings.TrimSpace(payload.Summary),
			DetailsMarkdown:  strings.TrimSpace(payload.DetailsMarkdown),
			ActionLabel:      strings.TrimSpace(payload.ActionLabel),
			ActionURL:        strings.TrimSpace(payload.ActionURL),
		}), nil
	case enumtypes.InteractionKindDecisionRequest:
		return s.buildDecisionInteractionDeliveryContent(ctx, request, binding, callbackEndpoint)
	default:
		return nil, fmt.Errorf("unsupported interaction kind %q", request.InteractionKind)
	}
}

func (s *Service) buildDecisionInteractionDeliveryContent(
	ctx context.Context,
	request entitytypes.InteractionRequest,
	binding interactionrequestrepo.ChannelBinding,
	callbackEndpoint *interactionCallbackEndpoint,
) (json.RawMessage, error) {
	if request.ResponseDeadlineAt == nil {
		return nil, fmt.Errorf("decision interaction missing response deadline")
	}

	var payload UserDecisionRequestInput
	if err := json.Unmarshal(request.RequestPayloadJSON, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal decision interaction payload: %w", err)
	}

	handleDefs, options, err := s.allocateDecisionInteractionHandles(ctx, request, binding.ID, payload)
	if err != nil {
		return nil, err
	}
	callbackEndpoint.Handles = handleDefs

	content := interactionDecisionContent{
		Question:            strings.TrimSpace(payload.Question),
		DetailsMarkdown:     strings.TrimSpace(payload.DetailsMarkdown),
		Options:             options,
		AllowFreeText:       payload.AllowFreeText,
		FreeTextPlaceholder: strings.TrimSpace(payload.FreeTextPlaceholder),
		ExpiresAt:           request.ResponseDeadlineAt.UTC().Format(time.RFC3339Nano),
	}
	if payload.AllowFreeText {
		content.ReplyInstruction = interactionFreeTextReplyInstruction
	}
	return marshalRawJSON(content), nil
}

func (s *Service) allocateDecisionInteractionHandles(
	ctx context.Context,
	request entitytypes.InteractionRequest,
	channelBindingID int64,
	payload UserDecisionRequestInput,
) ([]interactionCallbackHandle, []interactionDecisionOption, error) {
	if channelBindingID == 0 {
		return nil, nil, fmt.Errorf("decision interaction channel binding is required")
	}
	if request.ResponseDeadlineAt == nil {
		return nil, nil, fmt.Errorf("decision interaction missing response deadline")
	}

	responseDeadline := request.ResponseDeadlineAt.UTC()
	graceExpiresAt := responseDeadline.Add(interactionCallbackTokenGraceTTL)
	handleDefs := make([]interactionCallbackHandle, 0, len(payload.Options)+1)
	options := make([]interactionDecisionOption, 0, len(payload.Options))
	upserts := make([]querytypes.InteractionCallbackHandleUpsertItem, 0, len(payload.Options)+1)

	for _, option := range payload.Options {
		rawHandle, handleHash, err := generateInteractionCallbackHandle()
		if err != nil {
			return nil, nil, err
		}
		handleDefs = append(handleDefs, interactionCallbackHandle{
			Handle:      rawHandle,
			HandleKind:  enumtypes.InteractionCallbackHandleKindOption,
			ButtonLabel: strings.TrimSpace(option.Label),
			OptionID:    strings.TrimSpace(option.OptionID),
			ExpiresAt:   responseDeadline.Format(time.RFC3339Nano),
		})
		options = append(options, interactionDecisionOption{
			OptionID:       strings.TrimSpace(option.OptionID),
			Label:          strings.TrimSpace(option.Label),
			Description:    strings.TrimSpace(option.Description),
			CallbackHandle: rawHandle,
		})
		upserts = append(upserts, querytypes.InteractionCallbackHandleUpsertItem{
			HandleHash:         handleHash,
			HandleKind:         enumtypes.InteractionCallbackHandleKindOption,
			OptionID:           strings.TrimSpace(option.OptionID),
			ResponseDeadlineAt: responseDeadline,
			GraceExpiresAt:     graceExpiresAt,
		})
	}

	if payload.AllowFreeText {
		rawHandle, handleHash, err := generateInteractionCallbackHandle()
		if err != nil {
			return nil, nil, err
		}
		handleDefs = append(handleDefs, interactionCallbackHandle{
			Handle:     rawHandle,
			HandleKind: enumtypes.InteractionCallbackHandleKindFreeTextSession,
			ExpiresAt:  responseDeadline.Format(time.RFC3339Nano),
		})
		upserts = append(upserts, querytypes.InteractionCallbackHandleUpsertItem{
			HandleHash:         handleHash,
			HandleKind:         enumtypes.InteractionCallbackHandleKindFreeTextSession,
			ResponseDeadlineAt: responseDeadline,
			GraceExpiresAt:     graceExpiresAt,
		})
	}

	if _, err := s.interactions.UpsertCallbackHandles(ctx, interactionrequestrepo.UpsertCallbackHandlesParams{
		InteractionID:    request.ID,
		ChannelBindingID: channelBindingID,
		Items:            upserts,
	}); err != nil {
		return nil, nil, fmt.Errorf("upsert interaction callback handles: %w", err)
	}

	return handleDefs, options, nil
}

func generateInteractionCallbackHandle() (string, []byte, error) {
	random := make([]byte, interactionCallbackHandleRandomBytes)
	if _, err := rand.Read(random); err != nil {
		return "", nil, fmt.Errorf("generate interaction callback handle: %w", err)
	}
	handle := interactionCallbackHandlePrefix + base64.RawURLEncoding.EncodeToString(random)
	sum := sha256.Sum256([]byte(handle))
	return handle, sum[:], nil
}

func defaultInteractionContinuationPolicy() interactionContinuationPolicy {
	return interactionContinuationPolicy{
		PreferredMode:                   interactionContinuationPreferredModeEditInPlace,
		DisableKeyboardOnResolution:     true,
		SendFollowUpOnEditFailure:       true,
		ManualFallbackOnFollowUpFailure: true,
	}
}
