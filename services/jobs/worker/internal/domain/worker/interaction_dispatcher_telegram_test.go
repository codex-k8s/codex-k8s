package worker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTelegramInteractionDispatcherDispatchAccepted(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != telegramInteractionDeliveriesPath {
			t.Fatalf("path = %q, want %q", r.URL.Path, telegramInteractionDeliveriesPath)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer adapter-token"; got != want {
			t.Fatalf("authorization = %q, want %q", got, want)
		}
		_, _ = w.Write([]byte(`{
			"accepted":true,
			"adapter_delivery_id":"adapter-1",
			"provider_message_ref":{"message_id":"42"},
			"edit_capability":"editable"
		}`))
	}))
	defer server.Close()

	dispatcher, err := NewTelegramInteractionDispatcher(TelegramInteractionDispatcherConfig{
		BaseURL:     server.URL,
		BearerToken: "adapter-token",
		Timeout:     time.Second,
	})
	if err != nil {
		t.Fatalf("NewTelegramInteractionDispatcher returned error: %v", err)
	}

	ack, err := dispatcher.Dispatch(context.Background(), InteractionDispatchClaim{
		RecipientProvider: telegramInteractionAdapterKind,
		RequestEnvelopeJSON: []byte(`{
			"schema_version":"telegram-interaction-v1",
			"callback_endpoint":{"token_expires_at":"2026-03-13T17:00:00Z"}
		}`),
	})
	if err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if ack.AdapterDeliveryID != "adapter-1" {
		t.Fatalf("adapter_delivery_id = %q, want adapter-1", ack.AdapterDeliveryID)
	}
	if ack.EditCapability != telegramInteractionEditCapabilityEditable {
		t.Fatalf("edit_capability = %q, want %q", ack.EditCapability, telegramInteractionEditCapabilityEditable)
	}
	if got, want := strings.TrimSpace(string(ack.ProviderMessageRefJSON)), `{"message_id":"42"}`; got != want {
		t.Fatalf("provider_message_ref_json = %q, want %q", got, want)
	}
	if ack.CallbackTokenExpiresAt == nil || ack.CallbackTokenExpiresAt.UTC().Format(time.RFC3339) != "2026-03-13T17:00:00Z" {
		t.Fatalf("callback_token_expires_at = %v, want 2026-03-13T17:00:00Z", ack.CallbackTokenExpiresAt)
	}
}

func TestTelegramInteractionDispatcherDispatchRejected(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{
			"accepted":false,
			"retryable":true,
			"message":"temporary outage"
		}`))
	}))
	defer server.Close()

	dispatcher, err := NewTelegramInteractionDispatcher(TelegramInteractionDispatcherConfig{
		BaseURL: server.URL,
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewTelegramInteractionDispatcher returned error: %v", err)
	}

	ack, err := dispatcher.Dispatch(context.Background(), InteractionDispatchClaim{
		RecipientProvider: telegramInteractionAdapterKind,
		RequestEnvelopeJSON: []byte(`{
			"schema_version":"telegram-interaction-v1",
			"callback_endpoint":{"token_expires_at":"2026-03-13T17:00:00Z"}
		}`),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !ack.Retryable {
		t.Fatal("retryable = false, want true")
	}
	if ack.ErrorCode != telegramInteractionErrorHTTP5xx {
		t.Fatalf("error_code = %q, want %q", ack.ErrorCode, telegramInteractionErrorHTTP5xx)
	}
}
