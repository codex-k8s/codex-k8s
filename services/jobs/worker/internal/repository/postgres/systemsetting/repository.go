package systemsetting

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/get_boolean_by_key.sql
	queryGetBooleanByKey string
	//go:embed sql/listen.sql
	queryListen string
)

// Repository loads worker-visible runtime settings from PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func ListenQuery() string {
	return queryListen
}

func (r *Repository) GetBoolean(ctx context.Context, key string) (bool, bool, error) {
	var raw []byte
	err := r.db.QueryRow(ctx, queryGetBooleanByKey, strings.TrimSpace(key)).Scan(&raw)
	if err == nil {
		var value bool
		if unmarshalErr := json.Unmarshal(raw, &value); unmarshalErr != nil {
			return false, false, fmt.Errorf("decode worker system setting %q boolean value: %w", key, unmarshalErr)
		}
		return value, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return false, false, nil
	}
	return false, false, fmt.Errorf("load worker system setting %q: %w", key, err)
}
