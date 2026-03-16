package missioncontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	nextstepdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/nextstep"
	missioncontrolrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/missioncontrol"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

const workspaceProjectionLoadLimit = 1000

type workspaceGraph struct {
	projectID        string
	entities         []Entity
	entityByID       map[int64]Entity
	outgoing         map[int64][]Relation
	incoming         map[int64][]Relation
	openGaps         []missioncontrolrepo.ContinuityGap
	gapsBySubjectID  map[int64][]missioncontrolrepo.ContinuityGap
	latestWatermarks []missioncontrolrepo.WorkspaceWatermark
}

func (s *Service) loadWorkspaceGraph(ctx context.Context, projectID string) (workspaceGraph, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return workspaceGraph{}, errs.Validation{Field: "project_id", Msg: "is required"}
	}

	entities, err := s.repository.ListEntities(ctx, missioncontrolrepo.EntityListFilter{
		ProjectID: projectID,
		Limit:     workspaceProjectionLoadLimit,
	})
	if err != nil {
		return workspaceGraph{}, err
	}

	graph := workspaceGraph{
		projectID:       projectID,
		entities:        entities,
		entityByID:      make(map[int64]Entity, len(entities)),
		outgoing:        make(map[int64][]Relation, len(entities)),
		incoming:        make(map[int64][]Relation, len(entities)),
		gapsBySubjectID: make(map[int64][]missioncontrolrepo.ContinuityGap),
	}
	relationIDs := make(map[int64]struct{})
	for _, entity := range entities {
		graph.entityByID[entity.ID] = entity
		relations, relationErr := s.repository.ListRelationsForEntity(ctx, projectID, entity.ID)
		if relationErr != nil {
			return workspaceGraph{}, relationErr
		}
		for _, relation := range relations {
			if _, seen := relationIDs[relation.ID]; seen {
				continue
			}
			relationIDs[relation.ID] = struct{}{}
			graph.outgoing[relation.SourceEntityID] = append(graph.outgoing[relation.SourceEntityID], relation)
			graph.incoming[relation.TargetEntityID] = append(graph.incoming[relation.TargetEntityID], relation)
		}
	}
	for entityID := range graph.outgoing {
		sort.Slice(graph.outgoing[entityID], func(i, j int) bool { return graph.outgoing[entityID][i].ID < graph.outgoing[entityID][j].ID })
	}
	for entityID := range graph.incoming {
		sort.Slice(graph.incoming[entityID], func(i, j int) bool { return graph.incoming[entityID][i].ID < graph.incoming[entityID][j].ID })
	}

	openGaps, err := s.repository.ListContinuityGaps(ctx, missioncontrolrepo.ContinuityGapListFilter{
		ProjectID: projectID,
		Statuses:  []enumtypes.MissionControlGapStatus{enumtypes.MissionControlGapStatusOpen},
	})
	if err != nil {
		return workspaceGraph{}, err
	}
	graph.openGaps = openGaps
	for _, gap := range openGaps {
		graph.gapsBySubjectID[gap.SubjectEntityID] = append(graph.gapsBySubjectID[gap.SubjectEntityID], gap)
	}
	for entityID := range graph.gapsBySubjectID {
		sort.Slice(graph.gapsBySubjectID[entityID], func(i, j int) bool {
			left := graph.gapsBySubjectID[entityID][i]
			right := graph.gapsBySubjectID[entityID][j]
			if !left.DetectedAt.Equal(right.DetectedAt) {
				return left.DetectedAt.After(right.DetectedAt)
			}
			return left.ID > right.ID
		})
	}

	latestWatermarks, err := s.repository.ListLatestWorkspaceWatermarks(ctx, projectID)
	if err != nil {
		return workspaceGraph{}, err
	}
	graph.latestWatermarks = latestWatermarks
	return graph, nil
}

func deriveDesiredContinuityGapSeeds(graph workspaceGraph, observedAt time.Time) []missioncontrolrepo.ContinuityGapSeed {
	if observedAt.IsZero() {
		observedAt = time.Now().UTC()
	}
	desired := make(map[string]missioncontrolrepo.ContinuityGapSeed)
	for _, entity := range graph.entities {
		if entity.EntityKind == enumtypes.MissionControlEntityKindAgent || entity.CoverageClass == enumtypes.MissionControlCoverageClassOutOfScope {
			continue
		}
		if entity.ProviderKind == enumtypes.MissionControlProviderKindGitHub &&
			entity.StaleAfter != nil &&
			!entity.StaleAfter.IsZero() &&
			observedAt.After(entity.StaleAfter.UTC()) {
			desired[continuityGapSeedKey(entity.ID, enumtypes.MissionControlGapKindProviderStale)] = missioncontrolrepo.ContinuityGapSeed{
				SubjectEntityID: entity.ID,
				GapKind:         enumtypes.MissionControlGapKindProviderStale,
				Severity:        enumtypes.MissionControlGapSeverityWarning,
				ResolutionHint:  "Обновить provider mirror для этого узла, чтобы восстановить freshness и launch preview.",
				DetectedAt:      observedAt,
			}
		}
		if entity.EntityKind != enumtypes.MissionControlEntityKindRun {
			continue
		}
		if hasRelationKind(graph.outgoing[entity.ID], enumtypes.MissionControlRelationKindContinuesWith) {
			continue
		}
		pullRequests := producedPullRequestEntities(graph, entity.ID)
		if len(pullRequests) == 0 {
			desired[continuityGapSeedKey(entity.ID, enumtypes.MissionControlGapKindMissingPullRequest)] = missioncontrolrepo.ContinuityGapSeed{
				SubjectEntityID:    entity.ID,
				GapKind:            enumtypes.MissionControlGapKindMissingPullRequest,
				Severity:           enumtypes.MissionControlGapSeverityBlocking,
				ExpectedEntityKind: enumtypes.MissionControlEntityKindPullRequest,
				ExpectedStageLabel: runStageLabel(entity),
				ResolutionHint:     "Подготовить pull request как обязательный stage artifact перед следующим переходом.",
				PayloadJSON:        buildMissingPullRequestPayload(entity),
				DetectedAt:         lastRelevantActivity(entity, observedAt),
			}
			continue
		}
		if !runStageRequiresFollowUp(entity) || runHasLinkedFollowUpIssue(graph, entity, pullRequests) {
			continue
		}
		pullRequest := newestEntity(pullRequests)
		desired[continuityGapSeedKey(pullRequest.ID, enumtypes.MissionControlGapKindMissingFollowUpIssue)] = missioncontrolrepo.ContinuityGapSeed{
			SubjectEntityID:    pullRequest.ID,
			GapKind:            enumtypes.MissionControlGapKindMissingFollowUpIssue,
			Severity:           enumtypes.MissionControlGapSeverityBlocking,
			ExpectedEntityKind: enumtypes.MissionControlEntityKindWorkItem,
			ExpectedStageLabel: nextStageLabelForRun(entity),
			ResolutionHint:     "Нужна linked follow-up issue для следующего stage перед launch preview.",
			PayloadJSON:        buildMissingFollowUpPayload(entity, pullRequest),
			DetectedAt:         lastRelevantActivity(pullRequest, observedAt),
		}
	}

	keys := make([]string, 0, len(desired))
	for key := range desired {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]missioncontrolrepo.ContinuityGapSeed, 0, len(keys))
	for _, key := range keys {
		out = append(out, desired[key])
	}
	return out
}

