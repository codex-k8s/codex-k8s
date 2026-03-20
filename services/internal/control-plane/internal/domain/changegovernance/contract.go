package changegovernance

import (
	"context"

	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

type DraftSignalResult = valuetypes.ChangeGovernanceDraftSignalResult

type WaveMapResult = valuetypes.ChangeGovernanceWaveMapResult

type EvidenceSignalResult = valuetypes.ChangeGovernanceEvidenceSignalResult

// DomainService exposes canonical quality-governance foundation operations owned by control-plane.
type DomainService interface {
	ReportDraftSignal(ctx context.Context, params querytypes.ChangeGovernanceDraftSignalParams) (DraftSignalResult, error)
	PublishWaveMap(ctx context.Context, params querytypes.ChangeGovernanceWaveMapParams) (WaveMapResult, error)
	UpsertEvidenceSignal(ctx context.Context, params querytypes.ChangeGovernanceEvidenceSignalParams) (EvidenceSignalResult, error)
}
