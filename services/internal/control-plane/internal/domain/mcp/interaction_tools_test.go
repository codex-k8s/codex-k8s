package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	agentsessionrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentsession"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

func TestSetRunWaitContextFailsWhenRunWasNotUpdated(t *testing.T) {
	t.Parallel()

	runs := &interactionTestRunsRepository{setWaitContextUpdated: false}
	sessions := &interactionTestSessionsRepository{setWaitStateUpdated: true}
	service := &Service{
		runs:     runs,
		sessions: sessions,
		now:      time.Now,
	}

	err := service.setRunWaitContext(
		context.Background(),
		SessionContext{RunID: "run-1"},
		waitStateMCP,
		true,
		enumtypes.AgentRunWaitReasonInteractionReply,
		enumtypes.AgentRunWaitTargetKindInteractionRequest,
		"interaction-1",
		nil,
	)
	if err == nil {
		t.Fatal("expected error when run wait context update affects zero rows")
	}
	if !strings.Contains(err.Error(), "run run-1 not found") {
		t.Fatalf("error = %q, want run-not-found message", err)
	}
	if sessions.setWaitStateCalls != 0 {
		t.Fatalf("session wait state calls = %d, want 0", sessions.setWaitStateCalls)
	}
}

func TestClearInteractionWaitContextClearsMatchingRunWait(t *testing.T) {
	t.Parallel()

	runs := &interactionTestRunsRepository{clearWaitContextUpdated: true}
	sessions := &interactionTestSessionsRepository{setWaitStateUpdated: true}
	service := &Service{
		runs:     runs,
		sessions: sessions,
		now:      time.Now,
	}

	cleared, err := service.clearInteractionWaitContext(context.Background(), SessionContext{RunID: "run-1"}, "interaction-1", true)
	if err != nil {
		t.Fatalf("clearInteractionWaitContext returned error: %v", err)
	}
	if !cleared {
		t.Fatal("expected wait context to be cleared")
	}
	if runs.lastClearWaitContext.RunID != "run-1" {
		t.Fatalf("run id = %q, want run-1", runs.lastClearWaitContext.RunID)
	}
	if runs.lastClearWaitContext.WaitReason != enumtypes.AgentRunWaitReasonInteractionReply {
		t.Fatalf("wait reason = %q, want %q", runs.lastClearWaitContext.WaitReason, enumtypes.AgentRunWaitReasonInteractionReply)
	}
	if runs.lastClearWaitContext.WaitTargetKind != enumtypes.AgentRunWaitTargetKindInteractionRequest {
		t.Fatalf("wait target kind = %q, want %q", runs.lastClearWaitContext.WaitTargetKind, enumtypes.AgentRunWaitTargetKindInteractionRequest)
	}
	if runs.lastClearWaitContext.WaitTargetRef != "interaction-1" {
		t.Fatalf("wait target ref = %q, want interaction-1", runs.lastClearWaitContext.WaitTargetRef)
	}
	if sessions.lastSetWaitState.WaitState != string(waitStateNone) {
		t.Fatalf("session wait state = %q, want empty wait state", sessions.lastSetWaitState.WaitState)
	}
}

func TestClearInteractionWaitContextSkipsMissingDuplicateWait(t *testing.T) {
	t.Parallel()

	runs := &interactionTestRunsRepository{clearWaitContextUpdated: false}
	sessions := &interactionTestSessionsRepository{setWaitStateUpdated: true}
	service := &Service{
		runs:     runs,
		sessions: sessions,
		now:      time.Now,
	}

	cleared, err := service.clearInteractionWaitContext(context.Background(), SessionContext{RunID: "run-1"}, "interaction-1", false)
	if err != nil {
		t.Fatalf("clearInteractionWaitContext returned error: %v", err)
	}
	if cleared {
		t.Fatal("expected missing duplicate wait to be ignored")
	}
	if sessions.setWaitStateCalls != 0 {
		t.Fatalf("session wait state calls = %d, want 0", sessions.setWaitStateCalls)
	}
}

type interactionTestRunsRepository struct {
	setWaitContextUpdated   bool
	clearWaitContextUpdated bool
	createPendingCalls      int
	createPendingResult     agentrunrepo.CreateResult
	createPendingErr        error
	lastSetWaitContext      agentrunrepo.SetWaitContextParams
	lastClearWaitContext    agentrunrepo.ClearWaitContextParams
	lastCreatePending       agentrunrepo.CreateParams
	byID                    map[string]agentrunrepo.Run
}

