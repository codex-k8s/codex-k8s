package casters

import (
	"fmt"
	"strings"
	"time"

	"github.com/codex-k8s/kodex/libs/go/cast"
	controlplanev1 "github.com/codex-k8s/kodex/proto/gen/go/kodex/controlplane/v1"
	"github.com/codex-k8s/kodex/services/external/api-gateway/internal/transport/http/generated"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MissionControlDashboardSnapshot(
	item *controlplanev1.MissionControlDashboardSnapshot,
	resumeToken string,
) generated.MissionControlDashboardSnapshot {
	if item == nil {
		return generated.MissionControlDashboardSnapshot{}
	}

	out := generated.MissionControlDashboardSnapshot{
		SnapshotId:          item.GetSnapshotId(),
		ViewMode:            generated.MissionControlDashboardSnapshotViewMode(item.GetViewMode()),
		FreshnessStatus:     generated.MissionControlDashboardSnapshotFreshnessStatus(item.GetFreshnessStatus()),
		GeneratedAt:         item.GetGeneratedAt().AsTime().UTC(),
		StaleAfter:          item.GetStaleAfter().AsTime().UTC(),
		RealtimeResumeToken: strings.TrimSpace(resumeToken),
		Summary:             missionControlDashboardSummary(item.GetSummary()),
		Entities:            missionControlEntityCards(item.GetEntities()),
		Relations:           missionControlRelations(item.GetRelations()),
		NextPageCursor:      cast.OptionalTrimmedString(item.NextPageCursor),
	}
	return out
}

func MissionControlEntityDetails(item *controlplanev1.MissionControlEntityDetails) (generated.MissionControlEntityDetails, error) {
	if item == nil {
		return generated.MissionControlEntityDetails{}, nil
	}
	detailPayload, err := missionControlEntityDetailsPayload(item)
	if err != nil {
		return generated.MissionControlEntityDetails{}, err
	}
	return generated.MissionControlEntityDetails{
		Entity:            missionControlEntityCard(item.GetEntity()),
		DetailPayload:     detailPayload,
		Relations:         missionControlRelations(item.GetRelations()),
		TimelinePreview:   missionControlTimelineEntries(item.GetTimelinePreview()),
		AllowedActions:    missionControlAllowedActions(item.GetAllowedActions()),
		ProviderDeepLinks: missionControlProviderDeepLinks(item.GetProviderDeepLinks()),
	}, nil
}

func MissionControlTimelineItems(
	items []*controlplanev1.MissionControlTimelineEntry,
	nextCursor *string,
) generated.MissionControlTimelineItemsResponse {
	return generated.MissionControlTimelineItemsResponse{
		Items:      missionControlTimelineEntries(items),
		NextCursor: cast.OptionalTrimmedString(nextCursor),
	}
}

func MissionControlWorkspaceSnapshot(
	item *controlplanev1.MissionControlWorkspaceSnapshot,
	resumeToken string,
) generated.MissionControlWorkspaceSnapshot {
	if item == nil {
		return generated.MissionControlWorkspaceSnapshot{}
	}

	out := generated.MissionControlWorkspaceSnapshot{
		SnapshotId:          item.GetSnapshotId(),
		ViewMode:            generated.MissionControlWorkspaceSnapshotViewMode(item.GetViewMode()),
		GeneratedAt:         item.GetGeneratedAt().AsTime().UTC(),
		EffectiveFilters:    missionControlWorkspaceFilters(item.GetEffectiveFilters()),
		Summary:             missionControlWorkspaceSummary(item.GetSummary()),
		WorkspaceWatermarks: missionControlWorkspaceWatermarks(item.GetWorkspaceWatermarks()),
		RootGroups:          missionControlRootGroups(item.GetRootGroups()),
		Nodes:               missionControlNodes(item.GetNodes()),
		Edges:               missionControlEdges(item.GetEdges()),
		ResumeToken:         strings.TrimSpace(resumeToken),
		NextRootCursor:      cast.OptionalTrimmedString(item.NextRootCursor),
	}
	return out
}

func MissionControlNodeDetails(item *controlplanev1.MissionControlNodeDetails) (generated.MissionControlNodeDetails, error) {
	if item == nil {
		return generated.MissionControlNodeDetails{}, nil
	}

	detailPayload, err := missionControlNodeDetailsPayload(item)
	if err != nil {
		return generated.MissionControlNodeDetails{}, err
	}

	return generated.MissionControlNodeDetails{
		Node:              missionControlNode(item.GetNode()),
		AdjacentNodes:     missionControlNodes(item.GetAdjacentNodes()),
		AdjacentEdges:     missionControlEdges(item.GetAdjacentEdges()),
		ContinuityGaps:    missionControlContinuityGaps(item.GetContinuityGaps()),
		DetailPayload:     detailPayload,
		ActivityPreview:   missionControlActivityEntries(item.GetActivityPreview()),
		LaunchSurfaces:    missionControlLaunchSurfaces(item.GetLaunchSurfaces()),
		NodeWatermarks:    missionControlWorkspaceWatermarks(item.GetNodeWatermarks()),
		ProviderDeepLinks: missionControlProviderDeepLinks(item.GetProviderDeepLinks()),
	}, nil
}

func MissionControlNodeActivityItems(
	items []*controlplanev1.MissionControlActivityEntry,
	nextCursor *string,
) generated.MissionControlNodeActivityItemsResponse {
	return generated.MissionControlNodeActivityItemsResponse{
		Items:      missionControlActivityEntries(items),
		NextCursor: cast.OptionalTrimmedString(nextCursor),
	}
}

func MissionControlLaunchPreview(item *controlplanev1.MissionControlLaunchPreview) generated.MissionControlLaunchPreview {
	if item == nil {
		return generated.MissionControlLaunchPreview{}
	}

	out := generated.MissionControlLaunchPreview{
		PreviewId:           item.GetPreviewId(),
		ApprovalRequirement: generated.MissionControlLaunchPreviewApprovalRequirement(item.GetApprovalRequirement()),
		LabelDiff: generated.MissionControlLaunchPreviewLabelDiff{
			AddedLabels:   requiredStringSlice(item.GetLabelDiff().GetAddedLabels()),
			RemovedLabels: requiredStringSlice(item.GetLabelDiff().GetRemovedLabels()),
			FinalLabels:   requiredStringSlice(item.GetLabelDiff().GetFinalLabels()),
		},
		ContinuityEffect: generated.MissionControlLaunchPreviewContinuityEffect{
			ResolvedGapIds:    append([]int64{}, item.GetContinuityEffect().GetResolvedGapIds()...),
			RemainingGapIds:   append([]int64{}, item.GetContinuityEffect().GetRemainingGapIds()...),
			ResultingNodeRefs: missionControlNodeRefs(item.GetContinuityEffect().GetResultingNodeRefs()),
			ProviderRedirects: requiredStringSlice(item.GetContinuityEffect().GetProviderRedirects()),
		},
	}
	if blockedReason := strings.TrimSpace(item.GetBlockingReason()); blockedReason != "" {
		out.BlockingReason = &blockedReason
	}
	return out
}

func MissionControlCommandState(item *controlplanev1.MissionControlCommandState) generated.MissionControlCommandState {
	if item == nil {
		return generated.MissionControlCommandState{}
	}
	out := generated.MissionControlCommandState{
		CommandId:           item.GetCommandId(),
		CommandKind:         generated.MissionControlCommandStateCommandKind(item.GetCommandKind()),
		Status:              generated.MissionControlCommandStateStatus(item.GetStatus()),
		CorrelationId:       item.GetCorrelationId(),
		UpdatedAt:           item.GetUpdatedAt().AsTime().UTC(),
		BusinessIntentKey:   item.GetBusinessIntentKey(),
		EntityRefs:          missionControlEntityRefs(item.GetEntityRefs()),
		ProviderDeliveryIds: requiredStringSlice(item.GetProviderDeliveryIds()),
		BlockingReason:      cast.OptionalTrimmedString(item.BlockingReason),
		StatusMessage:       cast.OptionalTrimmedString(item.StatusMessage),
	}
	if failureReason := strings.TrimSpace(item.GetFailureReason()); failureReason != "" {
		value := generated.MissionControlCommandStateFailureReason(failureReason)
		out.FailureReason = &value
	}
	if item.GetReconciledAt() != nil {
		value := item.GetReconciledAt().AsTime().UTC()
		out.ReconciledAt = &value
	}
	if approval := missionControlCommandApproval(item.GetApproval()); approval != nil {
		out.Approval = approval
	}
	return out
}

func MissionControlCommandRequest(
	body generated.MissionControlCommandRequest,
	correlationID string,
	requestedAt time.Time,
) (*controlplanev1.SubmitMissionControlCommandRequest, error) {
	if item, err := body.AsMissionControlDiscussionCreateCommandRequest(); err == nil {
		req := newMissionControlCommandRequestBase(
			string(item.CommandKind),
			item.ProjectId,
			item.BusinessIntentKey,
			item.ExpectedProjectionVersion,
			item.TargetEntityKind,
			item.TargetEntityPublicId,
			correlationID,
			requestedAt,
		)
		req.Payload = &controlplanev1.SubmitMissionControlCommandRequest_DiscussionCreate{
			DiscussionCreate: &controlplanev1.MissionControlDiscussionCreatePayload{
				Title:                item.Payload.Title,
				BodyMarkdown:         cast.OptionalTrimmedString(item.Payload.BodyMarkdown),
				ParentEntityKind:     optionalStringPtr(item.Payload.ParentEntityKind),
				ParentEntityPublicId: cast.OptionalTrimmedString(item.Payload.ParentEntityPublicId),
			},
		}
		return req, nil
	}
	if item, err := body.AsMissionControlWorkItemCreateCommandRequest(); err == nil {
		req := newMissionControlCommandRequestBase(
			string(item.CommandKind),
			item.ProjectId,
			item.BusinessIntentKey,
			item.ExpectedProjectionVersion,
			item.TargetEntityKind,
			item.TargetEntityPublicId,
			correlationID,
			requestedAt,
		)
		req.Payload = &controlplanev1.SubmitMissionControlCommandRequest_WorkItemCreate{
			WorkItemCreate: &controlplanev1.MissionControlWorkItemCreatePayload{
				Title:             item.Payload.Title,
				BodyMarkdown:      cast.OptionalTrimmedString(item.Payload.BodyMarkdown),
				InitialLabels:     derefStringSlice(item.Payload.InitialLabels),
				RelatedEntityRefs: missionControlEntityRefsToProtoRequest(item.Payload.RelatedEntityRefs),
			},
		}
		return req, nil
	}
	if item, err := body.AsMissionControlDiscussionFormalizeCommandRequest(); err == nil {
		req := newMissionControlCommandRequestBase(
			string(item.CommandKind),
			item.ProjectId,
			item.BusinessIntentKey,
			item.ExpectedProjectionVersion,
			item.TargetEntityKind,
			item.TargetEntityPublicId,
			correlationID,
			requestedAt,
		)
		req.Payload = &controlplanev1.SubmitMissionControlCommandRequest_DiscussionFormalize{
			DiscussionFormalize: &controlplanev1.MissionControlDiscussionFormalizePayload{
				SourceEntityKind:     string(item.Payload.SourceEntityKind),
				SourceEntityPublicId: item.Payload.SourceEntityPublicId,
				FormalizedKind:       item.Payload.FormalizedKind,
				Title:                item.Payload.Title,
				BodyMarkdown:         cast.OptionalTrimmedString(item.Payload.BodyMarkdown),
			},
		}
		return req, nil
	}
	if item, err := body.AsMissionControlStageNextStepCommandRequest(); err == nil {
		req := newMissionControlCommandRequestBase(
			string(item.CommandKind),
			item.ProjectId,
			item.BusinessIntentKey,
			item.ExpectedProjectionVersion,
			item.TargetEntityKind,
			item.TargetEntityPublicId,
			correlationID,
			requestedAt,
		)
		req.Payload = &controlplanev1.SubmitMissionControlCommandRequest_StageNextStep{
			StageNextStep: &controlplanev1.MissionControlStageNextStepPayload{
				ThreadKind:          item.Payload.ThreadKind,
				ThreadNumber:        item.Payload.ThreadNumber,
				TargetLabel:         item.Payload.TargetLabel,
				RemovedLabels:       derefStringSlice(item.Payload.RemovedLabels),
				DisplayVariant:      cast.OptionalTrimmedString(item.Payload.DisplayVariant),
				ApprovalRequirement: optionalStringPtr(item.Payload.ApprovalRequirement),
			},
		}
		return req, nil
	}
	if item, err := body.AsMissionControlRetrySyncCommandRequest(); err == nil {
		req := newMissionControlCommandRequestBase(
			string(item.CommandKind),
			item.ProjectId,
			item.BusinessIntentKey,
			item.ExpectedProjectionVersion,
			item.TargetEntityKind,
			item.TargetEntityPublicId,
			correlationID,
			requestedAt,
		)
		req.Payload = &controlplanev1.SubmitMissionControlCommandRequest_RetrySync{
			RetrySync: &controlplanev1.MissionControlRetrySyncPayload{
				CommandId:      item.Payload.CommandId,
				RetryReason:    cast.OptionalTrimmedString(item.Payload.RetryReason),
				ExpectedStatus: optionalStringPtr(item.Payload.ExpectedStatus),
			},
		}
		return req, nil
	}
	return nil, fmt.Errorf("mission control command request variant is not supported")
}

func missionControlDashboardSummary(item *controlplanev1.MissionControlSnapshotSummary) generated.MissionControlDashboardSummary {
	if item == nil {
		return generated.MissionControlDashboardSummary{}
	}
	return generated.MissionControlDashboardSummary{
		TotalEntities:              item.GetTotalEntities(),
		WorkingCount:               item.GetWorkingCount(),
		WaitingCount:               item.GetWaitingCount(),
		BlockedCount:               item.GetBlockedCount(),
		ReviewCount:                item.GetReviewCount(),
		RecentCriticalUpdatesCount: item.GetRecentCriticalUpdatesCount(),
	}
}

func missionControlEntityCards(items []*controlplanev1.MissionControlEntityCard) []generated.MissionControlEntityCard {
	out := make([]generated.MissionControlEntityCard, 0, len(items))
	for _, item := range items {
		out = append(out, missionControlEntityCard(item))
	}
	return out
}

func missionControlEntityCard(item *controlplanev1.MissionControlEntityCard) generated.MissionControlEntityCard {
	if item == nil {
		return generated.MissionControlEntityCard{}
	}
	return generated.MissionControlEntityCard{
		EntityKind:        generated.MissionControlEntityCardEntityKind(item.GetEntityKind()),
		EntityPublicId:    item.GetEntityPublicId(),
		Title:             item.GetTitle(),
		State:             generated.MissionControlEntityCardState(item.GetState()),
		SyncStatus:        generated.MissionControlEntityCardSyncStatus(item.GetSyncStatus()),
		ProviderReference: missionControlProviderReference(item.GetProviderReference()),
		PrimaryActor:      missionControlPrimaryActor(item.GetPrimaryActor()),
		RelationCount:     item.GetRelationCount(),
		LastTimelineAt:    protoTimestampPtr(item.GetLastTimelineAt()),
		Badges:            missionControlEntityCardBadges(item.GetBadges()),
		ProjectionVersion: item.GetProjectionVersion(),
	}
}

func missionControlProviderReference(item *controlplanev1.MissionControlProviderReference) generated.MissionControlProviderReference {
	if item == nil {
		return generated.MissionControlProviderReference{}
	}
	return generated.MissionControlProviderReference{
		Provider:   generated.MissionControlProviderReferenceProvider(item.GetProvider()),
		ExternalId: item.GetExternalId(),
		Url:        cast.TrimmedStringPtr(item.GetUrl()),
	}
}

func missionControlPrimaryActor(item *controlplanev1.MissionControlPrimaryActor) *generated.MissionControlPrimaryActor {
	if item == nil {
		return nil
	}
	return &generated.MissionControlPrimaryActor{
		ActorType:   item.GetActorType(),
		ActorId:     item.GetActorId(),
		DisplayName: item.GetDisplayName(),
	}
}

func missionControlEntityCardBadges(items []string) []generated.MissionControlEntityCardBadges {
	return typedTrimmedStrings[generated.MissionControlEntityCardBadges](items)
}

func missionControlRelations(items []*controlplanev1.MissionControlRelation) []generated.MissionControlRelation {
	out := make([]generated.MissionControlRelation, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, generated.MissionControlRelation{
			RelationKind:         generated.MissionControlRelationRelationKind(item.GetRelationKind()),
			SourceKind:           generated.MissionControlRelationSourceKind(item.GetSourceKind()),
			SourceEntityKind:     generated.MissionControlRelationSourceEntityKind(item.GetSourceEntityKind()),
			SourceEntityPublicId: item.GetSourceEntityPublicId(),
			TargetEntityKind:     generated.MissionControlRelationTargetEntityKind(item.GetTargetEntityKind()),
			TargetEntityPublicId: item.GetTargetEntityPublicId(),
			Direction:            generated.MissionControlRelationDirection(item.GetDirection()),
		})
	}
	return out
}

