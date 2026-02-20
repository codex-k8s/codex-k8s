package worker

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIsRuntimeDeployTaskCanceledError(t *testing.T) {
	t.Parallel()

	err := status.Error(codes.Canceled, "runtime deploy task canceled for run_id=test")
	if !isRuntimeDeployTaskCanceledError(err) {
		t.Fatal("expected canceled runtime deploy task error to be detected")
	}
}

func TestIsRetryableRuntimeDeployError_TaskCanceledIsNotRetryable(t *testing.T) {
	t.Parallel()

	err := status.Error(codes.Canceled, "runtime deploy task canceled for run_id=test")
	if isRetryableRuntimeDeployError(err) {
		t.Fatal("expected canceled runtime deploy task error to be non-retryable")
	}
}
