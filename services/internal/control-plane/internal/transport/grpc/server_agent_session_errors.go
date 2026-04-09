package grpc

import (
	"strconv"

	agentsessionrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/agentsession"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const agentSessionSnapshotVersionConflictReason = "AGENT_SESSION_SNAPSHOT_VERSION_CONFLICT"

func agentSessionSnapshotVersionConflictStatus(err agentsessionrepo.SnapshotVersionConflict) error {
	st := status.New(codes.AlreadyExists, err.Error())
	details, detailsErr := st.WithDetails(&errdetails.ErrorInfo{
		Reason: agentSessionSnapshotVersionConflictReason,
		Metadata: map[string]string{
			"expected_snapshot_version": strconv.FormatInt(err.ExpectedSnapshotVersion, 10),
			"actual_snapshot_version":   strconv.FormatInt(err.ActualSnapshotVersion, 10),
		},
	})
	if detailsErr != nil {
		return st.Err()
	}
	return details.Err()
}
