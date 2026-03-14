package app

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestRegisterOrReplaceCollectorReplacesExistingCollector(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	firstCollector := newStubCollector(1)
	if err := registry.Register(firstCollector); err != nil {
		t.Fatalf("register first collector: %v", err)
	}

	secondCollector := newStubCollector(2)
	if err := registerOrReplaceCollector(registry, secondCollector); err != nil {
		t.Fatalf("registerOrReplaceCollector() error = %v", err)
	}

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	for _, family := range metricFamilies {
		if family.GetName() != "codexk8s_test_interaction_collector" {
			continue
		}
		if got := family.GetMetric()[0].GetGauge().GetValue(); got != 2 {
			t.Fatalf("collector gauge value = %v, want 2", got)
		}
		return
	}

	t.Fatal("replacement collector metric not found")
}

type stubCollector struct {
	desc  *prometheus.Desc
	value float64
}

func newStubCollector(value float64) *stubCollector {
	return &stubCollector{
		desc:  prometheus.NewDesc("codexk8s_test_interaction_collector", "test collector replacement", nil, nil),
		value: value,
	}
}

func (c *stubCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}

func (c *stubCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, c.value)
}
