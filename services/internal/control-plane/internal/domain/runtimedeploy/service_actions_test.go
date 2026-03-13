package runtimedeploy

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

func TestRequestTaskAction_CancelPersistsAuditAndLog(t *testing.T) {
	t.Parallel()

	tasks := &fakeTaskActionRepo{
		requestActionResult: runtimedeploytaskrepo.RequestActionResult{
			Task: runtimedeploytaskrepo.Task{
				RunID:  "run-1",
				Status: entitytypes.RuntimeDeployTaskStatusCanceled,
			},
			PreviousStatus: entitytypes.RuntimeDeployTaskStatusRunning,
			CurrentStatus:  entitytypes.RuntimeDeployTaskStatusCanceled,
		},
	}
	events := &fakeFlowEventRecorder{}
	svc := &Service{
		tasks:      tasks,
		runs:       &fakeTaskActionRunReader{run: agentrunrepo.Run{ID: "run-1", CorrelationID: "corr-1"}},
		flowEvents: events,
	}

	result, err := svc.RequestTaskAction(context.Background(), TaskActionParams{
		RunID:  "run-1",
		Action: TaskActionCancel,
		Reason: "stuck deployment",
		Actor: TaskActionActor{
			UserID: "user-1",
			Email:  "operator@example.com",
		},
		RequestedAt: time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("RequestTaskAction() error = %v", err)
	}
	if result.CurrentStatus != entitytypes.RuntimeDeployTaskStatusCanceled {
		t.Fatalf("unexpected current status: got %q", result.CurrentStatus)
	}
	if tasks.requestActionParams.Action != TaskActionCancel {
		t.Fatalf("unexpected action passed to repo: got %q", tasks.requestActionParams.Action)
	}
	if tasks.requestActionParams.RequestedBy != "operator@example.com" {
		t.Fatalf("unexpected requested_by: got %q", tasks.requestActionParams.RequestedBy)
	}
	if !strings.Contains(tasks.requestActionParams.Reason, "stuck deployment") {
		t.Fatalf("repo reason must include operator reason, got %q", tasks.requestActionParams.Reason)
	}
	if tasks.appendLogParams.Stage != "control" {
		t.Fatalf("unexpected log stage: got %q", tasks.appendLogParams.Stage)
	}
	if len(events.inserted) != 1 {
		t.Fatalf("expected exactly one audit event, got %d", len(events.inserted))
	}
	if events.inserted[0].EventType != floweventdomain.EventTypeRuntimeDeployCancelRequested {
		t.Fatalf("unexpected audit event type: got %q", events.inserted[0].EventType)
	}
	var payload taskActionAuditPayload
	if err := json.Unmarshal(events.inserted[0].Payload, &payload); err != nil {
		t.Fatalf("unmarshal audit payload: %v", err)
	}
	if payload.Reason != "stuck deployment" {
		t.Fatalf("unexpected audit reason: got %q", payload.Reason)
	}
	if payload.RequestedByEmail != "operator@example.com" {
		t.Fatalf("unexpected audit email: got %q", payload.RequestedByEmail)
	}
}

func TestRequestTaskAction_RequiresActorIdentity(t *testing.T) {
	t.Parallel()

	svc := &Service{}
	_, err := svc.RequestTaskAction(context.Background(), TaskActionParams{
		RunID:  "run-1",
		Action: TaskActionStop,
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validation errs.Validation
	if !strings.Contains(err.Error(), "requested_by") && !strings.Contains(err.Error(), "actor identity") {
		t.Fatalf("expected actor validation message, got %v", err)
	}
	if !errors.As(err, &validation) {
		t.Fatalf("expected errs.Validation, got %T", err)
	}
}

type fakeTaskActionRepo struct {
	requestActionParams runtimedeploytaskrepo.RequestActionParams
	requestActionResult runtimedeploytaskrepo.RequestActionResult
	appendLogParams     runtimedeploytaskrepo.AppendLogParams
}

func (*fakeTaskActionRepo) UpsertDesired(_ context.Context, _ runtimedeploytaskrepo.UpsertDesiredParams) (runtimedeploytaskrepo.Task, error) {
	return runtimedeploytaskrepo.Task{}, nil
}

func (*fakeTaskActionRepo) GetByRunID(_ context.Context, _ string) (runtimedeploytaskrepo.Task, bool, error) {
	return runtimedeploytaskrepo.Task{}, false, nil
}

func (*fakeTaskActionRepo) FindActiveByNamespace(_ context.Context, _ string) (runtimedeploytaskrepo.Task, bool, error) {
	return runtimedeploytaskrepo.Task{}, false, nil
}

func (*fakeTaskActionRepo) ClaimNext(_ context.Context, _ runtimedeploytaskrepo.ClaimParams) (runtimedeploytaskrepo.Task, bool, error) {
	return runtimedeploytaskrepo.Task{}, false, nil
}

func (*fakeTaskActionRepo) MarkSucceeded(_ context.Context, _ runtimedeploytaskrepo.MarkSucceededParams) (bool, error) {
	return false, nil
}

func (*fakeTaskActionRepo) MarkFailed(_ context.Context, _ runtimedeploytaskrepo.MarkFailedParams) (bool, error) {
	return false, nil
}

func (*fakeTaskActionRepo) RenewLease(_ context.Context, _ runtimedeploytaskrepo.RenewLeaseParams) (bool, error) {
	return false, nil
}

func (*fakeTaskActionRepo) Requeue(_ context.Context, _ runtimedeploytaskrepo.RequeueParams) (bool, error) {
	return false, nil
}

func (f *fakeTaskActionRepo) RequestAction(_ context.Context, params runtimedeploytaskrepo.RequestActionParams) (runtimedeploytaskrepo.RequestActionResult, error) {
	f.requestActionParams = params
	return f.requestActionResult, nil
}

func (*fakeTaskActionRepo) ListRecent(_ context.Context, _ runtimedeploytaskrepo.ListFilter) ([]runtimedeploytaskrepo.Task, int, error) {
	return nil, 0, nil
}

func (f *fakeTaskActionRepo) AppendLog(_ context.Context, params runtimedeploytaskrepo.AppendLogParams) error {
	f.appendLogParams = params
	return nil
}

func (*fakeTaskActionRepo) CleanupTaskLogsUpdatedBefore(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}

type fakeTaskActionRunReader struct {
	run agentrunrepo.Run
}

func (f *fakeTaskActionRunReader) GetByID(_ context.Context, _ string) (agentrunrepo.Run, bool, error) {
	return f.run, true, nil
}

type fakeFlowEventRecorder struct {
	inserted []floweventrepo.InsertParams
}

func (f *fakeFlowEventRecorder) Insert(_ context.Context, params floweventrepo.InsertParams) error {
	f.inserted = append(f.inserted, params)
	return nil
}
