package githubratelimit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

func TestReportSignalPrimaryLimitCreatesScheduledWaitAndProjection(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 14, 12, 0, 0, 0, time.UTC)
	resetAt := now.Add(55 * time.Second)

	runs := fakeRunReader{
		items: map[string]agentrunrepo.Run{
			"run-1": {ID: "run-1", ProjectID: "project-1", CorrelationID: "corr-1", Status: runStatusRunning},
		},
	}
	waits := newFakeWaitRepository(now)
	events := &fakeFlowEventRecorder{}
	service := newTestService(t, valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
	}, waits, runs, events, now)

	result, err := service.ReportSignal(context.Background(), ReportSignalParams{
		RunID: "run-1",
		Signal: Signal{
			SignalID:           "signal-primary",
			ContourKind:        enumtypes.GitHubRateLimitContourKindPlatformPAT,
			SignalOrigin:       enumtypes.GitHubRateLimitSignalOriginControlPlane,
			OperationClass:     enumtypes.GitHubRateLimitOperationClassRunStatusComment,
			ProviderStatusCode: 403,
			OccurredAt:         now,
			Headers: Headers{
				RateLimitRemaining: ptrInt(0),
				RateLimitResetAt:   &resetAt,
				GitHubRequestID:    "gh-req-1",
			},
		},
	})
	if err != nil {
		t.Fatalf("ReportSignal() error = %v", err)
	}

	if result.HardFailure {
		t.Fatal("expected recoverable wait, got hard failure")
	}
	if got, want := result.Classification.LimitKind, enumtypes.GitHubRateLimitLimitKindPrimary; got != want {
		t.Fatalf("classification.limit_kind = %q, want %q", got, want)
	}
	if got, want := result.Wait.State, enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled; got != want {
		t.Fatalf("wait.state = %q, want %q", got, want)
	}
	if got, want := result.Wait.ResumeActionKind, enumtypes.GitHubRateLimitResumeActionKindRunStatusCommentRetry; got != want {
		t.Fatalf("wait.resume_action_kind = %q, want %q", got, want)
	}
	if result.Wait.ResumeNotBefore == nil || !result.Wait.ResumeNotBefore.Equal(resetAt.Add(primaryLimitGuardDelay)) {
		t.Fatalf("wait.resume_not_before = %v, want %v", result.Wait.ResumeNotBefore, resetAt.Add(primaryLimitGuardDelay))
	}
	if got, want := result.Projection.DominantWait.WaitID, result.Wait.ID; got != want {
		t.Fatalf("projection.dominant_wait.wait_id = %q, want %q", got, want)
	}
	if got, want := result.Projection.CommentMirrorState, enumtypes.GitHubRateLimitCommentMirrorStatePendingRetry; got != want {
		t.Fatalf("projection.comment_mirror_state = %q, want %q", got, want)
	}
	if !strings.Contains(result.CommentRenderContext.Headline, "platform_pat") {
		t.Fatalf("comment headline must mention contour, got %q", result.CommentRenderContext.Headline)
	}
	if got := len(waits.evidence); got != 2 {
		t.Fatalf("evidence count = %d, want 2", got)
	}
	if got := len(events.items); got != 1 {
		t.Fatalf("flow event count = %d, want 1", got)
	}
	if got, want := events.items[0].EventType, floweventdomain.EventTypeRunWaitPaused; got != want {
		t.Fatalf("flow event type = %q, want %q", got, want)
	}
}

