package missioncontrol

import (
	"context"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	missioncontrolrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/missioncontrol"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

// ListActiveSet returns one typed active-set slice for future transport adapters.
func (s *Service) ListActiveSet(ctx context.Context, params ActiveSetQuery) (ActiveSet, error) {
	if err := s.ensureReadAllowed(); err != nil {
		return ActiveSet{}, err
	}
	entities, err := s.repository.ListEntities(ctx, missioncontrolrepo.EntityListFilter(params))
	if err != nil {
		return ActiveSet{}, err
	}

	entityIndex := make(map[int64]Entity, len(entities))
	for _, entity := range entities {
		entityIndex[entity.ID] = entity
	}
	relationIDs := make(map[int64]struct{})
	relations := make([]valuetypes.MissionControlRelationView, 0)
	for _, entity := range entities {
		entityRelations, relationErr := s.repository.ListRelationsForEntity(ctx, entity.ProjectID, entity.ID)
		if relationErr != nil {
			return ActiveSet{}, relationErr
		}
		for _, relation := range entityRelations {
			if _, seen := relationIDs[relation.ID]; seen {
				continue
			}
			relationIDs[relation.ID] = struct{}{}
			relationView, viewErr := s.buildRelationView(ctx, entity.ProjectID, relation, entityIndex)
			if viewErr != nil {
				return ActiveSet{}, viewErr
			}
			relations = append(relations, relationView)
		}
	}

	return ActiveSet{
		Entities:  entities,
		Relations: relations,
	}, nil
}

// GetEntityDetails returns one entity together with relation graph and timeline mirror.
func (s *Service) GetEntityDetails(ctx context.Context, params EntityDetailsQuery) (EntityDetails, error) {
	if err := s.ensureReadAllowed(); err != nil {
		return EntityDetails{}, err
	}
	entity, found, err := s.repository.GetEntityByPublicID(ctx, params.ProjectID, params.EntityKind, params.EntityPublicID)
	if err != nil {
		return EntityDetails{}, err
	}
	if !found {
		return EntityDetails{}, errs.NotFound{Msg: "mission control entity not found"}
	}

	relations, err := s.repository.ListRelationsForEntity(ctx, entity.ProjectID, entity.ID)
	if err != nil {
		return EntityDetails{}, err
	}
	entityIndex := map[int64]Entity{entity.ID: entity}
	relationViews, err := s.buildRelationViews(ctx, entity.ProjectID, relations, entityIndex)
	if err != nil {
		return EntityDetails{}, err
	}
	timeline, err := s.repository.ListTimelineEntries(ctx, missioncontrolrepo.TimelineListFilter{
		ProjectID: entity.ProjectID,
		EntityID:  entity.ID,
		Limit:     normalizeTimelineLimit(params.TimelineLimit, s.cfg.DefaultTimelineLimit),
	})
	if err != nil {
		return EntityDetails{}, err
	}

	return EntityDetails{
		Entity:    entity,
		Relations: relationViews,
		Timeline:  timeline,
	}, nil
}

func (s *Service) buildRelationViews(
	ctx context.Context,
	projectID string,
	relations []Relation,
	entityIndex map[int64]Entity,
) ([]valuetypes.MissionControlRelationView, error) {
	if len(relations) == 0 {
		return nil, nil
	}
	out := make([]valuetypes.MissionControlRelationView, 0, len(relations))
	for _, relation := range relations {
		relationView, err := s.buildRelationView(ctx, projectID, relation, entityIndex)
		if err != nil {
			return nil, err
		}
		out = append(out, relationView)
	}
	return out, nil
}

func (s *Service) buildRelationView(
	ctx context.Context,
	projectID string,
	relation Relation,
	entityIndex map[int64]Entity,
) (valuetypes.MissionControlRelationView, error) {
	sourceEntity, err := s.lookupRelationEntity(ctx, projectID, relation.SourceEntityID, entityIndex)
	if err != nil {
		return valuetypes.MissionControlRelationView{}, err
	}
	targetEntity, err := s.lookupRelationEntity(ctx, projectID, relation.TargetEntityID, entityIndex)
	if err != nil {
		return valuetypes.MissionControlRelationView{}, err
	}
	return valuetypes.MissionControlRelationView{
		RelationKind:    relation.RelationKind,
		SourceKind:      relation.SourceKind,
		SourceEntityRef: entityRefFromEntity(sourceEntity),
		TargetEntityRef: entityRefFromEntity(targetEntity),
	}, nil
}

func (s *Service) lookupRelationEntity(
	ctx context.Context,
	projectID string,
	entityID int64,
	entityIndex map[int64]Entity,
) (Entity, error) {
	if entity, ok := entityIndex[entityID]; ok {
		return entity, nil
	}
	entity, found, err := s.repository.GetEntityByID(ctx, projectID, entityID)
	if err != nil {
		return Entity{}, err
	}
	if !found {
		return Entity{}, errs.NotFound{Msg: "mission control relation entity not found"}
	}
	entityIndex[entityID] = entity
	return entity, nil
}

// GetCommandStatus returns one typed command status view for read-side polling fallback.
func (s *Service) GetCommandStatus(ctx context.Context, projectID string, commandID string) (CommandStatusView, error) {
	if err := s.ensureReadAllowed(); err != nil {
		return CommandStatusView{}, err
	}
	command, found, err := s.repository.GetCommandByID(ctx, projectID, commandID)
	if err != nil {
		return CommandStatusView{}, err
	}
	if !found {
		return CommandStatusView{}, errs.NotFound{Msg: "mission control command not found"}
	}
	resultPayload, err := decodeCommandResultPayload(command.ResultPayloadJSON)
	if err != nil {
		return CommandStatusView{}, err
	}
	return CommandStatusView{
		Command:             command,
		EntityRefs:          resultPayload.EntityRefs,
		Approval:            resultPayload.Approval,
		StatusMessage:       resultPayload.StatusMessage,
		ProviderDeliveryIDs: resultPayload.ProviderDeliveryIDs,
	}, nil
}