func missionControlTimelineEntries(items []*controlplanev1.MissionControlTimelineEntry) []generated.MissionControlTimelineEntry {
	out := make([]generated.MissionControlTimelineEntry, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, generated.MissionControlTimelineEntry{
			EntryId:        item.GetEntryId(),
			EntityKind:     generated.MissionControlTimelineEntryEntityKind(item.GetEntityKind()),
			EntityPublicId: item.GetEntityPublicId(),
			SourceKind:     generated.MissionControlTimelineEntrySourceKind(item.GetSourceKind()),
			SourceRef:      item.GetSourceRef(),
			OccurredAt:     item.GetOccurredAt().AsTime().UTC(),
			Summary:        item.GetSummary(),
			BodyMarkdown:   cast.OptionalTrimmedString(item.BodyMarkdown),
			CommandId:      cast.OptionalTrimmedString(item.CommandId),
			ProviderUrl:    cast.OptionalTrimmedString(item.ProviderUrl),
			IsReadOnly:     item.GetIsReadOnly(),
		})
	}
	return out
}

func missionControlAllowedActions(items []*controlplanev1.MissionControlAllowedAction) []generated.MissionControlAllowedAction {
	out := make([]generated.MissionControlAllowedAction, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, generated.MissionControlAllowedAction{
			ActionKind:          generated.MissionControlAllowedActionActionKind(item.GetActionKind()),
			Presentation:        generated.MissionControlAllowedActionPresentation(item.GetPresentation()),
			AllowedWhenDegraded: item.GetAllowedWhenDegraded(),
			ApprovalRequirement: generated.MissionControlAllowedActionApprovalRequirement(item.GetApprovalRequirement()),
			BlockedReason:       cast.OptionalTrimmedString(item.BlockedReason),
		})
	}
	return out
}

