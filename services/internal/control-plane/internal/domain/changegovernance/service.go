package changegovernance

import (
	"context"
	"fmt"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

// ReportDraftSignal records hidden-draft metadata and returns canonical package identity.
func (s *Service) ReportDraftSignal(ctx context.Context, params querytypes.ChangeGovernanceDraftSignalParams) (DraftSignalResult, error) {
	if s == nil {
		return DraftSignalResult{}, fmt.Errorf("change governance service is not configured")
	}
	if err := s.assertRunnerSignalsAllowed(); err != nil {
		return DraftSignalResult{}, err
	}
	if err := validateDraftSignalParams(params); err != nil {
		return DraftSignalResult{}, err
	}

	aggregate, duplicate, err := s.repo.RecordDraftSignal(ctx, params)
	if err != nil {
		return DraftSignalResult{}, err
	}

	nextStepKind := enumtypes.ChangeGovernanceNextStepKindWaveMapRequired
	if duplicate && len(aggregate.Waves) > 0 {
		nextStepKind = enumtypes.ChangeGovernanceNextStepKindNoOp
	}

	return DraftSignalResult{
		PackageID:    aggregate.Package.ID,
		DraftState:   enumtypes.ChangeGovernanceDraftStateHiddenRecorded,
		NextStepKind: nextStepKind,
	}, nil
}

// PublishWaveMap publishes semantic-wave lineage for one canonical package.
func (s *Service) PublishWaveMap(ctx context.Context, params querytypes.ChangeGovernanceWaveMapParams) (WaveMapResult, error) {
	if s == nil {
		return WaveMapResult{}, fmt.Errorf("change governance service is not configured")
	}
	if err := s.assertRunnerSignalsAllowed(); err != nil {
		return WaveMapResult{}, err
	}
	if err := validateWaveMapParams(params); err != nil {
		return WaveMapResult{}, err
	}

	aggregate, err := s.repo.PublishWaveMap(ctx, params)
	if err != nil {
		return WaveMapResult{}, err
	}
	return WaveMapResult{
		PackageID:         aggregate.Package.ID,
		PublicationState:  aggregate.Package.PublicationState,
		ProjectionVersion: aggregate.Package.ActiveProjectionVersion,
	}, nil
}

// UpsertEvidenceSignal records one package- or wave-scoped evidence block.
func (s *Service) UpsertEvidenceSignal(ctx context.Context, params querytypes.ChangeGovernanceEvidenceSignalParams) (EvidenceSignalResult, error) {
	if s == nil {
		return EvidenceSignalResult{}, fmt.Errorf("change governance service is not configured")
	}
	if err := s.assertRunnerSignalsAllowed(); err != nil {
		return EvidenceSignalResult{}, err
	}
	if err := validateEvidenceSignalParams(params); err != nil {
		return EvidenceSignalResult{}, err
	}

	aggregate, err := s.repo.UpsertEvidenceSignal(ctx, params)
	if err != nil {
		return EvidenceSignalResult{}, err
	}
	return EvidenceSignalResult{
		PackageID:                 aggregate.Package.ID,
		EvidenceCompletenessState: aggregate.Package.EvidenceCompletenessState,
		VerificationMinimumState:  aggregate.Package.VerificationMinimumState,
		ProjectionVersion:         aggregate.Package.ActiveProjectionVersion,
	}, nil
}

func (s *Service) effectiveRolloutState() valuetypes.ChangeGovernanceRolloutState {
	if s.rollout != nil {
		return s.rollout.CurrentChangeGovernanceRolloutState()
	}
	return s.cfg.RolloutState
}

func (s *Service) assertRunnerSignalsAllowed() error {
	caps, err := ResolveRolloutCapabilities(s.effectiveRolloutState())
	if err != nil {
		return err
	}
	if !caps.CanAcceptRunnerSignals {
		return errs.FailedPrecondition{Msg: "change governance runner signals require enabled schema and domain path"}
	}
	return nil
}

func validateDraftSignalParams(params querytypes.ChangeGovernanceDraftSignalParams) error {
	if strings.TrimSpace(params.RunID) == "" {
		return errs.Validation{Field: "run_id", Msg: "is required"}
	}
	if strings.TrimSpace(params.SignalID) == "" {
		return errs.Validation{Field: "signal_id", Msg: "is required"}
	}
	if strings.TrimSpace(params.CorrelationID) == "" {
		return errs.Validation{Field: "correlation_id", Msg: "is required"}
	}
	if strings.TrimSpace(params.ProjectID) == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if strings.TrimSpace(params.RepositoryFullName) == "" {
		return errs.Validation{Field: "repository_full_name", Msg: "is required"}
	}
	if params.IssueNumber <= 0 {
		return errs.Validation{Field: "issue_number", Msg: "must be > 0"}
	}
	if strings.TrimSpace(params.DraftRef) == "" {
		return errs.Validation{Field: "draft_ref", Msg: "is required"}
	}
	if params.DraftKind != enumtypes.ChangeGovernanceDraftKindInternalWorkingDraft {
		return errs.Validation{Field: "draft_kind", Msg: "must be internal_working_draft"}
	}
	if len(params.ChangeScopeHints) == 0 {
		return errs.Validation{Field: "change_scope_hints", Msg: "are required"}
	}
	for index, hint := range params.ChangeScopeHints {
		if strings.TrimSpace(hint.ContextKey) == "" {
			return errs.Validation{Field: fmt.Sprintf("change_scope_hints[%d].context_key", index), Msg: "is required"}
		}
		if strings.TrimSpace(string(hint.SurfaceKind)) == "" {
			return errs.Validation{Field: fmt.Sprintf("change_scope_hints[%d].surface_kind", index), Msg: "is required"}
		}
	}
	if params.OccurredAt.IsZero() {
		return errs.Validation{Field: "occurred_at", Msg: "is required"}
	}
	return nil
}

func validateWaveMapParams(params querytypes.ChangeGovernanceWaveMapParams) error {
	if strings.TrimSpace(params.PackageID) == "" {
		return errs.Validation{Field: "package_id", Msg: "is required"}
	}
	if strings.TrimSpace(params.ExpectedProjectID) == "" {
		return errs.Validation{Field: "expected_project_id", Msg: "is required"}
	}
	if strings.TrimSpace(params.WaveMapID) == "" {
		return errs.Validation{Field: "wave_map_id", Msg: "is required"}
	}
	if strings.TrimSpace(params.CorrelationID) == "" {
		return errs.Validation{Field: "correlation_id", Msg: "is required"}
	}
	if len(params.Waves) == 0 {
		return errs.Validation{Field: "waves", Msg: "are required"}
	}
	seenWaveKeys := make(map[string]struct{}, len(params.Waves))
	seenPublishOrder := make(map[int]struct{}, len(params.Waves))
	for _, wave := range params.Waves {
		if strings.TrimSpace(wave.WaveKey) == "" {
			return errs.Validation{Field: "wave_key", Msg: "is required"}
		}
		if wave.PublishOrder <= 0 {
			return errs.Validation{Field: "publish_order", Msg: "must be > 0"}
		}
		if strings.TrimSpace(wave.Summary) == "" {
			return errs.Validation{Field: "summary", Msg: "is required"}
		}
		if _, exists := seenWaveKeys[wave.WaveKey]; exists {
			return errs.Validation{Field: "wave_key", Msg: "must be unique within package"}
		}
		seenWaveKeys[wave.WaveKey] = struct{}{}
		if _, exists := seenPublishOrder[wave.PublishOrder]; exists {
			return errs.Validation{Field: "publish_order", Msg: "must be unique within package"}
		}
		seenPublishOrder[wave.PublishOrder] = struct{}{}
		if wave.BoundedScopeKind == enumtypes.ChangeGovernanceBoundedScopeKindMechanicalBoundedScope &&
			wave.DominantIntent != enumtypes.ChangeGovernanceDominantIntentMechanicalRefactor &&
			wave.DominantIntent != enumtypes.ChangeGovernanceDominantIntentDocsOnly {
			return errs.Validation{Field: "bounded_scope_kind", Msg: "mechanical_bounded_scope requires dominant_intent mechanical_refactor or docs_only"}
		}
	}
	if params.PublishedAt.IsZero() {
		return errs.Validation{Field: "published_at", Msg: "is required"}
	}
	return nil
}

func validateEvidenceSignalParams(params querytypes.ChangeGovernanceEvidenceSignalParams) error {
	if strings.TrimSpace(params.PackageID) == "" {
		return errs.Validation{Field: "package_id", Msg: "is required"}
	}
	if strings.TrimSpace(params.ExpectedProjectID) == "" {
		return errs.Validation{Field: "expected_project_id", Msg: "is required"}
	}
	if strings.TrimSpace(params.SignalID) == "" {
		return errs.Validation{Field: "signal_id", Msg: "is required"}
	}
	if strings.TrimSpace(params.CorrelationID) == "" {
		return errs.Validation{Field: "correlation_id", Msg: "is required"}
	}
	if params.ScopeKind != enumtypes.ChangeGovernanceEvidenceScopeKindPackage && params.ScopeKind != enumtypes.ChangeGovernanceEvidenceScopeKindWave {
		return errs.Validation{Field: "scope_kind", Msg: "must be package or wave"}
	}
	if strings.TrimSpace(params.ScopeRef) == "" {
		return errs.Validation{Field: "scope_ref", Msg: "is required"}
	}
	if strings.TrimSpace(string(params.BlockKind)) == "" {
		return errs.Validation{Field: "block_kind", Msg: "is required"}
	}
	if len(params.ArtifactLinks) == 0 {
		return errs.Validation{Field: "artifact_links", Msg: "are required"}
	}
	for index, artifactLink := range params.ArtifactLinks {
		if strings.TrimSpace(string(artifactLink.ArtifactKind)) == "" {
			return errs.Validation{Field: fmt.Sprintf("artifact_links[%d].artifact_kind", index), Msg: "is required"}
		}
		if strings.TrimSpace(artifactLink.ArtifactRef) == "" {
			return errs.Validation{Field: fmt.Sprintf("artifact_links[%d].artifact_ref", index), Msg: "is required"}
		}
		if strings.TrimSpace(string(artifactLink.RelationKind)) == "" {
			return errs.Validation{Field: fmt.Sprintf("artifact_links[%d].relation_kind", index), Msg: "is required"}
		}
	}
	if params.OccurredAt.IsZero() {
		return errs.Validation{Field: "occurred_at", Msg: "is required"}
	}
	return nil
}
