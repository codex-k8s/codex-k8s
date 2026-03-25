package githubratelimit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	sharedgithubratelimit "github.com/codex-k8s/codex-k8s/libs/go/domain/githubratelimit"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	waitrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/githubratelimitwait"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

const (
	githubRateLimitAutoResumeStaleClaimTimeout = 2 * time.Minute
	githubRateLimitResumeSucceededEventKey     = "github_rate_limit.resume_succeeded"
)

type githubRateLimitResumeRunPayload struct {
	Project struct {
		ID string `json:"id"`
	} `json:"project"`
	Agent struct {
		ID string `json:"id"`
	} `json:"agent"`
	LearningMode bool `json:"learning_mode"`
}

// ProcessNextAutoResume claims one due wait and executes deterministic replay or escalation.
func (s *Service) ProcessNextAutoResume(ctx context.Context, params ProcessNextAutoResumeParams) (ProcessNextAutoResumeResult, error) {
	if s == nil {
		return ProcessNextAutoResumeResult{}, fmt.Errorf("github rate-limit service is not configured")
	}
	if err := s.assertWorkerSweepAllowed(); err != nil {
		return ProcessNextAutoResumeResult{}, err
	}

	workerID := strings.TrimSpace(params.WorkerID)
	now := s.now().UTC()

	wait, found, err := s.waits.ClaimNextDueAutoResume(ctx, now, now.Add(-githubRateLimitAutoResumeStaleClaimTimeout))
	if err != nil {
		return ProcessNextAutoResumeResult{}, fmt.Errorf("claim due github rate-limit wait: %w", err)
	}
	if !found {
		return ProcessNextAutoResumeResult{}, nil
	}

	result := ProcessNextAutoResumeResult{
		Found:           true,
		Wait:            wait,
		AttemptNo:       wait.AutoResumeAttemptsUsed,
		ResumeNotBefore: wait.ResumeNotBefore,
	}

	if _, err := s.waits.AppendEvidence(ctx, querytypes.GitHubRateLimitWaitEvidenceCreateParams{
		WaitID:       wait.ID,
		EventKind:    enumtypes.GitHubRateLimitEvidenceEventResumeAttempted,
		SignalID:     wait.SignalID,
		SignalOrigin: enumtypes.GitHubRateLimitSignalOriginWorker,
		PayloadJSON: marshalJSONPayload(waitResumeAttemptEvidencePayload{
			WaitID:    wait.ID,
			AttemptNo: wait.AutoResumeAttemptsUsed,
			WorkerID:  workerID,
		}),
		ObservedAt: now,
	}); err != nil {
		return ProcessNextAutoResumeResult{}, fmt.Errorf("append github rate-limit resume-attempt evidence: %w", err)
	}

	cancelledWait, cancelled, err := s.cancelAutoResumeForTerminalRun(ctx, wait, now)
	if err != nil {
		return ProcessNextAutoResumeResult{}, err
	}
	if cancelled {
		result.Wait = cancelledWait
		result.ResolutionKind = enumtypes.GitHubRateLimitResolutionKindCancelled
		return result, nil
	}

	resumePayloadJSON, requeuedCorrelationID, attemptErr := s.executeAutoResume(ctx, wait, now)
	if attemptErr != nil {
		wait, err = s.handleAutoResumeFailure(ctx, wait, workerID, now, attemptErr)
		if err != nil {
			return ProcessNextAutoResumeResult{}, err
		}
		result.Wait = wait
		result.ResumeNotBefore = wait.ResumeNotBefore
		if wait.State == enumtypes.GitHubRateLimitWaitStateManualActionRequired {
			result.ManualActionKind = wait.ManualActionKind
		}
		return result, nil
	}

	wait, err = s.resolveAutoResumeSuccess(ctx, wait, now, resumePayloadJSON, requeuedCorrelationID)
	if err != nil {
		return ProcessNextAutoResumeResult{}, err
	}

	result.Wait = wait
	result.ResolutionKind = enumtypes.GitHubRateLimitResolutionKindAutoResumed
	result.RequeuedCorrelationID = requeuedCorrelationID
	return result, nil
}