func buildWorkspaceWatermarkParams(
	graph workspaceGraph,
	desiredGaps []missioncontrolrepo.ContinuityGapSeed,
	observedAt time.Time,
) []missioncontrolrepo.CreateWorkspaceWatermarkParams {
	if observedAt.IsZero() {
		observedAt = time.Now().UTC()
	}

	var (
		primaryOpenCount    int
		recentClosedCount   int
		runEntityCount      int
		legacyAgentCount    int
		providerEntityCount int
		staleProviderCount  int
		blockingGapCount    int
		recentClosedStart   *time.Time
		recentClosedEnd     *time.Time
		providerWindowStart *time.Time
		providerWindowEnd   *time.Time
	)

	for _, entity := range graph.entities {
		switch entity.CoverageClass {
		case enumtypes.MissionControlCoverageClassOpenPrimary:
			primaryOpenCount++
		case enumtypes.MissionControlCoverageClassRecentClosedContext:
			recentClosedCount++
			recentClosedStart = earlierTimePtr(recentClosedStart, entity.ProjectedAt)
			recentClosedEnd = laterTimePtr(recentClosedEnd, entity.ProjectedAt)
		}
		switch entity.EntityKind {
		case enumtypes.MissionControlEntityKindRun:
			runEntityCount++
		case enumtypes.MissionControlEntityKindAgent:
			legacyAgentCount++
		}
		if entity.ProviderKind != enumtypes.MissionControlProviderKindGitHub {
			continue
		}
		providerEntityCount++
		if entity.StaleAfter != nil && observedAt.After(entity.StaleAfter.UTC()) {
			staleProviderCount++
		}
		providerWindowStart = earlierTimePtr(providerWindowStart, projectionTimestamp(entity))
		providerWindowEnd = laterTimePtr(providerWindowEnd, projectionTimestamp(entity))
	}

	for _, gap := range desiredGaps {
		if gap.Severity == enumtypes.MissionControlGapSeverityBlocking {
			blockingGapCount++
		}
	}

	providerFreshnessStatus := enumtypes.MissionControlWorkspaceWatermarkStatusOutOfScope
	providerFreshnessSummary := "Workspace пока не содержит GitHub-backed узлов для оценки provider freshness."
	if providerEntityCount > 0 {
		if staleProviderCount > 0 {
			providerFreshnessStatus = enumtypes.MissionControlWorkspaceWatermarkStatusStale
			providerFreshnessSummary = fmt.Sprintf(
				"Provider freshness деградировал для %d из %d GitHub-backed узлов Mission Control.",
				staleProviderCount,
				providerEntityCount,
			)
		} else {
			providerFreshnessStatus = enumtypes.MissionControlWorkspaceWatermarkStatusFresh
			providerFreshnessSummary = fmt.Sprintf(
				"Provider freshness остаётся в staleness window для %d GitHub-backed узлов Mission Control.",
				providerEntityCount,
			)
		}
	}

	graphProjectionStatus := enumtypes.MissionControlWorkspaceWatermarkStatusFresh
	graphProjectionSummary := fmt.Sprintf(
		"Graph truth содержит %d узлов, %d run-узлов и %d открытых continuity gaps.",
		len(graph.entities),
		runEntityCount,
		len(desiredGaps),
	)
	if len(desiredGaps) > 0 {
		graphProjectionStatus = enumtypes.MissionControlWorkspaceWatermarkStatusDegraded
	}

	launchPolicyStatus := enumtypes.MissionControlWorkspaceWatermarkStatusFresh
	launchPolicySummary := "Launch preview остаётся read-only и использует current command ledger without side effects."
	if blockingGapCount > 0 {
		launchPolicyStatus = enumtypes.MissionControlWorkspaceWatermarkStatusDegraded
		launchPolicySummary = fmt.Sprintf(
			"Launch preview остаётся read-only, но %d blocking continuity gaps не позволяют безопасно открыть reconcile stream.",
			blockingGapCount,
		)
	}

	return []missioncontrolrepo.CreateWorkspaceWatermarkParams{
		{
			ProjectID:       graph.projectID,
			WatermarkKind:   enumtypes.MissionControlWorkspaceWatermarkKindProviderFreshness,
			Status:          providerFreshnessStatus,
			Summary:         providerFreshnessSummary,
			WindowStartedAt: providerWindowStart,
			WindowEndedAt:   providerWindowEnd,
			ObservedAt:      observedAt,
			PayloadJSON: mustMarshal(valuetypes.MissionControlWorkspaceWatermarkPayload{
				Source:            "control_plane",
				Scope:             "provider_freshness",
				EntityCount:       providerEntityCount,
				RunEntityCount:    runEntityCount,
				OpenGapCount:      len(desiredGaps),
				LegacyAgentCount:  legacyAgentCount,
				RecentClosedCount: recentClosedCount,
				PrimaryOpenCount:  primaryOpenCount,
			}),
		},
		{
			ProjectID:       graph.projectID,
			WatermarkKind:   enumtypes.MissionControlWorkspaceWatermarkKindProviderCoverage,
			Status:          enumtypes.MissionControlWorkspaceWatermarkStatusOutOfScope,
			Summary:         "Wave 2 использует bounded coverage из persisted graph truth; provider mirror coverage остаётся отдельным rollout stream.",
			WindowStartedAt: recentClosedStart,
			WindowEndedAt:   recentClosedEnd,
			ObservedAt:      observedAt,
			PayloadJSON: mustMarshal(valuetypes.MissionControlWorkspaceWatermarkPayload{
				Source:            "control_plane",
				Scope:             "bounded_recent_closed_context",
				EntityCount:       len(graph.entities),
				RunEntityCount:    runEntityCount,
				OpenGapCount:      len(desiredGaps),
				LegacyAgentCount:  legacyAgentCount,
				RecentClosedCount: recentClosedCount,
				PrimaryOpenCount:  primaryOpenCount,
			}),
		},
		{
			ProjectID:     graph.projectID,
			WatermarkKind: enumtypes.MissionControlWorkspaceWatermarkKindGraphProjection,
			Status:        graphProjectionStatus,
			Summary:       graphProjectionSummary,
			ObservedAt:    observedAt,
			PayloadJSON: mustMarshal(valuetypes.MissionControlWorkspaceWatermarkPayload{
				Source:            "control_plane",
				Scope:             graph.projectID,
				EntityCount:       len(graph.entities),
				RunEntityCount:    runEntityCount,
				OpenGapCount:      len(desiredGaps),
				LegacyAgentCount:  legacyAgentCount,
				RecentClosedCount: recentClosedCount,
				PrimaryOpenCount:  primaryOpenCount,
			}),
		},
		{
			ProjectID:     graph.projectID,
			WatermarkKind: enumtypes.MissionControlWorkspaceWatermarkKindLaunchPolicy,
			Status:        launchPolicyStatus,
			Summary:       launchPolicySummary,
			ObservedAt:    observedAt,
			PayloadJSON: mustMarshal(valuetypes.MissionControlWorkspaceWatermarkPayload{
				Source:            "control_plane",
				Scope:             "launch_policy",
				EntityCount:       len(graph.entities),
				RunEntityCount:    runEntityCount,
				OpenGapCount:      len(desiredGaps),
				LegacyAgentCount:  legacyAgentCount,
				RecentClosedCount: recentClosedCount,
				PrimaryOpenCount:  primaryOpenCount,
			}),
		},
	}
}

