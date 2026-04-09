package githubratelimit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	floweventdomain "github.com/codex-k8s/kodex/libs/go/domain/flowevent"
	sharedgithubratelimit "github.com/codex-k8s/kodex/libs/go/domain/githubratelimit"
	agentrunrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/agentrun"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

func TestProcessNextAutoResumeSchedulesAgentSessionResume(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 14, 18, 0, 0, 0, time.UTC)
	runPayload := json.RawMessage(`{
		"project":{"id":"project-1"},
		"agent":{"id":"agent-1"},
		"learning_mode":true
	}`)
	runs := &fakeRunRepository{
		items: map[string]agentrunrepo.Run{
			"run-1": {
				ID:            "run-1",
				ProjectID:     "project-1",
				CorrelationID: "corr-1",
				Status:        runStatusWaitingBackpressure,
				RunPayload:    runPayload,
			},
		},
	}
	waits := newFakeWaitRepository(now)
	waits.waits["wait-1"] = Wait{
		ID:                     "wait-1",
		ProjectID:              "project-1",
		RunID:                  "run-1",
		ContourKind:            enumtypes.GitHubRateLimitContourKindAgentBotToken,
		SignalOrigin:           enumtypes.GitHubRateLimitSignalOriginAgentRunner,
		OperationClass:         enumtypes.GitHubRateLimitOperationClassAgentGitHubCall,
		State:                  enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled,
		LimitKind:              enumtypes.GitHubRateLimitLimitKindSecondary,
		Confidence:             enumtypes.GitHubRateLimitConfidenceProviderUnclear,
		RecoveryHintKind:       enumtypes.GitHubRateLimitRecoveryHintKindExponentialBackoff,
		SignalID:               "signal-1",
		CorrelationID:          "corr-1",
		ResumeActionKind:       enumtypes.GitHubRateLimitResumeActionKindAgentSessionResume,
		AutoResumeAttemptsUsed: 0,
		MaxAutoResumeAttempts:  3,
		ResumeNotBefore:        ptrTime(now.Add(-time.Minute)),
		FirstDetectedAt:        now.Add(-5 * time.Minute),
		LastSignalAt:           now.Add(-4 * time.Minute),
		CreatedAt:              now.Add(-5 * time.Minute),
		UpdatedAt:              now.Add(-4 * time.Minute),
	}
	waits.claimIDs = []string{"wait-1"}
	events := &fakeFlowEventRecorder{}

	service := newTestService(t, valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
		WorkerReady:        true,
	}, waits, runs, events, now, nil, nil)

	result, err := service.ProcessNextAutoResume(context.Background(), ProcessNextAutoResumeParams{WorkerID: "worker-1"})
	if err != nil {
		t.Fatalf("ProcessNextAutoResume() error = %v", err)
	}
	if !result.Found {
		t.Fatal("expected due wait to be found")
	}
	if got, want := result.ResolutionKind, enumtypes.GitHubRateLimitResolutionKindAutoResumed; got != want {
		t.Fatalf("resolution_kind = %q, want %q", got, want)
	}
	if got, want := result.Wait.State, enumtypes.GitHubRateLimitWaitStateResolved; got != want {
		t.Fatalf("wait.state = %q, want %q", got, want)
	}
	if got, want := result.RequeuedCorrelationID, sharedgithubratelimit.ResumeCorrelationPrefix+"wait-1"; got != want {
		t.Fatalf("requeued_correlation_id = %q, want %q", got, want)
	}
	if got := len(runs.creates); got != 1 {
		t.Fatalf("pending resume runs = %d, want 1", got)
	}

	var pendingPayload map[string]json.RawMessage
	if err := json.Unmarshal(runs.creates[0].RunPayload, &pendingPayload); err != nil {
		t.Fatalf("decode pending run payload: %v", err)
	}
	if _, ok := pendingPayload[sharedgithubratelimit.ResumePayloadRunPayloadFieldName]; !ok {
		t.Fatalf("expected %q in pending run payload", sharedgithubratelimit.ResumePayloadRunPayloadFieldName)
	}
	if got, want := len(events.items), 1; got != want {
		t.Fatalf("flow events = %d, want %d", got, want)
	}
	if got, want := events.items[0].EventType, floweventdomain.EventTypeRunWaitResumed; got != want {
		t.Fatalf("flow event type = %q, want %q", got, want)
	}
}

