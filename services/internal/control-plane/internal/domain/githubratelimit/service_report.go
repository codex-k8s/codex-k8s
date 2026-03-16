package githubratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	waitrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/githubratelimitwait"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

// ReportSignal normalizes provider evidence, upserts wait aggregate, and returns canonical projection.
func (s *Service) ReportSignal(ctx context.Context, params ReportSignalParams) (ReportSignalResult, error) {
	if s == nil {
		return ReportSignalResult{}, fmt.Errorf("github rate-limit service is not configured")
	}

	trimmedRunID := strings.TrimSpace(params.RunID)
	if trimmedRunID == "" {
		return ReportSignalResult{}, fmt.Errorf("run_id is required")
	}

	signal, err := normalizeSignal(params.Signal, s.now())
	if err != nil {
		return ReportSignalResult{}, err
	}
	if err := s.assertSignalWriteAllowed(signal.SignalOrigin); err != nil {
		return ReportSignalResult{}, err
	}

	run, found, err := s.runs.GetByID(ctx, trimmedRunID)
	if err != nil {
		return ReportSignalResult{}, fmt.Errorf("load run for github rate-limit signal: %w", err)
	}
	if !found {
		return ReportSignalResult{}, fmt.Errorf("run %s not found", trimmedRunID)
	}
	if isTerminalRunStatus(run.Status) {
		return ReportSignalResult{}, fmt.Errorf("run %s is already terminal", trimmedRunID)
	}

	if existingBySignal, found, err := s.waits.GetBySignalID(ctx, signal.SignalID); err != nil {
		return ReportSignalResult{}, fmt.Errorf("lookup github rate-limit wait by signal id: %w", err)
	} else if found {
		if err := validateDuplicateSignalContext(trimmedRunID, signal, existingBySignal); err != nil {
			return ReportSignalResult{}, err
		}
		return s.buildExistingSignalResult(ctx, trimmedRunID, existingBySignal)
	}

	existingWait, found, err := s.waits.GetOpenByRunAndContour(ctx, trimmedRunID, signal.ContourKind)
	if err != nil {
		return ReportSignalResult{}, fmt.Errorf("lookup github rate-limit wait by run+contour: %w", err)
	}
	if !found {
		existingWait = Wait{}
	}

	classification, err := classifySignal(signal, existingWait, s.now())
	if err != nil {
		return ReportSignalResult{}, err
	}
	if classification.HardFailure {
		return ReportSignalResult{
			HardFailure:    true,
			Classification: classification,
		}, nil
	}

	correlationID := signal.CorrelationID
	if correlationID == "" {
		correlationID = strings.TrimSpace(run.CorrelationID)
	}
	if correlationID == "" {
		return ReportSignalResult{}, fmt.Errorf("correlation_id is required")
	}

	wait, err := s.upsertWait(ctx, run, signal, classification, existingWait, correlationID, params.ReplayPayloadJSON)
	if err != nil {
		return ReportSignalResult{}, err
	}

	if _, err := s.waits.AppendEvidence(ctx, querytypes.GitHubRateLimitWaitEvidenceCreateParams{
		WaitID:             wait.ID,
		EventKind:          enumtypes.GitHubRateLimitEvidenceEventSignalDetected,
		SignalID:           signal.SignalID,
		SignalOrigin:       signal.SignalOrigin,
		ProviderStatusCode: ptrInt(signal.ProviderStatusCode),
		RetryAfterSeconds:  signal.Headers.RetryAfterSeconds,
		RateLimitLimit:     signal.Headers.RateLimitLimit,
		RateLimitRemaining: signal.Headers.RateLimitRemaining,
		RateLimitUsed:      signal.Headers.RateLimitUsed,
		RateLimitResetAt:   signal.Headers.RateLimitResetAt,
		RateLimitResource:  signal.Headers.RateLimitResource,
		GitHubRequestID:    signal.Headers.GitHubRequestID,
		DocumentationURL:   signal.Headers.DocumentationURL,
		MessageExcerpt:     signal.MessageExcerpt,
		StderrExcerpt:      signal.StderrExcerpt,
		PayloadJSON:        marshalJSONPayload(waitSignalEvidencePayload{ContourKind: signal.ContourKind, OperationClass: signal.OperationClass, SessionSnapshotVersion: signal.SessionSnapshotVersion}),
		ObservedAt:         signal.OccurredAt,
	}); err != nil {
		return ReportSignalResult{}, fmt.Errorf("append github rate-limit signal evidence: %w", err)
	}

	refreshResult, err := s.waits.RefreshRunProjection(ctx, trimmedRunID)
	if err != nil {
		return ReportSignalResult{}, fmt.Errorf("refresh github rate-limit run projection: %w", err)
	}

	if _, err := s.waits.AppendEvidence(ctx, querytypes.GitHubRateLimitWaitEvidenceCreateParams{
		WaitID:       wait.ID,
		EventKind:    enumtypes.GitHubRateLimitEvidenceEventClassified,
		SignalID:     signal.SignalID,
		SignalOrigin: signal.SignalOrigin,
		PayloadJSON: marshalJSONPayload(waitClassificationEvidencePayload{
			LimitKind:           classification.LimitKind,
			Confidence:          classification.Confidence,
			RecoveryHintKind:    classification.RecoveryHintKind,
			RecoveryHintSource:  classification.RecoveryHintSource,
			NextStepKind:        classification.NextStepKind,
			ResumeActionKind:    classification.ResumeActionKind,
			ManualActionKind:    classification.ManualActionKind,
			ProjectionSyncState: refreshResult.SyncState,
		}),
		ObservedAt: signal.OccurredAt,
	}); err != nil {
		return ReportSignalResult{}, fmt.Errorf("append github rate-limit classification evidence: %w", err)
	}
	if err := s.appendPostClassificationEvidence(ctx, wait, signal, classification); err != nil {
		return ReportSignalResult{}, err
	}

	projection, commentRenderContext, err := s.loadProjectionAndCommentContext(ctx, trimmedRunID)
	if err != nil {
		return ReportSignalResult{}, err
	}
	s.insertRunWaitFlowEvent(ctx, correlationID, wait, classification, refreshResult, projection)

	return ReportSignalResult{
		Wait:                 wait,
		Classification:       classification,
		Projection:           projection,
		CommentRenderContext: commentRenderContext,
		ProjectionRefresh:    refreshResult,
	}, nil
}

