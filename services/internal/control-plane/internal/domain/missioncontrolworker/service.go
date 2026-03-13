package missioncontrolworker

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"sort"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/missioncontrol"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	missioncontrolrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/missioncontrol"
	projectrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/project"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	staffrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/staffrun"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

const (
	repositoryRoleOrchestrator = "orchestrator"
	repositoryRoleMixed        = "mixed"
	defaultWarmupRequestedBy   = "mission-control-worker"
	defaultTimelineText        = "Platform event"
)

// Service exposes worker-owned Mission Control warmup and execution lookup helpers.
type Service struct {
	cfg            Config
	projects       projectrepo.Repository
	repositories   repocfgrepo.Repository
	agentRuns      agentrunrepo.Repository
	staffRuns      staffrunrepo.Repository
	missionControl missioncontrol.DomainService
	projection     missioncontrolrepo.Repository
	now            func() time.Time
}

// NewService constructs Mission Control worker helpers.
func NewService(cfg Config, deps Dependencies) (*Service, error) {
	if deps.Projects == nil {
		return nil, fmt.Errorf("mission control worker projects repository is required")
	}
	if deps.Repositories == nil {
		return nil, fmt.Errorf("mission control worker repository bindings repository is required")
	}
	if deps.AgentRuns == nil {
		return nil, fmt.Errorf("mission control worker agent runs repository is required")
	}
	if deps.StaffRuns == nil {
		return nil, fmt.Errorf("mission control worker staff runs repository is required")
	}
	if deps.MissionControl == nil {
		return nil, fmt.Errorf("mission control worker domain service is required")
	}
	if deps.Projection == nil {
		return nil, fmt.Errorf("mission control worker projection repository is required")
	}
	if cfg.ProjectLimit <= 0 {
		cfg.ProjectLimit = 20
	}
	if cfg.RunLimit <= 0 {
		cfg.RunLimit = 200
	}
	if cfg.TimelineEventLimit <= 0 {
		cfg.TimelineEventLimit = 100
	}
	if cfg.StaleAfter <= 0 {
		cfg.StaleAfter = 24 * time.Hour
	}
	cfg.DefaultTimelineText = strings.TrimSpace(cfg.DefaultTimelineText)
	if cfg.DefaultTimelineText == "" {
		cfg.DefaultTimelineText = defaultTimelineText
	}

	return &Service{
		cfg:            cfg,
		projects:       deps.Projects,
		repositories:   deps.Repositories,
		agentRuns:      deps.AgentRuns,
		staffRuns:      deps.StaffRuns,
		missionControl: deps.MissionControl,
		projection:     deps.Projection,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}, nil
}

// ListWarmupProjects returns projects that have repository bindings and historical runs to backfill.
func (s *Service) ListWarmupProjects(ctx context.Context, limit int) ([]WarmupProject, error) {
	projectLimit := limit
	if projectLimit <= 0 {
		projectLimit = s.cfg.ProjectLimit
	}
	projects, err := s.projects.ListAll(ctx, projectLimit)
	if err != nil {
		return nil, err
	}

	items := make([]WarmupProject, 0, len(projects))
	for _, project := range projects {
		repositoryFullName, ok, err := s.resolveProjectRepositoryFullName(ctx, project.ID)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		recentRuns, err := s.agentRuns.ListRecentByProject(ctx, project.ID, repositoryFullName, 1, 0)
		if err != nil {
			return nil, err
		}
		if len(recentRuns) == 0 {
			continue
		}
		items = append(items, WarmupProject{
			ProjectID:          project.ID,
			ProjectName:        strings.TrimSpace(project.Name),
			RepositoryFullName: repositoryFullName,
		})
	}
	return items, nil
}