func TestProcessNextAutoResumeRetriesRunStatusComment(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 14, 18, 30, 0, 0, time.UTC)
	waits := newFakeWaitRepository(now)
	waits.waits["wait-comment"] = Wait{
		ID:               "wait-comment",
		ProjectID:        "project-1",
		RunID:            "run-comment",
		ContourKind:      enumtypes.GitHubRateLimitContourKindPlatformPAT,
		SignalOrigin:     enumtypes.GitHubRateLimitSignalOriginControlPlane,
		OperationClass:   enumtypes.GitHubRateLimitOperationClassRunStatusComment,
		State:            enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled,
		LimitKind:        enumtypes.GitHubRateLimitLimitKindPrimary,
		Confidence:       enumtypes.GitHubRateLimitConfidenceDeterministic,
		RecoveryHintKind: enumtypes.GitHubRateLimitRecoveryHintKindReset,
		SignalID:         "signal-comment",
		CorrelationID:    "corr-comment",
		ResumeActionKind: enumtypes.GitHubRateLimitResumeActionKindRunStatusCommentRetry,
		ResumePayloadJSON: marshalJSONPayload(valuetypes.GitHubRateLimitRunStatusCommentRetryPayload{
			GitHubRateLimitRunStatusCommentRetryTarget: valuetypes.GitHubRateLimitRunStatusCommentRetryTarget{
				RunID: "run-comment",
				Phase: "started",
			},
			GitHubRateLimitRunStatusCommentRetryRender: valuetypes.GitHubRateLimitRunStatusCommentRetryRender{
				RunStatus: "running",
			},
		}),
		MaxAutoResumeAttempts: 2,
		ResumeNotBefore:       ptrTime(now.Add(-time.Minute)),
		FirstDetectedAt:       now.Add(-2 * time.Minute),
		LastSignalAt:          now.Add(-time.Minute),
		CreatedAt:             now.Add(-2 * time.Minute),
		UpdatedAt:             now.Add(-time.Minute),
	}
	waits.claimIDs = []string{"wait-comment"}
	runStatus := &fakeRunStatusRetrier{}

	service := newTestService(t, valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
		WorkerReady:        true,
	}, waits, &fakeRunRepository{}, &fakeFlowEventRecorder{}, now, runStatus, nil)

	result, err := service.ProcessNextAutoResume(context.Background(), ProcessNextAutoResumeParams{WorkerID: "worker-1"})
	if err != nil {
		t.Fatalf("ProcessNextAutoResume() error = %v", err)
	}
	if !result.Found {
		t.Fatal("expected due wait to be found")
	}
	if got := len(runStatus.payloads); got != 1 {
		t.Fatalf("run status retries = %d, want 1", got)
	}
	if got, want := result.Wait.State, enumtypes.GitHubRateLimitWaitStateResolved; got != want {
		t.Fatalf("wait.state = %q, want %q", got, want)
	}
}

