package worker

import (
	"encoding/json"
	"io"
	"log/slog"
	"testing"
)

func TestResolveRunDebugPolicy_CleanupDisabled(t *testing.T) {
	t.Parallel()

	svc := NewService(Config{CleanupFullEnvNamespace: false}, Dependencies{Logger: slog.New(slog.NewJSONHandler(io.Discard, nil))})
	policy := svc.resolveRunDebugPolicy(nil)
	if !policy.SkipCleanup {
		t.Fatal("expected cleanup to be skipped when cleanup flag is disabled")
	}
	if policy.Reason != namespaceCleanupSkipReasonDisabledByConfig {
		t.Fatalf("expected disabled_by_config reason, got %q", policy.Reason)
	}
}

func TestResolveRunDebugPolicy_DebugLabel(t *testing.T) {
	t.Parallel()

	svc := NewService(Config{CleanupFullEnvNamespace: true, RunDebugLabel: "run:debug"}, Dependencies{Logger: slog.New(slog.NewJSONHandler(io.Discard, nil))})
	payload := json.RawMessage(`{"raw_payload":{"issue":{"labels":[{"name":"run:debug"}]}}}`)

	policy := svc.resolveRunDebugPolicy(payload)
	if !policy.SkipCleanup {
		t.Fatal("expected cleanup to be skipped for run:debug")
	}
	if policy.Reason != namespaceCleanupSkipReasonDebugLabel {
		t.Fatalf("expected debug_label_present reason, got %q", policy.Reason)
	}
}