func buildWorkspaceSnapshot(
	graph workspaceGraph,
	params WorkspaceQuery,
) WorkspaceSnapshot {
	effectivePreset := normalizeWorkspaceStatePreset(params.StatePreset)
	search := strings.ToLower(strings.TrimSpace(params.Search))
	effectiveContinuity := buildEffectiveContinuityMap(graph)

	primaryIDs := make(map[int64]struct{})
	for _, entity := range graph.entities {
		if entity.EntityKind == enumtypes.MissionControlEntityKindAgent {
			continue
		}
		if entity.CoverageClass != enumtypes.MissionControlCoverageClassOpenPrimary {
			continue
		}
		if !matchesWorkspaceStatePreset(entity, effectivePreset) {
			continue
		}
		if !matchesWorkspaceSearch(entity, search) {
			continue
		}
		primaryIDs[entity.ID] = struct{}{}
	}

	includedIDs := make(map[int64]struct{}, len(primaryIDs))
	for entityID := range primaryIDs {
		includedIDs[entityID] = struct{}{}
	}
	changed := true
	for changed {
		changed = false
		for _, relations := range graph.outgoing {
			for _, relation := range relations {
				sourceIncluded := idInSet(includedIDs, relation.SourceEntityID)
				targetIncluded := idInSet(includedIDs, relation.TargetEntityID)
				if sourceIncluded == targetIncluded {
					continue
				}
				var candidateID int64
				if sourceIncluded {
					candidateID = relation.TargetEntityID
				} else {
					candidateID = relation.SourceEntityID
				}
				candidate, ok := graph.entityByID[candidateID]
				if !ok || candidate.EntityKind == enumtypes.MissionControlEntityKindAgent || !workspaceSecondaryCandidate(candidate, effectivePreset, search) {
					continue
				}
				includedIDs[candidateID] = struct{}{}
				changed = true
			}
		}
	}

	rootIDs := selectWorkspaceRootIDs(graph, primaryIDs, includedIDs)
	rootByNode := make(map[int64]int64, len(includedIDs))
	for entityID := range includedIDs {
		rootByNode[entityID] = resolveWorkspaceRootID(graph, entityID, rootIDs)
	}

	nodes := make([]valuetypes.MissionControlWorkspaceNode, 0, len(includedIDs))
	rootGroups := make(map[int64]*valuetypes.MissionControlWorkspaceRootGroup, len(rootIDs))
	for _, rootID := range rootIDs {
		rootEntity := graph.entityByID[rootID]
		rootGroups[rootID] = &valuetypes.MissionControlWorkspaceRootGroup{
			RootNodeRef: valuetypes.MissionControlEntityRef{
				EntityKind:     rootEntity.EntityKind,
				EntityPublicID: rootEntity.EntityExternalKey,
			},
			RootTitle: strings.TrimSpace(rootEntity.Title),
		}
	}

	for entityID := range includedIDs {
		entity := graph.entityByID[entityID]
		rootID := rootByNode[entityID]
		rootEntity := graph.entityByID[rootID]
		visibility := enumtypes.MissionControlWorkspaceVisibilityTierSecondaryDimmed
		if idInSet(primaryIDs, entityID) {
			visibility = enumtypes.MissionControlWorkspaceVisibilityTierPrimary
		}
		node := valuetypes.MissionControlWorkspaceNode{
			NodeRef: valuetypes.MissionControlEntityRef{
				EntityKind:     entity.EntityKind,
				EntityPublicID: entity.EntityExternalKey,
			},
			Title:             strings.TrimSpace(entity.Title),
			VisibilityTier:    visibility,
			ActiveState:       entity.ActiveState,
			ContinuityStatus:  effectiveContinuity[entityID],
			CoverageClass:     entity.CoverageClass,
			RootNodePublicID:  rootEntity.EntityExternalKey,
			ColumnIndex:       workspaceColumnIndex(entity.EntityKind),
			LastActivityAt:    entityLastActivity(entity),
			HasBlockingGap:    continuityStatusBlocking(effectiveContinuity[entityID]),
			ProjectionVersion: entity.ProjectionVersion,
		}
		nodes = append(nodes, node)

		group := rootGroups[rootID]
		group.NodeRefs = append(group.NodeRefs, node.NodeRef)
		if node.HasBlockingGap {
			group.HasBlockingGap = true
		}
		group.LatestActivityAt = laterTimePtr(group.LatestActivityAt, derefTime(entityLastActivity(entity), entity.ProjectedAt))
	}

	sort.Slice(nodes, func(i, j int) bool {
		left := nodes[i]
		right := nodes[j]
		if left.RootNodePublicID != right.RootNodePublicID {
			return left.RootNodePublicID < right.RootNodePublicID
		}
		if left.ColumnIndex != right.ColumnIndex {
			return left.ColumnIndex < right.ColumnIndex
		}
		return left.NodeRef.EntityPublicID < right.NodeRef.EntityPublicID
	})

	edges := make([]valuetypes.MissionControlWorkspaceEdge, 0)
	for sourceID, relations := range graph.outgoing {
		if !idInSet(includedIDs, sourceID) {
			continue
		}
		for _, relation := range relations {
			if !idInSet(includedIDs, relation.TargetEntityID) {
				continue
			}
			sourceEntity := graph.entityByID[relation.SourceEntityID]
			targetEntity := graph.entityByID[relation.TargetEntityID]
			visibility := enumtypes.MissionControlWorkspaceVisibilityTierSecondaryDimmed
			if idInSet(primaryIDs, relation.SourceEntityID) && idInSet(primaryIDs, relation.TargetEntityID) {
				visibility = enumtypes.MissionControlWorkspaceVisibilityTierPrimary
			}
			edges = append(edges, valuetypes.MissionControlWorkspaceEdge{
				RelationKind: relation.RelationKind,
				SourceNodeRef: valuetypes.MissionControlEntityRef{
					EntityKind:     sourceEntity.EntityKind,
					EntityPublicID: sourceEntity.EntityExternalKey,
				},
				TargetNodeRef: valuetypes.MissionControlEntityRef{
					EntityKind:     targetEntity.EntityKind,
					EntityPublicID: targetEntity.EntityExternalKey,
				},
				VisibilityTier: visibility,
				SourceOfTruth:  relation.SourceKind,
				IsPrimaryPath:  relationIsPrimaryPath(relation.RelationKind),
			})
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		left := edges[i]
		right := edges[j]
		if left.SourceNodeRef.EntityPublicID != right.SourceNodeRef.EntityPublicID {
			return left.SourceNodeRef.EntityPublicID < right.SourceNodeRef.EntityPublicID
		}
		if left.TargetNodeRef.EntityPublicID != right.TargetNodeRef.EntityPublicID {
			return left.TargetNodeRef.EntityPublicID < right.TargetNodeRef.EntityPublicID
		}
		return left.RelationKind < right.RelationKind
	})

	rootList := make([]valuetypes.MissionControlWorkspaceRootGroup, 0, len(rootGroups))
	for _, rootID := range rootIDs {
		group := rootGroups[rootID]
		sort.Slice(group.NodeRefs, func(i, j int) bool { return group.NodeRefs[i].EntityPublicID < group.NodeRefs[j].EntityPublicID })
		rootList = append(rootList, *group)
	}
	sort.Slice(rootList, func(i, j int) bool {
		left := derefTime(rootList[i].LatestActivityAt, time.Time{})
		right := derefTime(rootList[j].LatestActivityAt, time.Time{})
		if !left.Equal(right) {
			return left.After(right)
		}
		return rootList[i].RootNodeRef.EntityPublicID < rootList[j].RootNodeRef.EntityPublicID
	})

	if params.RootLimit > 0 && len(rootList) > params.RootLimit {
		allowedRoots := make(map[string]struct{}, params.RootLimit)
		trimmed := rootList[:params.RootLimit]
		for _, root := range trimmed {
			allowedRoots[root.RootNodeRef.EntityPublicID] = struct{}{}
		}
		rootList = trimmed

		filteredNodes := nodes[:0]
		for _, node := range nodes {
			if _, ok := allowedRoots[node.RootNodePublicID]; ok {
				filteredNodes = append(filteredNodes, node)
			}
		}
		nodes = filteredNodes

		filteredEdges := edges[:0]
		for _, edge := range edges {
			sourceRoot := ""
			targetRoot := ""
			for _, node := range nodes {
				if node.NodeRef == edge.SourceNodeRef {
					sourceRoot = node.RootNodePublicID
				}
				if node.NodeRef == edge.TargetNodeRef {
					targetRoot = node.RootNodePublicID
				}
			}
			if sourceRoot != "" && sourceRoot == targetRoot {
				filteredEdges = append(filteredEdges, edge)
			}
		}
		edges = filteredEdges
	}

	summary := valuetypes.MissionControlWorkspaceSummary{}
	summary.RootCount = len(rootList)
	summary.NodeCount = len(nodes)
	for _, node := range nodes {
		if node.VisibilityTier == enumtypes.MissionControlWorkspaceVisibilityTierSecondaryDimmed {
			summary.SecondaryDimmedNodeCount++
		}
		if node.CoverageClass == enumtypes.MissionControlCoverageClassRecentClosedContext {
			summary.RecentClosedContextCount++
		}
	}
	for _, gap := range graph.openGaps {
		if !idInSet(includedIDs, gap.SubjectEntityID) {
			continue
		}
		switch gap.Severity {
		case enumtypes.MissionControlGapSeverityBlocking:
			summary.BlockingGapCount++
		case enumtypes.MissionControlGapSeverityWarning:
			summary.WarningGapCount++
		}
	}

	return WorkspaceSnapshot{
		Summary:             summary,
		WorkspaceWatermarks: append([]missioncontrolrepo.WorkspaceWatermark(nil), graph.latestWatermarks...),
		RootGroups:          rootList,
		Nodes:               nodes,
		Edges:               edges,
	}
}

