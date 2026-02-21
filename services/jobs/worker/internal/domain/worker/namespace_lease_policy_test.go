package worker

import (
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestResolveNamespaceLeaseContext(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"trigger":{"kind":"dev_revise"},"issue":{"number":74},"agent":{"key":"dev","name":"AI Developer"}}`)
	ctx := resolveNamespaceLeaseContext(payload)
	if got, want := ctx.AgentKey, "dev"; got != want {
		t.Fatalf("unexpected agent key: got %q want %q", got, want)
	}
	if got, want := ctx.IssueNumber, int64(74); got != want {
		t.Fatalf("unexpected issue number: got %d want %d", got, want)
	}
	if !ctx.IsRevise {
		t.Fatal("expected revise context for dev_revise trigger")
	}
}

func TestServiceResolveNamespaceTTL_ByRoleAndDefault(t *testing.T) {
	t.Parallel()

	svc := NewService(Config{
		DefaultNamespaceTTL: 12 * time.Hour,
		NamespaceTTLByRole: map[string]time.Duration{
			"dev": 24 * time.Hour,
		},
	}, Dependencies{Logger: slog.New(slog.NewJSONHandler(io.Discard, nil))})

	if got, want := svc.resolveNamespaceTTL("dev"), 24*time.Hour; got != want {
		t.Fatalf("unexpected dev ttl: got %s want %s", got, want)
	}
	if got, want := svc.resolveNamespaceTTL("qa"), 12*time.Hour; got != want {
		t.Fatalf("unexpected fallback ttl: got %s want %s", got, want)
	}
}

func TestNormalizeNamespaceTTLByRole(t *testing.T) {
	t.Parallel()

	normalized := normalizeNamespaceTTLByRole(map[string]time.Duration{
		"DEV": 24 * time.Hour,
		"qa":  0,
		"":    12 * time.Hour,
	})
	if got, want := len(normalized), 1; got != want {
		t.Fatalf("unexpected normalized role count: got %d want %d", got, want)
	}
	if got, want := normalized["dev"], 24*time.Hour; got != want {
		t.Fatalf("unexpected normalized ttl: got %s want %s", got, want)
	}
}
