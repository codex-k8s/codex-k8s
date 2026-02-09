package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	libslauncher "github.com/codex-k8s/codex-k8s/libs/go/k8s/joblauncher"
	k8slauncher "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/clients/kubernetes/launcher"
	"github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/worker"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/repository/postgres/flowevent"
	learningfeedbackrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/repository/postgres/learningfeedback"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/repository/postgres/runqueue"
)

// Run starts worker loop and blocks until termination signal.
func Run() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	pollInterval, err := time.ParseDuration(cfg.PollInterval)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_WORKER_POLL_INTERVAL: %w", err)
	}
	if pollInterval <= 0 {
		return fmt.Errorf("CODEXK8S_WORKER_POLL_INTERVAL must be > 0")
	}

	slotLeaseTTL, err := time.ParseDuration(cfg.SlotLeaseTTL)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_WORKER_SLOT_LEASE_TTL: %w", err)
	}
	if slotLeaseTTL <= 0 {
		return fmt.Errorf("CODEXK8S_WORKER_SLOT_LEASE_TTL must be > 0")
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	runs := runqueuerepo.NewRepository(db)
	events := floweventrepo.NewRepository(db)
	feedback := learningfeedbackrepo.NewRepository(db)
	launcher, err := k8slauncher.NewAdapter(libslauncher.Config{
		KubeconfigPath:        cfg.KubeconfigPath,
		Namespace:             cfg.K8sNamespace,
		Image:                 cfg.JobImage,
		Command:               cfg.JobCommand,
		TTLSeconds:            cfg.JobTTLSeconds,
		BackoffLimit:          cfg.JobBackoffLimit,
		ActiveDeadlineSeconds: cfg.JobActiveDeadlineSeconds,
	})
	if err != nil {
		return fmt.Errorf("create kubernetes launcher: %w", err)
	}

	service := worker.NewService(worker.Config{
		WorkerID:          cfg.WorkerID,
		ClaimLimit:        cfg.ClaimLimit,
		RunningCheckLimit: cfg.RunningCheckLimit,
		SlotsPerProject:   cfg.SlotsPerProject,
		SlotLeaseTTL:      slotLeaseTTL,
	}, runs, events, feedback, launcher, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer stop()

	logger.Info("worker started", "worker_id", cfg.WorkerID, "poll_interval", pollInterval.String())

	if err := service.Tick(ctx); err != nil {
		logger.Error("initial worker tick failed", "err", err)
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker stopped")
			return nil
		case <-ticker.C:
			tickCtx, cancel := context.WithTimeout(ctx, pollInterval)
			err := service.Tick(tickCtx)
			cancel()
			if err != nil {
				logger.Error("worker tick failed", "err", err)
			}
		}
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