func missionControlProviderDeepLinks(items []*controlplanev1.MissionControlProviderDeepLink) []generated.MissionControlProviderDeepLink {
	out := make([]generated.MissionControlProviderDeepLink, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, generated.MissionControlProviderDeepLink{
			ActionKind: generated.MissionControlProviderDeepLinkActionKind(item.GetActionKind()),
			Url:        item.GetUrl(),
			Reason:     generated.MissionControlProviderDeepLinkReason(item.GetReason()),
		})
	}
	return out
}

func missionControlWorkspaceFilters(item *controlplanev1.MissionControlWorkspaceFilters) generated.MissionControlWorkspaceFilters {
	if item == nil {
		return generated.MissionControlWorkspaceFilters{}
	}
	return generated.MissionControlWorkspaceFilters{
		OpenScope:       generated.MissionControlWorkspaceFiltersOpenScope(item.GetOpenScope()),
		AssignmentScope: generated.MissionControlWorkspaceFiltersAssignmentScope(item.GetAssignmentScope()),
		StatePreset:     generated.MissionControlWorkspaceFiltersStatePreset(item.GetStatePreset()),
		Search:          cast.OptionalTrimmedString(item.Search),
	}
}

func missionControlWorkspaceSummary(item *controlplanev1.MissionControlWorkspaceSummary) generated.MissionControlWorkspaceSummary {
	summary := generated.MissionControlWorkspaceSummary{}
	if item == nil {
		return summary
	}
	summary.RootCount = item.GetRootCount()
	summary.NodeCount = item.GetNodeCount()
	summary.BlockingGapCount = item.GetBlockingGapCount()
	summary.WarningGapCount = item.GetWarningGapCount()
	summary.RecentClosedContextCount = item.GetRecentClosedContextCount()
	summary.WorkingCount = item.GetWorkingCount()
	summary.WaitingCount = item.GetWaitingCount()
	summary.BlockedCount = item.GetBlockedCount()
	summary.ReviewCount = item.GetReviewCount()
	summary.RecentCriticalUpdatesCount = item.GetRecentCriticalUpdatesCount()
	return summary
}