func TestReportSignalSecondaryBackoffEscalatesToManualActionAfterBudget(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 14, 13, 0, 0, 0, time.UTC)
	existingWait := Wait{
		ID:                     "wait-existing",
		ProjectID:              "project-1",
		RunID:                  "run-2",
		ContourKind:            enumtypes.GitHubRateLimitContourKindAgentBotToken,
		SignalOrigin:           enumtypes.GitHubRateLimitSignalOriginAgentRunner,
		OperationClass:         enumtypes.GitHubRateLimitOperationClassAgentGitHubCall,
		State:                  enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled,
		LimitKind:              enumtypes.GitHubRateLimitLimitKindSecondary,
		Confidence:             enumtypes.GitHubRateLimitConfidenceProviderUnclear,
		RecoveryHintKind:       enumtypes.GitHubRateLimitRecoveryHintKindExponentialBackoff,
		SignalID:               "signal-old",
		CorrelationID:          "corr-2",
		ResumeActionKind:       enumtypes.GitHubRateLimitResumeActionKindAgentSessionResume,
		AutoResumeAttemptsUsed: 3,
		MaxAutoResumeAttempts:  3,
		ResumeNotBefore:        ptrTime(now.Add(4 * time.Minute)),
		FirstDetectedAt:        now.Add(-20 * time.Minute),
		LastSignalAt:           now.Add(-4 * time.Minute),
		CreatedAt:              now.Add(-20 * time.Minute),
		UpdatedAt:              now.Add(-4 * time.Minute),
	}

	runs := fakeRunReader{
		items: map[string]agentrunrepo.Run{
			"run-2": {ID: "run-2", ProjectID: "project-1", CorrelationID: "corr-2", Status: runStatusWaitingBackpressure},
		},
	}
	waits := newFakeWaitRepository(now)
	waits.waits[existingWait.ID] = existingWait
	service := newTestService(t, valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
		WorkerReady:        true,
		RunnerReady:        true,
	}, waits, runs, &fakeFlowEventRecorder{}, now)

	result, err := service.ReportSignal(context.Background(), ReportSignalParams{
		RunID: "run-2",
		Signal: Signal{
			SignalID:           "signal-new",
			ContourKind:        enumtypes.GitHubRateLimitContourKindAgentBotToken,
			SignalOrigin:       enumtypes.GitHubRateLimitSignalOriginAgentRunner,
			OperationClass:     enumtypes.GitHubRateLimitOperationClassAgentGitHubCall,
			ProviderStatusCode: 429,
			OccurredAt:         now,
			MessageExcerpt:     "You have exceeded a secondary rate limit",
		},
	})
	if err != nil {
		t.Fatalf("ReportSignal() error = %v", err)
	}

	if got, want := result.Wait.ID, existingWait.ID; got != want {
		t.Fatalf("wait.id = %q, want %q", got, want)
	}
	if got, want := result.Wait.State, enumtypes.GitHubRateLimitWaitStateManualActionRequired; got != want {
		t.Fatalf("wait.state = %q, want %q", got, want)
	}
	if got, want := result.Wait.ManualActionKind, enumtypes.GitHubRateLimitManualActionKindResumeAgentSession; got != want {
		t.Fatalf("wait.manual_action_kind = %q, want %q", got, want)
	}
	if result.Projection.DominantWait.ManualAction == nil {
		t.Fatal("expected manual action in projection")
	}
	if !strings.Contains(result.Projection.DominantWait.ManualAction.Summary, "snapshot") {
		t.Fatalf("manual action summary = %q, want snapshot guidance", result.Projection.DominantWait.ManualAction.Summary)
	}
	if got, want := result.Classification.NextStepKind, enumtypes.GitHubRateLimitNextStepKindManualActionRequired; got != want {
		t.Fatalf("classification.next_step_kind = %q, want %q", got, want)
	}
}

func TestReportSignalHardFailureDoesNotPersistWait(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 14, 14, 0, 0, 0, time.UTC)
	runs := fakeRunReader{
		items: map[string]agentrunrepo.Run{
			"run-3": {ID: "run-3", ProjectID: "project-1", CorrelationID: "corr-3", Status: runStatusRunning},
		},
	}
	waits := newFakeWaitRepository(now)
	service := newTestService(t, valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
	}, waits, runs, &fakeFlowEventRecorder{}, now)

	result, err := service.ReportSignal(context.Background(), ReportSignalParams{
		RunID: "run-3",
		Signal: Signal{
			SignalID:           "signal-hard-failure",
			ContourKind:        enumtypes.GitHubRateLimitContourKindPlatformPAT,
			SignalOrigin:       enumtypes.GitHubRateLimitSignalOriginControlPlane,
			OperationClass:     enumtypes.GitHubRateLimitOperationClassRepositoryProvider,
			ProviderStatusCode: 403,
			OccurredAt:         now,
			MessageExcerpt:     "Bad credentials",
		},
	})
	if err != nil {
		t.Fatalf("ReportSignal() error = %v", err)
	}
	if !result.HardFailure {
		t.Fatal("expected hard failure result")
	}
	if len(waits.waits) != 0 {
		t.Fatalf("waits persisted = %d, want 0", len(waits.waits))
	}
	if len(waits.evidence) != 0 {
		t.Fatalf("evidence persisted = %d, want 0", len(waits.evidence))
	}
}

