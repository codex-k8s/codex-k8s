package grpc

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	controlplanev1 "github.com/codex-k8s/kodex/proto/gen/go/kodex/controlplane/v1"
	changegovernancedomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/changegovernance"
	mcpdomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/mcp"
	agentrunrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/agentrun"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestReportChangeGovernanceDraftSignal_RejectsProjectScopeMismatch(t *testing.T) {
	t.Parallel()

	srv := NewServer(Dependencies{
		ChangeGovernance: fakeChangeGovernanceService{},
		MCP: fakeChangeGovernanceMCPService{
			session: mcpdomain.SessionContext{
				RunID:     "run-1",
				ProjectID: "project-from-token",
			},
		},
		Runs: fakeChangeGovernanceRunReader{
			run: agentrunrepo.Run{
				ID:            "run-1",
				ProjectID:     "project-from-token",
				CorrelationID: "corr-1",
				RunPayload: mustMarshalRunPayload(t, querytypes.RunPayload{
					Repository: querytypes.RunPayloadRepository{FullName: "codex-k8s/kodex"},
					Issue:      &querytypes.RunPayloadIssue{Number: 521},
				}),
			},
			found: true,
		},
	})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer test-token"))
	_, err := srv.ReportChangeGovernanceDraftSignal(ctx, &controlplanev1.ReportChangeGovernanceDraftSignalRequest{
		RunId:              "run-1",
		ProjectId:          "project-from-request",
		RepositoryFullName: "codex-k8s/kodex",
		IssueNumber:        521,
		SignalId:           "signal-1",
		CorrelationId:      "corr-1",
		DraftRef:           "draft-1",
		DraftKind:          "internal_working_draft",
		ChangeScopeHints: []*controlplanev1.ChangeGovernanceScopeHint{
			{ContextKey: "services/internal/control-plane", SurfaceKind: "domain"},
		},
		OccurredAt: timestamppb.New(time.Now().UTC()),
	})
	if err == nil {
		t.Fatal("expected permission denied error, got nil")
	}
	if got, want := status.Code(err), codes.PermissionDenied; got != want {
		t.Fatalf("status code = %s, want %s", got, want)
	}
}

func TestReportChangeGovernanceDraftSignal_RejectsRepositoryLineageMismatch(t *testing.T) {
	t.Parallel()

	srv := NewServer(Dependencies{
		ChangeGovernance: fakeChangeGovernanceService{},
		MCP: fakeChangeGovernanceMCPService{
			session: mcpdomain.SessionContext{
				RunID:     "run-1",
				ProjectID: "project-1",
			},
		},
		Runs: fakeChangeGovernanceRunReader{
			run: agentrunrepo.Run{
				ID:            "run-1",
				ProjectID:     "project-1",
				CorrelationID: "corr-1",
				RunPayload: mustMarshalRunPayload(t, querytypes.RunPayload{
					Repository: querytypes.RunPayloadRepository{FullName: "codex-k8s/kodex"},
					Issue:      &querytypes.RunPayloadIssue{Number: 521},
				}),
			},
			found: true,
		},
	})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer test-token"))
	_, err := srv.ReportChangeGovernanceDraftSignal(ctx, &controlplanev1.ReportChangeGovernanceDraftSignalRequest{
		RunId:              "run-1",
		RepositoryFullName: "other/repo",
		IssueNumber:        521,
		SignalId:           "signal-1",
		CorrelationId:      "corr-1",
		DraftRef:           "draft-1",
		DraftKind:          "internal_working_draft",
		ChangeScopeHints: []*controlplanev1.ChangeGovernanceScopeHint{
			{ContextKey: "services/internal/control-plane", SurfaceKind: "domain"},
		},
		OccurredAt: timestamppb.New(time.Now().UTC()),
	})
	if err == nil {
		t.Fatal("expected permission denied error, got nil")
	}
	if got, want := status.Code(err), codes.PermissionDenied; got != want {
		t.Fatalf("status code = %s, want %s", got, want)
	}
}

type fakeChangeGovernanceService struct{}

func (fakeChangeGovernanceService) ReportDraftSignal(context.Context, querytypes.ChangeGovernanceDraftSignalParams) (changegovernancedomain.DraftSignalResult, error) {
	return changegovernancedomain.DraftSignalResult{}, nil
}

func (fakeChangeGovernanceService) PublishWaveMap(context.Context, querytypes.ChangeGovernanceWaveMapParams) (changegovernancedomain.WaveMapResult, error) {
	return changegovernancedomain.WaveMapResult{}, nil
}

func (fakeChangeGovernanceService) UpsertEvidenceSignal(context.Context, querytypes.ChangeGovernanceEvidenceSignalParams) (changegovernancedomain.EvidenceSignalResult, error) {
	return changegovernancedomain.EvidenceSignalResult{}, nil
}

type fakeChangeGovernanceMCPService struct {
	session mcpdomain.SessionContext
}

func (f fakeChangeGovernanceMCPService) IssueRunToken(context.Context, mcpdomain.IssueRunTokenParams) (mcpdomain.IssuedToken, error) {
	return mcpdomain.IssuedToken{}, nil
}

func (f fakeChangeGovernanceMCPService) VerifyRunToken(context.Context, string) (mcpdomain.SessionContext, error) {
	return f.session, nil
}

func (f fakeChangeGovernanceMCPService) VerifyInteractionCallbackToken(context.Context, string, string) (mcpdomain.SessionContext, error) {
	return f.session, nil
}

func (fakeChangeGovernanceMCPService) ListPendingApprovals(context.Context, int) ([]mcpdomain.ApprovalListItem, error) {
	return nil, nil
}

func (fakeChangeGovernanceMCPService) ResolveApproval(context.Context, mcpdomain.ResolveApprovalParams) (mcpdomain.ResolveApprovalResult, error) {
	return mcpdomain.ResolveApprovalResult{}, nil
}

func (fakeChangeGovernanceMCPService) ClaimNextInteractionDispatch(context.Context, mcpdomain.ClaimNextInteractionDispatchParams) (mcpdomain.InteractionDispatchClaim, bool, error) {
	return mcpdomain.InteractionDispatchClaim{}, false, nil
}

func (fakeChangeGovernanceMCPService) CompleteInteractionDispatch(context.Context, mcpdomain.CompleteInteractionDispatchParams) (mcpdomain.CompleteInteractionDispatchResult, error) {
	return mcpdomain.CompleteInteractionDispatchResult{}, nil
}

func (fakeChangeGovernanceMCPService) ExpireNextDueInteraction(context.Context) (mcpdomain.ExpireNextInteractionResult, bool, error) {
	return mcpdomain.ExpireNextInteractionResult{}, false, nil
}

func (fakeChangeGovernanceMCPService) SubmitInteractionCallback(context.Context, mcpdomain.SubmitInteractionCallbackParams) (mcpdomain.SubmitInteractionCallbackResult, error) {
	return mcpdomain.SubmitInteractionCallbackResult{}, nil
}

type fakeChangeGovernanceRunReader struct {
	run   agentrunrepo.Run
	found bool
}

func (f fakeChangeGovernanceRunReader) GetByID(context.Context, string) (agentrunrepo.Run, bool, error) {
	return f.run, f.found, nil
}

func mustMarshalRunPayload(t *testing.T, payload querytypes.RunPayload) json.RawMessage {
	t.Helper()

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal run payload: %v", err)
	}
	return raw
}
