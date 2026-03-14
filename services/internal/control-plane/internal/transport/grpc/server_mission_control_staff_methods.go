package grpc

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	missioncontroldomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/missioncontrol"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/staff"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	missionControlSnapshotDefaultLimit  = 50
	missionControlSnapshotFetchLimit    = 500
	missionControlTimelineDefaultLimit  = 50
	missionControlVisibleProjectsLimit  = 1000
	missionControlDefaultSnapshotView   = "board"
	missionControlDefaultActiveFilter   = "all_active"
	missionControlDefaultFreshnessFresh = "fresh"
	missionControlFreshnessStale        = "stale"
	missionControlFreshnessDegraded     = "degraded"
)

type missionControlSnapshotQuery struct {
	viewMode     string
	activeFilter string
	search       string
	offset       int
	limit        int
}

type missionControlTimelineQuery struct {
	offset int
	limit  int
}

type missionControlSnapshotEntity struct {
	projectID string
	entity    missioncontroldomain.Entity
}

type missionControlDetailPayloadWorkItem struct {
	RepositoryFullName string `json:"repository_full_name"`
	IssueNumber        int64  `json:"issue_number"`
	IssueURL           string `json:"issue_url,omitempty"`
	LastRunID          string `json:"last_run_id,omitempty"`
	LastStatus         string `json:"last_status,omitempty"`
	TriggerKind        string `json:"trigger_kind,omitempty"`
}

type missionControlDetailPayloadPullRequest struct {
	RepositoryFullName string `json:"repository_full_name"`
	PullRequestNumber  int64  `json:"pull_request_number"`
	PullRequestURL     string `json:"pull_request_url,omitempty"`
	LastRunID          string `json:"last_run_id,omitempty"`
	LastStatus         string `json:"last_status,omitempty"`
}

type missionControlDetailPayloadAgent struct {
	AgentKey    string `json:"agent_key"`
	LastRunID   string `json:"last_run_id,omitempty"`
	LastStatus  string `json:"last_status,omitempty"`
	LastRunRepo string `json:"last_run_repository,omitempty"`
}

func (s *Server) GetMissionControlSnapshot(ctx context.Context, req *controlplanev1.GetMissionControlSnapshotRequest) (*controlplanev1.GetMissionControlSnapshotResponse, error) {
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
	query, err := parseMissionControlSnapshotQuery(req)
	if err != nil {
		return nil, toStatus(err)
	}
	projectIDs, err := s.visibleMissionControlProjectIDs(ctx, principal)
	if err != nil {
		return nil, toStatus(err)
	}

	entities, relations, err := s.collectMissionControlActiveSet(ctx, projectIDs, query.activeFilter, query.limit, query.search != "")
	if err != nil {
		return nil, toStatus(err)
	}

	filteredEntities := filterMissionControlEntities(entities, query.search)
	filteredRelations := filterMissionControlRelations(relations, filteredEntities)
	summary := missionControlSnapshotSummary(filteredEntities)
	pageEntities, nextCursor := paginateMissionControlEntities(filteredEntities, query.offset, query.limit)
	pageRelations := filterMissionControlRelations(filteredRelations, pageEntities)
	relationCounts := missionControlRelationCounts(pageRelations)
	freshnessStatus, generatedAt, staleAfter := missionControlSnapshotFreshness(pageEntities)
	snapshotID := buildMissionControlSnapshotID(query, pageEntities, pageRelations, generatedAt, staleAfter)

	resp := &controlplanev1.GetMissionControlSnapshotResponse{
		Snapshot: &controlplanev1.MissionControlDashboardSnapshot{
			SnapshotId:      snapshotID,
			ViewMode:        query.viewMode,
			FreshnessStatus: freshnessStatus,
			GeneratedAt:     timestamppb.New(generatedAt.UTC()),
			StaleAfter:      timestamppb.New(staleAfter.UTC()),
			Summary:         summary,
			Entities:        make([]*controlplanev1.MissionControlEntityCard, 0, len(pageEntities)),
			Relations:       make([]*controlplanev1.MissionControlRelation, 0, len(pageRelations)),
		},
	}
	if nextCursor != "" {
		resp.Snapshot.NextPageCursor = stringPtrOrNil(nextCursor)
	}
	for _, entity := range pageEntities {
		resp.Snapshot.Entities = append(resp.Snapshot.Entities, missionControlEntityCardToProto(entity.entity, relationCounts[missionControlEntityKey(entity.entity)]))
	}
	for _, relation := range pageRelations {
		resp.Snapshot.Relations = append(resp.Snapshot.Relations, missionControlRelationToProto(relation))
	}
	return resp, nil
}

func (s *Server) GetMissionControlEntity(ctx context.Context, req *controlplanev1.GetMissionControlEntityRequest) (*controlplanev1.MissionControlEntityDetails, error) {
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
	timelineLimit := clampLimit(req.GetTimelineLimit(), missionControlTimelineDefaultLimit)
	details, err := s.resolveMissionControlEntityDetails(
		ctx,
		projectIDs,
		enumtypes.MissionControlEntityKind(strings.TrimSpace(req.GetEntityKind())),
		strings.TrimSpace(req.GetEntityPublicId()),
		timelineLimit,
	)
	if err != nil {
		return nil, toStatus(err)
	}

	return missionControlEntityDetailsToProto(details), nil
}

