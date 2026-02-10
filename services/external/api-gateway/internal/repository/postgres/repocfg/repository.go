package repocfg

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	"github.com/codex-k8s/codex-k8s/libs/go/postgres"

	domainrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/repocfg"
)

var (
	//go:embed sql/list_for_project.sql
	queryListForProject string
	//go:embed sql/upsert.sql
	queryUpsert string
	//go:embed sql/delete.sql
	queryDelete string
	//go:embed sql/find_by_provider_external_id.sql
	queryFindByProviderExternalID string
	//go:embed sql/get_token_encrypted.sql
	queryGetTokenEncrypted string
)

// Repository stores project repository bindings in PostgreSQL.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs PostgreSQL repository binding repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ListForProject returns repository bindings for a project.
func (r *Repository) ListForProject(ctx context.Context, projectID string, limit int) ([]domainrepo.RepositoryBinding, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.QueryContext(ctx, queryListForProject, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]domainrepo.RepositoryBinding, 0, limit)
	for rows.Next() {
		var item domainrepo.RepositoryBinding
		if err := rows.Scan(&item.ID, &item.ProjectID, &item.Provider, &item.ExternalID, &item.Owner, &item.Name, &item.ServicesYAMLPath); err != nil {
			return nil, fmt.Errorf("scan repository row: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate repositories: %w", err)
	}
	return out, nil
}

// Upsert creates or updates a repository binding.
func (r *Repository) Upsert(ctx context.Context, params domainrepo.UpsertParams) (domainrepo.RepositoryBinding, error) {
	var item domainrepo.RepositoryBinding
	err := r.db.QueryRowContext(
		ctx,
		queryUpsert,
		params.ProjectID,
		params.Provider,
		params.ExternalID,
		params.Owner,
		params.Name,
		params.TokenEncrypted,
		params.ServicesYAMLPath,
	).Scan(&item.ID, &item.ProjectID, &item.Provider, &item.ExternalID, &item.Owner, &item.Name, &item.ServicesYAMLPath)
	if err == nil {
		return item, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domainrepo.RepositoryBinding{}, fmt.Errorf("repository is already attached to another project (provider=%s external_id=%d)", params.Provider, params.ExternalID)
	}
	return domainrepo.RepositoryBinding{}, fmt.Errorf("upsert repository binding: %w", err)
}

// Delete removes repository binding by id for a project.
func (r *Repository) Delete(ctx context.Context, projectID string, repositoryID string) error {
	return postgres.ExecRequireRowOrWrap(ctx, r.db, queryDelete, "delete repository binding", projectID, repositoryID)
}

// FindByProviderExternalID resolves binding by provider repo id.
func (r *Repository) FindByProviderExternalID(ctx context.Context, provider string, externalID int64) (domainrepo.FindResult, bool, error) {
	var res domainrepo.FindResult
	err := r.db.QueryRowContext(ctx, queryFindByProviderExternalID, provider, externalID).Scan(&res.ProjectID, &res.RepositoryID, &res.ServicesYAMLPath)
	if err == nil {
		return res, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domainrepo.FindResult{}, false, nil
	}
	return domainrepo.FindResult{}, false, fmt.Errorf("find repository binding: %w", err)
}

// GetTokenEncrypted returns encrypted token bytes for a repository binding.
func (r *Repository) GetTokenEncrypted(ctx context.Context, repositoryID string) ([]byte, bool, error) {
	var token []byte
	err := r.db.QueryRowContext(ctx, queryGetTokenEncrypted, repositoryID).Scan(&token)
	if err == nil {
		return token, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	return nil, false, fmt.Errorf("get repository token: %w", err)
}
