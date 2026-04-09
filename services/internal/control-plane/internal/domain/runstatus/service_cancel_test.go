package runstatus

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/codex-k8s/kodex/libs/go/crypto/tokencrypt"
	floweventdomain "github.com/codex-k8s/kodex/libs/go/domain/flowevent"
	mcpdomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/mcp"
	agentrunrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/agentrun"
	agentsessionrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/agentsession"
	githubratelimitwaitrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/githubratelimitwait"
	platformtokenrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/platformtoken"
	runtimedeploydomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/runtimedeploy"
	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

func TestCancelRun_CancelsArtifactsAndUpdatesComment(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tokenCrypt, err := tokencrypt.NewService("00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff")
	if err != nil {
		t.Fatalf("tokencrypt.NewService: %v", err)
	}
	botTokenEncrypted, err := tokenCrypt.EncryptString("bot-token")
	if err != nil {
		t.Fatalf("EncryptString: %v", err)
	}

	runPayload, err := json.Marshal(querytypes.RunPayload{
		Repository: querytypes.RunPayloadRepository{FullName: "codex-k8s/kodex"},
		Issue:      &querytypes.RunPayloadIssue{Number: 514, HTMLURL: "https://github.com/codex-k8s/kodex/issues/514"},
		Trigger: &querytypes.RunPayloadTrigger{
			Kind:  triggerKindDev,
			Label: "run:dev",
		},
		Runtime: &querytypes.RunPayloadRuntime{
			Mode:      runtimeModeFullEnv,
			Namespace: "run-514",
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal(runPayload): %v", err)
	}

	commentBody := testRunStatusCommentBody(t, commentState{
		RunID:        "run-514",
		Phase:        PhaseStarted,
		RuntimeMode:  runtimeModeFullEnv,
		Namespace:    "run-514",
		JobName:      "run-run-514",
		JobNamespace: "run-514",
		PromptLocale: localeRU,
		RunStatus:    "waiting_backpressure",
	})

	runs := &cancelRunTestRunsRepository{
		run: cancelRunTestRun{
			ID:            "run-514",
			CorrelationID: "corr-514",
			Status:        "waiting_backpressure",
			RunPayload:    runPayload,
		},
	}
	waits := &cancelRunTestWaitRepository{
		items: []githubratelimitwaitrepo.Wait{
			{
				ID:               "wait-1",
				RunID:            "run-514",
				CorrelationID:    "corr-514",
				State:            enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled,
				OperationClass:   enumtypes.GitHubRateLimitOperationClassAgentGitHubCall,
				LimitKind:        enumtypes.GitHubRateLimitLimitKindSecondary,
				Confidence:       enumtypes.GitHubRateLimitConfidenceConservative,
				RecoveryHintKind: enumtypes.GitHubRateLimitRecoveryHintKindRetryAfter,
				SignalID:         "sig-1",
				LastSignalAt:     time.Date(2026, 3, 16, 8, 0, 0, 0, time.UTC),
			},
		},
	}
	sessions := &cancelRunTestSessionsRepository{}
	k8s := &cancelRunTestKubernetesClient{
		jobExists: true,
	}
	runtimeDeploy := &cancelRunTestRuntimeDeployController{}
	events := &cancelRunTestFlowEvents{}
	github := &runstatusTestGitHub{
		listIssueCommentsFunc: func(context.Context, mcpdomain.GitHubListIssueCommentsParams) ([]mcpdomain.GitHubIssueComment, error) {
			return []mcpdomain.GitHubIssueComment{{
				ID:   99,
				Body: commentBody,
				URL:  "https://github.com/codex-k8s/kodex/issues/514#issuecomment-99",
			}}, nil
		},
		editIssueCommentFunc: func(_ context.Context, params mcpdomain.GitHubEditIssueCommentParams) (mcpdomain.GitHubIssueComment, error) {
			return mcpdomain.GitHubIssueComment{
				ID:   params.CommentID,
				Body: params.Body,
				URL:  "https://github.com/codex-k8s/kodex/issues/514#issuecomment-99",
			}, nil
		},
	}

	service, err := NewService(Config{
		PublicBaseURL: "https://platform.kodex.works",
		DefaultLocale: localeRU,
	}, Dependencies{
		Runs:                 runs,
		Sessions:             sessions,
		Platform:             &runstatusTestPlatformTokenRepository{item: platformtokenrepo.PlatformGitHubTokens{BotTokenEncrypted: botTokenEncrypted}},
		TokenCrypt:           tokenCrypt,
		GitHub:               github,
		Kubernetes:           k8s,
		FlowEvents:           events,
		StaffRuns:            &runstatusTestStaffRunsRepository{},
		GitHubRateLimitWaits: waits,
		RuntimeDeploy:        runtimeDeploy,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	result, err := service.CancelRun(ctx, CancelRunParams{
		RunID:             "run-514",
		Reason:            "manual stop",
		RequestedByType:   RequestedByTypeStaffUser,
		RequestedByID:     "user-1",
		RequestedByEmail:  "dev@example.com",
		RequestedByGitHub: "dev-user",
	})
	if err != nil {
		t.Fatalf("CancelRun() error = %v", err)
	}

	if !result.RuntimeDeployCancelRequested {
		t.Fatal("expected runtime deploy cancel request to be recorded")
	}
	if !result.JobStopped {
		t.Fatal("expected active run job to be deleted")
	}
	if result.CanceledGitHubWaits != 1 {
		t.Fatalf("canceled github waits = %d, want 1", result.CanceledGitHubWaits)
	}
	if got, want := result.CurrentStatus, runStatusCanceled; got != want {
		t.Fatalf("current status = %q, want %q", got, want)
	}
	if runs.run.Status != runStatusCanceled {
		t.Fatalf("stored run status = %q, want %q", runs.run.Status, runStatusCanceled)
	}
	if runtimeDeploy.calls != 1 {
		t.Fatalf("runtime deploy calls = %d, want 1", runtimeDeploy.calls)
	}
	if k8s.deleteJobCalls != 1 {
		t.Fatalf("delete job calls = %d, want 1", k8s.deleteJobCalls)
	}
	if sessions.calls != 1 {
		t.Fatalf("session wait-state calls = %d, want 1", sessions.calls)
	}
	if sessions.last.WaitState != string(enumtypes.AgentSessionWaitStateNone) {
		t.Fatalf("session wait_state = %q, want empty", sessions.last.WaitState)
	}
	if waits.items[0].State != enumtypes.GitHubRateLimitWaitStateCancelled {
		t.Fatalf("wait state = %q, want %q", waits.items[0].State, enumtypes.GitHubRateLimitWaitStateCancelled)
	}
	if waits.refreshCalls != 1 {
		t.Fatalf("refresh projection calls = %d, want 1", waits.refreshCalls)
	}
	foundCanceledEvent := false
	for _, item := range events.inserted {
		if item.EventType == floweventdomain.EventTypeRunCanceled {
			foundCanceledEvent = true
			break
		}
	}
	if !foundCanceledEvent {
		t.Fatalf("expected run.canceled event, got %#v", events.inserted)
	}

	state, ok := extractStateMarker(github.lastEditedBody)
	if !ok {
		t.Fatalf("edited comment body does not contain state marker: %q", github.lastEditedBody)
	}
	if state.RunStatus != runStatusCanceled {
		t.Fatalf("comment run_status = %q, want %q", state.RunStatus, runStatusCanceled)
	}
}

type cancelRunTestRun struct {
	ID            string
	CorrelationID string
	Status        string
	RunPayload    json.RawMessage
}

type cancelRunTestRunsRepository struct {
	run cancelRunTestRun
}

func (r *cancelRunTestRunsRepository) CreatePendingIfAbsent(context.Context, agentrunrepo.CreateParams) (agentrunrepo.CreateResult, error) {
	return agentrunrepo.CreateResult{}, nil
}

func (r *cancelRunTestRunsRepository) GetByID(context.Context, string) (agentrunrepo.Run, bool, error) {
	return agentrunrepo.Run{
		ID:            r.run.ID,
		CorrelationID: r.run.CorrelationID,
		Status:        r.run.Status,
		RunPayload:    r.run.RunPayload,
	}, true, nil
}

func (r *cancelRunTestRunsRepository) CancelActiveByID(context.Context, string) (bool, error) {
	r.run.Status = runStatusCanceled
	return true, nil
}

func (r *cancelRunTestRunsRepository) ListRecentByProject(context.Context, string, string, int, int) ([]agentrunrepo.RunLookupItem, error) {
	return nil, nil
}

func (r *cancelRunTestRunsRepository) SearchRecentByProjectIssueOrPullRequest(context.Context, string, string, int64, int64, int) ([]agentrunrepo.RunLookupItem, error) {
	return nil, nil
}

func (r *cancelRunTestRunsRepository) ListRunIDsByRepositoryIssue(context.Context, string, int64, int) ([]string, error) {
	return nil, nil
}

func (r *cancelRunTestRunsRepository) ListRunIDsByRepositoryPullRequest(context.Context, string, int64, int) ([]string, error) {
	return nil, nil
}

func (r *cancelRunTestRunsRepository) SetWaitContext(context.Context, agentrunrepo.SetWaitContextParams) (bool, error) {
	return false, nil
}

func (r *cancelRunTestRunsRepository) ClearWaitContextIfMatches(context.Context, agentrunrepo.ClearWaitContextParams) (bool, error) {
	return false, nil
}

type cancelRunTestSessionsRepository struct {
	calls int
	last  agentsessionrepo.SetWaitStateParams
}

func (r *cancelRunTestSessionsRepository) Upsert(context.Context, agentsessionrepo.UpsertParams) (valuetypes.AgentSessionSnapshotState, error) {
	return valuetypes.AgentSessionSnapshotState{}, nil
}

func (r *cancelRunTestSessionsRepository) SetWaitStateByRunID(_ context.Context, params agentsessionrepo.SetWaitStateParams) (bool, error) {
	r.calls++
	r.last = params
	return true, nil
}

func (r *cancelRunTestSessionsRepository) GetByRunID(context.Context, string) (agentsessionrepo.Session, bool, error) {
	return entitytypes.AgentSession{}, false, nil
}

func (r *cancelRunTestSessionsRepository) GetLatestByRepositoryBranchAndAgent(context.Context, string, string, string) (agentsessionrepo.Session, bool, error) {
	return entitytypes.AgentSession{}, false, nil
}

func (r *cancelRunTestSessionsRepository) CleanupSessionPayloadsFinishedBefore(context.Context, time.Time) (int64, error) {
	return 0, nil
}

type cancelRunTestWaitRepository struct {
	items        []githubratelimitwaitrepo.Wait
	refreshCalls int
}

func (r *cancelRunTestWaitRepository) Create(context.Context, githubratelimitwaitrepo.CreateWaitParams) (githubratelimitwaitrepo.Wait, error) {
	return githubratelimitwaitrepo.Wait{}, nil
}

func (r *cancelRunTestWaitRepository) Update(_ context.Context, params githubratelimitwaitrepo.UpdateWaitParams) (githubratelimitwaitrepo.Wait, bool, error) {
	for idx := range r.items {
		if r.items[idx].ID != params.ID {
			continue
		}
		r.items[idx].State = params.State
		r.items[idx].ResolvedAt = params.ResolvedAt
		r.items[idx].LastSignalAt = params.LastSignalAt
		r.items[idx].LastResumeAttemptAt = params.LastResumeAttemptAt
		return r.items[idx], true, nil
	}
	return githubratelimitwaitrepo.Wait{}, false, nil
}

func (r *cancelRunTestWaitRepository) GetByID(context.Context, string) (githubratelimitwaitrepo.Wait, bool, error) {
	return githubratelimitwaitrepo.Wait{}, false, nil
}

func (r *cancelRunTestWaitRepository) GetBySignalID(context.Context, string) (githubratelimitwaitrepo.Wait, bool, error) {
	return githubratelimitwaitrepo.Wait{}, false, nil
}

func (r *cancelRunTestWaitRepository) GetOpenByRunAndContour(context.Context, string, enumtypes.GitHubRateLimitContourKind) (githubratelimitwaitrepo.Wait, bool, error) {
	return githubratelimitwaitrepo.Wait{}, false, nil
}

func (r *cancelRunTestWaitRepository) ListByRunID(context.Context, string) ([]githubratelimitwaitrepo.Wait, error) {
	return append([]githubratelimitwaitrepo.Wait(nil), r.items...), nil
}

func (r *cancelRunTestWaitRepository) ClaimNextDueAutoResume(context.Context, time.Time, time.Time) (githubratelimitwaitrepo.Wait, bool, error) {
	return githubratelimitwaitrepo.Wait{}, false, nil
}

func (r *cancelRunTestWaitRepository) AppendEvidence(context.Context, githubratelimitwaitrepo.CreateEvidenceParams) (githubratelimitwaitrepo.Evidence, error) {
	return githubratelimitwaitrepo.Evidence{}, nil
}

func (r *cancelRunTestWaitRepository) RefreshRunProjection(context.Context, string) (valuetypes.GitHubRateLimitProjectionRefreshResult, error) {
	r.refreshCalls++
	return valuetypes.GitHubRateLimitProjectionRefreshResult{}, nil
}

type cancelRunTestKubernetesClient struct {
	jobExists      bool
	deleteJobCalls int
}

func (k *cancelRunTestKubernetesClient) DeleteManagedRunNamespace(context.Context, string) (bool, error) {
	return false, nil
}

func (k *cancelRunTestKubernetesClient) NamespaceExists(context.Context, string) (bool, error) {
	return true, nil
}

func (k *cancelRunTestKubernetesClient) JobExists(context.Context, string, string) (bool, error) {
	return k.jobExists, nil
}

func (k *cancelRunTestKubernetesClient) DeleteJobIfExists(context.Context, string, string) error {
	k.deleteJobCalls++
	k.jobExists = false
	return nil
}

func (k *cancelRunTestKubernetesClient) FindManagedRunNamespaceByRunID(context.Context, string) (string, bool, error) {
	return "", false, nil
}

type cancelRunTestRuntimeDeployController struct {
	calls int
}

func (c *cancelRunTestRuntimeDeployController) RequestTaskAction(context.Context, runtimedeploydomain.TaskActionParams) (runtimedeploydomain.TaskActionResult, error) {
	c.calls++
	return runtimedeploydomain.TaskActionResult{
		RunID:           "run-514",
		Action:          runtimedeploydomain.TaskActionCancel,
		PreviousStatus:  entitytypes.RuntimeDeployTaskStatusRunning,
		CurrentStatus:   entitytypes.RuntimeDeployTaskStatusCanceled,
		AlreadyTerminal: false,
	}, nil
}

type cancelRunTestFlowEvents struct {
	inserted []floweventdomain.InsertParams
}

func (r *cancelRunTestFlowEvents) Insert(_ context.Context, params floweventdomain.InsertParams) error {
	r.inserted = append(r.inserted, params)
	return nil
}