func TestReportSignalDuplicateSignalReturnsExistingWait(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 14, 15, 0, 0, 0, time.UTC)
	existingWait := Wait{
		ID:                    "wait-dup",
		ProjectID:             "project-1",
		RunID:                 "run-4",
		ContourKind:           enumtypes.GitHubRateLimitContourKindPlatformPAT,
		SignalOrigin:          enumtypes.GitHubRateLimitSignalOriginControlPlane,
		OperationClass:        enumtypes.GitHubRateLimitOperationClassIssueLabelTransition,
		State:                 enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled,
		LimitKind:             enumtypes.GitHubRateLimitLimitKindSecondary,
		Confidence:            enumtypes.GitHubRateLimitConfidenceConservative,
		RecoveryHintKind:      enumtypes.GitHubRateLimitRecoveryHintKindRetryAfter,
		SignalID:              "signal-dup",
		CorrelationID:         "corr-4",
		ResumeActionKind:      enumtypes.GitHubRateLimitResumeActionKindPlatformCallReplay,
		MaxAutoResumeAttempts: 2,
		ResumeNotBefore:       ptrTime(now.Add(30 * time.Second)),
		FirstDetectedAt:       now.Add(-2 * time.Minute),
		LastSignalAt:          now.Add(-30 * time.Second),
		CreatedAt:             now.Add(-2 * time.Minute),
		UpdatedAt:             now.Add(-30 * time.Second),
		DominantForRun:        true,
	}

	runs := fakeRunReader{
		items: map[string]agentrunrepo.Run{
			"run-4": {ID: "run-4", ProjectID: "project-1", CorrelationID: "corr-4", Status: runStatusWaitingBackpressure},
		},
	}
	waits := newFakeWaitRepository(now)
	waits.waits[existingWait.ID] = existingWait
	events := &fakeFlowEventRecorder{}
	service := newTestService(t, valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
	}, waits, runs, events, now)

	result, err := service.ReportSignal(context.Background(), ReportSignalParams{
		RunID: "run-4",
		Signal: Signal{
			SignalID:           "signal-dup",
			ContourKind:        enumtypes.GitHubRateLimitContourKindPlatformPAT,
			SignalOrigin:       enumtypes.GitHubRateLimitSignalOriginControlPlane,
			OperationClass:     enumtypes.GitHubRateLimitOperationClassIssueLabelTransition,
			ProviderStatusCode: 429,
			OccurredAt:         now,
		},
	})
	if err != nil {
		t.Fatalf("ReportSignal() error = %v", err)
	}
	if !result.DuplicateSignal {
		t.Fatal("expected duplicate signal fast-path")
	}
	if got := len(waits.evidence); got != 0 {
		t.Fatalf("evidence count = %d, want 0", got)
	}
	if got := len(events.items); got != 0 {
		t.Fatalf("flow event count = %d, want 0", got)
	}
}

