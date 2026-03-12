package worker

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/flowevent"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
)

func TestTickRunningLeaseOwnerStillActive_SkipsStaleReclaim(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	payload := json.RawMessage(`{"repository":{"full_name":"codex-k8s/codex-k8s"},"trigger":{"kind":"dev"},"issue":{"number":287},"agent":{"key":"dev","name":"AI Developer"}}`)
	runs := &fakeRunQueue{
		claimedRunning: []runqueuerepo.RunningRun{},
		running: []runqueuerepo.RunningRun{
			{
				RunID:         "run-active-owner",
				CorrelationID: "corr-active-owner",
				ProjectID:     "550e8400-e29b-41d4-a716-446655440000",
				RunPayload:    payload,
				StartedAt:     now.Add(-2 * time.Minute),
				LeaseOwner:    "worker-2",
				LeaseUntil:    now.Add(30 * time.Minute),
			},
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:          "worker-1",
		RunningCheckLimit: 10,
		SlotsPerProject:   2,
		SlotLeaseTTL:      time.Minute,
		RunLeaseTTL:       45 * time.Minute,
	}, Dependencies{
		Runs:           runs,
		Events:         events,
		Launcher:       launcher,
		WorkerPresence: fakeWorkerPresence{active: []string{"worker-1", "worker-2"}},
		Logger:         logger,
	})
	svc.now = func() time.Time { return now }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(runs.reclaimStale) != 0 {
		t.Fatalf("expected no stale lease reclaim attempts, got %d", len(runs.reclaimStale))
	}
	if len(events.inserted) != 0 {
		t.Fatalf("expected no reclaim events, got %d", len(events.inserted))
	}
	if len(launcher.launched) != 0 {
		t.Fatalf("expected no launched jobs, got %d", len(launcher.launched))
	}
}

func TestTickRunningStaleLeaseMissingWorker_ReclaimsAndLaunchesRun(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	payload := json.RawMessage(`{"repository":{"full_name":"codex-k8s/codex-k8s"},"trigger":{"kind":"dev"},"issue":{"number":287},"agent":{"key":"dev","name":"AI Developer"}}`)
	staleRun := runqueuerepo.RunningRun{
		RunID:         "run-stale-lease",
		CorrelationID: "corr-stale-lease",
		ProjectID:     "550e8400-e29b-41d4-a716-446655440000",
		RunPayload:    payload,
		StartedAt:     now.Add(-10 * time.Minute),
		LeaseOwner:    "worker-old-rs",
		LeaseUntil:    now.Add(35 * time.Minute),
	}
	reclaimedRun := staleRun
	reclaimedRun.LeaseOwner = "worker-1"
	reclaimedRun.LeaseUntil = now.Add(45 * time.Minute)

	runs := &fakeRunQueue{
		claimedRunning: []runqueuerepo.RunningRun{},
		running:        []runqueuerepo.RunningRun{staleRun},
		reclaimResults: map[string]runqueuerepo.RunningRun{
			staleRun.RunID: reclaimedRun,
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{
		states: map[string]JobState{
			staleRun.RunID: JobStateNotFound,
		},
	}
	deployer := &fakeRuntimePreparer{result: PrepareRunEnvironmentResult{Namespace: "codex-issue-runtime"}}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:                   "worker-1",
		RunningCheckLimit:          10,
		SlotsPerProject:            2,
		SlotLeaseTTL:               time.Minute,
		RunLeaseTTL:                45 * time.Minute,
		RuntimePrepareRetryTimeout: 5 * time.Minute,
		JobImage:                   "registry.local/agent-runner:test",
	}, Dependencies{
		Runs:            runs,
		Events:          events,
		Launcher:        launcher,
		RuntimePreparer: deployer,
		WorkerPresence:  fakeWorkerPresence{active: []string{"worker-1"}},
		Logger:          logger,
	})
	svc.now = func() time.Time { return now }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(runs.reclaimStale) != 1 {
		t.Fatalf("expected 1 stale reclaim attempt, got %d", len(runs.reclaimStale))
	}
	if got := runs.reclaimStale[0].PreviousLeaseOwner; got != "worker-old-rs" {
		t.Fatalf("expected previous lease owner %q, got %q", "worker-old-rs", got)
	}
	if len(launcher.launched) != 1 {
		t.Fatalf("expected 1 launched job after stale reclaim, got %d", len(launcher.launched))
	}
	if launcher.launched[0].RunID != staleRun.RunID {
		t.Fatalf("expected launched run %q, got %q", staleRun.RunID, launcher.launched[0].RunID)
	}
	if launcher.launched[0].Namespace != "codex-issue-runtime" {
		t.Fatalf("expected launched namespace %q, got %q", "codex-issue-runtime", launcher.launched[0].Namespace)
	}
	if len(deployer.prepared) != 1 {
		t.Fatalf("expected 1 runtime prepare call, got %d", len(deployer.prepared))
	}

	if got := collectEventTypes(events.inserted); !containsEvent(got, floweventdomain.EventTypeRunLeaseOwnerMissing) {
		t.Fatalf("expected %q event in %v", floweventdomain.EventTypeRunLeaseOwnerMissing, got)
	}
	if got := collectEventTypes(events.inserted); !containsEvent(got, floweventdomain.EventTypeRunLeaseStaleDetected) {
		t.Fatalf("expected %q event in %v", floweventdomain.EventTypeRunLeaseStaleDetected, got)
	}
	if got := collectEventTypes(events.inserted); !containsEvent(got, floweventdomain.EventTypeRunLeaseReclaimed) {
		t.Fatalf("expected %q event in %v", floweventdomain.EventTypeRunLeaseReclaimed, got)
	}
	if got := collectEventTypes(events.inserted); !containsEvent(got, floweventdomain.EventTypeRunStarted) {
		t.Fatalf("expected %q event in %v", floweventdomain.EventTypeRunStarted, got)
	}
}

