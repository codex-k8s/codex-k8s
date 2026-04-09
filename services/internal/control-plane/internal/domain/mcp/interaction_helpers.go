package mcp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/codex-k8s/kodex/libs/go/errs"
	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

const (
	interactionRecipientProviderTelegram = "telegram"
	interactionRecipientRoutingByGitHub  = "github_login:"
	userNotifySummaryMaxChars            = 200
	userDecisionQuestionMaxChars         = 500
	userDecisionOptionLabelMaxChars      = 64
	userDecisionResponseTTLMinSeconds    = 60
	userDecisionResponseTTLMaxSeconds    = 86400
	userDecisionOptionsMin               = 2
	userDecisionOptionsMax               = 5
)

type interactionContextLinks struct {
	RunID              string `json:"run_id"`
	RunURL             string `json:"run_url,omitempty"`
	IssueURL           string `json:"issue_url,omitempty"`
	PullRequestURL     string `json:"pull_request_url,omitempty"`
	RepositoryFullName string `json:"repository_full_name,omitempty"`
}

type interactionCallbackPayload struct {
	InteractionID           string                            `json:"interaction_id"`
	DeliveryID              string                            `json:"delivery_id,omitempty"`
	AdapterEventID          string                            `json:"adapter_event_id"`
	CallbackKind            enumtypes.InteractionCallbackKind `json:"callback_kind"`
	OccurredAt              string                            `json:"occurred_at"`
	CallbackHandle          string                            `json:"callback_handle,omitempty"`
	DeliveryStatus          string                            `json:"delivery_status,omitempty"`
	FreeText                string                            `json:"free_text,omitempty"`
	ResponderRef            string                            `json:"responder_ref,omitempty"`
	ProviderMessageRefJSON  json.RawMessage                   `json:"provider_message_ref_json,omitempty"`
	ProviderUpdateID        string                            `json:"provider_update_id,omitempty"`
	ProviderCallbackQueryID string                            `json:"provider_callback_query_id,omitempty"`
	TransportErrorCode      string                            `json:"transport_error_code,omitempty"`
	TransportRetryable      bool                              `json:"transport_retryable,omitempty"`
}

func normalizeUserNotifyInput(input UserNotifyInput) (UserNotifyInput, error) {
	input.NotificationKind = UserNotificationKind(strings.TrimSpace(string(input.NotificationKind)))
	switch input.NotificationKind {
	case UserNotificationKindCompletion, UserNotificationKindNextStep, UserNotificationKindStatusUpdate, UserNotificationKindWarning:
	default:
		return UserNotifyInput{}, errs.Validation{Field: "notification_kind", Msg: "must be one of completion|next_step|status_update|warning"}
	}

	input.Summary = strings.TrimSpace(input.Summary)
	if input.Summary == "" {
		return UserNotifyInput{}, errs.Validation{Field: "summary", Msg: "must not be empty"}
	}
	if len([]rune(input.Summary)) > userNotifySummaryMaxChars {
		return UserNotifyInput{}, errs.Validation{Field: "summary", Msg: "must be at most 200 characters"}
	}

	input.DetailsMarkdown = strings.TrimSpace(input.DetailsMarkdown)
	input.ActionLabel = strings.TrimSpace(input.ActionLabel)
	input.ActionURL = strings.TrimSpace(input.ActionURL)
	if input.ActionURL != "" {
		if input.ActionLabel == "" {
			return UserNotifyInput{}, errs.Validation{Field: "action_label", Msg: "is required when action_url is provided"}
		}
		parsed, err := url.Parse(input.ActionURL)
		if err != nil || parsed == nil || !strings.EqualFold(parsed.Scheme, "https") || strings.TrimSpace(parsed.Host) == "" {
			return UserNotifyInput{}, errs.Validation{Field: "action_url", Msg: "must be a valid https URL"}
		}
	}

	return input, nil
}

