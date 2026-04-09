package changegovernance

import (
	"context"

	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

type (
	Package            = entitytypes.ChangeGovernancePackage
	InternalDraft      = entitytypes.ChangeGovernanceInternalDraft
	Wave               = entitytypes.ChangeGovernanceWave
	EvidenceBlock      = entitytypes.ChangeGovernanceEvidenceBlock
	DecisionRecord     = entitytypes.ChangeGovernanceDecisionRecord
	FeedbackRecord     = entitytypes.ChangeGovernanceFeedbackRecord
	ProjectionSnapshot = entitytypes.ChangeGovernanceProjectionSnapshot
	ArtifactLink       = entitytypes.ChangeGovernanceArtifactLink
	Aggregate          = valuetypes.ChangeGovernanceAggregate
)

// Repository persists canonical change-governance aggregate state in PostgreSQL.
type Repository interface {
	RecordDraftSignal(ctx context.Context, params querytypes.ChangeGovernanceDraftSignalParams) (Aggregate, bool, error)
	PublishWaveMap(ctx context.Context, params querytypes.ChangeGovernanceWaveMapParams) (Aggregate, error)
	UpsertEvidenceSignal(ctx context.Context, params querytypes.ChangeGovernanceEvidenceSignalParams) (Aggregate, error)
	GetAggregateByPackageID(ctx context.Context, packageID string) (Aggregate, bool, error)
}
