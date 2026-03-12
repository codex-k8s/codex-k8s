package agentcallback

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/crypto/tokencrypt"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	repoprovider "github.com/codex-k8s/codex-k8s/libs/go/repo/provider"
	agentrunlogrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrunlog"
	agentsessionrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentsession"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
)

// Session mirrors persisted agent session snapshot entity.
type Session = agentsessionrepo.Session

// UpsertAgentSessionParams describes persistence payload for run session snapshot.
type UpsertAgentSessionParams = agentsessionrepo.UpsertParams

// UpsertAgentSessionResult returns persisted snapshot version metadata.
type UpsertAgentSessionResult = agentsessionrepo.UpsertResult

// GetLatestAgentSessionQuery describes latest snapshot lookup identity.
type GetLatestAgentSessionQuery struct {
	RepositoryFullName string
	BranchName         string
	AgentKey           string
}

// InsertRunFlowEventParams describes callback event persisted by agent-runner.
type InsertRunFlowEventParams struct {
	CorrelationID string
	EventType     floweventdomain.EventType
	Payload       json.RawMessage
	CreatedAt     time.Time
}

// LookupPullRequestQuery describes provider-neutral pull request recovery lookup.
type LookupPullRequestQuery struct {
	ProjectID          string
	RepositoryFullName string
	PullRequestNumber  int
	HeadBranch         string
}

// PullRequestLookupResult contains recovered pull request metadata.
type PullRequestLookupResult struct {
	Number int
	URL    string
	State  string
	Head   string
	Base   string
}

// Service encapsulates agent callback domain operations.
type Service struct {
	sessions   agentsessionrepo.Repository
	flowEvents floweventrepo.Repository
	runLogs    agentrunlogrepo.Repository
	repos      repocfgrepo.Repository
	tokenCrypt *tokencrypt.Service
	providers  map[string]repoprovider.RepositoryProvider
}

// NewService constructs callback domain service.
func NewService(
	sessions agentsessionrepo.Repository,
	flowEvents floweventrepo.Repository,
	runLogs agentrunlogrepo.Repository,
	repos repocfgrepo.Repository,
	tokenCrypt *tokencrypt.Service,
	providers map[repoprovider.Provider]repoprovider.RepositoryProvider,
) *Service {
	normalizedProviders := make(map[string]repoprovider.RepositoryProvider, len(providers))
	for providerID, providerClient := range providers {
		if providerClient == nil {
			continue
		}
		normalizedProviders[strings.TrimSpace(string(providerID))] = providerClient
	}

	return &Service{
		sessions:   sessions,
		flowEvents: flowEvents,
		runLogs:    runLogs,
		repos:      repos,
		tokenCrypt: tokenCrypt,
		providers:  normalizedProviders,
	}
}

// UpsertAgentSession stores or updates run session snapshot.
func (s *Service) UpsertAgentSession(ctx context.Context, params UpsertAgentSessionParams) (UpsertAgentSessionResult, error) {
	if s == nil || s.sessions == nil {
		return UpsertAgentSessionResult{}, errors.New("agent session repository is not configured")
	}
	result, err := s.sessions.Upsert(ctx, params)
	if err != nil {
		return UpsertAgentSessionResult{}, err
	}
	if s.runLogs != nil {
		if err := s.runLogs.UpsertRunAgentLogs(ctx, strings.TrimSpace(params.RunID), params.SessionJSON); err != nil {
			return UpsertAgentSessionResult{}, err
		}
	}
	return result, nil
}

// GetLatestAgentSession returns latest persisted snapshot by repo/branch/agent.
func (s *Service) GetLatestAgentSession(ctx context.Context, query GetLatestAgentSessionQuery) (Session, bool, error) {
	if s == nil || s.sessions == nil {
		return Session{}, false, errors.New("agent session repository is not configured")
	}
	return s.sessions.GetLatestByRepositoryBranchAndAgent(
		ctx,
		strings.TrimSpace(query.RepositoryFullName),
		strings.TrimSpace(query.BranchName),
		strings.TrimSpace(query.AgentKey),
	)
}

// InsertRunFlowEvent persists one run-bound callback event.
func (s *Service) InsertRunFlowEvent(ctx context.Context, params InsertRunFlowEventParams) error {
	if s == nil || s.flowEvents == nil {
		return errors.New("flow event repository is not configured")
	}
	payload := params.Payload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	createdAt := params.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: strings.TrimSpace(params.CorrelationID),
		ActorType:     floweventdomain.ActorTypeAgent,
		ActorID:       floweventdomain.ActorIDAgentRunner,
		EventType:     params.EventType,
		Payload:       payload,
		CreatedAt:     createdAt,
	})
}