func (r *interactionTestRunsRepository) CreatePendingIfAbsent(_ context.Context, params agentrunrepo.CreateParams) (agentrunrepo.CreateResult, error) {
	r.createPendingCalls++
	r.lastCreatePending = params
	if r.createPendingErr != nil {
		return agentrunrepo.CreateResult{}, r.createPendingErr
	}
	if strings.TrimSpace(r.createPendingResult.RunID) != "" {
		return r.createPendingResult, nil
	}
	return agentrunrepo.CreateResult{RunID: "resume-run-1", Inserted: true}, nil
}

func (r *interactionTestRunsRepository) GetByID(_ context.Context, runID string) (agentrunrepo.Run, bool, error) {
	if item, ok := r.byID[runID]; ok {
		return item, true, nil
	}
	return agentrunrepo.Run{}, false, nil
}

func (r *interactionTestRunsRepository) CancelActiveByID(context.Context, string) (bool, error) {
	return false, nil
}

func (r *interactionTestRunsRepository) ListRecentByProject(context.Context, string, string, int, int) ([]agentrunrepo.RunLookupItem, error) {
	return nil, nil
}

func (r *interactionTestRunsRepository) SearchRecentByProjectIssueOrPullRequest(context.Context, string, string, int64, int64, int) ([]agentrunrepo.RunLookupItem, error) {
	return nil, nil
}

func (r *interactionTestRunsRepository) ListRunIDsByRepositoryIssue(context.Context, string, int64, int) ([]string, error) {
	return nil, nil
}

func (r *interactionTestRunsRepository) ListRunIDsByRepositoryPullRequest(context.Context, string, int64, int) ([]string, error) {
	return nil, nil
}

func (r *interactionTestRunsRepository) SetWaitContext(_ context.Context, params agentrunrepo.SetWaitContextParams) (bool, error) {
	r.lastSetWaitContext = params
	return r.setWaitContextUpdated, nil
}

func (r *interactionTestRunsRepository) ClearWaitContextIfMatches(_ context.Context, params agentrunrepo.ClearWaitContextParams) (bool, error) {
	r.lastClearWaitContext = params
	return r.clearWaitContextUpdated, nil
}

type interactionTestSessionsRepository struct {
	setWaitStateUpdated bool
	setWaitStateCalls   int
	lastSetWaitState    agentsessionrepo.SetWaitStateParams
}

func (r *interactionTestSessionsRepository) Upsert(context.Context, agentsessionrepo.UpsertParams) (valuetypes.AgentSessionSnapshotState, error) {
	return valuetypes.AgentSessionSnapshotState{}, nil
}

func (r *interactionTestSessionsRepository) SetWaitStateByRunID(_ context.Context, params agentsessionrepo.SetWaitStateParams) (bool, error) {
	r.setWaitStateCalls++
	r.lastSetWaitState = params
	return r.setWaitStateUpdated, nil
}

func (r *interactionTestSessionsRepository) GetByRunID(context.Context, string) (agentsessionrepo.Session, bool, error) {
	return entitytypes.AgentSession{}, false, nil
}

func (r *interactionTestSessionsRepository) GetLatestByRepositoryBranchAndAgent(context.Context, string, string, string) (agentsessionrepo.Session, bool, error) {
	return entitytypes.AgentSession{}, false, nil
}

