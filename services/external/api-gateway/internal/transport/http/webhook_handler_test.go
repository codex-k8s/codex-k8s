package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/webhook"
)

func TestIngestGitHubWebhook_AcceptAndDuplicate(t *testing.T) {
	t.Parallel()

	secret := "test-secret"
	event := "push"
	deliveryID := "delivery-abc"
	payload := `{"action":"opened","repository":{"id":1,"full_name":"codex-k8s/codex-k8s"}}`

	fake := &fakeWebhookService{
		sequence: []webhook.IngestResult{
			{CorrelationID: deliveryID, RunID: "run-1", Status: "accepted", Duplicate: false},
			{CorrelationID: deliveryID, RunID: "run-1", Status: "duplicate", Duplicate: true},
		},
	}

	h := newWebhookHandler(ServerConfig{
		GitHubWebhookSecret: secret,
		MaxBodyBytes:        1024 * 1024,
	}, fake)

	e := echo.New()
	e.HTTPErrorHandler = newHTTPErrorHandler(slog.New(slog.NewTextHandler(ioDiscard{}, nil)))

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/github", strings.NewReader(payload))
	req1.Header.Set(headerGitHubEvent, event)
	req1.Header.Set(headerGitHubDelivery, deliveryID)
	req1.Header.Set(headerGitHubSignature256, sign(secret, payload))
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := h.IngestGitHubWebhook(c1); err != nil {
		t.Fatalf("unexpected error on first request: %v", err)
	}
	if rec1.Code != http.StatusAccepted {
		t.Fatalf("expected 202 on first request, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/github", strings.NewReader(payload))
	req2.Header.Set(headerGitHubEvent, event)
	req2.Header.Set(headerGitHubDelivery, deliveryID)
	req2.Header.Set(headerGitHubSignature256, sign(secret, payload))
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := h.IngestGitHubWebhook(c2); err != nil {
		t.Fatalf("unexpected error on second request: %v", err)
	}
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 on duplicate request, got %d", rec2.Code)
	}
}

func TestIngestGitHubWebhook_InvalidSignature(t *testing.T) {
	t.Parallel()

	secret := "test-secret"
	payload := `{"action":"opened"}`
	fake := &fakeWebhookService{}

	h := newWebhookHandler(ServerConfig{
		GitHubWebhookSecret: secret,
		MaxBodyBytes:        1024 * 1024,
	}, fake)

	e := echo.New()
	e.HTTPErrorHandler = newHTTPErrorHandler(slog.New(slog.NewTextHandler(ioDiscard{}, nil)))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/github", strings.NewReader(payload))
	req.Header.Set(headerGitHubEvent, "push")
	req.Header.Set(headerGitHubDelivery, "delivery-xyz")
	req.Header.Set(headerGitHubSignature256, "sha256=deadbeef")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.IngestGitHubWebhook(c)
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
	e.HTTPErrorHandler(c, err)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

type fakeWebhookService struct {
	sequence []webhook.IngestResult
	index    int
}

func (f *fakeWebhookService) IngestGitHubWebhook(_ context.Context, _ webhook.IngestCommand) (webhook.IngestResult, error) {
	if len(f.sequence) == 0 {
		return webhook.IngestResult{
			CorrelationID: "delivery-default",
			RunID:         "run-default",
			Status:        "accepted",
		}, nil
	}
	if f.index >= len(f.sequence) {
		return f.sequence[len(f.sequence)-1], nil
	}
	result := f.sequence[f.index]
	f.index++
	return result, nil
}

func sign(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (n int, err error) {
	return len(p), nil
}
