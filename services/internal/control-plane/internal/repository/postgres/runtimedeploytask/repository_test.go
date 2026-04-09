package runtimedeploytask

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/codex-k8s/kodex/libs/go/errs"
	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
)

func TestParseRuntimeDeployStatus_Canceled(t *testing.T) {
	t.Parallel()

	got, err := parseRuntimeDeployStatus("canceled")
	if err != nil {
		t.Fatalf("parseRuntimeDeployStatus() error = %v", err)
	}
	if got != entitytypes.RuntimeDeployTaskStatusCanceled {
		t.Fatalf("unexpected status: got %q want %q", got, entitytypes.RuntimeDeployTaskStatusCanceled)
	}
}

func TestValidateStopActionTask_ExpiredLease(t *testing.T) {
	t.Parallel()

	err := validateStopActionTask(entitytypes.RuntimeDeployTask{
		Status:     entitytypes.RuntimeDeployTaskStatusRunning,
		LeaseOwner: "worker-1",
		LeaseUntil: time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC),
	}, time.Date(2026, time.March, 10, 12, 0, 1, 0, time.UTC))
	if err == nil {
		t.Fatal("expected failed precondition, got nil")
	}

	var failedPrecondition errs.FailedPrecondition
	if !errors.As(err, &failedPrecondition) {
		t.Fatalf("expected errs.FailedPrecondition, got %T", err)
	}
}

func TestHasActiveLease(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)
	if hasActiveLease(entitytypes.RuntimeDeployTask{
		LeaseOwner: "worker-1",
		LeaseUntil: now.Add(30 * time.Second),
	}, now) != true {
		t.Fatal("expected active lease")
	}
	if hasActiveLease(entitytypes.RuntimeDeployTask{
		LeaseOwner: "worker-1",
		LeaseUntil: now,
	}, now) {
		t.Fatal("expected lease equal to request time to be inactive")
	}
	if hasActiveLease(entitytypes.RuntimeDeployTask{
		LeaseOwner: "",
		LeaseUntil: now.Add(30 * time.Second),
	}, now) {
		t.Fatal("expected empty owner to be inactive")
	}
}

func TestRealtimeListQueriesSupportPaginationAndCount(t *testing.T) {
	t.Parallel()

	if !strings.Contains(queryListRecent, "OFFSET $4") {
		t.Fatal("list_recent query must support pagination offset")
	}
	if !strings.Contains(queryCountRecent, "COUNT(*)") {
		t.Fatal("count_recent query must count filtered tasks")
	}
}