func buildEffectiveContinuityMap(graph workspaceGraph) map[int64]enumtypes.MissionControlContinuityStatus {
	direct := make(map[int64]enumtypes.MissionControlContinuityStatus, len(graph.entities))
	for _, entity := range graph.entities {
		direct[entity.ID] = directContinuityStatus(entity, graph.gapsBySubjectID[entity.ID])
	}
	memo := make(map[int64]enumtypes.MissionControlContinuityStatus, len(graph.entities))
	for _, entity := range graph.entities {
		memo[entity.ID] = effectiveContinuityStatus(graph, entity.ID, direct, memo)
	}
	return memo
}

func directContinuityStatus(entity Entity, gaps []missioncontrolrepo.ContinuityGap) enumtypes.MissionControlContinuityStatus {
	if entity.CoverageClass == enumtypes.MissionControlCoverageClassOutOfScope {
		return enumtypes.MissionControlContinuityStatusOutOfScope
	}
	status := enumtypes.MissionControlContinuityStatusComplete
	for _, gap := range gaps {
		if gap.Status != enumtypes.MissionControlGapStatusOpen {
			continue
		}
		status = worseContinuityStatus(status, continuityStatusFromGapKind(gap.GapKind))
	}
	return status
}

func effectiveContinuityStatus(
	graph workspaceGraph,
	entityID int64,
	direct map[int64]enumtypes.MissionControlContinuityStatus,
	memo map[int64]enumtypes.MissionControlContinuityStatus,
) enumtypes.MissionControlContinuityStatus {
	if status, ok := memo[entityID]; ok && status != "" {
		return status
	}
	entity, ok := graph.entityByID[entityID]
	if !ok {
		return enumtypes.MissionControlContinuityStatusOutOfScope
	}
	status := direct[entityID]
	switch entity.EntityKind {
	case enumtypes.MissionControlEntityKindRun:
		for _, relation := range graph.outgoing[entityID] {
			if relation.RelationKind != enumtypes.MissionControlRelationKindProducedPullRequest {
				continue
			}
			status = worseContinuityStatus(status, effectiveContinuityStatus(graph, relation.TargetEntityID, direct, memo))
		}
	case enumtypes.MissionControlEntityKindWorkItem, enumtypes.MissionControlEntityKindDiscussion:
		for _, relation := range graph.outgoing[entityID] {
			if relation.RelationKind != enumtypes.MissionControlRelationKindSpawnedRun &&
				relation.RelationKind != enumtypes.MissionControlRelationKindFormalizedFrom {
				continue
			}
			status = worseContinuityStatus(status, effectiveContinuityStatus(graph, relation.TargetEntityID, direct, memo))
		}
	}
	memo[entityID] = status
	return status
}

