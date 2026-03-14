package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const telegramInteractionErrorAdapterNotConfigured = "adapter_not_configured"

// TelegramInteractionErrorAdapterNotConfigured returns the canonical error code for missing adapter wiring.
func TelegramInteractionErrorAdapterNotConfigured() string {
	return telegramInteractionErrorAdapterNotConfigured
}

type unavailableInteractionDispatcher struct {
	adapterKind string
	errorCode   string
	message     string
}

type unavailableInteractionDispatcherPayload struct {
	Accepted  bool   `json:"accepted"`
	Retryable bool   `json:"retryable"`
	ErrorCode string `json:"error_code"`
	Message   string `json:"message,omitempty"`
}

// NewUnavailableInteractionDispatcher returns a dispatcher that reports a typed non-retryable failure.
func NewUnavailableInteractionDispatcher(adapterKind string, errorCode string, message string) InteractionDispatcher {
	if strings.TrimSpace(adapterKind) == "" {
		adapterKind = telegramInteractionAdapterKind
	}
	if strings.TrimSpace(errorCode) == "" {
		errorCode = telegramInteractionErrorAdapterNotConfigured
	}
	return unavailableInteractionDispatcher{
		adapterKind: strings.TrimSpace(adapterKind),
		errorCode:   strings.TrimSpace(errorCode),
		message:     strings.TrimSpace(message),
	}
}

func (d unavailableInteractionDispatcher) Dispatch(context.Context, InteractionDispatchClaim) (InteractionDispatchAck, error) {
	payload, _ := json.Marshal(unavailableInteractionDispatcherPayload{
		Accepted:  false,
		Retryable: false,
		ErrorCode: d.errorCode,
		Message:   d.message,
	})
	ack := InteractionDispatchAck{
		AdapterKind:    d.adapterKind,
		AckPayloadJSON: payload,
		ErrorCode:      d.errorCode,
	}
	return ack, fmt.Errorf("%s", d.message)
}