func (s *Service) cancelAutoResumeForTerminalRun(ctx context.Context, wait Wait, now time.Time) (Wait, bool, error) {
	run, found, err := s.runs.GetByID(ctx, wait.RunID)
	if err != nil {
		return Wait{}, false, fmt.Errorf("load run for github rate-limit auto-resume: %w", err)
	}
	if !found || !isTerminalRunStatus(run.Status) {
		return Wait{}, false, nil
	}

	resolvedAt := now.UTC()
	wait, found, err = s.waits.Update(ctx, querytypes.GitHubRateLimitWaitUpdateParams{
		ID:                     wait.ID,
		SignalOrigin:           enumtypes.GitHubRateLimitSignalOriginWorker,
		OperationClass:         wait.OperationClass,
		State:                  enumtypes.GitHubRateLimitWaitStateCancelled,
		LimitKind:              wait.LimitKind,
		Confidence:             wait.Confidence,
		RecoveryHintKind:       wait.RecoveryHintKind,
		SignalID:               wait.SignalID,
		RequestFingerprint:     wait.RequestFingerprint,
		CorrelationID:          wait.CorrelationID,
		ResumeActionKind:       wait.ResumeActionKind,
		ResumePayloadJSON:      wait.ResumePayloadJSON,
		ManualActionKind:       "",
		AutoResumeAttemptsUsed: wait.AutoResumeAttemptsUsed,
		MaxAutoResumeAttempts:  wait.MaxAutoResumeAttempts,
		ResumeNotBefore:        wait.ResumeNotBefore,
		LastResumeAttemptAt:    &resolvedAt,
		LastSignalAt:           resolvedAt,
		ResolvedAt:             &resolvedAt,
	})
	if err != nil {
		return Wait{}, false, fmt.Errorf("cancel github rate-limit wait for terminal run: %w", err)
	}
	if !found {
		return Wait{}, false, fmt.Errorf("wait %s not found while cancelling auto-resume for terminal run", wait.ID)
	}

	if err := s.appendResolvedEvidence(ctx, wait, resolvedAt, enumtypes.GitHubRateLimitResolutionKindCancelled, ""); err != nil {
		return Wait{}, false, fmt.Errorf("append github rate-limit cancel evidence: %w", err)
	}

	refreshResult, err := s.waits.RefreshRunProjection(ctx, wait.RunID)
	if err != nil {
		return Wait{}, false, fmt.Errorf("refresh github rate-limit projection after cancel: %w", err)
	}

	commentMirrorState := enumtypes.GitHubRateLimitCommentMirrorStateNotAttempted
	if projection, projectionFound, projectionErr := s.GetRunProjection(ctx, wait.RunID); projectionErr == nil && projectionFound {
		commentMirrorState = projection.CommentMirrorState
	}

	s.insertRunWaitResolvedFlowEvent(
		ctx,
		wait.CorrelationID,
		wait,
		enumtypes.GitHubRateLimitResolutionKindCancelled,
		wait.AutoResumeAttemptsUsed,
		"",
		commentMirrorState,
		refreshResult,
	)
	return wait, true, nil
}

func (s *Service) assertWorkerSweepAllowed() error {
	caps, err := s.capabilities()
	if err != nil {
		return err
	}
	if !caps.CanRunWorkerSweep {
		return fmt.Errorf("github rate-limit worker sweep requires worker rollout readiness")
	}
	return nil
}

func (s *Service) executeAutoResume(ctx context.Context, wait Wait, now time.Time) (json.RawMessage, string, error) {
	switch wait.ResumeActionKind {
	case enumtypes.GitHubRateLimitResumeActionKindAgentSessionResume:
		return s.scheduleAgentSessionResume(ctx, wait, now)
	case enumtypes.GitHubRateLimitResumeActionKindRunStatusCommentRetry:
		return nil, "", executeGitHubRateLimitReplay(
			s.runStatus,
			wait.ResumePayloadJSON,
			"run status retry service is not configured",
			"decode github rate-limit run-status retry payload",
			func(payload valuetypes.GitHubRateLimitRunStatusCommentRetryPayload) error {
				return s.runStatus.RetryGitHubRateLimitComment(ctx, payload)
			},
		)
	case enumtypes.GitHubRateLimitResumeActionKindPlatformCallReplay:
		return nil, "", executeGitHubRateLimitReplay(
			s.platform,
			wait.ResumePayloadJSON,
			"platform replay service is not configured",
			"decode github rate-limit platform replay payload",
			func(payload valuetypes.GitHubRateLimitPlatformCallReplayPayload) error {
				return s.platform.ReplayGitHubRateLimitPlatformCall(ctx, payload)
			},
		)
	default:
		return nil, "", fmt.Errorf("unsupported github rate-limit resume_action_kind %q", wait.ResumeActionKind)
	}
}

