package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	libslauncher "github.com/codex-k8s/codex-k8s/libs/go/k8s/joblauncher"
	"github.com/codex-k8s/codex-k8s/libs/go/postgres"
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

	appCtx := context.Background()

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

	learningDefault := false
	if strings.TrimSpace(cfg.LearningModeDefault) != "" {
		v, err := strconv.ParseBool(cfg.LearningModeDefault)
		if err != nil {
			return fmt.Errorf("parse CODEXK8S_LEARNING_MODE_DEFAULT=%q: %w", cfg.LearningModeDefault, err)
		}
		learningDefault = v
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := postgres.Open(appCtx, postgres.OpenParams{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		DBName:   cfg.DBName,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("db close failed", "err", err)
		}
	}()

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
		WorkerID:                   cfg.WorkerID,
		ClaimLimit:                 cfg.ClaimLimit,
		RunningCheckLimit:          cfg.RunningCheckLimit,
		SlotsPerProject:            cfg.SlotsPerProject,
		SlotLeaseTTL:               slotLeaseTTL,
		ProjectLearningModeDefault: learningDefault,
	}, runs, events, feedback, launcher, logger)

	ctx, stop := signal.NotifyContext(appCtx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
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
