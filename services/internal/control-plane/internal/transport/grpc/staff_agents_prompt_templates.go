package grpc

import (
	"context"
	"strings"
	"time"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/staff"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ListAgents(ctx context.Context, req *controlplanev1.ListAgentsRequest) (*controlplanev1.ListAgentsResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	items, err := s.staff.ListAgents(ctx, p, clampLimit(req.GetLimit(), 200))
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.Agent, 0, len(items))
	for _, item := range items {
		out = append(out, agentToProto(item))
	}
	return &controlplanev1.ListAgentsResponse{Items: out}, nil
}

func (s *Server) GetAgent(ctx context.Context, req *controlplanev1.GetAgentRequest) (*controlplanev1.Agent, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	item, err := s.staff.GetAgent(ctx, p, strings.TrimSpace(req.GetAgentId()))
	if err != nil {
		return nil, toStatus(err)
	}
	return agentToProto(item), nil
}

func (s *Server) UpdateAgentSettings(ctx context.Context, req *controlplanev1.UpdateAgentSettingsRequest) (*controlplanev1.Agent, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	item, err := s.staff.UpdateAgentSettings(ctx, p, querytypes.AgentUpdateSettingsParams{
		AgentID:         strings.TrimSpace(req.GetAgentId()),
		ExpectedVersion: int(req.GetExpectedVersion()),
		Settings:        agentSettingsFromProto(req.GetSettings()),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return agentToProto(item), nil
}

func (s *Server) ListPromptTemplateKeys(ctx context.Context, req *controlplanev1.ListPromptTemplateKeysRequest) (*controlplanev1.ListPromptTemplateKeysResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	items, err := s.staff.ListPromptTemplateKeys(ctx, p, querytypes.PromptTemplateKeyListFilter{
		Limit:     clampLimit(req.GetLimit(), 200),
		Scope:     optionalProtoString(req.Scope),
		ProjectID: optionalProtoString(req.ProjectId),
		Role:      optionalProtoString(req.Role),
		Kind:      optionalProtoString(req.Kind),
		Locale:    optionalProtoString(req.Locale),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.PromptTemplateKey, 0, len(items))
	for _, item := range items {
		out = append(out, promptTemplateKeyToProto(item))
	}
	return &controlplanev1.ListPromptTemplateKeysResponse{Items: out}, nil
}

func (s *Server) ListPromptTemplateVersions(ctx context.Context, req *controlplanev1.ListPromptTemplateVersionsRequest) (*controlplanev1.ListPromptTemplateVersionsResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	items, err := s.staff.ListPromptTemplateVersions(ctx, p, strings.TrimSpace(req.GetTemplateKey()), clampLimit(req.GetLimit(), 200))
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.PromptTemplateVersion, 0, len(items))
	for _, item := range items {
		out = append(out, promptTemplateVersionToProto(item))
	}
	return &controlplanev1.ListPromptTemplateVersionsResponse{Items: out}, nil
}

func (s *Server) CreatePromptTemplateVersion(ctx context.Context, req *controlplanev1.CreatePromptTemplateVersionRequest) (*controlplanev1.PromptTemplateVersion, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	key, err := staff.ParsePromptTemplateKey(strings.TrimSpace(req.GetTemplateKey()))
	if err != nil {
		return nil, toStatus(err)
	}
	item, err := s.staff.CreatePromptTemplateVersion(ctx, p, querytypes.PromptTemplateVersionCreateParams{
		Key:             key,
		BodyMarkdown:    req.GetBodyMarkdown(),
		ExpectedVersion: int(req.GetExpectedVersion()),
		Source:          enumtypes.PromptTemplateSource(optionalProtoString(req.Source)),
		ChangeReason:    optionalProtoString(req.ChangeReason),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return promptTemplateVersionToProto(item), nil
}

func (s *Server) ActivatePromptTemplateVersion(ctx context.Context, req *controlplanev1.ActivatePromptTemplateVersionRequest) (*controlplanev1.PromptTemplateVersion, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	key, err := staff.ParsePromptTemplateKey(strings.TrimSpace(req.GetTemplateKey()))
	if err != nil {
		return nil, toStatus(err)
	}
	item, err := s.staff.ActivatePromptTemplateVersion(ctx, p, querytypes.PromptTemplateVersionActivateParams{
		Key:             key,
		Version:         int(req.GetVersion()),
		ExpectedVersion: int(req.GetExpectedVersion()),
		ChangeReason:    strings.TrimSpace(req.GetChangeReason()),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return promptTemplateVersionToProto(item), nil
}

func (s *Server) SyncPromptTemplateSeeds(ctx context.Context, req *controlplanev1.PromptTemplateSeedSyncRequest) (*controlplanev1.PromptTemplateSeedSyncResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	result, err := s.staff.SyncPromptTemplateSeeds(ctx, p, querytypes.PromptTemplateSeedSyncParams{
		Mode:           req.GetMode(),
		Scope:          optionalProtoString(req.Scope),
		ProjectID:      optionalProtoString(req.ProjectId),
		IncludeLocales: req.GetIncludeLocales(),
		ForceOverwrite: req.GetForceOverwrite(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	items := make([]*controlplanev1.PromptTemplateSeedSyncItem, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, &controlplanev1.PromptTemplateSeedSyncItem{
			TemplateKey: item.TemplateKey,
			Action:      item.Action,
			Checksum:    stringPtrOrNil(item.Checksum),
			Reason:      stringPtrOrNil(item.Reason),
		})
	}
	return &controlplanev1.PromptTemplateSeedSyncResponse{
		CreatedCount: int32(result.CreatedCount),
		UpdatedCount: int32(result.UpdatedCount),
		SkippedCount: int32(result.SkippedCount),
		Items:        items,
	}, nil
}

func (s *Server) PreviewPromptTemplate(ctx context.Context, req *controlplanev1.PreviewPromptTemplateRequest) (*controlplanev1.PreviewPromptTemplateResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	key, err := staff.ParsePromptTemplateKey(strings.TrimSpace(req.GetTemplateKey()))
	if err != nil {
		return nil, toStatus(err)
	}
	if projectID := optionalProtoString(req.ProjectId); projectID != "" && key.Scope == enumtypes.PromptTemplateScopeGlobal {
		key.Scope = enumtypes.PromptTemplateScopeProject
		key.ScopeID = projectID
	}
	lookup := querytypes.PromptTemplatePreviewLookup{Key: key}
	if req.Version != nil {
		lookup.Version = int(req.GetVersion())
	}
	item, err := s.staff.PreviewPromptTemplate(ctx, p, lookup)
	if err != nil {
		return nil, toStatus(err)
	}
	return &controlplanev1.PreviewPromptTemplateResponse{
		TemplateKey:  item.TemplateKey,
		Version:      int32(item.Version),
		Source:       item.Source,
		Checksum:     item.Checksum,
		BodyMarkdown: item.BodyMarkdown,
	}, nil
}

func (s *Server) DiffPromptTemplateVersions(ctx context.Context, req *controlplanev1.DiffPromptTemplateVersionsRequest) (*controlplanev1.DiffPromptTemplateVersionsResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	key, err := staff.ParsePromptTemplateKey(strings.TrimSpace(req.GetTemplateKey()))
	if err != nil {
		return nil, toStatus(err)
	}
	fromItem, toItem, err := s.staff.DiffPromptTemplateVersions(ctx, p, key, int(req.GetFromVersion()), int(req.GetToVersion()))
	if err != nil {
		return nil, toStatus(err)
	}
	return &controlplanev1.DiffPromptTemplateVersionsResponse{
		TemplateKey:      fromItem.TemplateKey,
		FromVersion:      int32(fromItem.Version),
		ToVersion:        int32(toItem.Version),
		FromBodyMarkdown: fromItem.BodyMarkdown,
		ToBodyMarkdown:   toItem.BodyMarkdown,
	}, nil
}

func (s *Server) ListPromptTemplateAuditEvents(ctx context.Context, req *controlplanev1.ListPromptTemplateAuditEventsRequest) (*controlplanev1.ListPromptTemplateAuditEventsResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	items, err := s.staff.ListPromptTemplateAuditEvents(ctx, p, querytypes.PromptTemplateAuditListFilter{
		Limit:       clampLimit(req.GetLimit(), 200),
		ProjectID:   optionalProtoString(req.ProjectId),
		TemplateKey: optionalProtoString(req.TemplateKey),
		ActorID:     optionalProtoString(req.ActorId),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.PromptTemplateAuditEvent, 0, len(items))
	for _, item := range items {
		out = append(out, promptTemplateAuditEventToProto(item))
	}
	return &controlplanev1.ListPromptTemplateAuditEventsResponse{Items: out}, nil
}

func agentToProto(item entitytypes.Agent) *controlplanev1.Agent {
	return &controlplanev1.Agent{
		Id:              item.ID,
		AgentKey:        item.AgentKey,
		RoleKind:        item.RoleKind,
		ProjectId:       stringPtrOrNil(item.ProjectID),
		Name:            item.Name,
		IsActive:        item.IsActive,
		Settings:        agentSettingsToProto(item.Settings),
		SettingsVersion: int32(item.SettingsVersion),
	}
}

func agentSettingsToProto(item entitytypes.AgentSettings) *controlplanev1.AgentSettings {
	return &controlplanev1.AgentSettings{
		RuntimeMode:       item.RuntimeMode,
		TimeoutSeconds:    int32(item.TimeoutSeconds),
		MaxRetryCount:     int32(item.MaxRetryCount),
		PromptLocale:      item.PromptLocale,
		ApprovalsRequired: item.ApprovalsRequired,
	}
}

func agentSettingsFromProto(item *controlplanev1.AgentSettings) entitytypes.AgentSettings {
	if item == nil {
		return entitytypes.AgentSettings{}
	}
	return entitytypes.AgentSettings{
		RuntimeMode:       strings.TrimSpace(item.GetRuntimeMode()),
		TimeoutSeconds:    int(item.GetTimeoutSeconds()),
		MaxRetryCount:     int(item.GetMaxRetryCount()),
		PromptLocale:      strings.TrimSpace(item.GetPromptLocale()),
		ApprovalsRequired: item.GetApprovalsRequired(),
	}
}

func promptTemplateKeyToProto(item entitytypes.PromptTemplateKeyItem) *controlplanev1.PromptTemplateKey {
	return &controlplanev1.PromptTemplateKey{
		TemplateKey:   item.TemplateKey,
		Scope:         item.Scope,
		ProjectId:     stringPtrOrNil(item.ProjectID),
		Role:          item.Role,
		Kind:          item.Kind,
		Locale:        item.Locale,
		ActiveVersion: int32(item.ActiveVersion),
		UpdatedAt:     timestamppb.New(item.UpdatedAt.UTC()),
	}
}

func promptTemplateVersionToProto(item entitytypes.PromptTemplateVersion) *controlplanev1.PromptTemplateVersion {
	return &controlplanev1.PromptTemplateVersion{
		TemplateKey:       item.TemplateKey,
		Version:           int32(item.Version),
		Status:            item.Status,
		Source:            item.Source,
		Checksum:          item.Checksum,
		BodyMarkdown:      item.BodyMarkdown,
		ChangeReason:      stringPtrOrNil(item.ChangeReason),
		SupersedesVersion: int32PtrOrNil(int32Value(item.SupersedesVersion)),
		UpdatedBy:         item.UpdatedBy,
		UpdatedAt:         timestamppb.New(item.UpdatedAt.UTC()),
		ActivatedAt:       tsFromOptional(item.ActivatedAt),
	}
}

func promptTemplateAuditEventToProto(item entitytypes.PromptTemplateAuditEvent) *controlplanev1.PromptTemplateAuditEvent {
	return &controlplanev1.PromptTemplateAuditEvent{
		Id:            item.ID,
		CorrelationId: item.CorrelationID,
		ProjectId:     stringPtrOrNil(item.ProjectID),
		TemplateKey:   stringPtrOrNil(item.TemplateKey),
		Version:       int32PtrOrNil(int32Value(item.Version)),
		ActorId:       stringPtrOrNil(item.ActorID),
		EventType:     item.EventType,
		PayloadJson:   item.PayloadJSON,
		CreatedAt:     timestamppb.New(item.CreatedAt.UTC()),
	}
}

func int32Value(value *int) int32 {
	if value == nil || *value <= 0 {
		return 0
	}
	return int32(*value)
}

func tsFromOptional(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}
	return timestamppb.New(value.UTC())
}

var _ controlplanev1.ControlPlaneServiceServer = (*Server)(nil)
