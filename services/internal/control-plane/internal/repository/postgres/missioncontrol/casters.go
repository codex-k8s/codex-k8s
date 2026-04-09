package missioncontrol

import (
	"encoding/json"

	domainrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/missioncontrol"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
	"github.com/codex-k8s/kodex/services/internal/control-plane/internal/repository/postgres/missioncontrol/dbmodel"
)

func fromEntityRow(row dbmodel.EntityRow) domainrepo.Entity {
	item := domainrepo.Entity{
		ID:                row.ID,
		ProjectID:         row.ProjectID,
		EntityKind:        enumtypes.MissionControlEntityKind(row.EntityKind),
		EntityExternalKey: row.EntityExternalKey,
		ProviderKind:      enumtypes.MissionControlProviderKind(row.ProviderKind),
		Title:             row.Title,
		ActiveState:       enumtypes.MissionControlActiveState(row.ActiveState),
		SyncStatus:        enumtypes.MissionControlSyncStatus(row.SyncStatus),
		ContinuityStatus:  enumtypes.MissionControlContinuityStatus(row.ContinuityStatus),
		CoverageClass:     enumtypes.MissionControlCoverageClass(row.CoverageClass),
		ProjectionVersion: row.ProjectionVersion,
		CardPayloadJSON:   json.RawMessage(row.CardPayloadJSON),
		DetailPayloadJSON: json.RawMessage(row.DetailPayloadJSON),
		ProjectedAt:       row.ProjectedAt,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
	if row.ProviderURL.Valid {
		item.ProviderURL = row.ProviderURL.String
	}
	if row.LastTimelineAt.Valid {
		value := row.LastTimelineAt.Time
		item.LastTimelineAt = &value
	}
	if row.ProviderUpdatedAt.Valid {
		value := row.ProviderUpdatedAt.Time
		item.ProviderUpdatedAt = &value
	}
	if row.StaleAfter.Valid {
		value := row.StaleAfter.Time
		item.StaleAfter = &value
	}
	return item
}

func fromRelationRow(row dbmodel.RelationRow) domainrepo.Relation {
	return domainrepo.Relation{
		ID:             row.ID,
		ProjectID:      row.ProjectID,
		SourceEntityID: row.SourceEntityID,
		RelationKind:   enumtypes.MissionControlRelationKind(row.RelationKind),
		TargetEntityID: row.TargetEntityID,
		SourceKind:     enumtypes.MissionControlRelationSourceKind(row.SourceKind),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func fromTimelineEntryRow(row dbmodel.TimelineEntryRow) domainrepo.TimelineEntry {
	item := domainrepo.TimelineEntry{
		ID:               row.ID,
		ProjectID:        row.ProjectID,
		EntityID:         row.EntityID,
		SourceKind:       enumtypes.MissionControlTimelineSourceKind(row.SourceKind),
		EntryExternalKey: row.EntryExternalKey,
		Summary:          row.Summary,
		PayloadJSON:      json.RawMessage(row.PayloadJSON),
		OccurredAt:       row.OccurredAt,
		IsReadOnly:       row.IsReadOnly,
		CreatedAt:        row.CreatedAt,
	}
	if row.CommandID.Valid {
		item.CommandID = row.CommandID.String
	}
	if row.BodyMarkdown.Valid {
		item.BodyMarkdown = row.BodyMarkdown.String
	}
	if row.ProviderURL.Valid {
		item.ProviderURL = row.ProviderURL.String
	}
	return item
}

func fromContinuityGapRow(row dbmodel.ContinuityGapRow) domainrepo.ContinuityGap {
	item := domainrepo.ContinuityGap{
		ID:              row.ID,
		ProjectID:       row.ProjectID,
		SubjectEntityID: row.SubjectEntityID,
		GapKind:         enumtypes.MissionControlGapKind(row.GapKind),
		Severity:        enumtypes.MissionControlGapSeverity(row.Severity),
		Status:          enumtypes.MissionControlGapStatus(row.Status),
		ResolutionHint:  row.ResolutionHint.String,
		PayloadJSON:     json.RawMessage(row.PayloadJSON),
		DetectedAt:      row.DetectedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.ExpectedEntityKind.Valid {
		item.ExpectedEntityKind = enumtypes.MissionControlEntityKind(row.ExpectedEntityKind.String)
	}
	if row.ExpectedStageLabel.Valid {
		item.ExpectedStageLabel = row.ExpectedStageLabel.String
	}
	if row.ResolutionEntityID.Valid {
		value := row.ResolutionEntityID.Int64
		item.ResolutionEntityID = &value
	}
	if row.ResolvedAt.Valid {
		value := row.ResolvedAt.Time
		item.ResolvedAt = &value
	}
	return item
}

func fromWorkspaceWatermarkRow(row dbmodel.WorkspaceWatermarkRow) domainrepo.WorkspaceWatermark {
	item := domainrepo.WorkspaceWatermark{
		ID:            row.ID,
		ProjectID:     row.ProjectID,
		WatermarkKind: enumtypes.MissionControlWorkspaceWatermarkKind(row.WatermarkKind),
		Status:        enumtypes.MissionControlWorkspaceWatermarkStatus(row.Status),
		Summary:       row.Summary,
		ObservedAt:    row.ObservedAt,
		PayloadJSON:   json.RawMessage(row.PayloadJSON),
		CreatedAt:     row.CreatedAt,
	}
	if row.WindowStartedAt.Valid {
		value := row.WindowStartedAt.Time
		item.WindowStartedAt = &value
	}
	if row.WindowEndedAt.Valid {
		value := row.WindowEndedAt.Time
		item.WindowEndedAt = &value
	}
	return item
}

func fromCommandRow(row dbmodel.CommandRow) domainrepo.Command {
	item := domainrepo.Command{
		ID:                 row.ID,
		ProjectID:          row.ProjectID,
		CommandKind:        enumtypes.MissionControlCommandKind(row.CommandKind),
		ActorID:            row.ActorID,
		BusinessIntentKey:  row.BusinessIntentKey,
		CorrelationID:      row.CorrelationID,
		Status:             enumtypes.MissionControlCommandStatus(row.Status),
		ApprovalState:      enumtypes.MissionControlApprovalState(row.ApprovalState),
		PayloadJSON:        json.RawMessage(row.PayloadJSON),
		ResultPayloadJSON:  json.RawMessage(row.ResultPayloadJSON),
		ProviderDeliveries: json.RawMessage(row.ProviderDeliveries),
		RequestedAt:        row.RequestedAt,
		UpdatedAt:          row.UpdatedAt,
	}
	if row.TargetEntityID.Valid {
		value := row.TargetEntityID.Int64
		item.TargetEntityID = &value
	}
	if row.FailureReason.Valid {
		item.FailureReason = enumtypes.MissionControlCommandFailureReason(row.FailureReason.String)
	}
	if row.ApprovalRequestID.Valid {
		item.ApprovalRequestID = row.ApprovalRequestID.String
	}
	if row.ApprovalRequestedAt.Valid {
		value := row.ApprovalRequestedAt.Time
		item.ApprovalRequestedAt = &value
	}
	if row.ApprovalDecidedAt.Valid {
		value := row.ApprovalDecidedAt.Time
		item.ApprovalDecidedAt = &value
	}
	if row.LeaseOwner.Valid {
		item.LeaseOwner = row.LeaseOwner.String
	}
	if row.LeaseUntil.Valid {
		value := row.LeaseUntil.Time
		item.LeaseUntil = &value
	}
	if row.ReconciledAt.Valid {
		value := row.ReconciledAt.Time
		item.ReconciledAt = &value
	}
	return item
}

func fromWarmupSummaryRow(row dbmodel.WarmupSummaryRow) valuetypes.MissionControlWarmupSummary {
	return valuetypes.MissionControlWarmupSummary{
		ProjectID:                    row.ProjectID,
		EntityCount:                  row.EntityCount,
		RelationCount:                row.RelationCount,
		TimelineEntryCount:           row.TimelineEntryCount,
		CommandCount:                 row.CommandCount,
		MaxProjectionVersion:         row.MaxProjectionVersion,
		RunEntityCount:               row.RunEntityCount,
		LegacyAgentCount:             row.LegacyAgentCount,
		ContinuityGapCount:           row.ContinuityGapCount,
		OpenContinuityGapCount:       row.OpenContinuityGapCount,
		BlockingGapCount:             row.BlockingGapCount,
		MissingPullRequestGapCount:   row.MissingPullRequestGapCount,
		MissingFollowUpIssueGapCount: row.MissingFollowUpIssueGapCount,
		WatermarkCount:               row.WatermarkCount,
	}
}
