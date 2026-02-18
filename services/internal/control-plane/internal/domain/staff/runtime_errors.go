package staff

import (
	"context"
	"fmt"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	runtimeerrorrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimeerror"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

const (
	defaultRuntimeErrorListLimit = 100
	maxRuntimeErrorListLimit     = 1000
)

// ListRuntimeErrors returns runtime errors visible to principal.
func (s *Service) ListRuntimeErrors(ctx context.Context, principal Principal, filter querytypes.RuntimeErrorListFilter) ([]runtimeerrorrepo.Item, error) {
	if s.runtimeErrors == nil {
		return nil, fmt.Errorf("runtime error repository is not configured")
	}
	normalized := normalizeRuntimeErrorListFilter(filter)
	if principal.IsPlatformAdmin {
		return s.runtimeErrors.ListAll(ctx, normalized)
	}
	return s.runtimeErrors.ListForUser(ctx, principal.UserID, normalized)
}

// MarkRuntimeErrorViewed marks runtime error as viewed.
func (s *Service) MarkRuntimeErrorViewed(ctx context.Context, principal Principal, runtimeErrorID string) (runtimeerrorrepo.Item, error) {
	if s.runtimeErrors == nil {
		return runtimeerrorrepo.Item{}, fmt.Errorf("runtime error repository is not configured")
	}
	normalizedID := strings.TrimSpace(runtimeErrorID)
	if normalizedID == "" {
		return runtimeerrorrepo.Item{}, errs.Validation{Field: "runtime_error_id", Msg: "is required"}
	}

	item, ok, err := s.runtimeErrors.GetByID(ctx, normalizedID)
	if err != nil {
		return runtimeerrorrepo.Item{}, err
	}
	if !ok {
		return runtimeerrorrepo.Item{}, errs.Validation{Field: "runtime_error_id", Msg: "not found"}
	}
	if !principal.IsPlatformAdmin {
		if strings.TrimSpace(item.ProjectID) == "" {
			return runtimeerrorrepo.Item{}, errs.Forbidden{Msg: "platform admin required"}
		}
		_, hasRole, err := s.members.GetRole(ctx, item.ProjectID, principal.UserID)
		if err != nil {
			return runtimeerrorrepo.Item{}, err
		}
		if !hasRole {
			return runtimeerrorrepo.Item{}, errs.Forbidden{Msg: "project access required"}
		}
	}

	updated, updatedOK, err := s.runtimeErrors.MarkViewed(ctx, querytypes.RuntimeErrorMarkViewedParams{
		ID:       normalizedID,
		ViewerID: principal.UserID,
	})
	if err != nil {
		return runtimeerrorrepo.Item{}, err
	}
	if !updatedOK {
		return runtimeerrorrepo.Item{}, errs.Validation{Field: "runtime_error_id", Msg: "not found"}
	}
	return updated, nil
}

func normalizeRuntimeErrorListFilter(filter querytypes.RuntimeErrorListFilter) querytypes.RuntimeErrorListFilter {
	normalized := filter
	if normalized.Limit <= 0 {
		normalized.Limit = defaultRuntimeErrorListLimit
	}
	if normalized.Limit > maxRuntimeErrorListLimit {
		normalized.Limit = maxRuntimeErrorListLimit
	}
	normalized.Level = strings.TrimSpace(normalized.Level)
	normalized.Source = strings.TrimSpace(normalized.Source)
	normalized.RunID = strings.TrimSpace(normalized.RunID)
	normalized.CorrelationID = strings.TrimSpace(normalized.CorrelationID)
	switch normalized.State {
	case querytypes.RuntimeErrorListStateActive, querytypes.RuntimeErrorListStateViewed, querytypes.RuntimeErrorListStateAll:
	default:
		normalized.State = querytypes.RuntimeErrorListStateActive
	}
	return normalized
}
