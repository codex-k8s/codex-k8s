package user

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	domainrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/user"
)

var (
	//go:embed sql/ensure_owner.sql
	queryEnsureOwner string
	//go:embed sql/get_by_id.sql
	queryGetByID string
	//go:embed sql/get_by_email.sql
	queryGetByEmail string
	//go:embed sql/get_by_github_login.sql
	queryGetByGitHubLogin string
	//go:embed sql/update_github_identity.sql
	queryUpdateGitHubIdentity string
	//go:embed sql/create_allowed_user.sql
	queryCreateAllowedUser string
	//go:embed sql/list_users.sql
	queryListUsers string
	//go:embed sql/delete_by_id.sql
	queryDeleteByID string
)

// Repository stores staff users in PostgreSQL.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs PostgreSQL user repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// EnsureOwner inserts owner email as platform admin when missing.
func (r *Repository) EnsureOwner(ctx context.Context, email string) (domainrepo.User, error) {
	u, err := scanUser(r.db.QueryRowContext(ctx, queryEnsureOwner, email))
	if err != nil {
		return domainrepo.User{}, fmt.Errorf("ensure owner: %w", err)
	}
	return u, nil
}

// GetByID returns a user by id.
func (r *Repository) GetByID(ctx context.Context, userID string) (domainrepo.User, bool, error) {
	u, err := scanUser(r.db.QueryRowContext(ctx, queryGetByID, userID))
	if err == nil {
		return u, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domainrepo.User{}, false, nil
	}
	return domainrepo.User{}, false, fmt.Errorf("get by id: %w", err)
}

// GetByEmail returns a user by email.
func (r *Repository) GetByEmail(ctx context.Context, email string) (domainrepo.User, bool, error) {
	u, err := scanUser(r.db.QueryRowContext(ctx, queryGetByEmail, email))
	if err == nil {
		return u, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domainrepo.User{}, false, nil
	}
	return domainrepo.User{}, false, fmt.Errorf("get by email: %w", err)
}

// GetByGitHubLogin returns a user by GitHub login (case-insensitive).
func (r *Repository) GetByGitHubLogin(ctx context.Context, githubLogin string) (domainrepo.User, bool, error) {
	u, err := scanUser(r.db.QueryRowContext(ctx, queryGetByGitHubLogin, githubLogin))
	if err == nil {
		return u, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domainrepo.User{}, false, nil
	}
	return domainrepo.User{}, false, fmt.Errorf("get by github login: %w", err)
}

// UpdateGitHubIdentity updates GitHub user id/login for an existing user.
func (r *Repository) UpdateGitHubIdentity(ctx context.Context, userID string, githubUserID int64, githubLogin string) error {
	if _, err := r.db.ExecContext(ctx, queryUpdateGitHubIdentity, userID, githubUserID, githubLogin); err != nil {
		return fmt.Errorf("update github identity: %w", err)
	}
	return nil
}

// CreateAllowedUser creates or updates an allowed user record.
func (r *Repository) CreateAllowedUser(ctx context.Context, email string, isPlatformAdmin bool) (domainrepo.User, error) {
	u, err := scanUser(r.db.QueryRowContext(ctx, queryCreateAllowedUser, email, isPlatformAdmin))
	if err != nil {
		return domainrepo.User{}, fmt.Errorf("create allowed user: %w", err)
	}
	return u, nil
}

// List returns all users.
func (r *Repository) List(ctx context.Context, limit int) ([]domainrepo.User, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.QueryContext(ctx, queryListUsers, limit)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var out []domainrepo.User
	for rows.Next() {
		var u domainrepo.User
		if err := rows.Scan(&u.ID, &u.Email, &u.GitHubUserID, &u.GitHubLogin, &u.IsPlatformAdmin, &u.IsPlatformOwner); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return out, nil
}

// DeleteByID deletes a user by id.
func (r *Repository) DeleteByID(ctx context.Context, userID string) error {
	res, err := r.db.ExecContext(ctx, queryDeleteByID, userID)
	if err != nil {
		return fmt.Errorf("delete user by id: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("read rows affected for user delete: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (domainrepo.User, error) {
	var u domainrepo.User
	if err := row.Scan(&u.ID, &u.Email, &u.GitHubUserID, &u.GitHubLogin, &u.IsPlatformAdmin, &u.IsPlatformOwner); err != nil {
		return domainrepo.User{}, err
	}
	return u, nil
}
