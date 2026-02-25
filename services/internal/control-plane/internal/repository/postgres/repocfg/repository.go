package repocfg

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/repocfg/dbmodel"
)

var (
	//go:embed sql/list_for_project.sql
	queryListForProject string
	//go:embed sql/get_by_id.sql
	queryGetByID string
	//go:embed sql/upsert.sql
	queryUpsert string
	//go:embed sql/delete.sql
	queryDelete string
	//go:embed sql/find_by_provider_external_id.sql
	queryFindByProviderExternalID string
	//go:embed sql/find_by_provider_owner_name.sql
	queryFindByProviderOwnerName string
	//go:embed sql/get_token_encrypted.sql
	queryGetTokenEncrypted string
	//go:embed sql/get_bot_token_encrypted.sql
	queryGetBotTokenEncrypted string
	//go:embed sql/set_token_encrypted_for_all.sql
	querySetTokenEncryptedForAll string
	//go:embed sql/upsert_bot_params.sql
	queryUpsertBotParams string
	//go:embed sql/upsert_preflight_report.sql
	queryUpsertPreflightReport string
	//go:embed sql/acquire_preflight_lock.sql
	queryAcquirePreflightLock string
	//go:embed sql/release_preflight_lock.sql
	queryReleasePreflightLock string
)

// Repository stores project repository bindings in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL repository binding repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ListForProject returns repository bindings for a project.
func (r *Repository) ListForProject(ctx context.Context, projectID string, limit int) ([]domainrepo.RepositoryBinding, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.Query(ctx, queryListForProject, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}
	items, err := collectRepositoryBindingRows(rows, "repositories")
	if err != nil {
		return nil, err
	}

	out := make([]domainrepo.RepositoryBinding, 0, len(items))
	for _, item := range items {
		out = append(out, repositoryBindingFromDBModel(item))
	}
	return out, nil
}

// GetByID returns one repository binding by id.
func (r *Repository) GetByID(ctx context.Context, repositoryID string) (domainrepo.RepositoryBinding, bool, error) {
	item, ok, err := queryOneRepositoryBinding(ctx, r.db, queryGetByID, repositoryID)
	if err != nil {
		return domainrepo.RepositoryBinding{}, false, fmt.Errorf("get repository binding by id: %w", err)
	}
	if !ok {
		return domainrepo.RepositoryBinding{}, false, nil
	}
	return item, true, nil
}

// Upsert creates or updates a repository binding.
func (r *Repository) Upsert(ctx context.Context, params domainrepo.UpsertParams) (domainrepo.RepositoryBinding, error) {
	item, ok, err := queryOneRepositoryBinding(
		ctx,
		r.db,
		queryUpsert,
		params.ProjectID,
		params.Alias,
		params.Role,
		params.DefaultRef,
		params.Provider,
		params.ExternalID,
		params.Owner,
		params.Name,
		params.TokenEncrypted,
		params.ServicesYAMLPath,
		params.DocsRootPath,
	)
	if err != nil {
		return domainrepo.RepositoryBinding{}, fmt.Errorf("upsert repository binding: %w", err)
	}
	if !ok {
		return domainrepo.RepositoryBinding{}, fmt.Errorf("repository is already attached to another project (provider=%s external_id=%d)", params.Provider, params.ExternalID)
	}
	return item, nil
}

// Delete removes repository binding by id for a project.
func (r *Repository) Delete(ctx context.Context, projectID string, repositoryID string) error {
	return postgres.ExecRequireRowOrWrap(ctx, r.db, queryDelete, "delete repository binding", projectID, repositoryID)
}

// FindByProviderExternalID resolves binding by provider repo id.
func (r *Repository) FindByProviderExternalID(ctx context.Context, provider string, externalID int64) (domainrepo.FindResult, bool, error) {
	item, ok, err := queryOneLookupRow(ctx, r.db, queryFindByProviderExternalID, provider, externalID)
	if err != nil {
		return domainrepo.FindResult{}, false, fmt.Errorf("find repository binding: %w", err)
	}
	if !ok {
		return domainrepo.FindResult{}, false, nil
	}
	return findResultFromDBModel(item), true, nil
}