func (s *Service) scheduleAgentSessionResume(ctx context.Context, wait Wait, now time.Time) (json.RawMessage, string, error) {
	run, found, err := s.runs.GetByID(ctx, wait.RunID)
	if err != nil {
		return nil, "", fmt.Errorf("load run for github rate-limit resume: %w", err)
	}
	if !found {
		return nil, "", fmt.Errorf("run %s not found for github rate-limit resume", wait.RunID)
	}

	runMeta, err := parseGitHubRateLimitResumeRunPayload(run.RunPayload)
	if err != nil {
		return nil, "", err
	}
	resumePayload, err := s.BuildAgentSessionResumePayload(BuildResumePayloadParams{
		Wait:           wait,
		ResolutionKind: enumtypes.GitHubRateLimitResolutionKindAutoResumed,
		RecoveredAt:    now,
		AttemptNo:      wait.AutoResumeAttemptsUsed,
	})
	if err != nil {
		return nil, "", err
	}
	pendingRunPayload, err := buildGitHubRateLimitResumePendingRunPayload(run.RunPayload, resumePayload.Raw)
	if err != nil {
		return nil, "", err
	}

	projectID := strings.TrimSpace(run.ProjectID)
	if projectID == "" {
		projectID = strings.TrimSpace(runMeta.Project.ID)
	}
	agentID := strings.TrimSpace(runMeta.Agent.ID)
	if agentID == "" {
		return nil, "", fmt.Errorf("run payload missing agent.id for github rate-limit resume")
	}

	correlationID := buildGitHubRateLimitResumeCorrelationID(wait.ID)
	createResult, err := s.runs.CreatePendingIfAbsent(ctx, agentrunrepo.CreateParams{
		CorrelationID: correlationID,
		ProjectID:     projectID,
		AgentID:       agentID,
		RunPayload:    pendingRunPayload,
		LearningMode:  runMeta.LearningMode,
	})
	if err != nil {
		return nil, "", fmt.Errorf("create pending github rate-limit resume run: %w", err)
	}
	if strings.TrimSpace(createResult.RunID) == "" {
		return nil, "", fmt.Errorf("control-plane returned empty run id for github rate-limit resume")
	}

	return resumePayload.Raw, correlationID, nil
}

func (s *Service) resolveAutoResumeSuccess(
	ctx context.Context,
	wait Wait,
	now time.Time,
	resumePayloadJSON json.RawMessage,
	requeuedCorrelationID string,
) (Wait, error) {
	nextResumePayload := wait.ResumePayloadJSON
	if len(resumePayloadJSON) > 0 {
		nextResumePayload = append(json.RawMessage(nil), resumePayloadJSON...)
	}

	resolvedAt := now.UTC()
	wait, found, err := s.waits.Update(ctx, querytypes.GitHubRateLimitWaitUpdateParams{
		ID:                     wait.ID,
		SignalOrigin:           enumtypes.GitHubRateLimitSignalOriginWorker,
		OperationClass:         wait.OperationClass,
		State:                  enumtypes.GitHubRateLimitWaitStateResolved,
		LimitKind:              wait.LimitKind,
		Confidence:             wait.Confidence,
		RecoveryHintKind:       wait.RecoveryHintKind,
		SignalID:               wait.SignalID,
		RequestFingerprint:     wait.RequestFingerprint,
		CorrelationID:          wait.CorrelationID,
		ResumeActionKind:       wait.ResumeActionKind,
		ResumePayloadJSON:      nextResumePayload,
		ManualActionKind:       "",
		AutoResumeAttemptsUsed: wait.AutoResumeAttemptsUsed,
		MaxAutoResumeAttempts:  wait.MaxAutoResumeAttempts,
		ResumeNotBefore:        wait.ResumeNotBefore,
		LastResumeAttemptAt:    &resolvedAt,
		LastSignalAt:           wait.LastSignalAt,
		ResolvedAt:             &resolvedAt,
	})
	if err != nil {
		return Wait{}, fmt.Errorf("resolve github rate-limit wait: %w", err)
	}
	if !found {
		return Wait{}, fmt.Errorf("wait %s not found while resolving github rate-limit auto-resume", wait.ID)
	}

	if err := s.appendResolvedEvidence(ctx, wait, resolvedAt, enumtypes.GitHubRateLimitResolutionKindAutoResumed, requeuedCorrelationID); err != nil {
		return Wait{}, fmt.Errorf("append github rate-limit resolved evidence: %w", err)
	}

	refreshResult, err := s.waits.RefreshRunProjection(ctx, wait.RunID)
	if err != nil {
		return Wait{}, fmt.Errorf("refresh github rate-limit projection after resolve: %w", err)
	}

	commentMirrorState := enumtypes.GitHubRateLimitCommentMirrorStateNotAttempted
	if projection, found, err := s.GetRunProjection(ctx, wait.RunID); err == nil && found {
		commentMirrorState = projection.CommentMirrorState
	}

	s.insertRunWaitResolvedFlowEvent(ctx, wait.CorrelationID, wait, enumtypes.GitHubRateLimitResolutionKindAutoResumed, wait.AutoResumeAttemptsUsed, requeuedCorrelationID, commentMirrorState, refreshResult)
	return wait, nil
}

