package grpc

import (
	"context"
	"reflect"
	"strings"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/staff"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ListSystemSettings(ctx context.Context, req *controlplanev1.ListSystemSettingsRequest) (*controlplanev1.ListSystemSettingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	principal, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}

	items, err := s.staff.ListSystemSettings(ctx, principal)
	if err != nil {
		return nil, toStatus(err)
	}

	out := make([]*controlplanev1.SystemSetting, 0, len(items))
	for _, item := range items {
		out = append(out, systemSettingToProto(item))
	}
	return &controlplanev1.ListSystemSettingsResponse{Items: out}, nil
}

func (s *Server) GetSystemSetting(ctx context.Context, req *controlplanev1.GetSystemSettingRequest) (*controlplanev1.SystemSetting, error) {
	return s.handleSystemSettingRequest(ctx, req, s.staff.GetSystemSetting)
}

func (s *Server) UpdateSystemSettingBoolean(ctx context.Context, req *controlplanev1.UpdateSystemSettingBooleanRequest) (*controlplanev1.SystemSetting, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	principal, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}

	item, err := s.staff.UpdateSystemSettingBoolean(ctx, principal, strings.TrimSpace(req.GetSettingKey()), req.GetBooleanValue())
	if err != nil {
		return nil, toStatus(err)
	}
	return systemSettingToProto(item), nil
}

func (s *Server) ResetSystemSetting(ctx context.Context, req *controlplanev1.ResetSystemSettingRequest) (*controlplanev1.SystemSetting, error) {
	return s.handleSystemSettingRequest(ctx, req, s.staff.ResetSystemSetting)
}

func (s *Server) handleSystemSettingRequest(
	ctx context.Context,
	req systemSettingRequest,
	call staffSystemSettingCall,
) (*controlplanev1.SystemSetting, error) {
	if requestIsNil(req) {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	principal, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}

	item, err := call(ctx, principal, strings.TrimSpace(req.GetSettingKey()))
	if err != nil {
		return nil, toStatus(err)
	}
	return systemSettingToProto(item), nil
}

func systemSettingToProto(item entitytypes.SystemSetting) *controlplanev1.SystemSetting {
	out := &controlplanev1.SystemSetting{
		Key:                 string(item.Key),
		Section:             string(item.Section),
		ValueKind:           string(item.ValueKind),
		ReloadSemantics:     string(item.ReloadSemantics),
		Visibility:          string(item.Visibility),
		BooleanValue:        item.BooleanValue,
		DefaultBooleanValue: item.DefaultBooleanValue,
		Source:              string(item.Source),
		Version:             item.Version,
	}
	if item.UpdatedAt != nil {
		out.UpdatedAt = timestamppb.New(item.UpdatedAt.UTC())
	}
	if value := strings.TrimSpace(item.UpdatedByUserID); value != "" {
		out.UpdatedByUserId = &value
	}
	if value := strings.TrimSpace(item.UpdatedByEmail); value != "" {
		out.UpdatedByEmail = &value
	}
	return out
}

type staffSystemSettingCall func(context.Context, staff.Principal, string) (entitytypes.SystemSetting, error)

type systemSettingRequest interface {
	GetPrincipal() *controlplanev1.Principal
	GetSettingKey() string
}

func requestIsNil(req systemSettingRequest) bool {
	if req == nil {
		return true
	}
	value := reflect.ValueOf(req)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
