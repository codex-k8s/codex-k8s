package value

import (
	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

// ChangeGovernanceRolloutState captures domain enablement gates for the foundation path.
type ChangeGovernanceRolloutState struct {
	CoreFeatureEnabled bool
	SchemaReady        bool
	DomainReady        bool
	RunnerReady        bool
}

// ChangeGovernanceRolloutCapabilities describes what the current rollout state permits.
type ChangeGovernanceRolloutCapabilities struct {
	CanPersistFoundation   bool
	CanAcceptRunnerSignals bool
}

// ChangeGovernanceAggregate captures one hydrated package aggregate.
type ChangeGovernanceAggregate struct {
	Package            entitytypes.ChangeGovernancePackage
	Drafts             []entitytypes.ChangeGovernanceInternalDraft
	Waves              []entitytypes.ChangeGovernanceWave
	EvidenceBlocks     []entitytypes.ChangeGovernanceEvidenceBlock
	DecisionRecords    []entitytypes.ChangeGovernanceDecisionRecord
	FeedbackRecords    []entitytypes.ChangeGovernanceFeedbackRecord
	ArtifactLinks      []entitytypes.ChangeGovernanceArtifactLink
	CurrentProjections map[enumtypes.ChangeGovernanceProjectionKind]entitytypes.ChangeGovernanceProjectionSnapshot
}

// ChangeGovernanceDraftSignalResult describes draft-ingest outcome.
type ChangeGovernanceDraftSignalResult struct {
	PackageID    string
	DraftState   enumtypes.ChangeGovernanceDraftState
	NextStepKind enumtypes.ChangeGovernanceNextStepKind
}

// ChangeGovernanceWaveMapResult describes wave publication outcome.
type ChangeGovernanceWaveMapResult struct {
	PackageID         string
	PublicationState  enumtypes.ChangeGovernancePublicationState
	ProjectionVersion int64
}

// ChangeGovernanceEvidenceSignalResult describes evidence-upsert outcome.
type ChangeGovernanceEvidenceSignalResult struct {
	PackageID                 string
	EvidenceCompletenessState enumtypes.ChangeGovernanceEvidenceCompletenessState
	VerificationMinimumState  enumtypes.ChangeGovernanceVerificationMinimumState
	ProjectionVersion         int64
}
