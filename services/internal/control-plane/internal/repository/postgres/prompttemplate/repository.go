package prompttemplate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/prompttemplate"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/list_keys.sql
	queryListKeys string
	//go:embed sql/list_versions.sql
	queryListVersions string
	//go:embed sql/get_version.sql
	queryGetVersion string
	//go:embed sql/get_active_version.sql
	queryGetActiveVersion string
	//go:embed sql/lock_latest_version.sql
	queryLockLatestVersion string
	//go:embed sql/insert_version.sql
	queryInsertVersion string
	//go:embed sql/archive_active_versions.sql
	queryArchiveActiveVersions string
	//go:embed sql/activate_version.sql
	queryActivateVersion string
	//go:embed sql/list_audit_events.sql
	queryListAuditEvents string
	//go:embed sql/insert_flow_event.sql
	queryInsertFlowEvent string
)

const (
	eventTypePromptTemplateVersionCreated = "prompt_template.version_created"
	eventTypePromptTemplateActivated      = "prompt_template.activated"
	eventTypePromptTemplateSeedCreated    = "prompt_template.seed_created"
)

// Repository persists prompt template lifecycle state in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs prompt template repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListKeys(ctx context.Context, filter querytypes.PromptTemplateKeyListFilter) ([]domainrepo.KeyItem, error) {
	limit := normalizeLimit(filter.Limit)
	rows, err := r.db.Query(
		ctx,
		queryListKeys,
		strings.TrimSpace(filter.Scope),
		strings.TrimSpace(filter.ProjectID),
		strings.TrimSpace(filter.Role),
		strings.TrimSpace(filter.Kind),
		strings.TrimSpace(filter.Locale),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list prompt template keys: %w", err)
	}

	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domainrepo.KeyItem, error) {
		var (
			item      domainrepo.KeyItem
			projectID pgtype.Text
		)
		if err := row.Scan(
			&item.TemplateKey,
			&item.Scope,
			&projectID,
			&item.Role,
			&item.Kind,
			&item.Locale,
			&item.ActiveVersion,
			&item.UpdatedAt,
		); err != nil {
			return domainrepo.KeyItem{}, err
		}
		if projectID.Valid {
			item.ProjectID = strings.TrimSpace(projectID.String)
		}
		return item, nil
	})
	if err != nil {
		return nil, fmt.Errorf("collect prompt template keys: %w", err)
	}
	return items, nil
}

func (r *Repository) ListVersions(ctx context.Context, filter querytypes.PromptTemplateVersionListFilter) ([]domainrepo.Version, error) {
	limit := normalizeLimit(filter.Limit)
	rows, err := r.db.Query(ctx, queryListVersions, filter.Key.Scope, filter.Key.ScopeID, filter.Key.Role, filter.Key.Kind, filter.Key.Locale, limit)
	if err != nil {
		return nil, fmt.Errorf("list prompt template versions: %w", err)
	}
	items, err := pgx.CollectRows(rows, rowToPromptTemplateVersion)
	if err != nil {
		return nil, fmt.Errorf("collect prompt template versions: %w", err)
	}
	return items, nil
}

func (r *Repository) GetVersion(ctx context.Context, lookup querytypes.PromptTemplateVersionLookup) (domainrepo.Version, bool, error) {
	item, err := queryOneVersion(ctx, r.db, queryGetVersion, lookup.Key.Scope, lookup.Key.ScopeID, lookup.Key.Role, lookup.Key.Kind, lookup.Key.Locale, lookup.Version)
	if err != nil {
		return domainrepo.Version{}, false, fmt.Errorf("get prompt template version: %w", err)
	}
	if item == nil {
		return domainrepo.Version{}, false, nil
	}
	return *item, true, nil
}

func (r *Repository) GetActiveVersion(ctx context.Context, lookup querytypes.PromptTemplatePreviewLookup) (domainrepo.Version, bool, error) {
	item, err := queryOneVersion(ctx, r.db, queryGetActiveVersion, lookup.Key.Scope, lookup.Key.ScopeID, lookup.Key.Role, lookup.Key.Kind, lookup.Key.Locale)
	if err != nil {
		return domainrepo.Version{}, false, fmt.Errorf("get active prompt template version: %w", err)
	}
	if item == nil {
		return domainrepo.Version{}, false, nil
	}
	return *item, true, nil
}

