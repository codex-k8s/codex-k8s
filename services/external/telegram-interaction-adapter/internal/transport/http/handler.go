package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/codex-k8s/kodex/services/external/telegram-interaction-adapter/internal/service"
	"github.com/codex-k8s/kodex/services/external/telegram-interaction-adapter/internal/transport/http/casters"
	"github.com/codex-k8s/kodex/services/external/telegram-interaction-adapter/internal/transport/http/models"
)

const (
	headerAuthorization         = "Authorization"
	headerTelegramWebhookSecret = "X-Telegram-Bot-Api-Secret-Token"
	httpAuthPrefixBearer        = "Bearer "
)

type adapterService interface {
	DeliveryToken() string
	WebhookSecret() string
	Deliver(context.Context, service.DeliveryEnvelope) (service.DeliveryResponse, error)
	HandleWebhook(context.Context, []byte) error
}

type handler struct {
	svc          adapterService
	maxBodyBytes int64
	logger       *slog.Logger
}

func newHandler(svc adapterService, maxBodyBytes int64, logger *slog.Logger) *handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &handler{
		svc:          svc,
		maxBodyBytes: maxBodyBytes,
		logger:       logger,
	}
}

func (h *handler) PostTelegramInteractionDelivery(w http.ResponseWriter, r *http.Request) {
	deliveryToken := strings.TrimSpace(h.svc.DeliveryToken())
	if deliveryToken != "" && resolveBearerToken(r.Header.Get(headerAuthorization)) != deliveryToken {
		writeJSON(w, http.StatusUnauthorized, models.TelegramInteractionDeliveryResponse{
			Accepted:  false,
			Retryable: false,
			Message:   stringPtr("invalid telegram interaction adapter bearer token"),
		})
		return
	}

	body, err := readRequestBody(w, r, h.maxBodyBytes)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.TelegramInteractionDeliveryResponse{
			Accepted:  false,
			Retryable: false,
			Message:   stringPtr(err.Error()),
		})
		return
	}

	var request models.TelegramInteractionDeliveryEnvelope
	if err := decodeStrictJSON(body, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, models.TelegramInteractionDeliveryResponse{
			Accepted:  false,
			Retryable: false,
			Message:   stringPtr(fmt.Sprintf("decode telegram delivery request: %v", err)),
		})
		return
	}

	response, err := h.svc.Deliver(r.Context(), casters.DeliveryEnvelope(request))
	if err != nil {
		var deliveryErr *service.DeliveryError
		if errors.As(err, &deliveryErr) && deliveryErr != nil {
			writeJSON(w, deliveryErr.StatusCode, casters.DeliveryResponse(deliveryErr.Response))
			return
		}
		h.logger.Error("telegram delivery failed", "err", err)
		writeJSON(w, http.StatusServiceUnavailable, models.TelegramInteractionDeliveryResponse{
			Accepted:  false,
			Retryable: true,
			Message:   stringPtr("telegram interaction delivery failed"),
		})
		return
	}

	writeJSON(w, http.StatusOK, casters.DeliveryResponse(response))
}

func (h *handler) PostTelegramInteractionWebhook(w http.ResponseWriter, r *http.Request) {
	webhookSecret := strings.TrimSpace(h.svc.WebhookSecret())
	if webhookSecret == "" || strings.TrimSpace(r.Header.Get(headerTelegramWebhookSecret)) != webhookSecret {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := readRequestBody(w, r, h.maxBodyBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.HandleWebhook(r.Context(), body); err != nil {
		h.logger.Error("telegram webhook handling failed", "err", err)
		http.Error(w, "telegram webhook handling failed", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func readRequestBody(w http.ResponseWriter, r *http.Request, maxBodyBytes int64) ([]byte, error) {
	if r == nil {
		return nil, fmt.Errorf("request is nil")
	}
	reader := r.Body
	if maxBodyBytes > 0 {
		reader = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	}
	defer func() { _ = reader.Close() }()
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("request body is required")
	}
	return body, nil
}

func decodeStrictJSON(body []byte, output any) error {
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(output); err != nil {
		return err
	}
	if decoder.More() {
		return fmt.Errorf("unexpected trailing JSON tokens")
	}
	return nil
}

func resolveBearerToken(authorizationHeader string) string {
	authorization := strings.TrimSpace(authorizationHeader)
	if strings.HasPrefix(strings.ToLower(authorization), strings.ToLower(httpAuthPrefixBearer)) {
		return strings.TrimSpace(authorization[len(httpAuthPrefixBearer):])
	}
	return ""
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func stringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