func continuityStatusFromGapKind(kind enumtypes.MissionControlGapKind) enumtypes.MissionControlContinuityStatus {
	switch kind {
	case enumtypes.MissionControlGapKindMissingPullRequest:
		return enumtypes.MissionControlContinuityStatusMissingPullRequest
	case enumtypes.MissionControlGapKindMissingFollowUpIssue:
		return enumtypes.MissionControlContinuityStatusMissingFollowUpIssue
	case enumtypes.MissionControlGapKindProviderStale:
		return enumtypes.MissionControlContinuityStatusStaleProvider
	default:
		return enumtypes.MissionControlContinuityStatusComplete
	}
}

func worseContinuityStatus(
	current enumtypes.MissionControlContinuityStatus,
	candidate enumtypes.MissionControlContinuityStatus,
) enumtypes.MissionControlContinuityStatus {
	if continuityStatusSeverity(candidate) > continuityStatusSeverity(current) {
		return candidate
	}
	return current
}

func continuityStatusSeverity(status enumtypes.MissionControlContinuityStatus) int {
	switch status {
	case enumtypes.MissionControlContinuityStatusMissingPullRequest:
		return 4
	case enumtypes.MissionControlContinuityStatusMissingFollowUpIssue:
		return 3
	case enumtypes.MissionControlContinuityStatusStaleProvider:
		return 2
	case enumtypes.MissionControlContinuityStatusComplete:
		return 1
	default:
		return 0
	}
}

