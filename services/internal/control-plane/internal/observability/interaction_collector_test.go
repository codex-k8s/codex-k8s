package observability

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	interactionrequestrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/repository/postgres/interactionrequest"
)

func TestInteractionCollectorCollectsSnapshotMetrics(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collector := NewInteractionCollector(fakeInteractionMetricsSnapshotReader{
		snapshot: interactionrequestrepo.MetricsSnapshot{
			RequestStateTotals: []interactionrequestrepo.RequestStateTotal{
				{InteractionKind: "decision_request", State: "open", Total: 2},
			},
			PendingDispatchBacklog: []interactionrequestrepo.PendingDispatchBacklog{
				{InteractionKind: "decision_request", QueueKind: "retry", Total: 1},
			},
			OverdueWaitTotals: []interactionrequestrepo.OverdueWaitTotal{
				{InteractionKind: "decision_request", Total: 3},
			},
			CallbackEventTotals: []interactionrequestrepo.CallbackEventTotal{
				{CallbackKind: "decision_response", Classification: "duplicate", Total: 4},
			},
			DispatchAttemptTotals: []interactionrequestrepo.DispatchAttemptTotal{
				{InteractionKind: "decision_request", AdapterKind: "noop", Status: "failed", Total: 5},
			},
		},
	}, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	if err := registry.Register(collector); err != nil {
		t.Fatalf("register collector: %v", err)
	}

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	var builder strings.Builder
	for _, family := range metricFamilies {
		builder.WriteString(family.GetName())
		for _, metric := range family.GetMetric() {
			builder.WriteString("{")
			for idx, label := range metric.GetLabel() {
				if idx > 0 {
					builder.WriteString(",")
				}
				builder.WriteString(label.GetName())
				builder.WriteString("=\"")
				builder.WriteString(label.GetValue())
				builder.WriteString("\"")
			}
			builder.WriteString("} ")
			if metric.Gauge != nil {
				builder.WriteString(fmt.Sprintf("%g", metric.Gauge.GetValue()))
			}
			if metric.Counter != nil {
				builder.WriteString(fmt.Sprintf("%g", metric.Counter.GetValue()))
			}
			builder.WriteString("\n")
		}
	}

	rendered := builder.String()
	assertMetricContains(t, rendered, `kodex_interaction_requests_state{interaction_kind="decision_request",state="open"} 2`)
	assertMetricContains(t, rendered, `kodex_interaction_pending_dispatch_backlog{interaction_kind="decision_request",queue_kind="retry"} 1`)
	assertMetricContains(t, rendered, `kodex_interaction_wait_deadline_overdue{interaction_kind="decision_request"} 3`)
	assertMetricContains(t, rendered, `kodex_interaction_callback_events_total{callback_kind="decision_response",classification="duplicate"} 4`)
	assertMetricContains(t, rendered, `kodex_interaction_dispatch_attempts_total{adapter_kind="noop",interaction_kind="decision_request",status="failed"} 5`)
}

type fakeInteractionMetricsSnapshotReader struct {
	snapshot interactionrequestrepo.MetricsSnapshot
	err      error
}

func (f fakeInteractionMetricsSnapshotReader) LoadMetricsSnapshot(context.Context, time.Time) (interactionrequestrepo.MetricsSnapshot, error) {
	return f.snapshot, f.err
}

func assertMetricContains(t *testing.T, rendered string, want string) {
	t.Helper()

	if !strings.Contains(rendered, want) {
		t.Fatalf("metric output does not contain %q\nfull output:\n%s", want, rendered)
	}
}
