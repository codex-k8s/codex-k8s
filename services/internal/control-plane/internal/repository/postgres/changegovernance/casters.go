package changegovernance

import (
	"encoding/json"

	domainrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/changegovernance"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	"github.com/codex-k8s/kodex/services/internal/control-plane/internal/repository/postgres/changegovernance/dbmodel"
)

func fromPackageRow(row dbmodel.PackageRow) domainrepo.Package {
	item := domainrepo.Package{
		ID:                        row.ID,
		PackageKey:                row.PackageKey,
		ProjectID:                 row.ProjectID,
		RepositoryFullName:        row.RepositoryFullName,
		IssueNumber:               int(row.IssueNumber),
		RiskTier:                  enumtypes.ChangeGovernanceRiskTier(row.RiskTier.String),
		BundleAdmissibility:       enumtypes.ChangeGovernanceBundleAdmissibility(row.BundleAdmissibility),
		PublicationState:          enumtypes.ChangeGovernancePublicationState(row.PublicationState),
		EvidenceCompletenessState: enumtypes.ChangeGovernanceEvidenceCompletenessState(row.EvidenceCompletenessState),
		VerificationMinimumState:  enumtypes.ChangeGovernanceVerificationMinimumState(row.VerificationMinimumState),
		WaiverState:               enumtypes.ChangeGovernanceWaiverState(row.WaiverState),
		ReleaseReadinessState:     enumtypes.ChangeGovernanceReleaseReadinessState(row.ReleaseReadinessState),
		GovernanceFeedbackState:   enumtypes.ChangeGovernanceFeedbackState(row.GovernanceFeedbackState),
		ActiveProjectionVersion:   row.ActiveProjectionVersion,
		CreatedAt:                 row.CreatedAt,
		UpdatedAt:                 row.UpdatedAt,
	}
	if row.PRNumber.Valid {
		value := int(row.PRNumber.Int32)
		item.PRNumber = &value
	}
	if row.LatestCorrelationID.Valid {
		item.LatestCorrelationID = row.LatestCorrelationID.String
	}
	return item
}

func fromDraftRow(row dbmodel.InternalDraftRow) domainrepo.InternalDraft {
	item := domainrepo.InternalDraft{ID: row.ID, PackageID: row.PackageID}
	item.SignalID = row.SignalID
	item.DraftRef = row.DraftRef
	item.DraftKind = enumtypes.ChangeGovernanceDraftKind(row.DraftKind)
	item.MetadataJSON = json.RawMessage(row.MetadataJSON)
	item.IsLatest = row.IsLatest
	item.OccurredAt = row.OccurredAt
	item.CreatedAt = row.CreatedAt
	if row.RunID.Valid {
		item.RunID = row.RunID.String
	}
	if row.DraftChecksum.Valid {
		item.DraftChecksum = row.DraftChecksum.String
	}
	return item
}

func fromWaveRow(row dbmodel.WaveRow) domainrepo.Wave {
	return domainrepo.Wave{
		ID:                        row.ID,
		PackageID:                 row.PackageID,
		WaveKey:                   row.WaveKey,
		PublishOrder:              int(row.PublishOrder),
		DominantIntent:            enumtypes.ChangeGovernanceDominantIntent(row.DominantIntent),
		BoundedScopeKind:          enumtypes.ChangeGovernanceBoundedScopeKind(row.BoundedScopeKind),
		PublicationState:          enumtypes.ChangeGovernanceWavePublicationState(row.PublicationState),
		EvidenceCompletenessState: enumtypes.ChangeGovernanceEvidenceCompletenessState(row.EvidenceCompletenessState),
		VerificationMinimumState:  enumtypes.ChangeGovernanceVerificationMinimumState(row.VerificationMinimumState),
		Summary:                   row.Summary,
		VerificationTargetsJSON:   json.RawMessage(row.VerificationTargetsJSON),
		CreatedAt:                 row.CreatedAt,
		UpdatedAt:                 row.UpdatedAt,
	}
}