// LookupPullRequest recovers pull request metadata for one run-bound repository.
func (s *Service) LookupPullRequest(ctx context.Context, query LookupPullRequestQuery) (PullRequestLookupResult, bool, error) {
	if s == nil {
		return PullRequestLookupResult{}, false, errors.New("agent callback service is not configured")
	}
	if s.repos == nil {
		return PullRequestLookupResult{}, false, errors.New("repository binding repository is not configured")
	}
	if s.tokenCrypt == nil {
		return PullRequestLookupResult{}, false, errors.New("token crypto service is not configured")
	}

	projectID := strings.TrimSpace(query.ProjectID)
	if projectID == "" {
		return PullRequestLookupResult{}, false, errors.New("project_id is required")
	}
	repositoryFullName := strings.TrimSpace(query.RepositoryFullName)
	if repositoryFullName == "" {
		return PullRequestLookupResult{}, false, errors.New("repository_full_name is required")
	}
	if query.PullRequestNumber <= 0 && strings.TrimSpace(query.HeadBranch) == "" {
		return PullRequestLookupResult{}, false, errors.New("pull_request_number or head_branch is required")
	}

	repositoryBinding, found, err := s.findProjectRepositoryByFullName(ctx, projectID, repositoryFullName)
	if err != nil {
		return PullRequestLookupResult{}, false, err
	}
	if !found {
		return PullRequestLookupResult{}, false, nil
	}

	providerClient, found := s.providers[strings.TrimSpace(repositoryBinding.Provider)]
	if !found || providerClient == nil {
		return PullRequestLookupResult{}, false, fmt.Errorf("repository provider %q is not configured", strings.TrimSpace(repositoryBinding.Provider))
	}

	encryptedToken, found, err := s.repos.GetTokenEncrypted(ctx, repositoryBinding.ID)
	if err != nil {
		return PullRequestLookupResult{}, false, fmt.Errorf("get repository token: %w", err)
	}
	if !found || len(encryptedToken) == 0 {
		return PullRequestLookupResult{}, false, nil
	}

	token, err := s.tokenCrypt.DecryptString(encryptedToken)
	if err != nil {
		return PullRequestLookupResult{}, false, fmt.Errorf("decrypt repository token: %w", err)
	}

	if query.PullRequestNumber > 0 {
		item, found, err := providerClient.GetPullRequest(
			ctx,
			token,
			strings.TrimSpace(repositoryBinding.Owner),
			strings.TrimSpace(repositoryBinding.Name),
			query.PullRequestNumber,
		)
		if err != nil {
			return PullRequestLookupResult{}, false, err
		}
		if !found {
			return PullRequestLookupResult{}, false, nil
		}
		return pullRequestLookupResultFromProvider(item), true, nil
	}

	item, found, err := providerClient.FindPullRequestByHead(
		ctx,
		token,
		strings.TrimSpace(repositoryBinding.Owner),
		strings.TrimSpace(repositoryBinding.Name),
		strings.TrimSpace(query.HeadBranch),
	)
	if err != nil {
		return PullRequestLookupResult{}, false, err
	}
	if !found {
		return PullRequestLookupResult{}, false, nil
	}
	return pullRequestLookupResultFromProvider(item), true, nil
}

// CleanupRunAgentLogs clears stored run logs for finished runs older than cutoff.
func (s *Service) CleanupRunAgentLogs(ctx context.Context, finishedBefore time.Time) (int64, error) {
	if s == nil || s.runLogs == nil {
		return 0, errors.New("run logs repository is not configured")
	}
	return runCleanupWithCutoff(finishedBefore, func(cutoff time.Time) (int64, error) {
		return s.runLogs.CleanupRunAgentLogsFinishedBefore(ctx, cutoff)
	})
}

// CleanupSessionPayloads clears heavy session payloads for finished runs older than cutoff.
func (s *Service) CleanupSessionPayloads(ctx context.Context, finishedBefore time.Time) (int64, error) {
	if s == nil {
		return 0, errors.New("agent callback service is not configured")
	}
	if s.sessions == nil {
		return 0, errors.New("agent session repository is not configured")
	}
	cutoff := finishedBefore.UTC()
	if cutoff.IsZero() {
		return 0, errors.New("finished_before is required")
	}
	cleared, err := s.sessions.CleanupSessionPayloadsFinishedBefore(ctx, cutoff)
	if err != nil {
		return 0, err
	}
	return cleared, nil
}

func runCleanupWithCutoff(finishedBefore time.Time, execute func(cutoff time.Time) (int64, error)) (int64, error) {
	cutoff := finishedBefore.UTC()
	if cutoff.IsZero() {
		return 0, errors.New("finished_before is required")
	}
	return execute(cutoff)
}

func (s *Service) findProjectRepositoryByFullName(ctx context.Context, projectID string, repositoryFullName string) (repocfgrepo.RepositoryBinding, bool, error) {
	items, err := s.repos.ListForProject(ctx, projectID, 500)
	if err != nil {
		return repocfgrepo.RepositoryBinding{}, false, fmt.Errorf("list project repositories: %w", err)
	}
	for _, item := range items {
		fullName := strings.TrimSpace(item.Owner) + "/" + strings.TrimSpace(item.Name)
		if strings.EqualFold(fullName, strings.TrimSpace(repositoryFullName)) {
			return item, true, nil
		}
	}
	return repocfgrepo.RepositoryBinding{}, false, nil
}

func pullRequestLookupResultFromProvider(item repoprovider.PullRequestInfo) PullRequestLookupResult {
	return PullRequestLookupResult{
		Number: item.Number,
		URL:    strings.TrimSpace(item.URL),
		State:  strings.TrimSpace(item.State),
		Head:   strings.TrimSpace(item.Head),
		Base:   strings.TrimSpace(item.Base),
	}
}
