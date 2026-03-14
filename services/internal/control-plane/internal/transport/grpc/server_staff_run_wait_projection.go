package grpc

import (
	"context"
	"fmt"
	"strings"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	githubratelimitdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/githubratelimit"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	runWaitStateBackpressure = "waiting_backpressure"
	runWaitReasonGitHubLimit = "github_rate_limit"
)

func (s *Server) runToProtoWithWaitProjection(ctx context.Context, run entitytypes.StaffRun) (*controlplanev1.Run, error) {
	out := runToProto(run)
	if s.githubRateLimit == nil || !shouldAttachGitHubRateLimitProjection(run) {
		return out, nil
	}

	projection, found, err := s.githubRateLimit.GetRunProjection(ctx, strings.TrimSpace(run.ID))
	if err != nil {
		return nil, fmt.Errorf("load github rate-limit wait projection: %w", err)
	}
	if !found {
		return out, nil
	}

	out.WaitProjection = waitProjectionToProto(projection)
	return out, nil
}

func shouldAttachGitHubRateLimitProjection(run entitytypes.StaffRun) bool {
	waitState := strings.TrimSpace(run.WaitState)
	waitReason := strings.TrimSpace(run.WaitReason)
	status := strings.TrimSpace(run.Status)
	if strings.EqualFold(waitReason, runWaitReasonGitHubLimit) {
		return true
	}
	if strings.EqualFold(waitState, runWaitStateBackpressure) {
		return true
	}
	return strings.EqualFold(status, runWaitStateBackpressure)
}

func waitProjectionToProto(item githubratelimitdomain.WaitProjection) *controlplanev1.RunWaitProjection {
	if strings.TrimSpace(item.WaitState) == "" {
		return nil
	}

	out := &controlplanev1.RunWaitProjection{
		WaitState:          strings.TrimSpace(item.WaitState),
		WaitReason:         strings.TrimSpace(string(item.WaitReason)),
		DominantWait:       waitProjectionItemToProto(item.DominantWait),
		RelatedWaits:       make([]*controlplanev1.GitHubRateLimitWaitItem, 0, len(item.RelatedWaits)),
		CommentMirrorState: strings.TrimSpace(string(item.CommentMirrorState)),
	}
	for _, related := range item.RelatedWaits {
		out.RelatedWaits = append(out.RelatedWaits, waitProjectionItemToProto(related))
	}
	return out
}

func waitProjectionItemToProto(item githubratelimitdomain.WaitProjectionItem) *controlplanev1.GitHubRateLimitWaitItem {
	out := &controlplanev1.GitHubRateLimitWaitItem{
		WaitId:         strings.TrimSpace(item.WaitID),
		ContourKind:    strings.TrimSpace(string(item.ContourKind)),
		LimitKind:      strings.TrimSpace(string(item.LimitKind)),
		OperationClass: strings.TrimSpace(string(item.OperationClass)),
		State:          strings.TrimSpace(string(item.State)),
		Confidence:     strings.TrimSpace(string(item.Confidence)),
		EnteredAt:      timestamppb.New(item.EnteredAt.UTC()),
		AttemptsUsed:   int32(item.AttemptsUsed),
		MaxAttempts:    int32(item.MaxAttempts),
		RecoveryHint:   recoveryHintToProto(item.RecoveryHint),
	}
	if item.ResumeNotBefore != nil {
		out.ResumeNotBefore = timestamppb.New(item.ResumeNotBefore.UTC())
	}
	if item.ManualAction != nil {
		out.ManualAction = manualActionToProto(*item.ManualAction)
	}
	return out
}

func recoveryHintToProto(item githubratelimitdomain.RecoveryHint) *controlplanev1.GitHubRateLimitRecoveryHint {
	out := &controlplanev1.GitHubRateLimitRecoveryHint{
		HintKind:        strings.TrimSpace(string(item.HintKind)),
		SourceHeaders:   strings.TrimSpace(string(item.SourceHeaders)),
		DetailsMarkdown: item.DetailsMarkdown,
	}
	if item.ResumeNotBefore != nil {
		out.ResumeNotBefore = timestamppb.New(item.ResumeNotBefore.UTC())
	}
	return out
}

func manualActionToProto(item githubratelimitdomain.ManualAction) *controlplanev1.GitHubRateLimitManualAction {
	out := &controlplanev1.GitHubRateLimitManualAction{
		Kind:            strings.TrimSpace(string(item.Kind)),
		Summary:         item.Summary,
		DetailsMarkdown: item.DetailsMarkdown,
	}
	if item.SuggestedNotBefore != nil {
		out.SuggestedNotBefore = timestamppb.New(item.SuggestedNotBefore.UTC())
	}
	return out
}
