package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	SessionStore   SessionStore
	Recipients     *RecipientResolver
	Bot            BotClient
	CallbackClient *http.Client
	Logger         *slog.Logger
}

// Service owns Telegram transport logic and callback forwarding.
type Service struct {
	publicBaseURL string
	webhookSecret string
	deliveryToken string
	sessions      SessionStore
	recipients    *RecipientResolver
	bot           BotClient
	callbacks     *http.Client
	messages      *messageRenderer
	logger        *slog.Logger
}

// New builds the adapter service.
func New(cfg Config) (*Service, error) {
	renderer, err := newMessageRenderer()
	if err != nil {
		return nil, err
	}
	if cfg.CallbackClient == nil {
		cfg.CallbackClient = &http.Client{Timeout: 10 * time.Second}
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.SessionStore == nil {
		return nil, fmt.Errorf("session store is required")
	}
	if cfg.Recipients == nil {
		return nil, fmt.Errorf("recipient resolver is required")
	}
	if cfg.Bot == nil {
		return nil, fmt.Errorf("telegram bot client is required")
	}

	return &Service{
		publicBaseURL: strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/"),
		webhookSecret: strings.TrimSpace(cfg.WebhookSecret),
		deliveryToken: strings.TrimSpace(cfg.DeliveryToken),
		sessions:      cfg.SessionStore,
		recipients:    cfg.Recipients,
		bot:           cfg.Bot,
		callbacks:     cfg.CallbackClient,
		messages:      renderer,
		logger:        cfg.Logger,
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
	if err := s.sessions.CleanupExpired(time.Now().UTC()); err != nil {
		s.logger.Warn("cleanup adapter sessions failed", "err", err)
	}
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

	text, inlineOptions, actionLabel, actionURL, err := buildPrimaryMessage(envelope)
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

	providerRef := normalizeTelegramProviderMessageRef(sent.ChatID, sent.MessageID, sent.SentAt)
	response := DeliveryResponse{
		Accepted:           true,
		AdapterDeliveryID:  buildAdapterDeliveryID(envelope.DeliveryRole, sent.MessageID),
		ProviderMessageRef: providerRef,
		EditCapability:     resolveEditCapability(envelope),
		Retryable:          false,
	}

	if envelope.InteractionKind == InteractionKindDecisionRequest && envelope.CallbackEndpoint != nil {
		session := buildSessionRecord(envelope, sent, providerRef)
		if err := s.sessions.Upsert(session); err != nil {
			recordDispatchAttempt(envelope.DeliveryRole, metricStatusFailed)
			return DeliveryResponse{}, &DeliveryError{
				StatusCode: http.StatusServiceUnavailable,
				Response: DeliveryResponse{
					Accepted:  false,
					Retryable: true,
					Message:   fmt.Sprintf("persist adapter session: %v", err),
				},
			}
		}
	}

	recordDispatchAttempt(envelope.DeliveryRole, metricStatusAccepted)
	return response, nil
}

func (s *Service) deliverMessageEdit(ctx context.Context, envelope DeliveryEnvelope) (DeliveryResponse, error) {
	messageRef, err := s.resolveProviderMessageRef(envelope)
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
	chatID, err := s.resolveFollowUpChatID(envelope)
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
	if err := s.sessions.CleanupExpired(time.Now().UTC()); err != nil {
		s.logger.Warn("cleanup adapter sessions failed", "err", err)
	}

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

	chatID := int64(0)
	if query.Message != nil {
		chatID = query.Message.GetChat().ID
	}
	if chatID != 0 && !s.recipients.IsAllowedChat(chatID) {
		recordCallbackEvent(CallbackKindOptionSelected, metricStatusIgnored)
		return s.bot.AnswerCallbackQuery(ctx, AnswerCallbackQueryRequest{
			QueryID: query.ID,
			Text:    s.messages.Render("ru", "callback_ack_unavailable", nil),
		})
	}

	handle := strings.TrimSpace(query.Data)
	session, found := s.sessions.GetByHandle(handle)
	if !found {
		recordCallbackEvent(CallbackKindOptionSelected, metricStatusIgnored)
		return s.bot.AnswerCallbackQuery(ctx, AnswerCallbackQueryRequest{
			QueryID: query.ID,
			Text:    s.messages.Render("ru", "callback_ack_unavailable", nil),
		})
	}

	if err := s.bot.AnswerCallbackQuery(ctx, AnswerCallbackQueryRequest{
		QueryID: query.ID,
		Text:    s.messages.Render(session.Locale, "callback_ack_received", nil),
	}); err != nil {
		s.logger.Warn("answer telegram callback query failed", "interaction_id", session.InteractionID, "err", err)
	}

	envelope := CallbackEnvelope{
		SchemaVersion:           SchemaVersionTelegramInteractionV1,
		InteractionID:           session.InteractionID,
		DeliveryID:              session.DeliveryID,
		AdapterEventID:          "callback:" + strings.TrimSpace(query.ID),
		CallbackKind:            CallbackKindOptionSelected,
		OccurredAt:              time.Now().UTC().Format(time.RFC3339Nano),
		CallbackHandle:          handle,
		ResponderRef:            buildResponderRef(&query.From),
		ProviderMessageRef:      &session.ProviderMessageRef,
		ProviderUpdateID:        strconv.Itoa(update.UpdateID),
		ProviderCallbackQueryID: strings.TrimSpace(query.ID),
	}
	outcome, err := s.forwardCallback(ctx, session, envelope)
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

	freeText := strings.TrimSpace(message.Text)
	if freeText == "" {
		return nil
	}

	var (
		session SessionRecord
		found   bool
	)
	if message.ReplyToMessage != nil {
		session, found = s.sessions.GetByReply(message.Chat.ID, strconv.Itoa(message.ReplyToMessage.MessageID))
	}
	if !found {
		session, found = s.sessions.GetSingleOpenByChat(message.Chat.ID)
	}
	if !found || strings.TrimSpace(session.FreeTextHandle) == "" {
		recordCallbackEvent(CallbackKindFreeTextReceived, metricStatusIgnored)
		_, _ = s.bot.SendMessage(ctx, SendMessageRequest{
			ChatID: message.Chat.ID,
			Text:   s.messages.Render("ru", "free_text_unavailable", nil),
		})
		return nil
	}

	envelope := CallbackEnvelope{
		SchemaVersion:      SchemaVersionTelegramInteractionV1,
		InteractionID:      session.InteractionID,
		DeliveryID:         session.DeliveryID,
		AdapterEventID:     "message:" + strconv.Itoa(update.UpdateID),
		CallbackKind:       CallbackKindFreeTextReceived,
		OccurredAt:         time.Now().UTC().Format(time.RFC3339Nano),
		CallbackHandle:     session.FreeTextHandle,
		FreeText:           freeText,
		ResponderRef:       buildResponderRef(message.From),
		ProviderMessageRef: &session.ProviderMessageRef,
		ProviderUpdateID:   strconv.Itoa(update.UpdateID),
	}

	outcome, err := s.forwardCallback(ctx, session, envelope)
	if err != nil {
		recordCallbackEvent(CallbackKindFreeTextReceived, metricStatusFailed)
		_, _ = s.bot.SendMessage(ctx, SendMessageRequest{
			ChatID: message.Chat.ID,
			Text:   s.messages.Render(session.Locale, "free_text_failed", nil),
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
		Text:   s.messages.Render(session.Locale, confirmationKey, nil),
	})
	recordCallbackEvent(CallbackKindFreeTextReceived, outcome.Classification)
	return nil
}

func (s *Service) forwardCallback(ctx context.Context, session SessionRecord, envelope CallbackEnvelope) (CallbackOutcome, error) {
	payload, err := json.Marshal(envelope)
	if err != nil {
		return CallbackOutcome{}, fmt.Errorf("marshal callback envelope: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, session.CallbackURL, bytes.NewReader(payload))
	if err != nil {
		return CallbackOutcome{}, fmt.Errorf("create callback request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(session.CallbackBearerToken) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(session.CallbackBearerToken))
	}

	resp, err := s.callbacks.Do(req)
	if err != nil {
		return CallbackOutcome{}, fmt.Errorf("post callback envelope: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CallbackOutcome{}, fmt.Errorf("read callback response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return CallbackOutcome{}, fmt.Errorf("callback endpoint returned status %d", resp.StatusCode)
	}

	var outcome CallbackOutcome
	if len(body) != 0 {
		if err := json.Unmarshal(body, &outcome); err != nil {
			return CallbackOutcome{}, fmt.Errorf("decode callback outcome: %w", err)
		}
	}
	s.logger.Info(
		"telegram callback forwarded",
		"interaction_id", session.InteractionID,
		"delivery_id", session.DeliveryID,
		"classification", outcome.Classification,
		"resume_required", outcome.ResumeRequired,
	)
	return outcome, nil
}

func buildPrimaryMessage(envelope DeliveryEnvelope) (string, []InlineOption, string, string, error) {
	switch envelope.InteractionKind {
	case InteractionKindNotify:
		if strings.TrimSpace(envelope.Content.Summary) == "" {
			return "", nil, "", "", fmt.Errorf("notify content summary is required")
		}
		return joinNonEmptyParts(
			envelope.Content.Summary,
			envelope.Content.DetailsMarkdown,
			envelope.ContextLinks.RunURL,
			envelope.ContextLinks.IssueURL,
			envelope.ContextLinks.PullRequestURL,
		), nil, envelope.Content.ActionLabel, envelope.Content.ActionURL, nil
	case InteractionKindDecisionRequest:
		if strings.TrimSpace(envelope.Content.Question) == "" {
			return "", nil, "", "", fmt.Errorf("decision content question is required")
		}
		options := make([]InlineOption, 0, len(envelope.Content.Options))
		for _, option := range envelope.Content.Options {
			if strings.TrimSpace(option.Label) == "" || strings.TrimSpace(option.CallbackHandle) == "" {
				return "", nil, "", "", fmt.Errorf("decision options require label and callback_handle")
			}
			options = append(options, InlineOption{
				Label:        option.Label,
				CallbackData: option.CallbackHandle,
			})
		}
		return joinNonEmptyParts(
			envelope.Content.Question,
			envelope.Content.DetailsMarkdown,
			envelope.Content.ReplyInstruction,
			envelope.ContextLinks.RunURL,
			envelope.ContextLinks.IssueURL,
			envelope.ContextLinks.PullRequestURL,
		), options, "", "", nil
	default:
		return "", nil, "", "", fmt.Errorf("unsupported interaction_kind %q", envelope.InteractionKind)
	}
}

func buildSessionRecord(envelope DeliveryEnvelope, sent SentMessage, providerRef *ProviderMessageRef) SessionRecord {
	record := SessionRecord{
		InteractionID:       envelope.InteractionID,
		DeliveryID:          envelope.DeliveryID,
		RecipientRef:        envelope.RecipientRef,
		Locale:              envelope.Locale,
		CallbackURL:         envelope.CallbackEndpoint.URL,
		CallbackBearerToken: envelope.CallbackEndpoint.BearerToken,
		ChatID:              sent.ChatID,
		PrimaryMessageID:    strconv.Itoa(sent.MessageID),
		ProviderMessageRef:  *providerRef,
		OptionHandleHashes:  map[string]time.Time{},
		ExpiresAt:           sent.SentAt.UTC(),
	}
	for _, handle := range envelope.CallbackEndpoint.Handles {
		switch strings.TrimSpace(handle.HandleKind) {
		case HandleKindOption:
			record.OptionHandleHashes[hashInteractionHandle(handle.Handle)] = handle.ExpiresAt.UTC()
			if handle.ExpiresAt.After(record.ExpiresAt) {
				record.ExpiresAt = handle.ExpiresAt.UTC()
			}
		case HandleKindFreeTextSession:
			record.FreeTextHandle = handle.Handle
			record.FreeTextHandleHash = hashInteractionHandle(handle.Handle)
			expiresAt := handle.ExpiresAt.UTC()
			record.FreeTextExpiresAt = &expiresAt
			if handle.ExpiresAt.After(record.ExpiresAt) {
				record.ExpiresAt = handle.ExpiresAt.UTC()
			}
		}
	}
	return record
}

func (s *Service) resolveProviderMessageRef(envelope DeliveryEnvelope) (*ProviderMessageRef, error) {
	if envelope.ProviderMessageRef != nil && strings.TrimSpace(envelope.ProviderMessageRef.MessageID) != "" && strings.TrimSpace(envelope.ProviderMessageRef.ChatRef) != "" {
		return envelope.ProviderMessageRef, nil
	}
	if session, found := s.sessions.GetByInteractionID(envelope.InteractionID); found {
		return &session.ProviderMessageRef, nil
	}
	return nil, fmt.Errorf("provider_message_ref is required for continuation")
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

func (s *Service) resolveFollowUpChatID(envelope DeliveryEnvelope) (int64, error) {
	if envelope.ProviderMessageRef != nil && strings.TrimSpace(envelope.ProviderMessageRef.ChatRef) != "" {
		return strconv.ParseInt(strings.TrimSpace(envelope.ProviderMessageRef.ChatRef), 10, 64)
	}
	if session, found := s.sessions.GetByInteractionID(envelope.InteractionID); found {
		return session.ChatID, nil
	}
	return 0, fmt.Errorf("continuation provider chat_ref is required")
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

func joinNonEmptyParts(parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			nonEmpty = append(nonEmpty, trimmed)
		}
	}
	return strings.Join(nonEmpty, "\n\n")
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