// RunWarmup rebuilds the current coarse active-set projection from persisted runs and flow events.
func (s *Service) RunWarmup(ctx context.Context, params WarmupRequest) (WarmupResult, error) {
	projectID := strings.TrimSpace(params.ProjectID)
	if projectID == "" {
		return WarmupResult{}, errs.Validation{Field: "project_id", Msg: "is required"}
	}
	project, found, err := s.projects.GetByID(ctx, projectID)
	if err != nil {
		return WarmupResult{}, err
	}
	if !found {
		return WarmupResult{}, errs.NotFound{Msg: "project not found"}
	}
	repositoryFullName, ok, err := s.resolveProjectRepositoryFullName(ctx, projectID)
	if err != nil {
		return WarmupResult{}, err
	}
	if !ok {
		return WarmupResult{}, errs.FailedPrecondition{Msg: "mission control warmup requires at least one repository binding"}
	}

	runs, err := s.agentRuns.ListRecentByProject(ctx, projectID, repositoryFullName, s.cfg.RunLimit, 0)
	if err != nil {
		return WarmupResult{}, err
	}

	entitySeeds := make(map[string]projectionSeed)
	relationSeeds := make(map[string]relationSeed)
	timelineSeeds := make(map[string]timelineSeed)
	for _, run := range runs {
		if err := s.collectProjectionSeeds(ctx, project, repositoryFullName, run, entitySeeds, relationSeeds, timelineSeeds); err != nil {
			return WarmupResult{}, err
		}
	}

	entityIDs := make(map[string]int64, len(entitySeeds))
	entityKeys := sortedProjectionKeys(entitySeeds)
	for _, entityKey := range entityKeys {
		seed := entitySeeds[entityKey]
		entity, upsertErr := s.missionControl.UpsertEntity(ctx, missioncontrol.UpsertEntityParams{
			ProjectID:         seed.projectID,
			EntityKind:        seed.entityKind,
			EntityExternalKey: seed.entityExternalKey,
			ProviderKind:      seed.providerKind,
			ProviderURL:       seed.providerURL,
			Title:             seed.title,
			ActiveState:       seed.activeState,
			SyncStatus:        seed.syncStatus,
			ProjectionVersion: seed.projectionVersion,
			CardPayloadJSON:   seed.cardPayloadJSON,
			DetailPayloadJSON: seed.detailPayloadJSON,
			ProviderUpdatedAt: seed.providerUpdatedAt,
			ProjectedAt:       seed.projectedAt,
			StaleAfter:        seed.staleAfter,
		}, params.CorrelationID)
		if upsertErr != nil {
			return WarmupResult{}, upsertErr
		}
		entityIDs[entityKey] = entity.ID
	}

	relationsBySource := make(map[string][]missioncontrolrepo.RelationSeed)
	for _, relation := range relationSeeds {
		sourceID, sourceOK := entityIDs[relation.sourceEntityKey]
		targetID, targetOK := entityIDs[relation.targetEntityKey]
		if !sourceOK || !targetOK {
			continue
		}
		sourceKey := fmt.Sprintf("%d", sourceID)
		relationsBySource[sourceKey] = append(relationsBySource[sourceKey], missioncontrolrepo.RelationSeed{
			TargetEntityID: targetID,
			RelationKind:   relation.relationKind,
			SourceKind:     enumtypes.MissionControlRelationSourceKindPlatform,
		})
	}
	relationCount := 0
	for sourceKey, relations := range relationsBySource {
		sourceID := parseEntityID(sourceKey)
		if err := s.missionControl.ReplaceRelationsForSource(ctx, missioncontrol.ReplaceRelationsParams{
			ProjectID:      projectID,
			SourceEntityID: sourceID,
			Relations:      relations,
		}, params.CorrelationID); err != nil {
			return WarmupResult{}, err
		}
		relationCount += len(relations)
	}

	timelineKeys := sortedTimelineKeys(timelineSeeds)
	timelineCount := 0
	for _, timelineKey := range timelineKeys {
		seed := timelineSeeds[timelineKey]
		entityID, ok := entityIDs[seed.entityKey]
		if !ok {
			continue
		}
		if _, err := s.missionControl.UpsertTimelineEntry(ctx, missioncontrol.UpsertTimelineEntryParams{
			ProjectID:        projectID,
			EntityID:         entityID,
			SourceKind:       enumtypes.MissionControlTimelineSourceKindPlatform,
			EntryExternalKey: seed.entryExternalKey,
			Summary:          seed.summary,
			PayloadJSON:      seed.payloadJSON,
			OccurredAt:       seed.occurredAt,
			ProviderURL:      seed.providerURL,
			IsReadOnly:       true,
		}, params.CorrelationID); err != nil {
			return WarmupResult{}, err
		}
		timelineCount++
	}

	requestedBy := strings.TrimSpace(params.RequestedBy)
	if requestedBy == "" {
		requestedBy = defaultWarmupRequestedBy
	}
	summary, err := s.missionControl.RunWarmup(ctx, missioncontrol.WarmupRequest{
		ProjectID:     projectID,
		RequestedBy:   requestedBy,
		CorrelationID: params.CorrelationID,
		ForceRebuild:  params.ForceRebuild,
	})
	if err != nil {
		return WarmupResult{}, err
	}

	return WarmupResult{
		Summary:             summary,
		BackfilledEntities:  len(entityIDs),
		BackfilledRelations: relationCount,
		BackfilledTimelines: timelineCount,
	}, nil
}

