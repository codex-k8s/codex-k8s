package mcp

import (
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

const interactionMetricUnknownLabel = "unknown"

var (
	interactionRequestsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kodex_interaction_requests_created_total",
			Help: "Total interaction requests accepted by control-plane MCP tools.",
		},
		[]string{"tool_name", "interaction_kind"},
	)
	interactionResumeTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kodex_interaction_resume_total",
			Help: "Total interaction-driven resume payloads published by request status.",
		},
		[]string{"interaction_kind", "request_status"},
	)
	interactionDecisionTurnaround = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kodex_interaction_decision_turnaround_seconds",
			Help:    "Decision interaction turnaround from request creation to terminal outcome.",
			Buckets: prometheus.ExponentialBuckets(1, 2, 12),
		},
		[]string{"request_status"},
	)
)

func observeInteractionRequestCreated(request entitytypes.InteractionRequest, toolName ToolName) {
	interactionRequestsCreatedTotal.WithLabelValues(
		interactionMetricLabel(string(toolName)),
		interactionMetricLabel(string(request.InteractionKind)),
	).Inc()
}

func observeInteractionResume(request entitytypes.InteractionRequest, payload *valuetypes.InteractionResumePayload) {
	if payload == nil {
		return
	}

	requestStatus := interactionMetricLabel(string(payload.RequestStatus))
	interactionResumeTotal.WithLabelValues(
		interactionMetricLabel(string(request.InteractionKind)),
		requestStatus,
	).Inc()

	if request.InteractionKind != enumtypes.InteractionKindDecisionRequest || request.CreatedAt.IsZero() {
		return
	}

	resolvedAt, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(payload.ResolvedAt))
	if err != nil {
		return
	}
	turnaround := resolvedAt.UTC().Sub(request.CreatedAt.UTC())
	if turnaround < 0 {
		return
	}
	interactionDecisionTurnaround.WithLabelValues(requestStatus).Observe(turnaround.Seconds())
}

func interactionMetricLabel(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return interactionMetricUnknownLabel
	}
	return normalized
}
