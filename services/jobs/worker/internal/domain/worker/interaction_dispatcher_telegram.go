package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	telegramInteractionAdapterKind             = "telegram"
	telegramInteractionDeliveriesPath          = "/v1/telegram/interaction-deliveries"
	telegramInteractionEditCapabilityUnknown   = "unknown"
	telegramInteractionEditCapabilityEditable  = "editable"
	telegramInteractionEditCapabilityKeyboard  = "keyboard_only"
	telegramInteractionEditCapabilityFollowUp  = "follow_up_only"
	telegramInteractionErrorUnsupportedAdapter = "unsupported_adapter"
	telegramInteractionErrorTransportUnavailable = "transport_unavailable"
	telegramInteractionErrorRejected           = "adapter_rejected"
	telegramInteractionErrorHTTP4xx            = "adapter_http_4xx"
	telegramInteractionErrorHTTP5xx            = "adapter_http_5xx"
	telegramInteractionErrorInvalidResponse    = "invalid_adapter_response"
)

// TelegramInteractionDispatcherConfig configures the worker-side HTTP bridge to the Telegram adapter contour.
type TelegramInteractionDispatcherConfig struct {
	BaseURL     string
	BearerToken string
	Timeout     time.Duration
}

type telegramInteractionDispatcher struct {
	baseURL     string
	bearerToken string
	client      *http.Client
}

type telegramInteractionDeliveryEnvelope struct {
	CallbackEndpoint *struct {
		TokenExpiresAt string `json:"token_expires_at,omitempty"`
	} `json:"callback_endpoint,omitempty"`
}

type telegramInteractionDeliveryResponse struct {
	Accepted          bool            `json:"accepted"`
	AdapterDeliveryID string          `json:"adapter_delivery_id,omitempty"`
	ProviderMessageRef json.RawMessage `json:"provider_message_ref,omitempty"`
	EditCapability    string          `json:"edit_capability,omitempty"`
	Retryable         bool            `json:"retryable"`
	Message           string          `json:"message,omitempty"`
}

// NewTelegramInteractionDispatcher builds the worker-side HTTP dispatcher for Telegram interaction deliveries.
func NewTelegramInteractionDispatcher(cfg TelegramInteractionDispatcherConfig) (InteractionDispatcher, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("telegram interaction adapter base URL is required")
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &telegramInteractionDispatcher{
		baseURL:     baseURL,
		bearerToken: strings.TrimSpace(cfg.BearerToken),
		client: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (d *telegramInteractionDispatcher) Dispatch(ctx context.Context, claim InteractionDispatchClaim) (InteractionDispatchAck, error) {
	ack := InteractionDispatchAck{
		AdapterKind: telegramInteractionAdapterKind,
	}
	if tokenExpiresAt, err := interactionCallbackTokenExpiresAtFromEnvelope(claim.RequestEnvelopeJSON); err == nil {
		ack.CallbackTokenExpiresAt = tokenExpiresAt
	}

	adapterKind := strings.TrimSpace(claim.RecipientProvider)
	if adapterKind == "" {
		adapterKind = strings.TrimSpace(claim.Attempt.AdapterKind)
	}
	if adapterKind != "" && adapterKind != telegramInteractionAdapterKind {
		ack.AdapterKind = adapterKind
		ack.ErrorCode = telegramInteractionErrorUnsupportedAdapter
		return ack, fmt.Errorf("unsupported interaction adapter %q", adapterKind)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.baseURL+telegramInteractionDeliveriesPath, bytes.NewReader(claim.RequestEnvelopeJSON))
	if err != nil {
		ack.ErrorCode = telegramInteractionErrorTransportUnavailable
		ack.Retryable = true
		return ack, err
	}
	req.Header.Set("Content-Type", "application/json")
	if d.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+d.bearerToken)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		ack.ErrorCode = telegramInteractionErrorTransportUnavailable
		ack.Retryable = true
		return ack, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ack.ErrorCode = telegramInteractionErrorTransportUnavailable
		ack.Retryable = true
		return ack, err
	}
	if len(body) > 0 && json.Valid(body) {
		ack.AckPayloadJSON = body
	}

	var response telegramInteractionDeliveryResponse
	if len(body) > 0 {
		if err := json.Unmarshal(body, &response); err != nil {
			ack.ErrorCode = telegramInteractionErrorInvalidResponse
			ack.Retryable = resp.StatusCode >= http.StatusInternalServerError
			return ack, fmt.Errorf("decode telegram interaction delivery response: %w", err)
		}
	}

	ack.AdapterDeliveryID = strings.TrimSpace(response.AdapterDeliveryID)
	ack.ProviderMessageRefJSON = normalizedInteractionProviderMessageRefJSON(response.ProviderMessageRef)
	ack.EditCapability = normalizeTelegramInteractionEditCapability(response.EditCapability)
	ack.Retryable = response.Retryable

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices && response.Accepted {
		return ack, nil
	}

	ack.ErrorCode = interactionDispatchHTTPErrorCode(resp.StatusCode)
	if strings.TrimSpace(response.Message) == "" {
		response.Message = http.StatusText(resp.StatusCode)
	}
	if resp.StatusCode >= http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError {
		ack.Retryable = true
	}
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices && !response.Accepted {
		ack.ErrorCode = telegramInteractionErrorRejected
	}
	return ack, fmt.Errorf("telegram interaction delivery rejected: %s", strings.TrimSpace(response.Message))
}

func interactionCallbackTokenExpiresAtFromEnvelope(raw json.RawMessage) (*time.Time, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var envelope telegramInteractionDeliveryEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, err
	}
	if envelope.CallbackEndpoint == nil || strings.TrimSpace(envelope.CallbackEndpoint.TokenExpiresAt) == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(envelope.CallbackEndpoint.TokenExpiresAt))
	if err != nil {
		return nil, err
	}
	result := parsed.UTC()
	return &result, nil
}

func normalizedInteractionProviderMessageRefJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || !json.Valid(raw) {
		return nil
	}
	return raw
}

func normalizeTelegramInteractionEditCapability(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case telegramInteractionEditCapabilityEditable,
		telegramInteractionEditCapabilityKeyboard,
		telegramInteractionEditCapabilityFollowUp:
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return telegramInteractionEditCapabilityUnknown
	}
}

func interactionDispatchHTTPErrorCode(statusCode int) string {
	if statusCode >= http.StatusInternalServerError {
		return telegramInteractionErrorHTTP5xx
	}
	if statusCode >= http.StatusBadRequest {
		return telegramInteractionErrorHTTP4xx
	}
	return telegramInteractionErrorRejected
}