func missionControlWorkspaceWatermarks(items []*controlplanev1.MissionControlWorkspaceWatermark) []generated.MissionControlWorkspaceWatermark {
	out := make([]generated.MissionControlWorkspaceWatermark, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, generated.MissionControlWorkspaceWatermark{
			WatermarkKind:   generated.MissionControlWorkspaceWatermarkWatermarkKind(item.GetWatermarkKind()),
			Status:          generated.MissionControlWorkspaceWatermarkStatus(item.GetStatus()),
			Summary:         item.GetSummary(),
			ObservedAt:      item.GetObservedAt().AsTime().UTC(),
			WindowStartedAt: protoTimestampPtr(item.GetWindowStartedAt()),
			WindowEndedAt:   protoTimestampPtr(item.GetWindowEndedAt()),
		})
	}
	return out
}

func missionControlRootGroups(items []*controlplanev1.MissionControlRootGroup) []generated.MissionControlRootGroup {
	out := make([]generated.MissionControlRootGroup, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, generated.MissionControlRootGroup{
			RootNodeKind:     generated.MissionControlRootGroupRootNodeKind(item.GetRootNodeKind()),
			RootNodePublicId: item.GetRootNodePublicId(),
			RootTitle:        item.GetRootTitle(),
			LatestActivityAt: protoTimestampPtr(item.GetLatestActivityAt()),
			HasBlockingGap:   item.GetHasBlockingGap(),
			NodeRefs:         missionControlNodeRefs(item.GetNodeRefs()),
		})
	}
	return out
}