func continuityStatusBlocking(status enumtypes.MissionControlContinuityStatus) bool {
	return status == enumtypes.MissionControlContinuityStatusMissingPullRequest ||
		status == enumtypes.MissionControlContinuityStatusMissingFollowUpIssue
}

func producedPullRequestEntities(graph workspaceGraph, runEntityID int64) []Entity {
	out := make([]Entity, 0, 1)
	for _, relation := range graph.outgoing[runEntityID] {
		if relation.RelationKind != enumtypes.MissionControlRelationKindProducedPullRequest {
			continue
		}
		if entity, ok := graph.entityByID[relation.TargetEntityID]; ok {
			out = append(out, entity)
		}
	}
	return out
}

func newestEntity(entities []Entity) Entity {
	sort.Slice(entities, func(i, j int) bool {
		if !entities[i].ProjectedAt.Equal(entities[j].ProjectedAt) {
			return entities[i].ProjectedAt.After(entities[j].ProjectedAt)
		}
		return entities[i].ID > entities[j].ID
	})
	return entities[0]
}

func hasRelationKind(relations []Relation, kind enumtypes.MissionControlRelationKind) bool {
	for _, relation := range relations {
		if relation.RelationKind == kind {
			return true
		}
	}
	return false
}

func runHasLinkedFollowUpIssue(graph workspaceGraph, runEntity Entity, pullRequests []Entity) bool {
	sourceIssueRef := runIssueRef(runEntity)
	for _, pullRequest := range pullRequests {
		payload, ok := decodePullRequestPayload(pullRequest)
		if ok {
			for _, issueRef := range payload.LinkedIssueRefs {
				if issueRef != "" && issueRef != sourceIssueRef {
					return true
				}
			}
		}
		for _, relation := range graph.incoming[pullRequest.ID] {
			if relation.RelationKind != enumtypes.MissionControlRelationKindRelatedTo {
				continue
			}
			relatedEntity, ok := graph.entityByID[relation.SourceEntityID]
			if !ok || relatedEntity.EntityKind != enumtypes.MissionControlEntityKindWorkItem {
				continue
			}
			if relatedEntity.EntityExternalKey != sourceIssueRef {
				return true
			}
		}
	}
	return false
}

func runStageRequiresFollowUp(entity Entity) bool {
	return stageLabelRequiresFollowUp(runStageLabel(entity))
}

func stageLabelRequiresFollowUp(label string) bool {
	normalized := strings.ToLower(strings.TrimSpace(label))
	if normalized == "" {
		return false
	}
	if !strings.HasPrefix(normalized, "run:") {
		return false
	}
	stagePart := strings.TrimPrefix(normalized, "run:")
	stage := strings.Split(stagePart, ":")[0]
	switch stage {
	case "intake", "vision", "prd", "arch", "design", "plan", "dev":
		return true
	default:
		return false
	}
}

func runStageLabel(entity Entity) string {
	payload, ok := decodeRunPayload(entity)
	if !ok {
		return ""
	}
	return strings.TrimSpace(payload.StageLabel)
}

func nextStageLabelForRun(entity Entity) string {
	current := runStageLabel(entity)
	if !stageLabelRequiresFollowUp(current) {
		return ""
	}
	order := []string{"intake", "vision", "prd", "arch", "design", "plan", "dev", "doc-audit", "qa", "release", "postdeploy", "ops", "self-improve"}
	currentStage := strings.Split(strings.TrimPrefix(strings.ToLower(strings.TrimSpace(current)), "run:"), ":")[0]
	for idx, stage := range order {
		if stage != currentStage {
			continue
		}
		if idx+1 >= len(order) {
			return ""
		}
		if descriptor, ok := nextstepdomain.DefaultLabels().DescriptorByStage(order[idx+1]); ok {
			return descriptor.RunLabel
		}
		return "run:" + order[idx+1]
	}
	return ""
}

func runIssueRef(entity Entity) string {
	payload, ok := decodeRunPayload(entity)
	if !ok {
		return ""
	}
	return strings.TrimSpace(payload.IssueRef)
}

func buildMissingPullRequestPayload(entity Entity) []byte {
	payload, ok := decodeRunPayload(entity)
	if !ok {
		return nil
	}
	return mustMarshal(valuetypes.MissionControlMissingPullRequestGapPayload{
		RepositoryFullName: strings.TrimSpace(payload.RepositoryFullName),
		IssueRef:           strings.TrimSpace(payload.IssueRef),
		RunID:              strings.TrimSpace(payload.RunID),
		StageLabel:         strings.TrimSpace(payload.StageLabel),
	})
}

func buildMissingFollowUpPayload(runEntity Entity, pullRequestEntity Entity) []byte {
	runPayload, _ := decodeRunPayload(runEntity)
	return mustMarshal(valuetypes.MissionControlMissingFollowUpGapPayload{
		RepositoryFullName: strings.TrimSpace(runPayload.RepositoryFullName),
		PullRequestRef:     pullRequestEntity.EntityExternalKey,
		RunID:              strings.TrimSpace(runPayload.RunID),
		StageLabel:         strings.TrimSpace(runPayload.StageLabel),
	})
}

