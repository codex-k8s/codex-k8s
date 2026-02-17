package projecttoken

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projecttoken"
)

var (
	//go:embed sql/get_view.sql
	queryGetView string
	//go:embed sql/get_encrypted.sql
	queryGetEncrypted string
	//go:embed sql/upsert.sql
	queryUpsert string
	//go:embed sql/delete.sql
	queryDelete string
)

// Repository stores project GitHub tokens in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByProjectID(ctx context.Context, projectID string) (domainrepo.ProjectGitHubTokens, bool, error) {
	var item domainrepo.ProjectGitHubTokens
	var hasPlatform bool
	var hasBot bool
	err := r.db.QueryRow(ctx, queryGetView, projectID).Scan(
		&item.ProjectID,
		&hasPlatform,
		&hasBot,
		&item.BotUsername,
		&item.BotEmail,
	)
	if err == nil {
		item.HasPlatformToken = hasPlatform
		item.HasBotToken = hasBot
		return item, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domainrepo.ProjectGitHubTokens{}, false, nil
	}
	return domainrepo.ProjectGitHubTokens{}, false, fmt.Errorf("get project github tokens: %w", err)
}

func (r *Repository) GetEncryptedByProjectID(ctx context.Context, projectID string) (platformToken []byte, botToken []byte, botUsername string, botEmail string, ok bool, err error) {
	err = r.db.QueryRow(ctx, queryGetEncrypted, projectID).Scan(&platformToken, &botToken, &botUsername, &botEmail)
	if err == nil {
		return platformToken, botToken, botUsername, botEmail, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, "", "", false, nil
	}
	return nil, nil, "", "", false, fmt.Errorf("get project github tokens encrypted: %w", err)
}

func (r *Repository) Upsert(ctx context.Context, params domainrepo.UpsertParams) error {
	_, err := r.db.Exec(ctx, queryUpsert, params.ProjectID, params.PlatformTokenEncrypted, params.BotTokenEncrypted, params.BotUsername, params.BotEmail)
	if err != nil {
		return fmt.Errorf("upsert project github tokens: %w", err)
	}
	return nil
}

func (r *Repository) DeleteByProjectID(ctx context.Context, projectID string) error {
	_, err := r.db.Exec(ctx, queryDelete, projectID)
	if err != nil {
		return fmt.Errorf("delete project github tokens: %w", err)
	}
	return nil
}

