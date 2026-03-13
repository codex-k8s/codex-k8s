package worker

import (
	"context"
	"errors"
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
				RepositoryFullName:   "codex-k8s/codex-k8s",
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
		MissionControlEnabled:             true,
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
				RepositoryFullName:   "codex-k8s/codex-k8s",
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
		MissionControlEnabled:             true,
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
		WorkerID:              "worker-1",
		MissionControlEnabled: true,
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
			{ProjectID: "proj-1", ProjectName: "Project", RepositoryFullName: "codex-k8s/codex-k8s"},
		},
	}

	svc := NewService(Config{
		WorkerID:                     "worker-1",
		MissionControlEnabled:        true,
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

type fakeMissionControlClient struct {
	warmupProjects   []MissionControlWarmupProject
	pendingCommands  []MissionControlPendingCommand
	executeErrors    []error
	queueCalls       []MissionControlQueueCommandParams
	pendingSyncCalls []MissionControlPendingSyncParams
	reconciledCalls  []MissionControlReconciledParams
	failedCalls      []MissionControlFailedParams
	executeCalls     []NextStepExecuteParams
	runWarmupCalls   []struct {
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
	return MissionControlWarmupResult{ProjectID: projectID, EntityCount: 3}, nil
}

func (f *fakeMissionControlClient) ListMissionControlPendingCommands(_ context.Context, _ int) ([]MissionControlPendingCommand, error) {
	return append([]MissionControlPendingCommand(nil), f.pendingCommands...), nil
}

func (f *fakeMissionControlClient) QueueMissionControlCommand(_ context.Context, params MissionControlQueueCommandParams) (MissionControlCommandState, error) {
	f.queueCalls = append(f.queueCalls, params)
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