func TestProcessNextAutoResumeReschedulesReplayFailureWithinBudget(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 14, 19, 0, 0, 0, time.UTC)
	waits := newFakeWaitRepository(now)
	waits.waits["wait-platform"] = Wait{
		ID:               "wait-platform",
		ProjectID:        "project-1",
		RunID:            "run-platform",
		ContourKind:      enumtypes.GitHubRateLimitContourKindPlatformPAT,
		SignalOrigin:     enumtypes.GitHubRateLimitSignalOriginControlPlane,
		OperationClass:   enumtypes.GitHubRateLimitOperationClassIssueLabelTransition,
		State:            enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled,
		LimitKind:        enumtypes.GitHubRateLimitLimitKindSecondary,
		Confidence:       enumtypes.GitHubRateLimitConfidenceConservative,
		RecoveryHintKind: enumtypes.GitHubRateLimitRecoveryHintKindRetryAfter,
		SignalID:         "signal-platform",
		CorrelationID:    "corr-platform",
		ResumeActionKind: enumtypes.GitHubRateLimitResumeActionKindPlatformCallReplay,
		ResumePayloadJSON: marshalJSONPayload(valuetypes.GitHubRateLimitPlatformCallReplayPayload{
			OperationKind:            enumtypes.GitHubRateLimitPlatformReplayOperationKindIssueStageTransition,
			RepositoryFullName:       "codex-k8s/kodex",
			IssueNumber:              427,
			TargetLabel:              "run:qa",
			RequestFingerprint:       "issue:427:run:dev:revise->run:qa",
			CorrelationID:            "corr-platform",
			ExpectedCurrentRunLabels: []string{"run:dev:revise"},
		}),
		MaxAutoResumeAttempts: 2,
		ResumeNotBefore:       ptrTime(now.Add(-time.Minute)),
		FirstDetectedAt:       now.Add(-2 * time.Minute),
		LastSignalAt:          now.Add(-2*time.Minute + 15*time.Second),
		CreatedAt:             now.Add(-2 * time.Minute),
		UpdatedAt:             now.Add(-time.Minute),
	}
	waits.claimIDs = []string{"wait-platform"}
	platform := &fakePlatformCallReplayer{err: errors.New("github still rate limited")}

	service := newTestService(t, valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
		WorkerReady:        true,
	}, waits, &fakeRunRepository{}, &fakeFlowEventRecorder{}, now, nil, platform)

	result, err := service.ProcessNextAutoResume(context.Background(), ProcessNextAutoResumeParams{WorkerID: "worker-1"})
	if err != nil {
		t.Fatalf("ProcessNextAutoResume() error = %v", err)
	}
	if !result.Found {
		t.Fatal("expected due wait to be found")
	}
	if got, want := result.Wait.State, enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled; got != want {
		t.Fatalf("wait.state = %q, want %q", got, want)
	}
	if result.ManualActionKind != "" {
		t.Fatalf("manual_action_kind = %q, want empty", result.ManualActionKind)
	}
	if result.Wait.ResumeNotBefore == nil {
		t.Fatal("expected resume_not_before to be rescheduled")
	}
	if got, want := result.Wait.AutoResumeAttemptsUsed, 1; got != want {
		t.Fatalf("auto_resume_attempts_used = %d, want %d", got, want)
	}
	if got, want := result.Wait.Confidence, enumtypes.GitHubRateLimitConfidenceConservative; got != want {
		t.Fatalf("confidence = %q, want %q", got, want)
	}
	if got, want := result.Wait.RecoveryHintKind, enumtypes.GitHubRateLimitRecoveryHintKindRetryAfter; got != want {
		t.Fatalf("recovery_hint_kind = %q, want %q", got, want)
	}
	expectedResume := now.Add(45 * time.Second)
	if got := result.Wait.ResumeNotBefore.UTC(); !got.Equal(expectedResume) {
		t.Fatalf("resume_not_before = %s, want %s", got, expectedResume)
	}
	if got := len(waits.evidence); got != 3 {
		t.Fatalf("evidence count = %d, want 3", got)
	}
}

