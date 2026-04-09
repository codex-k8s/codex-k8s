package changegovernance

import (
	"time"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

type projectionPackageSummary struct {
	PackageID                 string                                              `json:"package_id"`
	RepositoryFullName        string                                              `json:"repository_full_name"`
	IssueNumber               int                                                 `json:"issue_number"`
	PRNumber                  *int                                                `json:"pr_number,omitempty"`
	RiskTier                  string                                              `json:"risk_tier,omitempty"`
	BundleAdmissibility       enumtypes.ChangeGovernanceBundleAdmissibility       `json:"bundle_admissibility"`
	PublicationState          enumtypes.ChangeGovernancePublicationState          `json:"publication_state"`
	EvidenceCompletenessState enumtypes.ChangeGovernanceEvidenceCompletenessState `json:"evidence_completeness_state"`
	VerificationMinimumState  enumtypes.ChangeGovernanceVerificationMinimumState  `json:"verification_minimum_state"`
	WaiverState               enumtypes.ChangeGovernanceWaiverState               `json:"waiver_state"`
	ReleaseReadinessState     enumtypes.ChangeGovernanceReleaseReadinessState     `json:"release_readiness_state"`
	GovernanceFeedbackState   enumtypes.ChangeGovernanceFeedbackState             `json:"governance_feedback_state"`
	OpenGapCount              int                                                 `json:"open_gap_count"`
	UpdatedAt                 time.Time                                           `json:"updated_at"`
}

type projectionArtifactLink struct {
	ArtifactKind string `json:"artifact_kind"`
	ArtifactRef  string `json:"artifact_ref"`
	RelationKind string `json:"relation_kind"`
	DisplayLabel string `json:"display_label"`
}

type projectionWaveItem struct {
	WaveKey                   string                                              `json:"wave_key"`
	PublishOrder              int                                                 `json:"publish_order"`
	DominantIntent            enumtypes.ChangeGovernanceDominantIntent            `json:"dominant_intent"`
	BoundedScopeKind          enumtypes.ChangeGovernanceBoundedScopeKind          `json:"bounded_scope_kind"`
	PublicationState          enumtypes.ChangeGovernanceWavePublicationState      `json:"publication_state"`
	EvidenceCompletenessState enumtypes.ChangeGovernanceEvidenceCompletenessState `json:"evidence_completeness_state"`
	VerificationMinimumState  enumtypes.ChangeGovernanceVerificationMinimumState  `json:"verification_minimum_state"`
	Summary                   string                                              `json:"summary"`
}

type projectionEvidenceBlock struct {
	BlockID           string                                             `json:"block_id"`
	ScopeKind         enumtypes.ChangeGovernanceEvidenceScopeKind        `json:"scope_kind"`
	ScopeRef          string                                             `json:"scope_ref"`
	BlockKind         enumtypes.ChangeGovernanceEvidenceBlockKind        `json:"block_kind"`
	State             enumtypes.ChangeGovernanceEvidenceBlockState       `json:"state"`
	RequiredByTier    bool                                               `json:"required_by_tier"`
	VerificationState enumtypes.ChangeGovernanceVerificationMinimumState `json:"verification_state"`
	ArtifactLinks     []projectionArtifactLink                           `json:"artifact_links"`
}

type projectionDecisionSummary struct {
	DecisionKind     enumtypes.ChangeGovernanceDecisionKind      `json:"decision_kind"`
	State            enumtypes.ChangeGovernanceDecisionState     `json:"state"`
	ActorKind        enumtypes.ChangeGovernanceDecisionActorKind `json:"actor_kind"`
	RecordedAt       time.Time                                   `json:"recorded_at"`
	ResidualRiskTier string                                      `json:"residual_risk_tier,omitempty"`
	Summary          string                                      `json:"summary"`
}

type projectionFeedbackRecord struct {
	GapID           string                                            `json:"gap_id"`
	GapKind         enumtypes.ChangeGovernanceFeedbackGapKind         `json:"gap_kind"`
	SourceKind      enumtypes.ChangeGovernanceFeedbackSourceKind      `json:"source_kind"`
	Severity        enumtypes.ChangeGovernanceFeedbackSeverity        `json:"severity"`
	State           enumtypes.ChangeGovernanceFeedbackRecordState     `json:"state"`
	SummaryMarkdown string                                            `json:"summary_markdown"`
	SuggestedAction enumtypes.ChangeGovernanceFeedbackSuggestedAction `json:"suggested_action"`
}

type packageDetailProjection struct {
	Package            projectionPackageSummary    `json:"package"`
	Waves              []projectionWaveItem        `json:"waves"`
	EvidenceBlocks     []projectionEvidenceBlock   `json:"evidence_blocks"`
	ActiveDecisions    []projectionDecisionSummary `json:"active_decisions"`
	FeedbackRecords    []projectionFeedbackRecord  `json:"feedback_records"`
	ArtifactLinks      []projectionArtifactLink    `json:"artifact_links"`
	CommentMirrorState string                      `json:"comment_mirror_state"`
}

type operatorGapQueueProjection struct {
	PackageID string                     `json:"package_id"`
	Items     []projectionFeedbackRecord `json:"items"`
}

type releaseGateProjection struct {
	Package projectionPackageSummary `json:"package"`
}

type githubStatusCommentProjection struct {
	Package       projectionPackageSummary `json:"package"`
	Waves         []projectionWaveItem     `json:"waves"`
	ArtifactLinks []projectionArtifactLink `json:"artifact_links"`
}
