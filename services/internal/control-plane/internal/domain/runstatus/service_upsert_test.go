package runstatus

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/codex-k8s/codex-k8s/libs/go/crypto/tokencrypt"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	platformtokenrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/platformtoken"
	staffrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/staffrun"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func TestUpsertRunStatusComment_UsesTrackedCommentIDWhenListIsStale(t *testing.T) {
	t.Parallel()

	const (
		runID            = "run-dev-149"
		correlationID    = "corr-dev-149"
		trackedCommentID = int64(98765)
	)

	github := &stubRunStatusGitHub{
		comments: []mcpdomain.GitHubIssueComment{},
		editResult: mcpdomain.GitHubIssueComment{
			ID:  trackedCommentID,
			URL: "https://example.invalid/comments/98765",
		},
	}

	svc := newRunStatusServiceForUpsertTest(t, runStatusTestDeps{
		runID:         runID,
		correlationID: correlationID,
		staffEvents:   []entitytypes.StaffFlowEvent{buildTrackedStatusCommentEvent(t, correlationID, runID, trackedCommentID, "https://example.invalid/comments/98765")},
		github:        github,
	})

	result, err := svc.UpsertRunStatusComment(context.Background(), UpsertCommentParams{
		RunID:       runID,
		Phase:       PhaseCreated,
		RuntimeMode: runtimeModeFullEnv,
		Namespace:   "codex-k8s-dev-1",
		RunStatus:   "running",
	})
	if err != nil {
		t.Fatalf("UpsertRunStatusComment returned error: %v", err)
	}

	if got := len(github.editCalls); got != 1 {
		t.Fatalf("expected one EditIssueComment call, got %d", got)
	}
	if github.editCalls[0].CommentID != trackedCommentID {
		t.Fatalf("expected tracked comment id %d, got %d", trackedCommentID, github.editCalls[0].CommentID)
	}
	if got := len(github.createCalls); got != 0 {
		t.Fatalf("expected no CreateIssueComment calls, got %d", got)
	}
	if result.CommentID != trackedCommentID {
		t.Fatalf("expected result comment id %d, got %d", trackedCommentID, result.CommentID)
	}
}

func TestUpsertRunStatusComment_RecreatesCommentWhenTrackedEditReturnsNotFound(t *testing.T) {
	t.Parallel()

	const (
		runID            = "run-revise-149"
		correlationID    = "corr-revise-149"
		trackedCommentID = int64(12345)
		createdCommentID = int64(22222)
	)

	github := &stubRunStatusGitHub{
		comments:  []mcpdomain.GitHubIssueComment{},
		editErr:   errors.New("404 Not Found"),
		createID:  createdCommentID,
		createURL: "https://example.invalid/comments/22222",
	}

	svc := newRunStatusServiceForUpsertTest(t, runStatusTestDeps{
		runID:         runID,
		correlationID: correlationID,
		staffEvents:   []entitytypes.StaffFlowEvent{buildTrackedStatusCommentEvent(t, correlationID, runID, trackedCommentID, "https://example.invalid/comments/12345")},
		github:        github,
	})

	result, err := svc.UpsertRunStatusComment(context.Background(), UpsertCommentParams{
		RunID:       runID,
		Phase:       PhaseCreated,
		RuntimeMode: runtimeModeFullEnv,
		Namespace:   "codex-k8s-dev-1",
		RunStatus:   "running",
	})
	if err != nil {
		t.Fatalf("UpsertRunStatusComment returned error: %v", err)
	}

	if got := len(github.editCalls); got != 1 {
		t.Fatalf("expected one EditIssueComment call, got %d", got)
	}
	if github.editCalls[0].CommentID != trackedCommentID {
		t.Fatalf("expected tracked comment id %d, got %d", trackedCommentID, github.editCalls[0].CommentID)
	}
	if got := len(github.createCalls); got != 1 {
		t.Fatalf("expected one CreateIssueComment call, got %d", got)
	}
	if result.CommentID != createdCommentID {
		t.Fatalf("expected created comment id %d, got %d", createdCommentID, result.CommentID)
	}
}

type runStatusTestDeps struct {
	runID         string
	correlationID string
	staffEvents   []entitytypes.StaffFlowEvent
	github        *stubRunStatusGitHub
}