func normalizeUserDecisionRequestInput(input UserDecisionRequestInput) (UserDecisionRequestInput, error) {
	input.Question = strings.TrimSpace(input.Question)
	if input.Question == "" {
		return UserDecisionRequestInput{}, errs.Validation{Field: "question", Msg: "must not be empty"}
	}
	if len([]rune(input.Question)) > userDecisionQuestionMaxChars {
		return UserDecisionRequestInput{}, errs.Validation{Field: "question", Msg: "must be at most 500 characters"}
	}

	if len(input.Options) < userDecisionOptionsMin || len(input.Options) > userDecisionOptionsMax {
		return UserDecisionRequestInput{}, errs.Validation{Field: "options", Msg: "must contain between 2 and 5 items"}
	}

	seenOptionIDs := make(map[string]struct{}, len(input.Options))
	normalizedOptions := make([]UserDecisionOption, 0, len(input.Options))
	for idx, option := range input.Options {
		option.OptionID = strings.TrimSpace(option.OptionID)
		option.Label = strings.TrimSpace(option.Label)
		option.Description = strings.TrimSpace(option.Description)
		if option.OptionID == "" {
			return UserDecisionRequestInput{}, errs.Validation{Field: fmt.Sprintf("options[%d].option_id", idx), Msg: "must not be empty"}
		}
		if option.Label == "" {
			return UserDecisionRequestInput{}, errs.Validation{Field: fmt.Sprintf("options[%d].label", idx), Msg: "must not be empty"}
		}
		if len([]rune(option.Label)) > userDecisionOptionLabelMaxChars {
			return UserDecisionRequestInput{}, errs.Validation{Field: fmt.Sprintf("options[%d].label", idx), Msg: "must be at most 64 characters"}
		}
		if _, exists := seenOptionIDs[option.OptionID]; exists {
			return UserDecisionRequestInput{}, errs.Validation{Field: "options", Msg: "duplicate option_id is not allowed"}
		}
		seenOptionIDs[option.OptionID] = struct{}{}
		normalizedOptions = append(normalizedOptions, option)
	}
	input.Options = normalizedOptions
	input.DetailsMarkdown = strings.TrimSpace(input.DetailsMarkdown)
	input.FreeTextPlaceholder = strings.TrimSpace(input.FreeTextPlaceholder)

	if !input.AllowFreeText && input.FreeTextPlaceholder != "" {
		return UserDecisionRequestInput{}, errs.Validation{Field: "free_text_placeholder", Msg: "requires allow_free_text=true"}
	}
	if input.ResponseTTLSeconds < userDecisionResponseTTLMinSeconds || input.ResponseTTLSeconds > userDecisionResponseTTLMaxSeconds {
		return UserDecisionRequestInput{}, errs.Validation{Field: "response_ttl_seconds", Msg: "must be between 60 and 86400 seconds"}
	}

	return input, nil
}

func resolveInteractionRecipient(runCtx resolvedRunContext) (string, string, error) {
	if runCtx.Payload.Issue != nil {
		if login := strings.TrimSpace(runCtx.Payload.Issue.User.Login); login != "" {
			return interactionRecipientProviderTelegram, interactionRecipientRoutingByGitHub + login, nil
		}
	}
	if runCtx.Payload.PullRequest != nil {
		if login := strings.TrimSpace(runCtx.Payload.PullRequest.User.Login); login != "" {
			return interactionRecipientProviderTelegram, interactionRecipientRoutingByGitHub + login, nil
		}
	}
	if login := strings.TrimSpace(runCtx.Payload.Sender.Login); login != "" {
		return interactionRecipientProviderTelegram, interactionRecipientRoutingByGitHub + login, nil
	}
	return "", "", errs.FailedPrecondition{Msg: "run context does not expose a resolvable recipient"}
}

