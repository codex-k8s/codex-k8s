package user

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/codex-k8s/codex-k8s/libs/go/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/user"
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
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL user repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// EnsureOwner inserts owner email as platform admin when missing.
func (r *Repository) EnsureOwner(ctx context.Context, email string) (domainrepo.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx, queryEnsureOwner, email))
	if err != nil {
		return domainrepo.User{}, fmt.Errorf("ensure owner: %w", err)
	}
	return u, nil
}

func (r *Repository) getOne(ctx context.Context, query string, errContext string, args ...any) (domainrepo.User, bool, error) {
	u, err := scanUser(r.db.QueryRow(ctx, query, args...))
	if err == nil {
		return u, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domainrepo.User{}, false, nil
	}
	return domainrepo.User{}, false, fmt.Errorf("%s: %w", errContext, err)
}

// GetByID returns a user by id.
func (r *Repository) GetByID(ctx context.Context, userID string) (domainrepo.User, bool, error) {
	return r.getOne(ctx, queryGetByID, "get by id", userID)
}

// GetByEmail returns a user by email.
func (r *Repository) GetByEmail(ctx context.Context, email string) (domainrepo.User, bool, error) {
	return r.getOne(ctx, queryGetByEmail, "get by email", email)
}

// GetByGitHubLogin returns a user by GitHub login (case-insensitive).
func (r *Repository) GetByGitHubLogin(ctx context.Context, githubLogin string) (domainrepo.User, bool, error) {
	return r.getOne(ctx, queryGetByGitHubLogin, "get by github login", githubLogin)
}

// UpdateGitHubIdentity updates GitHub user id/login for an existing user.
func (r *Repository) UpdateGitHubIdentity(ctx context.Context, userID string, githubUserID int64, githubLogin string) error {
	return postgres.ExecOrWrap(ctx, r.db, queryUpdateGitHubIdentity, "update github identity", userID, githubUserID, githubLogin)
}

// CreateAllowedUser creates or updates an allowed user record.
func (r *Repository) CreateAllowedUser(ctx context.Context, email string, isPlatformAdmin bool) (domainrepo.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx, queryCreateAllowedUser, email, isPlatformAdmin))
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
	rows, err := r.db.Query(ctx, queryListUsers, limit)
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
	return postgres.ExecRequireRowOrWrap(ctx, r.db, queryDeleteByID, "delete user by id", userID)
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
