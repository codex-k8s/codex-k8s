package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"
)

const telegramWebhookPath = "/api/v1/telegram/interactions/webhook"

// Config wires runtime dependencies for adapter service.
type Config struct {
	PublicBaseURL  string
	WebhookSecret  string
	DeliveryToken  string
	Recipients     *RecipientResolver
	Bot            BotClient
	CallbackSink   CallbackSink
	AudioConverter AudioConverter
	SpeechToText   SpeechToText
	Logger         *slog.Logger
}

// Service owns Telegram transport logic and callback forwarding.
type Service struct {
	publicBaseURL  string
	webhookSecret  string
	deliveryToken  string
	recipients     *RecipientResolver
	bot            BotClient
	callbacks      CallbackSink
	audioConverter AudioConverter
	speechToText   SpeechToText
	messages       *messageRenderer
	logger         *slog.Logger
}

// New builds the adapter service.
func New(cfg Config) (*Service, error) {
	renderer, err := newMessageRenderer()
	if err != nil {
		return nil, err
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Recipients == nil {
		return nil, fmt.Errorf("recipient resolver is required")
	}
	if cfg.Bot == nil {
		return nil, fmt.Errorf("telegram bot client is required")
	}
	if cfg.CallbackSink == nil {
		return nil, fmt.Errorf("callback sink is required")
	}

	return &Service{
		publicBaseURL:  strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/"),
		webhookSecret:  strings.TrimSpace(cfg.WebhookSecret),
		deliveryToken:  strings.TrimSpace(cfg.DeliveryToken),
		recipients:     cfg.Recipients,
		bot:            cfg.Bot,
		callbacks:      cfg.CallbackSink,
		audioConverter: cfg.AudioConverter,
		speechToText:   cfg.SpeechToText,
		messages:       renderer,
		logger:         cfg.Logger,
	}, nil
}

// DeliveryToken returns worker -> adapter bearer token expected by the service.
func (s *Service) DeliveryToken() string {
	return s.deliveryToken
}

// WebhookSecret returns the configured Telegram secret token.
func (s *Service) WebhookSecret() string {
	return s.webhookSecret
}

// SyncWebhook configures Telegram webhook when public base URL and bot token are available.
func (s *Service) SyncWebhook(ctx context.Context) error {
	if !s.bot.Ready() || s.publicBaseURL == "" {
		return nil
	}
	return s.bot.SetWebhook(ctx, SetWebhookRequest{
		URL:         s.publicBaseURL + telegramWebhookPath,
		SecretToken: s.webhookSecret,
	})
}

// Deliver handles one worker -> adapter delivery request.
func (s *Service) Deliver(ctx context.Context, envelope DeliveryEnvelope) (DeliveryResponse, error) {
	if !s.bot.Ready() {
		recordDispatchAttempt(envelope.DeliveryRole, metricStatusFailed)
		return DeliveryResponse{}, &DeliveryError{
			StatusCode: http.StatusServiceUnavailable,
			Response: DeliveryResponse{
				Accepted:  false,
				Retryable: false,
				Message:   "telegram bot token is not configured",
			},
		}
	}

	switch strings.TrimSpace(envelope.DeliveryRole) {
	case DeliveryRolePrimaryDispatch:
		return s.deliverPrimary(ctx, envelope)
	case DeliveryRoleMessageEdit:
		return s.deliverMessageEdit(ctx, envelope)
	case DeliveryRoleFollowUpNotify:
		return s.deliverFollowUp(ctx, envelope)
	default:
		recordDispatchAttempt(envelope.DeliveryRole, metricStatusRejected)
		return DeliveryResponse{}, &DeliveryError{
			StatusCode: http.StatusBadRequest,
			Response: DeliveryResponse{
				Accepted:  false,
				Retryable: false,
				Message:   fmt.Sprintf("unsupported delivery_role %q", envelope.DeliveryRole),
			},
		}
	}
}

func (s *Service) deliverPrimary(ctx context.Context, envelope DeliveryEnvelope) (DeliveryResponse, error) {
	chatID, err := s.recipients.Resolve(envelope.RecipientRef)
	if err != nil {
		recordDispatchAttempt(envelope.DeliveryRole, metricStatusRejected)
		return DeliveryResponse{}, &DeliveryError{
			StatusCode: http.StatusUnprocessableEntity,
			Response: DeliveryResponse{
				Accepted:  false,
				Retryable: false,
				Message:   err.Error(),
			},
		}
	}

	text, inlineOptions, actionLabel, actionURL, err := s.buildPrimaryMessage(envelope)
	if err != nil {
		recordDispatchAttempt(envelope.DeliveryRole, metricStatusRejected)
		return DeliveryResponse{}, &DeliveryError{
			StatusCode: http.StatusBadRequest,
			Response: DeliveryResponse{
				Accepted:  false,
				Retryable: false,
				Message:   err.Error(),
			},
		}
	}

	sent, err := s.bot.SendMessage(ctx, SendMessageRequest{
		ChatID:        chatID,
		Text:          text,
		ActionLabel:   actionLabel,
		ActionURL:     actionURL,
		InlineOptions: inlineOptions,
	})
	if err != nil {
		recordDispatchAttempt(envelope.DeliveryRole, metricStatusFailed)
		return DeliveryResponse{}, classifyTelegramDeliveryError(err)
	}

	recordDispatchAttempt(envelope.DeliveryRole, metricStatusAccepted)
	return DeliveryResponse{
		Accepted:           true,
		AdapterDeliveryID:  buildAdapterDeliveryID(envelope.DeliveryRole, sent.MessageID),
		ProviderMessageRef: normalizeTelegramProviderMessageRef(sent.ChatID, sent.MessageID, sent.SentAt),
		EditCapability:     resolveEditCapability(envelope),
		Retryable:          false,
	}, nil
}

func (s *Service) deliverMessageEdit(ctx context.Context, envelope DeliveryEnvelope) (DeliveryResponse, error) {
	messageRef, err := resolveProviderMessageRef(envelope.ProviderMessageRef)
	if err != nil {
		recordDispatchAttempt(envelope.DeliveryRole, metricStatusRejected)
		recordContinuationAttempt(ContinuationActionEditMessage, metricStatusRejected)
		return DeliveryResponse{}, &DeliveryError{
			StatusCode: http.StatusUnprocessableEntity,
			Response: DeliveryResponse{
				Accepted:  false,
				Retryable: false,
				Message:   err.Error(),
			},
		}
	}

	chatID, messageID, err := providerMessageIdentity(*messageRef)
	if err != nil {
		recordDispatchAttempt(envelope.DeliveryRole, metricStatusRejected)
		recordContinuationAttempt(ContinuationActionEditMessage, metricStatusRejected)
		return DeliveryResponse{}, &DeliveryError{
			StatusCode: http.StatusUnprocessableEntity,
			Response: DeliveryResponse{
				Accepted:  false,
				Retryable: false,
				Message:   err.Error(),
			},
		}
	}

	if err := s.bot.EditMessageKeyboard(ctx, EditMessageKeyboardRequest{
		ChatID:    chatID,
		MessageID: messageID,
	}); err != nil {
		recordDispatchAttempt(envelope.DeliveryRole, metricStatusFailed)
		recordContinuationAttempt(ContinuationActionEditMessage, metricStatusFailed)
		return DeliveryResponse{}, classifyTelegramDeliveryError(err)
	}

	recordDispatchAttempt(envelope.DeliveryRole, metricStatusAccepted)
	recordContinuationAttempt(ContinuationActionEditMessage, metricStatusAccepted)
	return DeliveryResponse{
		Accepted:           true,
		AdapterDeliveryID:  buildAdapterDeliveryID(envelope.DeliveryRole, messageID),
		ProviderMessageRef: messageRef,
		EditCapability:     EditCapabilityKeyboardOnly,
		Retryable:          false,
	}, nil
}

func (s *Service) deliverFollowUp(ctx context.Context, envelope DeliveryEnvelope) (DeliveryResponse, error) {
	chatID, err := resolveFollowUpChatID(envelope.ProviderMessageRef)
	if err != nil {
		recordDispatchAttempt(envelope.DeliveryRole, metricStatusRejected)
		recordContinuationAttempt(ContinuationActionSendFollowUp, metricStatusRejected)
		return DeliveryResponse{}, &DeliveryError{
			StatusCode: http.StatusUnprocessableEntity,
			Response: DeliveryResponse{
				Accepted:  false,
				Retryable: false,
				Message:   err.Error(),
			},
		}
	}

	text := s.messages.Render(envelope.Locale, followUpTemplateKey(envelope), followUpMessageData{
		RunURL:         envelope.ContextLinks.RunURL,
		IssueURL:       envelope.ContextLinks.IssueURL,
		PullRequestURL: envelope.ContextLinks.PullRequestURL,
	})
	if text == "" {
		text = s.messages.Render(envelope.Locale, "follow_up_applied_response", followUpMessageData{
			RunURL:         envelope.ContextLinks.RunURL,
			IssueURL:       envelope.ContextLinks.IssueURL,
			PullRequestURL: envelope.ContextLinks.PullRequestURL,
		})
	}

	sent, err := s.bot.SendMessage(ctx, SendMessageRequest{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		recordDispatchAttempt(envelope.DeliveryRole, metricStatusFailed)
		recordContinuationAttempt(ContinuationActionSendFollowUp, metricStatusFailed)
		return DeliveryResponse{}, classifyTelegramDeliveryError(err)
	}

	recordDispatchAttempt(envelope.DeliveryRole, metricStatusAccepted)
	recordContinuationAttempt(ContinuationActionSendFollowUp, metricStatusAccepted)
	return DeliveryResponse{
		Accepted:           true,
		AdapterDeliveryID:  buildAdapterDeliveryID(envelope.DeliveryRole, sent.MessageID),
		ProviderMessageRef: normalizeTelegramProviderMessageRef(sent.ChatID, sent.MessageID, sent.SentAt),
		EditCapability:     EditCapabilityFollowUpOnly,
		Retryable:          false,
	}, nil
}

// HandleWebhook processes one raw Telegram update.
func (s *Service) HandleWebhook(ctx context.Context, raw []byte) error {
	var update telego.Update
	if err := json.Unmarshal(raw, &update); err != nil {
		return fmt.Errorf("unmarshal telegram update: %w", err)
	}

	switch {
	case update.CallbackQuery != nil:
		return s.handleCallbackQuery(ctx, update)
	case update.Message != nil:
		return s.handleMessage(ctx, update)
	default:
		return nil
	}
}

func (s *Service) handleCallbackQuery(ctx context.Context, update telego.Update) error {
	query := update.CallbackQuery
	if query == nil {
		return nil
	}

	locale := telegramLocale(&query.From)
	chatID := int64(0)
	if query.Message != nil {
		chatID = query.Message.GetChat().ID
	}
	if chatID != 0 && !s.recipients.IsAllowedChat(chatID) {
		recordCallbackEvent(CallbackKindOptionSelected, metricStatusIgnored)
		return s.bot.AnswerCallbackQuery(ctx, AnswerCallbackQueryRequest{
			QueryID: query.ID,
			Text:    s.messages.Render(locale, "callback_ack_unavailable", nil),
		})
	}

	handle := strings.TrimSpace(query.Data)
	if handle == "" {
		recordCallbackEvent(CallbackKindOptionSelected, metricStatusIgnored)
		return s.bot.AnswerCallbackQuery(ctx, AnswerCallbackQueryRequest{
			QueryID: query.ID,
			Text:    s.messages.Render(locale, "callback_ack_unavailable", nil),
		})
	}

	if err := s.bot.AnswerCallbackQuery(ctx, AnswerCallbackQueryRequest{
		QueryID: query.ID,
		Text:    s.messages.Render(locale, "callback_ack_received", nil),
	}); err != nil {
		s.logger.Warn("answer telegram callback query failed", "provider_callback_query_id", query.ID, "err", err)
	}

	var providerMessageRef *ProviderMessageRef
	if query.Message != nil {
		providerMessageRef = providerMessageRefFromMaybeInaccessibleMessage(query.Message)
	}

	outcome, err := s.forwardCallback(ctx, CallbackEnvelope{
		SchemaVersion:           SchemaVersionTelegramInteractionV1,
		AdapterEventID:          "callback:" + strings.TrimSpace(query.ID),
		CallbackKind:            CallbackKindOptionSelected,
		OccurredAt:              time.Now().UTC().Format(time.RFC3339Nano),
		CallbackHandle:          handle,
		ResponderRef:            buildResponderRef(&query.From),
		ProviderMessageRef:      providerMessageRef,
		ProviderUpdateID:        strconv.Itoa(update.UpdateID),
		ProviderCallbackQueryID: strings.TrimSpace(query.ID),
	})
	if err != nil {
		recordCallbackEvent(CallbackKindOptionSelected, metricStatusFailed)
		return err
	}

	recordCallbackEvent(CallbackKindOptionSelected, outcome.Classification)
	return nil
}

func (s *Service) handleMessage(ctx context.Context, update telego.Update) error {
	message := update.Message
	if message == nil {
		return nil
	}
	if !s.recipients.IsAllowedChat(message.Chat.ID) {
		recordCallbackEvent(CallbackKindFreeTextReceived, metricStatusIgnored)
		return nil
	}

	locale := telegramLocale(message.From)
	freeText, err := s.resolveMessageText(ctx, message, locale)
	if err != nil {
		recordCallbackEvent(CallbackKindFreeTextReceived, metricStatusFailed)
		return err
	}
	if freeText == "" {
		return nil
	}

	outcome, err := s.forwardCallback(ctx, CallbackEnvelope{
		SchemaVersion:      SchemaVersionTelegramInteractionV1,
		AdapterEventID:     "message:" + strconv.Itoa(update.UpdateID),
		CallbackKind:       CallbackKindFreeTextReceived,
		OccurredAt:         time.Now().UTC().Format(time.RFC3339Nano),
		FreeText:           freeText,
		ResponderRef:       buildResponderRef(message.From),
		ProviderMessageRef: providerMessageRefForReply(message),
		ProviderUpdateID:   strconv.Itoa(update.UpdateID),
	})
	if err != nil {
		recordCallbackEvent(CallbackKindFreeTextReceived, metricStatusFailed)
		_, _ = s.bot.SendMessage(ctx, SendMessageRequest{
			ChatID: message.Chat.ID,
			Text:   s.messages.Render(locale, "free_text_failed", nil),
		})
		return err
	}

	confirmationKey := "free_text_received"
	switch strings.TrimSpace(outcome.Classification) {
	case "invalid", "expired", "stale", "duplicate":
		confirmationKey = "free_text_unavailable"
	}
	_, _ = s.bot.SendMessage(ctx, SendMessageRequest{
		ChatID: message.Chat.ID,
		Text:   s.messages.Render(locale, confirmationKey, nil),
	})
	recordCallbackEvent(CallbackKindFreeTextReceived, outcome.Classification)
	return nil
}

func (s *Service) resolveMessageText(ctx context.Context, message *telego.Message, locale string) (string, error) {
	if message == nil {
		return "", nil
	}
	if freeText := strings.TrimSpace(message.Text); freeText != "" {
		return freeText, nil
	}
	if message.Voice == nil {
		return "", nil
	}
	if s.speechToText == nil || s.audioConverter == nil {
		_, _ = s.bot.SendMessage(ctx, SendMessageRequest{
			ChatID: message.Chat.ID,
			Text:   s.messages.Render(locale, "voice_disabled", nil),
		})
		return "", nil
	}

	file, err := s.bot.DownloadFile(ctx, message.Voice.FileID)
	if err != nil {
		_, _ = s.bot.SendMessage(ctx, SendMessageRequest{
			ChatID: message.Chat.ID,
			Text:   s.messages.Render(locale, "transcription_failed", nil),
		})
		return "", fmt.Errorf("download telegram voice file: %w", err)
	}

	normalized, err := s.audioConverter.Convert(ctx, AudioPayload(file))
	if err != nil {
		_, _ = s.bot.SendMessage(ctx, SendMessageRequest{
			ChatID: message.Chat.ID,
			Text:   s.messages.Render(locale, "transcription_failed", nil),
		})
		return "", fmt.Errorf("normalize telegram voice file: %w", err)
	}

	transcript, err := s.speechToText.Transcribe(ctx, normalized, telegramSTTLanguage(locale))
	if err != nil {
		_, _ = s.bot.SendMessage(ctx, SendMessageRequest{
			ChatID: message.Chat.ID,
			Text:   s.messages.Render(locale, "transcription_failed", nil),
		})
		return "", fmt.Errorf("transcribe telegram voice file: %w", err)
	}
	return strings.TrimSpace(transcript), nil
}

func (s *Service) forwardCallback(ctx context.Context, envelope CallbackEnvelope) (CallbackOutcome, error) {
	outcome, err := s.callbacks.Submit(ctx, envelope)
	if err != nil {
		return CallbackOutcome{}, fmt.Errorf("submit adapter callback: %w", err)
	}
	s.logger.Info(
		"telegram callback forwarded",
		"interaction_id", envelope.InteractionID,
		"delivery_id", envelope.DeliveryID,
		"classification", outcome.Classification,
		"resume_required", outcome.ResumeRequired,
	)
	return outcome, nil
}

func (s *Service) buildPrimaryMessage(envelope DeliveryEnvelope) (string, []InlineOption, string, string, error) {
	switch envelope.InteractionKind {
	case InteractionKindNotify:
		if strings.TrimSpace(envelope.Content.Summary) == "" {
			return "", nil, "", "", fmt.Errorf("notify content summary is required")
		}
		text := s.messages.Render(envelope.Locale, "notify_message", notifyMessageData{
			Summary:         strings.TrimSpace(envelope.Content.Summary),
			DetailsMarkdown: strings.TrimSpace(envelope.Content.DetailsMarkdown),
			Links:           messageLinksFromEnvelope(envelope),
		})
		return text, nil, decorateActionLabel(strings.TrimSpace(envelope.Content.ActionLabel)), envelope.Content.ActionURL, nil
	case InteractionKindDecisionRequest:
		if strings.TrimSpace(envelope.Content.Question) == "" {
			return "", nil, "", "", fmt.Errorf("decision content question is required")
		}
		options := make([]InlineOption, 0, len(envelope.Content.Options))
		for idx, option := range envelope.Content.Options {
			if strings.TrimSpace(option.Label) == "" || strings.TrimSpace(option.CallbackHandle) == "" {
				return "", nil, "", "", fmt.Errorf("decision options require label and callback_handle")
			}
			options = append(options, InlineOption{
				Label:        decorateOptionLabel(option, idx),
				CallbackData: option.CallbackHandle,
			})
		}
		text := s.messages.Render(envelope.Locale, "decision_message", decisionMessageData{
			Question:         strings.TrimSpace(envelope.Content.Question),
			DetailsMarkdown:  strings.TrimSpace(envelope.Content.DetailsMarkdown),
			ReplyInstruction: strings.TrimSpace(envelope.Content.ReplyInstruction),
			Links:            messageLinksFromEnvelope(envelope),
		})
		return text, options, "", "", nil
	default:
		return "", nil, "", "", fmt.Errorf("unsupported interaction_kind %q", envelope.InteractionKind)
	}
}

func messageLinksFromEnvelope(envelope DeliveryEnvelope) messageLinks {
	return messageLinks{
		RunURL:         strings.TrimSpace(envelope.ContextLinks.RunURL),
		IssueURL:       strings.TrimSpace(envelope.ContextLinks.IssueURL),
		PullRequestURL: strings.TrimSpace(envelope.ContextLinks.PullRequestURL),
	}
}

func resolveProviderMessageRef(ref *ProviderMessageRef) (*ProviderMessageRef, error) {
	if ref == nil {
		return nil, fmt.Errorf("provider_message_ref is required for continuation")
	}
	if strings.TrimSpace(ref.ChatRef) == "" || strings.TrimSpace(ref.MessageID) == "" {
		return nil, fmt.Errorf("provider_message_ref.chat_ref and provider_message_ref.message_id are required for continuation")
	}
	return ref, nil
}

func providerMessageIdentity(ref ProviderMessageRef) (int64, int, error) {
	chatID, err := strconv.ParseInt(strings.TrimSpace(ref.ChatRef), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid provider chat_ref %q: %w", ref.ChatRef, err)
	}
	messageID, err := strconv.Atoi(strings.TrimSpace(ref.MessageID))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid provider message_id %q: %w", ref.MessageID, err)
	}
	return chatID, messageID, nil
}

func resolveFollowUpChatID(ref *ProviderMessageRef) (int64, error) {
	if ref == nil || strings.TrimSpace(ref.ChatRef) == "" {
		return 0, fmt.Errorf("continuation provider chat_ref is required")
	}
	return strconv.ParseInt(strings.TrimSpace(ref.ChatRef), 10, 64)
}

func resolveEditCapability(envelope DeliveryEnvelope) string {
	if envelope.InteractionKind == InteractionKindDecisionRequest && len(envelope.Content.Options) > 0 {
		return EditCapabilityKeyboardOnly
	}
	return EditCapabilityFollowUpOnly
}

func buildAdapterDeliveryID(role string, messageID int) string {
	return strings.TrimSpace(role) + ":" + strconv.Itoa(messageID)
}

func buildResponderRef(user *telego.User) string {
	if user == nil {
		return ""
	}
	return "telegram_user:" + strconv.FormatInt(user.ID, 10)
}

func providerMessageRefFromTelegramMessage(message *telego.Message) *ProviderMessageRef {
	if message == nil {
		return nil
	}
	return &ProviderMessageRef{
		ChatRef:   strconv.FormatInt(message.Chat.ID, 10),
		MessageID: strconv.Itoa(message.MessageID),
	}
}

func providerMessageRefFromMaybeInaccessibleMessage(message telego.MaybeInaccessibleMessage) *ProviderMessageRef {
	if message == nil {
		return nil
	}
	return &ProviderMessageRef{
		ChatRef:   strconv.FormatInt(message.GetChat().ID, 10),
		MessageID: strconv.Itoa(message.GetMessageID()),
	}
}

func providerMessageRefForReply(message *telego.Message) *ProviderMessageRef {
	if message == nil {
		return nil
	}
	if message.ReplyToMessage != nil {
		return providerMessageRefFromTelegramMessage(message.ReplyToMessage)
	}
	return &ProviderMessageRef{ChatRef: strconv.FormatInt(message.Chat.ID, 10)}
}

func telegramLocale(user *telego.User) string {
	if user == nil {
		return "ru"
	}
	locale := strings.TrimSpace(user.LanguageCode)
	if locale == "" {
		return "ru"
	}
	return locale
}

func telegramSTTLanguage(locale string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(locale)), "ru") {
		return "ru"
	}
	return "en"
}

