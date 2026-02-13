package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

	dsn := BuildDSN(params)

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

// OpenPGXPool opens a pgx native pool and verifies it via Ping.
func OpenPGXPool(ctx context.Context, params OpenParams) (*pgxpool.Pool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if params.PingTimeout <= 0 {
		params.PingTimeout = 5 * time.Second
	}

	cfg, err := pgxpool.ParseConfig(BuildDSN(params))
	if err != nil {
		return nil, fmt.Errorf("parse pgx pool config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open pgx pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, params.PingTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pgx pool: %w", err)
	}

	return pool, nil
}

// BuildDSN builds PostgreSQL connection string from open params.
func BuildDSN(params OpenParams) string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		params.Host,
		params.Port,
		params.DBName,
		params.User,
		params.Password,
		params.SSLMode,
	)
}
