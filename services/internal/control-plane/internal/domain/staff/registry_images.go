package staff

import (
	"context"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

// ListRegistryImages returns internal registry repositories and tags.
func (s *Service) ListRegistryImages(ctx context.Context, principal Principal, filter querytypes.RegistryImageListFilter) ([]entitytypes.RegistryImageRepository, error) {
	if !principal.IsPlatformAdmin {
		return nil, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.images == nil {
		return nil, errs.Validation{Field: "registry", Msg: "image service is not configured"}
	}
	return s.images.List(ctx, filter)
}

// DeleteRegistryImageTag deletes one internal registry tag.
func (s *Service) DeleteRegistryImageTag(ctx context.Context, principal Principal, params querytypes.RegistryImageDeleteParams) (entitytypes.RegistryImageDeleteResult, error) {
	if !principal.IsPlatformAdmin {
		return entitytypes.RegistryImageDeleteResult{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.images == nil {
		return entitytypes.RegistryImageDeleteResult{}, errs.Validation{Field: "registry", Msg: "image service is not configured"}
	}
	return s.images.DeleteTag(ctx, params)
}

// CleanupRegistryImages deletes stale tags in internal registry according to keep policy.
func (s *Service) CleanupRegistryImages(ctx context.Context, principal Principal, filter querytypes.RegistryImageCleanupFilter) (entitytypes.RegistryImageCleanupResult, error) {
	if !principal.IsPlatformAdmin {
		return entitytypes.RegistryImageCleanupResult{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.images == nil {
		return entitytypes.RegistryImageCleanupResult{}, errs.Validation{Field: "registry", Msg: "image service is not configured"}
	}
	return s.images.Cleanup(ctx, filter)
}
