package missioncontrol

import (
	"context"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
)

// RunWarmup returns current warmup summary once rollout guards allow owner-owned writes.
func (s *Service) RunWarmup(ctx context.Context, params WarmupRequest) (WarmupSummary, error) {
	if err := s.ensureDomainWriteAllowed(); err != nil {
		return WarmupSummary{}, err
	}
	summary, err := s.repository.GetWarmupSummary(ctx, params.ProjectID)
	if err != nil {
		return WarmupSummary{}, err
	}
	s.insertFlowEvent(ctx, params.CorrelationID, eventTypeMissionControlWarmupRequested, warmupEventPayload{
		ProjectID:            summary.ProjectID,
		RequestedBy:          params.RequestedBy,
		CorrelationID:        params.CorrelationID,
		EntityCount:          summary.EntityCount,
		RelationCount:        summary.RelationCount,
		TimelineEntryCount:   summary.TimelineEntryCount,
		CommandCount:         summary.CommandCount,
		MaxProjectionVersion: summary.MaxProjectionVersion,
	})
	return summary, nil
}

// UpsertEntity stores one projection row under control-plane ownership.
func (s *Service) UpsertEntity(ctx context.Context, params UpsertEntityParams, correlationID string) (Entity, error) {
	return storeProjectionValue(ctx, s, correlationID, eventTypeMissionControlEntityUpserted, func() (Entity, error) {
		return s.repository.UpsertEntity(ctx, params)
	}, func(entity Entity) any {
		return entityProjectionEventPayload{
			ProjectID:         entity.ProjectID,
			EntityKind:        entity.EntityKind,
			EntityPublicID:    entity.EntityExternalKey,
			ProjectionVersion: entity.ProjectionVersion,
		}
	})
}

// UpdateEntityProjection applies one optimistic projection mutation guarded by projection_version.
func (s *Service) UpdateEntityProjection(ctx context.Context, params UpdateEntityParams, correlationID string) (Entity, error) {
	return s.storeEntityProjection(ctx, correlationID, eventTypeMissionControlEntityUpdated, func() (Entity, error) { return s.repository.UpdateEntityProjection(ctx, params) })
}

// ReplaceRelationsForSource rewrites one entity relation set.
func (s *Service) ReplaceRelationsForSource(ctx context.Context, params ReplaceRelationsParams, correlationID string) error {
	if err := s.ensureDomainWriteAllowed(); err != nil {
		return err
	}
	if err := s.repository.ReplaceRelationsForSource(ctx, params); err != nil {
		return err
	}
	s.insertFlowEvent(ctx, correlationID, eventTypeMissionControlRelationsReplaced, relationReplaceEventPayload{
		ProjectID:      params.ProjectID,
		SourceEntityID: params.SourceEntityID,
		RelationCount:  len(params.Relations),
	})
	return nil
}

// UpsertTimelineEntry stores one timeline mirror entry.
func (s *Service) UpsertTimelineEntry(ctx context.Context, params UpsertTimelineEntryParams, correlationID string) (TimelineEntry, error) {
	if err := s.ensureDomainWriteAllowed(); err != nil {
		return TimelineEntry{}, err
	}
	entry, err := s.repository.UpsertTimelineEntry(ctx, params)
	if err != nil {
		return TimelineEntry{}, err
	}
	payload := timelineEventPayload{
		ProjectID:        entry.ProjectID,
		EntityID:         entry.EntityID,
		EntryExternalKey: entry.EntryExternalKey,
		SourceKind:       entry.SourceKind,
	}
	s.insertFlowEvent(ctx, correlationID, eventTypeMissionControlTimelineUpserted, payload)
	return entry, nil
}

var _ DomainService = (*Service)(nil)
var _ WarmupExecutor = (*Service)(nil)

func (s *Service) storeEntityProjection(
	ctx context.Context,
	correlationID string,
	eventType floweventdomain.EventType,
	op func() (Entity, error),
) (Entity, error) {
	if err := s.ensureDomainWriteAllowed(); err != nil {
		return Entity{}, err
	}
	entity, err := op()
	if err != nil {
		return Entity{}, err
	}
	s.insertFlowEvent(ctx, correlationID, eventType, entityProjectionEventPayload{
		ProjectID:         entity.ProjectID,
		EntityKind:        entity.EntityKind,
		EntityPublicID:    entity.EntityExternalKey,
		ProjectionVersion: entity.ProjectionVersion,
	})
	return entity, nil
}

func storeProjectionValue[T any](
	ctx context.Context,
	service *Service,
	correlationID string,
	eventType floweventdomain.EventType,
	op func() (T, error),
	payload func(T) any,
) (T, error) {
	var zero T
	if err := service.ensureDomainWriteAllowed(); err != nil {
		return zero, err
	}
	value, err := op()
	if err != nil {
		return zero, err
	}
	service.insertFlowEvent(ctx, correlationID, eventType, payload(value))
	return value, nil
}