func decodeRunPayload(entity Entity) (valuetypes.MissionControlRunProjectionPayload, bool) {
	return decodeProjectionPayload[valuetypes.MissionControlRunProjectionPayload](entity, enumtypes.MissionControlEntityKindRun)
}

func decodeWorkItemPayload(entity Entity) (valuetypes.MissionControlWorkItemProjectionPayload, bool) {
	return decodeProjectionPayload[valuetypes.MissionControlWorkItemProjectionPayload](entity, enumtypes.MissionControlEntityKindWorkItem)
}

func decodePullRequestPayload(entity Entity) (valuetypes.MissionControlPullRequestProjectionPayload, bool) {
	return decodeProjectionPayload[valuetypes.MissionControlPullRequestProjectionPayload](entity, enumtypes.MissionControlEntityKindPullRequest)
}

func decodeProjectionPayload[T any](entity Entity, kind enumtypes.MissionControlEntityKind) (T, bool) {
	var payload T
	if entity.EntityKind != kind || len(entity.DetailPayloadJSON) == 0 {
		return payload, false
	}
	if err := json.Unmarshal(entity.DetailPayloadJSON, &payload); err != nil {
		var zero T
		return zero, false
	}
	return payload, true
}

func projectionTimestamp(entity Entity) time.Time {
	if entity.ProviderUpdatedAt != nil && !entity.ProviderUpdatedAt.IsZero() {
		return entity.ProviderUpdatedAt.UTC()
	}
	return entity.ProjectedAt.UTC()
}

func entityLastActivity(entity Entity) *time.Time {
	if entity.LastTimelineAt != nil && !entity.LastTimelineAt.IsZero() {
		value := entity.LastTimelineAt.UTC()
		return &value
	}
	if entity.ProviderUpdatedAt != nil && !entity.ProviderUpdatedAt.IsZero() {
		value := entity.ProviderUpdatedAt.UTC()
		return &value
	}
	if entity.ProjectedAt.IsZero() {
		return nil
	}
	value := entity.ProjectedAt.UTC()
	return &value
}

func lastRelevantActivity(entity Entity, fallback time.Time) time.Time {
	if activity := entityLastActivity(entity); activity != nil {
		return activity.UTC()
	}
	if fallback.IsZero() {
		return time.Now().UTC()
	}
	return fallback.UTC()
}

func matchesWorkspaceStatePreset(entity Entity, preset enumtypes.MissionControlWorkspaceStatePreset) bool {
	if preset == enumtypes.MissionControlWorkspaceStatePresetAllActive {
		return entity.ActiveState != enumtypes.MissionControlActiveStateArchived
	}
	return string(entity.ActiveState) == string(preset)
}

func normalizeWorkspaceStatePreset(preset enumtypes.MissionControlWorkspaceStatePreset) enumtypes.MissionControlWorkspaceStatePreset {
	switch preset {
	case enumtypes.MissionControlWorkspaceStatePresetWorking,
		enumtypes.MissionControlWorkspaceStatePresetWaiting,
		enumtypes.MissionControlWorkspaceStatePresetBlocked,
		enumtypes.MissionControlWorkspaceStatePresetReview,
		enumtypes.MissionControlWorkspaceStatePresetRecentCriticalUpdates:
		return preset
	default:
		return enumtypes.MissionControlWorkspaceStatePresetAllActive
	}
}

func matchesWorkspaceSearch(entity Entity, search string) bool {
	if search == "" {
		return true
	}
	return strings.Contains(strings.ToLower(entity.EntityExternalKey), search) ||
		strings.Contains(strings.ToLower(entity.Title), search)
}

func workspaceSecondaryCandidate(
	entity Entity,
	preset enumtypes.MissionControlWorkspaceStatePreset,
	search string,
) bool {
	if !matchesWorkspaceSearch(entity, search) {
		return false
	}
	if entity.CoverageClass == enumtypes.MissionControlCoverageClassRecentClosedContext {
		return true
	}
	return entity.CoverageClass == enumtypes.MissionControlCoverageClassOpenPrimary &&
		!matchesWorkspaceStatePreset(entity, preset)
}

func selectWorkspaceRootIDs(graph workspaceGraph, primaryIDs map[int64]struct{}, includedIDs map[int64]struct{}) []int64 {
	roots := make([]int64, 0)
	for entityID := range primaryIDs {
		entity := graph.entityByID[entityID]
		if entity.EntityKind == enumtypes.MissionControlEntityKindDiscussion || entity.EntityKind == enumtypes.MissionControlEntityKindWorkItem {
			roots = append(roots, entityID)
		}
	}
	if len(roots) == 0 {
		for entityID := range primaryIDs {
			if len(primaryIncomingRelations(graph, entityID, includedIDs)) == 0 {
				roots = append(roots, entityID)
			}
		}
	}
	if len(roots) == 0 {
		for entityID := range includedIDs {
			roots = append(roots, entityID)
		}
	}
	sort.Slice(roots, func(i, j int) bool {
		left := graph.entityByID[roots[i]]
		right := graph.entityByID[roots[j]]
		leftActivity := derefTime(entityLastActivity(left), left.ProjectedAt)
		rightActivity := derefTime(entityLastActivity(right), right.ProjectedAt)
		if !leftActivity.Equal(rightActivity) {
			return leftActivity.After(rightActivity)
		}
		return left.ID < right.ID
	})
	return roots
}

