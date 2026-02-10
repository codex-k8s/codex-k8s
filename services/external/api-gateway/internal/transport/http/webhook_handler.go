package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/codex-k8s/codex-k8s/libs/go/crypto/githubsignature"
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
)

const (
	headerGitHubEvent        = "X-GitHub-Event"
	headerGitHubDelivery     = "X-GitHub-Delivery"
	headerGitHubSignature256 = "X-Hub-Signature-256"
)

var (
	webhookRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "codexk8s_webhook_requests_total",
			Help: "Total number of GitHub webhook requests handled by api-gateway.",
		},
		[]string{"result", "event"},
	)

	webhookDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "codexk8s_webhook_duration_seconds",
			Help:    "Duration of GitHub webhook ingestion in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"result", "event"},
	)
)

type webhookHandler struct {
	cp           webhookIngress
	secret       []byte
	maxBodyBytes int64
}

type webhookIngress interface {
	IngestGitHubWebhook(ctx context.Context, correlationID string, eventType string, deliveryID string, receivedAt time.Time, payloadJSON []byte) (*controlplanev1.IngestGitHubWebhookResponse, error)
}

func newWebhookHandler(cfg ServerConfig, cp webhookIngress) *webhookHandler {
	return &webhookHandler{
		cp:           cp,
		secret:       []byte(cfg.GitHubWebhookSecret),
		maxBodyBytes: cfg.MaxBodyBytes,
	}
}

func (h *webhookHandler) IngestGitHubWebhook(c *echo.Context) error {
	startedAt := time.Now().UTC()
	req := c.Request()
	ctx := req.Context()

	deliveryID := req.Header.Get(headerGitHubDelivery)
	if deliveryID == "" {
		return errs.Validation{Field: "X-GitHub-Delivery", Msg: "header is required"}
	}

	eventType := req.Header.Get(headerGitHubEvent)
	if eventType == "" {
		return errs.Validation{Field: "X-GitHub-Event", Msg: "header is required"}
	}

	signature := req.Header.Get(headerGitHubSignature256)
	if signature == "" {
		return errs.Unauthorized{Msg: "missing webhook signature"}
	}

	rawPayload, err := readRequestBody(req.Body, h.maxBodyBytes)
	if err != nil {
		return err
	}

	if err := githubsignature.VerifySHA256(h.secret, rawPayload, signature); err != nil {
		return errs.Unauthorized{Msg: "invalid webhook signature"}
	}

	if !json.Valid(rawPayload) {
		return errs.Validation{Field: "body", Msg: "payload must be valid JSON"}
	}

	if h.cp == nil {
		return errs.Unauthorized{Msg: "webhook ingress misconfigured"}
	}
	result, err := h.cp.IngestGitHubWebhook(ctx, deliveryID, eventType, deliveryID, startedAt, rawPayload)
	if err != nil {
		webhookRequestsTotal.WithLabelValues("error", eventType).Inc()
		webhookDuration.WithLabelValues("error", eventType).Observe(time.Since(startedAt).Seconds())
		return err
	}

	if result.GetDuplicate() {
		webhookRequestsTotal.WithLabelValues("duplicate", eventType).Inc()
		webhookDuration.WithLabelValues("duplicate", eventType).Observe(time.Since(startedAt).Seconds())
		return c.JSON(http.StatusOK, map[string]any{
			"correlation_id": result.GetCorrelationId(),
			"run_id":         result.GetRunId(),
			"status":         result.GetStatus(),
			"duplicate":      true,
		})
	}

	webhookRequestsTotal.WithLabelValues("accepted", eventType).Inc()
	webhookDuration.WithLabelValues("accepted", eventType).Observe(time.Since(startedAt).Seconds())
	return c.JSON(http.StatusAccepted, map[string]any{
		"correlation_id": result.GetCorrelationId(),
		"run_id":         result.GetRunId(),
		"status":         result.GetStatus(),
		"duplicate":      false,
	})
}

// readRequestBody lives in request_body.go (shared by webhook + potential future public endpoints).
