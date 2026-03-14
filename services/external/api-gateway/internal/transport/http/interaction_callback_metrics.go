package http

import (
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	interactionCallbackMetricError   = "error"
	interactionCallbackMetricUnknown = "unknown"
)

var (
	interactionCallbackRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "codexk8s_interaction_callback_requests_total",
			Help: "Total number of interaction callback HTTP requests handled by api-gateway.",
		},
		[]string{"callback_kind", "classification"},
	)
	interactionCallbackDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "codexk8s_interaction_callback_duration_seconds",
			Help:    "Duration of interaction callback handling in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"callback_kind", "classification"},
	)
)

func recordInteractionCallbackMetrics(callbackKind string, classification string, startedAt time.Time) {
	kindLabel := normalizeInteractionCallbackMetricLabel(callbackKind)
	classificationLabel := normalizeInteractionCallbackMetricLabel(classification)
	interactionCallbackRequestsTotal.WithLabelValues(kindLabel, classificationLabel).Inc()
	interactionCallbackDuration.WithLabelValues(kindLabel, classificationLabel).Observe(time.Since(startedAt).Seconds())
}

func normalizeInteractionCallbackMetricLabel(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		return interactionCallbackMetricUnknown
	}
	return normalized
}
