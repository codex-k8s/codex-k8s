package grpc

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/codex-k8s/kodex/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/kodex/proto/gen/go/kodex/controlplane/v1"
	missioncontroldomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/missioncontrol"
	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	missionControlWorkspaceDefaultRootLimit   = 50
	missionControlWorkspaceDefaultViewMode    = "graph"
	missionControlWorkspaceDefaultStatePreset = "all_active"
	missionControlActivityPreviewLimit        = 10
)

type missionControlWorkspaceQuery struct {
	viewMode    string
	statePreset string
	search      string
	offset      int
	rootLimit   int
}

type missionControlResolvedNodeDetails struct {
	projectID string
	details   missioncontroldomain.EntityDetails
}

type missionControlWorkspaceProjectSnapshot struct {
	projectID string
	snapshot  missioncontroldomain.WorkspaceSnapshot
}

func (s *Server) GetMissionControlWorkspace(ctx context.Context, req *controlplanev1.GetMissionControlWorkspaceRequest) (*controlplanev1.GetMissionControlWorkspaceResponse, error) {
	if s.missionControlDomain == nil {
		return nil, status.Error(codes.FailedPrecondition, "mission control domain service is not configured")
	}
	if s.staff == nil {
		return nil, status.Error(codes.FailedPrecondition, "staff service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	principal, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	query, err := parseMissionControlWorkspaceQuery(req)
	if err != nil {
		return nil, toStatus(err)
	}
	projectIDs, err := s.visibleMissionControlProjectIDs(ctx, principal)
	if err != nil {
		return nil, toStatus(err)
	}

	projectSnapshots := make([]missionControlWorkspaceProjectSnapshot, 0, len(projectIDs))
	for _, projectID := range projectIDs {
		snapshot, getErr := s.missionControlDomain.GetWorkspace(ctx, missioncontroldomain.WorkspaceQuery{
			ProjectID:   projectID,
			StatePreset: enumtypes.MissionControlWorkspaceStatePreset(query.statePreset),
			Search:      query.search,
		})
		if getErr != nil {
			return nil, toStatus(getErr)
		}
		projectSnapshots = append(projectSnapshots, missionControlWorkspaceProjectSnapshot{
			projectID: projectID,
			snapshot:  snapshot,
		})
	}

	return buildMissionControlWorkspaceResponse(query, projectSnapshots), nil
}

func (s *Server) GetMissionControlNode(ctx context.Context, req *controlplanev1.GetMissionControlNodeRequest) (*controlplanev1.MissionControlNodeDetails, error) {
	if s.missionControlDomain == nil {
		return nil, status.Error(codes.FailedPrecondition, "mission control domain service is not configured")
	}
	if s.staff == nil {
		return nil, status.Error(codes.FailedPrecondition, "staff service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	principal, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	projectIDs, err := s.visibleMissionControlProjectIDs(ctx, principal)
	if err != nil {
		return nil, toStatus(err)
	}
	resolved, err := s.resolveMissionControlNodeDetails(
		ctx,
		projectIDs,
		enumtypes.MissionControlEntityKind(strings.TrimSpace(req.GetNodeKind())),
		strings.TrimSpace(req.GetNodePublicId()),
		missionControlActivityPreviewLimit,
	)
	if err != nil {
		return nil, toStatus(err)
	}
	return missionControlNodeDetailsToProto(resolved.details), nil
}

func (s *Server) ListMissionControlNodeActivity(ctx context.Context, req *controlplanev1.ListMissionControlNodeActivityRequest) (*controlplanev1.ListMissionControlNodeActivityResponse, error) {
	if s.missionControlDomain == nil {
		return nil, status.Error(codes.FailedPrecondition, "mission control domain service is not configured")
	}
	if s.staff == nil {
		return nil, status.Error(codes.FailedPrecondition, "staff service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	principal, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	query, err := parseMissionControlTimelineQuery(&controlplanev1.ListMissionControlTimelineRequest{
		Cursor: req.Cursor,
		Limit:  req.GetLimit(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	projectIDs, err := s.visibleMissionControlProjectIDs(ctx, principal)
	if err != nil {
		return nil, toStatus(err)
	}
	resolved, err := s.resolveMissionControlNodeDetails(
		ctx,
		projectIDs,
		enumtypes.MissionControlEntityKind(strings.TrimSpace(req.GetNodeKind())),
		strings.TrimSpace(req.GetNodePublicId()),
		query.offset+query.limit,
	)
	if err != nil {
		return nil, toStatus(err)
	}

	pageItems, nextCursor := paginateMissionControlTimeline(resolved.details.Timeline, query.offset, query.limit)
	resp := &controlplanev1.ListMissionControlNodeActivityResponse{
		Items: make([]*controlplanev1.MissionControlActivityEntry, 0, len(pageItems)),
	}
	if nextCursor != "" {
		resp.NextCursor = stringPtrOrNil(nextCursor)
	}
	for _, item := range pageItems {
		resp.Items = append(resp.Items, missionControlActivityEntryToProto(resolved.details.Node.NodeRef, item))
	}
	return resp, nil
}

func (s *Server) PreviewMissionControlLaunch(ctx context.Context, req *controlplanev1.PreviewMissionControlLaunchRequest) (*controlplanev1.MissionControlLaunchPreview, error) {
	if s.missionControlDomain == nil {
		return nil, status.Error(codes.FailedPrecondition, "mission control domain service is not configured")
	}
	if s.staff == nil {
		return nil, status.Error(codes.FailedPrecondition, "staff service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	principal, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	projectIDs, err := s.visibleMissionControlProjectIDs(ctx, principal)
	if err != nil {
		return nil, toStatus(err)
	}
	resolved, err := s.resolveMissionControlNodeDetails(
		ctx,
		projectIDs,
		enumtypes.MissionControlEntityKind(strings.TrimSpace(req.GetNodeKind())),
		strings.TrimSpace(req.GetNodePublicId()),
		0,
	)
	if err != nil {
		return nil, toStatus(err)
	}

	preview, err := s.missionControlDomain.PreviewLaunch(ctx, missioncontroldomain.LaunchPreviewParams{
		ProjectID:                 resolved.projectID,
		NodeKind:                  enumtypes.MissionControlEntityKind(strings.TrimSpace(req.GetNodeKind())),
		NodePublicID:              strings.TrimSpace(req.GetNodePublicId()),
		ThreadKind:                strings.TrimSpace(req.GetThreadKind()),
		ThreadNumber:              int(req.GetThreadNumber()),
		TargetLabel:               strings.TrimSpace(req.GetTargetLabel()),
		RemovedLabels:             append([]string(nil), req.GetRemovedLabels()...),
		ExpectedProjectionVersion: req.GetExpectedProjectionVersion(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return missionControlLaunchPreviewToProto(preview), nil
}

func (s *Server) resolveMissionControlNodeDetails(
	ctx context.Context,
	projectIDs []string,
	nodeKind enumtypes.MissionControlEntityKind,
	nodePublicID string,
	timelineLimit int,
) (missionControlResolvedNodeDetails, error) {
	if strings.TrimSpace(string(nodeKind)) == "" {
		return missionControlResolvedNodeDetails{}, errs.Validation{Field: "node_kind", Msg: "is required"}
	}
	if strings.TrimSpace(nodePublicID) == "" {
		return missionControlResolvedNodeDetails{}, errs.Validation{Field: "node_public_id", Msg: "is required"}
	}

	var (
		found        bool
		lastNotFound error
		result       missionControlResolvedNodeDetails
	)
	for _, projectID := range projectIDs {
		item, getErr := s.missionControlDomain.GetEntityDetails(ctx, missioncontroldomain.EntityDetailsQuery{
			ProjectID:      projectID,
			EntityKind:     nodeKind,
			EntityPublicID: nodePublicID,
			TimelineLimit:  timelineLimit,
		})
		if getErr == nil {
			if found {
				return missionControlResolvedNodeDetails{}, errs.FailedPrecondition{Msg: "mission control node scope is ambiguous across projects"}
			}
			found = true
			result = missionControlResolvedNodeDetails{projectID: projectID, details: item}
			continue
		}
		var notFound errs.NotFound
		if errors.As(getErr, &notFound) {
			lastNotFound = getErr
			continue
		}
		return missionControlResolvedNodeDetails{}, getErr
	}
	if found {
		return result, nil
	}
	if lastNotFound != nil {
		return missionControlResolvedNodeDetails{}, lastNotFound
	}
	return missionControlResolvedNodeDetails{}, errs.NotFound{Msg: "mission control node not found"}
}

func parseMissionControlWorkspaceQuery(req *controlplanev1.GetMissionControlWorkspaceRequest) (missionControlWorkspaceQuery, error) {
	offset, err := decodeMissionControlCursor(req.GetCursor())
	if err != nil {
		return missionControlWorkspaceQuery{}, errs.Validation{Field: "cursor", Msg: "must be a valid opaque cursor"}
	}

	viewMode := strings.TrimSpace(req.GetViewMode())
	if viewMode == "" {
		viewMode = missionControlWorkspaceDefaultViewMode
	}
	switch viewMode {
	case "graph", "list":
	default:
		return missionControlWorkspaceQuery{}, errs.Validation{Field: "view_mode", Msg: "must be graph or list"}
	}

	statePreset := strings.TrimSpace(req.GetStatePreset())
	if statePreset == "" {
		statePreset = missionControlWorkspaceDefaultStatePreset
	}
	if _, ok := missionControlActiveFilterMap()[statePreset]; !ok {
		return missionControlWorkspaceQuery{}, errs.Validation{Field: "state_preset", Msg: "is not supported"}
	}

	return missionControlWorkspaceQuery{
		viewMode:    viewMode,
		statePreset: statePreset,
		search:      strings.ToLower(strings.TrimSpace(req.GetSearch())),
		offset:      offset,
		rootLimit:   clampLimit(req.GetRootLimit(), missionControlWorkspaceDefaultRootLimit),
	}, nil
}

func buildMissionControlWorkspaceResponse(
	query missionControlWorkspaceQuery,
	projectSnapshots []missionControlWorkspaceProjectSnapshot,
) *controlplanev1.GetMissionControlWorkspaceResponse {
	rootGroups := make([]*controlplanev1.MissionControlRootGroup, 0)
	nodes := make([]*controlplanev1.MissionControlNode, 0)
	edges := make([]*controlplanev1.MissionControlEdge, 0)
	watermarks := make([]*controlplanev1.MissionControlWorkspaceWatermark, 0)

	nodeRoots := make(map[string]string)
	for _, item := range projectSnapshots {
		for _, rootGroup := range item.snapshot.RootGroups {
			rootGroups = append(rootGroups, missionControlRootGroupToProto(rootGroup))
		}
		for _, node := range item.snapshot.Nodes {
			nodeKey := missionControlNodeRefKey(node.NodeRef)
			nodeRoots[nodeKey] = node.RootNodePublicID
			nodes = append(nodes, missionControlWorkspaceNodeToProto(node))
		}
		for _, edge := range item.snapshot.Edges {
			edges = append(edges, missionControlWorkspaceEdgeToProto(edge))
		}
		for _, watermark := range item.snapshot.WorkspaceWatermarks {
			watermarks = append(watermarks, missionControlWorkspaceWatermarkToProto(watermark))
		}
	}

	sort.Slice(rootGroups, func(i, j int) bool {
		left := tsToTime(rootGroups[i].LatestActivityAt)
		right := tsToTime(rootGroups[j].LatestActivityAt)
		if !left.Equal(right) {
			return left.After(right)
		}
		return rootGroups[i].GetRootNodePublicId() < rootGroups[j].GetRootNodePublicId()
	})
	sort.Slice(nodes, func(i, j int) bool {
		leftRoot := nodeRoots[missionControlProtoNodeRefKey(nodes[i].GetNodeKind(), nodes[i].GetNodePublicId())]
		rightRoot := nodeRoots[missionControlProtoNodeRefKey(nodes[j].GetNodeKind(), nodes[j].GetNodePublicId())]
		if leftRoot != rightRoot {
			return leftRoot < rightRoot
		}
		if nodes[i].GetColumnIndex() != nodes[j].GetColumnIndex() {
			return nodes[i].GetColumnIndex() < nodes[j].GetColumnIndex()
		}
		return nodes[i].GetNodePublicId() < nodes[j].GetNodePublicId()
	})
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].GetSourceNodePublicId() != edges[j].GetSourceNodePublicId() {
			return edges[i].GetSourceNodePublicId() < edges[j].GetSourceNodePublicId()
		}
		if edges[i].GetTargetNodePublicId() != edges[j].GetTargetNodePublicId() {
			return edges[i].GetTargetNodePublicId() < edges[j].GetTargetNodePublicId()
		}
		return edges[i].GetEdgeKind() < edges[j].GetEdgeKind()
	})

	summary := missionControlWorkspaceSummary(projectSnapshots)
	pageRootGroups, nextCursor := paginateMissionControlSlice(rootGroups, query.offset, query.rootLimit)
	allowedRoots := make(map[string]struct{}, len(pageRootGroups))
	for _, root := range pageRootGroups {
		allowedRoots[root.GetRootNodePublicId()] = struct{}{}
	}
	pageNodes := make([]*controlplanev1.MissionControlNode, 0, len(nodes))
	for _, node := range nodes {
		rootID := nodeRoots[missionControlProtoNodeRefKey(node.GetNodeKind(), node.GetNodePublicId())]
		if _, ok := allowedRoots[rootID]; ok {
			pageNodes = append(pageNodes, node)
		}
	}
	pageEdges := make([]*controlplanev1.MissionControlEdge, 0, len(edges))
	for _, edge := range edges {
		sourceRoot := nodeRoots[missionControlProtoNodeRefKey(edge.GetSourceNodeKind(), edge.GetSourceNodePublicId())]
		targetRoot := nodeRoots[missionControlProtoNodeRefKey(edge.GetTargetNodeKind(), edge.GetTargetNodePublicId())]
		if sourceRoot == "" || sourceRoot != targetRoot {
			continue
		}
		if _, ok := allowedRoots[sourceRoot]; ok {
			pageEdges = append(pageEdges, edge)
		}
	}

	snapshot := &controlplanev1.MissionControlWorkspaceSnapshot{
		SnapshotId:          buildMissionControlWorkspaceSnapshotID(query, summary, pageRootGroups, pageNodes, pageEdges, watermarks),
		ViewMode:            query.viewMode,
		GeneratedAt:         timestamppb.New(missionControlWorkspaceGeneratedAt(watermarks)),
		EffectiveFilters:    missionControlWorkspaceFiltersToProto(query),
		Summary:             summary,
		WorkspaceWatermarks: watermarks,
		RootGroups:          pageRootGroups,
		Nodes:               pageNodes,
		Edges:               pageEdges,
	}
	if nextCursor != "" {
		snapshot.NextRootCursor = stringPtrOrNil(nextCursor)
	}
	return &controlplanev1.GetMissionControlWorkspaceResponse{Snapshot: snapshot}
}

func missionControlWorkspaceGeneratedAt(watermarks []*controlplanev1.MissionControlWorkspaceWatermark) time.Time {
	generatedAt := time.Time{}
	for _, watermark := range watermarks {
		candidate := tsToTime(watermark.ObservedAt)
		if candidate.After(generatedAt) {
			generatedAt = candidate
		}
	}
	if generatedAt.IsZero() {
		return time.Now().UTC()
	}
	return generatedAt.UTC()
}

func buildMissionControlWorkspaceSnapshotID(
	query missionControlWorkspaceQuery,
	summary *controlplanev1.MissionControlWorkspaceSummary,
	rootGroups []*controlplanev1.MissionControlRootGroup,
	nodes []*controlplanev1.MissionControlNode,
	edges []*controlplanev1.MissionControlEdge,
	watermarks []*controlplanev1.MissionControlWorkspaceWatermark,
) string {
	if len(rootGroups) == 0 && len(nodes) == 0 && len(edges) == 0 && len(watermarks) == 0 {
		return "mc-workspace-empty"
	}
	var builder strings.Builder
	builder.WriteString(query.viewMode)
	builder.WriteString("|")
	builder.WriteString(query.statePreset)
	builder.WriteString("|")
	builder.WriteString(query.search)
	builder.WriteString("|")
	if summary != nil {
		builder.WriteString(fmt.Sprintf("%d:%d:%d:%d:%d:%d:%d:%d:%d:%d|",
			summary.GetRootCount(),
			summary.GetNodeCount(),
			summary.GetBlockingGapCount(),
			summary.GetWarningGapCount(),
			summary.GetRecentClosedContextCount(),
			summary.GetWorkingCount(),
			summary.GetWaitingCount(),
			summary.GetBlockedCount(),
			summary.GetReviewCount(),
			summary.GetRecentCriticalUpdatesCount(),
		))
	}
	for _, root := range rootGroups {
		builder.WriteString(root.GetRootNodeKind())
		builder.WriteString(":")
		builder.WriteString(root.GetRootNodePublicId())
		builder.WriteString("|")
	}
	for _, node := range nodes {
		builder.WriteString("node=")
		builder.WriteString(node.GetNodeKind())
		builder.WriteString("/")
		builder.WriteString(node.GetNodePublicId())
		builder.WriteString("/")
		builder.WriteString(node.GetContinuityStatus())
		builder.WriteString("|")
	}
	for _, edge := range edges {
		_, _ = fmt.Fprintf(
			&builder,
			"edge=%s>%s>%s|",
			edge.GetEdgeKind(),
			edge.GetSourceNodePublicId(),
			edge.GetTargetNodePublicId(),
		)
	}
	for _, watermark := range watermarks {
		builder.WriteString(watermark.GetWatermarkKind())
		builder.WriteString(":")
		builder.WriteString(watermark.GetStatus())
		builder.WriteString("|")
	}
	sum := sha1.Sum([]byte(builder.String()))
	return "mcws_" + base64.RawURLEncoding.EncodeToString(sum[:8])
}

func missionControlWorkspaceSummary(projectSnapshots []missionControlWorkspaceProjectSnapshot) *controlplanev1.MissionControlWorkspaceSummary {
	summary := &controlplanev1.MissionControlWorkspaceSummary{}
	for _, item := range projectSnapshots {
		summary.RootCount += int32(item.snapshot.Summary.RootCount)
		summary.NodeCount += int32(item.snapshot.Summary.NodeCount)
		summary.BlockingGapCount += int32(item.snapshot.Summary.BlockingGapCount)
		summary.WarningGapCount += int32(item.snapshot.Summary.WarningGapCount)
		summary.RecentClosedContextCount += int32(item.snapshot.Summary.RecentClosedContextCount)
		summary.WorkingCount += int32(item.snapshot.Summary.WorkingCount)
		summary.WaitingCount += int32(item.snapshot.Summary.WaitingCount)
		summary.BlockedCount += int32(item.snapshot.Summary.BlockedCount)
		summary.ReviewCount += int32(item.snapshot.Summary.ReviewCount)
		summary.RecentCriticalUpdatesCount += int32(item.snapshot.Summary.RecentCriticalUpdatesCount)
	}
	return summary
}

func missionControlWorkspaceFiltersToProto(query missionControlWorkspaceQuery) *controlplanev1.MissionControlWorkspaceFilters {
	out := &controlplanev1.MissionControlWorkspaceFilters{
		OpenScope:       "open_only",
		AssignmentScope: "assigned_to_me_or_unassigned",
		StatePreset:     query.statePreset,
	}
	if query.search != "" {
		out.Search = stringPtrOrNil(query.search)
	}
	return out
}

func missionControlRootGroupToProto(group valuetypes.MissionControlWorkspaceRootGroup) *controlplanev1.MissionControlRootGroup {
	out := &controlplanev1.MissionControlRootGroup{
		RootNodeKind:     string(group.RootNodeRef.EntityKind),
		RootNodePublicId: group.RootNodeRef.EntityPublicID,
		RootTitle:        group.RootTitle,
		HasBlockingGap:   group.HasBlockingGap,
		NodeRefs:         missionControlNodeRefsToProto(group.NodeRefs),
	}
	if group.LatestActivityAt != nil {
		out.LatestActivityAt = timestamppb.New(group.LatestActivityAt.UTC())
	}
	return out
}

func missionControlWorkspaceNodeToProto(node valuetypes.MissionControlWorkspaceNode) *controlplanev1.MissionControlNode {
	out := &controlplanev1.MissionControlNode{
		NodeKind:          string(node.NodeRef.EntityKind),
		NodePublicId:      node.NodeRef.EntityPublicID,
		Title:             node.Title,
		VisibilityTier:    string(node.VisibilityTier),
		ActiveState:       string(node.ActiveState),
		ContinuityStatus:  string(node.ContinuityStatus),
		CoverageClass:     string(node.CoverageClass),
		RootNodePublicId:  node.RootNodePublicID,
		ColumnIndex:       node.ColumnIndex,
		HasBlockingGap:    node.HasBlockingGap,
		Badges:            append([]string(nil), node.Badges...),
		ProjectionVersion: node.ProjectionVersion,
	}
	if node.LastActivityAt != nil {
		out.LastActivityAt = timestamppb.New(node.LastActivityAt.UTC())
	}
	if node.ProviderReference != nil {
		out.ProviderReference = &controlplanev1.MissionControlProviderReference{
			Provider:   string(node.ProviderReference.Provider),
			ExternalId: node.ProviderReference.ExternalID,
			Url:        strings.TrimSpace(node.ProviderReference.URL),
		}
	}
	return out
}

func missionControlWorkspaceEdgeToProto(edge valuetypes.MissionControlWorkspaceEdge) *controlplanev1.MissionControlEdge {
	return &controlplanev1.MissionControlEdge{
		EdgeKind:           string(edge.RelationKind),
		SourceNodeKind:     string(edge.SourceNodeRef.EntityKind),
		SourceNodePublicId: edge.SourceNodeRef.EntityPublicID,
		TargetNodeKind:     string(edge.TargetNodeRef.EntityKind),
		TargetNodePublicId: edge.TargetNodeRef.EntityPublicID,
		VisibilityTier:     string(edge.VisibilityTier),
		SourceOfTruth:      string(edge.SourceOfTruth),
		IsPrimaryPath:      edge.IsPrimaryPath,
	}
}

func missionControlWorkspaceWatermarkToProto(item entitytypes.MissionControlWorkspaceWatermark) *controlplanev1.MissionControlWorkspaceWatermark {
	out := &controlplanev1.MissionControlWorkspaceWatermark{
		WatermarkKind: string(item.WatermarkKind),
		Status:        string(item.Status),
		Summary:       strings.TrimSpace(item.Summary),
		ObservedAt:    timestamppb.New(item.ObservedAt.UTC()),
	}
	if item.WindowStartedAt != nil {
		out.WindowStartedAt = timestamppb.New(item.WindowStartedAt.UTC())
	}
	if item.WindowEndedAt != nil {
		out.WindowEndedAt = timestamppb.New(item.WindowEndedAt.UTC())
	}
	return out
}

func missionControlNodeDetailsToProto(details missioncontroldomain.EntityDetails) *controlplanev1.MissionControlNodeDetails {
	out := &controlplanev1.MissionControlNodeDetails{
		Node:              missionControlWorkspaceNodeToProto(details.Node),
		AdjacentNodes:     make([]*controlplanev1.MissionControlNode, 0, len(details.AdjacentNodes)),
		AdjacentEdges:     make([]*controlplanev1.MissionControlEdge, 0, len(details.AdjacentEdges)),
		ContinuityGaps:    make([]*controlplanev1.MissionControlContinuityGap, 0, len(details.ContinuityGaps)),
		NodeWatermarks:    make([]*controlplanev1.MissionControlWorkspaceWatermark, 0, len(details.NodeWatermarks)),
		ActivityPreview:   make([]*controlplanev1.MissionControlActivityEntry, 0, len(details.Timeline)),
		LaunchSurfaces:    make([]*controlplanev1.MissionControlLaunchSurface, 0, len(details.LaunchSurfaces)),
		ProviderDeepLinks: missionControlProviderDeepLinks(details.Entity),
	}
	for _, node := range details.AdjacentNodes {
		out.AdjacentNodes = append(out.AdjacentNodes, missionControlWorkspaceNodeToProto(node))
	}
	for _, edge := range details.AdjacentEdges {
		out.AdjacentEdges = append(out.AdjacentEdges, missionControlWorkspaceEdgeToProto(edge))
	}
	for _, gap := range details.ContinuityGaps {
		out.ContinuityGaps = append(out.ContinuityGaps, missionControlContinuityGapToProto(gap))
	}
	for _, watermark := range details.NodeWatermarks {
		out.NodeWatermarks = append(out.NodeWatermarks, missionControlWorkspaceWatermarkToProto(watermark))
	}
	for _, item := range details.Timeline {
		out.ActivityPreview = append(out.ActivityPreview, missionControlActivityEntryToProto(details.Node.NodeRef, item))
	}
	for _, surface := range details.LaunchSurfaces {
		out.LaunchSurfaces = append(out.LaunchSurfaces, missionControlLaunchSurfaceToProto(surface))
	}
	missionControlApplyNodeDetailPayload(out, details)
	return out
}

func missionControlApplyNodeDetailPayload(out *controlplanev1.MissionControlNodeDetails, details missioncontroldomain.EntityDetails) {
	entity := details.Entity
	switch entity.EntityKind {
	case enumtypes.MissionControlEntityKindDiscussion:
		var payload valuetypes.MissionControlDiscussionProjectionPayload
		if err := json.Unmarshal(entity.DetailPayloadJSON, &payload); err != nil {
			return
		}
		nodeDetails := &controlplanev1.MissionControlDiscussionNodeDetails{
			DiscussionKind:       strings.TrimSpace(payload.DiscussionKind),
			Status:               strings.TrimSpace(payload.Status),
			Author:               strings.TrimSpace(payload.Author),
			ParticipantCount:     payload.ParticipantCount,
			LatestCommentExcerpt: strings.TrimSpace(payload.LatestCommentExcerpt),
		}
		for _, relation := range details.Relations {
			if relation.RelationKind != enumtypes.MissionControlRelationKindFormalizedFrom {
				continue
			}
			nodeDetails.FormalizationTargetRefs = append(nodeDetails.FormalizationTargetRefs, &controlplanev1.MissionControlNodeRef{
				NodeKind:     string(relation.TargetEntityRef.EntityKind),
				NodePublicId: relation.TargetEntityRef.EntityPublicID,
			})
		}
		out.DetailPayload = &controlplanev1.MissionControlNodeDetails_Discussion{Discussion: nodeDetails}
	case enumtypes.MissionControlEntityKindWorkItem:
		var payload valuetypes.MissionControlWorkItemProjectionPayload
		if err := json.Unmarshal(entity.DetailPayloadJSON, &payload); err != nil {
			return
		}
		nodeDetails := &controlplanev1.MissionControlWorkItemNodeDetails{
			RepositoryFullName: strings.TrimSpace(payload.RepositoryFullName),
			IssueNumber:        payload.IssueNumber,
			StageLabel:         strings.TrimSpace(payload.StageLabel),
			Labels:             append([]string(nil), payload.Labels...),
			Assignees:          append([]string(nil), payload.Assignees...),
		}
		for _, relation := range details.AdjacentEdges {
			switch relation.RelationKind {
			case enumtypes.MissionControlRelationKindSpawnedRun:
				if relation.SourceNodeRef.EntityPublicID == entity.EntityExternalKey {
					nodeDetails.LinkedRunRefs = append(nodeDetails.LinkedRunRefs, missionControlNodeRefToProto(relation.TargetNodeRef))
				}
			case enumtypes.MissionControlRelationKindContinuesWith:
				if relation.SourceNodeRef.EntityPublicID == entity.EntityExternalKey {
					nodeDetails.LinkedFollowUpRefs = append(nodeDetails.LinkedFollowUpRefs, missionControlNodeRefToProto(relation.TargetNodeRef))
				}
			}
		}
		if payload.LastProviderSyncAt != nil {
			nodeDetails.LastProviderSyncAt = timestamppb.New(payload.LastProviderSyncAt.UTC())
		}
		out.DetailPayload = &controlplanev1.MissionControlNodeDetails_WorkItem{WorkItem: nodeDetails}
	case enumtypes.MissionControlEntityKindRun:
		var payload valuetypes.MissionControlRunProjectionPayload
		if err := json.Unmarshal(entity.DetailPayloadJSON, &payload); err != nil {
			return
		}
		nodeDetails := &controlplanev1.MissionControlRunNodeDetails{
			RunId:              strings.TrimSpace(payload.RunID),
			AgentKey:           strings.TrimSpace(payload.AgentKey),
			RunStatus:          strings.TrimSpace(payload.LastStatus),
			RuntimeMode:        strings.TrimSpace(payload.RuntimeMode),
			TriggerLabel:       strings.TrimSpace(payload.TriggerLabel),
			BuildRef:           strings.TrimSpace(payload.BranchHead),
			CandidateNamespace: strings.TrimSpace(payload.CandidateNamespace),
		}
		if payload.StartedAt != nil {
			nodeDetails.StartedAt = timestamppb.New(payload.StartedAt.UTC())
		}
		if payload.FinishedAt != nil {
			nodeDetails.FinishedAt = timestamppb.New(payload.FinishedAt.UTC())
		}
		if linkedPR := strings.TrimSpace(payload.PullRequestRef); linkedPR != "" {
			nodeDetails.LinkedPullRequestRefs = append(nodeDetails.LinkedPullRequestRefs, &controlplanev1.MissionControlNodeRef{
				NodeKind:     string(enumtypes.MissionControlEntityKindPullRequest),
				NodePublicId: linkedPR,
			})
		}
		if issueRef := strings.TrimSpace(payload.IssueRef); issueRef != "" {
			nodeDetails.ProducedIssueRefs = append(nodeDetails.ProducedIssueRefs, &controlplanev1.MissionControlNodeRef{
				NodeKind:     string(enumtypes.MissionControlEntityKindWorkItem),
				NodePublicId: issueRef,
			})
		}
		out.DetailPayload = &controlplanev1.MissionControlNodeDetails_Run{Run: nodeDetails}
	case enumtypes.MissionControlEntityKindPullRequest:
		var payload valuetypes.MissionControlPullRequestProjectionPayload
		if err := json.Unmarshal(entity.DetailPayloadJSON, &payload); err != nil {
			return
		}
		nodeDetails := &controlplanev1.MissionControlPullRequestNodeDetails{
			RepositoryFullName: strings.TrimSpace(payload.RepositoryFullName),
			PullRequestNumber:  payload.PullRequestNumber,
			BranchHead:         strings.TrimSpace(payload.BranchHead),
			BranchBase:         strings.TrimSpace(payload.BranchBase),
			MergeState:         strings.TrimSpace(payload.MergeState),
			ReviewDecision:     strings.TrimSpace(payload.ReviewDecision),
			ChecksSummary:      strings.TrimSpace(payload.ChecksSummary),
		}
		for _, ref := range payload.LinkedIssueRefs {
			nodeDetails.LinkedIssueRefs = append(nodeDetails.LinkedIssueRefs, &controlplanev1.MissionControlNodeRef{
				NodeKind:     string(enumtypes.MissionControlEntityKindWorkItem),
				NodePublicId: strings.TrimSpace(ref),
			})
		}
		for _, relation := range details.AdjacentEdges {
			if relation.RelationKind == enumtypes.MissionControlRelationKindProducedPullRequest &&
				relation.TargetNodeRef.EntityPublicID == entity.EntityExternalKey {
				nodeDetails.LinkedRunRef = missionControlNodeRefToProto(relation.SourceNodeRef)
				break
			}
		}
		out.DetailPayload = &controlplanev1.MissionControlNodeDetails_PullRequest{PullRequest: nodeDetails}
	}
}

func missionControlContinuityGapToProto(item valuetypes.MissionControlContinuityGapView) *controlplanev1.MissionControlContinuityGap {
	out := &controlplanev1.MissionControlContinuityGap{
		GapId:               item.GapID,
		GapKind:             string(item.GapKind),
		Severity:            string(item.Severity),
		Status:              string(item.Status),
		SubjectNodeKind:     string(item.SubjectNodeRef.EntityKind),
		SubjectNodePublicId: item.SubjectNodeRef.EntityPublicID,
		DetectedAt:          timestamppb.New(item.DetectedAt.UTC()),
	}
	if item.ExpectedNodeKind != "" {
		out.ExpectedNodeKind = stringPtrOrNil(string(item.ExpectedNodeKind))
	}
	if item.ExpectedStageLabel != "" {
		out.ExpectedStageLabel = stringPtrOrNil(item.ExpectedStageLabel)
	}
	if item.ResolvedAt != nil {
		out.ResolvedAt = timestamppb.New(item.ResolvedAt.UTC())
	}
	if item.ResolutionHint != "" {
		out.ResolutionHint = stringPtrOrNil(item.ResolutionHint)
	}
	return out
}

func missionControlLaunchSurfaceToProto(item valuetypes.MissionControlLaunchSurface) *controlplanev1.MissionControlLaunchSurface {
	out := &controlplanev1.MissionControlLaunchSurface{
		ActionKind:          item.ActionKind,
		Presentation:        item.Presentation,
		ApprovalRequirement: string(item.ApprovalRequirement),
	}
	if item.BlockedReason != "" {
		out.BlockedReason = stringPtrOrNil(item.BlockedReason)
	}
	if item.CommandTemplate != nil {
		out.CommandTemplate = &controlplanev1.MissionControlStageNextStepTemplate{
			ThreadKind:          item.CommandTemplate.ThreadKind,
			ThreadNumber:        int32(item.CommandTemplate.ThreadNumber),
			TargetLabel:         item.CommandTemplate.TargetLabel,
			RemovedLabels:       append([]string(nil), item.CommandTemplate.RemovedLabels...),
			ApprovalRequirement: string(item.CommandTemplate.ApprovalRequirement),
			ExpectedGapIds:      append([]int64(nil), item.CommandTemplate.ExpectedGapIDs...),
		}
		if item.CommandTemplate.DisplayVariant != "" {
			out.CommandTemplate.DisplayVariant = stringPtrOrNil(item.CommandTemplate.DisplayVariant)
		}
	}
	return out
}

func missionControlActivityEntryToProto(ref valuetypes.MissionControlEntityRef, item missioncontroldomain.TimelineEntry) *controlplanev1.MissionControlActivityEntry {
	out := &controlplanev1.MissionControlActivityEntry{
		EntryId:      item.EntryExternalKey,
		NodeKind:     string(ref.EntityKind),
		NodePublicId: ref.EntityPublicID,
		SourceKind:   string(item.SourceKind),
		SourceRef:    item.EntryExternalKey,
		OccurredAt:   timestamppb.New(item.OccurredAt.UTC()),
		Summary:      item.Summary,
		IsReadOnly:   item.IsReadOnly,
	}
	if body := strings.TrimSpace(item.BodyMarkdown); body != "" {
		out.BodyMarkdown = stringPtrOrNil(body)
	}
	if providerURL := strings.TrimSpace(item.ProviderURL); providerURL != "" {
		out.ProviderUrl = stringPtrOrNil(providerURL)
	}
	return out
}

func missionControlLaunchPreviewToProto(item missioncontroldomain.LaunchPreview) *controlplanev1.MissionControlLaunchPreview {
	out := &controlplanev1.MissionControlLaunchPreview{
		PreviewId:           item.PreviewID,
		ApprovalRequirement: string(item.ApprovalRequirement),
		LabelDiff: &controlplanev1.MissionControlLaunchPreviewLabelDiff{
			RemovedLabels: append([]string(nil), item.LabelDiff.RemovedLabels...),
			AddedLabels:   append([]string(nil), item.LabelDiff.AddedLabels...),
			FinalLabels:   append([]string(nil), item.LabelDiff.FinalLabels...),
		},
		ContinuityEffect: &controlplanev1.MissionControlLaunchPreviewContinuityEffect{
			ResolvedGapIds:    append([]int64(nil), item.ContinuityEffect.ResolvedGapIDs...),
			RemainingGapIds:   append([]int64(nil), item.ContinuityEffect.RemainingGapIDs...),
			ResultingNodeRefs: missionControlNodeRefsToProto(item.ContinuityEffect.ResultingNodeRefs),
			ProviderRedirects: append([]string(nil), item.ContinuityEffect.ProviderRedirects...),
		},
	}
	if item.BlockingReason != "" {
		out.BlockingReason = stringPtrOrNil(item.BlockingReason)
	}
	return out
}

func missionControlNodeRefToProto(ref valuetypes.MissionControlEntityRef) *controlplanev1.MissionControlNodeRef {
	return &controlplanev1.MissionControlNodeRef{
		NodeKind:     string(ref.EntityKind),
		NodePublicId: ref.EntityPublicID,
	}
}

func missionControlNodeRefsToProto(items []valuetypes.MissionControlEntityRef) []*controlplanev1.MissionControlNodeRef {
	out := make([]*controlplanev1.MissionControlNodeRef, 0, len(items))
	for _, item := range items {
		out = append(out, missionControlNodeRefToProto(item))
	}
	return out
}

func missionControlNodeRefKey(ref valuetypes.MissionControlEntityRef) string {
	return string(ref.EntityKind) + ":" + ref.EntityPublicID
}

func missionControlProtoNodeRefKey(kind string, publicID string) string {
	return strings.TrimSpace(kind) + ":" + strings.TrimSpace(publicID)
}
