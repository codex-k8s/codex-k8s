package dbmodel

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// RunAccessKeyRow maps run_access_keys PostgreSQL row.
type RunAccessKeyRow struct {
	RunID         string             `db:"run_id"`
	ProjectID     pgtype.Text        `db:"project_id"`
	CorrelationID string             `db:"correlation_id"`
	RuntimeMode   string             `db:"runtime_mode"`
	Namespace     pgtype.Text        `db:"namespace"`
	TargetEnv     pgtype.Text        `db:"target_env"`
	KeyHash       []byte             `db:"key_hash"`
	Status        string             `db:"status"`
	IssuedAt      time.Time          `db:"issued_at"`
	ExpiresAt     time.Time          `db:"expires_at"`
	RevokedAt     pgtype.Timestamptz `db:"revoked_at"`
	LastUsedAt    pgtype.Timestamptz `db:"last_used_at"`
	CreatedBy     string             `db:"created_by"`
	CreatedAt     time.Time          `db:"created_at"`
	UpdatedAt     time.Time          `db:"updated_at"`
}