func newRunStatusServiceForUpsertTest(t *testing.T, deps runStatusTestDeps) *Service {
	t.Helper()

	crypt, err := tokencrypt.NewService("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("tokencrypt.NewService returned error: %v", err)
	}
	encryptedBotToken, err := crypt.EncryptString("bot-token")
	if err != nil {
		t.Fatalf("EncryptString returned error: %v", err)
	}

	runPayload := querytypes.RunPayload{
		Project: querytypes.RunPayloadProject{ID: "project-1"},
		Repository: querytypes.RunPayloadRepository{
			FullName: "codex-k8s/codex-k8s",
			Name:     "codex-k8s",
		},
		Issue: &querytypes.RunPayloadIssue{
			Number:  149,
			Title:   "issue-149",
			State:   "open",
			HTMLURL: "https://github.com/codex-k8s/codex-k8s/issues/149",
		},
		Trigger: &querytypes.RunPayloadTrigger{
			Source: triggerSourceIssueLabel,
			Label:  "run:dev",
			Kind:   triggerKindDev,
		},
	}

	runPayloadJSON, err := json.Marshal(runPayload)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	svc, err := NewService(Config{
		PublicBaseURL: "https://platform.codex-k8s.dev",
		DefaultLocale: "ru",
		AIDomain:      "ai.platform.codex-k8s.dev",
	}, Dependencies{
		Runs: &stubRunStatusRuns{
			run: agentrunrepo.Run{
				ID:            deps.runID,
				CorrelationID: deps.correlationID,
				Status:        "running",
				RunPayload:    runPayloadJSON,
			},
		},
		Platform:   &stubRunStatusPlatformTokens{tokens: platformtokenrepo.PlatformGitHubTokens{BotTokenEncrypted: encryptedBotToken}},
		TokenCrypt: crypt,
		GitHub:     deps.github,
		Kubernetes: &stubRunStatusKubernetes{},
		FlowEvents: &stubRunStatusFlowEvents{},
		StaffRuns:  &stubRunStatusStaffRuns{events: deps.staffEvents},
	})
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	return svc
}

func mustJSON(t *testing.T, payload any) []byte {
	t.Helper()

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	return raw
}

func buildTrackedStatusCommentEvent(t *testing.T, correlationID string, runID string, commentID int64, commentURL string) entitytypes.StaffFlowEvent {
	t.Helper()

	return entitytypes.StaffFlowEvent{
		CorrelationID: correlationID,
		EventType:     string(floweventdomain.EventTypeRunStatusCommentUpserted),
		PayloadJSON: mustJSON(t, runStatusCommentUpsertedPayload{
			RunID:              runID,
			CommentID:          commentID,
			CommentURL:         commentURL,
			IssueNumber:        149,
			RepositoryFullName: "codex-k8s/codex-k8s",
			Phase:              PhaseStarted,
		}),
	}
}

type stubRunStatusRuns struct {
	run agentrunrepo.Run
}

func (s *stubRunStatusRuns) CreatePendingIfAbsent(context.Context, querytypes.AgentRunCreateParams) (querytypes.AgentRunCreateResult, error) {
	return querytypes.AgentRunCreateResult{}, errors.New("not implemented")
}

func (s *stubRunStatusRuns) GetByID(context.Context, string) (agentrunrepo.Run, bool, error) {
	return s.run, true, nil
}

func (s *stubRunStatusRuns) ListRecentByProject(context.Context, string, string, int, int) ([]querytypes.AgentRunLookupItem, error) {
	return nil, nil
}

func (s *stubRunStatusRuns) SearchRecentByProjectIssueOrPullRequest(context.Context, string, string, int64, int64, int) ([]querytypes.AgentRunLookupItem, error) {
	return nil, nil
}

func (s *stubRunStatusRuns) ListRunIDsByRepositoryIssue(context.Context, string, int64, int) ([]string, error) {
	return nil, nil
}

func (s *stubRunStatusRuns) ListRunIDsByRepositoryPullRequest(context.Context, string, int64, int) ([]string, error) {
	return nil, nil
}

type stubRunStatusPlatformTokens struct {
	tokens platformtokenrepo.PlatformGitHubTokens
}

func (s *stubRunStatusPlatformTokens) Get(context.Context) (platformtokenrepo.PlatformGitHubTokens, bool, error) {
	return s.tokens, true, nil
}

func (s *stubRunStatusPlatformTokens) Upsert(context.Context, querytypes.PlatformGitHubTokensUpsertParams) (platformtokenrepo.PlatformGitHubTokens, error) {
	return s.tokens, nil
}

type stubRunStatusGitHub struct {
	comments   []mcpdomain.GitHubIssueComment
	editResult mcpdomain.GitHubIssueComment
	editErr    error
	createID   int64
	createURL  string

	editCalls   []mcpdomain.GitHubEditIssueCommentParams
	createCalls []mcpdomain.GitHubCreateIssueCommentParams
}

func (s *stubRunStatusGitHub) ListIssueComments(context.Context, mcpdomain.GitHubListIssueCommentsParams) ([]mcpdomain.GitHubIssueComment, error) {
	return append([]mcpdomain.GitHubIssueComment(nil), s.comments...), nil
}

func (s *stubRunStatusGitHub) CreateIssueComment(_ context.Context, params mcpdomain.GitHubCreateIssueCommentParams) (mcpdomain.GitHubIssueComment, error) {
	s.createCalls = append(s.createCalls, params)
	id := s.createID
	if id == 0 {
		id = 1
	}
	url := s.createURL
	if url == "" {
		url = "https://example.invalid/comments/1"
	}
	return mcpdomain.GitHubIssueComment{ID: id, URL: url, Body: params.Body}, nil
}