func decorateActionLabel(label string) string {
	label = strings.TrimSpace(label)
	if label == "" {
		return ""
	}
	if strings.ContainsAny(label, "✅❌🔗🟢🟡🔵🟣🟠⚪️⚫️") {
		return label
	}
	return "🔗 " + label
}

func decorateOptionLabel(option DecisionOption, index int) string {
	label := strings.TrimSpace(option.Label)
	if label == "" {
		return ""
	}
	if strings.ContainsAny(label, "✅❌🟢🟡🔵🟣🟠⚪️⚫️❓⏳") {
		return label
	}

	token := strings.ToLower(strings.TrimSpace(option.OptionID + " " + option.Label))
	switch {
	case strings.Contains(token, "approve"), strings.Contains(token, "accept"), strings.Contains(token, "yes"), strings.Contains(token, "ok"), strings.Contains(token, "merge"), strings.Contains(token, "confirm"), strings.Contains(token, "да"), strings.Contains(token, "подтверд"), strings.Contains(token, "одобр"):
		return "✅ " + label
	case strings.Contains(token, "reject"), strings.Contains(token, "deny"), strings.Contains(token, "no"), strings.Contains(token, "cancel"), strings.Contains(token, "decline"), strings.Contains(token, "нет"), strings.Contains(token, "отклон"), strings.Contains(token, "отмен"):
		return "❌ " + label
	case strings.Contains(token, "later"), strings.Contains(token, "wait"), strings.Contains(token, "позже"), strings.Contains(token, "потом"):
		return "⏳ " + label
	case strings.Contains(token, "question"), strings.Contains(token, "help"), strings.Contains(token, "info"), strings.Contains(token, "справ"), strings.Contains(token, "вопрос"):
		return "❓ " + label
	default:
		prefixes := []string{"🟢", "🟡", "🔵", "🟣", "🟠", "⚪️", "⚫️"}
		return prefixes[index%len(prefixes)] + " " + label
	}
}

