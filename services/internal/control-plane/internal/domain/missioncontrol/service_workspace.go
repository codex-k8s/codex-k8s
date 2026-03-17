package missioncontrol

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	nextstepdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/nextstep"
	missioncontrolrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/missioncontrol"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

// RefreshWorkspaceProjection recalculates continuity gaps and effective watermarks under control-plane ownership.
func (s *Service) RefreshWorkspaceProjection(ctx context.Context, params WorkspaceRefreshParams) (WorkspaceProjectionSummary, error) {
	if err := s.ensureDomainWriteAllowed(); err != nil {
		return WorkspaceProjectionSummary{}, err
	}

	projectID := strings.TrimSpace(params.ProjectID)
	if projectID == "" {
		return WorkspaceProjectionSummary{}, errs.Validation{Field: "project_id", Msg: "is required"}
	}
	observedAt := params.ObservedAt
	if observedAt.IsZero() {
		observedAt = s.now()
	}

	graph, err := s.loadWorkspaceGraph(ctx, projectID)
	if err != nil {
		return WorkspaceProjectionSummary{}, err
	}

	desiredGaps := deriveDesiredContinuityGapSeeds(graph, observedAt)
	if err := s.repository.SyncContinuityGaps(ctx, missioncontrolrepo.SyncContinuityGapsParams{
		ProjectID:   projectID,
		ResolvedAt:  observedAt,
		DesiredOpen: desiredGaps,
	}); err != nil {
		return WorkspaceProjectionSummary{}, err
	}

	watermarkParams := buildWorkspaceWatermarkParams(graph, desiredGaps, observedAt)
	for _, item := range watermarkParams {
		if _, err := s.repository.CreateWorkspaceWatermark(ctx, item); err != nil {
			return WorkspaceProjectionSummary{}, err
		}
	}

	refreshedGraph, err := s.loadWorkspaceGraph(ctx, projectID)
	if err != nil {
		return WorkspaceProjectionSummary{}, err
	}
	snapshot := buildWorkspaceSnapshot(refreshedGraph, WorkspaceQuery{
		ProjectID:   projectID,
		StatePreset: enumtypes.MissionControlWorkspaceStatePresetAllActive,
	})
	summary := buildWorkspaceProjectionSummary(projectID, snapshot, desiredGaps, observedAt)
	s.insertFlowEvent(ctx, params.CorrelationID, eventTypeMissionControlWorkspaceRefreshed, workspaceProjectionEventPayload(summary))
	return summary, nil
}

// GetWorkspace returns one graph-first snapshot derived from persisted Mission Control truth.
func (s *Service) GetWorkspace(ctx context.Context, params WorkspaceQuery) (WorkspaceSnapshot, error) {
	if err := s.ensureReadAllowed(); err != nil {
		return WorkspaceSnapshot{}, err
	}
	graph, err := s.loadWorkspaceGraph(ctx, params.ProjectID)
	if err != nil {
		return WorkspaceSnapshot{}, err
	}
	return buildWorkspaceSnapshot(graph, params), nil
}

// PreviewLaunch returns one read-only deterministic preview for stage.next_step.execute.
func (s *Service) PreviewLaunch(ctx context.Context, params LaunchPreviewParams) (LaunchPreview, error) {
	if err := s.ensureReadAllowed(); err != nil {
		return LaunchPreview{}, err
	}
	projectID := strings.TrimSpace(params.ProjectID)
	if projectID == "" {
		return LaunchPreview{}, errs.Validation{Field: "project_id", Msg: "is required"}
	}
	graph, err := s.loadWorkspaceGraph(ctx, projectID)
	if err != nil {
		return LaunchPreview{}, err
	}
	entity, ok := entityByPublicRef(graph, params.NodeKind, params.NodePublicID)
	if !ok {
		return LaunchPreview{}, errs.NotFound{Msg: "mission control workspace node not found"}
	}
	preview, err := s.previewLaunchAgainstEntity(ctx, graph, entity, params)
	if err != nil {
		return LaunchPreview{}, err
	}
	eventType := eventTypeMissionControlPreviewGenerated
	if strings.TrimSpace(preview.BlockingReason) != "" {
		eventType = eventTypeMissionControlPreviewBlocked
	}
	s.insertFlowEvent(ctx, preview.PreviewID, eventType, workspacePreviewEventPayload{
		ProjectID: projectID,
		NodeRef: valuetypes.MissionControlEntityRef{
			EntityKind:     entity.EntityKind,
			EntityPublicID: entity.EntityExternalKey,
		},
		ThreadKind:      strings.TrimSpace(params.ThreadKind),
		ThreadNumber:    params.ThreadNumber,
		TargetLabel:     strings.TrimSpace(params.TargetLabel),
		PreviewID:       preview.PreviewID,
		BlockingReason:  preview.BlockingReason,
		ResolvedGapIDs:  preview.ContinuityEffect.ResolvedGapIDs,
		RemainingGapIDs: preview.ContinuityEffect.RemainingGapIDs,
	})
	return preview, nil
}

