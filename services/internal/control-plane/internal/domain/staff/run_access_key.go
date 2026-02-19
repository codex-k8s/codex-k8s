package staff

import (
	"context"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	runaccessdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runaccess"
)

// GetRunAccessKeyStatus returns run-scoped OAuth bypass key status snapshot.
func (s *Service) GetRunAccessKeyStatus(ctx context.Context, principal Principal, runID string) (runaccessdomain.KeyStatus, error) {
	if s.runAccess == nil {
		return runaccessdomain.KeyStatus{}, errs.Validation{Field: "run_access_key", Msg: "service is not configured"}
	}
	normalizedRunID := strings.TrimSpace(runID)
	if _, _, err := s.resolveRunAccess(ctx, principal, normalizedRunID); err != nil {
		return runaccessdomain.KeyStatus{}, err
	}
	return s.runAccess.GetStatus(ctx, normalizedRunID)
}

// RegenerateRunAccessKey rotates run-scoped OAuth bypass key and returns plaintext value.
func (s *Service) RegenerateRunAccessKey(ctx context.Context, principal Principal, runID string, ttl time.Duration) (runaccessdomain.IssuedKey, error) {
	if s.runAccess == nil {
		return runaccessdomain.IssuedKey{}, errs.Validation{Field: "run_access_key", Msg: "service is not configured"}
	}
	normalizedRunID := strings.TrimSpace(runID)
	if _, _, err := s.resolveRunAccess(ctx, principal, normalizedRunID); err != nil {
		return runaccessdomain.IssuedKey{}, err
	}

	status, err := s.runAccess.GetStatus(ctx, normalizedRunID)
	if err != nil {
		return runaccessdomain.IssuedKey{}, err
	}

	return s.runAccess.Regenerate(ctx, runaccessdomain.IssueParams{
		RunID:       normalizedRunID,
		RuntimeMode: strings.TrimSpace(status.RuntimeMode),
		Namespace:   strings.TrimSpace(status.Namespace),
		TargetEnv:   strings.TrimSpace(status.TargetEnv),
		CreatedBy:   resolveRunAccessActor(principal),
		TTL:         ttl,
	})
}

// RevokeRunAccessKey revokes active run-scoped OAuth bypass key.
func (s *Service) RevokeRunAccessKey(ctx context.Context, principal Principal, runID string) (runaccessdomain.KeyStatus, error) {
	if s.runAccess == nil {
		return runaccessdomain.KeyStatus{}, errs.Validation{Field: "run_access_key", Msg: "service is not configured"}
	}
	normalizedRunID := strings.TrimSpace(runID)
	if _, _, err := s.resolveRunAccess(ctx, principal, normalizedRunID); err != nil {
		return runaccessdomain.KeyStatus{}, err
	}
	return s.runAccess.Revoke(ctx, normalizedRunID, resolveRunAccessActor(principal))
}

func resolveRunAccessActor(principal Principal) string {
	if value := strings.TrimSpace(principal.Email); value != "" {
		return value
	}
	if value := strings.TrimSpace(principal.GitHubLogin); value != "" {
		return value
	}
	if value := strings.TrimSpace(principal.UserID); value != "" {
		return value
	}
	return "staff"
}
