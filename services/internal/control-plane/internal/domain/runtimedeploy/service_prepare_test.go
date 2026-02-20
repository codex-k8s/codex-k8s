package runtimedeploy

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

func TestPrepareRunEnvironment_ReturnsCanceledError(t *testing.T) {
	t.Parallel()

	repo := &fakePrepareRuntimeDeployTasksRepo{
		upsertTask: runtimedeploytaskrepo.Task{
			RunID:   "run-canceled",
			Status:  entitytypes.RuntimeDeployTaskStatusPending,
		},
		getTask: runtimedeploytaskrepo.Task{
			RunID:     "run-canceled",
			Status:    entitytypes.RuntimeDeployTaskStatusCanceled,
			LastError: "superseded by newer deploy task",
		},
		getTaskFound: true,
	}
	svc := &Service{
		cfg:   Config{WaitPollInterval: time.Millisecond},
		tasks: repo,
	}

	_, err := svc.PrepareRunEnvironment(context.Background(), PrepareParams{RunID: "run-canceled"})
	if err == nil {
		t.Fatal("expected canceled error, got nil")
	}
	if !errors.Is(err, ErrTaskCanceled) {
		t.Fatalf("expected ErrTaskCanceled, got %v", err)
	}
	if !strings.Contains(err.Error(), "superseded by newer deploy task") {
		t.Fatalf("expected cancellation reason in error, got %q", err.Error())
	}
}

type fakePrepareRuntimeDeployTasksRepo struct {
	upsertTask    runtimedeploytaskrepo.Task
	upsertErr     error
	getTask       runtimedeploytaskrepo.Task
	getTaskFound  bool
	getTaskErr    error
}

func (f *fakePrepareRuntimeDeployTasksRepo) UpsertDesired(_ context.Context, _ runtimedeploytaskrepo.UpsertDesiredParams) (runtimedeploytaskrepo.Task, error) {
	if f.upsertErr != nil {
		return runtimedeploytaskrepo.Task{}, f.upsertErr
	}
	return f.upsertTask, nil
}

func (f *fakePrepareRuntimeDeployTasksRepo) GetByRunID(_ context.Context, _ string) (runtimedeploytaskrepo.Task, bool, error) {
	if f.getTaskErr != nil {
		return runtimedeploytaskrepo.Task{}, false, f.getTaskErr
	}
	return f.getTask, f.getTaskFound, nil
}

func (*fakePrepareRuntimeDeployTasksRepo) ClaimNext(_ context.Context, _ runtimedeploytaskrepo.ClaimParams) (runtimedeploytaskrepo.Task, bool, error) {
	return runtimedeploytaskrepo.Task{}, false, nil
}

func (*fakePrepareRuntimeDeployTasksRepo) MarkSucceeded(_ context.Context, _ runtimedeploytaskrepo.MarkSucceededParams) (bool, error) {
	return false, nil
}

func (*fakePrepareRuntimeDeployTasksRepo) MarkFailed(_ context.Context, _ runtimedeploytaskrepo.MarkFailedParams) (bool, error) {
	return false, nil
}

func (*fakePrepareRuntimeDeployTasksRepo) RenewLease(_ context.Context, _ runtimedeploytaskrepo.RenewLeaseParams) (bool, error) {
	return false, nil
}

func (*fakePrepareRuntimeDeployTasksRepo) Requeue(_ context.Context, _ runtimedeploytaskrepo.RequeueParams) (bool, error) {
	return false, nil
}

func (*fakePrepareRuntimeDeployTasksRepo) ListRecent(_ context.Context, _ runtimedeploytaskrepo.ListFilter) ([]runtimedeploytaskrepo.Task, error) {
	return nil, nil
}

func (*fakePrepareRuntimeDeployTasksRepo) AppendLog(_ context.Context, _ runtimedeploytaskrepo.AppendLogParams) error {
	return nil
}