// ListPendingCommands returns accepted/queued commands that worker may execute.
func (s *Service) ListPendingCommands(ctx context.Context, limit int) ([]PendingCommand, error) {
	commandLimit := limit
	if commandLimit <= 0 {
		commandLimit = s.cfg.RunLimit
	}
	commands, err := s.projection.ListCommandsAll(ctx, missioncontrolrepo.GlobalCommandListFilter{
		Statuses: []enumtypes.MissionControlCommandStatus{
			enumtypes.MissionControlCommandStatusAccepted,
			enumtypes.MissionControlCommandStatusQueued,
		},
		Limit: commandLimit,
	})
	if err != nil {
		return nil, err
	}

	repositoryCache := make(map[string]string, len(commands))
	items := make([]PendingCommand, 0, len(commands))
	for _, command := range commands {
		item, ok, err := s.resolvePendingCommand(ctx, command, repositoryCache)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) resolvePendingCommand(
	ctx context.Context,
	command missioncontrolrepo.Command,
	repositoryCache map[string]string,
) (PendingCommand, bool, error) {
	item := PendingCommand{
		ProjectID:            command.ProjectID,
		CommandID:            command.ID,
		CommandKind:          command.CommandKind,
		EffectiveCommandKind: command.CommandKind,
		Status:               command.Status,
		CorrelationID:        command.CorrelationID,
		BusinessIntentKey:    command.BusinessIntentKey,
		RequestedAt:          command.RequestedAt.UTC(),
		UpdatedAt:            command.UpdatedAt.UTC(),
	}
	if repositoryFullName, ok := repositoryCache[command.ProjectID]; ok {
		item.RepositoryFullName = repositoryFullName
	} else {
		repositoryFullName, found, err := s.resolveProjectRepositoryFullName(ctx, command.ProjectID)
		if err != nil {
			return PendingCommand{}, false, err
		}
		if found {
			item.RepositoryFullName = repositoryFullName
			repositoryCache[command.ProjectID] = repositoryFullName
		}
	}

	payload, err := decodePendingCommandPayload(command.PayloadJSON)
	if err != nil {
		return item, true, nil
	}
	switch command.CommandKind {
	case enumtypes.MissionControlCommandKindStageNextStep:
		if payload.StageNextStep != nil {
			item.StageNextStep = &PendingStageNextStep{
				ThreadKind:  strings.TrimSpace(payload.StageNextStep.ThreadKind),
				ThreadNo:    payload.StageNextStep.ThreadNumber,
				TargetLabel: strings.TrimSpace(payload.StageNextStep.TargetLabel),
			}
		}
	case enumtypes.MissionControlCommandKindRetrySync:
		if payload.RetrySync == nil {
			return item, true, nil
		}
		target, found, err := s.projection.GetCommandByID(ctx, command.ProjectID, payload.RetrySync.CommandID)
		if err != nil {
			return PendingCommand{}, false, err
		}
		if !found {
			return item, true, nil
		}
		item.RetryTargetCommandID = target.ID
		item.EffectiveCommandKind = target.CommandKind
		targetPayload, decodeErr := decodePendingCommandPayload(target.PayloadJSON)
		if decodeErr != nil {
			return item, true, nil
		}
		if target.CommandKind == enumtypes.MissionControlCommandKindStageNextStep && targetPayload.StageNextStep != nil {
			item.StageNextStep = &PendingStageNextStep{
				ThreadKind:  strings.TrimSpace(targetPayload.StageNextStep.ThreadKind),
				ThreadNo:    targetPayload.StageNextStep.ThreadNumber,
				TargetLabel: strings.TrimSpace(targetPayload.StageNextStep.TargetLabel),
			}
		}
	}
	return item, true, nil
}

func (s *Service) collectProjectionSeeds(
	ctx context.Context,
	project projectrepo.Project,
	repositoryFullName string,
	run agentrunrepo.RunLookupItem,
	entitySeeds map[string]projectionSeed,
	relationSeeds map[string]relationSeed,
	timelineSeeds map[string]timelineSeed,
) error {
	projectedAt := safeProjectionTime(run.CreatedAt, s.now)
	staleAfter := projectedAt.Add(s.cfg.StaleAfter)

	agentKey := strings.TrimSpace(run.AgentKey)
	agentEntityKey := ""
	if agentKey != "" {
		agentEntityKey = agentEntityProjectionKey(agentKey)
		seedProjection(entitySeeds, agentEntityKey, projectionSeed{
			projectID:         project.ID,
			entityKind:        enumtypes.MissionControlEntityKindAgent,
			entityExternalKey: agentEntityKey,
			providerKind:      enumtypes.MissionControlProviderKindPlatform,
			title:             agentDisplayTitle(agentKey),
			activeState:       enumtypes.MissionControlActiveStateWorking,
			syncStatus:        enumtypes.MissionControlSyncStatusSynced,
			projectionVersion: projectedAt.UnixMilli(),
			cardPayloadJSON:   mustMarshal(agentCardPayload{AgentKey: agentKey, LastRunID: run.RunID, LastStatus: run.Status, LastRunRepo: repositoryFullName}),
			detailPayloadJSON: mustMarshal(agentCardPayload{AgentKey: agentKey, LastRunID: run.RunID, LastStatus: run.Status, LastRunRepo: repositoryFullName}),
			projectedAt:       projectedAt,
			staleAfter:        &staleAfter,
		})
	}

	workItemEntityKey := ""
	if run.IssueNumber > 0 {
		workItemEntityKey = workItemProjectionKey(repositoryFullName, run.IssueNumber)
		issueURL := strings.TrimSpace(run.IssueURL)
		if issueURL == "" {
			issueURL = githubIssueURL(repositoryFullName, run.IssueNumber)
		}
		seedProjection(entitySeeds, workItemEntityKey, projectionSeed{
			projectID:         project.ID,
			entityKind:        enumtypes.MissionControlEntityKindWorkItem,
			entityExternalKey: workItemEntityKey,
			providerKind:      enumtypes.MissionControlProviderKindGitHub,
			providerURL:       issueURL,
			title:             fmt.Sprintf("Issue #%d", run.IssueNumber),
			activeState:       warmupActiveState(run.Status, false, run.PullRequestNumber > 0),
			syncStatus:        enumtypes.MissionControlSyncStatusSynced,
			projectionVersion: projectedAt.UnixMilli(),
			cardPayloadJSON: mustMarshal(workItemCardPayload{
				RepositoryFullName: repositoryFullName,
				IssueNumber:        run.IssueNumber,
				IssueURL:           issueURL,
				LastRunID:          run.RunID,
				LastStatus:         run.Status,
				TriggerKind:        run.TriggerKind,
			}),
			detailPayloadJSON: mustMarshal(workItemCardPayload{
				RepositoryFullName: repositoryFullName,
				IssueNumber:        run.IssueNumber,
				IssueURL:           issueURL,
				LastRunID:          run.RunID,
				LastStatus:         run.Status,
				TriggerKind:        run.TriggerKind,
			}),
			providerUpdatedAt: timePointer(projectedAt),
			projectedAt:       projectedAt,
			staleAfter:        &staleAfter,
		})
		if agentEntityKey != "" {
			relationSeeds[relationCompositeKey(workItemEntityKey, agentEntityKey, enumtypes.MissionControlRelationKindAssignedTo)] = relationSeed{
				sourceEntityKey: workItemEntityKey,
				targetEntityKey: agentEntityKey,
				relationKind:    enumtypes.MissionControlRelationKindAssignedTo,
			}
		}
	}

	pullRequestEntityKey := ""
	if run.PullRequestNumber > 0 {
		pullRequestEntityKey = pullRequestProjectionKey(repositoryFullName, run.PullRequestNumber)
		pullRequestURL := strings.TrimSpace(run.PullRequestURL)
		if pullRequestURL == "" {
			pullRequestURL = githubPullRequestURL(repositoryFullName, run.PullRequestNumber)
		}
		seedProjection(entitySeeds, pullRequestEntityKey, projectionSeed{
			projectID:         project.ID,
			entityKind:        enumtypes.MissionControlEntityKindPullRequest,
			entityExternalKey: pullRequestEntityKey,
			providerKind:      enumtypes.MissionControlProviderKindGitHub,
			providerURL:       pullRequestURL,
			title:             fmt.Sprintf("PR #%d", run.PullRequestNumber),
			activeState:       warmupActiveState(run.Status, true, false),
			syncStatus:        enumtypes.MissionControlSyncStatusSynced,
			projectionVersion: projectedAt.UnixMilli(),
			cardPayloadJSON: mustMarshal(pullRequestCardPayload{
				RepositoryFullName: repositoryFullName,
				PullRequestNumber:  run.PullRequestNumber,
				PullRequestURL:     pullRequestURL,
				LastRunID:          run.RunID,
				LastStatus:         run.Status,
			}),
			detailPayloadJSON: mustMarshal(pullRequestCardPayload{
				RepositoryFullName: repositoryFullName,
				PullRequestNumber:  run.PullRequestNumber,
				PullRequestURL:     pullRequestURL,
				LastRunID:          run.RunID,
				LastStatus:         run.Status,
			}),
			providerUpdatedAt: timePointer(projectedAt),
			projectedAt:       projectedAt,
			staleAfter:        &staleAfter,
		})
		if agentEntityKey != "" {
			relationSeeds[relationCompositeKey(pullRequestEntityKey, agentEntityKey, enumtypes.MissionControlRelationKindAssignedTo)] = relationSeed{
				sourceEntityKey: pullRequestEntityKey,
				targetEntityKey: agentEntityKey,
				relationKind:    enumtypes.MissionControlRelationKindAssignedTo,
			}
		}
		if workItemEntityKey != "" {
			relationSeeds[relationCompositeKey(workItemEntityKey, pullRequestEntityKey, enumtypes.MissionControlRelationKindLinkedTo)] = relationSeed{
				sourceEntityKey: workItemEntityKey,
				targetEntityKey: pullRequestEntityKey,
				relationKind:    enumtypes.MissionControlRelationKindLinkedTo,
			}
		}
	}

	eventEntityKeys := relatedProjectionKeys(workItemEntityKey, pullRequestEntityKey, agentEntityKey)
	if len(eventEntityKeys) == 0 || strings.TrimSpace(run.CorrelationID) == "" {
		return nil
	}
	events, err := s.staffRuns.ListEventsByCorrelation(ctx, run.CorrelationID, s.cfg.TimelineEventLimit)
	if err != nil {
		return err
	}
	for _, event := range events {
		s.collectTimelineSeeds(repositoryFullName, run.RunID, event, eventEntityKeys, timelineSeeds)
	}
	return nil
}

func (s *Service) collectTimelineSeeds(
	repositoryFullName string,
	runID string,
	event staffrunrepo.FlowEvent,
	entityKeys []string,
	timelineSeeds map[string]timelineSeed,
) {
	payload := timelinePayload{
		RunID:          strings.TrimSpace(runID),
		CorrelationID:  strings.TrimSpace(event.CorrelationID),
		EventType:      strings.TrimSpace(event.EventType),
		EventPayload:   cloneJSON(event.PayloadJSON),
		RepositoryFull: repositoryFullName,
	}
	for _, entityKey := range entityKeys {
		entryExternalKey := buildTimelineEntryExternalKey(entityKey, event)
		timelineSeeds[entityKey+"::"+entryExternalKey] = timelineSeed{
			entityKey:         entityKey,
			entryExternalKey:  entryExternalKey,
			summary:           buildTimelineSummary(event.EventType, s.cfg.DefaultTimelineText),
			payloadJSON:       mustMarshal(payload),
			occurredAt:        safeProjectionTime(event.CreatedAt, s.now),
			repositoryFullRef: repositoryFullName,
		}
	}
}

func (s *Service) resolveProjectRepositoryFullName(ctx context.Context, projectID string) (string, bool, error) {
	bindings, err := s.repositories.ListForProject(ctx, projectID, 20)
	if err != nil {
		return "", false, err
	}
	binding, ok := selectRepositoryBinding(bindings)
	if !ok {
		return "", false, nil
	}
	return strings.TrimSpace(binding.Owner) + "/" + strings.TrimSpace(binding.Name), true, nil
}

func decodePendingCommandPayload(raw []byte) (valuetypes.MissionControlCommandPayload, error) {
	if len(raw) == 0 {
		return valuetypes.MissionControlCommandPayload{}, fmt.Errorf("empty command payload")
	}
	var payload valuetypes.MissionControlCommandPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return valuetypes.MissionControlCommandPayload{}, err
	}
	return payload, nil
}