func (r *Repository) CreateVersion(ctx context.Context, params querytypes.PromptTemplateVersionCreateParams) (domainrepo.Version, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.Version{}, fmt.Errorf("begin tx for create prompt template version: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	latestVersion, exists, err := lockLatestVersion(ctx, tx, params.Key)
	if err != nil {
		return domainrepo.Version{}, fmt.Errorf("lock latest prompt template version: %w", err)
	}
	if params.ExpectedVersion != latestVersion {
		return domainrepo.Version{}, errs.Conflict{Msg: fmt.Sprintf("expected_version=%d does not match actual_version=%d", params.ExpectedVersion, latestVersion)}
	}

	source := strings.TrimSpace(string(params.Source))
	if source == "" {
		if params.Key.Scope == enumtypes.PromptTemplateScopeProject {
			source = string(enumtypes.PromptTemplateSourceProjectOverride)
		} else {
			source = string(enumtypes.PromptTemplateSourceGlobalOverride)
		}
	}

	newVersion := latestVersion + 1
	checksum := checksumSHA256(params.BodyMarkdown)
	changeReason := strings.TrimSpace(params.ChangeReason)
	var supersedesVersion interface{}
	if exists {
		supersedesVersion = latestVersion
	}

	item, err := queryOneVersion(
		ctx,
		tx,
		queryInsertVersion,
		params.Key.Scope,
		params.Key.ScopeID,
		params.Key.Role,
		params.Key.Kind,
		params.Key.Locale,
		params.BodyMarkdown,
		source,
		newVersion,
		false,
		string(enumtypes.PromptTemplateStatusDraft),
		checksum,
		changeReason,
		supersedesVersion,
		strings.TrimSpace(params.UpdatedByUserID),
		nil,
	)
	if err != nil {
		return domainrepo.Version{}, fmt.Errorf("insert prompt template version: %w", err)
	}
	if item == nil {
		return domainrepo.Version{}, errors.New("insert prompt template version returned no row")
	}

	if err := insertAuditEvent(ctx, tx, strings.TrimSpace(params.UpdatedByUserID), eventTypePromptTemplateVersionCreated, buildAuditPayload(params.Key, *item)); err != nil {
		return domainrepo.Version{}, fmt.Errorf("insert prompt template create audit event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domainrepo.Version{}, fmt.Errorf("commit create prompt template version: %w", err)
	}
	return *item, nil
}

func (r *Repository) ActivateVersion(ctx context.Context, params querytypes.PromptTemplateVersionActivateParams) (domainrepo.Version, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.Version{}, fmt.Errorf("begin tx for activate prompt template version: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	latestVersion, exists, err := lockLatestVersion(ctx, tx, params.Key)
	if err != nil {
		return domainrepo.Version{}, fmt.Errorf("lock latest prompt template version for activate: %w", err)
	}
	if !exists {
		return domainrepo.Version{}, errs.Validation{Field: "template_key", Msg: "not found"}
	}
	if params.ExpectedVersion != latestVersion {
		return domainrepo.Version{}, errs.Conflict{Msg: fmt.Sprintf("expected_version=%d does not match actual_version=%d", params.ExpectedVersion, latestVersion)}
	}

	updatedBy := strings.TrimSpace(params.UpdatedByUserID)
	if _, err := tx.Exec(ctx, queryArchiveActiveVersions, params.Key.Scope, params.Key.ScopeID, params.Key.Role, params.Key.Kind, params.Key.Locale, updatedBy); err != nil {
		return domainrepo.Version{}, fmt.Errorf("archive active prompt template versions: %w", err)
	}

	item, err := queryOneVersion(
		ctx,
		tx,
		queryActivateVersion,
		params.Key.Scope,
		params.Key.ScopeID,
		params.Key.Role,
		params.Key.Kind,
		params.Key.Locale,
		updatedBy,
		strings.TrimSpace(params.ChangeReason),
		params.Version,
	)
	if err != nil {
		return domainrepo.Version{}, fmt.Errorf("activate prompt template version: %w", err)
	}
	if item == nil {
		return domainrepo.Version{}, errs.Validation{Field: "version", Msg: "not found"}
	}

	if err := insertAuditEvent(ctx, tx, updatedBy, eventTypePromptTemplateActivated, buildAuditPayload(params.Key, *item)); err != nil {
		return domainrepo.Version{}, fmt.Errorf("insert prompt template activate audit event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domainrepo.Version{}, fmt.Errorf("commit activate prompt template version: %w", err)
	}
	return *item, nil
}

func (r *Repository) CreateSeedIfMissing(ctx context.Context, params querytypes.PromptTemplateSeedCreateParams) (domainrepo.Version, bool, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.Version{}, false, fmt.Errorf("begin tx for seed prompt template create: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, exists, err := lockLatestVersion(ctx, tx, params.Key)
	if err != nil {
		return domainrepo.Version{}, false, fmt.Errorf("lock latest prompt template version for seed: %w", err)
	}
	if exists {
		return domainrepo.Version{}, false, nil
	}

	checksum := checksumSHA256(params.BodyMarkdown)
	activatedAt := time.Now().UTC()
	item, err := queryOneVersion(
		ctx,
		tx,
		queryInsertVersion,
		params.Key.Scope,
		params.Key.ScopeID,
		params.Key.Role,
		params.Key.Kind,
		params.Key.Locale,
		params.BodyMarkdown,
		string(enumtypes.PromptTemplateSourceRepoSeed),
		1,
		true,
		string(enumtypes.PromptTemplateStatusActive),
		checksum,
		"seed bootstrap",
		nil,
		strings.TrimSpace(params.UpdatedByUserID),
		activatedAt,
	)
	if err != nil {
		return domainrepo.Version{}, false, fmt.Errorf("insert prompt template seed version: %w", err)
	}
	if item == nil {
		return domainrepo.Version{}, false, errors.New("insert prompt template seed version returned no row")
	}

	if err := insertAuditEvent(ctx, tx, strings.TrimSpace(params.UpdatedByUserID), eventTypePromptTemplateSeedCreated, buildAuditPayload(params.Key, *item)); err != nil {
		return domainrepo.Version{}, false, fmt.Errorf("insert prompt template seed audit event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domainrepo.Version{}, false, fmt.Errorf("commit seed prompt template create: %w", err)
	}
	return *item, true, nil
}

func (r *Repository) ListAuditEvents(ctx context.Context, filter querytypes.PromptTemplateAuditListFilter) ([]domainrepo.AuditEvent, error) {
	limit := normalizeLimit(filter.Limit)
	rows, err := r.db.Query(
		ctx,
		queryListAuditEvents,
		strings.TrimSpace(filter.ProjectID),
		strings.TrimSpace(filter.TemplateKey),
		strings.TrimSpace(filter.ActorID),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list prompt template audit events: %w", err)
	}
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domainrepo.AuditEvent, error) {
		var (
			item      domainrepo.AuditEvent
			versionDB pgtype.Int4
		)
		if err := row.Scan(
			&item.ID,
			&item.CorrelationID,
			&item.ProjectID,
			&item.TemplateKey,
			&versionDB,
			&item.ActorID,
			&item.EventType,
			&item.PayloadJSON,
			&item.CreatedAt,
		); err != nil {
			return domainrepo.AuditEvent{}, err
		}
		item.ProjectID = strings.TrimSpace(item.ProjectID)
		item.TemplateKey = strings.TrimSpace(item.TemplateKey)
		item.ActorID = strings.TrimSpace(item.ActorID)
		if versionDB.Valid {
			value := int(versionDB.Int32)
			item.Version = &value
		}
		return item, nil
	})
	if err != nil {
		return nil, fmt.Errorf("collect prompt template audit events: %w", err)
	}
	return items, nil
}

func lockLatestVersion(ctx context.Context, tx pgx.Tx, key querytypes.PromptTemplateKey) (int, bool, error) {
	var version int
	err := tx.QueryRow(ctx, queryLockLatestVersion, key.Scope, key.ScopeID, key.Role, key.Kind, key.Locale).Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return version, true, nil
}

func insertAuditEvent(ctx context.Context, tx pgx.Tx, actorID string, eventType string, payload map[string]interface{}) error {
	payloadRaw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal audit payload: %w", err)
	}
	actorType := "system"
	if strings.TrimSpace(actorID) != "" {
		actorType = "human"
	}
	if _, err := tx.Exec(
		ctx,
		queryInsertFlowEvent,
		uuid.NewString(),
		actorType,
		strings.TrimSpace(actorID),
		eventType,
		payloadRaw,
	); err != nil {
		return err
	}
	return nil
}

func buildAuditPayload(key querytypes.PromptTemplateKey, item domainrepo.Version) map[string]interface{} {
	payload := map[string]interface{}{
		"template_key": item.TemplateKey,
		"version":      item.Version,
		"status":       item.Status,
		"source":       item.Source,
	}
	if key.Scope == enumtypes.PromptTemplateScopeProject && strings.TrimSpace(key.ScopeID) != "" {
		payload["project_id"] = strings.TrimSpace(key.ScopeID)
	}
	return payload
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 200
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}

func checksumSHA256(body string) string {
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])
}

type queryable interface {
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
}

func queryOneVersion(ctx context.Context, q queryable, query string, args ...interface{}) (*domainrepo.Version, error) {
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	items, err := pgx.CollectRows(rows, rowToPromptTemplateVersion)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func rowToPromptTemplateVersion(row pgx.CollectableRow) (domainrepo.Version, error) {
	var (
		item              domainrepo.Version
		changeReason      pgtype.Text
		supersedesVersion pgtype.Int4
		activatedAt       pgtype.Timestamptz
	)
	if err := row.Scan(
		&item.TemplateKey,
		&item.Version,
		&item.Status,
		&item.Source,
		&item.Checksum,
		&item.BodyMarkdown,
		&changeReason,
		&supersedesVersion,
		&item.UpdatedBy,
		&item.UpdatedAt,
		&activatedAt,
	); err != nil {
		return domainrepo.Version{}, err
	}
	if changeReason.Valid {
		item.ChangeReason = strings.TrimSpace(changeReason.String)
	}
	if supersedesVersion.Valid {
		value := int(supersedesVersion.Int32)
		item.SupersedesVersion = &value
	}
	if activatedAt.Valid {
		value := activatedAt.Time.UTC()
		item.ActivatedAt = &value
	}
	return item, nil
}