// FindByProviderOwnerName resolves binding by provider repo slug (owner/name).
func (r *Repository) FindByProviderOwnerName(ctx context.Context, provider string, owner string, name string) (domainrepo.FindResult, bool, error) {
	item, ok, err := queryOneLookupRow(ctx, r.db, queryFindByProviderOwnerName, provider, owner, name)
	if err != nil {
		return domainrepo.FindResult{}, false, fmt.Errorf("find repository binding by owner/name: %w", err)
	}
	if !ok {
		return domainrepo.FindResult{}, false, nil
	}
	return findResultFromDBModel(item), true, nil
}

// GetTokenEncrypted returns encrypted token bytes for a repository binding.
func (r *Repository) GetTokenEncrypted(ctx context.Context, repositoryID string) ([]byte, bool, error) {
	return r.getEncryptedToken(ctx, repositoryID, queryGetTokenEncrypted, "get repository token")
}

// GetBotTokenEncrypted returns encrypted bot token bytes for a repository binding.
func (r *Repository) GetBotTokenEncrypted(ctx context.Context, repositoryID string) ([]byte, bool, error) {
	return r.getEncryptedToken(ctx, repositoryID, queryGetBotTokenEncrypted, "get repository bot token")
}

func (r *Repository) getEncryptedToken(ctx context.Context, repositoryID string, query string, op string) ([]byte, bool, error) {
	token, ok, err := queryOneBytes(ctx, r.db, query, repositoryID)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", op, err)
	}
	if !ok {
		return nil, false, nil
	}
	return token, true, nil
}

// UpsertBotParams updates bot token + params for a repository binding.
func (r *Repository) UpsertBotParams(ctx context.Context, params domainrepo.RepositoryBotParamsUpsertParams) error {
	_, err := r.db.Exec(ctx, queryUpsertBotParams, params.RepositoryID, params.BotTokenEncrypted, params.BotUsername, params.BotEmail)
	if err != nil {
		return fmt.Errorf("upsert repository bot params: %w", err)
	}
	return nil
}

// UpsertPreflightReport updates stored preflight report for a repository binding.
func (r *Repository) UpsertPreflightReport(ctx context.Context, params domainrepo.RepositoryPreflightReportUpsertParams) error {
	_, err := r.db.Exec(ctx, queryUpsertPreflightReport, params.RepositoryID, params.ReportJSON)
	if err != nil {
		return fmt.Errorf("upsert repository preflight report: %w", err)
	}
	return nil
}

func (r *Repository) AcquirePreflightLock(ctx context.Context, params domainrepo.RepositoryPreflightLockAcquireParams) (string, bool, error) {
	repositoryID := strings.TrimSpace(params.RepositoryID)
	lockToken := strings.TrimSpace(params.LockToken)
	if repositoryID == "" || lockToken == "" {
		return "", false, fmt.Errorf("repository id and lock token are required")
	}

	acquiredToken, ok, err := queryOneString(
		ctx,
		r.db,
		queryAcquirePreflightLock,
		repositoryID,
		lockToken,
		strings.TrimSpace(params.LockedByUserID),
		params.LockedUntilUTC,
	)
	if err != nil {
		return "", false, fmt.Errorf("acquire repository preflight lock: %w", err)
	}
	if !ok {
		return "", false, nil
	}
	return strings.TrimSpace(acquiredToken), true, nil
}

func (r *Repository) ReleasePreflightLock(ctx context.Context, repositoryID string, lockToken string) error {
	repositoryID = strings.TrimSpace(repositoryID)
	lockToken = strings.TrimSpace(lockToken)
	if repositoryID == "" || lockToken == "" {
		return nil
	}
	if _, err := r.db.Exec(ctx, queryReleasePreflightLock, repositoryID, lockToken); err != nil {
		return fmt.Errorf("release repository preflight lock: %w", err)
	}
	return nil
}

