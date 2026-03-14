package casters

import (
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/services/external/telegram-interaction-adapter/internal/service"
	"github.com/codex-k8s/codex-k8s/services/external/telegram-interaction-adapter/internal/transport/http/models"
)

// DeliveryEnvelope converts transport DTO into adapter service input.
func DeliveryEnvelope(input models.TelegramInteractionDeliveryEnvelope) service.DeliveryEnvelope {
	output := service.DeliveryEnvelope{
		SchemaVersion:      strings.TrimSpace(input.SchemaVersion),
		DeliveryID:         strings.TrimSpace(input.DeliveryID),
		DeliveryRole:       strings.TrimSpace(input.DeliveryRole),
		InteractionID:      strings.TrimSpace(input.InteractionID),
		InteractionKind:    strings.TrimSpace(input.InteractionKind),
		RecipientProvider:  strings.TrimSpace(input.RecipientProvider),
		RecipientRef:       strings.TrimSpace(input.RecipientRef),
		Locale:             strings.TrimSpace(input.Locale),
		ContextLinks:       contextLinks(input.ContextLinks),
		Content:            interactionContent(input.Content),
		CallbackEndpoint:   callbackEndpoint(input.CallbackEndpoint),
		ProviderMessageRef: providerMessageRef(input.ProviderMessageRef),
		Continuation:       continuation(input.Continuation),
		ContinuationPolicy: continuationPolicy(input.ContinuationPolicy),
		DeliveryDeadlineAt: input.DeliveryDeadlineAt,
	}
	return output
}

// DeliveryResponse converts adapter service output into transport DTO.
func DeliveryResponse(input service.DeliveryResponse) models.TelegramInteractionDeliveryResponse {
	output := models.TelegramInteractionDeliveryResponse{
		Accepted:           input.Accepted,
		AdapterDeliveryID:  optionalTrimmedString(input.AdapterDeliveryID),
		ProviderMessageRef: providerMessageRefModel(input.ProviderMessageRef),
		EditCapability:     optionalTrimmedString(input.EditCapability),
		Retryable:          input.Retryable,
		Message:            optionalTrimmedString(input.Message),
	}
	return output
}

func contextLinks(input models.InteractionContextLinks) service.ContextLinks {
	return service.ContextLinks{
		RunID:              strings.TrimSpace(input.RunID),
		RunURL:             stringValue(input.RunURL),
		IssueURL:           stringValue(input.IssueURL),
		PullRequestURL:     stringValue(input.PullRequestURL),
		RepositoryFullName: stringValue(input.RepositoryFullName),
	}
}

func interactionContent(input *models.TelegramInteractionContent) service.InteractionContent {
	if input == nil {
		return service.InteractionContent{}
	}
	output := service.InteractionContent{
		NotificationKind:    stringValue(input.NotificationKind),
		Summary:             stringValue(input.Summary),
		DetailsMarkdown:     stringValue(input.DetailsMarkdown),
		ActionLabel:         stringValue(input.ActionLabel),
		ActionURL:           stringValue(input.ActionURL),
		Question:            stringValue(input.Question),
		AllowFreeText:       boolValue(input.AllowFreeText),
		FreeTextPlaceholder: stringValue(input.FreeTextPlaceholder),
		ExpiresAt:           input.ExpiresAt,
		ReplyInstruction:    stringValue(input.ReplyInstruction),
	}
	if len(input.Options) > 0 {
		output.Options = make([]service.DecisionOption, 0, len(input.Options))
		for _, option := range input.Options {
			output.Options = append(output.Options, service.DecisionOption{
				OptionID:       strings.TrimSpace(option.OptionID),
				Label:          strings.TrimSpace(option.Label),
				Description:    stringValue(option.Description),
				CallbackHandle: strings.TrimSpace(option.CallbackHandle),
			})
		}
	}
	return output
}

func callbackEndpoint(input *models.TelegramCallbackEndpoint) *service.CallbackEndpoint {
	if input == nil {
		return nil
	}
	output := &service.CallbackEndpoint{
		URL:            strings.TrimSpace(input.URL),
		BearerToken:    strings.TrimSpace(input.BearerToken),
		TokenExpiresAt: input.TokenExpiresAt,
	}
	if len(input.Handles) > 0 {
		output.Handles = make([]service.CallbackHandle, 0, len(input.Handles))
		for _, handle := range input.Handles {
			expiresAt := time.Time{}
			if handle.ExpiresAt != nil {
				expiresAt = handle.ExpiresAt.UTC()
			}
			output.Handles = append(output.Handles, service.CallbackHandle{
				Handle:      strings.TrimSpace(handle.Handle),
				HandleKind:  strings.TrimSpace(handle.HandleKind),
				ButtonLabel: stringValue(handle.ButtonLabel),
				OptionID:    stringValue(handle.OptionID),
				ExpiresAt:   expiresAt,
			})
		}
	}
	return output
}

func providerMessageRef(input *models.TelegramProviderMessageRef) *service.ProviderMessageRef {
	if input == nil {
		return nil
	}
	chatRef, messageID, inlineMessageID := providerMessageRefStrings(
		stringValue(input.ChatRef),
		stringValue(input.MessageID),
		stringValue(input.InlineMessageID),
	)
	return &service.ProviderMessageRef{
		ChatRef:         chatRef,
		MessageID:       messageID,
		InlineMessageID: inlineMessageID,
		SentAt:          input.SentAt,
	}
}

func providerMessageRefModel(input *service.ProviderMessageRef) *models.TelegramProviderMessageRef {
	if input == nil {
		return nil
	}
	chatRef, messageID, inlineMessageID := providerMessageRefPointers(
		input.ChatRef,
		input.MessageID,
		input.InlineMessageID,
	)
	return &models.TelegramProviderMessageRef{
		ChatRef:         chatRef,
		MessageID:       messageID,
		InlineMessageID: inlineMessageID,
		SentAt:          input.SentAt,
	}
}

func continuation(input *models.TelegramInteractionContinuation) *service.Continuation {
	if input == nil {
		return nil
	}
	return &service.Continuation{
		Action:         strings.TrimSpace(input.Action),
		Reason:         stringValue(input.Reason),
		ResolutionKind: stringValue(input.ResolutionKind),
		ResolvedAt:     input.ResolvedAt,
	}
}

func continuationPolicy(input models.TelegramContinuationPolicy) service.ContinuationPolicy {
	return service.ContinuationPolicy{
		PreferredMode:                   strings.TrimSpace(input.PreferredMode),
		DisableKeyboardOnResolution:     input.DisableKeyboardOnResolution,
		SendFollowUpOnEditFailure:       input.SendFollowUpOnEditFailure,
		ManualFallbackOnFollowUpFailure: input.ManualFallbackOnFollowUpFailure,
	}
}

func stringValue(input *string) string {
	if input == nil {
		return ""
	}
	return strings.TrimSpace(*input)
}

func boolValue(input *bool) bool {
	if input == nil {
		return false
	}
	return *input
}

func optionalTrimmedString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func providerMessageRefStrings(chatRef string, messageID string, inlineMessageID string) (string, string, string) {
	return strings.TrimSpace(chatRef), strings.TrimSpace(messageID), strings.TrimSpace(inlineMessageID)
}

func providerMessageRefPointers(chatRef string, messageID string, inlineMessageID string) (*string, *string, *string) {
	return optionalTrimmedString(chatRef), optionalTrimmedString(messageID), optionalTrimmedString(inlineMessageID)
}
