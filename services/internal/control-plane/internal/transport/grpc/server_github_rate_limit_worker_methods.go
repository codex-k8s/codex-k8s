package grpc

import (
	"context"
	"strings"

	controlplanev1 "github.com/codex-k8s/kodex/proto/gen/go/kodex/controlplane/v1"
	githubratelimitdomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/githubratelimit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ProcessNextGitHubRateLimitWait(
	ctx context.Context,
	req *controlplanev1.ProcessNextGitHubRateLimitWaitRequest,
) (*controlplanev1.ProcessNextGitHubRateLimitWaitResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if s.githubRateLimit == nil {
		return nil, status.Error(codes.FailedPrecondition, "github rate-limit worker service is not configured")
	}

	result, err := s.githubRateLimit.ProcessNextAutoResume(ctx, githubratelimitdomain.ProcessNextAutoResumeParams{
		WorkerID: strings.TrimSpace(req.GetWorkerId()),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	resp := &controlplanev1.ProcessNextGitHubRateLimitWaitResponse{
		Found:                 result.Found,
		ResolutionKind:        strings.TrimSpace(string(result.ResolutionKind)),
		AttemptNo:             int32(result.AttemptNo),
		ManualActionKind:      strings.TrimSpace(string(result.ManualActionKind)),
		RequeuedCorrelationId: strings.TrimSpace(result.RequeuedCorrelationID),
	}
	if !result.Found {
		return resp, nil
	}

	resp.WaitId = strings.TrimSpace(result.Wait.ID)
	resp.RunId = strings.TrimSpace(result.Wait.RunID)
	resp.State = strings.TrimSpace(string(result.Wait.State))
	if result.ResumeNotBefore != nil && !result.ResumeNotBefore.IsZero() {
		resp.ResumeNotBefore = timestamppb.New(result.ResumeNotBefore.UTC())
	}
	return resp, nil
}
