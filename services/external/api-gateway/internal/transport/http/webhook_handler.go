package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/codex-k8s/codex-k8s/libs/go/crypto/githubsignature"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/errs"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/webhook"
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
	webhookService webhookIngress
	secret         []byte
	maxBodyBytes   int64
}

func newWebhookHandler(cfg ServerConfig, webhookService webhookIngress) *webhookHandler {
	return &webhookHandler{
		webhookService: webhookService,
		secret:         []byte(cfg.GitHubWebhookSecret),
		maxBodyBytes:   cfg.MaxBodyBytes,
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

	result, err := h.webhookService.IngestGitHubWebhook(ctx, webhook.IngestCommand{
		CorrelationID: deliveryID,
		EventType:     eventType,
		DeliveryID:    deliveryID,
		ReceivedAt:    startedAt,
		Payload:       rawPayload,
	})
	if err != nil {
		webhookRequestsTotal.WithLabelValues("error", eventType).Inc()
		webhookDuration.WithLabelValues("error", eventType).Observe(time.Since(startedAt).Seconds())
		return fmt.Errorf("ingest github webhook: %w", err)
	}

	if result.Duplicate {
		webhookRequestsTotal.WithLabelValues("duplicate", eventType).Inc()
		webhookDuration.WithLabelValues("duplicate", eventType).Observe(time.Since(startedAt).Seconds())
		return c.JSON(http.StatusOK, result)
	}

	webhookRequestsTotal.WithLabelValues("accepted", eventType).Inc()
	webhookDuration.WithLabelValues("accepted", eventType).Observe(time.Since(startedAt).Seconds())
	return c.JSON(http.StatusAccepted, result)
}

func readRequestBody(body io.ReadCloser, maxBodyBytes int64) ([]byte, error) {
	defer func() { _ = body.Close() }()

	if maxBodyBytes <= 0 {
		maxBodyBytes = 1024 * 1024
	}

	limitedReader := io.LimitReader(body, maxBodyBytes+1)
	payload, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if int64(len(payload)) > maxBodyBytes {
		return nil, errs.Validation{
			Field: "body",
			Msg:   fmt.Sprintf("payload too large (max %d bytes)", maxBodyBytes),
		}
	}
	return payload, nil
}