func (s *Service) buildExistingSignalResult(ctx context.Context, runID string, wait Wait) (ReportSignalResult, error) {
	projection, commentRenderContext, err := s.loadProjectionAndCommentContext(ctx, runID)
	if err != nil {
		return ReportSignalResult{}, err
	}

	return ReportSignalResult{
		DuplicateSignal:      true,
		Wait:                 wait,
		Classification:       classificationFromWait(wait),
		Projection:           projection,
		CommentRenderContext: commentRenderContext,
	}, nil
}

func validateDuplicateSignalContext(runID string, signal Signal, wait Wait) error {
	if strings.TrimSpace(wait.RunID) != runID {
		return errs.Conflict{
			Msg: fmt.Sprintf(
				"stale github rate-limit duplicate signal %q belongs to run %q instead of %q",
				signal.SignalID,
				strings.TrimSpace(wait.RunID),
				runID,
			),
		}
	}
	if wait.ContourKind != signal.ContourKind {
		return errs.Conflict{
			Msg: fmt.Sprintf(
				"stale github rate-limit duplicate signal %q belongs to contour %q instead of %q",
				signal.SignalID,
				wait.ContourKind,
				signal.ContourKind,
			),
		}
	}
	if wait.OperationClass != signal.OperationClass {
		return errs.Conflict{
			Msg: fmt.Sprintf(
				"stale github rate-limit duplicate signal %q belongs to operation_class %q instead of %q",
				signal.SignalID,
				wait.OperationClass,
				signal.OperationClass,
			),
		}
	}
	return nil
}