// SetTokenEncryptedForAll updates encrypted token for all repository bindings.
func (r *Repository) SetTokenEncryptedForAll(ctx context.Context, tokenEncrypted []byte) (int64, error) {
	res, err := r.db.Exec(ctx, querySetTokenEncryptedForAll, tokenEncrypted)
	if err != nil {
		return 0, fmt.Errorf("set repository token for all: %w", err)
	}
	return res.RowsAffected(), nil
}

func queryOneRepositoryBinding(ctx context.Context, db *pgxpool.Pool, query string, args ...any) (domainrepo.RepositoryBinding, bool, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return domainrepo.RepositoryBinding{}, false, err
	}
	items, err := collectRepositoryBindingRows(rows, "repository binding")
	if err != nil {
		return domainrepo.RepositoryBinding{}, false, err
	}
	if len(items) == 0 {
		return domainrepo.RepositoryBinding{}, false, nil
	}
	return repositoryBindingFromDBModel(items[0]), true, nil
}

func collectRepositoryBindingRows(rows pgx.Rows, operationLabel string) ([]dbmodel.RepositoryBindingRow, error) {
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.RepositoryBindingRow])
	if err != nil {
		return nil, fmt.Errorf("collect %s: %w", operationLabel, err)
	}
	return items, nil
}

func repositoryBindingFromDBModel(row dbmodel.RepositoryBindingRow) domainrepo.RepositoryBinding {
	return domainrepo.RepositoryBinding{
		ID:                 strings.TrimSpace(row.ID),
		ProjectID:          strings.TrimSpace(row.ProjectID),
		Alias:              strings.TrimSpace(row.Alias),
		Role:               strings.TrimSpace(row.Role),
		DefaultRef:         strings.TrimSpace(row.DefaultRef),
		Provider:           strings.TrimSpace(row.Provider),
		ExternalID:         row.ExternalID,
		Owner:              strings.TrimSpace(row.Owner),
		Name:               strings.TrimSpace(row.Name),
		ServicesYAMLPath:   strings.TrimSpace(row.ServicesYAMLPath),
		DocsRootPath:       strings.TrimSpace(row.DocsRootPath),
		BotUsername:        strings.TrimSpace(row.BotUsername),
		BotEmail:           strings.TrimSpace(row.BotEmail),
		PreflightUpdatedAt: strings.TrimSpace(row.PreflightUpdatedAt),
	}
}

func queryOneLookupRow(ctx context.Context, db *pgxpool.Pool, query string, args ...any) (dbmodel.RepositoryBindingLookupRow, bool, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return dbmodel.RepositoryBindingLookupRow{}, false, err
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByPos[dbmodel.RepositoryBindingLookupRow])
	if err != nil {
		return dbmodel.RepositoryBindingLookupRow{}, false, fmt.Errorf("collect repository binding lookup: %w", err)
	}
	if len(items) == 0 {
		return dbmodel.RepositoryBindingLookupRow{}, false, nil
	}
	return items[0], true, nil
}

func findResultFromDBModel(row dbmodel.RepositoryBindingLookupRow) domainrepo.FindResult {
	return domainrepo.FindResult{
		ProjectID:        strings.TrimSpace(row.ProjectID),
		RepositoryID:     strings.TrimSpace(row.RepositoryID),
		ServicesYAMLPath: strings.TrimSpace(row.ServicesYAMLPath),
		DefaultRef:       strings.TrimSpace(row.DefaultRef),
	}
}

func queryOneString(ctx context.Context, db *pgxpool.Pool, query string, args ...any) (string, bool, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return "", false, err
	}
	items, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return "", false, fmt.Errorf("collect string row: %w", err)
	}
	if len(items) == 0 {
		return "", false, nil
	}
	return items[0], true, nil
}

func queryOneBytes(ctx context.Context, db *pgxpool.Pool, query string, args ...any) ([]byte, bool, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	items, err := pgx.CollectRows(rows, pgx.RowTo[[]byte])
	if err != nil {
		return nil, false, fmt.Errorf("collect bytes row: %w", err)
	}
	if len(items) == 0 {
		return nil, false, nil
	}
	return items[0], true, nil
}