func TestReportSignalDuplicateSignalRejectsStaleContext(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 14, 15, 5, 0, 0, time.UTC)
	existingWait := Wait{
		ID:                    "wait-stale-dup",
		ProjectID:             "project-1",
		RunID:                 "run-other",
		ContourKind:           enumtypes.GitHubRateLimitContourKindPlatformPAT,
		SignalOrigin:          enumtypes.GitHubRateLimitSignalOriginControlPlane,
		OperationClass:        enumtypes.GitHubRateLimitOperationClassIssueLabelTransition,
		State:                 enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled,
		LimitKind:             enumtypes.GitHubRateLimitLimitKindSecondary,
		Confidence:            enumtypes.GitHubRateLimitConfidenceConservative,
		RecoveryHintKind:      enumtypes.GitHubRateLimitRecoveryHintKindRetryAfter,
		SignalID:              "signal-stale-dup",
		CorrelationID:         "corr-stale",
		ResumeActionKind:      enumtypes.GitHubRateLimitResumeActionKindPlatformCallReplay,
		MaxAutoResumeAttempts: 2,
		ResumeNotBefore:       ptrTime(now.Add(30 * time.Second)),
		FirstDetectedAt:       now.Add(-2 * time.Minute),
		LastSignalAt:          now.Add(-30 * time.Second),
		CreatedAt:             now.Add(-2 * time.Minute),
		UpdatedAt:             now.Add(-30 * time.Second),
		DominantForRun:        true,
	}

	runs := fakeRunReader{
		items: map[string]agentrunrepo.Run{
			"run-5": {ID: "run-5", ProjectID: "project-1", CorrelationID: "corr-5", Status: runStatusWaitingBackpressure},
		},
	}
	waits := newFakeWaitRepository(now)
	waits.waits[existingWait.ID] = existingWait
	service := newTestService(t, valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
	}, waits, runs, &fakeFlowEventRecorder{}, now)

	_, err := service.ReportSignal(context.Background(), ReportSignalParams{
		RunID: "run-5",
		Signal: Signal{
			SignalID:           "signal-stale-dup",
			ContourKind:        enumtypes.GitHubRateLimitContourKindPlatformPAT,
			SignalOrigin:       enumtypes.GitHubRateLimitSignalOriginControlPlane,
			OperationClass:     enumtypes.GitHubRateLimitOperationClassIssueLabelTransition,
			ProviderStatusCode: 429,
			OccurredAt:         now,
		},
	})
	if err == nil {
		t.Fatal("expected stale duplicate conflict error")
	}

	var conflictErr errs.Conflict
	if !errors.As(err, &conflictErr) {
		t.Fatalf("expected errs.Conflict, got %T (%v)", err, err)
	}
}

func TestBuildRunProjectionFromWaitsPrefersManualDominantAndBuildsRelatedWaits(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 14, 16, 0, 0, 0, time.UTC)
	manualWait := Wait{
		ID:                     "wait-manual",
		ContourKind:            enumtypes.GitHubRateLimitContourKindPlatformPAT,
		OperationClass:         enumtypes.GitHubRateLimitOperationClassRunStatusComment,
		State:                  enumtypes.GitHubRateLimitWaitStateManualActionRequired,
		LimitKind:              enumtypes.GitHubRateLimitLimitKindPrimary,
		Confidence:             enumtypes.GitHubRateLimitConfidenceConservative,
		RecoveryHintKind:       enumtypes.GitHubRateLimitRecoveryHintKindManualOnly,
		ResumeActionKind:       enumtypes.GitHubRateLimitResumeActionKindRunStatusCommentRetry,
		ManualActionKind:       enumtypes.GitHubRateLimitManualActionKindRequeuePlatformOperation,
		AutoResumeAttemptsUsed: 2,
		MaxAutoResumeAttempts:  2,
		ResumeNotBefore:        ptrTime(now.Add(time.Minute)),
		FirstDetectedAt:        now.Add(-10 * time.Minute),
		UpdatedAt:              now.Add(-30 * time.Second),
	}
	relatedWait := Wait{
		ID:                     "wait-related",
		ContourKind:            enumtypes.GitHubRateLimitContourKindAgentBotToken,
		OperationClass:         enumtypes.GitHubRateLimitOperationClassAgentGitHubCall,
		State:                  enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled,
		LimitKind:              enumtypes.GitHubRateLimitLimitKindSecondary,
		Confidence:             enumtypes.GitHubRateLimitConfidenceProviderUnclear,
		RecoveryHintKind:       enumtypes.GitHubRateLimitRecoveryHintKindExponentialBackoff,
		ResumeActionKind:       enumtypes.GitHubRateLimitResumeActionKindAgentSessionResume,
		AutoResumeAttemptsUsed: 1,
		MaxAutoResumeAttempts:  3,
		ResumeNotBefore:        ptrTime(now.Add(2 * time.Minute)),
		FirstDetectedAt:        now.Add(-5 * time.Minute),
		UpdatedAt:              now.Add(-time.Minute),
	}

	projection, found, err := BuildRunProjectionFromWaits([]Wait{relatedWait, manualWait})
	if err != nil {
		t.Fatalf("BuildRunProjectionFromWaits() error = %v", err)
	}
	if !found {
		t.Fatal("expected projection to be present")
	}
	if got, want := projection.DominantWait.WaitID, manualWait.ID; got != want {
		t.Fatalf("dominant wait = %q, want %q", got, want)
	}
	if got, want := len(projection.RelatedWaits), 1; got != want {
		t.Fatalf("related waits = %d, want %d", got, want)
	}
	if got, want := projection.CommentMirrorState, enumtypes.GitHubRateLimitCommentMirrorStatePendingRetry; got != want {
		t.Fatalf("comment mirror state = %q, want %q", got, want)
	}
	if projection.DominantWait.ManualAction == nil {
		t.Fatal("expected manual action on dominant wait")
	}
}