func missionControlNodes(items []*controlplanev1.MissionControlNode) []generated.MissionControlNode {
	out := make([]generated.MissionControlNode, 0, len(items))
	for _, item := range items {
		out = append(out, missionControlNode(item))
	}
	return out
}

func missionControlNode(item *controlplanev1.MissionControlNode) generated.MissionControlNode {
	if item == nil {
		return generated.MissionControlNode{}
	}
	return generated.MissionControlNode{
		NodeKind:          generated.MissionControlNodeNodeKind(item.GetNodeKind()),
		NodePublicId:      item.GetNodePublicId(),
		Title:             item.GetTitle(),
		VisibilityTier:    generated.MissionControlNodeVisibilityTier(item.GetVisibilityTier()),
		ActiveState:       generated.MissionControlNodeActiveState(item.GetActiveState()),
		ContinuityStatus:  generated.MissionControlNodeContinuityStatus(item.GetContinuityStatus()),
		CoverageClass:     generated.MissionControlNodeCoverageClass(item.GetCoverageClass()),
		RootNodePublicId:  item.GetRootNodePublicId(),
		ColumnIndex:       item.GetColumnIndex(),
		LastActivityAt:    protoTimestampPtr(item.GetLastActivityAt()),
		HasBlockingGap:    item.GetHasBlockingGap(),
		ProviderReference: missionControlProviderReferencePtr(item.GetProviderReference()),
		Badges:            missionControlNodeBadges(item.GetBadges()),
		ProjectionVersion: item.GetProjectionVersion(),
	}
}

func missionControlProviderReferencePtr(item *controlplanev1.MissionControlProviderReference) *generated.MissionControlProviderReference {
	if item == nil {
		return nil
	}
	out := missionControlProviderReference(item)
	return &out
}

func missionControlNodeBadges(items []string) []generated.MissionControlNodeBadges {
	return typedTrimmedStrings[generated.MissionControlNodeBadges](items)
}

func missionControlEdges(items []*controlplanev1.MissionControlEdge) []generated.MissionControlEdge {
	out := make([]generated.MissionControlEdge, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, generated.MissionControlEdge{
			EdgeKind:           generated.MissionControlEdgeEdgeKind(item.GetEdgeKind()),
			SourceNodeKind:     generated.MissionControlEdgeSourceNodeKind(item.GetSourceNodeKind()),
			SourceNodePublicId: item.GetSourceNodePublicId(),
			TargetNodeKind:     generated.MissionControlEdgeTargetNodeKind(item.GetTargetNodeKind()),
			TargetNodePublicId: item.GetTargetNodePublicId(),
			VisibilityTier:     generated.MissionControlEdgeVisibilityTier(item.GetVisibilityTier()),
			SourceOfTruth:      generated.MissionControlEdgeSourceOfTruth(item.GetSourceOfTruth()),
			IsPrimaryPath:      item.GetIsPrimaryPath(),
		})
	}
	return out
}

func missionControlContinuityGaps(items []*controlplanev1.MissionControlContinuityGap) []generated.MissionControlContinuityGap {
	out := make([]generated.MissionControlContinuityGap, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		gap := generated.MissionControlContinuityGap{
			GapId:               item.GetGapId(),
			GapKind:             generated.MissionControlContinuityGapGapKind(item.GetGapKind()),
			Severity:            generated.MissionControlContinuityGapSeverity(item.GetSeverity()),
			Status:              generated.MissionControlContinuityGapStatus(item.GetStatus()),
			SubjectNodeKind:     generated.MissionControlContinuityGapSubjectNodeKind(item.GetSubjectNodeKind()),
			SubjectNodePublicId: item.GetSubjectNodePublicId(),
			DetectedAt:          item.GetDetectedAt().AsTime().UTC(),
			ExpectedStageLabel:  cast.OptionalTrimmedString(item.ExpectedStageLabel),
			ResolutionHint:      cast.OptionalTrimmedString(item.ResolutionHint),
			ResolvedAt:          protoTimestampPtr(item.GetResolvedAt()),
		}
		if expectedNodeKind := strings.TrimSpace(item.GetExpectedNodeKind()); expectedNodeKind != "" {
			value := generated.MissionControlContinuityGapExpectedNodeKind(expectedNodeKind)
			gap.ExpectedNodeKind = &value
		}
		out = append(out, gap)
	}
	return out
}

