package githubratelimit

import (
	"context"
	"fmt"
	"sort"
	"strings"

	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// GetRunProjection returns canonical dominant/related wait read model for one run.
func (s *Service) GetRunProjection(ctx context.Context, runID string) (WaitProjection, bool, error) {
	if err := s.assertReadAllowed(); err != nil {
		return WaitProjection{}, false, err
	}

	trimmedRunID := strings.TrimSpace(runID)
	if trimmedRunID == "" {
		return WaitProjection{}, false, fmt.Errorf("run_id is required")
	}

	waits, err := s.waits.ListByRunID(ctx, trimmedRunID)
	if err != nil {
		return WaitProjection{}, false, fmt.Errorf("list github rate-limit waits by run id: %w", err)
	}
	return BuildRunProjectionFromWaits(waits)
}

// BuildRunProjectionFromWaits converts persisted waits into canonical dominant/related read model.
func BuildRunProjectionFromWaits(waits []Wait) (WaitProjection, bool, error) {
	openWaits := make([]Wait, 0, len(waits))
	for _, wait := range waits {
		if wait.State.IsOpen() {
			openWaits = append(openWaits, wait)
		}
	}
	if len(openWaits) == 0 {
		return WaitProjection{}, false, nil
	}

	sortOpenWaits(openWaits)
	dominant, found := selectDominantProjectionWait(openWaits)
	if !found {
		return WaitProjection{}, false, nil
	}

	dominantItem, err := buildProjectionItem(dominant)
	if err != nil {
		return WaitProjection{}, false, err
	}

	related := make([]WaitProjectionItem, 0, len(openWaits)-1)
	for _, wait := range openWaits {
		if wait.ID == dominant.ID {
			continue
		}
		item, err := buildProjectionItem(wait)
		if err != nil {
			return WaitProjection{}, false, err
		}
		related = append(related, item)
	}

	return WaitProjection{
		WaitState:          waitStateWaitingBackpressure,
		WaitReason:         enumtypes.AgentRunWaitReasonGitHubRateLimit,
		DominantWait:       dominantItem,
		RelatedWaits:       related,
		CommentMirrorState: deriveCommentMirrorState(openWaits),
	}, true, nil
}

// BuildCommentRenderContext derives best-effort GitHub service-comment data strictly from typed projection.
func (s *Service) BuildCommentRenderContext(projection WaitProjection) (CommentRenderContext, error) {
	return BuildCommentRenderContext(projection)
}

// BuildCommentRenderContext derives best-effort GitHub service-comment data strictly from typed projection.
func BuildCommentRenderContext(projection WaitProjection) (CommentRenderContext, error) {
	if strings.TrimSpace(projection.DominantWait.WaitID) == "" {
		return CommentRenderContext{}, fmt.Errorf("dominant_wait is required")
	}

	headlineTemplate := "comment_headline_auto_resume"
	nextStepKind := nextStepKindForWaitState(projection.DominantWait.State)
	manualActionSummary := ""
	if projection.DominantWait.ManualAction != nil {
		headlineTemplate = "comment_headline_manual_action"
		manualActionSummary = projection.DominantWait.ManualAction.Summary
	}

	headline, err := renderMessageTemplate(headlineTemplate, messageTemplateData{
		ContourKind:            string(projection.DominantWait.ContourKind),
		LimitKind:              string(projection.DominantWait.LimitKind),
		OperationClass:         string(projection.DominantWait.OperationClass),
		NextStepKind:           string(nextStepKind),
		ResumeNotBeforeRFC3339: formatTemplateTime(projection.DominantWait.ResumeNotBefore),
	})
	if err != nil {
		return CommentRenderContext{}, err
	}

	badges := make([]CommentContourBadge, 0, len(projection.RelatedWaits))
	for _, related := range projection.RelatedWaits {
		badges = append(badges, CommentContourBadge{
			ContourKind: related.ContourKind,
			LimitKind:   related.LimitKind,
			State:       related.State,
		})
	}

	return CommentRenderContext{
		Headline:             headline,
		DominantContour:      projection.DominantWait.ContourKind,
		LimitKind:            projection.DominantWait.LimitKind,
		OperationClass:       projection.DominantWait.OperationClass,
		NextStepKind:         nextStepKind,
		ResumeNotBefore:      projection.DominantWait.ResumeNotBefore,
		ManualActionSummary:  manualActionSummary,
		RelatedContourBadges: badges,
	}, nil
}

func buildProjectionItem(wait Wait) (WaitProjectionItem, error) {
	recoveryHint, err := buildRecoveryHint(wait)
	if err != nil {
		return WaitProjectionItem{}, err
	}

	var manualAction *ManualAction
	if wait.State == enumtypes.GitHubRateLimitWaitStateManualActionRequired {
		action, err := buildManualAction(wait)
		if err != nil {
			return WaitProjectionItem{}, err
		}
		manualAction = &action
	}

	return WaitProjectionItem{
		WaitID:          wait.ID,
		ContourKind:     wait.ContourKind,
		LimitKind:       wait.LimitKind,
		OperationClass:  wait.OperationClass,
		State:           wait.State,
		Confidence:      wait.Confidence,
		EnteredAt:       wait.FirstDetectedAt.UTC(),
		ResumeNotBefore: wait.ResumeNotBefore,
		AttemptsUsed:    wait.AutoResumeAttemptsUsed,
		MaxAttempts:     wait.MaxAutoResumeAttempts,
		RecoveryHint:    recoveryHint,
		ManualAction:    manualAction,
	}, nil
}

func buildRecoveryHint(wait Wait) (RecoveryHint, error) {
	source := recoveryHintSourceForWait(wait)
	details, err := renderMessageTemplate(templateNameForRecoveryHint(wait.RecoveryHintKind), messageTemplateDataFromWait(wait))
	if err != nil {
		return RecoveryHint{}, err
	}

	return RecoveryHint{
		HintKind:        wait.RecoveryHintKind,
		ResumeNotBefore: wait.ResumeNotBefore,
		SourceHeaders:   source,
		DetailsMarkdown: details,
	}, nil
}

func buildManualAction(wait Wait) (ManualAction, error) {
	data := messageTemplateDataFromWait(wait)
	data.SuggestedNotBeforeRFC3339 = formatTemplateTime(wait.ResumeNotBefore)

	summary, err := renderMessageTemplate(templateNameForManualAction(wait.ManualActionKind, true), data)
	if err != nil {
		return ManualAction{}, err
	}
	details, err := renderMessageTemplate(templateNameForManualAction(wait.ManualActionKind, false), data)
	if err != nil {
		return ManualAction{}, err
	}

	return ManualAction{
		Kind:               wait.ManualActionKind,
		Summary:            summary,
		DetailsMarkdown:    details,
		SuggestedNotBefore: wait.ResumeNotBefore,
	}, nil
}

func messageTemplateDataFromWait(wait Wait) messageTemplateData {
	return messageTemplateData{
		ContourKind:            string(wait.ContourKind),
		LimitKind:              string(wait.LimitKind),
		OperationClass:         string(wait.OperationClass),
		WaitState:              string(wait.State),
		Confidence:             string(wait.Confidence),
		RecoveryHintKind:       string(wait.RecoveryHintKind),
		ResumeNotBeforeRFC3339: formatTemplateTime(wait.ResumeNotBefore),
		AttemptsUsed:           wait.AutoResumeAttemptsUsed,
		MaxAttempts:            wait.MaxAutoResumeAttempts,
	}
}

func templateNameForRecoveryHint(kind enumtypes.GitHubRateLimitRecoveryHintKind) string {
	return "recovery_hint_" + string(kind)
}

func templateNameForManualAction(kind enumtypes.GitHubRateLimitManualActionKind, summary bool) string {
	suffix := "details"
	if summary {
		suffix = "summary"
	}
	return "manual_action_" + string(kind) + "_" + suffix
}

func recoveryHintSourceForWait(wait Wait) enumtypes.GitHubRateLimitRecoveryHintSource {
	switch wait.RecoveryHintKind {
	case enumtypes.GitHubRateLimitRecoveryHintKindReset:
		return enumtypes.GitHubRateLimitRecoveryHintSourceResetAt
	case enumtypes.GitHubRateLimitRecoveryHintKindRetryAfter:
		return enumtypes.GitHubRateLimitRecoveryHintSourceRetryAfter
	case enumtypes.GitHubRateLimitRecoveryHintKindManualOnly:
		if wait.LimitKind == enumtypes.GitHubRateLimitLimitKindPrimary {
			return enumtypes.GitHubRateLimitRecoveryHintSourceResetAt
		}
		if wait.Confidence == enumtypes.GitHubRateLimitConfidenceConservative {
			return enumtypes.GitHubRateLimitRecoveryHintSourceRetryAfter
		}
		fallthrough
	default:
		return enumtypes.GitHubRateLimitRecoveryHintSourceProviderUncertain
	}
}

func deriveCommentMirrorState(openWaits []Wait) enumtypes.GitHubRateLimitCommentMirrorState {
	if len(openWaits) == 0 {
		return enumtypes.GitHubRateLimitCommentMirrorStateNotAttempted
	}
	for _, wait := range openWaits {
		if wait.ResumeActionKind == enumtypes.GitHubRateLimitResumeActionKindRunStatusCommentRetry ||
			wait.OperationClass == enumtypes.GitHubRateLimitOperationClassRunStatusComment {
			return enumtypes.GitHubRateLimitCommentMirrorStatePendingRetry
		}
	}
	return enumtypes.GitHubRateLimitCommentMirrorStateNotAttempted
}

func selectDominantProjectionWait(openWaits []Wait) (Wait, bool) {
	flagged := make([]Wait, 0, len(openWaits))
	for _, wait := range openWaits {
		if wait.DominantForRun {
			flagged = append(flagged, wait)
		}
	}
	if len(flagged) == 1 {
		return flagged[0], true
	}
	if len(flagged) > 1 {
		sortOpenWaits(flagged)
		return flagged[0], true
	}
	return ElectDominantWait(openWaits)
}

func sortOpenWaits(openWaits []Wait) {
	sort.SliceStable(openWaits, func(i int, j int) bool {
		return compareDominantWait(openWaits[i], openWaits[j]) < 0
	})
}