func (s *Service) appendResolvedEvidence(
	ctx context.Context,
	wait Wait,
	observedAt time.Time,
	resolutionKind enumtypes.GitHubRateLimitResolutionKind,
	requeuedCorrelationID string,
) error {
	_, err := s.waits.AppendEvidence(ctx, querytypes.GitHubRateLimitWaitEvidenceCreateParams{
		WaitID:       wait.ID,
		EventKind:    enumtypes.GitHubRateLimitEvidenceEventResolved,
		SignalID:     wait.SignalID,
		SignalOrigin: enumtypes.GitHubRateLimitSignalOriginWorker,
		PayloadJSON: marshalJSONPayload(waitResolvedEvidencePayload{
			WaitID:                wait.ID,
			AttemptNo:             wait.AutoResumeAttemptsUsed,
			ResolutionKind:        resolutionKind,
			RequeuedCorrelationID: requeuedCorrelationID,
		}),
		ObservedAt: observedAt,
	})
	return err
}

func (s *Service) handleAutoResumeFailure(
	ctx context.Context,
	wait Wait,
	workerID string,
	now time.Time,
	attemptErr error,
) (Wait, error) {
	attemptNo := wait.AutoResumeAttemptsUsed
	classification := classifyAutoResumeReplayFailure(wait, now)
	nextResumeNotBefore := classification.ResumeNotBefore

	if _, err := s.waits.AppendEvidence(ctx, querytypes.GitHubRateLimitWaitEvidenceCreateParams{
		WaitID:       wait.ID,
		EventKind:    enumtypes.GitHubRateLimitEvidenceEventResumeFailed,
		SignalID:     wait.SignalID,
		SignalOrigin: enumtypes.GitHubRateLimitSignalOriginWorker,
		PayloadJSON: marshalJSONPayload(waitResumeFailureEvidencePayload{
			WaitID:          wait.ID,
			AttemptNo:       attemptNo,
			WorkerID:        workerID,
			Error:           trimToMaxBytes(strings.TrimSpace(attemptErr.Error()), signalExcerptMaxBytes),
			NextStepKind:    classification.NextStepKind,
			ResumeNotBefore: nextResumeNotBefore,
		}),
		ObservedAt: now,
	}); err != nil {
		return Wait{}, fmt.Errorf("append github rate-limit resume-failed evidence: %w", err)
	}

	lastResumeAttemptAt := now.UTC()
	wait, found, err := s.waits.Update(ctx, querytypes.GitHubRateLimitWaitUpdateParams{
		ID:                     wait.ID,
		SignalOrigin:           enumtypes.GitHubRateLimitSignalOriginWorker,
		OperationClass:         wait.OperationClass,
		State:                  classification.State,
		LimitKind:              classification.LimitKind,
		Confidence:             classification.Confidence,
		RecoveryHintKind:       classification.RecoveryHintKind,
		SignalID:               wait.SignalID,
		RequestFingerprint:     wait.RequestFingerprint,
		CorrelationID:          wait.CorrelationID,
		ResumeActionKind:       wait.ResumeActionKind,
		ResumePayloadJSON:      wait.ResumePayloadJSON,
		ManualActionKind:       classification.ManualActionKind,
		AutoResumeAttemptsUsed: attemptNo,
		MaxAutoResumeAttempts:  classification.MaxAutoResumeAttempts,
		ResumeNotBefore:        nextResumeNotBefore,
		LastResumeAttemptAt:    &lastResumeAttemptAt,
		LastSignalAt:           wait.LastSignalAt,
		ResolvedAt:             nil,
	})
	if err != nil {
		return Wait{}, fmt.Errorf("update github rate-limit wait after replay failure: %w", err)
	}
	if !found {
		return Wait{}, fmt.Errorf("wait %s not found while handling github rate-limit replay failure", wait.ID)
	}

	if classification.NextStepKind == enumtypes.GitHubRateLimitNextStepKindAutoResumeScheduled {
		if err := s.appendResumeScheduledEvidence(
			ctx,
			wait,
			wait.SignalID,
			enumtypes.GitHubRateLimitSignalOriginWorker,
			now,
			classification.NextStepKind,
		); err != nil {
			return Wait{}, fmt.Errorf("append github rate-limit resume-scheduled evidence after replay failure: %w", err)
		}
	} else {
		if _, err := s.waits.AppendEvidence(ctx, querytypes.GitHubRateLimitWaitEvidenceCreateParams{
			WaitID:       wait.ID,
			EventKind:    enumtypes.GitHubRateLimitEvidenceEventManualActionRequired,
			SignalID:     wait.SignalID,
			SignalOrigin: enumtypes.GitHubRateLimitSignalOriginWorker,
			PayloadJSON: marshalJSONPayload(waitManualActionEvidencePayload{
				WaitID:           wait.ID,
				AttemptNo:        attemptNo,
				ManualActionKind: wait.ManualActionKind,
				WorkerID:         workerID,
			}),
			ObservedAt: now,
		}); err != nil {
			return Wait{}, fmt.Errorf("append github rate-limit manual-action evidence after replay failure: %w", err)
		}
	}

	refreshResult, err := s.waits.RefreshRunProjection(ctx, wait.RunID)
	if err != nil {
		return Wait{}, fmt.Errorf("refresh github rate-limit projection after replay failure: %w", err)
	}
	projection, _, err := s.loadProjectionAndCommentContext(ctx, wait.RunID)
	if err != nil {
		return Wait{}, err
	}
	s.insertRunWaitFlowEvent(ctx, wait.CorrelationID, wait, classification, refreshResult, projection)

	return wait, nil
}