func TestBuildRunProjectionFromWaitsDefaultsCommentMirrorToNotAttempted(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 14, 16, 5, 0, 0, time.UTC)
	wait := Wait{
		ID:                     "wait-no-mirror",
		ContourKind:            enumtypes.GitHubRateLimitContourKindAgentBotToken,
		OperationClass:         enumtypes.GitHubRateLimitOperationClassAgentGitHubCall,
		State:                  enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled,
		LimitKind:              enumtypes.GitHubRateLimitLimitKindSecondary,
		Confidence:             enumtypes.GitHubRateLimitConfidenceConservative,
		RecoveryHintKind:       enumtypes.GitHubRateLimitRecoveryHintKindRetryAfter,
		ResumeActionKind:       enumtypes.GitHubRateLimitResumeActionKindAgentSessionResume,
		AutoResumeAttemptsUsed: 1,
		MaxAutoResumeAttempts:  2,
		ResumeNotBefore:        ptrTime(now.Add(time.Minute)),
		FirstDetectedAt:        now.Add(-2 * time.Minute),
		UpdatedAt:              now.Add(-30 * time.Second),
	}

	projection, found, err := BuildRunProjectionFromWaits([]Wait{wait})
	if err != nil {
		t.Fatalf("BuildRunProjectionFromWaits() error = %v", err)
	}
	if !found {
		t.Fatal("expected projection to be present")
	}
	if got, want := projection.CommentMirrorState, enumtypes.GitHubRateLimitCommentMirrorStateNotAttempted; got != want {
		t.Fatalf("comment mirror state = %q, want %q", got, want)
	}
}

func TestBuildAgentSessionResumePayload(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 14, 17, 0, 0, 0, time.UTC)
	service := newTestService(t, valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
	}, newFakeWaitRepository(now), fakeRunReader{}, &fakeFlowEventRecorder{}, now)

	result, err := service.BuildAgentSessionResumePayload(BuildResumePayloadParams{
		Wait: Wait{
			ID:               "wait-resume",
			ContourKind:      enumtypes.GitHubRateLimitContourKindAgentBotToken,
			LimitKind:        enumtypes.GitHubRateLimitLimitKindSecondary,
			OperationClass:   enumtypes.GitHubRateLimitOperationClassAgentGitHubCall,
			ResumeActionKind: enumtypes.GitHubRateLimitResumeActionKindAgentSessionResume,
		},
		ResolutionKind: enumtypes.GitHubRateLimitResolutionKindAutoResumed,
		RecoveredAt:    now,
		AttemptNo:      2,
	})
	if err != nil {
		t.Fatalf("BuildAgentSessionResumePayload() error = %v", err)
	}
	if !json.Valid(result.Raw) {
		t.Fatalf("resume payload raw JSON is invalid: %s", string(result.Raw))
	}
	if got, want := result.Payload.WaitReason, enumtypes.AgentRunWaitReasonGitHubRateLimit; got != want {
		t.Fatalf("payload.wait_reason = %q, want %q", got, want)
	}
	if !strings.Contains(result.Payload.Guidance, "snapshot") {
		t.Fatalf("payload.guidance = %q, want snapshot guidance", result.Payload.Guidance)
	}
}