func followUpTemplateKey(envelope DeliveryEnvelope) string {
	if envelope.Continuation == nil {
		return "follow_up_applied_response"
	}
	switch strings.TrimSpace(envelope.Continuation.Reason) {
	case "edit_failed":
		return "follow_up_edit_failed"
	case "expired_wait":
		return "follow_up_expired_wait"
	case "operator_fallback":
		return "follow_up_operator_fallback"
	default:
		return "follow_up_applied_response"
	}
}

func classifyTelegramDeliveryError(err error) error {
	message := strings.TrimSpace(err.Error())
	statusCode := http.StatusServiceUnavailable
	retryable := true

	var telegramErr *telegoapi.Error
	if errors.As(err, &telegramErr) && telegramErr != nil {
		message = strings.TrimSpace(telegramErr.Description)
		switch telegramErr.ErrorCode {
		case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
			statusCode = http.StatusUnprocessableEntity
			retryable = false
		case http.StatusTooManyRequests:
			statusCode = http.StatusTooManyRequests
			retryable = true
		default:
			if telegramErr.ErrorCode >= http.StatusInternalServerError {
				statusCode = http.StatusServiceUnavailable
			}
		}
	}

	return &DeliveryError{
		StatusCode: statusCode,
		Response: DeliveryResponse{
			Accepted:  false,
			Retryable: retryable,
			Message:   message,
		},
	}
}
