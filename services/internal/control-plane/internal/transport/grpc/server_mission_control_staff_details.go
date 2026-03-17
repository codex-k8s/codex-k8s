package grpc

import (
	"encoding/json"
	"strings"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	missioncontroldomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/missioncontrol"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const missionControlActionBlockedReasonDegraded = "entity is degraded; retry sync or refresh before using this action"

func missionControlProviderDeepLinks(entity missioncontroldomain.Entity) []*controlplanev1.MissionControlProviderDeepLink {
	url := strings.TrimSpace(entity.ProviderURL)
	if url == "" {
		return nil
	}
	actionKind := "provider.open_issue"
	switch entity.EntityKind {
	case enumtypes.MissionControlEntityKindPullRequest:
		actionKind = "provider.open_pr"
	case enumtypes.MissionControlEntityKindDiscussion:
		actionKind = "provider.reply_comment"
	}
	return []*controlplanev1.MissionControlProviderDeepLink{{
		ActionKind: actionKind,
		Url:        url,
		Reason:     "not_in_mvp_inline_scope",
	}}
}

func missionControlApplyDetailPayload(out *controlplanev1.MissionControlEntityDetails, details missioncontroldomain.EntityDetails) {
	entity := details.Entity
	switch entity.EntityKind {
	case enumtypes.MissionControlEntityKindWorkItem:
		payload := &controlplanev1.MissionControlWorkItemDetailsPayload{}
		var decoded valuetypes.MissionControlWorkItemProjectionPayload
		if err := json.Unmarshal(entity.DetailPayloadJSON, &decoded); err == nil {
			payload.RepositoryFullName = strings.TrimSpace(decoded.RepositoryFullName)
			payload.IssueNumber = decoded.IssueNumber
			payload.IssueUrl = stringPtrOrNil(strings.TrimSpace(decoded.IssueURL))
			payload.LastRunId = stringPtrOrNil(strings.TrimSpace(decoded.LastRunID))
			payload.LastStatus = stringPtrOrNil(strings.TrimSpace(decoded.LastStatus))
			payload.TriggerKind = stringPtrOrNil(strings.TrimSpace(decoded.TriggerKind))
			payload.WorkItemType = stringPtrOrNil(strings.TrimSpace(decoded.WorkItemType))
			payload.StageLabel = stringPtrOrNil(strings.TrimSpace(decoded.StageLabel))
			payload.Labels = append([]string(nil), decoded.Labels...)
			payload.Owner = stringPtrOrNil(strings.TrimSpace(decoded.Owner))
			payload.Assignees = append([]string(nil), decoded.Assignees...)
			if decoded.LastProviderSyncAt != nil {
				payload.LastProviderSyncAt = timestamppb.New(decoded.LastProviderSyncAt.UTC())
			}
		}
		if payload.LastProviderSyncAt == nil && entity.ProviderUpdatedAt != nil {
			payload.LastProviderSyncAt = timestamppb.New(entity.ProviderUpdatedAt.UTC())
		}
		out.DetailPayload = &controlplanev1.MissionControlEntityDetails_WorkItem{WorkItem: payload}
	case enumtypes.MissionControlEntityKindPullRequest:
		payload := &controlplanev1.MissionControlPullRequestDetailsPayload{}
		var decoded valuetypes.MissionControlPullRequestProjectionPayload
		if err := json.Unmarshal(entity.DetailPayloadJSON, &decoded); err == nil {
			payload.RepositoryFullName = strings.TrimSpace(decoded.RepositoryFullName)
			payload.PullRequestNumber = decoded.PullRequestNumber
			payload.PullRequestUrl = stringPtrOrNil(strings.TrimSpace(decoded.PullRequestURL))
			payload.LastRunId = stringPtrOrNil(strings.TrimSpace(decoded.LastRunID))
			payload.LastStatus = stringPtrOrNil(strings.TrimSpace(decoded.LastStatus))
			payload.BranchHead = stringPtrOrNil(strings.TrimSpace(decoded.BranchHead))
			payload.BranchBase = stringPtrOrNil(strings.TrimSpace(decoded.BranchBase))
			payload.MergeState = stringPtrOrNil(strings.TrimSpace(decoded.MergeState))
			payload.ReviewDecision = stringPtrOrNil(strings.TrimSpace(decoded.ReviewDecision))
			payload.ChecksSummary = stringPtrOrNil(strings.TrimSpace(decoded.ChecksSummary))
			payload.LinkedIssueRefs = mergeMissionControlStringLists(decoded.LinkedIssueRefs, missionControlLinkedIssueRefs(details.Relations))
		}
		if len(payload.LinkedIssueRefs) == 0 {
			payload.LinkedIssueRefs = missionControlLinkedIssueRefs(details.Relations)
		}
		out.DetailPayload = &controlplanev1.MissionControlEntityDetails_PullRequest{PullRequest: payload}
	case enumtypes.MissionControlEntityKindAgent:
		payload := &controlplanev1.MissionControlAgentDetailsPayload{}
		var decoded valuetypes.MissionControlAgentProjectionPayload
		if err := json.Unmarshal(entity.DetailPayloadJSON, &decoded); err == nil {
			payload.AgentKey = strings.TrimSpace(decoded.AgentKey)
			payload.RunStatus = stringPtrOrNil(strings.TrimSpace(decoded.LastStatus))
			payload.RuntimeMode = stringPtrOrNil(strings.TrimSpace(decoded.RuntimeMode))
			payload.WaitingReason = stringPtrOrNil(strings.TrimSpace(decoded.WaitingReason))
			payload.ActiveRunId = stringPtrOrNil(strings.TrimSpace(decoded.LastRunID))
			payload.LastRunRepository = stringPtrOrNil(strings.TrimSpace(decoded.LastRunRepository))
			if decoded.LastHeartbeatAt != nil {
				payload.LastHeartbeatAt = timestamppb.New(decoded.LastHeartbeatAt.UTC())
			}
		}
		out.DetailPayload = &controlplanev1.MissionControlEntityDetails_Agent{Agent: payload}
	case enumtypes.MissionControlEntityKindDiscussion:
		payload := &controlplanev1.MissionControlDiscussionDetailsPayload{}
		var decoded valuetypes.MissionControlDiscussionProjectionPayload
		if err := json.Unmarshal(entity.DetailPayloadJSON, &decoded); err == nil {
			payload.DiscussionKind = stringPtrOrNil(strings.TrimSpace(decoded.DiscussionKind))
			payload.Status = stringPtrOrNil(strings.TrimSpace(decoded.Status))
			payload.Author = stringPtrOrNil(strings.TrimSpace(decoded.Author))
			payload.ParticipantCount = decoded.ParticipantCount
			payload.LatestCommentExcerpt = stringPtrOrNil(strings.TrimSpace(decoded.LatestCommentExcerpt))
			payload.FormalizationTarget = stringPtrOrNil(strings.TrimSpace(decoded.FormalizationTarget))
		}
		if payload.GetDiscussionKind() == "" {
			payload.DiscussionKind = stringPtrOrNil("discussion")
		}
		if payload.GetStatus() == "" {
			payload.Status = stringPtrOrNil(strings.TrimSpace(string(entity.ActiveState)))
		}
		if payload.GetLatestCommentExcerpt() == "" && len(details.Timeline) > 0 {
			payload.LatestCommentExcerpt = stringPtrOrNil(strings.TrimSpace(details.Timeline[0].Summary))
		}
		out.DetailPayload = &controlplanev1.MissionControlEntityDetails_Discussion{Discussion: payload}
	}
}

func missionControlAllowedActions(entity missioncontroldomain.Entity) []*controlplanev1.MissionControlAllowedAction {
	blockedReason := ""
	if entity.SyncStatus == enumtypes.MissionControlSyncStatusDegraded || entity.SyncStatus == enumtypes.MissionControlSyncStatusFailed {
		blockedReason = missionControlActionBlockedReasonDegraded
	}
	stageNextStepApproval := string(missioncontroldomain.StageNextStepApprovalRequirement(entity))

	out := make([]*controlplanev1.MissionControlAllowedAction, 0, 4)
	switch entity.EntityKind {
	case enumtypes.MissionControlEntityKindWorkItem:
		out = append(out, missionControlAllowedAction("discussion.create", "secondary", false, "none", blockedReason))
		out = append(out, missionControlAllowedAction("stage.next_step.execute", "primary", false, stageNextStepApproval, blockedReason))
	case enumtypes.MissionControlEntityKindDiscussion:
		out = append(out, missionControlAllowedAction("work_item.create", "primary", false, "none", blockedReason))
		out = append(out, missionControlAllowedAction("discussion.formalize", "secondary", false, "none", blockedReason))
		out = append(out, missionControlAllowedAction("stage.next_step.execute", "secondary", false, stageNextStepApproval, blockedReason))
	case enumtypes.MissionControlEntityKindPullRequest:
		out = append(out, missionControlAllowedAction("stage.next_step.execute", "primary", false, stageNextStepApproval, blockedReason))
	}
	if entity.SyncStatus == enumtypes.MissionControlSyncStatusDegraded || entity.SyncStatus == enumtypes.MissionControlSyncStatusFailed {
		out = append(out, missionControlAllowedAction("command.retry_sync", "secondary", true, "none", ""))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func missionControlAllowedAction(
	actionKind string,
	presentation string,
	allowedWhenDegraded bool,
	approvalRequirement string,
	blockedReason string,
) *controlplanev1.MissionControlAllowedAction {
	out := &controlplanev1.MissionControlAllowedAction{
		ActionKind:          actionKind,
		Presentation:        presentation,
		AllowedWhenDegraded: allowedWhenDegraded,
		ApprovalRequirement: approvalRequirement,
	}
	if trimmed := strings.TrimSpace(blockedReason); trimmed != "" {
		out.BlockedReason = stringPtrOrNil(trimmed)
	}
	return out
}

func missionControlLinkedIssueRefs(relations []valuetypes.MissionControlRelationView) []string {
	out := make([]string, 0, len(relations))
	seen := make(map[string]struct{}, len(relations))
	for _, relation := range relations {
		for _, ref := range []valuetypes.MissionControlEntityRef{relation.SourceEntityRef, relation.TargetEntityRef} {
			if ref.EntityKind != enumtypes.MissionControlEntityKindWorkItem {
				continue
			}
			publicID := strings.TrimSpace(ref.EntityPublicID)
			if publicID == "" {
				continue
			}
			if _, ok := seen[publicID]; ok {
				continue
			}
			seen[publicID] = struct{}{}
			out = append(out, publicID)
		}
	}
	return out
}

func mergeMissionControlStringLists(primary []string, secondary []string) []string {
	if len(primary) == 0 && len(secondary) == 0 {
		return nil
	}
	out := make([]string, 0, len(primary)+len(secondary))
	seen := make(map[string]struct{}, len(primary)+len(secondary))
	for _, list := range [][]string{primary, secondary} {
		for _, raw := range list {
			item := strings.TrimSpace(raw)
			if item == "" {
				continue
			}
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			out = append(out, item)
		}
	}
	return out
}
