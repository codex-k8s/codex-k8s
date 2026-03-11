package grpc

import (
	"context"
	"testing"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	agentsessionrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentsession"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
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

func TestAgentSessionSnapshotVersionConflictStatus(t *testing.T) {
	t.Parallel()

	err := agentSessionSnapshotVersionConflictStatus(agentsessionrepo.SnapshotVersionConflict{
		ExpectedSnapshotVersion: 2,
		ActualSnapshotVersion:   4,
	})

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status, got %T", err)
	}
	if st.Code() != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists, got %s", st.Code())
	}

	details := st.Details()
	if len(details) != 1 {
		t.Fatalf("expected one detail, got %d", len(details))
	}

	info, ok := details[0].(*errdetails.ErrorInfo)
	if !ok {
		t.Fatalf("expected ErrorInfo detail, got %T", details[0])
	}
	if info.Reason != agentSessionSnapshotVersionConflictReason {
		t.Fatalf("unexpected reason %q", info.Reason)
	}
	if info.Metadata["actual_snapshot_version"] != "4" {
		t.Fatalf("unexpected actual_snapshot_version %q", info.Metadata["actual_snapshot_version"])
	}
}