func classifyAutoResumeReplayFailure(wait Wait, now time.Time) Classification {
	attemptNo := wait.AutoResumeAttemptsUsed
	switch wait.LimitKind {
	case enumtypes.GitHubRateLimitLimitKindPrimary:
		return classifyPrimaryReset(derivePrimaryResetAt(wait, now), attemptNo, wait.ResumeActionKind, now)
	case enumtypes.GitHubRateLimitLimitKindSecondary:
		if wait.RecoveryHintKind == enumtypes.GitHubRateLimitRecoveryHintKindRetryAfter {
			if retryAfterSeconds, ok := deriveRetryAfterSeconds(wait); ok {
				return classifySecondaryRetryAfter(retryAfterSeconds, attemptNo, wait.ResumeActionKind, now)
			}
		}
		return classifySecondaryBackoff(attemptNo, wait.ResumeActionKind, now)
	default:
		return Classification{
			LimitKind:              wait.LimitKind,
			Confidence:             wait.Confidence,
			RecoveryHintKind:       enumtypes.GitHubRateLimitRecoveryHintKindManualOnly,
			State:                  enumtypes.GitHubRateLimitWaitStateManualActionRequired,
			NextStepKind:           enumtypes.GitHubRateLimitNextStepKindManualActionRequired,
			ResumeActionKind:       wait.ResumeActionKind,
			ManualActionKind:       resolveManualActionKind(wait.ResumeActionKind, wait.Confidence, enumtypes.GitHubRateLimitRecoveryHintKindManualOnly),
			AutoResumeAttemptsUsed: attemptNo,
			MaxAutoResumeAttempts:  wait.MaxAutoResumeAttempts,
		}
	}
}

func derivePrimaryResetAt(wait Wait, now time.Time) time.Time {
	if wait.ResumeNotBefore == nil {
		return now
	}
	return wait.ResumeNotBefore.Add(-primaryLimitGuardDelay).UTC()
}

