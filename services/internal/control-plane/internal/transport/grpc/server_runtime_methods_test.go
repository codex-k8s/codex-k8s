package grpc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	agentcallbackdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/agentcallback"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	agentsessionrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentsession"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

func TestGetRunInteractionResumePayload_ReturnsRunScopedPayload(t *testing.T) {
	t.Parallel()

	var gotRunID string
	srv := &Server{
		agentCallbacks: fakeRuntimeAgentCallbackService{
			getRunInteractionResumePayload: func(ctx context.Context, runID string) (json.RawMessage, bool, error) {
				gotRunID = runID
				return json.RawMessage(`{"interaction_id":"interaction-1"}`), true, nil
			},
		},
		mcp: fakeRuntimeMCPRunTokenService{
			verifyRunToken: func(ctx context.Context, rawToken string) (mcpdomain.SessionContext, error) {
				if rawToken != "token-1" {
					t.Fatalf("unexpected token %q", rawToken)
				}
				return mcpdomain.SessionContext{RunID: "run-1"}, nil
			},
		},
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer token-1"))
	resp, err := srv.GetRunInteractionResumePayload(ctx, &controlplanev1.GetRunInteractionResumePayloadRequest{})
	if err != nil {
		t.Fatalf("GetRunInteractionResumePayload() error = %v", err)
	}
	if gotRunID != "run-1" {
		t.Fatalf("runID = %q, want run-1", gotRunID)
	}
	if !resp.GetFound() {
		t.Fatal("expected found=true")
	}
	if got, want := string(resp.GetPayloadJson()), `{"interaction_id":"interaction-1"}`; got != want {
		t.Fatalf("payload_json = %q, want %q", got, want)
	}
}

type fakeRuntimeAgentCallbackService struct {
	getRunInteractionResumePayload func(ctx context.Context, runID string) (json.RawMessage, bool, error)
}

func (f fakeRuntimeAgentCallbackService) UpsertAgentSession(context.Context, agentcallbackdomain.UpsertAgentSessionParams) (agentcallbackdomain.UpsertAgentSessionResult, error) {
	return agentcallbackdomain.UpsertAgentSessionResult{}, nil
}

func (f fakeRuntimeAgentCallbackService) GetLatestAgentSession(context.Context, agentcallbackdomain.GetLatestAgentSessionQuery) (agentcallbackdomain.Session, bool, error) {
	return agentcallbackdomain.Session{}, false, nil
}

func (f fakeRuntimeAgentCallbackService) GetRunInteractionResumePayload(ctx context.Context, runID string) (json.RawMessage, bool, error) {
	if f.getRunInteractionResumePayload != nil {
		return f.getRunInteractionResumePayload(ctx, runID)
	}
	return nil, false, nil
}

func (f fakeRuntimeAgentCallbackService) LookupPullRequest(context.Context, agentcallbackdomain.LookupPullRequestQuery) (agentcallbackdomain.PullRequestLookupResult, bool, error) {
	return agentcallbackdomain.PullRequestLookupResult{}, false, nil
}

func (f fakeRuntimeAgentCallbackService) InsertRunFlowEvent(context.Context, agentcallbackdomain.InsertRunFlowEventParams) error {
	return nil
}

type fakeRuntimeMCPRunTokenService struct {
	verifyRunToken func(ctx context.Context, rawToken string) (mcpdomain.SessionContext, error)
}

func (f fakeRuntimeMCPRunTokenService) IssueRunToken(context.Context, mcpdomain.IssueRunTokenParams) (mcpdomain.IssuedToken, error) {
	return mcpdomain.IssuedToken{}, nil
}

func (f fakeRuntimeMCPRunTokenService) VerifyRunToken(ctx context.Context, rawToken string) (mcpdomain.SessionContext, error) {
	if f.verifyRunToken != nil {
		return f.verifyRunToken(ctx, rawToken)
	}
	return mcpdomain.SessionContext{}, nil
}

func (f fakeRuntimeMCPRunTokenService) VerifyInteractionCallbackToken(context.Context, string, string) (mcpdomain.SessionContext, error) {
	return mcpdomain.SessionContext{}, nil
}

func (f fakeRuntimeMCPRunTokenService) ListPendingApprovals(context.Context, int) ([]mcpdomain.ApprovalListItem, error) {
	return nil, nil
}

func (f fakeRuntimeMCPRunTokenService) ResolveApproval(context.Context, mcpdomain.ResolveApprovalParams) (mcpdomain.ResolveApprovalResult, error) {
	return mcpdomain.ResolveApprovalResult{}, nil
}

func (f fakeRuntimeMCPRunTokenService) ClaimNextInteractionDispatch(context.Context, mcpdomain.ClaimNextInteractionDispatchParams) (mcpdomain.InteractionDispatchClaim, bool, error) {
	return mcpdomain.InteractionDispatchClaim{}, false, nil
}

func (f fakeRuntimeMCPRunTokenService) CompleteInteractionDispatch(context.Context, mcpdomain.CompleteInteractionDispatchParams) (mcpdomain.CompleteInteractionDispatchResult, error) {
	return mcpdomain.CompleteInteractionDispatchResult{}, nil
}

func (f fakeRuntimeMCPRunTokenService) ExpireNextDueInteraction(context.Context) (mcpdomain.ExpireNextInteractionResult, bool, error) {
	return mcpdomain.ExpireNextInteractionResult{}, false, nil
}

func (f fakeRuntimeMCPRunTokenService) SubmitInteractionCallback(context.Context, mcpdomain.SubmitInteractionCallbackParams) (mcpdomain.SubmitInteractionCallbackResult, error) {
	return mcpdomain.SubmitInteractionCallbackResult{}, nil
}
