package agentcallback

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	agentsessionrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentsession"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
)

// Session mirrors persisted agent session snapshot entity.
type Session = agentsessionrepo.Session

// UpsertAgentSessionParams describes persistence payload for run session snapshot.
type UpsertAgentSessionParams = agentsessionrepo.UpsertParams

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

// Service encapsulates agent callback domain operations.
type Service struct {
	sessions   agentsessionrepo.Repository
	flowEvents floweventrepo.Repository
}

// NewService constructs callback domain service.
func NewService(sessions agentsessionrepo.Repository, flowEvents floweventrepo.Repository) *Service {
	return &Service{sessions: sessions, flowEvents: flowEvents}
}

// UpsertAgentSession stores or updates run session snapshot.
func (s *Service) UpsertAgentSession(ctx context.Context, params UpsertAgentSessionParams) error {
	if s == nil || s.sessions == nil {
		return errors.New("agent session repository is not configured")
	}
	return s.sessions.Upsert(ctx, params)
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
