package casters

import (
	"fmt"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/cast"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/generated"
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
	out := make([]generated.MissionControlEntityCardBadges, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			out = append(out, generated.MissionControlEntityCardBadges(trimmed))
		}
	}
	return out
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
	out := make([]generated.MissionControlEntityRef, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, generated.MissionControlEntityRef{
			EntityKind:     generated.MissionControlEntityRefEntityKind(item.GetEntityKind()),
			EntityPublicId: item.GetEntityPublicId(),
		})
	}
	return out
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