func (s *Service) previewLaunchAgainstEntity(
	_ context.Context,
	graph workspaceGraph,
	entity Entity,
	params LaunchPreviewParams,
) (LaunchPreview, error) {
	threadKind := strings.ToLower(strings.TrimSpace(params.ThreadKind))
	targetLabel := strings.TrimSpace(params.TargetLabel)
	if threadKind == "" {
		return LaunchPreview{}, errs.Validation{Field: "thread_kind", Msg: "is required"}
	}
	if threadKind != "issue" && threadKind != "pull_request" {
		return LaunchPreview{}, errs.Validation{Field: "thread_kind", Msg: "must be issue or pull_request"}
	}
	if params.ThreadNumber <= 0 {
		return LaunchPreview{}, errs.Validation{Field: "thread_number", Msg: "must be positive"}
	}
	if targetLabel == "" {
		return LaunchPreview{}, errs.Validation{Field: "target_label", Msg: "is required"}
	}
	if params.ExpectedProjectionVersion > 0 && entity.ProjectionVersion != params.ExpectedProjectionVersion {
		return LaunchPreview{}, errs.FailedPrecondition{Msg: "mission control preview projection is stale"}
	}
	if !nextstepLabelKnown(threadKind, targetLabel) {
		return LaunchPreview{}, errs.Validation{Field: "target_label", Msg: "must be a known next-step label for this thread kind"}
	}

	relevantGapIDs := previewRelevantGapIDs(graph, entity.ID)
	resolvedGapIDs := make([]int64, 0, len(relevantGapIDs))
	remainingGapIDs := make([]int64, 0, len(relevantGapIDs))
	for _, gapID := range relevantGapIDs {
		gap := openGapByID(graph, gapID)
		if gap.ID == 0 {
			continue
		}
		if gap.GapKind == enumtypes.MissionControlGapKindMissingFollowUpIssue &&
			previewResolvesFollowUpGap(graph, gap, params.ThreadKind, params.ThreadNumber, targetLabel, params.RemovedLabels) {
			resolvedGapIDs = append(resolvedGapIDs, gap.ID)
			continue
		}
		remainingGapIDs = append(remainingGapIDs, gap.ID)
	}
	slices.Sort(resolvedGapIDs)
	slices.Sort(remainingGapIDs)

	currentLabels := previewCurrentLabels(graph, entity, threadKind, params.ThreadNumber)
	finalLabels := previewFinalLabels(currentLabels, params.RemovedLabels, targetLabel)
	resultingNodeRefs := make([]valuetypes.MissionControlEntityRef, 0, 1)
	if guessed := guessedThreadEntityRef(entity, threadKind, params.ThreadNumber); guessed != nil {
		resultingNodeRefs = append(resultingNodeRefs, *guessed)
	}

	preview := LaunchPreview{
		ApprovalRequirement: StageNextStepApprovalRequirement(entity),
		LabelDiff: valuetypes.MissionControlLaunchPreviewLabelDiff{
			RemovedLabels: normalizeStringSlice(params.RemovedLabels),
			AddedLabels:   []string{targetLabel},
			FinalLabels:   finalLabels,
		},
		ContinuityEffect: valuetypes.MissionControlLaunchPreviewContinuityEffect{
			ResolvedGapIDs:    resolvedGapIDs,
			RemainingGapIDs:   remainingGapIDs,
			ResultingNodeRefs: resultingNodeRefs,
		},
	}
	if len(remainingGapIDs) > 0 {
		firstGap := openGapByID(graph, remainingGapIDs[0])
		preview.BlockingReason = string(firstGap.GapKind)
	}
	preview.PreviewID = buildPreviewID(
		graph.projectID,
		valuetypes.MissionControlEntityRef{EntityKind: entity.EntityKind, EntityPublicID: entity.EntityExternalKey},
		targetLabel,
		params.ExpectedProjectionVersion,
		resolvedGapIDs,
		remainingGapIDs,
	)
	return preview, nil
}

