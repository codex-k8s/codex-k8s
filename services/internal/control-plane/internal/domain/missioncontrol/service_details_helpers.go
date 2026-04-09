package missioncontrol

import (
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	nextstepdomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/nextstep"
	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

func (s *Service) enrichEntityDetailsWithWorkspaceContext(details *EntityDetails, graph workspaceGraph) {
	if details == nil {
		return
	}
	entity := details.Entity
	if entity.ID == 0 {
		return
	}

	rootByNode := buildWorkspaceDetailRootMap(graph)
	effectiveContinuity := buildEffectiveContinuityMap(graph)

	details.Node = buildWorkspaceDetailNode(graph, rootByNode, effectiveContinuity, entity.ID)
	details.AdjacentNodes = buildWorkspaceAdjacentNodes(graph, rootByNode, effectiveContinuity, entity.ID)
	details.AdjacentEdges = buildWorkspaceAdjacentEdges(graph, effectiveContinuity, entity.ID)
	details.ContinuityGaps = buildWorkspaceDetailContinuityGaps(graph, entity.ID)
	details.NodeWatermarks = append([]entitytypes.MissionControlWorkspaceWatermark(nil), graph.latestWatermarks...)
	details.LaunchSurfaces = buildWorkspaceLaunchSurfaces(graph, entity, details.ContinuityGaps, s.cfg.NextStepLabels)
}

func buildWorkspaceDetailRootMap(graph workspaceGraph) map[int64]int64 {
	includedIDs := make(map[int64]struct{}, len(graph.entities))
	primaryIDs := make(map[int64]struct{}, len(graph.entities))
	for _, entity := range graph.entities {
		if entity.EntityKind == enumtypes.MissionControlEntityKindAgent {
			continue
		}
		includedIDs[entity.ID] = struct{}{}
		primaryIDs[entity.ID] = struct{}{}
	}
	rootIDs := selectWorkspaceRootIDs(graph, primaryIDs, includedIDs)
	rootByNode := make(map[int64]int64, len(includedIDs))
	for entityID := range includedIDs {
		rootByNode[entityID] = resolveWorkspaceRootID(graph, entityID, rootIDs)
	}
	return rootByNode
}

func buildWorkspaceDetailNode(
	graph workspaceGraph,
	rootByNode map[int64]int64,
	effectiveContinuity map[int64]enumtypes.MissionControlContinuityStatus,
	entityID int64,
) valuetypes.MissionControlWorkspaceNode {
	entity, ok := graph.entityByID[entityID]
	if !ok {
		return valuetypes.MissionControlWorkspaceNode{}
	}
	rootID := rootByNode[entityID]
	rootEntity := graph.entityByID[rootID]
	return valuetypes.MissionControlWorkspaceNode{
		NodeRef: valuetypes.MissionControlEntityRef{
			EntityKind:     entity.EntityKind,
			EntityPublicID: entity.EntityExternalKey,
		},
		Title:             strings.TrimSpace(entity.Title),
		VisibilityTier:    workspaceDetailVisibility(entity),
		ActiveState:       entity.ActiveState,
		ContinuityStatus:  effectiveContinuity[entityID],
		CoverageClass:     entity.CoverageClass,
		ProviderReference: workspaceProviderReference(entity),
		RootNodePublicID:  rootEntity.EntityExternalKey,
		ColumnIndex:       workspaceColumnIndex(entity.EntityKind),
		LastActivityAt:    entityLastActivity(entity),
		HasBlockingGap:    continuityStatusBlocking(effectiveContinuity[entityID]),
		Badges:            workspaceNodeBadges(entity, effectiveContinuity[entityID]),
		ProjectionVersion: entity.ProjectionVersion,
	}
}

func buildWorkspaceAdjacentNodes(
	graph workspaceGraph,
	rootByNode map[int64]int64,
	effectiveContinuity map[int64]enumtypes.MissionControlContinuityStatus,
	entityID int64,
) []valuetypes.MissionControlWorkspaceNode {
	adjacentIDs := make(map[int64]struct{})
	for _, relation := range graph.outgoing[entityID] {
		adjacentIDs[relation.TargetEntityID] = struct{}{}
	}
	for _, relation := range graph.incoming[entityID] {
		adjacentIDs[relation.SourceEntityID] = struct{}{}
	}

	out := make([]valuetypes.MissionControlWorkspaceNode, 0, len(adjacentIDs))
	for adjacentID := range adjacentIDs {
		if adjacentID == entityID {
			continue
		}
		out = append(out, buildWorkspaceDetailNode(graph, rootByNode, effectiveContinuity, adjacentID))
	}
	sort.Slice(out, func(i, j int) bool {
		left := out[i]
		right := out[j]
		if left.ColumnIndex != right.ColumnIndex {
			return left.ColumnIndex < right.ColumnIndex
		}
		return left.NodeRef.EntityPublicID < right.NodeRef.EntityPublicID
	})
	return out
}

func buildWorkspaceAdjacentEdges(
	graph workspaceGraph,
	effectiveContinuity map[int64]enumtypes.MissionControlContinuityStatus,
	entityID int64,
) []valuetypes.MissionControlWorkspaceEdge {
	edgeMap := make(map[string]valuetypes.MissionControlWorkspaceEdge)
	for _, relation := range graph.outgoing[entityID] {
		appendWorkspaceAdjacentEdge(edgeMap, graph, effectiveContinuity, relation)
	}
	for _, relation := range graph.incoming[entityID] {
		appendWorkspaceAdjacentEdge(edgeMap, graph, effectiveContinuity, relation)
	}

	out := make([]valuetypes.MissionControlWorkspaceEdge, 0, len(edgeMap))
	for _, item := range edgeMap {
		out = append(out, item)
	}
	sortWorkspaceEdges(out)
	return out
}

func appendWorkspaceAdjacentEdge(
	edgeMap map[string]valuetypes.MissionControlWorkspaceEdge,
	graph workspaceGraph,
	effectiveContinuity map[int64]enumtypes.MissionControlContinuityStatus,
	relation Relation,
) {
	sourceEntity, sourceOK := graph.entityByID[relation.SourceEntityID]
	targetEntity, targetOK := graph.entityByID[relation.TargetEntityID]
	if !sourceOK || !targetOK {
		return
	}
	key := strings.Join([]string{
		string(relation.RelationKind),
		sourceEntity.EntityExternalKey,
		targetEntity.EntityExternalKey,
		string(relation.SourceKind),
	}, ":")
	visibility := enumtypes.MissionControlWorkspaceVisibilityTierSecondaryDimmed
	if !continuityStatusBlocking(effectiveContinuity[relation.SourceEntityID]) &&
		!continuityStatusBlocking(effectiveContinuity[relation.TargetEntityID]) {
		visibility = enumtypes.MissionControlWorkspaceVisibilityTierPrimary
	}
	edgeMap[key] = workspaceEdgeForRelation(sourceEntity, targetEntity, relation, visibility)
}

func buildWorkspaceDetailContinuityGaps(graph workspaceGraph, entityID int64) []valuetypes.MissionControlContinuityGapView {
	relevantGapIDs := previewRelevantGapIDs(graph, entityID)
	if len(relevantGapIDs) == 0 {
		return nil
	}

	out := make([]valuetypes.MissionControlContinuityGapView, 0, len(relevantGapIDs))
	for _, gapID := range relevantGapIDs {
		gap := openGapByID(graph, gapID)
		if gap.ID == 0 {
			continue
		}
		subjectEntity, ok := graph.entityByID[gap.SubjectEntityID]
		if !ok {
			continue
		}
		out = append(out, valuetypes.MissionControlContinuityGapView{
			GapID:    gap.ID,
			GapKind:  gap.GapKind,
			Severity: gap.Severity,
			Status:   gap.Status,
			SubjectNodeRef: valuetypes.MissionControlEntityRef{
				EntityKind:     subjectEntity.EntityKind,
				EntityPublicID: subjectEntity.EntityExternalKey,
			},
			ExpectedNodeKind:   gap.ExpectedEntityKind,
			ExpectedStageLabel: strings.TrimSpace(gap.ExpectedStageLabel),
			DetectedAt:         gap.DetectedAt.UTC(),
			ResolvedAt:         normalizeTimePtr(gap.ResolvedAt),
			ResolutionHint:     strings.TrimSpace(gap.ResolutionHint),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		left := out[i]
		right := out[j]
		if left.DetectedAt.Equal(right.DetectedAt) {
			return left.GapID > right.GapID
		}
		return left.DetectedAt.After(right.DetectedAt)
	})
	return out
}

func buildWorkspaceLaunchSurfaces(
	graph workspaceGraph,
	entity Entity,
	gaps []valuetypes.MissionControlContinuityGapView,
	labels nextstepdomain.Labels,
) []valuetypes.MissionControlLaunchSurface {
	surfaces := make([]valuetypes.MissionControlLaunchSurface, 0, 5)

	if template := buildWorkspaceStageNextStepTemplate(graph, entity, gaps, labels); template != nil {
		surface := valuetypes.MissionControlLaunchSurface{
			ActionKind:          "preview_next_stage",
			Presentation:        "primary",
			ApprovalRequirement: template.ApprovalRequirement,
			CommandTemplate:     template,
		}
		if len(template.ExpectedGapIDs) > 0 {
			surface.BlockedReason = "open_continuity_gaps"
		}
		if entity.EntityKind == enumtypes.MissionControlEntityKindDiscussion || entity.EntityKind == enumtypes.MissionControlEntityKindRun {
			surface.Presentation = "secondary"
		}
		surfaces = append(surfaces, surface)
	}

	if entity.ProviderURL != "" {
		surfaces = append(surfaces, valuetypes.MissionControlLaunchSurface{
			ActionKind:          "open_provider_context",
			Presentation:        "link",
			ApprovalRequirement: enumtypes.MissionControlApprovalRequirementNone,
		})
	}

	if entity.EntityKind == enumtypes.MissionControlEntityKindRun {
		surfaces = append(surfaces, valuetypes.MissionControlLaunchSurface{
			ActionKind:          "inspect_run_context",
			Presentation:        "secondary",
			ApprovalRequirement: enumtypes.MissionControlApprovalRequirementNone,
		})
	}

	if len(producedPullRequestEntities(graph, entity.ID)) > 0 || linkedPullRequestRef(entity) != "" {
		surfaces = append(surfaces, valuetypes.MissionControlLaunchSurface{
			ActionKind:          "open_linked_pull_request",
			Presentation:        "link",
			ApprovalRequirement: enumtypes.MissionControlApprovalRequirementNone,
		})
	}

	if len(linkedFollowUpRefs(graph, entity)) > 0 {
		surfaces = append(surfaces, valuetypes.MissionControlLaunchSurface{
			ActionKind:          "open_linked_follow_up_issue",
			Presentation:        "link",
			ApprovalRequirement: enumtypes.MissionControlApprovalRequirementNone,
		})
	}

	return surfaces
}

func buildWorkspaceStageNextStepTemplate(
	graph workspaceGraph,
	entity Entity,
	gaps []valuetypes.MissionControlContinuityGapView,
	labels nextstepdomain.Labels,
) *valuetypes.MissionControlStageNextStepTemplate {
	threadKind, threadNumber, ok := workspaceStageThreadTarget(entity)
	if !ok {
		return nil
	}

	targetLabel := strings.TrimSpace(nextStageLabelForEntity(graph, entity, labels))
	if targetLabel == "" {
		return nil
	}

	template := &valuetypes.MissionControlStageNextStepTemplate{
		ThreadKind:          threadKind,
		ThreadNumber:        threadNumber,
		TargetLabel:         targetLabel,
		DisplayVariant:      "default",
		ApprovalRequirement: StageNextStepApprovalRequirement(entity),
	}
	for _, gap := range gaps {
		if gap.Status == enumtypes.MissionControlGapStatusOpen {
			template.ExpectedGapIDs = append(template.ExpectedGapIDs, gap.GapID)
		}
	}
	return template
}

func nextStageLabelForEntity(graph workspaceGraph, entity Entity, labels nextstepdomain.Labels) string {
	switch entity.EntityKind {
	case enumtypes.MissionControlEntityKindRun:
		return nextStageLabelForRun(entity, labels)
	case enumtypes.MissionControlEntityKindPullRequest:
		for _, relation := range graph.incoming[entity.ID] {
			if relation.RelationKind != enumtypes.MissionControlRelationKindProducedPullRequest {
				continue
			}
			runEntity, ok := graph.entityByID[relation.SourceEntityID]
			if !ok {
				continue
			}
			return nextStageLabelForRun(runEntity, labels)
		}
	case enumtypes.MissionControlEntityKindWorkItem:
		payload, ok := decodeWorkItemPayload(entity)
		if !ok {
			return ""
		}
		if nextLabel, ok := labels.NextMainPathRunLabel(strings.TrimSpace(payload.StageLabel)); ok {
			return nextLabel
		}
	case enumtypes.MissionControlEntityKindDiscussion:
		for _, relation := range graph.outgoing[entity.ID] {
			if relation.RelationKind != enumtypes.MissionControlRelationKindFormalizedFrom {
				continue
			}
			workItem, ok := graph.entityByID[relation.TargetEntityID]
			if !ok {
				continue
			}
			return nextStageLabelForEntity(graph, workItem, labels)
		}
	}
	return ""
}

func workspaceStageThreadTarget(entity Entity) (string, int, bool) {
	switch entity.EntityKind {
	case enumtypes.MissionControlEntityKindWorkItem:
		payload, ok := decodeWorkItemPayload(entity)
		if !ok || payload.IssueNumber <= 0 {
			return "", 0, false
		}
		return "issue", int(payload.IssueNumber), true
	case enumtypes.MissionControlEntityKindPullRequest:
		payload, ok := decodePullRequestPayload(entity)
		if !ok || payload.PullRequestNumber <= 0 {
			return "", 0, false
		}
		return "pull_request", int(payload.PullRequestNumber), true
	case enumtypes.MissionControlEntityKindRun:
		payload, ok := decodeRunPayload(entity)
		if !ok {
			return "", 0, false
		}
		if ref := strings.TrimSpace(payload.PullRequestRef); ref != "" {
			if number := parseTrailingNumber(ref); number > 0 {
				return "pull_request", number, true
			}
		}
		if ref := strings.TrimSpace(payload.IssueRef); ref != "" {
			if number := parseTrailingNumber(ref); number > 0 {
				return "issue", number, true
			}
		}
	}
	return "", 0, false
}

func linkedPullRequestRef(entity Entity) string {
	payload, ok := decodeRunPayload(entity)
	if !ok {
		return ""
	}
	return strings.TrimSpace(payload.PullRequestRef)
}

func linkedFollowUpRefs(graph workspaceGraph, entity Entity) []string {
	out := make([]string, 0)
	seen := make(map[string]struct{})
	appendRef := func(raw string) {
		value := strings.TrimSpace(raw)
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}

	switch entity.EntityKind {
	case enumtypes.MissionControlEntityKindPullRequest:
		payload, ok := decodePullRequestPayload(entity)
		if ok {
			for _, ref := range payload.LinkedIssueRefs {
				appendRef(ref)
			}
		}
	case enumtypes.MissionControlEntityKindRun:
		for _, relation := range graph.outgoing[entity.ID] {
			if relation.RelationKind != enumtypes.MissionControlRelationKindProducedPullRequest {
				continue
			}
			pullRequest, ok := graph.entityByID[relation.TargetEntityID]
			if !ok {
				continue
			}
			for _, ref := range linkedFollowUpRefs(graph, pullRequest) {
				appendRef(ref)
			}
		}
	}
	return out
}

func workspaceProviderReference(entity Entity) *valuetypes.MissionControlProviderReferenceView {
	if entity.ProviderKind == "" {
		return nil
	}
	return &valuetypes.MissionControlProviderReferenceView{
		Provider:   entity.ProviderKind,
		ExternalID: strings.TrimSpace(entity.EntityExternalKey),
		URL:        strings.TrimSpace(entity.ProviderURL),
	}
}

func workspaceNodeBadges(entity Entity, continuityStatus enumtypes.MissionControlContinuityStatus) []string {
	badges := make([]string, 0, 3)
	if continuityStatusBlocking(continuityStatus) {
		badges = append(badges, "continuity_gap")
	}
	if continuityStatus == enumtypes.MissionControlContinuityStatusStaleProvider {
		badges = append(badges, "provider_stale")
	}
	if entity.CoverageClass == enumtypes.MissionControlCoverageClassRecentClosedContext {
		badges = append(badges, "recent_closed_context")
	}
	if entity.ActiveState == enumtypes.MissionControlActiveStateWaiting {
		badges = append(badges, "waiting_mcp")
	}
	if entity.ActiveState == enumtypes.MissionControlActiveStateReview {
		badges = append(badges, "review_required")
	}
	return slices.Compact(badges)
}

func workspaceDetailVisibility(entity Entity) enumtypes.MissionControlWorkspaceVisibilityTier {
	if entity.CoverageClass == enumtypes.MissionControlCoverageClassOpenPrimary &&
		entity.ActiveState != enumtypes.MissionControlActiveStateArchived {
		return enumtypes.MissionControlWorkspaceVisibilityTierPrimary
	}
	return enumtypes.MissionControlWorkspaceVisibilityTierSecondaryDimmed
}

func normalizeTimePtr(value *time.Time) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}

func parseTrailingNumber(ref string) int {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return 0
	}
	hashIndex := strings.LastIndex(ref, "#")
	slashIndex := strings.LastIndex(ref, "/")
	start := hashIndex
	if slashIndex > start {
		start = slashIndex
	}
	if start < 0 || start+1 >= len(ref) {
		return 0
	}
	number, err := strconv.Atoi(strings.TrimSpace(ref[start+1:]))
	if err != nil || number <= 0 {
		return 0
	}
	return number
}
