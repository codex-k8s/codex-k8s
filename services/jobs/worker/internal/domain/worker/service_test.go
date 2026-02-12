package worker

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/flowevent"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
)

func TestTickLaunchesPendingRun(t *testing.T) {
	t.Parallel()

	runs := &fakeRunQueue{
		claims: []runqueuerepo.ClaimedRun{
			{RunID: "run-1", CorrelationID: "corr-1", ProjectID: "proj-1", SlotNo: 1},
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{states: map[string]JobState{}}
	mcpTokens := &fakeMCPTokenIssuer{token: "token-run-1"}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:               "worker-1",
		ClaimLimit:             2,
		RunningCheckLimit:      10,
		SlotsPerProject:        2,
		SlotLeaseTTL:           time.Minute,
		ControlPlaneMCPBaseURL: "http://codex-k8s-control-plane.test.svc:8081/mcp",
	}, Dependencies{
		Runs:           runs,
		Events:         events,
		Launcher:       launcher,
		MCPTokenIssuer: mcpTokens,
		Logger:         logger,
	})
	svc.now = func() time.Time { return time.Date(2026, 2, 9, 10, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(launcher.launched) != 1 {
		t.Fatalf("expected 1 launched job, got %d", len(launcher.launched))
	}
	if launcher.launched[0].MCPBaseURL != "http://codex-k8s-control-plane.test.svc:8081/mcp" {
		t.Fatalf("expected mcp base url to be propagated, got %q", launcher.launched[0].MCPBaseURL)
	}
	if launcher.launched[0].MCPBearerToken != "token-run-1" {
		t.Fatalf("expected mcp token to be propagated, got %q", launcher.launched[0].MCPBearerToken)
	}
	if len(events.inserted) != 1 {
		t.Fatalf("expected 1 flow event, got %d", len(events.inserted))
	}
	if events.inserted[0].EventType != floweventdomain.EventTypeRunStarted {
		t.Fatalf("expected run.started event, got %s", events.inserted[0].EventType)
	}
}

func TestTickFinalizesSucceededRun(t *testing.T) {
	t.Parallel()

	runs := &fakeRunQueue{
		running: []runqueuerepo.RunningRun{{RunID: "run-2", CorrelationID: "corr-2", ProjectID: "proj-2"}},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{states: map[string]JobState{"run-2": JobStateSucceeded}}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{WorkerID: "worker-1", ClaimLimit: 1, RunningCheckLimit: 10, SlotsPerProject: 2, SlotLeaseTTL: time.Minute}, Dependencies{
		Runs:     runs,
		Events:   events,
		Launcher: launcher,
		Logger:   logger,
	})
	svc.now = func() time.Time { return time.Date(2026, 2, 9, 11, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(runs.finished) != 1 {
		t.Fatalf("expected 1 finished run, got %d", len(runs.finished))
	}
	if runs.finished[0].Status != rundomain.StatusSucceeded {
		t.Fatalf("expected succeeded status, got %s", runs.finished[0].Status)
	}
	if len(events.inserted) != 1 {
		t.Fatalf("expected 1 flow event, got %d", len(events.inserted))
	}
	if events.inserted[0].EventType != floweventdomain.EventTypeRunSucceeded {
		t.Fatalf("expected run.succeeded event, got %s", events.inserted[0].EventType)
	}
}

func TestTickLaunchesFullEnvRunWithNamespacePreparation(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"trigger":{"kind":"dev"},"issue":{"number":77}}`)
	runs := &fakeRunQueue{
		claims: []runqueuerepo.ClaimedRun{
			{RunID: "run-3", CorrelationID: "corr-3", ProjectID: "550e8400-e29b-41d4-a716-446655440000", RunPayload: payload, SlotNo: 1},
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{states: map[string]JobState{}}
	mcpTokens := &fakeMCPTokenIssuer{token: "token-run-3"}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:                "worker-1",
		ClaimLimit:              1,
		RunningCheckLimit:       10,
		SlotsPerProject:         2,
		SlotLeaseTTL:            time.Minute,
		RunNamespacePrefix:      "codex-issue",
		CleanupFullEnvNamespace: true,
		ControlPlaneMCPBaseURL:  "http://codex-k8s-control-plane.test.svc:8081/mcp",
	}, Dependencies{
		Runs:           runs,
		Events:         events,
		Launcher:       launcher,
		MCPTokenIssuer: mcpTokens,
		Logger:         logger,
	})
	svc.now = func() time.Time { return time.Date(2026, 2, 11, 10, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(launcher.prepared) != 1 {
		t.Fatalf("expected 1 prepared namespace, got %d", len(launcher.prepared))
	}
	if launcher.prepared[0].RuntimeMode != agentdomain.RuntimeModeFullEnv {
		t.Fatalf("expected full-env runtime mode, got %q", launcher.prepared[0].RuntimeMode)
	}
	if launcher.prepared[0].Namespace == "" {
		t.Fatal("expected non-empty namespace for full-env run")
	}
	if len(launcher.launched) != 1 {
		t.Fatalf("expected 1 launched job, got %d", len(launcher.launched))
	}
	if launcher.launched[0].RuntimeMode != agentdomain.RuntimeModeFullEnv {
		t.Fatalf("expected launched runtime mode full-env, got %q", launcher.launched[0].RuntimeMode)
	}
	if launcher.launched[0].Namespace == "" {
		t.Fatal("expected launched job namespace to be set")
	}
	if launcher.launched[0].MCPBearerToken != "token-run-3" {
		t.Fatalf("expected mcp token to be set, got %q", launcher.launched[0].MCPBearerToken)
	}
}

func TestTickFinalizesFullEnvRunAndCleansNamespace(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"trigger":{"kind":"dev_revise"},"issue":{"number":10}}`)
	runs := &fakeRunQueue{
		running: []runqueuerepo.RunningRun{{
			RunID:         "run-4",
			CorrelationID: "corr-4",
			ProjectID:     "550e8400-e29b-41d4-a716-446655440000",
			RunPayload:    payload,
		}},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{states: map[string]JobState{"run-4": JobStateSucceeded}}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:                "worker-1",
		ClaimLimit:              1,
		RunningCheckLimit:       10,
		SlotsPerProject:         2,
		SlotLeaseTTL:            time.Minute,
		RunNamespacePrefix:      "codex-issue",
		CleanupFullEnvNamespace: true,
	}, Dependencies{
		Runs:     runs,
		Events:   events,
		Launcher: launcher,
		Logger:   logger,
	})
	svc.now = func() time.Time { return time.Date(2026, 2, 11, 11, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(launcher.cleaned) != 1 {
		t.Fatalf("expected 1 cleaned namespace, got %d", len(launcher.cleaned))
	}
	if launcher.cleaned[0].Namespace == "" {
		t.Fatal("expected cleaned namespace to be set")
	}
}

type fakeRunQueue struct {
	claims     []runqueuerepo.ClaimedRun
	claimCalls int
	running    []runqueuerepo.RunningRun
	finished   []runqueuerepo.FinishParams
	claimErr   error
	listErr    error
	finishErr  error
}

func (f *fakeRunQueue) ClaimNextPending(_ context.Context, _ runqueuerepo.ClaimParams) (runqueuerepo.ClaimedRun, bool, error) {
	if f.claimErr != nil {
		return runqueuerepo.ClaimedRun{}, false, f.claimErr
	}
	if f.claimCalls >= len(f.claims) {
		return runqueuerepo.ClaimedRun{}, false, nil
	}
	item := f.claims[f.claimCalls]
	f.claimCalls++
	return item, true, nil
}

func (f *fakeRunQueue) ListRunning(_ context.Context, _ int) ([]runqueuerepo.RunningRun, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.running, nil
}

func (f *fakeRunQueue) FinishRun(_ context.Context, params runqueuerepo.FinishParams) (bool, error) {
	if f.finishErr != nil {
		return false, f.finishErr
	}
	f.finished = append(f.finished, params)
	return true, nil
}

type fakeFlowEvents struct {
	inserted []floweventrepo.InsertParams
	err      error
}

func (f *fakeFlowEvents) Insert(_ context.Context, params floweventrepo.InsertParams) error {
	if f.err != nil {
		return f.err
	}
	f.inserted = append(f.inserted, params)
	return nil
}

type fakeLauncher struct {
	states    map[string]JobState
	launched  []JobSpec
	prepared  []NamespaceSpec
	cleaned   []NamespaceSpec
	launchErr error
	statusErr error
}

type fakeMCPTokenIssuer struct {
	token string
	err   error
}

func (f *fakeMCPTokenIssuer) IssueRunMCPToken(_ context.Context, _ IssueMCPTokenParams) (IssuedMCPToken, error) {
	if f.err != nil {
		return IssuedMCPToken{}, f.err
	}
	return IssuedMCPToken{Token: f.token, ExpiresAt: time.Now().Add(time.Hour)}, nil
}

func (f *fakeLauncher) JobRef(runID string, namespace string) JobRef {
	if namespace == "" {
		namespace = "ns"
	}
	return JobRef{Namespace: namespace, Name: "job-" + runID}
}

func (f *fakeLauncher) EnsureNamespace(_ context.Context, spec NamespaceSpec) error {
	f.prepared = append(f.prepared, spec)
	return nil
}

func (f *fakeLauncher) CleanupNamespace(_ context.Context, spec NamespaceSpec) error {
	f.cleaned = append(f.cleaned, spec)
	return nil
}

func (f *fakeLauncher) Launch(_ context.Context, spec JobSpec) (JobRef, error) {
	if f.launchErr != nil {
		return JobRef{}, f.launchErr
	}
	f.launched = append(f.launched, spec)
	return f.JobRef(spec.RunID, spec.Namespace), nil
}

func (f *fakeLauncher) Status(_ context.Context, ref JobRef) (JobState, error) {
	if f.statusErr != nil {
		return "", f.statusErr
	}
	runID := strings.TrimPrefix(ref.Name, "job-")
	if state, ok := f.states[runID]; ok {
		return state, nil
	}
	return JobStatePending, nil
}
