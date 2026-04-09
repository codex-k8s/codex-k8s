package missioncontrolworker

import (
	"encoding/json"
	"time"

	"github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/missioncontrol"
	agentrunrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/agentrun"
	missioncontrolrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/missioncontrol"
	projectrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/project"
	repocfgrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/repocfg"
	staffrunrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/staffrun"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
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

type timelinePayload struct {
	RunID          string          `json:"run_id,omitempty"`
	CorrelationID  string          `json:"correlation_id"`
	EventType      string          `json:"event_type"`
	EventPayload   json.RawMessage `json:"event_payload,omitempty"`
	RepositoryFull string          `json:"repository_full_name,omitempty"`
}

type projectionSeed = missioncontrolrepo.UpsertEntityParams

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

type continuityGapSeed struct {
	subjectEntityKey string
	gapKind          enumtypes.MissionControlGapKind
	severity         enumtypes.MissionControlGapSeverity
	detectedAt       time.Time
}

type workspaceWatermarkSeed struct {
	watermarkKind   enumtypes.MissionControlWorkspaceWatermarkKind
	status          enumtypes.MissionControlWorkspaceWatermarkStatus
	summary         string
	windowStartedAt *time.Time
	windowEndedAt   *time.Time
	payloadJSON     json.RawMessage
}