func (s *stubRunStatusGitHub) EditIssueComment(_ context.Context, params mcpdomain.GitHubEditIssueCommentParams) (mcpdomain.GitHubIssueComment, error) {
	s.editCalls = append(s.editCalls, params)
	if s.editErr != nil {
		return mcpdomain.GitHubIssueComment{}, s.editErr
	}
	if s.editResult.ID != 0 {
		return s.editResult, nil
	}
	return mcpdomain.GitHubIssueComment{ID: params.CommentID, URL: "https://example.invalid/comments/edited", Body: params.Body}, nil
}

func (s *stubRunStatusGitHub) ListIssueReactions(context.Context, mcpdomain.GitHubListIssueReactionsParams) ([]mcpdomain.GitHubIssueReaction, error) {
	return nil, nil
}

func (s *stubRunStatusGitHub) CreateIssueReaction(context.Context, mcpdomain.GitHubCreateIssueReactionParams) (mcpdomain.GitHubIssueReaction, error) {
	return mcpdomain.GitHubIssueReaction{}, nil
}

type stubRunStatusKubernetes struct{}

func (s *stubRunStatusKubernetes) DeleteManagedRunNamespace(context.Context, string) (bool, error) {
	return false, nil
}

func (s *stubRunStatusKubernetes) NamespaceExists(context.Context, string) (bool, error) {
	return false, nil
}

func (s *stubRunStatusKubernetes) JobExists(context.Context, string, string) (bool, error) {
	return false, nil
}

func (s *stubRunStatusKubernetes) FindManagedRunNamespaceByRunID(context.Context, string) (string, bool, error) {
	return "", false, nil
}

type stubRunStatusFlowEvents struct {
	inserted []floweventrepo.InsertParams
}

func (s *stubRunStatusFlowEvents) Insert(_ context.Context, params floweventrepo.InsertParams) error {
	s.inserted = append(s.inserted, params)
	return nil
}

type stubRunStatusStaffRuns struct {
	events []entitytypes.StaffFlowEvent
}

func (s *stubRunStatusStaffRuns) ListAll(context.Context, int) ([]staffrunrepo.Run, error) {
	return nil, nil
}

func (s *stubRunStatusStaffRuns) ListForUser(context.Context, string, int) ([]staffrunrepo.Run, error) {
	return nil, nil
}

func (s *stubRunStatusStaffRuns) ListJobsAll(context.Context, querytypes.StaffRunListFilter) ([]staffrunrepo.Run, error) {
	return nil, nil
}

func (s *stubRunStatusStaffRuns) ListJobsForUser(context.Context, string, querytypes.StaffRunListFilter) ([]staffrunrepo.Run, error) {
	return nil, nil
}

func (s *stubRunStatusStaffRuns) ListWaitsAll(context.Context, querytypes.StaffRunListFilter) ([]staffrunrepo.Run, error) {
	return nil, nil
}

func (s *stubRunStatusStaffRuns) ListWaitsForUser(context.Context, string, querytypes.StaffRunListFilter) ([]staffrunrepo.Run, error) {
	return nil, nil
}

func (s *stubRunStatusStaffRuns) GetByID(context.Context, string) (staffrunrepo.Run, bool, error) {
	return staffrunrepo.Run{}, false, nil
}

func (s *stubRunStatusStaffRuns) GetLogsByRunID(context.Context, string) (staffrunrepo.RunLogs, bool, error) {
	return staffrunrepo.RunLogs{}, false, nil
}

func (s *stubRunStatusStaffRuns) ListEventsByCorrelation(context.Context, string, int) ([]staffrunrepo.FlowEvent, error) {
	items := make([]staffrunrepo.FlowEvent, 0, len(s.events))
	for _, event := range s.events {
		items = append(items, staffrunrepo.FlowEvent(event))
	}
	return items, nil
}

func (s *stubRunStatusStaffRuns) DeleteFlowEventsByProjectID(context.Context, string) error {
	return nil
}

func (s *stubRunStatusStaffRuns) GetCorrelationByRunID(context.Context, string) (string, string, bool, error) {
	return "", "", false, nil
}

var _ agentrunrepo.Repository = (*stubRunStatusRuns)(nil)
var _ platformtokenrepo.Repository = (*stubRunStatusPlatformTokens)(nil)
var _ GitHubClient = (*stubRunStatusGitHub)(nil)
var _ KubernetesClient = (*stubRunStatusKubernetes)(nil)
var _ floweventrepo.Repository = (*stubRunStatusFlowEvents)(nil)
var _ staffrunrepo.Repository = (*stubRunStatusStaffRuns)(nil)
