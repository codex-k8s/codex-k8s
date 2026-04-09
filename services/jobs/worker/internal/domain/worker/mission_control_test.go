package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestReconcileMissionControlExecutesAcceptedStageNextStep(t *testing.T) {
	t.Parallel()

	missionControl := &fakeMissionControlClient{
		pendingCommands: []MissionControlPendingCommand{
			{
				ProjectID:            "proj-1",
				CommandID:            "cmd-1",
				Status:               "accepted",
				EffectiveCommandKind: "stage.next_step.execute",
				RepositoryFullName:   "codex-k8s/kodex",
				StageNextStep: &MissionControlStageNextStepPayload{
					ThreadKind:  "issue",
					ThreadNo:    371,
					TargetLabel: "run:qa",
				},
			},
		},
	}

	svc := NewService(Config{
		WorkerID:                          "worker-1",
		MissionControlPendingCommandLimit: 10,
		MissionControlRetryMaxAttempts:    3,
		MissionControlRetryBaseInterval:   time.Millisecond,
	}, Dependencies{
		MissionControl: missionControl,
	})

	if err := svc.reconcileMissionControl(context.Background()); err != nil {
		t.Fatalf("reconcileMissionControl() error = %v", err)
	}
	if len(missionControl.queueCalls) != 1 {
		t.Fatalf("expected 1 queue call, got %d", len(missionControl.queueCalls))
	}
	if len(missionControl.executeCalls) != 1 {
		t.Fatalf("expected 1 execute call, got %d", len(missionControl.executeCalls))
	}
	if len(missionControl.pendingSyncCalls) != 1 {
		t.Fatalf("expected 1 pending_sync call, got %d", len(missionControl.pendingSyncCalls))
	}
	if len(missionControl.reconciledCalls) != 1 {
		t.Fatalf("expected 1 reconciled call, got %d", len(missionControl.reconciledCalls))
	}
	if len(missionControl.failedCalls) != 0 {
		t.Fatalf("expected no failed calls, got %d", len(missionControl.failedCalls))
	}
}

func TestReconcileMissionControlFailsAfterRetryBudget(t *testing.T) {
	t.Parallel()

	missionControl := &fakeMissionControlClient{
		pendingCommands: []MissionControlPendingCommand{
			{
				ProjectID:            "proj-1",
				CommandID:            "cmd-2",
				Status:               "queued",
				EffectiveCommandKind: "stage.next_step.execute",
				RepositoryFullName:   "codex-k8s/kodex",
				StageNextStep: &MissionControlStageNextStepPayload{
					ThreadKind:  "issue",
					ThreadNo:    371,
					TargetLabel: "run:qa",
				},
			},
		},
		executeErrors: []error{
			errors.New("temporary github error"),
			errors.New("temporary github error"),
			errors.New("temporary github error"),
		},
	}

	svc := NewService(Config{
		WorkerID:                          "worker-1",
		MissionControlPendingCommandLimit: 10,
		MissionControlRetryMaxAttempts:    3,
		MissionControlRetryBaseInterval:   time.Millisecond,
	}, Dependencies{
		MissionControl: missionControl,
	})

	if err := svc.reconcileMissionControl(context.Background()); err != nil {
		t.Fatalf("reconcileMissionControl() error = %v", err)
	}
	if len(missionControl.executeCalls) != 3 {
		t.Fatalf("expected 3 execute attempts, got %d", len(missionControl.executeCalls))
	}
	if len(missionControl.failedCalls) != 1 {
		t.Fatalf("expected 1 failed call, got %d", len(missionControl.failedCalls))
	}
	if got := missionControl.failedCalls[0].FailureReason; got != "provider_error" {
		t.Fatalf("failure reason = %q, want provider_error", got)
	}
}

func TestReconcileMissionControlFailsUnsupportedCommandKind(t *testing.T) {
	t.Parallel()

	missionControl := &fakeMissionControlClient{
		pendingCommands: []MissionControlPendingCommand{
			{
				ProjectID:            "proj-1",
				CommandID:            "cmd-3",
				Status:               "accepted",
				EffectiveCommandKind: "discussion.create",
			},
		},
	}

	svc := NewService(Config{
		WorkerID: "worker-1",
	}, Dependencies{
		MissionControl: missionControl,
	})

	if err := svc.reconcileMissionControl(context.Background()); err != nil {
		t.Fatalf("reconcileMissionControl() error = %v", err)
	}
	if len(missionControl.executeCalls) != 0 {
		t.Fatalf("expected no execute calls, got %d", len(missionControl.executeCalls))
	}
	if len(missionControl.failedCalls) != 1 {
		t.Fatalf("expected 1 failed call, got %d", len(missionControl.failedCalls))
	}
	if got := missionControl.failedCalls[0].FailureReason; got != "unknown" {
		t.Fatalf("failure reason = %q, want unknown", got)
	}
}

