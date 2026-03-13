package mcp

import (
	"context"
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
	lastSetWaitContext      agentrunrepo.SetWaitContextParams
	lastClearWaitContext    agentrunrepo.ClearWaitContextParams
}

func (r *interactionTestRunsRepository) CreatePendingIfAbsent(context.Context, agentrunrepo.CreateParams) (agentrunrepo.CreateResult, error) {
	return agentrunrepo.CreateResult{}, nil
}

func (r *interactionTestRunsRepository) GetByID(context.Context, string) (agentrunrepo.Run, bool, error) {
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