func newTestService(t *testing.T, rollout valuetypes.GitHubRateLimitRolloutState, waits *fakeWaitRepository, runs fakeRunReader, events *fakeFlowEventRecorder, now time.Time) *Service {
	t.Helper()

	service, err := NewService(Config{RolloutState: rollout}, Dependencies{
		Runs:       runs,
		Waits:      waits,
		FlowEvents: events,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	service.now = func() time.Time { return now }
	return service
}

type fakeRunReader struct {
	items map[string]agentrunrepo.Run
}

func (f fakeRunReader) GetByID(_ context.Context, runID string) (agentrunrepo.Run, bool, error) {
	item, found := f.items[runID]
	return item, found, nil
}

type fakeWaitRepository struct {
	waits    map[string]Wait
	evidence []querytypes.GitHubRateLimitWaitEvidenceCreateParams
	nextID   int
	now      time.Time
}

func newFakeWaitRepository(now time.Time) *fakeWaitRepository {
	return &fakeWaitRepository{
		waits:  make(map[string]Wait),
		nextID: 1,
		now:    now,
	}
}

func (f *fakeWaitRepository) Create(_ context.Context, params querytypes.GitHubRateLimitWaitCreateParams) (Wait, error) {
	id := fmt.Sprintf("wait-%d", f.nextID)
	f.nextID++
	item := Wait{
		ID:                     id,
		ProjectID:              params.ProjectID,
		RunID:                  params.RunID,
		ContourKind:            params.ContourKind,
		SignalOrigin:           params.SignalOrigin,
		OperationClass:         params.OperationClass,
		State:                  params.State,
		LimitKind:              params.LimitKind,
		Confidence:             params.Confidence,
		RecoveryHintKind:       params.RecoveryHintKind,
		SignalID:               params.SignalID,
		RequestFingerprint:     params.RequestFingerprint,
		CorrelationID:          params.CorrelationID,
		ResumeActionKind:       params.ResumeActionKind,
		ResumePayloadJSON:      params.ResumePayloadJSON,
		ManualActionKind:       params.ManualActionKind,
		AutoResumeAttemptsUsed: params.AutoResumeAttemptsUsed,
		MaxAutoResumeAttempts:  params.MaxAutoResumeAttempts,
		ResumeNotBefore:        params.ResumeNotBefore,
		LastResumeAttemptAt:    params.LastResumeAttemptAt,
		FirstDetectedAt:        params.FirstDetectedAt,
		LastSignalAt:           params.LastSignalAt,
		ResolvedAt:             params.ResolvedAt,
		CreatedAt:              params.FirstDetectedAt,
		UpdatedAt:              params.LastSignalAt,
	}
	f.waits[item.ID] = item
	return item, nil
}

func (f *fakeWaitRepository) Update(_ context.Context, params querytypes.GitHubRateLimitWaitUpdateParams) (Wait, bool, error) {
	item, found := f.waits[params.ID]
	if !found {
		return Wait{}, false, nil
	}
	item.SignalOrigin = params.SignalOrigin
	item.OperationClass = params.OperationClass
	item.State = params.State
	item.LimitKind = params.LimitKind
	item.Confidence = params.Confidence
	item.RecoveryHintKind = params.RecoveryHintKind
	item.SignalID = params.SignalID
	item.RequestFingerprint = params.RequestFingerprint
	item.CorrelationID = params.CorrelationID
	item.ResumeActionKind = params.ResumeActionKind
	item.ResumePayloadJSON = params.ResumePayloadJSON
	item.ManualActionKind = params.ManualActionKind
	item.AutoResumeAttemptsUsed = params.AutoResumeAttemptsUsed
	item.MaxAutoResumeAttempts = params.MaxAutoResumeAttempts
	item.ResumeNotBefore = params.ResumeNotBefore
	item.LastResumeAttemptAt = params.LastResumeAttemptAt
	item.LastSignalAt = params.LastSignalAt
	item.ResolvedAt = params.ResolvedAt
	item.UpdatedAt = params.LastSignalAt
	f.waits[item.ID] = item
	return item, true, nil
}

func (f *fakeWaitRepository) GetByID(_ context.Context, waitID string) (Wait, bool, error) {
	item, found := f.waits[waitID]
	return item, found, nil
}

func (f *fakeWaitRepository) GetBySignalID(_ context.Context, signalID string) (Wait, bool, error) {
	for _, item := range f.waits {
		if item.SignalID == signalID {
			return item, true, nil
		}
	}
	return Wait{}, false, nil
}

func (f *fakeWaitRepository) GetOpenByRunAndContour(_ context.Context, runID string, contourKind enumtypes.GitHubRateLimitContourKind) (Wait, bool, error) {
	for _, item := range f.waits {
		if item.RunID == runID && item.ContourKind == contourKind && item.State.IsOpen() {
			return item, true, nil
		}
	}
	return Wait{}, false, nil
}

func (f *fakeWaitRepository) ListByRunID(_ context.Context, runID string) ([]Wait, error) {
	items := make([]Wait, 0)
	for _, item := range f.waits {
		if item.RunID == runID {
			items = append(items, item)
		}
	}
	slices.SortStableFunc(items, func(left Wait, right Wait) int {
		if left.UpdatedAt.Equal(right.UpdatedAt) {
			return strings.Compare(left.ID, right.ID)
		}
		if left.UpdatedAt.After(right.UpdatedAt) {
			return -1
		}
		return 1
	})
	return items, nil
}

func (f *fakeWaitRepository) AppendEvidence(_ context.Context, params querytypes.GitHubRateLimitWaitEvidenceCreateParams) (entitytypes.GitHubRateLimitWaitEvidence, error) {
	f.evidence = append(f.evidence, params)
	return entitytypes.GitHubRateLimitWaitEvidence{
		ID:         int64(len(f.evidence)),
		WaitID:     params.WaitID,
		EventKind:  params.EventKind,
		SignalID:   params.SignalID,
		ObservedAt: params.ObservedAt,
	}, nil
}

func (f *fakeWaitRepository) RefreshRunProjection(_ context.Context, runID string) (valuetypes.GitHubRateLimitProjectionRefreshResult, error) {
	openWaits := make([]Wait, 0)
	for _, item := range f.waits {
		if item.RunID == runID {
			item.DominantForRun = false
			f.waits[item.ID] = item
			if item.State.IsOpen() {
				openWaits = append(openWaits, item)
			}
		}
	}

	result := valuetypes.GitHubRateLimitProjectionRefreshResult{
		RunID:         runID,
		OpenWaitCount: len(openWaits),
		SyncState:     enumtypes.GitHubRateLimitProjectionSyncStateCleared,
	}
	dominant, found := ElectDominantWait(openWaits)
	if !found {
		return result, nil
	}

	item := f.waits[dominant.ID]
	item.DominantForRun = true
	f.waits[item.ID] = item
	result.DominantWaitID = dominant.ID
	result.WaitDeadlineAt = dominant.ResumeNotBefore
	result.SyncState = enumtypes.GitHubRateLimitProjectionSyncStateApplied
	return result, nil
}

type fakeFlowEventRecorder struct {
	items []floweventrepo.InsertParams
}

func (f *fakeFlowEventRecorder) Insert(_ context.Context, params floweventrepo.InsertParams) error {
	f.items = append(f.items, params)
	return nil
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