func missionControlLaunchSurfaces(items []*controlplanev1.MissionControlLaunchSurface) []generated.MissionControlLaunchSurface {
	out := make([]generated.MissionControlLaunchSurface, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		surface := generated.MissionControlLaunchSurface{
			ActionKind:          generated.MissionControlLaunchSurfaceActionKind(item.GetActionKind()),
			Presentation:        generated.MissionControlLaunchSurfacePresentation(item.GetPresentation()),
			ApprovalRequirement: generated.MissionControlLaunchSurfaceApprovalRequirement(item.GetApprovalRequirement()),
			BlockedReason:       cast.OptionalTrimmedString(item.BlockedReason),
		}
		if template := missionControlStageNextStepTemplate(item.GetCommandTemplate()); template != nil {
			surface.CommandTemplate = template
		}
		out = append(out, surface)
	}
	return out
}

func missionControlStageNextStepTemplate(item *controlplanev1.MissionControlStageNextStepTemplate) *generated.MissionControlStageNextStepTemplate {
	if item == nil {
		return nil
	}
	return &generated.MissionControlStageNextStepTemplate{
		ThreadKind:          generated.MissionControlStageNextStepTemplateThreadKind(item.GetThreadKind()),
		ThreadNumber:        item.GetThreadNumber(),
		TargetLabel:         item.GetTargetLabel(),
		RemovedLabels:       requiredStringSlice(item.GetRemovedLabels()),
		DisplayVariant:      cast.OptionalTrimmedString(item.DisplayVariant),
		ApprovalRequirement: generated.MissionControlStageNextStepTemplateApprovalRequirement(item.GetApprovalRequirement()),
		ExpectedGapIds:      append([]int64{}, item.GetExpectedGapIds()...),
	}
}

func missionControlActivityEntries(items []*controlplanev1.MissionControlActivityEntry) []generated.MissionControlActivityEntry {
	out := make([]generated.MissionControlActivityEntry, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, generated.MissionControlActivityEntry{
			EntryId:      item.GetEntryId(),
			NodeKind:     generated.MissionControlActivityEntryNodeKind(item.GetNodeKind()),
			NodePublicId: item.GetNodePublicId(),
			SourceKind:   generated.MissionControlActivityEntrySourceKind(item.GetSourceKind()),
			SourceRef:    item.GetSourceRef(),
			OccurredAt:   item.GetOccurredAt().AsTime().UTC(),
			Summary:      item.GetSummary(),
			BodyMarkdown: cast.OptionalTrimmedString(item.BodyMarkdown),
			ProviderUrl:  cast.OptionalTrimmedString(item.ProviderUrl),
			IsReadOnly:   item.GetIsReadOnly(),
		})
	}
	return out
}

func missionControlNodeRefs(items []*controlplanev1.MissionControlNodeRef) []generated.MissionControlNodeRef {
	return mapNonNil(items, func(item *controlplanev1.MissionControlNodeRef) (generated.MissionControlNodeRef, bool) {
		if item == nil {
			return generated.MissionControlNodeRef{}, false
		}
		return generated.MissionControlNodeRef{
			NodeKind:     generated.MissionControlNodeRefNodeKind(item.GetNodeKind()),
			NodePublicId: item.GetNodePublicId(),
		}, true
	})
}

func missionControlNodeDetailsPayload(item *controlplanev1.MissionControlNodeDetails) (generated.MissionControlNodeDetailsPayload, error) {
	var out generated.MissionControlNodeDetailsPayload
	switch {
	case item.GetDiscussion() != nil:
		err := out.FromMissionControlDiscussionNodeDetails(generated.MissionControlDiscussionNodeDetails{
			DiscussionKind:          item.GetDiscussion().GetDiscussionKind(),
			Status:                  item.GetDiscussion().GetStatus(),
			Author:                  item.GetDiscussion().GetAuthor(),
			ParticipantCount:        item.GetDiscussion().GetParticipantCount(),
			LatestCommentExcerpt:    item.GetDiscussion().GetLatestCommentExcerpt(),
			FormalizationTargetRefs: missionControlNodeRefs(item.GetDiscussion().GetFormalizationTargetRefs()),
		})
		return out, err
	case item.GetWorkItem() != nil:
		err := out.FromMissionControlWorkItemNodeDetails(generated.MissionControlWorkItemNodeDetails{
			RepositoryFullName: item.GetWorkItem().GetRepositoryFullName(),
			IssueNumber:        item.GetWorkItem().GetIssueNumber(),
			StageLabel:         item.GetWorkItem().GetStageLabel(),
			Labels:             requiredStringSlice(item.GetWorkItem().GetLabels()),
			Assignees:          requiredStringSlice(item.GetWorkItem().GetAssignees()),
			LastProviderSyncAt: protoTimestampPtr(item.GetWorkItem().GetLastProviderSyncAt()),
			LinkedRunRefs:      missionControlNodeRefs(item.GetWorkItem().GetLinkedRunRefs()),
			LinkedFollowUpRefs: missionControlNodeRefs(item.GetWorkItem().GetLinkedFollowUpRefs()),
		})
		return out, err
	case item.GetRun() != nil:
		err := out.FromMissionControlRunNodeDetails(generated.MissionControlRunNodeDetails{
			RunId:                 item.GetRun().GetRunId(),
			AgentKey:              item.GetRun().GetAgentKey(),
			RunStatus:             item.GetRun().GetRunStatus(),
			RuntimeMode:           item.GetRun().GetRuntimeMode(),
			TriggerLabel:          item.GetRun().GetTriggerLabel(),
			BuildRef:              item.GetRun().GetBuildRef(),
			CandidateNamespace:    item.GetRun().GetCandidateNamespace(),
			StartedAt:             protoTimestampPtr(item.GetRun().GetStartedAt()),
			FinishedAt:            protoTimestampPtr(item.GetRun().GetFinishedAt()),
			LinkedPullRequestRefs: missionControlNodeRefs(item.GetRun().GetLinkedPullRequestRefs()),
			ProducedIssueRefs:     missionControlNodeRefs(item.GetRun().GetProducedIssueRefs()),
		})
		return out, err
	case item.GetPullRequest() != nil:
		err := out.FromMissionControlPullRequestNodeDetails(generated.MissionControlPullRequestNodeDetails{
			RepositoryFullName: item.GetPullRequest().GetRepositoryFullName(),
			PullRequestNumber:  item.GetPullRequest().GetPullRequestNumber(),
			BranchHead:         item.GetPullRequest().GetBranchHead(),
			BranchBase:         item.GetPullRequest().GetBranchBase(),
			MergeState:         item.GetPullRequest().GetMergeState(),
			ReviewDecision:     item.GetPullRequest().GetReviewDecision(),
			ChecksSummary:      item.GetPullRequest().GetChecksSummary(),
			LinkedIssueRefs:    missionControlNodeRefs(item.GetPullRequest().GetLinkedIssueRefs()),
			LinkedRunRef:       missionControlNodeRefPtr(item.GetPullRequest().GetLinkedRunRef()),
		})
		return out, err
	default:
		return generated.MissionControlNodeDetailsPayload{}, fmt.Errorf("mission control node detail payload variant is missing")
	}
}