func (s *Service) upsertWait(ctx context.Context, run agentrunrepo.Run, signal Signal, classification Classification, existing Wait, correlationID string, replayPayload json.RawMessage) (Wait, error) {
	resumePayload, err := resolveWaitResumePayload(classification.ResumeActionKind, existing.ResumePayloadJSON, replayPayload)
	if err != nil {
		return Wait{}, err
	}

	if strings.TrimSpace(existing.ID) != "" {
		wait, found, err := s.waits.Update(ctx, querytypes.GitHubRateLimitWaitUpdateParams{
			ID:                     existing.ID,
			SignalOrigin:           signal.SignalOrigin,
			OperationClass:         signal.OperationClass,
			State:                  classification.State,
			LimitKind:              classification.LimitKind,
			Confidence:             classification.Confidence,
			RecoveryHintKind:       classification.RecoveryHintKind,
			SignalID:               signal.SignalID,
			RequestFingerprint:     signal.RequestFingerprint,
			CorrelationID:          correlationID,
			ResumeActionKind:       classification.ResumeActionKind,
			ResumePayloadJSON:      resumePayload,
			ManualActionKind:       classification.ManualActionKind,
			AutoResumeAttemptsUsed: existing.AutoResumeAttemptsUsed,
			MaxAutoResumeAttempts:  classification.MaxAutoResumeAttempts,
			ResumeNotBefore:        classification.ResumeNotBefore,
			LastSignalAt:           signal.OccurredAt,
		})
		if err != nil {
			return Wait{}, fmt.Errorf("update github rate-limit wait: %w", err)
		}
		if !found {
			return Wait{}, fmt.Errorf("wait %s not found for update", existing.ID)
		}
		return wait, nil
	}

	wait, err := s.waits.Create(ctx, querytypes.GitHubRateLimitWaitCreateParams{
		ProjectID:              strings.TrimSpace(run.ProjectID),
		RunID:                  strings.TrimSpace(run.ID),
		ContourKind:            signal.ContourKind,
		SignalOrigin:           signal.SignalOrigin,
		OperationClass:         signal.OperationClass,
		State:                  classification.State,
		LimitKind:              classification.LimitKind,
		Confidence:             classification.Confidence,
		RecoveryHintKind:       classification.RecoveryHintKind,
		SignalID:               signal.SignalID,
		RequestFingerprint:     signal.RequestFingerprint,
		CorrelationID:          correlationID,
		ResumeActionKind:       classification.ResumeActionKind,
		ResumePayloadJSON:      resumePayload,
		ManualActionKind:       classification.ManualActionKind,
		AutoResumeAttemptsUsed: 0,
		MaxAutoResumeAttempts:  classification.MaxAutoResumeAttempts,
		ResumeNotBefore:        classification.ResumeNotBefore,
		FirstDetectedAt:        signal.OccurredAt,
		LastSignalAt:           signal.OccurredAt,
	})
	if err != nil {
		return Wait{}, fmt.Errorf("create github rate-limit wait: %w", err)
	}
	return wait, nil
}

func (s *Service) appendPostClassificationEvidence(ctx context.Context, wait Wait, signal Signal, classification Classification) error {
	observedAt := signal.OccurredAt
	if classification.NextStepKind == enumtypes.GitHubRateLimitNextStepKindAutoResumeScheduled {
		return s.appendResumeScheduledEvidence(ctx, wait, signal.SignalID, signal.SignalOrigin, observedAt, classification.NextStepKind)
	}
	if classification.NextStepKind == enumtypes.GitHubRateLimitNextStepKindManualActionRequired {
		if _, err := s.waits.AppendEvidence(ctx, querytypes.GitHubRateLimitWaitEvidenceCreateParams{
			WaitID:       wait.ID,
			EventKind:    enumtypes.GitHubRateLimitEvidenceEventManualActionRequired,
			SignalID:     signal.SignalID,
			SignalOrigin: signal.SignalOrigin,
			PayloadJSON: marshalJSONPayload(waitManualActionEvidencePayload{
				WaitID:           wait.ID,
				AttemptNo:        wait.AutoResumeAttemptsUsed,
				ManualActionKind: wait.ManualActionKind,
			}),
			ObservedAt: observedAt,
		}); err != nil {
			return fmt.Errorf("append github rate-limit manual-action evidence: %w", err)
		}
	}
	return nil
}

