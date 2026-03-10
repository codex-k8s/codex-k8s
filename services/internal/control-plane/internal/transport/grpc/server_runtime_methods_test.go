package grpc

import (
	"context"
	"testing"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStopRuntimeDeployTask_RequiresForce(t *testing.T) {
	t.Parallel()

	srv := &Server{}
	_, err := srv.StopRuntimeDeployTask(context.Background(), &controlplanev1.StopRuntimeDeployTaskRequest{})
	if err == nil {
		t.Fatal("expected invalid argument, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status, got %T", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %s", st.Code())
	}
	if st.Message() != "force must be true for stop" {
		t.Fatalf("unexpected message: %q", st.Message())
	}
}

func TestToStatus_MapsNotFound(t *testing.T) {
	t.Parallel()

	err := toStatus(errs.NotFound{Msg: "run_id: not found"})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status, got %T", err)
	}
	if st.Code() != codes.NotFound {
		t.Fatalf("expected NotFound, got %s", st.Code())
	}
	if st.Message() != "run_id: not found" {
		t.Fatalf("unexpected message: %q", st.Message())
	}
}
