package dbmodel

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// PackageRow mirrors one change_governance_packages row.
type PackageRow struct {
	ID                        string      `db:"id"`
	PackageKey                string      `db:"package_key"`
	ProjectID                 string      `db:"project_id"`
	RepositoryFullName        string      `db:"repository_full_name"`
	IssueNumber               int32       `db:"issue_number"`
	PRNumber                  pgtype.Int4 `db:"pr_number"`
	RiskTier                  pgtype.Text `db:"risk_tier"`
	BundleAdmissibility       string      `db:"bundle_admissibility"`
	PublicationState          string      `db:"publication_state"`
	EvidenceCompletenessState string      `db:"evidence_completeness_state"`
	VerificationMinimumState  string      `db:"verification_minimum_state"`
	WaiverState               string      `db:"waiver_state"`
	ReleaseReadinessState     string      `db:"release_readiness_state"`
	GovernanceFeedbackState   string      `db:"governance_feedback_state"`
	ActiveProjectionVersion   int64       `db:"active_projection_version"`
	LatestCorrelationID       pgtype.Text `db:"latest_correlation_id"`
	CreatedAt                 time.Time   `db:"created_at"`
	UpdatedAt                 time.Time   `db:"updated_at"`
}

// InternalDraftRow mirrors one change_governance_internal_drafts row.
type InternalDraftRow struct {
	ID            string      `db:"id"`
	PackageID     string      `db:"package_id"`
	RunID         pgtype.Text `db:"run_id"`
	SignalID      string      `db:"signal_id"`
	DraftRef      string      `db:"draft_ref"`
	DraftChecksum pgtype.Text `db:"draft_checksum"`
	DraftKind     string      `db:"draft_kind"`
	MetadataJSON  []byte      `db:"metadata_json"`
	IsLatest      bool        `db:"is_latest"`
	OccurredAt    time.Time   `db:"occurred_at"`
	CreatedAt     time.Time   `db:"created_at"`
}

// WaveRow mirrors one change_governance_waves row.
type WaveRow struct {
	ID                        string    `db:"id"`
	PackageID                 string    `db:"package_id"`
	WaveKey                   string    `db:"wave_key"`
	PublishOrder              int32     `db:"publish_order"`
	DominantIntent            string    `db:"dominant_intent"`
	BoundedScopeKind          string    `db:"bounded_scope_kind"`
	PublicationState          string    `db:"publication_state"`
	EvidenceCompletenessState string    `db:"evidence_completeness_state"`
	VerificationMinimumState  string    `db:"verification_minimum_state"`
	Summary                   string    `db:"summary"`
	VerificationTargetsJSON   []byte    `db:"verification_targets_json"`
	CreatedAt                 time.Time `db:"created_at"`
	UpdatedAt                 time.Time `db:"updated_at"`
}

// EvidenceBlockRow mirrors one change_governance_evidence_blocks row.
type EvidenceBlockRow struct {
	ID                string      `db:"id"`
	PackageID         string      `db:"package_id"`
	WaveID            pgtype.Text `db:"wave_id"`
	BlockKind         string      `db:"block_kind"`
	State             string      `db:"state"`
	VerificationState string      `db:"verification_state"`
	RequiredByTier    bool        `db:"required_by_tier"`
	SourceKind        string      `db:"source_kind"`
	ArtifactLinksJSON []byte      `db:"artifact_links_json"`
	LatestSignalID    pgtype.Text `db:"latest_signal_id"`
	ObservedAt        time.Time   `db:"observed_at"`
	CreatedAt         time.Time   `db:"created_at"`
	UpdatedAt         time.Time   `db:"updated_at"`
}

// DecisionRecordRow mirrors one change_governance_decision_records row.
type DecisionRecordRow struct {
	ID                  string      `db:"id"`
	PackageID           string      `db:"package_id"`
	ScopeKind           string      `db:"scope_kind"`
	ScopeRef            string      `db:"scope_ref"`
	DecisionID          string      `db:"decision_id"`
	DecisionKind        string      `db:"decision_kind"`
	State               string      `db:"state"`
	ActorKind           string      `db:"actor_kind"`
	ResidualRiskTier    pgtype.Text `db:"residual_risk_tier"`
	SummaryMarkdown     string      `db:"summary_markdown"`
	DecisionPayloadJSON []byte      `db:"decision_payload_json"`
	RecordedAt          time.Time   `db:"recorded_at"`
	CreatedAt           time.Time   `db:"created_at"`
}

// FeedbackRecordRow mirrors one change_governance_feedback_records row.
type FeedbackRecordRow struct {
	ID                 string             `db:"id"`
	PackageID          string             `db:"package_id"`
	FeedbackID         string             `db:"feedback_id"`
	GapKind            string             `db:"gap_kind"`
	SourceKind         string             `db:"source_kind"`
	Severity           string             `db:"severity"`
	State              string             `db:"state"`
	SuggestedAction    string             `db:"suggested_action"`
	SummaryMarkdown    string             `db:"summary_markdown"`
	RelatedArtifactRef pgtype.Text        `db:"related_artifact_ref"`
	OpenedAt           time.Time          `db:"opened_at"`
	ClosedAt           pgtype.Timestamptz `db:"closed_at"`
	CreatedAt          time.Time          `db:"created_at"`
	UpdatedAt          time.Time          `db:"updated_at"`
}

// ProjectionSnapshotRow mirrors one change_governance_projection_snapshots row.
type ProjectionSnapshotRow struct {
	ID                int64     `db:"id"`
	PackageID         string    `db:"package_id"`
	ProjectionKind    string    `db:"projection_kind"`
	ProjectionVersion int64     `db:"projection_version"`
	IsCurrent         bool      `db:"is_current"`
	PayloadJSON       []byte    `db:"payload_json"`
	RefreshedAt       time.Time `db:"refreshed_at"`
	CreatedAt         time.Time `db:"created_at"`
}

// ArtifactLinkRow mirrors one change_governance_artifact_links row.
type ArtifactLinkRow struct {
	ID           int64     `db:"id"`
	PackageID    string    `db:"package_id"`
	ArtifactKind string    `db:"artifact_kind"`
	ArtifactRef  string    `db:"artifact_ref"`
	RelationKind string    `db:"relation_kind"`
	DisplayLabel string    `db:"display_label"`
	CreatedAt    time.Time `db:"created_at"`
}
