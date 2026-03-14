package grpc

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	githubratelimitdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/githubratelimit"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func TestRunToProtoWithWaitProjection_FallsBackToBaseRunOnProjectionError(t *testing.T) {
	t.Parallel()

	recorder := &captureRuntimeErrorRecorder{}
	server := &Server{
		githubRateLimit: stubRunWaitProjectionGitHubRateLimitService{
			err: errors.New("projection store unavailable"),
		},
		runtimeErrors: recorder,
		logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	run := testStaffRunForWaitProjection()
	got, err := server.runToProtoWithWaitProjection(context.Background(), run)
	if err != nil {
		t.Fatalf("runToProtoWithWaitProjection() error = %v", err)
	}
	if got.GetId() != run.ID {
		t.Fatalf("run id = %q, want %q", got.GetId(), run.ID)
	}
	if got.WaitProjection != nil {
		t.Fatal("wait_projection must stay nil when projection lookup fails")
	}
	if len(recorder.calls) != 1 {
		t.Fatalf("runtime error records = %d, want 1", len(recorder.calls))
	}
	if got, want := recorder.calls[0].Source, runWaitProjectionSource; got != want {
		t.Fatalf("runtime error source = %q, want %q", got, want)
	}
	if got, want := recorder.calls[0].RunID, run.ID; got != want {
		t.Fatalf("runtime error run_id = %q, want %q", got, want)
	}
}

func TestRunToProtoWithWaitProjection_AttachesProjectionWhenAvailable(t *testing.T) {
	t.Parallel()

	server := &Server{
		githubRateLimit: stubRunWaitProjectionGitHubRateLimitService{
			found: true,
			projection: githubratelimitdomain.WaitProjection{
				WaitState:  runWaitStateBackpressure,
				WaitReason: enumtypes.AgentRunWaitReasonGitHubRateLimit,
				DominantWait: githubratelimitdomain.WaitProjectionItem{
					WaitID:         "wait-1",
					ContourKind:    enumtypes.GitHubRateLimitContourKindPlatformPAT,
					LimitKind:      enumtypes.GitHubRateLimitLimitKindSecondary,
					OperationClass: enumtypes.GitHubRateLimitOperationClassAgentGitHubCall,
					State:          enumtypes.GitHubRateLimitWaitStateOpen,
					Confidence:     enumtypes.GitHubRateLimitConfidenceDeterministic,
					EnteredAt:      time.Date(2026, 3, 14, 19, 0, 0, 0, time.UTC),
					AttemptsUsed:   1,
					MaxAttempts:    3,
					RecoveryHint: githubratelimitdomain.RecoveryHint{
						HintKind:        enumtypes.GitHubRateLimitRecoveryHintKindRetryAfter,
						SourceHeaders:   enumtypes.GitHubRateLimitRecoveryHintSourceRetryAfter,
						DetailsMarkdown: "Retry after the provider window expires.",
					},
				},
				RelatedWaits:       []githubratelimitdomain.WaitProjectionItem{},
				CommentMirrorState: enumtypes.GitHubRateLimitCommentMirrorStatePendingRetry,
			},
		},
	}

	got, err := server.runToProtoWithWaitProjection(context.Background(), testStaffRunForWaitProjection())
	if err != nil {
		t.Fatalf("runToProtoWithWaitProjection() error = %v", err)
	}
	if got.WaitProjection == nil {
		t.Fatal("wait_projection is nil")
	}
	if got, want := got.WaitProjection.GetDominantWait().GetWaitId(), "wait-1"; got != want {
		t.Fatalf("dominant wait id = %q, want %q", got, want)
	}
	if got, want := got.WaitProjection.GetCommentMirrorState(), "pending_retry"; got != want {
		t.Fatalf("comment mirror state = %q, want %q", got, want)
	}
}

func testStaffRunForWaitProjection() entitytypes.StaffRun {
	return entitytypes.StaffRun{
		ID:            "run-1",
		CorrelationID: "corr-1",
		ProjectID:     "proj-1",
		JobName:       "job-1",
		Namespace:     "ns-1",
		WaitState:     runWaitStateBackpressure,
		WaitReason:    runWaitReasonGitHubLimit,
		Status:        runWaitStateBackpressure,
		CreatedAt:     time.Date(2026, 3, 14, 18, 0, 0, 0, time.UTC),
	}
}

type stubRunWaitProjectionGitHubRateLimitService struct {
	projection githubratelimitdomain.WaitProjection
	found      bool
	err        error
}

func (s stubRunWaitProjectionGitHubRateLimitService) ReportSignal(context.Context, githubratelimitdomain.ReportSignalParams) (githubratelimitdomain.ReportSignalResult, error) {
	return githubratelimitdomain.ReportSignalResult{}, nil
}

func (s stubRunWaitProjectionGitHubRateLimitService) ProcessNextAutoResume(context.Context, githubratelimitdomain.ProcessNextAutoResumeParams) (githubratelimitdomain.ProcessNextAutoResumeResult, error) {
	return githubratelimitdomain.ProcessNextAutoResumeResult{}, nil
}

func (s stubRunWaitProjectionGitHubRateLimitService) GetRunProjection(context.Context, string) (githubratelimitdomain.WaitProjection, bool, error) {
	return s.projection, s.found, s.err
}

type captureRuntimeErrorRecorder struct {
	calls []querytypes.RuntimeErrorRecordParams
}

func (r *captureRuntimeErrorRecorder) RecordBestEffort(_ context.Context, params querytypes.RuntimeErrorRecordParams) {
	r.calls = append(r.calls, params)
}