func buildWorkspaceProjectionSummary(
	projectID string,
	snapshot WorkspaceSnapshot,
	desiredGaps []missioncontrolrepo.ContinuityGapSeed,
	observedAt time.Time,
) WorkspaceProjectionSummary {
	summary := WorkspaceProjectionSummary{
		ProjectID:         projectID,
		EntityCount:       snapshot.Summary.NodeCount,
		RootCount:         snapshot.Summary.RootCount,
		NodeCount:         snapshot.Summary.NodeCount,
		OpenGapCount:      len(desiredGaps),
		BlockingGapCount:  snapshot.Summary.BlockingGapCount,
		WarningGapCount:   snapshot.Summary.WarningGapCount,
		WatermarkCount:    len(snapshot.WorkspaceWatermarks),
		ReadyForReconcile: snapshot.Summary.BlockingGapCount == 0,
		ObservedAt:        observedAt.UTC(),
	}
	if !summary.ReadyForReconcile {
		summary.GatingReason = "open_blocking_continuity_gaps"
	}
	for _, gap := range desiredGaps {
		switch gap.GapKind {
		case enumtypes.MissionControlGapKindMissingPullRequest:
			summary.MissingPullRequestGapCount++
		case enumtypes.MissionControlGapKindMissingFollowUpIssue:
			summary.MissingFollowUpIssueGapCount++
		}
	}
	return summary
}

func entityByPublicRef(graph workspaceGraph, kind enumtypes.MissionControlEntityKind, publicID string) (Entity, bool) {
	publicID = strings.TrimSpace(publicID)
	for _, entity := range graph.entities {
		if entity.EntityKind == kind && entity.EntityExternalKey == publicID {
			return entity, true
		}
	}
	return Entity{}, false
}

func previewRelevantGapIDs(graph workspaceGraph, entityID int64) []int64 {
	relevantEntityIDs := map[int64]struct{}{entityID: {}}
	entity, ok := graph.entityByID[entityID]
	if !ok {
		return nil
	}
	switch entity.EntityKind {
	case enumtypes.MissionControlEntityKindWorkItem, enumtypes.MissionControlEntityKindDiscussion:
		for _, relation := range graph.outgoing[entityID] {
			if relation.RelationKind != enumtypes.MissionControlRelationKindSpawnedRun &&
				relation.RelationKind != enumtypes.MissionControlRelationKindFormalizedFrom {
				continue
			}
			relevantEntityIDs[relation.TargetEntityID] = struct{}{}
			for _, childRelation := range graph.outgoing[relation.TargetEntityID] {
				if childRelation.RelationKind == enumtypes.MissionControlRelationKindProducedPullRequest {
					relevantEntityIDs[childRelation.TargetEntityID] = struct{}{}
				}
			}
		}
	case enumtypes.MissionControlEntityKindRun:
		for _, relation := range graph.outgoing[entityID] {
			if relation.RelationKind == enumtypes.MissionControlRelationKindProducedPullRequest {
				relevantEntityIDs[relation.TargetEntityID] = struct{}{}
			}
		}
	}

	gapIDs := make([]int64, 0)
	for subjectEntityID := range relevantEntityIDs {
		for _, gap := range graph.gapsBySubjectID[subjectEntityID] {
			if gap.Status != enumtypes.MissionControlGapStatusOpen || gap.Severity != enumtypes.MissionControlGapSeverityBlocking {
				continue
			}
			gapIDs = append(gapIDs, gap.ID)
		}
	}
	slices.Sort(gapIDs)
	return gapIDs
}

