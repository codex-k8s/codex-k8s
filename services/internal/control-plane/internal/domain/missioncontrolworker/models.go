package missioncontrolworker

import (
	"encoding/json"
	"time"

	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/missioncontrol"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	missioncontrolrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/missioncontrol"
	projectrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/project"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	staffrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/staffrun"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// Config controls worker-facing Mission Control orchestration helpers.
type Config struct {
	ProjectLimit        int
	RunLimit            int
	TimelineEventLimit  int
	StaleAfter          time.Duration
	DefaultTimelineText string
}

// Dependencies wires worker-facing Mission Control helpers.
type Dependencies struct {
	Projects       projectrepo.Repository
	Repositories   repocfgrepo.Repository
	AgentRuns      agentrunrepo.Repository
	StaffRuns      staffrunrepo.Repository
	MissionControl missioncontrol.DomainService
	Projection     missioncontrolrepo.Repository
}

// WarmupProject identifies one project eligible for Mission Control warmup.
type WarmupProject struct {
	ProjectID          string
	ProjectName        string
	RepositoryFullName string
}

// WarmupRequest triggers one warmup/backfill execution for a project.
type WarmupRequest = missioncontrol.WarmupRequest

// WarmupResult captures backfill diagnostics together with the final summary.
type WarmupResult struct {
	Summary             missioncontrol.WarmupSummary
	BackfilledEntities  int
	BackfilledRelations int
	BackfilledTimelines int
}

// PendingStageNextStep describes the currently executable stage transition payload.
type PendingStageNextStep struct {
	ThreadKind  string
	ThreadNo    int
	TargetLabel string
}

// PendingCommand is one worker-owned Mission Control execution candidate.
type PendingCommand struct {
	ProjectID            string
	CommandID            string
	CommandKind          enumtypes.MissionControlCommandKind
	EffectiveCommandKind enumtypes.MissionControlCommandKind
	Status               enumtypes.MissionControlCommandStatus
	CorrelationID        string
	BusinessIntentKey    string
	RepositoryFullName   string
	RetryTargetCommandID string
	StageNextStep        *PendingStageNextStep
	RequestedAt          time.Time
	UpdatedAt            time.Time
}

type workItemCardPayload struct {
	RepositoryFullName string `json:"repository_full_name"`
	IssueNumber        int64  `json:"issue_number"`
	IssueURL           string `json:"issue_url,omitempty"`
	LastRunID          string `json:"last_run_id,omitempty"`
	LastStatus         string `json:"last_status,omitempty"`
	TriggerKind        string `json:"trigger_kind,omitempty"`
}

type pullRequestCardPayload struct {
	RepositoryFullName string `json:"repository_full_name"`
	PullRequestNumber  int64  `json:"pull_request_number"`
	PullRequestURL     string `json:"pull_request_url,omitempty"`
	LastRunID          string `json:"last_run_id,omitempty"`
	LastStatus         string `json:"last_status,omitempty"`
}

type agentCardPayload struct {
	AgentKey    string `json:"agent_key"`
	LastRunID   string `json:"last_run_id,omitempty"`
	LastStatus  string `json:"last_status,omitempty"`
	LastRunRepo string `json:"last_run_repository,omitempty"`
}

type timelinePayload struct {
	RunID          string          `json:"run_id,omitempty"`
	CorrelationID  string          `json:"correlation_id"`
	EventType      string          `json:"event_type"`
	EventPayload   json.RawMessage `json:"event_payload,omitempty"`
	RepositoryFull string          `json:"repository_full_name,omitempty"`
}

type projectionSeed struct {
	projectID         string
	entityKind        enumtypes.MissionControlEntityKind
	entityExternalKey string
	providerKind      enumtypes.MissionControlProviderKind
	providerURL       string
	title             string
	activeState       enumtypes.MissionControlActiveState
	syncStatus        enumtypes.MissionControlSyncStatus
	projectionVersion int64
	cardPayloadJSON   json.RawMessage
	detailPayloadJSON json.RawMessage
	providerUpdatedAt *time.Time
	projectedAt       time.Time
	staleAfter        *time.Time
}

type relationSeed struct {
	sourceEntityKey string
	targetEntityKey string
	relationKind    enumtypes.MissionControlRelationKind
}

type timelineSeed struct {
	entityKey         string
	entryExternalKey  string
	summary           string
	payloadJSON       json.RawMessage
	occurredAt        time.Time
	providerURL       string
	repositoryFullRef string
}