func TestReconcileMissionControlWarmupUsesThrottle(t *testing.T) {
	t.Parallel()

	missionControl := &fakeMissionControlClient{
		warmupProjects: []MissionControlWarmupProject{
			{ProjectID: "proj-1", ProjectName: "Project", RepositoryFullName: "codex-k8s/kodex"},
		},
	}

	svc := NewService(Config{
		WorkerID:                     "worker-1",
		MissionControlWarmupInterval: time.Hour,
	}, Dependencies{
		MissionControl: missionControl,
	})
	svc.now = func() time.Time { return time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC) }

	if err := svc.reconcileMissionControl(context.Background()); err != nil {
		t.Fatalf("first reconcileMissionControl() error = %v", err)
	}
	if err := svc.reconcileMissionControl(context.Background()); err != nil {
		t.Fatalf("second reconcileMissionControl() error = %v", err)
	}
	if len(missionControl.runWarmupCalls) != 1 {
		t.Fatalf("expected 1 warmup call, got %d", len(missionControl.runWarmupCalls))
	}
}

func TestReconcileMissionControlContinuesCommandsWhenWarmupFails(t *testing.T) {
	t.Parallel()

	missionControl := &fakeMissionControlClient{
		warmupProjects: []MissionControlWarmupProject{
			{ProjectID: "proj-1", ProjectName: "Project", RepositoryFullName: "codex-k8s/kodex"},
		},
		warmupErrorsByProject: map[string]error{
			"proj-1": errors.New("warmup failed"),
		},
		pendingCommands: []MissionControlPendingCommand{
			{
				ProjectID:            "proj-1",
				CommandID:            "cmd-1",
				Status:               "accepted",
				EffectiveCommandKind: "stage.next_step.execute",
				RepositoryFullName:   "codex-k8s/kodex",
				StageNextStep: &MissionControlStageNextStepPayload{
					ThreadKind:  "issue",
					ThreadNo:    544,
					TargetLabel: "run:qa",
				},
			},
		},
	}

	svc := NewService(Config{
		WorkerID:                          "worker-1",
		MissionControlPendingCommandLimit: 10,
		MissionControlRetryMaxAttempts:    1,
	}, Dependencies{
		MissionControl: missionControl,
	})

	if err := svc.reconcileMissionControl(context.Background()); err == nil {
		t.Fatal("reconcileMissionControl() error = nil, want joined warmup error")
	}
	if len(missionControl.runWarmupCalls) != 1 {
		t.Fatalf("expected 1 warmup call, got %d", len(missionControl.runWarmupCalls))
	}
	if len(missionControl.executeCalls) != 1 {
		t.Fatalf("expected command processing to continue after warmup failure, got %d execute calls", len(missionControl.executeCalls))
	}
}

