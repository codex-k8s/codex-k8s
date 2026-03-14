package observability

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	interactionrequestrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/interactionrequest"
)

const defaultCollectorQueryTimeout = 3 * time.Second

type interactionMetricsSnapshotReader interface {
	LoadMetricsSnapshot(ctx context.Context, now time.Time) (interactionrequestrepo.MetricsSnapshot, error)
}

// InteractionCollector publishes persisted interaction observability metrics from control-plane state.
type InteractionCollector struct {
	reader       interactionMetricsSnapshotReader
	logger       *slog.Logger
	queryTimeout time.Duration

	requestsStateDesc    *prometheus.Desc
	pendingBacklogDesc   *prometheus.Desc
	overdueWaitDesc      *prometheus.Desc
	callbackEventsDesc   *prometheus.Desc
	dispatchAttemptsDesc *prometheus.Desc
}

// NewInteractionCollector constructs interaction metrics collector.
func NewInteractionCollector(reader interactionMetricsSnapshotReader, logger *slog.Logger) *InteractionCollector {
	if logger == nil {
		logger = slog.Default()
	}

	return &InteractionCollector{
		reader:       reader,
		logger:       logger,
		queryTimeout: defaultCollectorQueryTimeout,
		requestsStateDesc: prometheus.NewDesc(
			"codexk8s_interaction_requests_state",
			"Current number of interaction requests by interaction kind and state.",
			[]string{"interaction_kind", "state"},
			nil,
		),
		pendingBacklogDesc: prometheus.NewDesc(
			"codexk8s_interaction_pending_dispatch_backlog",
			"Current number of pending interaction dispatch items split into initial and retry backlog.",
			[]string{"interaction_kind", "queue_kind"},
			nil,
		),
		overdueWaitDesc: prometheus.NewDesc(
			"codexk8s_interaction_wait_deadline_overdue",
			"Current number of interaction waits past deadline without terminal outcome.",
			[]string{"interaction_kind"},
			nil,
		),
		callbackEventsDesc: prometheus.NewDesc(
			"codexk8s_interaction_callback_events_total",
			"Persisted interaction callback evidence counts by callback kind and classification.",
			[]string{"callback_kind", "classification"},
			nil,
		),
		dispatchAttemptsDesc: prometheus.NewDesc(
			"codexk8s_interaction_dispatch_attempts_total",
			"Persisted interaction dispatch attempt counts by interaction kind, adapter kind and status.",
			[]string{"interaction_kind", "adapter_kind", "status"},
			nil,
		),
	}
}

// Describe implements prometheus.Collector.
func (c *InteractionCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.requestsStateDesc
	ch <- c.pendingBacklogDesc
	ch <- c.overdueWaitDesc
	ch <- c.callbackEventsDesc
	ch <- c.dispatchAttemptsDesc
}

// Collect implements prometheus.Collector.
func (c *InteractionCollector) Collect(ch chan<- prometheus.Metric) {
	if c == nil {
		return
	}
	if c.reader == nil {
		ch <- prometheus.NewInvalidMetric(c.requestsStateDesc, context.DeadlineExceeded)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.queryTimeout)
	defer cancel()

	snapshot, err := c.reader.LoadMetricsSnapshot(ctx, time.Now().UTC())
	if err != nil {
		c.logger.Error("collect interaction metrics snapshot failed", "err", err)
		ch <- prometheus.NewInvalidMetric(c.requestsStateDesc, err)
		return
	}

	for _, item := range snapshot.RequestStateTotals {
		ch <- prometheus.MustNewConstMetric(
			c.requestsStateDesc,
			prometheus.GaugeValue,
			float64(item.Total),
			normalizedMetricLabel(item.InteractionKind, "unknown"),
			normalizedMetricLabel(item.State, "unknown"),
		)
	}
	for _, item := range snapshot.PendingDispatchBacklog {
		ch <- prometheus.MustNewConstMetric(
			c.pendingBacklogDesc,
			prometheus.GaugeValue,
			float64(item.Total),
			normalizedMetricLabel(item.InteractionKind, "unknown"),
			normalizedMetricLabel(item.QueueKind, "unknown"),
		)
	}
	for _, item := range snapshot.OverdueWaitTotals {
		ch <- prometheus.MustNewConstMetric(
			c.overdueWaitDesc,
			prometheus.GaugeValue,
			float64(item.Total),
			normalizedMetricLabel(item.InteractionKind, "unknown"),
		)
	}
	for _, item := range snapshot.CallbackEventTotals {
		ch <- prometheus.MustNewConstMetric(
			c.callbackEventsDesc,
			prometheus.CounterValue,
			float64(item.Total),
			normalizedMetricLabel(item.CallbackKind, "unknown"),
			normalizedMetricLabel(item.Classification, "unknown"),
		)
	}
	for _, item := range snapshot.DispatchAttemptTotals {
		ch <- prometheus.MustNewConstMetric(
			c.dispatchAttemptsDesc,
			prometheus.CounterValue,
			float64(item.Total),
			normalizedMetricLabel(item.InteractionKind, "unknown"),
			normalizedMetricLabel(item.AdapterKind, "unknown"),
			normalizedMetricLabel(item.Status, "unknown"),
		)
	}
}

func normalizedMetricLabel(value string, fallback string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return fallback
	}
	return normalized
}