func (s *Server) ListMissionControlTimeline(ctx context.Context, req *controlplanev1.ListMissionControlTimelineRequest) (*controlplanev1.ListMissionControlTimelineResponse, error) {
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
	query, err := parseMissionControlTimelineQuery(req)
	if err != nil {
		return nil, toStatus(err)
	}
	projectIDs, err := s.visibleMissionControlProjectIDs(ctx, principal)
	if err != nil {
		return nil, toStatus(err)
	}
	details, err := s.resolveMissionControlEntityDetails(
		ctx,
		projectIDs,
		enumtypes.MissionControlEntityKind(strings.TrimSpace(req.GetEntityKind())),
		strings.TrimSpace(req.GetEntityPublicId()),
		query.offset+query.limit,
	)
	if err != nil {
		return nil, toStatus(err)
	}
	pageItems, nextCursor := paginateMissionControlTimeline(details.Timeline, query.offset, query.limit)
	resp := &controlplanev1.ListMissionControlTimelineResponse{
		Items: make([]*controlplanev1.MissionControlTimelineEntry, 0, len(pageItems)),
	}
	if nextCursor != "" {
		resp.NextCursor = stringPtrOrNil(nextCursor)
	}
	for _, item := range pageItems {
		resp.Items = append(resp.Items, missionControlTimelineEntryToProto(details.Entity.EntityKind, details.Entity.EntityExternalKey, item))
	}
	return resp, nil
}

