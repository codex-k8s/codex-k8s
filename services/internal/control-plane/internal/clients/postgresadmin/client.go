package postgresadmin

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultAdminDatabase = "postgres"

var databaseNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]{0,62}$`)

// Config defines PostgreSQL admin connection parameters.
type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	SSLMode      string
	AdminDBName  string
	ProtectedDBs []string
}

// Client performs idempotent database lifecycle operations.
type Client struct {
	pool         *pgxpool.Pool
	protectedDBs map[string]struct{}
}

// NewClient creates PostgreSQL admin client for database lifecycle operations.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	adminDB := strings.TrimSpace(cfg.AdminDBName)
	if adminDB == "" {
		adminDB = defaultAdminDatabase
	}

	connString := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		strings.TrimSpace(cfg.Host),
		cfg.Port,
		adminDB,
		strings.TrimSpace(cfg.User),
		cfg.Password,
		strings.TrimSpace(cfg.SSLMode),
	)
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("open postgres admin pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres admin pool: %w", err)
	}

	protected := make(map[string]struct{}, len(cfg.ProtectedDBs)+1)
	protected[strings.ToLower(adminDB)] = struct{}{}
	for _, item := range cfg.ProtectedDBs {
		name := strings.TrimSpace(strings.ToLower(item))
		if name == "" {
			continue
		}
		protected[name] = struct{}{}
	}

	return &Client{pool: pool, protectedDBs: protected}, nil
}

// Close releases admin pool.
func (c *Client) Close() {
	if c == nil || c.pool == nil {
		return
	}
	c.pool.Close()
}

// EnsureDatabase creates database when missing and returns whether it was created.
func (c *Client) EnsureDatabase(ctx context.Context, databaseName string) (bool, error) {
	name, err := normalizeDatabaseName(databaseName)
	if err != nil {
		return false, err
	}

	exists, err := c.databaseExists(ctx, name)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	query := "CREATE DATABASE " + pgx.Identifier{name}.Sanitize()
	if _, err := c.pool.Exec(ctx, query); err != nil {
		return false, fmt.Errorf("create database %s: %w", name, err)
	}
	return true, nil
}

// DropDatabase drops database when present and returns whether it was deleted.
func (c *Client) DropDatabase(ctx context.Context, databaseName string) (bool, error) {
	name, err := normalizeDatabaseName(databaseName)
	if err != nil {
		return false, err
	}
	if _, blocked := c.protectedDBs[strings.ToLower(name)]; blocked {
		return false, fmt.Errorf("database %q is protected", name)
	}

	exists, err := c.databaseExists(ctx, name)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	query := "DROP DATABASE " + pgx.Identifier{name}.Sanitize() + " WITH (FORCE)"
	if _, err := c.pool.Exec(ctx, query); err != nil {
		return false, fmt.Errorf("drop database %s: %w", name, err)
	}
	return true, nil
}

func (c *Client) databaseExists(ctx context.Context, databaseName string) (bool, error) {
	var exists bool
	if err := c.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", databaseName).Scan(&exists); err != nil {
		return false, fmt.Errorf("check database %s existence: %w", databaseName, err)
	}
	return exists, nil
}

func normalizeDatabaseName(databaseName string) (string, error) {
	name := strings.TrimSpace(databaseName)
	if name == "" {
		return "", fmt.Errorf("database_name is required")
	}
	if !databaseNamePattern.MatchString(name) {
		return "", fmt.Errorf("database_name %q is invalid", name)
	}
	return name, nil
}
