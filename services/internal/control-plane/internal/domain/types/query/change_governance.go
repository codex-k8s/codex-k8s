package query

import (
	"time"

	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// ChangeGovernanceScopeHint captures one bounded-context or surface hint from runner.
type ChangeGovernanceScopeHint struct {
	ContextKey  string                                `json:"context_key"`
	SurfaceKind enumtypes.ChangeGovernanceSurfaceKind `json:"surface_kind"`
}

// ChangeGovernanceVerificationTarget captures one expected verification target for a wave.
type ChangeGovernanceVerificationTarget struct {
	TargetKind enumtypes.ChangeGovernanceVerificationTargetKind `json:"target_kind"`
	TargetRef  string                                           `json:"target_ref"`
}

// ChangeGovernanceWaveDraft captures one semantic wave from runner.
type ChangeGovernanceWaveDraft struct {
	WaveKey             string                                     `json:"wave_key"`
	PublishOrder        int                                        `json:"publish_order"`
	DominantIntent      enumtypes.ChangeGovernanceDominantIntent   `json:"dominant_intent"`
	BoundedScopeKind    enumtypes.ChangeGovernanceBoundedScopeKind `json:"bounded_scope_kind"`
	Summary             string                                     `json:"summary"`
	VerificationTargets []ChangeGovernanceVerificationTarget       `json:"verification_targets"`
}

// ChangeGovernanceArtifactLinkSeed captures one artifact lineage input.
type ChangeGovernanceArtifactLinkSeed struct {
	ArtifactKind enumtypes.ChangeGovernanceArtifactKind         `json:"artifact_kind"`
	ArtifactRef  string                                         `json:"artifact_ref"`
	RelationKind enumtypes.ChangeGovernanceArtifactRelationKind `json:"relation_kind"`
	DisplayLabel string                                         `json:"display_label"`
}

// ChangeGovernanceDraftSignalParams captures draft ingestion input.
type ChangeGovernanceDraftSignalParams struct {
	RunID                string
	SignalID             string
	CorrelationID        string
	ProjectID            string
	RepositoryFullName   string
	IssueNumber          int
	PRNumber             *int
	BranchName           string
	DraftRef             string
	DraftKind            enumtypes.ChangeGovernanceDraftKind
	ChangeScopeHints     []ChangeGovernanceScopeHint
	CandidateRiskDrivers []enumtypes.ChangeGovernanceRiskDriver
	DraftChecksum        string
	OccurredAt           time.Time
}

// ChangeGovernanceWaveMapParams captures wave-map publication input.
type ChangeGovernanceWaveMapParams struct {
	PackageID         string
	ExpectedProjectID string
	WaveMapID         string
	CorrelationID     string
	Waves             []ChangeGovernanceWaveDraft
	PublishedAt       time.Time
}

// ChangeGovernanceEvidenceSignalParams captures one evidence upsert signal.
type ChangeGovernanceEvidenceSignalParams struct {
	PackageID             string
	ExpectedProjectID     string
	SignalID              string
	CorrelationID         string
	ScopeKind             enumtypes.ChangeGovernanceEvidenceScopeKind
	ScopeRef              string
	BlockKind             enumtypes.ChangeGovernanceEvidenceBlockKind
	ArtifactLinks         []ChangeGovernanceArtifactLinkSeed
	VerificationStateHint enumtypes.ChangeGovernanceVerificationMinimumState
	RequiredByTier        bool
	OccurredAt            time.Time
}