func (s *Server) SubmitMissionControlCommand(ctx context.Context, req *controlplanev1.SubmitMissionControlCommandRequest) (*controlplanev1.MissionControlCommandState, error) {
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
	projectID := strings.TrimSpace(req.GetProjectId())
	if projectID == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if _, err := s.staff.GetProject(ctx, principal, projectID); err != nil {
		return nil, toStatus(err)
	}
	payload, err := missionControlCommandPayloadFromProto(req)
	if err != nil {
		return nil, toStatus(err)
	}

	admission, err := s.missionControlDomain.SubmitCommand(ctx, missioncontroldomain.SubmitCommandParams{
		ProjectID:                 projectID,
		ActorID:                   principal.UserID,
		CorrelationID:             strings.TrimSpace(req.GetCorrelationId()),
		CommandKind:               enumtypes.MissionControlCommandKind(strings.TrimSpace(req.GetCommandKind())),
		TargetEntityRef:           missionControlTargetRefFromProto(req),
		BusinessIntentKey:         strings.TrimSpace(req.GetBusinessIntentKey()),
		ExpectedProjectionVersion: req.GetExpectedProjectionVersion(),
		Payload:                   payload,
		RequestedAt:               tsToTime(req.GetRequestedAt()),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return s.commandStateToProto(ctx, admission.Command)
}

func (s *Server) GetMissionControlCommand(ctx context.Context, req *controlplanev1.GetMissionControlCommandRequest) (*controlplanev1.MissionControlCommandState, error) {
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
	commandID := strings.TrimSpace(req.GetCommandId())
	if commandID == "" {
		return nil, status.Error(codes.InvalidArgument, "command_id is required")
	}

	var lastNotFound error
	for _, projectID := range projectIDs {
		statusView, getErr := s.missionControlDomain.GetCommandStatus(ctx, projectID, commandID)
		if getErr == nil {
			return missionControlCommandStatusViewToProto(statusView), nil
		}
		var notFound errs.NotFound
		if errors.As(getErr, &notFound) {
			lastNotFound = getErr
			continue
		}
		return nil, toStatus(getErr)
	}
	if lastNotFound != nil {
		return nil, toStatus(lastNotFound)
	}
	return nil, status.Error(codes.NotFound, "mission control command not found")
}

func (s *Server) visibleMissionControlProjectIDs(ctx context.Context, principal staff.Principal) ([]string, error) {
	projects, err := s.staff.ListProjects(ctx, principal, missionControlVisibleProjectsLimit)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(projects))
	for _, project := range projects {
		projectID := strings.TrimSpace(project.ID)
		if projectID == "" {
			continue
		}
		out = append(out, projectID)
	}
	return out, nil
}

func (s *Server) collectMissionControlActiveSet(
	ctx context.Context,
	projectIDs []string,
	activeFilter string,
	requestLimit int,
	searchRequested bool,
) ([]missionControlSnapshotEntity, []valuetypes.MissionControlRelationView, error) {
	entities := make([]missionControlSnapshotEntity, 0)
	relations := make([]valuetypes.MissionControlRelationView, 0)
	fetchLimit := requestLimit
	if fetchLimit <= 0 {
		fetchLimit = missionControlSnapshotDefaultLimit
	}
	if searchRequested || fetchLimit < missionControlSnapshotFetchLimit {
		fetchLimit = missionControlSnapshotFetchLimit
	}

	for _, projectID := range projectIDs {
		activeSet, err := s.missionControlDomain.ListActiveSet(ctx, missioncontroldomain.ActiveSetQuery{
			ProjectID:    projectID,
			ActiveStates: missionControlActiveStates(activeFilter),
			Limit:        fetchLimit,
		})
		if err != nil {
			return nil, nil, err
		}
		for _, entity := range activeSet.Entities {
			entities = append(entities, missionControlSnapshotEntity{projectID: projectID, entity: entity})
		}
		relations = append(relations, activeSet.Relations...)
	}

	sort.SliceStable(entities, func(i int, j int) bool {
		left := entities[i].entity
		right := entities[j].entity
		leftTime := time.Time{}
		rightTime := time.Time{}
		if left.LastTimelineAt != nil {
			leftTime = left.LastTimelineAt.UTC()
		}
		if right.LastTimelineAt != nil {
			rightTime = right.LastTimelineAt.UTC()
		}
		if !leftTime.Equal(rightTime) {
			return leftTime.After(rightTime)
		}
		if left.ActiveState != right.ActiveState {
			return strings.Compare(string(left.ActiveState), string(right.ActiveState)) < 0
		}
		return strings.Compare(left.EntityExternalKey, right.EntityExternalKey) < 0
	})

	return entities, relations, nil
}

func (s *Server) resolveMissionControlEntityDetails(
	ctx context.Context,
	projectIDs []string,
	entityKind enumtypes.MissionControlEntityKind,
	entityPublicID string,
	timelineLimit int,
) (missioncontroldomain.EntityDetails, error) {
	if strings.TrimSpace(string(entityKind)) == "" {
		return missioncontroldomain.EntityDetails{}, errs.Validation{Field: "entity_kind", Msg: "is required"}
	}
	if strings.TrimSpace(entityPublicID) == "" {
		return missioncontroldomain.EntityDetails{}, errs.Validation{Field: "entity_public_id", Msg: "is required"}
	}

	var (
		found        bool
		lastNotFound error
		result       missioncontroldomain.EntityDetails
	)
	for _, projectID := range projectIDs {
		item, err := s.missionControlDomain.GetEntityDetails(ctx, missioncontroldomain.EntityDetailsQuery{
			ProjectID:      projectID,
			EntityKind:     entityKind,
			EntityPublicID: entityPublicID,
			TimelineLimit:  timelineLimit,
		})
		if err == nil {
			if found {
				return missioncontroldomain.EntityDetails{}, errs.FailedPrecondition{Msg: "mission control entity scope is ambiguous across projects"}
			}
			found = true
			result = item
			continue
		}
		var notFound errs.NotFound
		if errors.As(err, &notFound) {
			lastNotFound = err
			continue
		}
		return missioncontroldomain.EntityDetails{}, err
	}
	if found {
		return result, nil
	}
	if lastNotFound != nil {
		return missioncontroldomain.EntityDetails{}, lastNotFound
	}
	return missioncontroldomain.EntityDetails{}, errs.NotFound{Msg: "mission control entity not found"}
}

func parseMissionControlSnapshotQuery(req *controlplanev1.GetMissionControlSnapshotRequest) (missionControlSnapshotQuery, error) {
	limit := clampLimit(req.GetLimit(), missionControlSnapshotDefaultLimit)
	offset, err := decodeMissionControlCursor(req.GetCursor())
	if err != nil {
		return missionControlSnapshotQuery{}, errs.Validation{Field: "cursor", Msg: "must be a valid opaque cursor"}
	}

	viewMode := strings.TrimSpace(req.GetViewMode())
	if viewMode == "" {
		viewMode = missionControlDefaultSnapshotView
	}
	switch viewMode {
	case "board", "list":
	default:
		return missionControlSnapshotQuery{}, errs.Validation{Field: "view_mode", Msg: "must be board or list"}
	}

	activeFilter := strings.TrimSpace(req.GetActiveFilter())
	if activeFilter == "" {
		activeFilter = missionControlDefaultActiveFilter
	}
	if _, ok := missionControlActiveFilterMap()[activeFilter]; !ok {
		return missionControlSnapshotQuery{}, errs.Validation{Field: "active_filter", Msg: "is not supported"}
	}

	return missionControlSnapshotQuery{
		viewMode:     viewMode,
		activeFilter: activeFilter,
		search:       strings.ToLower(strings.TrimSpace(req.GetSearch())),
		offset:       offset,
		limit:        limit,
	}, nil
}

func parseMissionControlTimelineQuery(req *controlplanev1.ListMissionControlTimelineRequest) (missionControlTimelineQuery, error) {
	offset, err := decodeMissionControlCursor(req.GetCursor())
	if err != nil {
		return missionControlTimelineQuery{}, errs.Validation{Field: "cursor", Msg: "must be a valid opaque cursor"}
	}
	return missionControlTimelineQuery{
		offset: offset,
		limit:  clampLimit(req.GetLimit(), missionControlTimelineDefaultLimit),
	}, nil
}

func missionControlActiveFilterMap() map[string][]enumtypes.MissionControlActiveState {
	return map[string][]enumtypes.MissionControlActiveState{
		"all_active": {
			enumtypes.MissionControlActiveStateWorking,
			enumtypes.MissionControlActiveStateWaiting,
			enumtypes.MissionControlActiveStateBlocked,
			enumtypes.MissionControlActiveStateReview,
			enumtypes.MissionControlActiveStateRecentCriticalUpdates,
		},
		"working":                 {enumtypes.MissionControlActiveStateWorking},
		"waiting":                 {enumtypes.MissionControlActiveStateWaiting},
		"blocked":                 {enumtypes.MissionControlActiveStateBlocked},
		"review":                  {enumtypes.MissionControlActiveStateReview},
		"recent_critical_updates": {enumtypes.MissionControlActiveStateRecentCriticalUpdates},
	}
}

func missionControlActiveStates(activeFilter string) []enumtypes.MissionControlActiveState {
	items := missionControlActiveFilterMap()[strings.TrimSpace(activeFilter)]
	out := make([]enumtypes.MissionControlActiveState, 0, len(items))
	out = append(out, items...)
	return out
}

func filterMissionControlEntities(items []missionControlSnapshotEntity, search string) []missionControlSnapshotEntity {
	if search == "" {
		return items
	}
	out := make([]missionControlSnapshotEntity, 0, len(items))
	for _, item := range items {
		if missionControlEntityMatchesSearch(item.entity, search) {
			out = append(out, item)
		}
	}
	return out
}

func filterMissionControlRelations(
	relations []valuetypes.MissionControlRelationView,
	entities []missionControlSnapshotEntity,
) []valuetypes.MissionControlRelationView {
	if len(relations) == 0 || len(entities) == 0 {
		return nil
	}
	allowed := make(map[string]struct{}, len(entities))
	for _, entity := range entities {
		allowed[missionControlEntityKey(entity.entity)] = struct{}{}
	}
	out := make([]valuetypes.MissionControlRelationView, 0, len(relations))
	for _, relation := range relations {
		sourceKey := missionControlEntityRefKey(relation.SourceEntityRef)
		targetKey := missionControlEntityRefKey(relation.TargetEntityRef)
		if _, ok := allowed[sourceKey]; !ok {
			continue
		}
		if _, ok := allowed[targetKey]; !ok {
			continue
		}
		out = append(out, relation)
	}
	return out
}

func missionControlSnapshotSummary(items []missionControlSnapshotEntity) *controlplanev1.MissionControlSnapshotSummary {
	summary := &controlplanev1.MissionControlSnapshotSummary{
		TotalEntities: int32(len(items)),
	}
	for _, item := range items {
		switch item.entity.ActiveState {
		case enumtypes.MissionControlActiveStateWorking:
			summary.WorkingCount++
		case enumtypes.MissionControlActiveStateWaiting:
			summary.WaitingCount++
		case enumtypes.MissionControlActiveStateBlocked:
			summary.BlockedCount++
		case enumtypes.MissionControlActiveStateReview:
			summary.ReviewCount++
		case enumtypes.MissionControlActiveStateRecentCriticalUpdates:
			summary.RecentCriticalUpdatesCount++
		}
	}
	return summary
}

func paginateMissionControlEntities(items []missionControlSnapshotEntity, offset int, limit int) ([]missionControlSnapshotEntity, string) {
	return paginateMissionControlSlice(items, offset, limit)
}

func paginateMissionControlTimeline(items []missioncontroldomain.TimelineEntry, offset int, limit int) ([]missioncontroldomain.TimelineEntry, string) {
	return paginateMissionControlSlice(items, offset, limit)
}

func paginateMissionControlSlice[T any](items []T, offset int, limit int) ([]T, string) {
	if offset >= len(items) {
		return nil, ""
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	page := items[offset:end]
	nextCursor := ""
	if end < len(items) {
		nextCursor = encodeMissionControlCursor(end)
	}
	return page, nextCursor
}

func missionControlRelationCounts(relations []valuetypes.MissionControlRelationView) map[string]int32 {
	out := make(map[string]int32, len(relations)*2)
	for _, relation := range relations {
		out[missionControlEntityRefKey(relation.SourceEntityRef)]++
		out[missionControlEntityRefKey(relation.TargetEntityRef)]++
	}
	return out
}

func missionControlSnapshotFreshness(items []missionControlSnapshotEntity) (string, time.Time, time.Time) {
	now := time.Now().UTC()
	if len(items) == 0 {
		return missionControlDefaultFreshnessFresh, now, now
	}

	status := missionControlDefaultFreshnessFresh
	generatedAt := time.Time{}
	staleAfter := time.Time{}
	for _, item := range items {
		entity := item.entity
		if entity.ProjectedAt.After(generatedAt) {
			generatedAt = entity.ProjectedAt.UTC()
		}
		if entity.StaleAfter != nil {
			candidate := entity.StaleAfter.UTC()
			if staleAfter.IsZero() || candidate.Before(staleAfter) {
				staleAfter = candidate
			}
		}
		switch entity.SyncStatus {
		case enumtypes.MissionControlSyncStatusDegraded, enumtypes.MissionControlSyncStatusFailed:
			status = missionControlFreshnessDegraded
		case enumtypes.MissionControlSyncStatusPendingSync:
			if status != missionControlFreshnessDegraded {
				status = missionControlFreshnessStale
			}
		}
		if status != missionControlFreshnessDegraded && entity.StaleAfter != nil && now.After(entity.StaleAfter.UTC()) {
			status = missionControlFreshnessStale
		}
	}
	if generatedAt.IsZero() {
		generatedAt = now
	}
	if staleAfter.IsZero() {
		staleAfter = generatedAt
	}
	return status, generatedAt, staleAfter
}

func buildMissionControlSnapshotID(
	query missionControlSnapshotQuery,
	entities []missionControlSnapshotEntity,
	relations []valuetypes.MissionControlRelationView,
	generatedAt time.Time,
	staleAfter time.Time,
) string {
	var builder strings.Builder
	builder.WriteString(query.viewMode)
	builder.WriteByte('|')
	builder.WriteString(query.activeFilter)
	builder.WriteByte('|')
	builder.WriteString(query.search)
	builder.WriteByte('|')
	builder.WriteString(strconv.Itoa(query.limit))
	builder.WriteByte('|')
	builder.WriteString(generatedAt.UTC().Format(time.RFC3339Nano))
	builder.WriteByte('|')
	builder.WriteString(staleAfter.UTC().Format(time.RFC3339Nano))
	for _, entity := range entities {
		builder.WriteByte('|')
		builder.WriteString(entity.entity.EntityExternalKey)
		builder.WriteByte(':')
		builder.WriteString(strconv.FormatInt(entity.entity.ProjectionVersion, 10))
		builder.WriteByte(':')
		builder.WriteString(string(entity.entity.SyncStatus))
	}
	for _, relation := range relations {
		builder.WriteByte('|')
		builder.WriteString(string(relation.RelationKind))
		builder.WriteByte(':')
		builder.WriteString(missionControlEntityRefKey(relation.SourceEntityRef))
		builder.WriteByte('>')
		builder.WriteString(missionControlEntityRefKey(relation.TargetEntityRef))
	}
	sum := sha1.Sum([]byte(builder.String()))
	return fmt.Sprintf("%x", sum)
}

func missionControlEntityCardToProto(entity missioncontroldomain.Entity, relationCount int32) *controlplanev1.MissionControlEntityCard {
	card := &controlplanev1.MissionControlEntityCard{
		EntityKind:        strings.TrimSpace(string(entity.EntityKind)),
		EntityPublicId:    strings.TrimSpace(entity.EntityExternalKey),
		Title:             strings.TrimSpace(entity.Title),
		State:             strings.TrimSpace(string(entity.ActiveState)),
		SyncStatus:        strings.TrimSpace(string(entity.SyncStatus)),
		RelationCount:     relationCount,
		Badges:            missionControlEntityBadges(entity),
		ProjectionVersion: entity.ProjectionVersion,
	}
	if providerRef := missionControlProviderReference(entity); providerRef != nil {
		card.ProviderReference = providerRef
	}
	if actor := missionControlPrimaryActor(entity); actor != nil {
		card.PrimaryActor = actor
	}
	if entity.LastTimelineAt != nil {
		card.LastTimelineAt = timestamppb.New(entity.LastTimelineAt.UTC())
	}
	return card
}

func missionControlProviderReference(entity missioncontroldomain.Entity) *controlplanev1.MissionControlProviderReference {
	if strings.TrimSpace(entity.ProviderURL) == "" && strings.TrimSpace(entity.EntityExternalKey) == "" {
		return nil
	}
	return &controlplanev1.MissionControlProviderReference{
		Provider:   strings.TrimSpace(string(entity.ProviderKind)),
		ExternalId: strings.TrimSpace(entity.EntityExternalKey),
		Url:        strings.TrimSpace(entity.ProviderURL),
	}
}

func missionControlPrimaryActor(entity missioncontroldomain.Entity) *controlplanev1.MissionControlPrimaryActor {
	if entity.EntityKind != enumtypes.MissionControlEntityKindAgent {
		return nil
	}
	return &controlplanev1.MissionControlPrimaryActor{
		ActorType:   "agent",
		ActorId:     strings.TrimPrefix(strings.TrimSpace(entity.EntityExternalKey), "agent/"),
		DisplayName: strings.TrimSpace(entity.Title),
	}
}

func missionControlEntityBadges(entity missioncontroldomain.Entity) []string {
	out := make([]string, 0, 2)
	if entity.ActiveState == enumtypes.MissionControlActiveStateBlocked {
		out = append(out, "blocked")
	}
	if entity.SyncStatus == enumtypes.MissionControlSyncStatusDegraded || entity.SyncStatus == enumtypes.MissionControlSyncStatusFailed {
		out = append(out, "realtime_stale")
	} else if entity.StaleAfter != nil && time.Now().UTC().After(entity.StaleAfter.UTC()) {
		out = append(out, "realtime_stale")
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func missionControlRelationToProto(relation valuetypes.MissionControlRelationView) *controlplanev1.MissionControlRelation {
	return &controlplanev1.MissionControlRelation{
		RelationKind:         strings.TrimSpace(string(relation.RelationKind)),
		SourceKind:           strings.TrimSpace(string(relation.SourceKind)),
		SourceEntityKind:     strings.TrimSpace(string(relation.SourceEntityRef.EntityKind)),
		SourceEntityPublicId: strings.TrimSpace(relation.SourceEntityRef.EntityPublicID),
		TargetEntityKind:     strings.TrimSpace(string(relation.TargetEntityRef.EntityKind)),
		TargetEntityPublicId: strings.TrimSpace(relation.TargetEntityRef.EntityPublicID),
		Direction:            missionControlRelationDirection(relation.RelationKind),
	}
}

func missionControlRelationDirection(kind enumtypes.MissionControlRelationKind) string {
	switch kind {
	case enumtypes.MissionControlRelationKindLinkedTo:
		return "bidirectional"
	case enumtypes.MissionControlRelationKindBlockedBy:
		return "inbound"
	default:
		return "outbound"
	}
}

func missionControlEntityDetailsToProto(details missioncontroldomain.EntityDetails) *controlplanev1.MissionControlEntityDetails {
	relationCounts := missionControlRelationCounts(details.Relations)
	out := &controlplanev1.MissionControlEntityDetails{
		Entity:            missionControlEntityCardToProto(details.Entity, relationCounts[missionControlEntityKey(details.Entity)]),
		Relations:         make([]*controlplanev1.MissionControlRelation, 0, len(details.Relations)),
		TimelinePreview:   make([]*controlplanev1.MissionControlTimelineEntry, 0, len(details.Timeline)),
		AllowedActions:    nil,
		ProviderDeepLinks: missionControlProviderDeepLinks(details.Entity),
	}
	for _, relation := range details.Relations {
		out.Relations = append(out.Relations, missionControlRelationToProto(relation))
	}
	for _, item := range details.Timeline {
		out.TimelinePreview = append(out.TimelinePreview, missionControlTimelineEntryToProto(details.Entity.EntityKind, details.Entity.EntityExternalKey, item))
	}
	missionControlApplyDetailPayload(out, details.Entity)
	return out
}

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

func missionControlApplyDetailPayload(out *controlplanev1.MissionControlEntityDetails, entity missioncontroldomain.Entity) {
	switch entity.EntityKind {
	case enumtypes.MissionControlEntityKindWorkItem:
		payload := &controlplanev1.MissionControlWorkItemDetailsPayload{}
		var decoded missionControlDetailPayloadWorkItem
		if err := json.Unmarshal(entity.DetailPayloadJSON, &decoded); err == nil {
			payload.RepositoryFullName = strings.TrimSpace(decoded.RepositoryFullName)
			payload.IssueNumber = decoded.IssueNumber
			payload.IssueUrl = stringPtrOrNil(strings.TrimSpace(decoded.IssueURL))
			payload.LastRunId = stringPtrOrNil(strings.TrimSpace(decoded.LastRunID))
			payload.LastStatus = stringPtrOrNil(strings.TrimSpace(decoded.LastStatus))
			payload.TriggerKind = stringPtrOrNil(strings.TrimSpace(decoded.TriggerKind))
		}
		if entity.ProviderUpdatedAt != nil {
			payload.LastProviderSyncAt = timestamppb.New(entity.ProviderUpdatedAt.UTC())
		}
		out.DetailPayload = &controlplanev1.MissionControlEntityDetails_WorkItem{WorkItem: payload}
	case enumtypes.MissionControlEntityKindPullRequest:
		payload := &controlplanev1.MissionControlPullRequestDetailsPayload{}
		var decoded missionControlDetailPayloadPullRequest
		if err := json.Unmarshal(entity.DetailPayloadJSON, &decoded); err == nil {
			payload.RepositoryFullName = strings.TrimSpace(decoded.RepositoryFullName)
			payload.PullRequestNumber = decoded.PullRequestNumber
			payload.PullRequestUrl = stringPtrOrNil(strings.TrimSpace(decoded.PullRequestURL))
			payload.LastRunId = stringPtrOrNil(strings.TrimSpace(decoded.LastRunID))
			payload.LastStatus = stringPtrOrNil(strings.TrimSpace(decoded.LastStatus))
		}
		out.DetailPayload = &controlplanev1.MissionControlEntityDetails_PullRequest{PullRequest: payload}
	case enumtypes.MissionControlEntityKindAgent:
		payload := &controlplanev1.MissionControlAgentDetailsPayload{}
		var decoded missionControlDetailPayloadAgent
		if err := json.Unmarshal(entity.DetailPayloadJSON, &decoded); err == nil {
			payload.AgentKey = strings.TrimSpace(decoded.AgentKey)
			payload.RunStatus = stringPtrOrNil(strings.TrimSpace(decoded.LastStatus))
			payload.ActiveRunId = stringPtrOrNil(strings.TrimSpace(decoded.LastRunID))
			payload.LastRunRepository = stringPtrOrNil(strings.TrimSpace(decoded.LastRunRepo))
		}
		out.DetailPayload = &controlplanev1.MissionControlEntityDetails_Agent{Agent: payload}
	case enumtypes.MissionControlEntityKindDiscussion:
		out.DetailPayload = &controlplanev1.MissionControlEntityDetails_Discussion{
			Discussion: &controlplanev1.MissionControlDiscussionDetailsPayload{},
		}
		out.GetDiscussion().DiscussionKind = stringPtrOrNil("discussion")
	}
}

func missionControlTimelineEntryToProto(
	entityKind enumtypes.MissionControlEntityKind,
	entityPublicID string,
	entry missioncontroldomain.TimelineEntry,
) *controlplanev1.MissionControlTimelineEntry {
	out := &controlplanev1.MissionControlTimelineEntry{
		EntryId:        strings.TrimSpace(entry.EntryExternalKey),
		EntityKind:     strings.TrimSpace(string(entityKind)),
		EntityPublicId: strings.TrimSpace(entityPublicID),
		SourceKind:     strings.TrimSpace(string(entry.SourceKind)),
		SourceRef:      strings.TrimSpace(entry.EntryExternalKey),
		OccurredAt:     timestamppb.New(entry.OccurredAt.UTC()),
		Summary:        strings.TrimSpace(entry.Summary),
		IsReadOnly:     entry.IsReadOnly,
	}
	if trimmed := strings.TrimSpace(entry.BodyMarkdown); trimmed != "" {
		out.BodyMarkdown = stringPtrOrNil(trimmed)
	}
	if trimmed := strings.TrimSpace(entry.CommandID); trimmed != "" {
		out.CommandId = stringPtrOrNil(trimmed)
		out.SourceRef = trimmed
	}
	if trimmed := strings.TrimSpace(entry.ProviderURL); trimmed != "" {
		out.ProviderUrl = stringPtrOrNil(trimmed)
	}
	return out
}

func missionControlCommandPayloadFromProto(req *controlplanev1.SubmitMissionControlCommandRequest) (valuetypes.MissionControlCommandPayload, error) {
	switch payload := req.GetPayload().(type) {
	case *controlplanev1.SubmitMissionControlCommandRequest_DiscussionCreate:
		return valuetypes.MissionControlCommandPayload{
			DiscussionCreate: &valuetypes.MissionControlDiscussionCreatePayload{
				Title:        strings.TrimSpace(payload.DiscussionCreate.GetTitle()),
				BodyMarkdown: strings.TrimSpace(payload.DiscussionCreate.GetBodyMarkdown()),
				ParentEntityRef: missionControlOptionalEntityRef(
					payload.DiscussionCreate.GetParentEntityKind(),
					payload.DiscussionCreate.GetParentEntityPublicId(),
				),
			},
		}, nil
	case *controlplanev1.SubmitMissionControlCommandRequest_WorkItemCreate:
		return valuetypes.MissionControlCommandPayload{
			WorkItemCreate: &valuetypes.MissionControlWorkItemCreatePayload{
				Title:             strings.TrimSpace(payload.WorkItemCreate.GetTitle()),
				BodyMarkdown:      strings.TrimSpace(payload.WorkItemCreate.GetBodyMarkdown()),
				InitialLabels:     trimStringSlice(payload.WorkItemCreate.GetInitialLabels()),
				RelatedEntityRefs: missionControlEntityRefsFromProto(payload.WorkItemCreate.GetRelatedEntityRefs()),
			},
		}, nil
	case *controlplanev1.SubmitMissionControlCommandRequest_DiscussionFormalize:
		return valuetypes.MissionControlCommandPayload{
			DiscussionFormalize: &valuetypes.MissionControlDiscussionFormalizePayload{
				SourceEntityRef: valuetypes.MissionControlEntityRef{
					EntityKind:     enumtypes.MissionControlEntityKind(strings.TrimSpace(payload.DiscussionFormalize.GetSourceEntityKind())),
					EntityPublicID: strings.TrimSpace(payload.DiscussionFormalize.GetSourceEntityPublicId()),
				},
				FormalizedKind: strings.TrimSpace(payload.DiscussionFormalize.GetFormalizedKind()),
				Title:          strings.TrimSpace(payload.DiscussionFormalize.GetTitle()),
				BodyMarkdown:   strings.TrimSpace(payload.DiscussionFormalize.GetBodyMarkdown()),
			},
		}, nil
	case *controlplanev1.SubmitMissionControlCommandRequest_StageNextStep:
		return valuetypes.MissionControlCommandPayload{
			StageNextStep: &valuetypes.MissionControlStageNextStepExecutePayload{
				ThreadKind:          strings.TrimSpace(payload.StageNextStep.GetThreadKind()),
				ThreadNumber:        int(payload.StageNextStep.GetThreadNumber()),
				TargetLabel:         strings.TrimSpace(payload.StageNextStep.GetTargetLabel()),
				RemovedLabels:       trimStringSlice(payload.StageNextStep.GetRemovedLabels()),
				DisplayVariant:      strings.TrimSpace(payload.StageNextStep.GetDisplayVariant()),
				ApprovalRequirement: enumtypes.MissionControlApprovalRequirement(strings.TrimSpace(payload.StageNextStep.GetApprovalRequirement())),
			},
		}, nil
	case *controlplanev1.SubmitMissionControlCommandRequest_RetrySync:
		return valuetypes.MissionControlCommandPayload{
			RetrySync: &valuetypes.MissionControlRetrySyncPayload{
				CommandID:      strings.TrimSpace(payload.RetrySync.GetCommandId()),
				RetryReason:    strings.TrimSpace(payload.RetrySync.GetRetryReason()),
				ExpectedStatus: enumtypes.MissionControlCommandStatus(strings.TrimSpace(payload.RetrySync.GetExpectedStatus())),
			},
		}, nil
	default:
		return valuetypes.MissionControlCommandPayload{}, errs.Validation{Field: "payload", Msg: "is required"}
	}
}

func missionControlTargetRefFromProto(req *controlplanev1.SubmitMissionControlCommandRequest) *valuetypes.MissionControlEntityRef {
	return missionControlOptionalEntityRef(req.GetTargetEntityKind(), req.GetTargetEntityPublicId())
}

func missionControlOptionalEntityRef(entityKind string, entityPublicID string) *valuetypes.MissionControlEntityRef {
	entityKind = strings.TrimSpace(entityKind)
	entityPublicID = strings.TrimSpace(entityPublicID)
	if entityKind == "" && entityPublicID == "" {
		return nil
	}
	return &valuetypes.MissionControlEntityRef{
		EntityKind:     enumtypes.MissionControlEntityKind(entityKind),
		EntityPublicID: entityPublicID,
	}
}

func missionControlEntityRefsFromProto(items []*controlplanev1.MissionControlEntityRef) []valuetypes.MissionControlEntityRef {
	if len(items) == 0 {
		return nil
	}
	out := make([]valuetypes.MissionControlEntityRef, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		entityRef := missionControlOptionalEntityRef(item.GetEntityKind(), item.GetEntityPublicId())
		if entityRef == nil {
			continue
		}
		out = append(out, *entityRef)
	}
	return out
}

func missionControlCommandStatusViewToProto(view missioncontroldomain.CommandStatusView) *controlplanev1.MissionControlCommandState {
	command := view.Command
	resp := &controlplanev1.MissionControlCommandState{
		ProjectId:           strings.TrimSpace(command.ProjectID),
		CommandId:           strings.TrimSpace(command.ID),
		CommandKind:         strings.TrimSpace(string(command.CommandKind)),
		Status:              strings.TrimSpace(string(command.Status)),
		FailureReason:       strings.TrimSpace(string(command.FailureReason)),
		CorrelationId:       strings.TrimSpace(command.CorrelationID),
		ProviderDeliveryIds: trimStringSlice(view.ProviderDeliveryIDs),
		UpdatedAt:           timestamppb.New(command.UpdatedAt.UTC()),
		BusinessIntentKey:   strings.TrimSpace(command.BusinessIntentKey),
		EntityRefs:          missionControlEntityRefsToProto(view.EntityRefs),
	}
	if strings.TrimSpace(view.StatusMessage) != "" {
		resp.StatusMessage = stringPtrOrNil(strings.TrimSpace(view.StatusMessage))
	}
	if command.ReconciledAt != nil {
		resp.ReconciledAt = timestamppb.New(command.ReconciledAt.UTC())
	}
	if approval := missionControlApprovalToProto(view.Approval); approval != nil {
		resp.Approval = approval
	}
	if command.Status == enumtypes.MissionControlCommandStatusBlocked || command.Status == enumtypes.MissionControlCommandStatusFailed {
		if reason := strings.TrimSpace(string(command.FailureReason)); reason != "" {
			resp.BlockingReason = stringPtrOrNil(reason)
		}
	}
	return resp
}

func missionControlApprovalToProto(approval *valuetypes.MissionControlApprovalSnapshot) *controlplanev1.MissionControlCommandApproval {
	if approval == nil {
		return nil
	}
	out := &controlplanev1.MissionControlCommandApproval{
		ApprovalState: strings.TrimSpace(string(approval.ApprovalState)),
	}
	if trimmed := strings.TrimSpace(approval.ApprovalRequestID); trimmed != "" {
		out.ApprovalRequestId = stringPtrOrNil(trimmed)
	}
	if approval.RequestedAt != nil {
		out.RequestedAt = timestamppb.New(approval.RequestedAt.UTC())
	}
	if approval.DecidedAt != nil {
		out.DecidedAt = timestamppb.New(approval.DecidedAt.UTC())
	}
	if trimmed := strings.TrimSpace(approval.ApproverActorID); trimmed != "" {
		out.ApproverActorId = stringPtrOrNil(trimmed)
	}
	return out
}

func missionControlEntityRefsToProto(items []valuetypes.MissionControlEntityRef) []*controlplanev1.MissionControlEntityRef {
	if len(items) == 0 {
		return nil
	}
	out := make([]*controlplanev1.MissionControlEntityRef, 0, len(items))
	for _, item := range items {
		out = append(out, &controlplanev1.MissionControlEntityRef{
			EntityKind:     strings.TrimSpace(string(item.EntityKind)),
			EntityPublicId: strings.TrimSpace(item.EntityPublicID),
		})
	}
	return out
}

func decodeMissionControlCursor(cursor string) (int, error) {
	if strings.TrimSpace(cursor) == "" {
		return 0, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(cursor))
	if err != nil {
		return 0, err
	}
	offset, err := strconv.Atoi(string(decoded))
	if err != nil || offset < 0 {
		return 0, fmt.Errorf("invalid cursor")
	}
	return offset, nil
}

func encodeMissionControlCursor(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func missionControlEntityMatchesSearch(entity missioncontroldomain.Entity, search string) bool {
	if search == "" {
		return true
	}
	search = strings.ToLower(search)
	candidates := []string{
		entity.Title,
		entity.EntityExternalKey,
		entity.ProviderURL,
		string(entity.CardPayloadJSON),
		string(entity.DetailPayloadJSON),
	}
	for _, candidate := range candidates {
		if strings.Contains(strings.ToLower(candidate), search) {
			return true
		}
	}
	return false
}

func missionControlEntityKey(entity missioncontroldomain.Entity) string {
	return string(entity.EntityKind) + "/" + entity.EntityExternalKey
}

func missionControlEntityRefKey(entityRef valuetypes.MissionControlEntityRef) string {
	return string(entityRef.EntityKind) + "/" + entityRef.EntityPublicID
}