func TestProcessNextAutoResumeEscalatesReplayFailureAfterBudgetExhausted(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 14, 19, 10, 0, 0, time.UTC)
	waits := newFakeWaitRepository(now)
	waits.waits["wait-platform"] = Wait{
		ID:                     "wait-platform",
		ProjectID:              "project-1",
		RunID:                  "run-platform",
		ContourKind:            enumtypes.GitHubRateLimitContourKindPlatformPAT,
		SignalOrigin:           enumtypes.GitHubRateLimitSignalOriginControlPlane,
		OperationClass:         enumtypes.GitHubRateLimitOperationClassIssueLabelTransition,
		State:                  enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled,
		LimitKind:              enumtypes.GitHubRateLimitLimitKindSecondary,
		Confidence:             enumtypes.GitHubRateLimitConfidenceConservative,
		RecoveryHintKind:       enumtypes.GitHubRateLimitRecoveryHintKindRetryAfter,
		SignalID:               "signal-platform-2",
		CorrelationID:          "corr-platform-2",
		ResumeActionKind:       enumtypes.GitHubRateLimitResumeActionKindPlatformCallReplay,
		AutoResumeAttemptsUsed: 1,
		ResumePayloadJSON: marshalJSONPayload(valuetypes.GitHubRateLimitPlatformCallReplayPayload{
			OperationKind:            enumtypes.GitHubRateLimitPlatformReplayOperationKindIssueStageTransition,
			RepositoryFullName:       "codex-k8s/kodex",
			IssueNumber:              427,
			TargetLabel:              "run:qa",
			RequestFingerprint:       "issue:427:run:dev:revise->run:qa",
			CorrelationID:            "corr-platform-2",
			ExpectedCurrentRunLabels: []string{"run:dev:revise"},
		}),
		MaxAutoResumeAttempts: 2,
		ResumeNotBefore:       ptrTime(now.Add(-time.Minute)),
		FirstDetectedAt:       now.Add(-3 * time.Minute),
		LastSignalAt:          now.Add(-2*time.Minute + 15*time.Second),
		CreatedAt:             now.Add(-3 * time.Minute),
		UpdatedAt:             now.Add(-time.Minute),
	}
	waits.claimIDs = []string{"wait-platform"}
	platform := &fakePlatformCallReplayer{err: errors.New("github still rate limited")}

	service := newTestService(t, valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
		WorkerReady:        true,
	}, waits, &fakeRunRepository{}, &fakeFlowEventRecorder{}, now, nil, platform)

	result, err := service.ProcessNextAutoResume(context.Background(), ProcessNextAutoResumeParams{WorkerID: "worker-1"})
	if err != nil {
		t.Fatalf("ProcessNextAutoResume() error = %v", err)
	}
	if !result.Found {
		t.Fatal("expected due wait to be found")
	}
	if got, want := result.Wait.State, enumtypes.GitHubRateLimitWaitStateManualActionRequired; got != want {
		t.Fatalf("wait.state = %q, want %q", got, want)
	}
	if got, want := result.ManualActionKind, enumtypes.GitHubRateLimitManualActionKindRequeuePlatformOperation; got != want {
		t.Fatalf("manual_action_kind = %q, want %q", got, want)
	}
	if got, want := result.Wait.AutoResumeAttemptsUsed, 2; got != want {
		t.Fatalf("auto_resume_attempts_used = %d, want %d", got, want)
	}
	if got := len(waits.evidence); got != 3 {
		t.Fatalf("evidence count = %d, want 3", got)
	}
}

func TestReportSignalRequiresReplayPayloadForRunStatusCommentRetry(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 14, 19, 15, 0, 0, time.UTC)
	runs := &fakeRunRepository{
		items: map[string]agentrunrepo.Run{
			"run-1": {ID: "run-1", ProjectID: "project-1", CorrelationID: "corr-1", Status: runStatusRunning},
		},
	}

	service := newTestService(t, valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
	}, newFakeWaitRepository(now), runs, &fakeFlowEventRecorder{}, now, nil, nil)

	_, err := service.ReportSignal(context.Background(), ReportSignalParams{
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
				RateLimitResetAt:   ptrTime(now.Add(time.Minute)),
			},
		},
	})
	if !errors.Is(err, errGitHubRateLimitReplayPayloadMissing) {
		t.Fatalf("expected missing replay payload error, got %v", err)
	}
}
