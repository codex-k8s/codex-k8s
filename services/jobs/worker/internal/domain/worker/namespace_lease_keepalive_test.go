package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"
	"testing"
	"time"

	agentdomain "github.com/codex-k8s/kodex/libs/go/domain/agent"
	runqueuerepo "github.com/codex-k8s/kodex/services/jobs/worker/internal/domain/repository/runqueue"
)

func TestTickExtendsNamespaceLeaseBeforeCleanupForRunningFullEnvRun(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{
		"repository":{"full_name":"codex-k8s/kodex"},
		"trigger":{"kind":"dev"},
		"issue":{"number":393},
		"agent":{"key":"dev","name":"AI Developer"}
	}`)
	runs := &fakeRunQueue{
		running: []runqueuerepo.RunningRun{{
			RunID:         "run-393",
			CorrelationID: "corr-393",
			ProjectID:     "project-1",
			RunPayload:    payload,
			StartedAt:     time.Date(2026, 3, 13, 18, 0, 0, 0, time.UTC),
		}},
	}
	launcher := &fakeLauncher{
		states: map[string]JobState{
			"run-393": JobStateRunning,
		},
	}
	svc := NewService(Config{
		WorkerID:                   "worker-1",
		RunningCheckLimit:          1,
		RunLeaseTTL:                5 * time.Minute,
		RunNamespaceCleanupEnabled: true,
		DefaultNamespaceTTL:        24 * time.Hour,
		RunNamespacePrefix:         defaultRunNamespacePrefix,
	}, Dependencies{
		Runs:     runs,
		Launcher: launcher,
		Logger:   slog.Default(),
	})
	now := time.Date(2026, 3, 13, 18, 30, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick returned error: %v", err)
	}
	if len(launcher.prepared) != 1 {
		t.Fatalf("prepared namespaces = %d, want 1", len(launcher.prepared))
	}
	gotSpec := launcher.prepared[0]
	if gotSpec.Namespace == "" {
		t.Fatal("expected namespace lease keepalive to target managed namespace")
	}
	if gotSpec.RuntimeMode != agentdomain.RuntimeModeFullEnv {
		t.Fatalf("runtime mode = %q, want %q", gotSpec.RuntimeMode, agentdomain.RuntimeModeFullEnv)
	}
	if gotSpec.LeaseTTL != 24*time.Hour {
		t.Fatalf("lease ttl = %s, want 24h", gotSpec.LeaseTTL)
	}
	wantLeaseExpiry := now.Add(24 * time.Hour)
	if !gotSpec.LeaseExpiresAt.Equal(wantLeaseExpiry) {
		t.Fatalf("lease expires at = %s, want %s", gotSpec.LeaseExpiresAt.Format(time.RFC3339), wantLeaseExpiry.Format(time.RFC3339))
	}
	if got := slices.Index(launcher.callLog, "list_managed_namespaces"); got == -1 {
		t.Fatalf("expected managed namespace listing in call log, got %v", launcher.callLog)
	}

	ensureIdx := slices.Index(launcher.callLog, "ensure_namespace")
	cleanupIdx := slices.Index(launcher.callLog, "list_managed_namespaces")
	if ensureIdx == -1 || cleanupIdx == -1 {
		t.Fatalf("call log missing ensure/cleanup markers: %v", launcher.callLog)
	}
	if ensureIdx >= cleanupIdx {
		t.Fatalf("expected namespace keepalive before cleanup, got %v", launcher.callLog)
	}
}
