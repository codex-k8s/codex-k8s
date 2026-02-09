package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/webhook"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/repository/postgres/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/repository/postgres/flowevent"
	httptransport "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http"
)

// Run starts api-gateway and blocks until shutdown or fatal server error.
func Run() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	agentRuns := agentrunrepo.NewRepository(db)
	flowEvents := floweventrepo.NewRepository(db)
	webhookService := webhook.NewService(agentRuns, flowEvents)

	server := httptransport.NewServer(httptransport.ServerConfig{
		HTTPAddr:            cfg.HTTPAddr,
		GitHubWebhookSecret: cfg.GitHubWebhookSecret,
		MaxBodyBytes:        cfg.WebhookMaxBodyBytes,
	}, webhookService, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("api-gateway started", "addr", cfg.HTTPAddr)
		serverErr <- server.Start()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		logger.Info("shutting down api-gateway")
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown api-gateway: %w", err)
		}
		return nil
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("api-gateway server failed: %w", err)
		}
		return nil
	}
}

func openDB(cfg Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBSSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}
