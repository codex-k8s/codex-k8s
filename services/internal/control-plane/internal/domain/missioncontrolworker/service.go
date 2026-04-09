package missioncontrolworker

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"sort"
	"strings"
	"time"

	"github.com/codex-k8s/kodex/libs/go/errs"
	"github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/missioncontrol"
	agentrunrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/agentrun"
	missioncontrolrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/missioncontrol"
	projectrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/project"
	repocfgrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/repocfg"
	staffrunrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/staffrun"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
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
	lineageState := make(map[string]string)
	for _, run := range runs {
		if err := s.collectProjectionSeeds(ctx, project, repositoryFullName, run, entitySeeds, relationSeeds, timelineSeeds, lineageState); err != nil {
			return WarmupResult{}, err
		}
	}

	entityIDs := make(map[string]int64, len(entitySeeds))
	entityKeys := sortedProjectionKeys(entitySeeds)
	for _, entityKey := range entityKeys {
		seed := entitySeeds[entityKey]
		entity, upsertErr := s.missionControl.UpsertEntity(ctx, missioncontrol.UpsertEntityParams(seed), params.CorrelationID)
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

	if _, err := s.missionControl.RefreshWorkspaceProjection(ctx, missioncontrol.WorkspaceRefreshParams{
		ProjectID:     projectID,
		CorrelationID: params.CorrelationID,
		ObservedAt:    s.now(),
	}); err != nil {
		return WarmupResult{}, err
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

// ClaimPendingCommands atomically leases accepted/queued commands for one worker instance.
func (s *Service) ClaimPendingCommands(ctx context.Context, workerID string, leaseTTL time.Duration, limit int) ([]PendingCommand, error) {
	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return nil, errs.Validation{Field: "worker_id", Msg: "is required"}
	}
	commandLimit := limit
	if commandLimit <= 0 {
		commandLimit = s.cfg.RunLimit
	}
	commands, err := s.projection.ClaimCommandsAll(ctx, missioncontrolrepo.ClaimCommandParams{
		WorkerID: workerID,
		LeaseTTL: leaseTTL,
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
	lineageState map[string]string,
) error {
	projectedAt := safeProjectionTime(run.CreatedAt, s.now)
	staleAfter := projectedAt.Add(s.cfg.StaleAfter)
	runContext, err := s.loadProjectionRunContext(ctx, run.RunID)
	if err != nil {
		return err
	}

	workItemEntityKey := ""
	workItemCoverage := enumtypes.MissionControlCoverageClassOutOfScope
	if run.IssueNumber > 0 {
		workItemEntityKey = workItemProjectionKey(repositoryFullName, run.IssueNumber)
		workItemCoverage = coverageClassForIssueState(runContext.IssueState)
	}

	pullRequestEntityKey := ""
	pullRequestCoverage := enumtypes.MissionControlCoverageClassOutOfScope
	if run.PullRequestNumber > 0 {
		pullRequestEntityKey = pullRequestProjectionKey(repositoryFullName, run.PullRequestNumber)
		pullRequestCoverage = coverageClassForPullRequestState(runContext.PullRequestState)
	}

	runEntityKey := runProjectionKey(run.RunID)
	nextRunEntityKey := trackRunLineage(lineageState, repositoryFullName, run, runEntityKey)
	runCoverage := coverageClassForRun(workItemCoverage, pullRequestCoverage)
	runContinuity := runContinuityStatus(run.Status, runCoverage, pullRequestEntityKey, nextRunEntityKey)
	workItemContinuity := workItemContinuityStatus(runContinuity)
	pullRequestContinuity := pullRequestContinuityStatus(runContinuity, pullRequestCoverage)
	lastTimelineAt := timePointer(projectedAt)

	if workItemEntityKey != "" {
		issueURL := strings.TrimSpace(run.IssueURL)
		if issueURL == "" {
			issueURL = githubIssueURL(repositoryFullName, run.IssueNumber)
		}
		workItemPayload := valuetypes.MissionControlWorkItemProjectionPayload{
			RepositoryFullName: repositoryFullName,
			IssueNumber:        run.IssueNumber,
			IssueURL:           issueURL,
			LastRunID:          run.RunID,
			LastStatus:         run.Status,
			TriggerKind:        run.TriggerKind,
			WorkItemType:       "issue",
			StageLabel:         runContext.StageLabel,
			Labels:             append([]string(nil), runContext.IssueLabels...),
			Owner:              runContext.IssueOwner,
			LastProviderSyncAt: timePointer(projectedAt),
		}
		workItemTitle := strings.TrimSpace(runContext.IssueTitle)
		if workItemTitle == "" {
			workItemTitle = fmt.Sprintf("Issue #%d", run.IssueNumber)
		}
		seedProjection(entitySeeds, workItemEntityKey, projectionSeed{
			ProjectID:         project.ID,
			EntityKind:        enumtypes.MissionControlEntityKindWorkItem,
			EntityExternalKey: workItemEntityKey,
			ProviderKind:      enumtypes.MissionControlProviderKindGitHub,
			ProviderURL:       issueURL,
			Title:             workItemTitle,
			ActiveState:       warmupActiveState(run.Status, false, run.PullRequestNumber > 0),
			SyncStatus:        enumtypes.MissionControlSyncStatusSynced,
			ContinuityStatus:  workItemContinuity,
			CoverageClass:     workItemCoverage,
			ProjectionVersion: projectedAt.UnixMilli(),
			CardPayloadJSON:   mustMarshal(workItemPayload),
			DetailPayloadJSON: mustMarshal(workItemPayload),
			LastTimelineAt:    lastTimelineAt,
			ProviderUpdatedAt: timePointer(projectedAt),
			ProjectedAt:       projectedAt,
			StaleAfter:        &staleAfter,
		})
		if runEntityKey != "" {
			relationSeeds[relationCompositeKey(workItemEntityKey, runEntityKey, enumtypes.MissionControlRelationKindSpawnedRun)] = relationSeed{
				sourceEntityKey: workItemEntityKey,
				targetEntityKey: runEntityKey,
				relationKind:    enumtypes.MissionControlRelationKindSpawnedRun,
			}
		}
	}

	if runEntityKey != "" {
		runPayload := valuetypes.MissionControlRunProjectionPayload{
			RunID:              strings.TrimSpace(run.RunID),
			RepositoryFullName: repositoryFullName,
			AgentKey:           strings.TrimSpace(run.AgentKey),
			LastStatus:         strings.TrimSpace(run.Status),
			RuntimeMode:        strings.TrimSpace(runContext.RuntimeMode),
			WaitingReason:      strings.TrimSpace(runContext.WaitReason),
			TriggerLabel:       strings.TrimSpace(runContext.TriggerLabel),
			StageLabel:         strings.TrimSpace(runContext.StageLabel),
			IssueRef:           workItemEntityKey,
			PullRequestRef:     pullRequestEntityKey,
			BranchHead:         strings.TrimSpace(runContext.PullRequestHead),
			BranchBase:         strings.TrimSpace(runContext.PullRequestBase),
			CandidateNamespace: strings.TrimSpace(runContext.CandidateNamespace),
			StartedAt:          cloneProjectionTime(runContext.StartedAt),
			FinishedAt:         cloneProjectionTime(runContext.FinishedAt),
			LastHeartbeatAt:    cloneProjectionTime(runContext.LastHeartbeatAt),
			LastProviderSyncAt: timePointer(projectedAt),
		}
		seedProjection(entitySeeds, runEntityKey, projectionSeed{
			ProjectID:         project.ID,
			EntityKind:        enumtypes.MissionControlEntityKindRun,
			EntityExternalKey: runEntityKey,
			ProviderKind:      enumtypes.MissionControlProviderKindPlatform,
			Title:             runDisplayTitle(run.RunID),
			ActiveState:       warmupActiveState(run.Status, pullRequestEntityKey != "", false),
			SyncStatus:        enumtypes.MissionControlSyncStatusSynced,
			ContinuityStatus:  runContinuity,
			CoverageClass:     runCoverage,
			ProjectionVersion: projectedAt.UnixMilli(),
			CardPayloadJSON:   mustMarshal(runPayload),
			DetailPayloadJSON: mustMarshal(runPayload),
			LastTimelineAt:    lastTimelineAt,
			ProjectedAt:       projectedAt,
			StaleAfter:        &staleAfter,
		})
		if nextRunEntityKey != "" {
			relationSeeds[relationCompositeKey(runEntityKey, nextRunEntityKey, enumtypes.MissionControlRelationKindContinuesWith)] = relationSeed{
				sourceEntityKey: runEntityKey,
				targetEntityKey: nextRunEntityKey,
				relationKind:    enumtypes.MissionControlRelationKindContinuesWith,
			}
		}
	}

	if pullRequestEntityKey != "" {
		pullRequestURL := strings.TrimSpace(run.PullRequestURL)
		if pullRequestURL == "" {
			pullRequestURL = githubPullRequestURL(repositoryFullName, run.PullRequestNumber)
		}
		pullRequestPayload := valuetypes.MissionControlPullRequestProjectionPayload{
			RepositoryFullName: repositoryFullName,
			PullRequestNumber:  run.PullRequestNumber,
			PullRequestURL:     pullRequestURL,
			LastRunID:          run.RunID,
			LastStatus:         run.Status,
			BranchHead:         runContext.PullRequestHead,
			BranchBase:         runContext.PullRequestBase,
		}
		if workItemEntityKey != "" {
			pullRequestPayload.LinkedIssueRefs = []string{workItemEntityKey}
		}
		pullRequestTitle := strings.TrimSpace(runContext.PullRequestTitle)
		if pullRequestTitle == "" {
			pullRequestTitle = fmt.Sprintf("PR #%d", run.PullRequestNumber)
		}
		seedProjection(entitySeeds, pullRequestEntityKey, projectionSeed{
			ProjectID:         project.ID,
			EntityKind:        enumtypes.MissionControlEntityKindPullRequest,
			EntityExternalKey: pullRequestEntityKey,
			ProviderKind:      enumtypes.MissionControlProviderKindGitHub,
			ProviderURL:       pullRequestURL,
			Title:             pullRequestTitle,
			ActiveState:       warmupActiveState(run.Status, true, false),
			SyncStatus:        enumtypes.MissionControlSyncStatusSynced,
			ContinuityStatus:  pullRequestContinuity,
			CoverageClass:     pullRequestCoverage,
			ProjectionVersion: projectedAt.UnixMilli(),
			CardPayloadJSON:   mustMarshal(pullRequestPayload),
			DetailPayloadJSON: mustMarshal(pullRequestPayload),
			LastTimelineAt:    lastTimelineAt,
			ProviderUpdatedAt: timePointer(projectedAt),
			ProjectedAt:       projectedAt,
			StaleAfter:        &staleAfter,
		})
		if runEntityKey != "" {
			relationSeeds[relationCompositeKey(runEntityKey, pullRequestEntityKey, enumtypes.MissionControlRelationKindProducedPullRequest)] = relationSeed{
				sourceEntityKey: runEntityKey,
				targetEntityKey: pullRequestEntityKey,
				relationKind:    enumtypes.MissionControlRelationKindProducedPullRequest,
			}
		}
		if workItemEntityKey != "" {
			relationSeeds[relationCompositeKey(workItemEntityKey, pullRequestEntityKey, enumtypes.MissionControlRelationKindRelatedTo)] = relationSeed{
				sourceEntityKey: workItemEntityKey,
				targetEntityKey: pullRequestEntityKey,
				relationKind:    enumtypes.MissionControlRelationKindRelatedTo,
			}
		}
	}

	eventEntityKeys := relatedProjectionKeys(workItemEntityKey, runEntityKey, pullRequestEntityKey)
	if len(eventEntityKeys) == 0 || strings.TrimSpace(run.CorrelationID) == "" {
		return nil
	}
	events, err := s.staffRuns.ListEventsByCorrelation(ctx, run.CorrelationID, s.cfg.TimelineEventLimit)
	if err != nil {
		return err
	}
	latestTimelineAt := time.Time{}
	for _, event := range events {
		s.collectTimelineSeeds(repositoryFullName, run.RunID, event, eventEntityKeys, timelineSeeds)
		occurredAt := safeProjectionTime(event.CreatedAt, s.now)
		if latestTimelineAt.IsZero() || occurredAt.After(latestTimelineAt) {
			latestTimelineAt = occurredAt
		}
	}
	if latestTimelineAt.IsZero() {
		return nil
	}
	for _, entityKey := range eventEntityKeys {
		updateSeedLastTimelineAt(entitySeeds, entityKey, latestTimelineAt)
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
	if !found || candidate.ProjectionVersion >= current.ProjectionVersion {
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
