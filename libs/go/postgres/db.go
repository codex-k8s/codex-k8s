package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// OpenParams defines the minimal required connection settings for Postgres.
type OpenParams struct {
	Host     string
	Port     int
	DBName   string
	User     string
	Password string
	SSLMode  string

	PingTimeout time.Duration
}

// Open opens a Postgres connection using pgx stdlib driver and verifies it via Ping.
func Open(ctx context.Context, params OpenParams) (*sql.DB, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if params.PingTimeout <= 0 {
		params.PingTimeout = 5 * time.Second
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		params.Host,
		params.Port,
		params.DBName,
		params.User,
		params.Password,
		params.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, params.PingTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