func TestTickRunningExpiredLeaseClaimedWithoutStaleReclaim(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	payload := json.RawMessage(`{"runtime":{"mode":"code-only"}}`)
	expiredRun := runqueuerepo.RunningRun{
		RunID:         "run-expired-lease",
		CorrelationID: "corr-expired-lease",
		ProjectID:     "550e8400-e29b-41d4-a716-446655440001",
		RunPayload:    payload,
		StartedAt:     now.Add(-30 * time.Minute),
		LeaseOwner:    "worker-1",
		LeaseUntil:    now.Add(45 * time.Minute),
	}

	runs := &fakeRunQueue{
		claimedRunning: []runqueuerepo.RunningRun{expiredRun},
		running: []runqueuerepo.RunningRun{
			{
				RunID:         expiredRun.RunID,
				CorrelationID: expiredRun.CorrelationID,
				ProjectID:     expiredRun.ProjectID,
				RunPayload:    expiredRun.RunPayload,
				StartedAt:     expiredRun.StartedAt,
				LeaseOwner:    "worker-dead-rs",
				LeaseUntil:    now.Add(-time.Minute),
			},
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{
		states: map[string]JobState{
			expiredRun.RunID: JobStateSucceeded,
		},
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:          "worker-1",
		RunningCheckLimit: 10,
		SlotsPerProject:   2,
		SlotLeaseTTL:      time.Minute,
		RunLeaseTTL:       45 * time.Minute,
	}, Dependencies{
		Runs:           runs,
		Events:         events,
		Launcher:       launcher,
		WorkerPresence: fakeWorkerPresence{active: []string{"worker-1"}},
		Logger:         logger,
	})
	svc.now = func() time.Time { return now }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(runs.reclaimStale) != 0 {
		t.Fatalf("expected expired lease to be handled by regular claim path, got %d stale reclaim attempts", len(runs.reclaimStale))
	}
	if len(runs.finished) != 1 {
		t.Fatalf("expected 1 finished run after expired lease claim, got %d", len(runs.finished))
	}
	if runs.finished[0].Status != rundomain.StatusSucceeded {
		t.Fatalf("expected status %q, got %q", rundomain.StatusSucceeded, runs.finished[0].Status)
	}
}

type fakeWorkerPresence struct {
	active []string
	err    error
}

func (f fakeWorkerPresence) ListActiveWorkerIDs(_ context.Context) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	out := make([]string, len(f.active))
	copy(out, f.active)
	return out, nil
}

func collectEventTypes(items []floweventrepo.InsertParams) []floweventdomain.EventType {
	result := make([]floweventdomain.EventType, 0, len(items))
	for _, item := range items {
		result = append(result, item.EventType)
	}
	return result
}

func containsEvent(items []floweventdomain.EventType, want floweventdomain.EventType) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
