package query

import (
	"time"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

// MissionControlContinuityGapListFilter defines one continuity-gap lookup.
type MissionControlContinuityGapListFilter struct {
	ProjectID        string
	SubjectEntityIDs []int64
	Statuses         []enumtypes.MissionControlGapStatus
}

// MissionControlWorkspaceRefreshParams triggers one owner-owned workspace continuity refresh.
type MissionControlWorkspaceRefreshParams struct {
	ProjectID     string
	CorrelationID string
	ObservedAt    time.Time
}

// MissionControlWorkspaceQuery defines one read-only graph workspace lookup.
type MissionControlWorkspaceQuery struct {
	ProjectID   string
	StatePreset enumtypes.MissionControlWorkspaceStatePreset
	Search      string
	RootLimit   int
}

// MissionControlLaunchPreviewParams defines one read-only launch preview request.
type MissionControlLaunchPreviewParams struct {
	ProjectID                 string
	NodeKind                  enumtypes.MissionControlEntityKind
	NodePublicID              string
	ThreadKind                string
	ThreadNumber              int
	TargetLabel               string
	RemovedLabels             []string
	ExpectedProjectionVersion int64
}