func missionControlNodeRefPtr(item *controlplanev1.MissionControlNodeRef) *generated.MissionControlNodeRef {
	if item == nil {
		return nil
	}
	out := generated.MissionControlNodeRef{
		NodeKind:     generated.MissionControlNodeRefNodeKind(item.GetNodeKind()),
		NodePublicId: item.GetNodePublicId(),
	}
	return &out
}

func missionControlEntityDetailsPayload(item *controlplanev1.MissionControlEntityDetails) (generated.MissionControlEntityDetailsPayload, error) {
	var out generated.MissionControlEntityDetailsPayload
	switch {
	case item.GetWorkItem() != nil:
		err := out.FromWorkItemDetailsPayload(generated.WorkItemDetailsPayload{
			RepositoryFullName: cast.TrimmedStringPtr(item.GetWorkItem().RepositoryFullName),
			IssueNumber:        item.GetWorkItem().GetIssueNumber(),
			IssueUrl:           cast.OptionalTrimmedString(item.GetWorkItem().IssueUrl),
			LastRunId:          cast.OptionalTrimmedString(item.GetWorkItem().LastRunId),
			LastStatus:         cast.OptionalTrimmedString(item.GetWorkItem().LastStatus),
			TriggerKind:        cast.OptionalTrimmedString(item.GetWorkItem().TriggerKind),
			WorkItemType:       cast.OptionalTrimmedString(item.GetWorkItem().WorkItemType),
			StageLabel:         cast.OptionalTrimmedString(item.GetWorkItem().StageLabel),
			Labels:             stringSlicePtr(item.GetWorkItem().GetLabels()),
			Owner:              cast.OptionalTrimmedString(item.GetWorkItem().Owner),
			Assignees:          stringSlicePtr(item.GetWorkItem().GetAssignees()),
			LastProviderSyncAt: protoTimestampPtr(item.GetWorkItem().GetLastProviderSyncAt()),
		})
		return out, err
	case item.GetDiscussion() != nil:
		err := out.FromDiscussionDetailsPayload(generated.DiscussionDetailsPayload{
			DiscussionKind:       item.GetDiscussion().GetDiscussionKind(),
			Status:               cast.OptionalTrimmedString(item.GetDiscussion().Status),
			Author:               cast.OptionalTrimmedString(item.GetDiscussion().Author),
			ParticipantCount:     int32Ptr(item.GetDiscussion().GetParticipantCount()),
			LatestCommentExcerpt: cast.OptionalTrimmedString(item.GetDiscussion().LatestCommentExcerpt),
			FormalizationTarget:  cast.OptionalTrimmedString(item.GetDiscussion().FormalizationTarget),
		})
		return out, err
	case item.GetPullRequest() != nil:
		err := out.FromPullRequestDetailsPayload(generated.PullRequestDetailsPayload{
			RepositoryFullName: cast.TrimmedStringPtr(item.GetPullRequest().RepositoryFullName),
			PullRequestNumber:  item.GetPullRequest().GetPullRequestNumber(),
			PullRequestUrl:     cast.OptionalTrimmedString(item.GetPullRequest().PullRequestUrl),
			LastRunId:          cast.OptionalTrimmedString(item.GetPullRequest().LastRunId),
			LastStatus:         cast.OptionalTrimmedString(item.GetPullRequest().LastStatus),
			BranchHead:         cast.OptionalTrimmedString(item.GetPullRequest().BranchHead),
			BranchBase:         cast.OptionalTrimmedString(item.GetPullRequest().BranchBase),
			MergeState:         cast.OptionalTrimmedString(item.GetPullRequest().MergeState),
			ReviewDecision:     cast.OptionalTrimmedString(item.GetPullRequest().ReviewDecision),
			ChecksSummary:      cast.OptionalTrimmedString(item.GetPullRequest().ChecksSummary),
			LinkedIssueRefs:    stringSlicePtr(item.GetPullRequest().GetLinkedIssueRefs()),
		})
		return out, err
	case item.GetAgent() != nil:
		err := out.FromAgentDetailsPayload(generated.AgentDetailsPayload{
			AgentKey:          item.GetAgent().GetAgentKey(),
			RunStatus:         cast.OptionalTrimmedString(item.GetAgent().RunStatus),
			RuntimeMode:       cast.OptionalTrimmedString(item.GetAgent().RuntimeMode),
			WaitingReason:     cast.OptionalTrimmedString(item.GetAgent().WaitingReason),
			ActiveRunId:       cast.OptionalTrimmedString(item.GetAgent().ActiveRunId),
			LastHeartbeatAt:   protoTimestampPtr(item.GetAgent().GetLastHeartbeatAt()),
			LastRunRepository: cast.OptionalTrimmedString(item.GetAgent().LastRunRepository),
		})
		return out, err
	default:
		return generated.MissionControlEntityDetailsPayload{}, fmt.Errorf("mission control detail payload variant is missing")
	}
}