func resolveWorkspaceRootID(graph workspaceGraph, entityID int64, rootIDs []int64) int64 {
	if len(rootIDs) == 0 {
		return entityID
	}
	rootSet := make(map[int64]struct{}, len(rootIDs))
	for _, rootID := range rootIDs {
		rootSet[rootID] = struct{}{}
	}
	if _, ok := rootSet[entityID]; ok {
		return entityID
	}

	queue := append([]int64(nil), entityID)
	seen := map[int64]struct{}{entityID: {}}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, relation := range graph.incoming[current] {
			if !relationIsPrimaryPath(relation.RelationKind) {
				continue
			}
			if _, ok := seen[relation.SourceEntityID]; ok {
				continue
			}
			if _, ok := rootSet[relation.SourceEntityID]; ok {
				return relation.SourceEntityID
			}
			seen[relation.SourceEntityID] = struct{}{}
			queue = append(queue, relation.SourceEntityID)
		}
	}
	return rootIDs[0]
}

func primaryIncomingRelations(graph workspaceGraph, entityID int64, includedIDs map[int64]struct{}) []Relation {
	out := make([]Relation, 0, len(graph.incoming[entityID]))
	for _, relation := range graph.incoming[entityID] {
		if !relationIsPrimaryPath(relation.RelationKind) || !idInSet(includedIDs, relation.SourceEntityID) {
			continue
		}
		out = append(out, relation)
	}
	return out
}

func relationIsPrimaryPath(kind enumtypes.MissionControlRelationKind) bool {
	switch kind {
	case enumtypes.MissionControlRelationKindFormalizedFrom,
		enumtypes.MissionControlRelationKindSpawnedRun,
		enumtypes.MissionControlRelationKindProducedPullRequest,
		enumtypes.MissionControlRelationKindContinuesWith:
		return true
	default:
		return false
	}
}

func workspaceColumnIndex(kind enumtypes.MissionControlEntityKind) int32 {
	switch kind {
	case enumtypes.MissionControlEntityKindDiscussion, enumtypes.MissionControlEntityKindWorkItem:
		return 0
	case enumtypes.MissionControlEntityKindRun:
		return 1
	case enumtypes.MissionControlEntityKindPullRequest:
		return 2
	default:
		return 3
	}
}

func continuityGapSeedKey(subjectEntityID int64, gapKind enumtypes.MissionControlGapKind) string {
	return fmt.Sprintf("%d/%s", subjectEntityID, string(gapKind))
}

func idInSet(set map[int64]struct{}, id int64) bool {
	_, ok := set[id]
	return ok
}

func derefTime(value *time.Time, fallback time.Time) time.Time {
	if value == nil || value.IsZero() {
		return fallback.UTC()
	}
	return value.UTC()
}

func earlierTimePtr(current *time.Time, candidate time.Time) *time.Time {
	return selectTimePtr(current, candidate, false)
}

func laterTimePtr(current *time.Time, candidate time.Time) *time.Time {
	return selectTimePtr(current, candidate, true)
}

func selectTimePtr(current *time.Time, candidate time.Time, preferLater bool) *time.Time {
	if candidate.IsZero() {
		return current
	}
	if current == nil {
		value := candidate.UTC()
		return &value
	}
	if (!preferLater && candidate.Before(*current)) || (preferLater && candidate.After(*current)) {
		value := candidate.UTC()
		return &value
	}
	return current
}

func buildPreviewID(
	projectID string,
	nodeRef valuetypes.MissionControlEntityRef,
	targetLabel string,
	expectedProjectionVersion int64,
	resolvedGapIDs []int64,
	remainingGapIDs []int64,
) string {
	raw := strings.Builder{}
	raw.WriteString(strings.TrimSpace(projectID))
	raw.WriteString("|")
	raw.WriteString(string(nodeRef.EntityKind))
	raw.WriteString("|")
	raw.WriteString(nodeRef.EntityPublicID)
	raw.WriteString("|")
	raw.WriteString(strings.TrimSpace(targetLabel))
	raw.WriteString("|")
	raw.WriteString(strconv.FormatInt(expectedProjectionVersion, 10))
	raw.WriteString("|")
	for _, gapID := range resolvedGapIDs {
		raw.WriteString(strconv.FormatInt(gapID, 10))
		raw.WriteString(",")
	}
	raw.WriteString("|")
	for _, gapID := range remainingGapIDs {
		raw.WriteString(strconv.FormatInt(gapID, 10))
		raw.WriteString(",")
	}
	return fmt.Sprintf("preview-%08x", crc32.ChecksumIEEE([]byte(raw.String())))
}

func guessedThreadEntityRef(entity Entity, threadKind string, threadNumber int) *valuetypes.MissionControlEntityRef {
	if threadNumber <= 0 {
		return nil
	}
	repositoryFullName := repositoryFullNameFromEntity(entity)
	if repositoryFullName == "" {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(threadKind)) {
	case "issue":
		return &valuetypes.MissionControlEntityRef{
			EntityKind:     enumtypes.MissionControlEntityKindWorkItem,
			EntityPublicID: repositoryFullName + "#" + strconv.Itoa(threadNumber),
		}
	case "pull_request":
		return &valuetypes.MissionControlEntityRef{
			EntityKind:     enumtypes.MissionControlEntityKindPullRequest,
			EntityPublicID: repositoryFullName + "/pull/" + strconv.Itoa(threadNumber),
		}
	default:
		return nil
	}
}

func repositoryFullNameFromEntity(entity Entity) string {
	if payload, ok := decodeRunPayload(entity); ok {
		return strings.TrimSpace(payload.RepositoryFullName)
	}
	if payload, ok := decodeWorkItemPayload(entity); ok {
		return strings.TrimSpace(payload.RepositoryFullName)
	}
	if payload, ok := decodePullRequestPayload(entity); ok {
		return strings.TrimSpace(payload.RepositoryFullName)
	}
	return ""
}

func mustMarshal(value any) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return raw
}
