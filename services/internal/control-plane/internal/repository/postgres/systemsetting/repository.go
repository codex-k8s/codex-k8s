package systemsetting

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/systemsetting"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

var (
	//go:embed sql/list.sql
	queryList string
	//go:embed sql/get_by_key_for_update.sql
	queryGetByKeyForUpdate string
	//go:embed sql/upsert.sql
	queryUpsert string
	//go:embed sql/insert_change.sql
	queryInsertChange string
	//go:embed sql/notify_change.sql
	queryNotifyChange string
	//go:embed sql/listen.sql
	queryListen string
)

type row struct {
	Key             string    `db:"key"`
	ValueKind       string    `db:"value_kind"`
	ValueJSON       []byte    `db:"value_json"`
	Source          string    `db:"source"`
	Version         int64     `db:"version"`
	UpdatedByUserID *string   `db:"updated_by_user_id"`
	UpdatedByEmail  *string   `db:"updated_by_email"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// Repository stores control-plane-owned platform settings in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL system settings repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ListenQuery returns one embedded LISTEN statement used by reload loops.
func ListenQuery() string {
	return queryListen
}

// List returns current persisted platform settings snapshot.
func (r *Repository) List(ctx context.Context) ([]domainrepo.SystemSettingRecord, error) {
	rows, err := r.db.Query(ctx, queryList)
	if err != nil {
		return nil, fmt.Errorf("list system settings: %w", err)
	}
	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[row])
	if err != nil {
		return nil, fmt.Errorf("collect system settings: %w", err)
	}

	out := make([]domainrepo.SystemSettingRecord, 0, len(items))
	for _, item := range items {
		decoded, err := recordFromRow(item)
		if err != nil {
			return nil, err
		}
		out = append(out, decoded)
	}
	return out, nil
}

// UpsertBoolean persists one typed boolean setting and writes versioned audit row + NOTIFY.
func (r *Repository) UpsertBoolean(ctx context.Context, params domainrepo.BooleanWriteParams) (domainrepo.SystemSettingRecord, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.SystemSettingRecord{}, fmt.Errorf("begin system settings tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	current, found, err := loadCurrentRow(ctx, tx, string(params.Key))
	if err != nil {
		return domainrepo.SystemSettingRecord{}, err
	}

	valueJSON, err := json.Marshal(params.BooleanValue)
	if err != nil {
		return domainrepo.SystemSettingRecord{}, fmt.Errorf("marshal system setting %q boolean value: %w", params.Key, err)
	}

	nextVersion := int64(1)
	var previousValueJSON []byte
	if found {
		nextVersion = current.Version + 1
		previousValueJSON = append([]byte(nil), current.ValueJSON...)
	}

	nextRow := row{}
	if err := tx.QueryRow(
		ctx,
		queryUpsert,
		string(params.Key),
		string(enumtypes.SystemSettingValueKindBoolean),
		valueJSON,
		string(params.Source),
		nextVersion,
		strings.TrimSpace(params.ActorUserID),
		strings.TrimSpace(params.ActorEmail),
	).Scan(
		&nextRow.Key,
		&nextRow.ValueKind,
		&nextRow.ValueJSON,
		&nextRow.Source,
		&nextRow.Version,
		&nextRow.UpdatedByUserID,
		&nextRow.UpdatedByEmail,
		&nextRow.UpdatedAt,
	); err != nil {
		return domainrepo.SystemSettingRecord{}, fmt.Errorf("upsert system setting %q: %w", params.Key, err)
	}

	if _, err := tx.Exec(
		ctx,
		queryInsertChange,
		string(params.Key),
		string(enumtypes.SystemSettingValueKindBoolean),
		valueJSON,
		nullableJSON(previousValueJSON),
		string(params.Source),
		nextVersion,
		string(params.ChangeKind),
		strings.TrimSpace(params.ActorUserID),
		strings.TrimSpace(params.ActorEmail),
	); err != nil {
		return domainrepo.SystemSettingRecord{}, fmt.Errorf("insert system setting %q change row: %w", params.Key, err)
	}

	if _, err := tx.Exec(ctx, queryNotifyChange, string(params.Key), nextVersion); err != nil {
		return domainrepo.SystemSettingRecord{}, fmt.Errorf("notify system setting %q change: %w", params.Key, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domainrepo.SystemSettingRecord{}, fmt.Errorf("commit system setting %q mutation: %w", params.Key, err)
	}

	return recordFromRow(nextRow)
}

func loadCurrentRow(ctx context.Context, tx pgx.Tx, key string) (row, bool, error) {
	current := row{}
	err := tx.QueryRow(ctx, queryGetByKeyForUpdate, strings.TrimSpace(key)).Scan(
		&current.Key,
		&current.ValueKind,
		&current.ValueJSON,
		&current.Source,
		&current.Version,
		&current.UpdatedByUserID,
		&current.UpdatedByEmail,
		&current.UpdatedAt,
	)
	if err == nil {
		return current, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return row{}, false, nil
	}
	return row{}, false, fmt.Errorf("load system setting %q for update: %w", key, err)
}

func recordFromRow(item row) (entitytypes.SystemSettingRecord, error) {
	if strings.TrimSpace(item.ValueKind) != string(enumtypes.SystemSettingValueKindBoolean) {
		return entitytypes.SystemSettingRecord{}, fmt.Errorf("unsupported system setting %q value kind %q", item.Key, item.ValueKind)
	}

	var booleanValue bool
	if err := json.Unmarshal(item.ValueJSON, &booleanValue); err != nil {
		return entitytypes.SystemSettingRecord{}, fmt.Errorf("decode system setting %q boolean value: %w", item.Key, err)
	}

	return entitytypes.SystemSettingRecord{
		Key:             enumtypes.SystemSettingKey(strings.TrimSpace(item.Key)),
		ValueKind:       enumtypes.SystemSettingValueKind(strings.TrimSpace(item.ValueKind)),
		BooleanValue:    booleanValue,
		Source:          enumtypes.SystemSettingSource(strings.TrimSpace(item.Source)),
		Version:         item.Version,
		UpdatedAt:       item.UpdatedAt.UTC(),
		UpdatedByUserID: optionalStringValue(item.UpdatedByUserID),
		UpdatedByEmail:  optionalStringValue(item.UpdatedByEmail),
	}, nil
}

func nullableJSON(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	return json.RawMessage(raw)
}

func optionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