func (r *interactionTestSessionsRepository) CleanupSessionPayloadsFinishedBefore(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func TestFinalizeInteractionResumeSchedulesPendingRun(t *testing.T) {
	t.Parallel()

	runs := &interactionTestRunsRepository{
		clearWaitContextUpdated: true,
		byID: map[string]agentrunrepo.Run{
			"run-1": {
				ID:            "run-1",
				CorrelationID: "corr-1",
				ProjectID:     "project-1",
				RunPayload: json.RawMessage(`{
					"project":{"id":"project-1"},
					"agent":{"id":"agent-dev"},
					"learning_mode":true
				}`),
			},
		},
	}
	sessions := &interactionTestSessionsRepository{setWaitStateUpdated: true}
	service := &Service{
		cfg: Config{
			TokenSigningKey: "test-signing-key",
			TokenIssuer:     "codex-k8s/test",
			PublicBaseURL:   "https://platform.codex-k8s.dev",
		},
		runs:     runs,
		sessions: sessions,
		now: func() time.Time {
			return time.Date(2026, 3, 13, 16, 0, 0, 0, time.UTC)
		},
	}

	scheduled, err := service.finalizeInteractionResume(context.Background(), entitytypes.InteractionRequest{
		ID:              "interaction-1",
		RunID:           "run-1",
		InteractionKind: enumtypes.InteractionKindDecisionRequest,
		State:           enumtypes.InteractionStateDeliveryExhausted,
		UpdatedAt:       time.Date(2026, 3, 13, 16, 5, 0, 0, time.UTC),
	}, nil, true)
	if err != nil {
		t.Fatalf("finalizeInteractionResume returned error: %v", err)
	}
	if !scheduled {
		t.Fatal("expected interaction resume to be scheduled")
	}
	if runs.createPendingCalls != 1 {
		t.Fatalf("create pending calls = %d, want 1", runs.createPendingCalls)
	}
	if got, want := runs.lastCreatePending.CorrelationID, buildInteractionResumeCorrelationID("interaction-1"); got != want {
		t.Fatalf("correlation id = %q, want %q", got, want)
	}
	if got, want := runs.lastCreatePending.AgentID, "agent-dev"; got != want {
		t.Fatalf("agent id = %q, want %q", got, want)
	}
	if got, want := runs.lastCreatePending.ProjectID, "project-1"; got != want {
		t.Fatalf("project id = %q, want %q", got, want)
	}
	if !runs.lastCreatePending.LearningMode {
		t.Fatal("expected learning mode to be preserved for resume run")
	}
	var pendingRunPayload map[string]json.RawMessage
	if err := json.Unmarshal(runs.lastCreatePending.RunPayload, &pendingRunPayload); err != nil {
		t.Fatalf("unmarshal pending run payload: %v", err)
	}
	resumePayloadRaw, ok := pendingRunPayload["interaction_resume_payload"]
	if !ok {
		t.Fatal("expected interaction_resume_payload in pending run payload")
	}
	var resumePayload valuetypes.InteractionResumePayload
	if err := json.Unmarshal(resumePayloadRaw, &resumePayload); err != nil {
		t.Fatalf("unmarshal interaction resume payload: %v", err)
	}
	if got, want := resumePayload.InteractionID, "interaction-1"; got != want {
		t.Fatalf("resume payload interaction_id = %q, want %q", got, want)
	}
	if got, want := string(resumePayload.RequestStatus), string(enumtypes.InteractionRequestStatusDeliveryExhausted); got != want {
		t.Fatalf("resume payload request_status = %q, want %q", got, want)
	}
	if runs.lastClearWaitContext.WaitTargetRef != "interaction-1" {
		t.Fatalf("wait target ref = %q, want interaction-1", runs.lastClearWaitContext.WaitTargetRef)
	}
}

func TestFinalizeInteractionResumeReschedulesAfterWaitWasAlreadyCleared(t *testing.T) {
	t.Parallel()

	runs := &interactionTestRunsRepository{
		clearWaitContextUpdated: false,
		byID: map[string]agentrunrepo.Run{
			"run-1": {
				ID:            "run-1",
				CorrelationID: "corr-1",
				RunPayload:    json.RawMessage(`{"agent":{"id":"agent-dev"}}`),
			},
		},
	}
	sessions := &interactionTestSessionsRepository{setWaitStateUpdated: true}
	service := &Service{
		cfg: Config{
			TokenSigningKey: "test-signing-key",
			TokenIssuer:     "codex-k8s/test",
		},
		runs:     runs,
		sessions: sessions,
		now:      time.Now,
	}

	scheduled, err := service.finalizeInteractionResume(context.Background(), entitytypes.InteractionRequest{
		ID:              "interaction-1",
		RunID:           "run-1",
		InteractionKind: enumtypes.InteractionKindDecisionRequest,
		State:           enumtypes.InteractionStateExpired,
		UpdatedAt:       time.Date(2026, 3, 13, 16, 5, 0, 0, time.UTC),
	}, nil, false)
	if err != nil {
		t.Fatalf("finalizeInteractionResume returned error: %v", err)
	}
	if !scheduled {
		t.Fatal("expected interaction resume to remain scheduled when wait was already cleared")
	}
	if runs.createPendingCalls != 1 {
		t.Fatalf("create pending calls = %d, want 1", runs.createPendingCalls)
	}
	if sessions.setWaitStateCalls != 0 {
		t.Fatalf("session wait state calls = %d, want 0 when wait was already cleared", sessions.setWaitStateCalls)
	}
}

func TestBuildInteractionDeliveryEnvelopeIncludesCallbackContractFields(t *testing.T) {
	t.Parallel()

	deadline := time.Date(2026, 3, 13, 17, 0, 0, 0, time.UTC)
	interactions := &interactionTestRepository{}
	service := &Service{
		cfg: Config{
			TokenSigningKey: "test-signing-key",
			TokenIssuer:     "codex-k8s/test",
			PublicBaseURL:   "https://platform.codex-k8s.dev",
		},
		interactions: interactions,
		now: func() time.Time {
			return time.Date(2026, 3, 13, 16, 0, 0, 0, time.UTC)
		},
	}

	raw, err := service.buildInteractionDeliveryEnvelope(context.Background(), entitytypes.AgentRun{
		ID:            "run-1",
		CorrelationID: "corr-1",
		ProjectID:     "project-1",
	}, entitytypes.InteractionRequest{
		ID:                "interaction-1",
		RunID:             "run-1",
		ChannelFamily:     enumtypes.InteractionChannelFamilyTelegram,
		InteractionKind:   enumtypes.InteractionKindDecisionRequest,
		RecipientProvider: interactionRecipientProviderTelegram,
		RecipientRef:      interactionRecipientRoutingByGitHub + "ai-da-stas",
		RequestPayloadJSON: json.RawMessage(`{
			"question":"Ship it?",
			"options":[{"option_id":"approve","label":"Approve"},{"option_id":"deny","label":"Deny"}]
		}`),
		ContextLinksJSON:   json.RawMessage(`{"run_id":"run-1"}`),
		ResponseDeadlineAt: &deadline,
	}, entitytypes.InteractionDeliveryAttempt{
		DeliveryID: "delivery-1",
	})
	if err != nil {
		t.Fatalf("buildInteractionDeliveryEnvelope returned error: %v", err)
	}

	var envelope interactionDeliveryEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("unmarshal interaction delivery envelope: %v", err)
	}

	if got, want := envelope.SchemaVersion, interactionDeliveryEnvelopeSchemaVersion; got != want {
		t.Fatalf("schema_version = %q, want %q", got, want)
	}
	if got, want := envelope.Locale, interactionDeliveryLocaleDefault; got != want {
		t.Fatalf("locale = %q, want %q", got, want)
	}
	if envelope.CallbackEndpoint == nil {
		t.Fatal("expected callback endpoint to be populated")
	}
	if got, want := envelope.CallbackEndpoint.URL, "https://platform.codex-k8s.dev"+interactionCallbackPath; got != want {
		t.Fatalf("callback url = %q, want %q", got, want)
	}
	if strings.TrimSpace(envelope.CallbackEndpoint.BearerToken) == "" {
		t.Fatal("expected callback bearer token to be populated")
	}
	if got, want := envelope.CallbackEndpoint.TokenExpiresAt, deadline.Add(interactionCallbackTokenGraceTTL).UTC().Format(time.RFC3339Nano); got != want {
		t.Fatalf("token_expires_at = %q, want %q", got, want)
	}
	if got, want := envelope.DeliveryDeadlineAt, deadline.UTC().Format(time.RFC3339Nano); got != want {
		t.Fatalf("delivery_deadline_at = %q, want %q", got, want)
	}
	if len(envelope.CallbackEndpoint.Handles) != 2 {
		t.Fatalf("callback handles = %d, want 2", len(envelope.CallbackEndpoint.Handles))
	}
	if interactions.ensureBindingParams.AdapterKind != interactionRecipientProviderTelegram {
		t.Fatalf("binding adapter kind = %q, want %q", interactions.ensureBindingParams.AdapterKind, interactionRecipientProviderTelegram)
	}
	if len(interactions.upsertHandleParams.Items) != 2 {
		t.Fatalf("upserted handle items = %d, want 2", len(interactions.upsertHandleParams.Items))
	}

	var content interactionDecisionContent
	if err := json.Unmarshal(envelope.Content, &content); err != nil {
		t.Fatalf("unmarshal decision content: %v", err)
	}
	if len(content.Options) != 2 {
		t.Fatalf("decision options = %d, want 2", len(content.Options))
	}
	for idx, option := range content.Options {
		if strings.TrimSpace(option.CallbackHandle) == "" {
			t.Fatalf("option %d callback handle is empty", idx)
		}
	}
}