func buildInteractionContextLinks(runCtx resolvedRunContext, publicBaseURL string) interactionContextLinks {
	links := interactionContextLinks{
		RunID:              runCtx.Session.RunID,
		RepositoryFullName: strings.TrimSpace(runCtx.Repository.Owner) + "/" + strings.TrimSpace(runCtx.Repository.Name),
	}
	if baseURL := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/"); baseURL != "" && strings.TrimSpace(runCtx.Session.RunID) != "" {
		links.RunURL = baseURL + "/runs/" + strings.TrimSpace(runCtx.Session.RunID)
	}
	if runCtx.Payload.Issue != nil {
		links.IssueURL = strings.TrimSpace(runCtx.Payload.Issue.HTMLURL)
	}
	if runCtx.Payload.PullRequest != nil {
		links.PullRequestURL = strings.TrimSpace(runCtx.Payload.PullRequest.HTMLURL)
	}
	return links
}

func buildInteractionCallbackNormalizedPayload(params SubmitInteractionCallbackParams) json.RawMessage {
	return marshalRawJSON(interactionCallbackPayload{
		InteractionID:           strings.TrimSpace(params.InteractionID),
		DeliveryID:              strings.TrimSpace(params.DeliveryID),
		AdapterEventID:          strings.TrimSpace(params.AdapterEventID),
		CallbackKind:            params.CallbackKind,
		OccurredAt:              params.OccurredAt.UTC().Format(time.RFC3339Nano),
		CallbackHandle:          strings.TrimSpace(params.CallbackHandle),
		DeliveryStatus:          strings.TrimSpace(params.DeliveryStatus),
		FreeText:                strings.TrimSpace(params.FreeText),
		ResponderRef:            strings.TrimSpace(params.ResponderRef),
		ProviderMessageRefJSON:  params.ProviderMessageRefJSON,
		ProviderUpdateID:        strings.TrimSpace(params.ProviderUpdateID),
		ProviderCallbackQueryID: strings.TrimSpace(params.ProviderCallbackQueryID),
		TransportErrorCode:      strings.TrimSpace(params.TransportErrorCode),
		TransportRetryable:      params.TransportRetryable,
	})
}

func buildInteractionResumePayload(request entitytypes.InteractionRequest, response *entitytypes.InteractionResponseRecord) *valuetypes.InteractionResumePayload {
	if request.InteractionKind != enumtypes.InteractionKindDecisionRequest {
		return nil
	}

	payload := &valuetypes.InteractionResumePayload{
		InteractionID: request.ID,
		ToolName:      string(ToolMCPUserDecisionRequest),
	}

	resolvedAt := request.UpdatedAt.UTC()
	switch request.State {
	case enumtypes.InteractionStateResolved:
		payload.RequestStatus = enumtypes.InteractionRequestStatusAnswered
		payload.ResolutionReason = string(enumtypes.InteractionCallbackResultClassificationAccepted)
		if response != nil {
			payload.ResponseKind = response.ResponseKind
			payload.SelectedOptionID = response.SelectedOptionID
			payload.FreeText = response.FreeText
			resolvedAt = response.RespondedAt.UTC()
		} else {
			payload.ResponseKind = enumtypes.InteractionResponseKindNone
		}
	case enumtypes.InteractionStateExpired:
		payload.RequestStatus = enumtypes.InteractionRequestStatusExpired
		payload.ResponseKind = enumtypes.InteractionResponseKindNone
		payload.ResolutionReason = string(enumtypes.InteractionRequestStatusExpired)
	case enumtypes.InteractionStateDeliveryExhausted:
		payload.RequestStatus = enumtypes.InteractionRequestStatusDeliveryExhausted
		payload.ResponseKind = enumtypes.InteractionResponseKindNone
		payload.ResolutionReason = string(enumtypes.InteractionRequestStatusDeliveryExhausted)
	case enumtypes.InteractionStateCancelled:
		payload.RequestStatus = enumtypes.InteractionRequestStatusCancelled
		payload.ResponseKind = enumtypes.InteractionResponseKindNone
		payload.ResolutionReason = string(enumtypes.InteractionRequestStatusCancelled)
	default:
		return nil
	}

	payload.ResolvedAt = resolvedAt.Format(time.RFC3339Nano)
	return payload
}
