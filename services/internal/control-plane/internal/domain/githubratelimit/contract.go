package githubratelimit

import (
	"context"

	waitrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/githubratelimitwait"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

type Signal = valuetypes.GitHubRateLimitSignal

type Headers = valuetypes.GitHubRateLimitHeaders

type Classification = valuetypes.GitHubRateLimitClassification

type WaitProjection = valuetypes.GitHubRateLimitRunWaitProjection

type WaitProjectionItem = valuetypes.GitHubRateLimitWaitProjectionItem

type RecoveryHint = valuetypes.GitHubRateLimitRecoveryHint

type ManualAction = valuetypes.GitHubRateLimitManualAction

type CommentRenderContext = valuetypes.GitHubRateLimitCommentRenderContext

type CommentContourBadge = valuetypes.GitHubRateLimitCommentContourBadge

type ResumePayloadBuildResult = valuetypes.GitHubRateLimitResumePayloadBuildResult

type Wait = waitrepo.Wait

// DomainService exposes canonical GitHub rate-limit semantics owned by control-plane.
type DomainService interface {
	ReportSignal(ctx context.Context, params ReportSignalParams) (ReportSignalResult, error)
	GetRunProjection(ctx context.Context, runID string) (WaitProjection, bool, error)
	BuildCommentRenderContext(projection WaitProjection) (CommentRenderContext, error)
	BuildAgentSessionResumePayload(params BuildResumePayloadParams) (ResumePayloadBuildResult, error)
	ProcessNextAutoResume(ctx context.Context, params ProcessNextAutoResumeParams) (ProcessNextAutoResumeResult, error)
}