func openGapByID(graph workspaceGraph, gapID int64) missioncontrolrepo.ContinuityGap {
	for _, gap := range graph.openGaps {
		if gap.ID == gapID {
			return gap
		}
	}
	return missioncontrolrepo.ContinuityGap{}
}

func previewResolvesFollowUpGap(
	graph workspaceGraph,
	gap missioncontrolrepo.ContinuityGap,
	threadKind string,
	threadNumber int,
	targetLabel string,
	removedLabels []string,
) bool {
	if !strings.EqualFold(strings.TrimSpace(threadKind), "issue") || threadNumber <= 0 {
		return false
	}
	subjectEntity, ok := graph.entityByID[gap.SubjectEntityID]
	if !ok || subjectEntity.EntityKind != enumtypes.MissionControlEntityKindPullRequest {
		return false
	}
	targetRef := guessedThreadEntityRef(subjectEntity, threadKind, threadNumber)
	if targetRef == nil || targetRef.EntityKind != enumtypes.MissionControlEntityKindWorkItem {
		return false
	}
	targetEntity, ok := entityByPublicRef(graph, targetRef.EntityKind, targetRef.EntityPublicID)
	if !ok {
		return false
	}
	if targetEntity.EntityExternalKey == pullRequestSourceIssueRef(graph, subjectEntity) {
		return false
	}
	if !pullRequestReferencesIssue(graph, subjectEntity, targetEntity.EntityExternalKey) {
		return false
	}
	return workItemMatchesStageAfterLabelChange(targetEntity, removedLabels, targetLabel, gap.ExpectedStageLabel)
}

func currentIssueNumber(entity Entity) int64 {
	switch entity.EntityKind {
	case enumtypes.MissionControlEntityKindWorkItem:
		if payload, ok := decodeWorkItemPayload(entity); ok {
			return payload.IssueNumber
		}
	case enumtypes.MissionControlEntityKindRun:
		if payload, ok := decodeRunPayload(entity); ok && payload.IssueRef != "" {
			return issueNumberFromPublicID(payload.IssueRef)
		}
	case enumtypes.MissionControlEntityKindPullRequest:
		if payload, ok := decodePullRequestPayload(entity); ok && len(payload.LinkedIssueRefs) > 0 {
			return issueNumberFromPublicID(payload.LinkedIssueRefs[0])
		}
	}
	return 0
}

func issueNumberFromPublicID(publicID string) int64 {
	parts := strings.Split(strings.TrimSpace(publicID), "#")
	if len(parts) != 2 {
		return 0
	}
	number, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0
	}
	return number
}

func previewCurrentLabels(graph workspaceGraph, sourceEntity Entity, threadKind string, threadNumber int) []string {
	ref := guessedThreadEntityRef(sourceEntity, threadKind, threadNumber)
	if ref == nil {
		return nil
	}
	target, ok := entityByPublicRef(graph, ref.EntityKind, ref.EntityPublicID)
	if !ok {
		return nil
	}
	switch target.EntityKind {
	case enumtypes.MissionControlEntityKindWorkItem:
		if payload, ok := decodeWorkItemPayload(target); ok {
			return normalizeStringSlice(payload.Labels)
		}
	}
	return nil
}

func previewFinalLabels(current []string, removed []string, targetLabel string) []string {
	removed = normalizeStringSlice(removed)
	final := make([]string, 0, len(current)+1)
	for _, label := range normalizeStringSlice(current) {
		if slices.Contains(removed, label) {
			continue
		}
		final = append(final, label)
	}
	if targetLabel != "" && !slices.Contains(final, targetLabel) {
		final = append(final, targetLabel)
	}
	slices.Sort(final)
	return final
}

func nextstepLabelKnown(threadKind string, label string) bool {
	return nextstepdomain.DefaultLabels().IsKnownLabelForThreadKind(threadKind, label)
}
