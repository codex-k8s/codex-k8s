package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/codex-k8s/kodex/services/external/telegram-interaction-adapter/internal/service"
)

func TestPostTelegramInteractionDelivery_RejectsInvalidBearer(t *testing.T) {
	t.Parallel()

	handler := newHandler(&fakeAdapterService{
		deliveryToken: "expected-token",
	}, defaultMaxBodyBytes, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1/telegram/interaction-deliveries", strings.NewReader(`{}`))
	req.Header.Set(headerAuthorization, "Bearer wrong-token")
	rec := httptest.NewRecorder()

	handler.PostTelegramInteractionDelivery(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	var response map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if accepted, ok := response["accepted"].(bool); !ok || accepted {
		t.Fatalf("accepted = %v, want false", response["accepted"])
	}
}

func TestPostTelegramInteractionDelivery_DelegatesToService(t *testing.T) {
	t.Parallel()

	fake := &fakeAdapterService{
		deliveryToken: "expected-token",
		deliveryResponse: service.DeliveryResponse{
			Accepted:          true,
			AdapterDeliveryID: "primary_dispatch:55",
			EditCapability:    service.EditCapabilityKeyboardOnly,
			Retryable:         false,
			ProviderMessageRef: &service.ProviderMessageRef{
				ChatRef:   stringPtrValue("101"),
				MessageID: stringPtrValue("55"),
			},
		},
	}
	handler := newHandler(fake, defaultMaxBodyBytes, nil)

	body := `{
		"schema_version":"telegram-interaction-v1",
		"delivery_id":"5e76210a-bcdd-4877-973f-70577298d080",
		"delivery_role":"primary_dispatch",
		"interaction_id":"e54f63ce-d9fc-4f85-a8ee-66563d6dff7d",
		"interaction_kind":"notify",
		"recipient_provider":"telegram",
		"recipient_ref":"telegram_chat_id:101",
		"context_links":{"run_id":"run-1"},
		"content":{"summary":"hello"},
		"continuation_policy":{
			"preferred_mode":"edit_in_place_first",
			"disable_keyboard_on_resolution":true,
			"send_follow_up_on_edit_failure":true,
			"manual_fallback_on_follow_up_failure":true
		}
	}`
	req := httptest.NewRequest(http.MethodPost, "/v1/telegram/interaction-deliveries", strings.NewReader(body))
	req.Header.Set(headerAuthorization, "Bearer expected-token")
	rec := httptest.NewRecorder()

	handler.PostTelegramInteractionDelivery(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if fake.lastEnvelope.InteractionID != "e54f63ce-d9fc-4f85-a8ee-66563d6dff7d" {
		t.Fatalf("InteractionID = %q", fake.lastEnvelope.InteractionID)
	}
	if fake.lastEnvelope.Content.Summary != "hello" {
		t.Fatalf("Content.Summary = %q", fake.lastEnvelope.Content.Summary)
	}
}

func TestPostTelegramInteractionWebhook_ValidatesSecretAndBody(t *testing.T) {
	t.Parallel()

	fake := &fakeAdapterService{
		webhookSecret: "secret-token",
	}
	handler := newHandler(fake, defaultMaxBodyBytes, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/interactions/webhook", strings.NewReader(`{"update_id":1}`))
	req.Header.Set(headerTelegramWebhookSecret, "secret-token")
	rec := httptest.NewRecorder()

	handler.PostTelegramInteractionWebhook(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if string(fake.lastWebhookBody) != `{"update_id":1}` {
		t.Fatalf("lastWebhookBody = %s", string(fake.lastWebhookBody))
	}
}

func TestPostTelegramInteractionWebhook_PropagatesFailures(t *testing.T) {
	t.Parallel()

	handler := newHandler(&fakeAdapterService{
		webhookSecret: "secret-token",
		webhookError:  errors.New("callback endpoint unavailable"),
	}, defaultMaxBodyBytes, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/interactions/webhook", strings.NewReader(`{"update_id":1}`))
	req.Header.Set(headerTelegramWebhookSecret, "secret-token")
	rec := httptest.NewRecorder()

	handler.PostTelegramInteractionWebhook(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

type fakeAdapterService struct {
	deliveryToken    string
	webhookSecret    string
	deliveryResponse service.DeliveryResponse
	deliveryError    error
	webhookError     error
	lastEnvelope     service.DeliveryEnvelope
	lastWebhookBody  []byte
}

func (f *fakeAdapterService) DeliveryToken() string {
	return f.deliveryToken
}

func (f *fakeAdapterService) WebhookSecret() string {
	return f.webhookSecret
}

func (f *fakeAdapterService) Deliver(_ context.Context, envelope service.DeliveryEnvelope) (service.DeliveryResponse, error) {
	f.lastEnvelope = envelope
	return f.deliveryResponse, f.deliveryError
}

func (f *fakeAdapterService) HandleWebhook(_ context.Context, body []byte) error {
	f.lastWebhookBody = append([]byte(nil), body...)
	return f.webhookError
}

func stringPtrValue(value string) string {
	return value
}