func fromEvidenceBlockRow(row dbmodel.EvidenceBlockRow) domainrepo.EvidenceBlock {
	item := domainrepo.EvidenceBlock{
		ID:                row.ID,
		PackageID:         row.PackageID,
		BlockKind:         enumtypes.ChangeGovernanceEvidenceBlockKind(row.BlockKind),
		State:             enumtypes.ChangeGovernanceEvidenceBlockState(row.State),
		VerificationState: enumtypes.ChangeGovernanceVerificationMinimumState(row.VerificationState),
		RequiredByTier:    row.RequiredByTier,
		SourceKind:        enumtypes.ChangeGovernanceEvidenceSourceKind(row.SourceKind),
		ArtifactLinksJSON: json.RawMessage(row.ArtifactLinksJSON),
		ObservedAt:        row.ObservedAt,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
	if row.WaveID.Valid {
		item.WaveID = row.WaveID.String
	}
	if row.LatestSignalID.Valid {
		item.LatestSignalID = row.LatestSignalID.String
	}
	return item
}

func fromDecisionRecordRow(row dbmodel.DecisionRecordRow) domainrepo.DecisionRecord {
	item := domainrepo.DecisionRecord{
		ID:                  row.ID,
		PackageID:           row.PackageID,
		ScopeKind:           enumtypes.ChangeGovernanceDecisionScopeKind(row.ScopeKind),
		ScopeRef:            row.ScopeRef,
		DecisionID:          row.DecisionID,
		DecisionKind:        enumtypes.ChangeGovernanceDecisionKind(row.DecisionKind),
		State:               enumtypes.ChangeGovernanceDecisionState(row.State),
		ActorKind:           enumtypes.ChangeGovernanceDecisionActorKind(row.ActorKind),
		SummaryMarkdown:     row.SummaryMarkdown,
		DecisionPayloadJSON: json.RawMessage(row.DecisionPayloadJSON),
		RecordedAt:          row.RecordedAt,
		CreatedAt:           row.CreatedAt,
	}
	if row.ResidualRiskTier.Valid {
		item.ResidualRiskTier = enumtypes.ChangeGovernanceRiskTier(row.ResidualRiskTier.String)
	}
	return item
}

func fromFeedbackRecordRow(row dbmodel.FeedbackRecordRow) domainrepo.FeedbackRecord {
	item := domainrepo.FeedbackRecord{
		ID:              row.ID,
		PackageID:       row.PackageID,
		FeedbackID:      row.FeedbackID,
		GapKind:         enumtypes.ChangeGovernanceFeedbackGapKind(row.GapKind),
		SourceKind:      enumtypes.ChangeGovernanceFeedbackSourceKind(row.SourceKind),
		Severity:        enumtypes.ChangeGovernanceFeedbackSeverity(row.Severity),
		State:           enumtypes.ChangeGovernanceFeedbackRecordState(row.State),
		SuggestedAction: enumtypes.ChangeGovernanceFeedbackSuggestedAction(row.SuggestedAction),
		SummaryMarkdown: row.SummaryMarkdown,
		OpenedAt:        row.OpenedAt,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.RelatedArtifactRef.Valid {
		item.RelatedArtifactRef = row.RelatedArtifactRef.String
	}
	if row.ClosedAt.Valid {
		value := row.ClosedAt.Time
		item.ClosedAt = &value
	}
	return item
}

func fromProjectionSnapshotRow(row dbmodel.ProjectionSnapshotRow) domainrepo.ProjectionSnapshot {
	return domainrepo.ProjectionSnapshot{
		ID:                row.ID,
		PackageID:         row.PackageID,
		ProjectionKind:    enumtypes.ChangeGovernanceProjectionKind(row.ProjectionKind),
		ProjectionVersion: row.ProjectionVersion,
		IsCurrent:         row.IsCurrent,
		PayloadJSON:       json.RawMessage(row.PayloadJSON),
		RefreshedAt:       row.RefreshedAt,
		CreatedAt:         row.CreatedAt,
	}
}

func fromArtifactLinkRow(row dbmodel.ArtifactLinkRow) domainrepo.ArtifactLink {
	return domainrepo.ArtifactLink{
		ID:           row.ID,
		PackageID:    row.PackageID,
		ArtifactKind: enumtypes.ChangeGovernanceArtifactKind(row.ArtifactKind),
		ArtifactRef:  row.ArtifactRef,
		RelationKind: enumtypes.ChangeGovernanceArtifactRelationKind(row.RelationKind),
		DisplayLabel: row.DisplayLabel,
		CreatedAt:    row.CreatedAt,
	}
}
