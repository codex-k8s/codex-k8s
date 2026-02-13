package platformtoken

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/platformtoken"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/get.sql
	queryGet string
	//go:embed sql/upsert.sql
	queryUpsert string
)

// Repository stores singleton platform GitHub tokens in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL platform token repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Get returns singleton token row.
func (r *Repository) Get(ctx context.Context) (domainrepo.PlatformGitHubTokens, bool, error) {
	var item domainrepo.PlatformGitHubTokens
	err := r.db.QueryRow(ctx, queryGet).Scan(&item.PlatformTokenEncrypted, &item.BotTokenEncrypted)
	if err == nil {
		return item, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domainrepo.PlatformGitHubTokens{}, false, nil
	}
	return domainrepo.PlatformGitHubTokens{}, false, fmt.Errorf("get platform github tokens: %w", err)
}

// Upsert writes singleton token row.
func (r *Repository) Upsert(ctx context.Context, params domainrepo.UpsertParams) (domainrepo.PlatformGitHubTokens, error) {
	var item domainrepo.PlatformGitHubTokens
	err := r.db.QueryRow(ctx, queryUpsert, params.PlatformTokenEncrypted, params.BotTokenEncrypted).Scan(&item.PlatformTokenEncrypted, &item.BotTokenEncrypted)
	if err != nil {
		return domainrepo.PlatformGitHubTokens{}, fmt.Errorf("upsert platform github tokens: %w", err)
	}
	return item, nil
}
