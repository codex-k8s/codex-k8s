package worker

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const interactionMetricUnknownLabel = "unknown"

var (
	interactionDispatchAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kodex_interaction_dispatch_attempt_total",
			Help: "Total number of interaction delivery attempts grouped by adapter and completion status.",
		},
		[]string{"adapter", "status"},
	)

	interactionDispatchRetryScheduledTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kodex_interaction_dispatch_retry_scheduled_total",
			Help: "Total number of retry schedules created for interaction delivery attempts.",
		},
		[]string{"adapter", "error_code"},
	)
)

func init() {
	for _, status := range []string{
		interactionAttemptStatusAccepted,
		interactionAttemptStatusFailed,
		interactionAttemptStatusExhausted,
	} {
		interactionDispatchAttemptsTotal.WithLabelValues(noopInteractionAdapterKind, status)
	}

	for _, errorCode := range []string{
		"dispatch_failed",
		"transport_unavailable",
	} {
		interactionDispatchRetryScheduledTotal.WithLabelValues(noopInteractionAdapterKind, errorCode)
	}
}

func recordInteractionDispatchAttempt(adapter string, status string) {
	interactionDispatchAttemptsTotal.WithLabelValues(
		normalizeInteractionWorkerLabel(adapter),
		normalizeInteractionWorkerLabel(status),
	).Inc()
}

func recordInteractionDispatchRetryScheduled(adapter string, errorCode string) {
	interactionDispatchRetryScheduledTotal.WithLabelValues(
		normalizeInteractionWorkerLabel(adapter),
		normalizeInteractionWorkerLabel(errorCode),
	).Inc()
}

func normalizeInteractionWorkerLabel(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return interactionMetricUnknownLabel
	}
	return trimmed
}
