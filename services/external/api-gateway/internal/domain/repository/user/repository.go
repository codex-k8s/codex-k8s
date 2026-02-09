package user

import "context"

// User is a persisted staff user record.
type User struct {
	ID              string
	Email           string
	GitHubUserID    int64
	GitHubLogin     string
	IsPlatformAdmin bool
}

// Repository stores and loads users.
type Repository interface {
	// EnsureOwner makes sure the owner email exists and is a platform admin.
	EnsureOwner(ctx context.Context, email string) (User, error)
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
}
