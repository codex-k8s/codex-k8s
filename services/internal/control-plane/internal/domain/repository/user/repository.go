package user

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

type User = entitytypes.User

// Repository stores and loads users.
type Repository interface {
	// EnsureOwner makes sure the owner email exists and is a platform admin.
	EnsureOwner(ctx context.Context, email string) (User, error)
	// GetByID returns user by id.
	GetByID(ctx context.Context, userID string) (User, bool, error)
	// GetByEmail returns user by email.
	GetByEmail(ctx context.Context, email string) (User, bool, error)
	// GetByGitHubLogin returns user by GitHub login (case-insensitive).
	GetByGitHubLogin(ctx context.Context, githubLogin string) (User, bool, error)
	// UpdateGitHubIdentity updates GitHub identity fields.
	UpdateGitHubIdentity(ctx context.Context, userID string, githubUserID int64, githubLogin string) error
	// CreateAllowedUser creates a new allowed user (without OAuth identity yet).
	CreateAllowedUser(ctx context.Context, email string, isPlatformAdmin bool) (User, error)
	// List returns all users.
	List(ctx context.Context, limit int) ([]User, error)
	// DeleteByID removes a user by id.
	DeleteByID(ctx context.Context, userID string) error
}
