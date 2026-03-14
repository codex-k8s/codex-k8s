package service

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricLabelUnknown    = "unknown"
	metricStatusAccepted  = "accepted"
	metricStatusRejected  = "rejected"
	metricStatusFailed    = "failed"
	metricStatusForwarded = "forwarded"
	metricStatusIgnored   = "ignored"
)

var (
	telegramInteractionDispatchAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "codexk8s_telegram_interaction_dispatch_attempt_total",
			Help: "Total number of Telegram interaction delivery attempts handled by the adapter.",
		},
		[]string{"delivery_role", "status"},
	)
	telegramInteractionCallbackTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "codexk8s_telegram_interaction_callback_total",
			Help: "Total number of Telegram callback and free-text webhook events handled by the adapter.",
		},
		[]string{"callback_kind", "classification"},
	)
	telegramInteractionContinuationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "codexk8s_telegram_interaction_continuation_total",
			Help: "Total number of Telegram continuation actions processed by the adapter.",
		},
		[]string{"action_kind", "status"},
	)
)

func recordDispatchAttempt(deliveryRole string, status string) {
	telegramInteractionDispatchAttemptsTotal.WithLabelValues(
		normalizeMetricLabel(deliveryRole),
		normalizeMetricLabel(status),
	).Inc()
}

func recordCallbackEvent(callbackKind string, classification string) {
	telegramInteractionCallbackTotal.WithLabelValues(
		normalizeMetricLabel(callbackKind),
		normalizeMetricLabel(classification),
	).Inc()
}

func recordContinuationAttempt(actionKind string, status string) {
	telegramInteractionContinuationTotal.WithLabelValues(
		normalizeMetricLabel(actionKind),
		normalizeMetricLabel(status),
	).Inc()
}

func normalizeMetricLabel(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		return metricLabelUnknown
	}
	return normalized
}