func deriveRetryAfterSeconds(wait Wait) (int, bool) {
	if wait.ResumeNotBefore == nil || wait.LastSignalAt.IsZero() {
		return 0, false
	}

	delay := wait.ResumeNotBefore.Sub(wait.LastSignalAt.UTC()) - secondaryRetryAfterGuardDelay
	if delay <= 0 {
		return 0, false
	}

	seconds := int((delay + time.Second - 1) / time.Second)
	if seconds <= 0 {
		return 0, false
	}
	return seconds, true
}

func decodeGitHubRateLimitReplayPayload[T any](raw json.RawMessage, contextMessage string) (T, error) {
	var zero T
	if len(raw) == 0 {
		return zero, errGitHubRateLimitReplayPayloadMissing
	}

	var payload T
	if err := json.Unmarshal(raw, &payload); err != nil {
		return zero, fmt.Errorf("%s: %w", contextMessage, err)
	}
	return payload, nil
}

func decodeAndExecuteGitHubRateLimitReplay[T any](raw json.RawMessage, contextMessage string, execute func(T) error) error {
	payload, err := decodeGitHubRateLimitReplayPayload[T](raw, contextMessage)
	if err != nil {
		return err
	}
	return execute(payload)
}

func executeGitHubRateLimitReplay[T any](
	dependency any,
	raw json.RawMessage,
	dependencyError string,
	contextMessage string,
	execute func(T) error,
) error {
	if dependency == nil {
		return fmt.Errorf("%s", dependencyError)
	}
	return decodeAndExecuteGitHubRateLimitReplay(raw, contextMessage, execute)
}

func parseGitHubRateLimitResumeRunPayload(raw json.RawMessage) (githubRateLimitResumeRunPayload, error) {
	return querytypes.DecodeRunPayloadInto[githubRateLimitResumeRunPayload](bytes.TrimSpace(raw), "decode run payload for github rate-limit resume")
}

func buildGitHubRateLimitResumePendingRunPayload(raw json.RawMessage, resumePayload json.RawMessage) (json.RawMessage, error) {
	if len(resumePayload) == 0 {
		return nil, fmt.Errorf("github rate-limit resume payload is required")
	}
	if len(resumePayload) > rateLimitResumePayloadMaxBytes {
		return nil, fmt.Errorf("github rate-limit resume payload exceeds %d bytes", rateLimitResumePayloadMaxBytes)
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("decode run payload for github rate-limit resume persistence: %w", err)
	}
	envelope[sharedgithubratelimit.ResumePayloadRunPayloadFieldName] = append(json.RawMessage(nil), resumePayload...)

	encodedRunPayload, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("marshal run payload with github rate-limit resume payload: %w", err)
	}
	return json.RawMessage(encodedRunPayload), nil
}

func buildGitHubRateLimitResumeCorrelationID(waitID string) string {
	return sharedgithubratelimit.ResumeCorrelationPrefix + strings.TrimSpace(waitID)
}

func (s *Service) insertRunWaitResolvedFlowEvent(
	ctx context.Context,
	correlationID string,
	wait Wait,
	resolutionKind enumtypes.GitHubRateLimitResolutionKind,
	attemptNo int,
	requeuedCorrelationID string,
	commentMirrorState enumtypes.GitHubRateLimitCommentMirrorState,
	refresh waitrepo.RefreshProjectionResult,
) {
	if s.flowEvents == nil || strings.TrimSpace(correlationID) == "" {
		return
	}

	_ = s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: correlationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorIDControlPlane,
		EventType:     floweventdomain.EventTypeRunWaitResumed,
		Payload: marshalJSONPayload(runWaitResolvedFlowEventPayload{
			RunID:                 wait.RunID,
			WaitID:                wait.ID,
			ContourKind:           wait.ContourKind,
			LimitKind:             wait.LimitKind,
			OperationClass:        wait.OperationClass,
			ResolutionKind:        resolutionKind,
			AttemptNo:             attemptNo,
			EventKey:              githubRateLimitResumeSucceededEventKey,
			RequeuedCorrelationID: requeuedCorrelationID,
			CommentMirrorState:    resolveCommentMirrorStateAfterRefresh(commentMirrorState, refresh.OpenWaitCount),
		}),
		CreatedAt: s.now(),
	})
}

func resolveCommentMirrorStateAfterRefresh(
	commentMirrorState enumtypes.GitHubRateLimitCommentMirrorState,
	openWaitCount int,
) enumtypes.GitHubRateLimitCommentMirrorState {
	if openWaitCount == 0 {
		return enumtypes.GitHubRateLimitCommentMirrorStateNotAttempted
	}
	return commentMirrorState
}