func missionControlCommandApproval(item *controlplanev1.MissionControlCommandApproval) *generated.MissionControlCommandApproval {
	if item == nil {
		return nil
	}
	return &generated.MissionControlCommandApproval{
		ApprovalState:     generated.MissionControlCommandApprovalApprovalState(item.GetApprovalState()),
		ApprovalRequestId: cast.OptionalTrimmedString(item.ApprovalRequestId),
		RequestedAt:       protoTimestampPtr(item.GetRequestedAt()),
		DecidedAt:         protoTimestampPtr(item.GetDecidedAt()),
		ApproverActorId:   cast.OptionalTrimmedString(item.ApproverActorId),
	}
}

func missionControlEntityRefs(items []*controlplanev1.MissionControlEntityRef) []generated.MissionControlEntityRef {
	refs := mapNonNil(items, func(item *controlplanev1.MissionControlEntityRef) (generated.MissionControlEntityRef, bool) {
		if item == nil {
			return generated.MissionControlEntityRef{}, false
		}
		ref := generated.MissionControlEntityRef{}
		ref.EntityKind = generated.MissionControlEntityRefEntityKind(item.GetEntityKind())
		ref.EntityPublicId = item.GetEntityPublicId()
		return ref, true
	})
	return refs
}

func trimStringSlice(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func requiredStringSlice(items []string) []string {
	normalized := trimStringSlice(items)
	if normalized == nil {
		return []string{}
	}
	return normalized
}

func stringSlicePtr(items []string) *[]string {
	normalized := trimStringSlice(items)
	if len(normalized) == 0 {
		return nil
	}
	return &normalized
}

func int32Ptr(value int32) *int32 {
	if value == 0 {
		return nil
	}
	result := value
	return &result
}

func timestampPtr(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	result := value.UTC()
	return &result
}

func protoTimestampPtr(value *timestamppb.Timestamp) *time.Time {
	if value == nil {
		return nil
	}
	return timestampPtr(value.AsTime())
}

func typedTrimmedStrings[T ~string](items []string) []T {
	out := make([]T, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			out = append(out, T(trimmed))
		}
	}
	return out
}

func mapNonNil[T any, R any](items []T, mapFn func(T) (R, bool)) []R {
	out := make([]R, 0, len(items))
	for _, item := range items {
		mapped, ok := mapFn(item)
		if ok {
			out = append(out, mapped)
		}
	}
	return out
}

func newMissionControlCommandRequestBase[T ~string](
	commandKind string,
	projectID string,
	businessIntentKey string,
	expectedProjectionVersion *int64,
	targetEntityKind *T,
	targetEntityPublicID *string,
	correlationID string,
	requestedAt time.Time,
) *controlplanev1.SubmitMissionControlCommandRequest {
	req := &controlplanev1.SubmitMissionControlCommandRequest{
		CommandKind:       strings.TrimSpace(commandKind),
		ProjectId:         strings.TrimSpace(projectID),
		BusinessIntentKey: strings.TrimSpace(businessIntentKey),
		CorrelationId:     strings.TrimSpace(correlationID),
	}
	if expectedProjectionVersion != nil {
		req.ExpectedProjectionVersion = *expectedProjectionVersion
	}
	if trimmed := strings.TrimSpace(derefString(targetEntityPublicID)); trimmed != "" {
		req.TargetEntityPublicId = &trimmed
	}
	if trimmed := strings.TrimSpace(fmt.Sprint(derefStringEnum(targetEntityKind))); trimmed != "" && trimmed != "<nil>" {
		req.TargetEntityKind = &trimmed
	}
	if !requestedAt.IsZero() {
		req.RequestedAt = timestamppb.New(requestedAt.UTC())
	}
	return req
}

func missionControlEntityRefsToProtoRequest(items *[]generated.MissionControlEntityRef) []*controlplanev1.MissionControlEntityRef {
	if items == nil || len(*items) == 0 {
		return nil
	}
	out := make([]*controlplanev1.MissionControlEntityRef, 0, len(*items))
	for _, item := range *items {
		out = append(out, &controlplanev1.MissionControlEntityRef{
			EntityKind:     string(item.EntityKind),
			EntityPublicId: item.EntityPublicId,
		})
	}
	return out
}

func optionalStringPtr[T ~string](value *T) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(string(*value))
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func derefStringSlice(value *[]string) []string {
	if value == nil {
		return nil
	}
	return trimStringSlice(*value)
}

func derefStringEnum[T ~string](value *T) T {
	if value == nil {
		var zero T
		return zero
	}
	return *value
}