func resolveWaitResumePayload(resumeActionKind enumtypes.GitHubRateLimitResumeActionKind, existing json.RawMessage, incoming json.RawMessage) (json.RawMessage, error) {
	if len(incoming) == 0 {
		if len(existing) == 0 {
			switch resumeActionKind {
			case enumtypes.GitHubRateLimitResumeActionKindRunStatusCommentRetry,
				enumtypes.GitHubRateLimitResumeActionKindPlatformCallReplay:
				return nil, errGitHubRateLimitReplayPayloadMissing
			default:
				return json.RawMessage(`{}`), nil
			}
		}
		return existing, nil
	}
	if !json.Valid(incoming) {
		return nil, fmt.Errorf("github rate-limit replay payload must be valid JSON")
	}
	switch resumeActionKind {
	case enumtypes.GitHubRateLimitResumeActionKindRunStatusCommentRetry,
		enumtypes.GitHubRateLimitResumeActionKindPlatformCallReplay:
		return incoming, nil
	default:
		if len(existing) == 0 {
			return json.RawMessage(`{}`), nil
		}
		return existing, nil
	}
}

func (s *Service) loadProjectionAndCommentContext(ctx context.Context, runID string) (WaitProjection, CommentRenderContext, error) {
	projection, found, err := s.GetRunProjection(ctx, runID)
	if err != nil {
		return WaitProjection{}, CommentRenderContext{}, err
	}
	if !found {
		return WaitProjection{}, CommentRenderContext{}, nil
	}
	commentRenderContext, err := s.BuildCommentRenderContext(projection)
	if err != nil {
		return WaitProjection{}, CommentRenderContext{}, err
	}
	return projection, commentRenderContext, nil
}

func (s *Service) insertRunWaitFlowEvent(ctx context.Context, correlationID string, wait Wait, classification Classification, refresh waitrepo.RefreshProjectionResult, projection WaitProjection) {
	if s.flowEvents == nil || strings.TrimSpace(correlationID) == "" {
		return
	}

	eventKey := githubRateLimitWaitEnteredEventKey
	if classification.NextStepKind == enumtypes.GitHubRateLimitNextStepKindManualActionRequired {
		eventKey = githubRateLimitManualActionEventKey
	}

	_ = s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: correlationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorIDControlPlane,
		EventType:     waitPausedEventType(),
		Payload: marshalJSONPayload(runWaitFlowEventPayload{
			RunID:              wait.RunID,
			WaitID:             wait.ID,
			ContourKind:        wait.ContourKind,
			LimitKind:          wait.LimitKind,
			State:              wait.State,
			NextStepKind:       classification.NextStepKind,
			ResumeNotBefore:    wait.ResumeNotBefore,
			OpenWaitCount:      refresh.OpenWaitCount,
			DominantWaitID:     refresh.DominantWaitID,
			EventKey:           eventKey,
			CommentMirrorState: projection.CommentMirrorState,
		}),
		CreatedAt: s.now(),
	})
}

func (s *Service) assertSignalWriteAllowed(origin enumtypes.GitHubRateLimitSignalOrigin) error {
	caps, err := s.capabilities()
	if err != nil {
		return err
	}
	switch origin {
	case enumtypes.GitHubRateLimitSignalOriginAgentRunner:
		if !caps.CanAcceptSignals {
			return fmt.Errorf("github rate-limit agent-runner signals require runner rollout readiness")
		}
	default:
		if !caps.CanPersistWaits {
			return fmt.Errorf("github rate-limit persistence requires core feature flag and domain rollout readiness")
		}
	}
	return nil
}

func (s *Service) assertReadAllowed() error {
	caps, err := s.capabilities()
	if err != nil {
		return err
	}
	if !caps.CanPersistWaits {
		return fmt.Errorf("github rate-limit read path requires core feature flag and domain rollout readiness")
	}
	return nil
}

func isTerminalRunStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case runStatusSucceeded, runStatusFailed, runStatusCanceled:
		return true
	default:
		return false
	}
}

func ptrInt(value int) *int {
	return &value
}
