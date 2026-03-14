package grpc

import (
	"context"
	"encoding/json"
	"strings"

	sharedgithubratelimit "github.com/codex-k8s/codex-k8s/libs/go/domain/githubratelimit"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	githubratelimitdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/githubratelimit"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ReportGitHubRateLimitSignal(
	ctx context.Context,
	req *controlplanev1.ReportGitHubRateLimitSignalRequest,
) (*controlplanev1.ReportGitHubRateLimitSignalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if s.githubRateLimit == nil {
		return nil, status.Error(codes.FailedPrecondition, "github rate-limit service is not configured")
	}

	runSession, err := s.authenticateRunToken(ctx)
	if err != nil {
		return nil, err
	}

	runID := strings.TrimSpace(req.GetRunId())
	if runID == "" {
		return nil, status.Error(codes.InvalidArgument, "run_id is required")
	}
	if runID != strings.TrimSpace(runSession.RunID) {
		return nil, status.Error(codes.PermissionDenied, "run_id does not match authenticated run token")
	}
	if strings.TrimSpace(req.GetSignalId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "signal_id is required")
	}
	if strings.TrimSpace(req.GetCorrelationId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "correlation_id is required")
	}
	if req.GetProviderStatusCode() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "provider_status_code must be > 0")
	}
	if req.GetOccurredAt() == nil || !req.GetOccurredAt().IsValid() {
		return nil, status.Error(codes.InvalidArgument, "occurred_at is required")
	}

	result, err := s.githubRateLimit.ReportSignal(ctx, githubratelimitdomain.ReportSignalParams{
		RunID: runID,
		Signal: githubratelimitdomain.Signal{
			SignalID:               strings.TrimSpace(req.GetSignalId()),
			CorrelationID:          strings.TrimSpace(req.GetCorrelationId()),
			ContourKind:            enumtypes.GitHubRateLimitContourKind(strings.TrimSpace(req.GetContourKind())),
			SignalOrigin:           enumtypes.GitHubRateLimitSignalOrigin(strings.TrimSpace(req.GetSignalOrigin())),
			OperationClass:         enumtypes.GitHubRateLimitOperationClass(strings.TrimSpace(req.GetOperationClass())),
			ProviderStatusCode:     int(req.GetProviderStatusCode()),
			OccurredAt:             req.GetOccurredAt().AsTime().UTC(),
			RequestFingerprint:     strings.TrimSpace(req.GetRequestFingerprint()),
			StderrExcerpt:          strings.TrimSpace(req.GetStderrExcerpt()),
			MessageExcerpt:         strings.TrimSpace(req.GetMessageExcerpt()),
			Headers:                githubHeadersFromProto(req.GetGithubHeaders()),
			SessionSnapshotVersion: int64PtrFromProto(req.SessionSnapshotVersion),
		},
	})
	if err != nil {
		return nil, toStatus(err)
	}
	if result.HardFailure {
		failureReason := strings.TrimSpace(result.Classification.FailureReason)
		if failureReason == "" {
			failureReason = "signal is not recoverable as github rate-limit wait"
		}
		return nil, status.Error(codes.FailedPrecondition, failureReason)
	}

	resp := &controlplanev1.ReportGitHubRateLimitSignalResponse{
		WaitId:       strings.TrimSpace(result.Wait.ID),
		WaitState:    strings.TrimSpace(string(result.Wait.State)),
		WaitReason:   strings.TrimSpace(string(enumtypes.AgentRunWaitReasonGitHubRateLimit)),
		NextStepKind: strings.TrimSpace(string(result.Classification.NextStepKind)),
		RunnerAction: sharedgithubratelimit.RunnerActionPersistSessionAndExitWait,
	}
	if result.Wait.ResumeNotBefore != nil && !result.Wait.ResumeNotBefore.IsZero() {
		resp.ResumeNotBefore = timestamppb.New(result.Wait.ResumeNotBefore.UTC())
	}
	return resp, nil
}

func (s *Server) GetRunGitHubRateLimitResumePayload(
	ctx context.Context,
	req *controlplanev1.GetRunGitHubRateLimitResumePayloadRequest,
) (*controlplanev1.GetRunGitHubRateLimitResumePayloadResponse, error) {
	return executeRunScopedPayloadSpec(ctx, req, s.loadRunScopedPayload, githubRateLimitResumePayloadSpec(s))
}

func githubRateLimitResumePayloadSpec(s *Server) runScopedPayloadSpec[*controlplanev1.GetRunGitHubRateLimitResumePayloadResponse] {
	return runScopedPayloadSpec[*controlplanev1.GetRunGitHubRateLimitResumePayloadResponse]{
		label: "github rate-limit resume payload",
		load:  s.agentCallbacks.GetRunGitHubRateLimitResumePayload,
		build: githubRateLimitResumePayloadResponseBuilder,
	}
}

var githubRateLimitResumePayloadResponseBuilder = func(found bool, payload json.RawMessage) *controlplanev1.GetRunGitHubRateLimitResumePayloadResponse {
	return &controlplanev1.GetRunGitHubRateLimitResumePayloadResponse{Found: found, PayloadJson: payload}
}

func githubHeadersFromProto(item *controlplanev1.GitHubRateLimitHeaders) githubratelimitdomain.Headers {
	if item == nil {
		return githubratelimitdomain.Headers{}
	}

	headers := githubratelimitdomain.Headers{
		RateLimitLimit:     intPtrFromOptionalInt32(item.RateLimitLimit),
		RateLimitRemaining: intPtrFromOptionalInt32(item.RateLimitRemaining),
		RateLimitUsed:      intPtrFromOptionalInt32(item.RateLimitUsed),
		RateLimitResource:  strings.TrimSpace(item.GetRateLimitResource()),
		RetryAfterSeconds:  intPtrFromOptionalInt32(item.RetryAfterSeconds),
		GitHubRequestID:    strings.TrimSpace(item.GetGithubRequestId()),
		DocumentationURL:   strings.TrimSpace(item.GetDocumentationUrl()),
	}
	if item.GetRateLimitResetAt() != nil && item.GetRateLimitResetAt().IsValid() {
		resetAt := item.GetRateLimitResetAt().AsTime().UTC()
		headers.RateLimitResetAt = &resetAt
	}
	return headers
}

func intPtrFromOptionalInt32(value *int32) *int {
	if value == nil {
		return nil
	}
	converted := int(*value)
	return &converted
}

func int64PtrFromProto(value *int64) *int64 {
	if value == nil {
		return nil
	}
	converted := *value
	return &converted
}