func TestReconcileMissionControlWarmupsDoesNotLogPerProjectFailure(t *testing.T) {
	t.Parallel()

	logger, buf := newMissionControlTestLogger()
	missionControl := &fakeMissionControlClient{
		warmupProjects: []MissionControlWarmupProject{
			{ProjectID: "proj-1", ProjectName: "Project", RepositoryFullName: "codex-k8s/kodex"},
		},
		warmupErrorsByProject: map[string]error{
			"proj-1": errors.New("warmup failed"),
		},
	}

	svc := NewService(Config{
		WorkerID:                     "worker-1",
		MissionControlWarmupInterval: time.Hour,
	}, Dependencies{
		Logger:         logger,
		MissionControl: missionControl,
	})

	err := svc.reconcileMissionControlWarmups(context.Background())
	if err == nil {
		t.Fatal("reconcileMissionControlWarmups() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "project proj-1 (codex-k8s/kodex)") {
		t.Fatalf("reconcileMissionControlWarmups() error = %q, want project context", err)
	}
	if got := strings.TrimSpace(buf.String()); got != "" {
		t.Fatalf("expected no per-project logs below boundary, got %q", got)
	}
}

func TestReconcileMissionControlCommandsContinueAfterQueueFailure(t *testing.T) {
	t.Parallel()

	missionControl := &fakeMissionControlClient{
		pendingCommands: []MissionControlPendingCommand{
			{
				ProjectID:            "proj-1",
				CommandID:            "cmd-broken",
				Status:               "accepted",
				EffectiveCommandKind: "stage.next_step.execute",
				RepositoryFullName:   "codex-k8s/kodex",
				StageNextStep: &MissionControlStageNextStepPayload{
					ThreadKind:  "issue",
					ThreadNo:    544,
					TargetLabel: "run:qa",
				},
			},
			{
				ProjectID:            "proj-1",
				CommandID:            "cmd-ok",
				Status:               "accepted",
				EffectiveCommandKind: "stage.next_step.execute",
				RepositoryFullName:   "codex-k8s/kodex",
				StageNextStep: &MissionControlStageNextStepPayload{
					ThreadKind:  "issue",
					ThreadNo:    545,
					TargetLabel: "run:qa",
				},
			},
		},
		queueErrorsByCommand: map[string]error{
			"cmd-broken": errors.New("queue failed"),
		},
	}

	svc := NewService(Config{
		WorkerID:                          "worker-1",
		MissionControlPendingCommandLimit: 10,
		MissionControlRetryMaxAttempts:    1,
	}, Dependencies{
		MissionControl: missionControl,
	})

	if err := svc.reconcileMissionControlCommands(context.Background()); err == nil {
		t.Fatal("reconcileMissionControlCommands() error = nil, want joined queue error")
	}
	if len(missionControl.queueCalls) != 2 {
		t.Fatalf("expected both commands to be queued, got %d calls", len(missionControl.queueCalls))
	}
	if len(missionControl.executeCalls) != 1 {
		t.Fatalf("expected second command to continue after queue failure, got %d execute calls", len(missionControl.executeCalls))
	}
}

func TestReconcileMissionControlCommandsDoesNotLogPerCommandFailure(t *testing.T) {
	t.Parallel()

	logger, buf := newMissionControlTestLogger()
	missionControl := &fakeMissionControlClient{
		pendingCommands: []MissionControlPendingCommand{
			{
				ProjectID:            "proj-1",
				CommandID:            "cmd-broken",
				Status:               "accepted",
				EffectiveCommandKind: "stage.next_step.execute",
				RepositoryFullName:   "codex-k8s/kodex",
				StageNextStep: &MissionControlStageNextStepPayload{
					ThreadKind:  "issue",
					ThreadNo:    544,
					TargetLabel: "run:qa",
				},
			},
		},
		queueErrorsByCommand: map[string]error{
			"cmd-broken": errors.New("queue failed"),
		},
	}

	svc := NewService(Config{
		WorkerID:                          "worker-1",
		MissionControlPendingCommandLimit: 10,
		MissionControlRetryMaxAttempts:    1,
	}, Dependencies{
		Logger:         logger,
		MissionControl: missionControl,
	})

	err := svc.reconcileMissionControlCommands(context.Background())
	if err == nil {
		t.Fatal("reconcileMissionControlCommands() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "project proj-1 command cmd-broken (stage.next_step.execute)") {
		t.Fatalf("reconcileMissionControlCommands() error = %q, want command context", err)
	}
	if got := strings.TrimSpace(buf.String()); got != "" {
		t.Fatalf("expected no per-command logs below boundary, got %q", got)
	}
}

func TestLogMissionControlWarmupResultUsesInfoForOpenRolloutGates(t *testing.T) {
	t.Parallel()

	logger, buf := newMissionControlTestLogger()
	svc := NewService(Config{
		WorkerID: "worker-1",
	}, Dependencies{
		Logger: logger,
	})

	svc.logMissionControlWarmupResult(
		MissionControlWarmupProject{
			ProjectID:          "proj-1",
			ProjectName:        "Project",
			RepositoryFullName: "codex-k8s/kodex",
		},
		"corr-1",
		MissionControlWarmupResult{
			ProjectID:               "proj-1",
			EntityCount:             3,
			RunEntityCount:          1,
			ContinuityGapCount:      0,
			OpenGapCount:            0,
			BlockingGapCount:        0,
			WatermarkCount:          1,
			ReadyForReconcile:       true,
			ReadyForTransport:       false,
			TransportGatingReason:   "provider_coverage_not_ready",
			ProviderFreshnessStatus: "ready",
			ProviderCoverageStatus:  "out_of_scope",
			GraphProjectionStatus:   "ready",
			LaunchPolicyStatus:      "ready",
		},
	)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 log line, got %d", len(lines))
	}

	var record map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &record); err != nil {
		t.Fatalf("json.Unmarshal() log record error = %v", err)
	}
	if got, want := record["level"], "INFO"; got != want {
		t.Fatalf("log level = %v, want %q", got, want)
	}
	if got, want := record["msg"], "mission control warmup completed with open rollout gates"; got != want {
		t.Fatalf("log message = %v, want %q", got, want)
	}
}

func newMissionControlTestLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewJSONHandler(&buf, nil)), &buf
}

type fakeMissionControlClient struct {
	warmupProjects        []MissionControlWarmupProject
	warmupErrorsByProject map[string]error
	pendingCommands       []MissionControlPendingCommand
	claimErr              error
	executeErrors         []error
	queueErrorsByCommand  map[string]error
	queueCalls            []MissionControlQueueCommandParams
	pendingSyncCalls      []MissionControlPendingSyncParams
	reconciledCalls       []MissionControlReconciledParams
	failedCalls           []MissionControlFailedParams
	executeCalls          []NextStepExecuteParams
	runWarmupCalls        []struct {
		projectID     string
		requestedBy   string
		correlationID string
		forceRebuild  bool
	}
}

func (f *fakeMissionControlClient) ListMissionControlWarmupProjects(_ context.Context, _ int) ([]MissionControlWarmupProject, error) {
	return append([]MissionControlWarmupProject(nil), f.warmupProjects...), nil
}

func (f *fakeMissionControlClient) RunMissionControlWarmup(_ context.Context, projectID string, requestedBy string, correlationID string, forceRebuild bool) (MissionControlWarmupResult, error) {
	f.runWarmupCalls = append(f.runWarmupCalls, struct {
		projectID     string
		requestedBy   string
		correlationID string
		forceRebuild  bool
	}{
		projectID:     projectID,
		requestedBy:   requestedBy,
		correlationID: correlationID,
		forceRebuild:  forceRebuild,
	})
	if err, ok := f.warmupErrorsByProject[projectID]; ok {
		return MissionControlWarmupResult{}, err
	}
	return MissionControlWarmupResult{ProjectID: projectID, EntityCount: 3}, nil
}

func (f *fakeMissionControlClient) ClaimMissionControlPendingCommands(_ context.Context, _ string, _ time.Duration, _ int) ([]MissionControlPendingCommand, error) {
	if f.claimErr != nil {
		return nil, f.claimErr
	}
	return append([]MissionControlPendingCommand(nil), f.pendingCommands...), nil
}

func (f *fakeMissionControlClient) QueueMissionControlCommand(_ context.Context, params MissionControlQueueCommandParams) (MissionControlCommandState, error) {
	f.queueCalls = append(f.queueCalls, params)
	if err, ok := f.queueErrorsByCommand[params.CommandID]; ok {
		return MissionControlCommandState{}, err
	}
	return MissionControlCommandState{ProjectID: params.ProjectID, CommandID: params.CommandID, Status: "queued"}, nil
}

func (f *fakeMissionControlClient) MarkMissionControlCommandPendingSync(_ context.Context, params MissionControlPendingSyncParams) (MissionControlCommandState, error) {
	f.pendingSyncCalls = append(f.pendingSyncCalls, params)
	return MissionControlCommandState{ProjectID: params.ProjectID, CommandID: params.CommandID, Status: "pending_sync"}, nil
}

func (f *fakeMissionControlClient) MarkMissionControlCommandReconciled(_ context.Context, params MissionControlReconciledParams) (MissionControlCommandState, error) {
	f.reconciledCalls = append(f.reconciledCalls, params)
	return MissionControlCommandState{ProjectID: params.ProjectID, CommandID: params.CommandID, Status: "reconciled"}, nil
}

func (f *fakeMissionControlClient) MarkMissionControlCommandFailed(_ context.Context, params MissionControlFailedParams) (MissionControlCommandState, error) {
	f.failedCalls = append(f.failedCalls, params)
	return MissionControlCommandState{ProjectID: params.ProjectID, CommandID: params.CommandID, Status: "failed"}, nil
}

func (f *fakeMissionControlClient) ExecuteNextStepAction(_ context.Context, params NextStepExecuteParams) error {
	f.executeCalls = append(f.executeCalls, params)
	if len(f.executeErrors) == 0 {
		return nil
	}
	err := f.executeErrors[0]
	f.executeErrors = f.executeErrors[1:]
	return err
}
