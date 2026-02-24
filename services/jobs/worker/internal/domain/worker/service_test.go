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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	deployer := &fakeRuntimePreparer{
		result: PrepareRunEnvironmentResult{
			Namespace: "codex-k8s-dev-1",
			TargetEnv: "ai",
		},
	}
	mcpTokens := &fakeMCPTokenIssuer{token: "token-run-1"}
	runStatus := &fakeRunStatusNotifier{}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:               "worker-1",
		ClaimLimit:             2,
		RunningCheckLimit:      10,
		SlotsPerProject:        2,
		SlotLeaseTTL:           time.Minute,
		ControlPlaneMCPBaseURL: "http://codex-k8s-control-plane.test.svc:8081/mcp",
	}, Dependencies{
		Runs:            runs,
		Events:          events,
		Launcher:        launcher,
		RuntimePreparer: deployer,
		MCPTokenIssuer:  mcpTokens,
		RunStatus:       runStatus,
		Logger:          logger,
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
	if len(events.inserted) != 4 {
		t.Fatalf("expected run.namespace.prepared + run.namespace.ttl_scheduled + run.profile.resolved + run.started events, got %d", len(events.inserted))
	}
	if events.inserted[0].EventType != floweventdomain.EventTypeRunNamespacePrepared {
		t.Fatalf("expected first event run.namespace.prepared, got %s", events.inserted[0].EventType)
	}
	if events.inserted[1].EventType != floweventdomain.EventTypeRunNamespaceTTLScheduled {
		t.Fatalf("expected second event run.namespace.ttl_scheduled, got %s", events.inserted[1].EventType)
	}
	if events.inserted[2].EventType != floweventdomain.EventTypeRunProfileResolved {
		t.Fatalf("expected third event run.profile.resolved, got %s", events.inserted[2].EventType)
	}
	if events.inserted[3].EventType != floweventdomain.EventTypeRunStarted {
		t.Fatalf("expected fourth event run.started, got %s", events.inserted[3].EventType)
	}
	if len(runStatus.upserts) < 1 {
		t.Fatalf("expected run status upserts, got %d", len(runStatus.upserts))
	}
	if runStatus.upserts[0].Phase != RunStatusPhasePreparingRuntime {
		t.Fatalf("expected first run status phase %q, got %q", RunStatusPhasePreparingRuntime, runStatus.upserts[0].Phase)
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
	if runs.finished[0].LeaseOwner != "worker-1" {
		t.Fatalf("expected finish lease owner %q, got %q", "worker-1", runs.finished[0].LeaseOwner)
	}
	if len(events.inserted) != 1 || events.inserted[0].EventType != floweventdomain.EventTypeRunSucceeded {
		t.Fatalf("expected single run.succeeded event, got %#v", events.inserted)
	}
}

func TestTickLaunchesAIRepairCodeOnlyRunAsPodWorkload(t *testing.T) {
	t.Parallel()

	runs := &fakeRunQueue{
		claims: []runqueuerepo.ClaimedRun{
			{
				RunID:         "run-ai-repair",
				CorrelationID: "corr-ai-repair",
				ProjectID:     "proj-1",
				RunPayload: json.RawMessage(`{
					"repository":{"full_name":"codex-k8s/codex-k8s"},
					"trigger":{"kind":"ai_repair"},
					"issue":{"number":45},
					"agent":{"key":"sre","name":"AI SRE"},
					"runtime":{"mode":"code-only","namespace":"codex-k8s-prod"}
				}`),
				SlotNo: 1,
			},
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{states: map[string]JobState{}}
	mcpTokens := &fakeMCPTokenIssuer{token: "token-ai-repair"}
	deployer := &fakeRuntimePreparer{}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:               "worker-1",
		ClaimLimit:             1,
		RunningCheckLimit:      10,
		SlotsPerProject:        2,
		SlotLeaseTTL:           time.Minute,
		AgentBaseBranch:        "main",
		AIRepairNamespace:      "codex-k8s-prod",
		AIRepairServiceAccount: "codex-k8s-control-plane",
	}, Dependencies{
		Runs:            runs,
		Events:          events,
		Launcher:        launcher,
		MCPTokenIssuer:  mcpTokens,
		RuntimePreparer: deployer,
		Logger:          logger,
	})
	svc.now = func() time.Time { return time.Date(2026, 2, 24, 11, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(deployer.prepared) != 0 {
		t.Fatalf("expected no runtime deploy calls for ai-repair code-only run, got %d", len(deployer.prepared))
	}
	if len(launcher.launched) != 1 {
		t.Fatalf("expected 1 launched workload, got %d", len(launcher.launched))
	}
	if got, want := launcher.launched[0].Namespace, "codex-k8s-prod"; got != want {
		t.Fatalf("expected ai-repair namespace %q, got %q", want, got)
	}
	if got, want := launcher.launched[0].ServiceAccountName, "codex-k8s-control-plane"; got != want {
		t.Fatalf("expected ai-repair service account %q, got %q", want, got)
	}
	if got, want := launcher.launched[0].TargetBranch, "main"; got != want {
		t.Fatalf("expected ai-repair target branch %q, got %q", want, got)
	}
	if len(events.inserted) != 2 {
		t.Fatalf("expected run.profile.resolved + run.started events, got %d", len(events.inserted))
	}
	if events.inserted[0].EventType != floweventdomain.EventTypeRunProfileResolved {
		t.Fatalf("expected first event run.profile.resolved, got %s", events.inserted[0].EventType)
	}
	if events.inserted[1].EventType != floweventdomain.EventTypeRunStarted {
		t.Fatalf("expected second event run.started, got %s", events.inserted[1].EventType)
	}
}

func TestTickDeployOnlyRun_PreparesEnvironmentWithoutLaunchingJob(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{
		"repository":{"full_name":"codex-k8s/codex-k8s"},
		"runtime":{
			"mode":"full-env",
			"target_env":"production",
			"namespace":"codex-k8s-prod",
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
			Namespace: "codex-k8s-prod",
			TargetEnv: "production",
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
	if got, want := deployer.prepared[0].Namespace, "codex-k8s-prod"; got != want {
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

func TestTickDeployOnlyRun_RuntimeTaskCanceled_FinishesRunCanceled(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{
		"repository":{"full_name":"codex-k8s/codex-k8s"},
		"runtime":{
			"mode":"full-env",
			"target_env":"production",
			"namespace":"codex-k8s-prod",
			"build_ref":"0123456789abcdef0123456789abcdef01234567",
			"deploy_only":true
		}
	}`)
	runs := &fakeRunQueue{
		claims: []runqueuerepo.ClaimedRun{
			{
				RunID:         "run-deploy-only-canceled",
				CorrelationID: "corr-deploy-only-canceled",
				ProjectID:     "proj-1",
				RunPayload:    payload,
				SlotNo:        1,
			},
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{states: map[string]JobState{}}
	deployer := &fakeRuntimePreparer{
		err: status.Error(codes.Canceled, "runtime deploy task canceled for run_id=run-deploy-only-canceled: superseded"),
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
	svc.now = func() time.Time { return time.Date(2026, 2, 20, 9, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(runs.finished) != 1 {
		t.Fatalf("expected 1 finished run, got %d", len(runs.finished))
	}
	if runs.finished[0].Status != rundomain.StatusCanceled {
		t.Fatalf("expected deploy-only run to finish as canceled, got %s", runs.finished[0].Status)
	}
	if len(events.inserted) != 1 || events.inserted[0].EventType != floweventdomain.EventTypeRunCanceled {
		t.Fatalf("expected one run.canceled event, got %#v", events.inserted)
	}
}

func TestTickDeployOnlyRunningRun_IsReconciledWithoutKubernetesJob(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{
		"repository":{"full_name":"codex-k8s/codex-k8s"},
		"runtime":{
			"mode":"full-env",
			"target_env":"production",
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
			Namespace: "codex-k8s-prod",
			TargetEnv: "production",
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

func TestTickAIRepairRunningRun_IsReconciledByWorkloadState(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{
		"repository":{"full_name":"codex-k8s/codex-k8s"},
		"trigger":{"kind":"ai_repair"},
		"issue":{"number":45},
		"runtime":{"mode":"code-only","namespace":"codex-k8s-prod"},
		"agent":{"key":"sre","name":"AI SRE"}
	}`)
	runs := &fakeRunQueue{
		running: []runqueuerepo.RunningRun{
			{
				RunID:         "run-ai-repair",
				CorrelationID: "corr-ai-repair",
				ProjectID:     "proj-1",
				SlotNo:        1,
				RunPayload:    payload,
			},
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{states: map[string]JobState{"run-ai-repair": JobStateSucceeded}}
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
	svc.now = func() time.Time { return time.Date(2026, 2, 24, 12, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(runs.finished) != 1 {
		t.Fatalf("expected 1 finished run, got %d", len(runs.finished))
	}
	if runs.finished[0].Status != rundomain.StatusSucceeded {
		t.Fatalf("expected ai-repair run to finish as succeeded, got %s", runs.finished[0].Status)
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
		WorkerID:               "worker-1",
		ClaimLimit:             1,
		RunningCheckLimit:      10,
		SlotsPerProject:        2,
		SlotLeaseTTL:           time.Minute,
		RunNamespacePrefix:     "codex-issue",
		ControlPlaneMCPBaseURL: "http://codex-k8s-control-plane.test.svc:8081/mcp",
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
	if len(events.inserted) != 4 {
		t.Fatalf("expected run.namespace.prepared + run.namespace.ttl_scheduled + run.profile.resolved + run.started events, got %d", len(events.inserted))
	}
	if events.inserted[0].EventType != floweventdomain.EventTypeRunNamespacePrepared {
		t.Fatalf("expected first event %q, got %q", floweventdomain.EventTypeRunNamespacePrepared, events.inserted[0].EventType)
	}
	if events.inserted[1].EventType != floweventdomain.EventTypeRunNamespaceTTLScheduled {
		t.Fatalf("expected second event %q, got %q", floweventdomain.EventTypeRunNamespaceTTLScheduled, events.inserted[1].EventType)
	}
	if events.inserted[2].EventType != floweventdomain.EventTypeRunProfileResolved {
		t.Fatalf("expected third event %q, got %q", floweventdomain.EventTypeRunProfileResolved, events.inserted[2].EventType)
	}
	if events.inserted[3].EventType != floweventdomain.EventTypeRunStarted {
		t.Fatalf("expected fourth event %q, got %q", floweventdomain.EventTypeRunStarted, events.inserted[3].EventType)
	}
}

func TestTickFinalizesFullEnvRunRetainsNamespaceByTTL(t *testing.T) {
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
	runStatus := &fakeRunStatusNotifier{}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:           "worker-1",
		ClaimLimit:         1,
		RunningCheckLimit:  10,
		SlotsPerProject:    2,
		SlotLeaseTTL:       time.Minute,
		RunNamespacePrefix: "codex-issue",
	}, Dependencies{
		Runs:      runs,
		Events:    events,
		Launcher:  launcher,
		RunStatus: runStatus,
		Logger:    logger,
	})
	svc.now = func() time.Time { return time.Date(2026, 2, 11, 11, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}
	if len(events.inserted) != 1 || events.inserted[0].EventType != floweventdomain.EventTypeRunSucceeded {
		t.Fatalf("expected one run.succeeded event, got %#v", events.inserted)
	}
	if len(runStatus.upserts) < 2 {
		t.Fatalf("expected finished + retained-namespace status upserts, got %d", len(runStatus.upserts))
	}
	if got := runStatus.upserts[len(runStatus.upserts)-1]; got.Phase != RunStatusPhaseNamespaceDeleted || got.Deleted {
		t.Fatalf("expected retained namespace marker (phase=%q deleted=false), got phase=%q deleted=%v", RunStatusPhaseNamespaceDeleted, got.Phase, got.Deleted)
	}
}

func TestTickLaunchesReviseRunReusesNamespaceAndExtendsTTL(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"repository":{"full_name":"codex-k8s/codex-k8s"},"trigger":{"kind":"dev_revise"},"issue":{"number":74},"agent":{"key":"dev","name":"AI Developer"}}`)
	runs := &fakeRunQueue{
		claims: []runqueuerepo.ClaimedRun{
			{RunID: "run-revise", CorrelationID: "corr-revise", ProjectID: "550e8400-e29b-41d4-a716-446655440000", RunPayload: payload, SlotNo: 1},
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{
		states:        map[string]JobState{},
		reusableFound: true,
		reusable: NamespaceReuseResult{
			Namespace: "codex-k8s-dev-1",
			ExpiresAt: time.Date(2026, 2, 11, 15, 0, 0, 0, time.UTC),
		},
		ensureResult: NamespaceEnsureResult{Reused: true, LeaseExpiresAt: time.Date(2026, 2, 12, 10, 0, 0, 0, time.UTC)},
	}
	mcpTokens := &fakeMCPTokenIssuer{token: "token-run-revise"}
	deployer := &fakeRuntimePreparer{
		result: PrepareRunEnvironmentResult{
			Namespace: "codex-k8s-dev-1",
			TargetEnv: "ai",
		},
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:           "worker-1",
		ClaimLimit:         1,
		RunningCheckLimit:  10,
		SlotsPerProject:    2,
		SlotLeaseTTL:       time.Minute,
		RunNamespacePrefix: "codex-issue",
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
	if len(deployer.prepared) != 1 {
		t.Fatalf("expected one runtime prepare call, got %d", len(deployer.prepared))
	}
	if got, want := deployer.prepared[0].Namespace, "codex-k8s-dev-1"; got != want {
		t.Fatalf("expected revise run to reuse namespace %q, got %q", want, got)
	}
	if len(events.inserted) < 3 {
		t.Fatalf("expected namespace prepared + ttl extended + run started events, got %d", len(events.inserted))
	}
	if events.inserted[1].EventType != floweventdomain.EventTypeRunNamespaceTTLExtended {
		t.Fatalf("expected second event %q, got %q", floweventdomain.EventTypeRunNamespaceTTLExtended, events.inserted[1].EventType)
	}
}

func TestTickCleanupExpiredNamespaces_UpdatesRunStatusComment(t *testing.T) {
	t.Parallel()

	runs := &fakeRunQueue{}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{
		states: map[string]JobState{},
		expiredCleanups: []NamespaceCleanupResult{
			{
				Namespace: "codex-k8s-dev-1",
				RunID:     "run-expired",
				ExpiresAt: time.Date(2026, 2, 11, 9, 0, 0, 0, time.UTC),
			},
		},
	}
	runStatus := &fakeRunStatusNotifier{}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:          "worker-1",
		ClaimLimit:        1,
		RunningCheckLimit: 10,
		SlotsPerProject:   2,
		SlotLeaseTTL:      time.Minute,
	}, Dependencies{
		Runs:      runs,
		Events:    events,
		Launcher:  launcher,
		RunStatus: runStatus,
		Logger:    logger,
	})
	svc.now = func() time.Time { return time.Date(2026, 2, 11, 10, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}
	if len(runStatus.upserts) != 1 {
		t.Fatalf("expected one status upsert for ttl cleanup, got %d", len(runStatus.upserts))
	}
	if got := runStatus.upserts[0]; got.RunID != "run-expired" || got.Phase != RunStatusPhaseNamespaceDeleted || !got.Deleted {
		t.Fatalf("unexpected ttl cleanup run-status params: %+v", got)
	}
}

func TestTickPendingFullEnvPreparing_DoesNotBlockNextClaim(t *testing.T) {
	t.Parallel()

	fullEnvPayload := json.RawMessage(`{"repository":{"full_name":"codex-k8s/codex-k8s"},"trigger":{"kind":"dev"},"issue":{"number":74},"agent":{"key":"dev","name":"AI Developer"}}`)
	codeOnlyPayload := json.RawMessage(`{"repository":{"full_name":"codex-k8s/codex-k8s"},"issue":{"number":75},"agent":{"key":"km","name":"Knowledge Manager"}}`)
	runs := &fakeRunQueue{
		claims: []runqueuerepo.ClaimedRun{
			{
				RunID:         "run-fullenv",
				CorrelationID: "corr-fullenv",
				ProjectID:     "550e8400-e29b-41d4-a716-446655440000",
				RunPayload:    fullEnvPayload,
				SlotNo:        1,
			},
			{
				RunID:         "run-code-only",
				CorrelationID: "corr-code-only",
				ProjectID:     "550e8400-e29b-41d4-a716-446655440000",
				RunPayload:    codeOnlyPayload,
				SlotNo:        2,
			},
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{states: map[string]JobState{}}
	deployer := &fakeRuntimePreparer{result: PrepareRunEnvironmentResult{}}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:          "worker-1",
		ClaimLimit:        2,
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
	svc.now = func() time.Time { return time.Date(2026, 2, 20, 1, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(deployer.prepared) != 1 {
		t.Fatalf("expected 1 runtime prepare call for first full-env run, got %d", len(deployer.prepared))
	}
	if len(runs.finished) != 1 {
		t.Fatalf("expected second claimed code-only run to be finished, got %d finished runs", len(runs.finished))
	}
	if runs.finished[0].RunID != "run-code-only" {
		t.Fatalf("expected finished run %q, got %q", "run-code-only", runs.finished[0].RunID)
	}
	if runs.finished[0].Status != rundomain.StatusSucceeded {
		t.Fatalf("expected code-only run to finish as succeeded, got %s", runs.finished[0].Status)
	}
}

func TestTickRunningFullEnvJobNotFound_RuntimePreparingKeepsRunRunning(t *testing.T) {
	t.Parallel()

	fullEnvPayload := json.RawMessage(`{"repository":{"full_name":"codex-k8s/codex-k8s"},"trigger":{"kind":"dev"},"issue":{"number":74},"agent":{"key":"dev","name":"AI Developer"}}`)
	runs := &fakeRunQueue{
		running: []runqueuerepo.RunningRun{
			{
				RunID:         "run-fullenv",
				CorrelationID: "corr-fullenv",
				ProjectID:     "550e8400-e29b-41d4-a716-446655440000",
				SlotNo:        1,
				RunPayload:    fullEnvPayload,
				StartedAt:     time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	events := &fakeFlowEvents{}
	launcher := &fakeLauncher{states: map[string]JobState{
		"run-fullenv": JobStateNotFound,
	}}
	deployer := &fakeRuntimePreparer{result: PrepareRunEnvironmentResult{}}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	svc := NewService(Config{
		WorkerID:                   "worker-1",
		ClaimLimit:                 1,
		RunningCheckLimit:          10,
		SlotsPerProject:            2,
		SlotLeaseTTL:               time.Minute,
		RuntimePrepareRetryTimeout: 5 * time.Minute,
	}, Dependencies{
		Runs:            runs,
		Events:          events,
		Launcher:        launcher,
		RuntimePreparer: deployer,
		Logger:          logger,
	})
	svc.now = func() time.Time { return time.Date(2026, 2, 20, 1, 0, 0, 0, time.UTC) }

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}

	if len(deployer.prepared) != 1 {
		t.Fatalf("expected one runtime prepare poll call, got %d", len(deployer.prepared))
	}
	if runs.claimRunningCalls != 1 {
		t.Fatalf("expected one running-claim call, got %d", runs.claimRunningCalls)
	}
	if got := runs.claimRunning[0].WorkerID; got != "worker-1" {
		t.Fatalf("expected running claim worker_id %q, got %q", "worker-1", got)
	}
	if len(runs.finished) != 0 {
		t.Fatalf("expected run to stay running while runtime prepare is in progress, got %d finished runs", len(runs.finished))
	}
	if len(events.inserted) != 0 {
		t.Fatalf("expected no terminal events while runtime prepare is in progress, got %d", len(events.inserted))
	}
}

type fakeRunQueue struct {
	claims            []runqueuerepo.ClaimedRun
	claimCalls        int
	running           []runqueuerepo.RunningRun
	claimRunningCalls int
	claimRunning      []runqueuerepo.ClaimRunningParams
	finished          []runqueuerepo.FinishParams
	extended          []runqueuerepo.ExtendLeaseParams
	claimErr          error
	claimRunningErr   error
	listErr           error
	finishErr         error
	extendErr         error
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

func (f *fakeRunQueue) ClaimRunning(_ context.Context, params runqueuerepo.ClaimRunningParams) ([]runqueuerepo.RunningRun, error) {
	if f.claimRunningErr != nil {
		return nil, f.claimRunningErr
	}
	f.claimRunningCalls++
	f.claimRunning = append(f.claimRunning, params)
	return f.running, nil
}

func (f *fakeRunQueue) ExtendLease(_ context.Context, params runqueuerepo.ExtendLeaseParams) (bool, error) {
	return appendIfNoError(&f.extended, params, f.extendErr)
}

func (f *fakeRunQueue) FinishRun(_ context.Context, params runqueuerepo.FinishParams) (bool, error) {
	return appendIfNoError(&f.finished, params, f.finishErr)
}

func appendIfNoError[T any](dst *[]T, value T, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	*dst = append(*dst, value)
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
	states           map[string]JobState
	launched         []JobSpec
	prepared         []NamespaceSpec
	ensureResult     NamespaceEnsureResult
	reusable         NamespaceReuseResult
	reusableFound    bool
	expiredCleanups  []NamespaceCleanupResult
	cleanupSweepCall []NamespaceCleanupParams
	launchErr        error
	statusErr        error
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

type fakeRunStatusNotifier struct {
	upserts []RunStatusCommentParams
}

func (f *fakeRunStatusNotifier) UpsertRunStatusComment(_ context.Context, params RunStatusCommentParams) (RunStatusCommentResult, error) {
	f.upserts = append(f.upserts, params)
	return RunStatusCommentResult{CommentID: int64(len(f.upserts))}, nil
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

func (f *fakeLauncher) FindRunJobRefByRunID(_ context.Context, _ string) (JobRef, bool, error) {
	return JobRef{}, false, nil
}

func (f *fakeLauncher) FindReusableNamespace(_ context.Context, _ NamespaceReuseLookup) (NamespaceReuseResult, bool, error) {
	if !f.reusableFound {
		return NamespaceReuseResult{}, false, nil
	}
	return f.reusable, true, nil
}

func (f *fakeLauncher) EnsureNamespace(_ context.Context, spec NamespaceSpec) (NamespaceEnsureResult, error) {
	f.prepared = append(f.prepared, spec)
	if f.ensureResult.LeaseExpiresAt.IsZero() && spec.LeaseExpiresAt.IsZero() {
		f.ensureResult.LeaseExpiresAt = time.Now().UTC().Add(24 * time.Hour)
	}
	if f.ensureResult.LeaseExpiresAt.IsZero() {
		f.ensureResult.LeaseExpiresAt = spec.LeaseExpiresAt
	}
	return f.ensureResult, nil
}

func (f *fakeLauncher) CleanupExpiredNamespaces(_ context.Context, params NamespaceCleanupParams) ([]NamespaceCleanupResult, error) {
	f.cleanupSweepCall = append(f.cleanupSweepCall, params)
	if len(f.expiredCleanups) == 0 {
		return nil, nil
	}
	out := make([]NamespaceCleanupResult, len(f.expiredCleanups))
	copy(out, f.expiredCleanups)
	return out, nil
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
