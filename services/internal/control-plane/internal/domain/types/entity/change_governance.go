package entity

import (
	"encoding/json"
	"time"

	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// ChangeGovernancePackage stores one canonical package root.
type ChangeGovernancePackage struct {
	ID                        string
	PackageKey                string
	ProjectID                 string
	RepositoryFullName        string
	IssueNumber               int
	PRNumber                  *int
	RiskTier                  enumtypes.ChangeGovernanceRiskTier
	BundleAdmissibility       enumtypes.ChangeGovernanceBundleAdmissibility
	PublicationState          enumtypes.ChangeGovernancePublicationState
	EvidenceCompletenessState enumtypes.ChangeGovernanceEvidenceCompletenessState
	VerificationMinimumState  enumtypes.ChangeGovernanceVerificationMinimumState
	WaiverState               enumtypes.ChangeGovernanceWaiverState
	ReleaseReadinessState     enumtypes.ChangeGovernanceReleaseReadinessState
	GovernanceFeedbackState   enumtypes.ChangeGovernanceFeedbackState
	ActiveProjectionVersion   int64
	LatestCorrelationID       string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

// ChangeGovernanceInternalDraft stores one hidden draft ledger row.
type ChangeGovernanceInternalDraft struct {
	ID            string
	PackageID     string
	RunID         string
	SignalID      string
	DraftRef      string
	DraftChecksum string
	DraftKind     enumtypes.ChangeGovernanceDraftKind
	MetadataJSON  json.RawMessage
	IsLatest      bool
	OccurredAt    time.Time
	CreatedAt     time.Time
}

// ChangeGovernanceWave stores one semantic wave.
type ChangeGovernanceWave struct {
	ID                        string
	PackageID                 string
	WaveKey                   string
	PublishOrder              int
	DominantIntent            enumtypes.ChangeGovernanceDominantIntent
	BoundedScopeKind          enumtypes.ChangeGovernanceBoundedScopeKind
	PublicationState          enumtypes.ChangeGovernanceWavePublicationState
	EvidenceCompletenessState enumtypes.ChangeGovernanceEvidenceCompletenessState
	VerificationMinimumState  enumtypes.ChangeGovernanceVerificationMinimumState
	Summary                   string
	VerificationTargetsJSON   json.RawMessage
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

// ChangeGovernanceEvidenceBlock stores one typed evidence block.
type ChangeGovernanceEvidenceBlock struct {
	ID                string
	PackageID         string
	WaveID            string
	BlockKind         enumtypes.ChangeGovernanceEvidenceBlockKind
	State             enumtypes.ChangeGovernanceEvidenceBlockState
	VerificationState enumtypes.ChangeGovernanceVerificationMinimumState
	RequiredByTier    bool
	SourceKind        enumtypes.ChangeGovernanceEvidenceSourceKind
	ArtifactLinksJSON json.RawMessage
	LatestSignalID    string
	ObservedAt        time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ChangeGovernanceDecisionRecord stores one append-only decision.
type ChangeGovernanceDecisionRecord struct {
	ID                  string
	PackageID           string
	ScopeKind           enumtypes.ChangeGovernanceDecisionScopeKind
	ScopeRef            string
	DecisionID          string
	DecisionKind        enumtypes.ChangeGovernanceDecisionKind
	State               enumtypes.ChangeGovernanceDecisionState
	ActorKind           enumtypes.ChangeGovernanceDecisionActorKind
	ResidualRiskTier    enumtypes.ChangeGovernanceRiskTier
	SummaryMarkdown     string
	DecisionPayloadJSON json.RawMessage
	RecordedAt          time.Time
	CreatedAt           time.Time
}

// ChangeGovernanceFeedbackRecord stores one append-only gap record.
type ChangeGovernanceFeedbackRecord struct {
	ID                 string
	PackageID          string
	FeedbackID         string
	GapKind            enumtypes.ChangeGovernanceFeedbackGapKind
	SourceKind         enumtypes.ChangeGovernanceFeedbackSourceKind
	Severity           enumtypes.ChangeGovernanceFeedbackSeverity
	State              enumtypes.ChangeGovernanceFeedbackRecordState
	SuggestedAction    enumtypes.ChangeGovernanceFeedbackSuggestedAction
	SummaryMarkdown    string
	RelatedArtifactRef string
	OpenedAt           time.Time
	ClosedAt           *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// ChangeGovernanceProjectionSnapshot stores one persisted projection payload.
type ChangeGovernanceProjectionSnapshot struct {
	ID                int64
	PackageID         string
	ProjectionKind    enumtypes.ChangeGovernanceProjectionKind
	ProjectionVersion int64
	IsCurrent         bool
	PayloadJSON       json.RawMessage
	RefreshedAt       time.Time
	CreatedAt         time.Time
}

// ChangeGovernanceArtifactLink stores one auditable artifact lineage row.
type ChangeGovernanceArtifactLink struct {
	ID           int64
	PackageID    string
	ArtifactKind enumtypes.ChangeGovernanceArtifactKind
	ArtifactRef  string
	RelationKind enumtypes.ChangeGovernanceArtifactRelationKind
	DisplayLabel string
	CreatedAt    time.Time
}
