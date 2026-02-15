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
			{
				RunID:         "run-1",
				CorrelationID: "corr-1",
				ProjectID:     "proj-1",
				RunPayload:    json.RawMessage(`{"repository":{"full_name":"codex-k8s/codex-k8s"},"trigger":{"kind":"dev"},"issue":{"number":1},"agent":{"key":"dev","name":"AI Developer"}}`),
				SlotNo:        1,
			},
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
	if len(events.inserted) != 2 {
		t.Fatalf("expected run.namespace.prepared + run.started events, got %d", len(events.inserted))
	}
	if events.inserted[0].EventType != floweventdomain.EventTypeRunNamespacePrepared {
		t.Fatalf("expected first event run.namespace.prepared, got %s", events.inserted[0].EventType)
	}
	if events.inserted[1].EventType != floweventdomain.EventTypeRunStarted {
		t.Fatalf("expected second event run.started, got %s", events.inserted[1].EventType)
	}
}

func TestTickSkipsCodeOnlyRun(t *testing.T) {
	t.Parallel()

	runs := &fakeRunQueue{
		claims: []runqueuerepo.ClaimedRun{
			{
				RunID:         "run-code-only",
				CorrelationID: "corr-code-only",
				ProjectID:     "proj-1",
				RunPayload:    json.RawMessage(`{"repository":{"full_name":"codex-k8s/codex-k8s"},"issue":{"number":123},"agent":{"key":"dev","name":"AI Developer"}}`),
				SlotNo:        1,
			},
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{states: map[string]JobState{}}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:          "worker-1",
		ClaimLimit:        1,
		RunningCheckLimit: 10,
		SlotsPerProject:   2,
		SlotLeaseTTL:      time.Minute,
	}, Dependencies{
		Runs:     runs,
		Events:   events,
		Launcher: launcher,
		Logger:   logger,
	})
	svc.now = func() time.Time { return time.Date(2026, 2, 12, 10, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(launcher.launched) != 0 {
		t.Fatalf("expected no launched jobs for code-only run, got %d", len(launcher.launched))
	}
	if len(runs.finished) != 1 {
		t.Fatalf("expected 1 finished run, got %d", len(runs.finished))
	}
	if runs.finished[0].Status != rundomain.StatusSucceeded {
		t.Fatalf("expected code-only run to finish as succeeded, got %s", runs.finished[0].Status)
	}
	if len(events.inserted) != 1 || events.inserted[0].EventType != floweventdomain.EventTypeRunSucceeded {
		t.Fatalf("expected single run.succeeded event, got %#v", events.inserted)
	}
}

func TestTickDeployOnlyRun_PreparesEnvironmentWithoutLaunchingJob(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{
		"repository":{"full_name":"codex-k8s/codex-k8s"},
		"runtime":{
			"mode":"full-env",
			"target_env":"ai-staging",
			"namespace":"codex-k8s-ai-staging",
			"build_ref":"0123456789abcdef0123456789abcdef01234567",
			"deploy_only":true
		}
	}`)
	runs := &fakeRunQueue{
		claims: []runqueuerepo.ClaimedRun{
			{
				RunID:         "run-deploy-only",
				CorrelationID: "corr-deploy-only",
				ProjectID:     "proj-1",
				RunPayload:    payload,
				SlotNo:        1,
			},
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{states: map[string]JobState{}}
	deployer := &fakeRuntimePreparer{
		result: PrepareRunEnvironmentResult{
			Namespace: "codex-k8s-ai-staging",
			TargetEnv: "ai-staging",
		},
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:          "worker-1",
		ClaimLimit:        1,
		RunningCheckLimit: 10,
		SlotsPerProject:   2,
		SlotLeaseTTL:      time.Minute,
	}, Dependencies{
		Runs:            runs,
		Events:          events,
		Launcher:        launcher,
		RuntimePreparer: deployer,
		Logger:          logger,
	})
	svc.now = func() time.Time { return time.Date(2026, 2, 14, 10, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(deployer.prepared) != 1 {
		t.Fatalf("expected 1 runtime deploy call, got %d", len(deployer.prepared))
	}
	if !deployer.prepared[0].DeployOnly {
		t.Fatal("expected deploy-only runtime deploy params")
	}
	if got, want := deployer.prepared[0].Namespace, "codex-k8s-ai-staging"; got != want {
		t.Fatalf("unexpected deploy namespace: got %q want %q", got, want)
	}
	if len(launcher.prepared) != 0 {
		t.Fatalf("expected no runtime namespace preparation for deploy-only run, got %d", len(launcher.prepared))
	}
	if len(launcher.launched) != 0 {
		t.Fatalf("expected no launched jobs for deploy-only run, got %d", len(launcher.launched))
	}
	if len(runs.finished) != 1 {
		t.Fatalf("expected 1 finished run, got %d", len(runs.finished))
	}
	if runs.finished[0].Status != rundomain.StatusSucceeded {
		t.Fatalf("expected deploy-only run to finish as succeeded, got %s", runs.finished[0].Status)
	}
	if len(events.inserted) != 1 || events.inserted[0].EventType != floweventdomain.EventTypeRunSucceeded {
		t.Fatalf("expected one run.succeeded event, got %#v", events.inserted)
	}
}

func TestTickDeployOnlyRunningRun_IsReconciledWithoutKubernetesJob(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{
		"repository":{"full_name":"codex-k8s/codex-k8s"},
		"runtime":{
			"mode":"full-env",
			"target_env":"ai-staging",
			"build_ref":"0123456789abcdef0123456789abcdef01234567",
			"deploy_only":true
		}
	}`)
	runs := &fakeRunQueue{
		running: []runqueuerepo.RunningRun{
			{
				RunID:         "run-deploy-only",
				CorrelationID: "corr-deploy-only",
				ProjectID:     "proj-1",
				SlotNo:        1,
				RunPayload:    payload,
			},
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{states: map[string]JobState{}, statusErr: context.Canceled}
	deployer := &fakeRuntimePreparer{
		result: PrepareRunEnvironmentResult{
			Namespace: "codex-k8s-ai-staging",
			TargetEnv: "ai-staging",
		},
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:          "worker-1",
		ClaimLimit:        1,
		RunningCheckLimit: 10,
		SlotsPerProject:   2,
		SlotLeaseTTL:      time.Minute,
	}, Dependencies{
		Runs:            runs,
		Events:          events,
		Launcher:        launcher,
		RuntimePreparer: deployer,
		Logger:          logger,
	})
	svc.now = func() time.Time { return time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(deployer.prepared) != 1 {
		t.Fatalf("expected 1 runtime deploy call, got %d", len(deployer.prepared))
	}
	if len(runs.finished) != 1 {
		t.Fatalf("expected 1 finished run, got %d", len(runs.finished))
	}
	if runs.finished[0].Status != rundomain.StatusSucceeded {
		t.Fatalf("expected deploy-only running run to finish as succeeded, got %s", runs.finished[0].Status)
	}
	if len(launcher.cleaned) != 0 {
		t.Fatalf("expected no namespace cleanup for deploy-only run, got %d", len(launcher.cleaned))
	}
	if len(events.inserted) != 1 || events.inserted[0].EventType != floweventdomain.EventTypeRunSucceeded {
		t.Fatalf("expected one run.succeeded event, got %#v", events.inserted)
	}
}

func TestTickCodeOnlyRunningRun_IsReconciledWithoutKubernetesJob(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"repository":{"full_name":"codex-k8s/codex-k8s"}}`)
	runs := &fakeRunQueue{
		running: []runqueuerepo.RunningRun{
			{
				RunID:         "run-code-only",
				CorrelationID: "corr-code-only",
				ProjectID:     "proj-1",
				SlotNo:        1,
				RunPayload:    payload,
			},
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{states: map[string]JobState{}, statusErr: context.Canceled}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:          "worker-1",
		ClaimLimit:        1,
		RunningCheckLimit: 10,
		SlotsPerProject:   2,
		SlotLeaseTTL:      time.Minute,
	}, Dependencies{
		Runs:     runs,
		Events:   events,
		Launcher: launcher,
		Logger:   logger,
	})
	svc.now = func() time.Time { return time.Date(2026, 2, 15, 11, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(runs.finished) != 1 {
		t.Fatalf("expected 1 finished run, got %d", len(runs.finished))
	}
	if runs.finished[0].Status != rundomain.StatusSucceeded {
		t.Fatalf("expected code-only running run to finish as succeeded, got %s", runs.finished[0].Status)
	}
	if len(events.inserted) != 1 || events.inserted[0].EventType != floweventdomain.EventTypeRunSucceeded {
		t.Fatalf("expected one run.succeeded event, got %#v", events.inserted)
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

	payload := json.RawMessage(`{"repository":{"full_name":"codex-k8s/codex-k8s"},"trigger":{"kind":"dev"},"issue":{"number":77},"agent":{"key":"dev","name":"AI Developer"}}`)
	runs := &fakeRunQueue{
		claims: []runqueuerepo.ClaimedRun{
			{RunID: "run-3", CorrelationID: "corr-3", ProjectID: "550e8400-e29b-41d4-a716-446655440000", RunPayload: payload, SlotNo: 1},
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{states: map[string]JobState{}}
	mcpTokens := &fakeMCPTokenIssuer{token: "token-run-3"}
	deployer := &fakeRuntimePreparer{
		result: PrepareRunEnvironmentResult{
			Namespace: "codex-k8s-dev-1",
			TargetEnv: "ai",
		},
	}
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
		Runs:            runs,
		Events:          events,
		Launcher:        launcher,
		RuntimePreparer: deployer,
		MCPTokenIssuer:  mcpTokens,
		Logger:          logger,
	})
	svc.now = func() time.Time { return time.Date(2026, 2, 11, 10, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(launcher.prepared) != 1 {
		t.Fatalf("expected 1 prepared namespace, got %d", len(launcher.prepared))
	}
	if len(deployer.prepared) != 1 {
		t.Fatalf("expected 1 runtime deploy call, got %d", len(deployer.prepared))
	}
	if got, want := deployer.prepared[0].RunID, "run-3"; got != want {
		t.Fatalf("unexpected deploy run id: got %q want %q", got, want)
	}
	if got := deployer.prepared[0].Namespace; got != "" {
		t.Fatalf("expected empty namespace in deploy request for slot-mode full-env run, got %q", got)
	}
	if got := deployer.prepared[0].DeployOnly; got {
		t.Fatal("expected deploy_only=false for full-env agent run")
	}
	if launcher.prepared[0].RuntimeMode != agentdomain.RuntimeModeFullEnv {
		t.Fatalf("expected full-env runtime mode, got %q", launcher.prepared[0].RuntimeMode)
	}
	if got, want := launcher.prepared[0].Namespace, "codex-k8s-dev-1"; got != want {
		t.Fatalf("expected prepared namespace %q, got %q", want, got)
	}
	if len(launcher.launched) != 1 {
		t.Fatalf("expected 1 launched job, got %d", len(launcher.launched))
	}
	if launcher.launched[0].RuntimeMode != agentdomain.RuntimeModeFullEnv {
		t.Fatalf("expected launched runtime mode full-env, got %q", launcher.launched[0].RuntimeMode)
	}
	if got, want := launcher.launched[0].Namespace, "codex-k8s-dev-1"; got != want {
		t.Fatalf("expected launched namespace %q, got %q", want, got)
	}
	if launcher.launched[0].MCPBearerToken != "token-run-3" {
		t.Fatalf("expected mcp token to be set, got %q", launcher.launched[0].MCPBearerToken)
	}
}

func TestTickFinalizesFullEnvRunAndCleansNamespace(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"repository":{"full_name":"codex-k8s/codex-k8s"},"trigger":{"kind":"dev_revise"},"issue":{"number":10},"agent":{"key":"dev","name":"AI Developer"}}`)
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

func TestTickFinalizesFullEnvRunSkipsCleanupForDebugLabel(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"repository":{"full_name":"codex-k8s/codex-k8s"},"trigger":{"kind":"dev"},"issue":{"number":10},"agent":{"key":"dev","name":"AI Developer"},"raw_payload":{"issue":{"labels":[{"name":"run:debug"}]}}}`)
	runs := &fakeRunQueue{
		running: []runqueuerepo.RunningRun{{
			RunID:         "run-5",
			CorrelationID: "corr-5",
			ProjectID:     "550e8400-e29b-41d4-a716-446655440000",
			RunPayload:    payload,
		}},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{states: map[string]JobState{"run-5": JobStateSucceeded}}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:                "worker-1",
		ClaimLimit:              1,
		RunningCheckLimit:       10,
		SlotsPerProject:         2,
		SlotLeaseTTL:            time.Minute,
		RunNamespacePrefix:      "codex-issue",
		CleanupFullEnvNamespace: true,
		RunDebugLabel:           "run:debug",
	}, Dependencies{
		Runs:     runs,
		Events:   events,
		Launcher: launcher,
		Logger:   logger,
	})
	svc.now = func() time.Time { return time.Date(2026, 2, 11, 12, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(launcher.cleaned) != 0 {
		t.Fatalf("expected namespace cleanup to be skipped, got %d cleanups", len(launcher.cleaned))
	}

	if len(events.inserted) != 2 {
		t.Fatalf("expected run.succeeded + run.namespace.cleanup_skipped events, got %d", len(events.inserted))
	}
	if events.inserted[0].EventType != floweventdomain.EventTypeRunSucceeded {
		t.Fatalf("expected first event run.succeeded, got %s", events.inserted[0].EventType)
	}
	if events.inserted[1].EventType != floweventdomain.EventTypeRunNamespaceCleanupSkipped {
		t.Fatalf("expected second event run.namespace.cleanup_skipped, got %s", events.inserted[1].EventType)
	}
	if !strings.Contains(string(events.inserted[1].Payload), "debug_label_present") {
		t.Fatalf("expected cleanup_skipped payload to include debug reason, got %s", string(events.inserted[1].Payload))
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

type fakeRuntimePreparer struct {
	prepared []PrepareRunEnvironmentParams
	result   PrepareRunEnvironmentResult
	err      error
}

func (f *fakeRuntimePreparer) PrepareRunEnvironment(_ context.Context, params PrepareRunEnvironmentParams) (PrepareRunEnvironmentResult, error) {
	if f.err != nil {
		return PrepareRunEnvironmentResult{}, f.err
	}
	f.prepared = append(f.prepared, params)
	return f.result, nil
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