func selectRepositoryBinding(bindings []repocfgrepo.RepositoryBinding) (repocfgrepo.RepositoryBinding, bool) {
	for _, role := range []string{repositoryRoleOrchestrator, repositoryRoleMixed} {
		for _, binding := range bindings {
			if strings.EqualFold(strings.TrimSpace(binding.Role), role) {
				return binding, true
			}
		}
	}
	if len(bindings) == 0 {
		return repocfgrepo.RepositoryBinding{}, false
	}
	sort.Slice(bindings, func(i, j int) bool {
		left := strings.TrimSpace(bindings[i].Alias)
		right := strings.TrimSpace(bindings[j].Alias)
		if left == right {
			return bindings[i].ID < bindings[j].ID
		}
		return left < right
	})
	return bindings[0], true
}

func seedProjection(seeds map[string]projectionSeed, key string, candidate projectionSeed) {
	current, found := seeds[key]
	if !found || candidate.projectionVersion >= current.projectionVersion {
		seeds[key] = candidate
	}
}

func sortedProjectionKeys(seeds map[string]projectionSeed) []string {
	keys := make([]string, 0, len(seeds))
	for key := range seeds {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedTimelineKeys(seeds map[string]timelineSeed) []string {
	keys := make([]string, 0, len(seeds))
	for key := range seeds {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func relationCompositeKey(source string, target string, kind enumtypes.MissionControlRelationKind) string {
	return source + "::" + string(kind) + "::" + target
}

func relatedProjectionKeys(keys ...string) []string {
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		if strings.TrimSpace(key) == "" {
			continue
		}
		out = append(out, key)
	}
	return out
}

func workItemProjectionKey(repositoryFullName string, issueNumber int64) string {
	return strings.TrimSpace(repositoryFullName) + "#" + fmt.Sprintf("%d", issueNumber)
}

func pullRequestProjectionKey(repositoryFullName string, pullRequestNumber int64) string {
	return strings.TrimSpace(repositoryFullName) + "/pull/" + fmt.Sprintf("%d", pullRequestNumber)
}

func agentEntityProjectionKey(agentKey string) string {
	return "agent/" + strings.TrimSpace(agentKey)
}

func buildTimelineEntryExternalKey(entityKey string, event staffrunrepo.FlowEvent) string {
	checksum := crc32.ChecksumIEEE(event.PayloadJSON)
	occurredAtUnixNano := event.CreatedAt.UTC().UnixNano()
	if event.CreatedAt.IsZero() {
		occurredAtUnixNano = 0
	}
	return fmt.Sprintf(
		"%s:%s:%d:%08x",
		entityKey,
		strings.TrimSpace(event.EventType),
		occurredAtUnixNano,
		checksum,
	)
}

func buildTimelineSummary(eventType string, fallback string) string {
	summary := strings.TrimSpace(eventType)
	if summary != "" {
		return summary
	}
	return fallback
}

func safeProjectionTime(value time.Time, fallback func() time.Time) time.Time {
	if value.IsZero() {
		return fallback().UTC()
	}
	return value.UTC()
}

func warmupActiveState(status string, hasPullRequest bool, linkedPullRequest bool) enumtypes.MissionControlActiveState {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "failed", "canceled":
		return enumtypes.MissionControlActiveStateRecentCriticalUpdates
	case "running", "pending":
		return enumtypes.MissionControlActiveStateWorking
	case "succeeded":
		if hasPullRequest || linkedPullRequest {
			return enumtypes.MissionControlActiveStateReview
		}
		return enumtypes.MissionControlActiveStateWorking
	default:
		if hasPullRequest || linkedPullRequest {
			return enumtypes.MissionControlActiveStateReview
		}
		return enumtypes.MissionControlActiveStateWaiting
	}
}

func agentDisplayTitle(agentKey string) string {
	return "Agent " + strings.TrimSpace(agentKey)
}

func githubIssueURL(repositoryFullName string, issueNumber int64) string {
	if strings.TrimSpace(repositoryFullName) == "" || issueNumber <= 0 {
		return ""
	}
	return "https://github.com/" + strings.TrimSpace(repositoryFullName) + "/issues/" + fmt.Sprintf("%d", issueNumber)
}

func githubPullRequestURL(repositoryFullName string, pullRequestNumber int64) string {
	if strings.TrimSpace(repositoryFullName) == "" || pullRequestNumber <= 0 {
		return ""
	}
	return "https://github.com/" + strings.TrimSpace(repositoryFullName) + "/pull/" + fmt.Sprintf("%d", pullRequestNumber)
}

func timePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	resolved := value.UTC()
	return &resolved
}

func parseEntityID(value string) int64 {
	var entityID int64
	_, _ = fmt.Sscanf(value, "%d", &entityID)
	return entityID
}

func mustMarshal(value any) json.RawMessage {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return raw
}

func cloneJSON(value []byte) json.RawMessage {
	if len(value) == 0 {
		return nil
	}
	cloned := make([]byte, len(value))
	copy(cloned, value)
	return cloned
}
