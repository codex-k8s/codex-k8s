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

	libslauncher "github.com/codex-k8s/codex-k8s/libs/go/k8s/joblauncher"
	"github.com/codex-k8s/codex-k8s/libs/go/postgres"
	k8slauncher "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/clients/kubernetes/launcher"
	"github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/controlplane"
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
	runLeaseTTL, err := time.ParseDuration(cfg.RunLeaseTTL)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_WORKER_RUN_LEASE_TTL: %w", err)
	}
	if runLeaseTTL <= 0 {
		return fmt.Errorf("CODEXK8S_WORKER_RUN_LEASE_TTL must be > 0")
	}
	tickTimeout, err := time.ParseDuration(cfg.TickTimeout)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_WORKER_TICK_TIMEOUT: %w", err)
	}
	if tickTimeout <= 0 {
		return fmt.Errorf("CODEXK8S_WORKER_TICK_TIMEOUT must be > 0")
	}
	runtimePrepareRetryTimeout, err := time.ParseDuration(cfg.RuntimePrepareRetryTimeout)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_WORKER_RUNTIME_PREPARE_RETRY_TIMEOUT: %w", err)
	}
	if runtimePrepareRetryTimeout <= 0 {
		return fmt.Errorf("CODEXK8S_WORKER_RUNTIME_PREPARE_RETRY_TIMEOUT must be > 0")
	}
	runtimePrepareRetryInterval, err := time.ParseDuration(cfg.RuntimePrepareRetryInterval)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_WORKER_RUNTIME_PREPARE_RETRY_INTERVAL: %w", err)
	}
	if runtimePrepareRetryInterval <= 0 {
		return fmt.Errorf("CODEXK8S_WORKER_RUNTIME_PREPARE_RETRY_INTERVAL must be > 0")
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
	namespaceLeaseDefaultTTL, namespaceLeaseTTLByRole := loadWebhookRuntimeNamespaceTTLPolicy(cfg, logger)

	dialCtx, cancelDial := context.WithTimeout(appCtx, 30*time.Second)
	defer cancelDial()
	controlPlane, err := controlplane.Dial(dialCtx, cfg.ControlPlaneGRPCTarget)
	if err != nil {
		return fmt.Errorf("dial control-plane grpc: %w", err)
	}
	defer func() { _ = controlPlane.Close() }()

	db, err := postgres.OpenPGXPool(appCtx, postgres.OpenParams{
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
	defer db.Close()

	runs := runqueuerepo.NewRepository(db)
	events := floweventrepo.NewRepository(db)
	feedback := learningfeedbackrepo.NewRepository(db)
	launcher, err := k8slauncher.NewAdapter(libslauncher.Config{
		KubeconfigPath:           cfg.KubeconfigPath,
		Namespace:                cfg.K8sNamespace,
		Image:                    cfg.JobImage,
		Command:                  cfg.JobCommand,
		TTLSeconds:               cfg.JobTTLSeconds,
		BackoffLimit:             cfg.JobBackoffLimit,
		ActiveDeadlineSeconds:    cfg.JobActiveDeadlineSeconds,
		RunServiceAccountName:    cfg.RunServiceAccountName,
		RunRoleName:              cfg.RunRoleName,
		RunRoleBindingName:       cfg.RunRoleBindingName,
		RunResourceQuotaName:     cfg.RunResourceQuotaName,
		RunLimitRangeName:        cfg.RunLimitRangeName,
		RunCredentialsSecretName: cfg.RunCredentialsSecretName,
		RunResourceQuotaPods:     cfg.RunResourceQuotaPods,
	})
	if err != nil {
		return fmt.Errorf("create kubernetes launcher: %w", err)
	}

	service := worker.NewService(worker.Config{
		WorkerID:                    cfg.WorkerID,
		ClaimLimit:                  cfg.ClaimLimit,
		RunningCheckLimit:           cfg.RunningCheckLimit,
		SlotsPerProject:             cfg.SlotsPerProject,
		SlotLeaseTTL:                slotLeaseTTL,
		RunLeaseTTL:                 runLeaseTTL,
		RuntimePrepareRetryTimeout:  runtimePrepareRetryTimeout,
		RuntimePrepareRetryInterval: runtimePrepareRetryInterval,
		ProjectLearningModeDefault:  learningDefault,
		RunNamespacePrefix:          cfg.RunNamespacePrefix,
		DefaultNamespaceTTL:         namespaceLeaseDefaultTTL,
		NamespaceTTLByRole:          namespaceLeaseTTLByRole,
		NamespaceLeaseSweepLimit:    cfg.NamespaceLeaseSweepLimit,
		StateInReviewLabel:          cfg.StateInReviewLabel,
		ControlPlaneGRPCTarget:      cfg.ControlPlaneGRPCTarget,
		ControlPlaneMCPBaseURL:      cfg.ControlPlaneMCPBaseURL,
		OpenAIAPIKey:                cfg.OpenAIAPIKey,
		Context7APIKey:              cfg.Context7APIKey,
		GitBotToken:                 cfg.GitBotToken,
		GitBotUsername:              cfg.GitBotUsername,
		GitBotMail:                  cfg.GitBotMail,
		AgentDefaultModel:           cfg.AgentDefaultModel,
		AgentDefaultReasoningEffort: cfg.AgentDefaultReasoningEffort,
		AgentDefaultLocale:          cfg.AgentDefaultLocale,
		AgentBaseBranch:             cfg.AgentBaseBranch,
		AIModelGPT53CodexLabel:      cfg.AIModelGPT53CodexLabel,
		AIModelGPT53CodexSparkLabel: cfg.AIModelGPT53CodexSparkLabel,
		AIModelGPT52CodexLabel:      cfg.AIModelGPT52CodexLabel,
		AIModelGPT52Label:           cfg.AIModelGPT52Label,
		AIModelGPT51CodexMaxLabel:   cfg.AIModelGPT51CodexMaxLabel,
		AIModelGPT51CodexMiniLabel:  cfg.AIModelGPT51CodexMiniLabel,
		AIReasoningLowLabel:         cfg.AIReasoningLowLabel,
		AIReasoningMediumLabel:      cfg.AIReasoningMediumLabel,
		AIReasoningHighLabel:        cfg.AIReasoningHighLabel,
		AIReasoningExtraHighLabel:   cfg.AIReasoningExtraHighLabel,
	}, worker.Dependencies{
		Runs:            runs,
		Events:          events,
		Feedback:        feedback,
		Launcher:        launcher,
		RuntimePreparer: controlPlane,
		MCPTokenIssuer:  controlPlane,
		RunStatus:       controlPlane,
		Logger:          logger,
	})

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
			tickCtx, cancel := context.WithTimeout(ctx, tickTimeout)
			err := service.Tick(tickCtx)
			cancel()
			if err != nil {
				logger.Error("worker tick failed", "err", err)
			}
		}
	}
}
