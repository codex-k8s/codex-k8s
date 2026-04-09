package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	libslauncher "github.com/codex-k8s/kodex/libs/go/k8s/joblauncher"
	"github.com/codex-k8s/kodex/libs/go/postgres"
	sharedsystemsettings "github.com/codex-k8s/kodex/libs/go/systemsettings"
	k8slauncher "github.com/codex-k8s/kodex/services/jobs/worker/internal/clients/kubernetes/launcher"
	"github.com/codex-k8s/kodex/services/jobs/worker/internal/controlplane"
	systemsettingsdomain "github.com/codex-k8s/kodex/services/jobs/worker/internal/domain/systemsettings"
	"github.com/codex-k8s/kodex/services/jobs/worker/internal/domain/worker"
	floweventrepo "github.com/codex-k8s/kodex/services/jobs/worker/internal/repository/postgres/flowevent"
	learningfeedbackrepo "github.com/codex-k8s/kodex/services/jobs/worker/internal/repository/postgres/learningfeedback"
	runqueuerepo "github.com/codex-k8s/kodex/services/jobs/worker/internal/repository/postgres/runqueue"
	systemsettingrepo "github.com/codex-k8s/kodex/services/jobs/worker/internal/repository/postgres/systemsetting"
	workerinstancerepo "github.com/codex-k8s/kodex/services/jobs/worker/internal/repository/postgres/workerinstance"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
		return fmt.Errorf("parse KODEX_WORKER_POLL_INTERVAL: %w", err)
	}
	if pollInterval <= 0 {
		return fmt.Errorf("KODEX_WORKER_POLL_INTERVAL must be > 0")
	}
	workerHeartbeatInterval, err := time.ParseDuration(cfg.WorkerHeartbeatInterval)
	if err != nil {
		return fmt.Errorf("parse KODEX_WORKER_HEARTBEAT_INTERVAL: %w", err)
	}
	if workerHeartbeatInterval <= 0 {
		return fmt.Errorf("KODEX_WORKER_HEARTBEAT_INTERVAL must be > 0")
	}
	workerInstanceTTL, err := time.ParseDuration(cfg.WorkerInstanceTTL)
	if err != nil {
		return fmt.Errorf("parse KODEX_WORKER_INSTANCE_TTL: %w", err)
	}
	if workerInstanceTTL <= 0 {
		return fmt.Errorf("KODEX_WORKER_INSTANCE_TTL must be > 0")
	}
	if workerInstanceTTL <= workerHeartbeatInterval {
		return fmt.Errorf("KODEX_WORKER_INSTANCE_TTL must be greater than KODEX_WORKER_HEARTBEAT_INTERVAL")
	}

	slotLeaseTTL, err := time.ParseDuration(cfg.SlotLeaseTTL)
	if err != nil {
		return fmt.Errorf("parse KODEX_WORKER_SLOT_LEASE_TTL: %w", err)
	}
	if slotLeaseTTL <= 0 {
		return fmt.Errorf("KODEX_WORKER_SLOT_LEASE_TTL must be > 0")
	}
	runLeaseTTL, err := time.ParseDuration(cfg.RunLeaseTTL)
	if err != nil {
		return fmt.Errorf("parse KODEX_WORKER_RUN_LEASE_TTL: %w", err)
	}
	if runLeaseTTL <= 0 {
		return fmt.Errorf("KODEX_WORKER_RUN_LEASE_TTL must be > 0")
	}
	tickTimeout, err := time.ParseDuration(cfg.TickTimeout)
	if err != nil {
		return fmt.Errorf("parse KODEX_WORKER_TICK_TIMEOUT: %w", err)
	}
	if tickTimeout <= 0 {
		return fmt.Errorf("KODEX_WORKER_TICK_TIMEOUT must be > 0")
	}
	runtimePrepareRetryTimeout, err := time.ParseDuration(cfg.RuntimePrepareRetryTimeout)
	if err != nil {
		return fmt.Errorf("parse KODEX_WORKER_RUNTIME_PREPARE_RETRY_TIMEOUT: %w", err)
	}
	if runtimePrepareRetryTimeout <= 0 {
		return fmt.Errorf("KODEX_WORKER_RUNTIME_PREPARE_RETRY_TIMEOUT must be > 0")
	}
	runtimePrepareRetryInterval, err := time.ParseDuration(cfg.RuntimePrepareRetryInterval)
	if err != nil {
		return fmt.Errorf("parse KODEX_WORKER_RUNTIME_PREPARE_RETRY_INTERVAL: %w", err)
	}
	if runtimePrepareRetryInterval <= 0 {
		return fmt.Errorf("KODEX_WORKER_RUNTIME_PREPARE_RETRY_INTERVAL must be > 0")
	}
	missionControlWarmupInterval, err := time.ParseDuration(cfg.MissionControlWarmupInterval)
	if err != nil {
		return fmt.Errorf("parse KODEX_WORKER_MISSION_CONTROL_WARMUP_INTERVAL: %w", err)
	}
	if missionControlWarmupInterval <= 0 {
		return fmt.Errorf("KODEX_WORKER_MISSION_CONTROL_WARMUP_INTERVAL must be > 0")
	}
	missionControlClaimTTL, err := time.ParseDuration(cfg.MissionControlClaimTTL)
	if err != nil {
		return fmt.Errorf("parse KODEX_WORKER_MISSION_CONTROL_CLAIM_TTL: %w", err)
	}
	if missionControlClaimTTL <= 0 {
		return fmt.Errorf("KODEX_WORKER_MISSION_CONTROL_CLAIM_TTL must be > 0")
	}
	missionControlRetryBaseInterval, err := time.ParseDuration(cfg.MissionControlRetryBaseInterval)
	if err != nil {
		return fmt.Errorf("parse KODEX_WORKER_MISSION_CONTROL_RETRY_BASE_INTERVAL: %w", err)
	}
	if missionControlRetryBaseInterval <= 0 {
		return fmt.Errorf("KODEX_WORKER_MISSION_CONTROL_RETRY_BASE_INTERVAL must be > 0")
	}
	jobImageCheckTimeout, err := time.ParseDuration(cfg.JobImageCheckTimeout)
	if err != nil {
		return fmt.Errorf("parse KODEX_WORKER_JOB_IMAGE_CHECK_TIMEOUT: %w", err)
	}
	if jobImageCheckTimeout <= 0 {
		return fmt.Errorf("KODEX_WORKER_JOB_IMAGE_CHECK_TIMEOUT must be > 0")
	}
	telegramInteractionAdapterTimeout, err := time.ParseDuration(cfg.TelegramInteractionAdapterTimeout)
	if err != nil {
		return fmt.Errorf("parse KODEX_TELEGRAM_INTERACTION_ADAPTER_TIMEOUT: %w", err)
	}
	if telegramInteractionAdapterTimeout <= 0 {
		return fmt.Errorf("KODEX_TELEGRAM_INTERACTION_ADAPTER_TIMEOUT must be > 0")
	}

	learningDefault := false
	if strings.TrimSpace(cfg.LearningModeDefault) != "" {
		v, err := strconv.ParseBool(cfg.LearningModeDefault)
		if err != nil {
			return fmt.Errorf("parse KODEX_LEARNING_MODE_DEFAULT=%q: %w", cfg.LearningModeDefault, err)
		}
		learningDefault = v
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	namespaceLeaseDefaultTTL, namespaceLeaseTTLByRole := loadWebhookRuntimeNamespaceTTLPolicy(cfg, logger)
	ctx, stop := signal.NotifyContext(appCtx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer stop()

	dialCtx, cancelDial := context.WithTimeout(appCtx, 30*time.Second)
	defer cancelDial()
	controlPlane, err := controlplane.Dial(dialCtx, cfg.ControlPlaneGRPCTarget)
	if err != nil {
		return fmt.Errorf("dial control-plane grpc: %w", err)
	}
	defer func() { _ = controlPlane.Close() }()

	dbOpenParams := postgres.OpenParams{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		DBName:   cfg.DBName,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		SSLMode:  cfg.DBSSLMode,
	}
	db, err := postgres.OpenPGXPool(appCtx, dbOpenParams)
	if err != nil {
		return err
	}
	defer db.Close()

	runs := runqueuerepo.NewRepository(db)
	events := floweventrepo.NewRepository(db)
	feedback := learningfeedbackrepo.NewRepository(db)
	workerSystemSettingsRepo := systemsettingrepo.NewRepository(db)
	workerInstances := workerinstancerepo.NewRepository(db)
	workerSystemSettings, err := systemsettingsdomain.NewService(workerSystemSettingsRepo)
	if err != nil {
		return fmt.Errorf("init worker system settings service: %w", err)
	}
	if err := sharedsystemsettings.StartReloadLoop(
		ctx,
		sharedsystemsettings.ReloadLoopConfig{
			DSN:         postgres.BuildDSN(dbOpenParams),
			ListenQuery: systemsettingrepo.ListenQuery(),
		},
		logger,
		workerSystemSettings.RefreshCache,
	); err != nil {
		return fmt.Errorf("start worker system settings reload loop: %w", err)
	}
	launcher, err := k8slauncher.NewAdapter(libslauncher.Config{
		KubeconfigPath:                cfg.KubeconfigPath,
		Namespace:                     cfg.K8sNamespace,
		Image:                         cfg.JobImage,
		Command:                       cfg.JobCommand,
		TTLSeconds:                    cfg.JobTTLSeconds,
		BackoffLimit:                  cfg.JobBackoffLimit,
		ActiveDeadlineSeconds:         cfg.JobActiveDeadlineSeconds,
		RunServiceAccountName:         cfg.RunServiceAccountName,
		RunRoleName:                   cfg.RunRoleName,
		RunRoleBindingName:            cfg.RunRoleBindingName,
		RunReadOnlyServiceAccountName: cfg.RunReadOnlyServiceAccountName,
		RunReadOnlyRoleName:           cfg.RunReadOnlyRoleName,
		RunReadOnlyRoleBindingName:    cfg.RunReadOnlyRoleBindingName,
		RunResourceQuotaName:          cfg.RunResourceQuotaName,
		RunLimitRangeName:             cfg.RunLimitRangeName,
		RunCredentialsSecretName:      cfg.RunCredentialsSecretName,
		RunResourceQuotaPods:          cfg.RunResourceQuotaPods,
	})
	if err != nil {
		return fmt.Errorf("create kubernetes launcher: %w", err)
	}
	jobImageChecker, err := newRegistryJobImageChecker(cfg.InternalRegistryScheme, cfg.InternalRegistryHost, jobImageCheckTimeout)
	if err != nil {
		return fmt.Errorf("create worker job image checker: %w", err)
	}
	var interactionDispatcher worker.InteractionDispatcher
	if strings.TrimSpace(cfg.TelegramInteractionAdapterBaseURL) != "" {
		interactionDispatcher, err = worker.NewTelegramInteractionDispatcher(worker.TelegramInteractionDispatcherConfig{
			BaseURL:     cfg.TelegramInteractionAdapterBaseURL,
			BearerToken: cfg.TelegramInteractionAdapterBearerToken,
			Timeout:     telegramInteractionAdapterTimeout,
		})
		if err != nil {
			return fmt.Errorf("create telegram interaction dispatcher: %w", err)
		}
	} else {
		logger.Warn("telegram interaction adapter is not configured; dispatch attempts will fail with typed non-retryable outcome")
		interactionDispatcher = worker.NewUnavailableInteractionDispatcher(
			"telegram",
			worker.TelegramInteractionErrorAdapterNotConfigured(),
			"telegram interaction adapter base URL is not configured",
		)
	}

	service := worker.NewService(worker.Config{
		WorkerID:                          cfg.WorkerID,
		ClaimLimit:                        cfg.ClaimLimit,
		RunningCheckLimit:                 cfg.RunningCheckLimit,
		StaleLeaseSweepLimit:              cfg.StaleLeaseSweepLimit,
		SlotsPerProject:                   cfg.SlotsPerProject,
		SlotLeaseTTL:                      slotLeaseTTL,
		RunLeaseTTL:                       runLeaseTTL,
		RuntimePrepareRetryTimeout:        runtimePrepareRetryTimeout,
		RuntimePrepareRetryInterval:       runtimePrepareRetryInterval,
		GitHubRateLimitSweepLimit:         cfg.GitHubRateLimitSweepLimit,
		MissionControlWarmupInterval:      missionControlWarmupInterval,
		MissionControlWarmupProjectLimit:  cfg.MissionControlWarmupProjectLimit,
		MissionControlPendingCommandLimit: cfg.MissionControlPendingCommandLimit,
		MissionControlClaimTTL:            missionControlClaimTTL,
		MissionControlRetryMaxAttempts:    cfg.MissionControlRetryMaxAttempts,
		MissionControlRetryBaseInterval:   missionControlRetryBaseInterval,
		ProjectLearningModeDefault:        learningDefault,
		RunNamespacePrefix:                cfg.RunNamespacePrefix,
		RunNamespaceCleanupEnabled:        cfg.RunNamespaceCleanup,
		DefaultNamespaceTTL:               namespaceLeaseDefaultTTL,
		NamespaceTTLByRole:                namespaceLeaseTTLByRole,
		NamespaceLeaseSweepLimit:          cfg.NamespaceLeaseSweepLimit,
		StateInReviewLabel:                cfg.StateInReviewLabel,
		ControlPlaneGRPCTarget:            cfg.ControlPlaneGRPCTarget,
		ControlPlaneMCPBaseURL:            cfg.ControlPlaneMCPBaseURL,
		OpenAIAPIKey:                      cfg.OpenAIAPIKey,
		Context7APIKey:                    cfg.Context7APIKey,
		GitBotToken:                       cfg.GitBotToken,
		GitBotUsername:                    cfg.GitBotUsername,
		GitBotMail:                        cfg.GitBotMail,
		AgentDefaultModel:                 cfg.AgentDefaultModel,
		AgentDefaultReasoningEffort:       cfg.AgentDefaultReasoningEffort,
		AgentDefaultLocale:                cfg.AgentDefaultLocale,
		AgentBaseBranch:                   cfg.AgentBaseBranch,
		JobImage:                          cfg.JobImage,
		JobImageFallback:                  cfg.JobImageFallback,
		KubernetesNamespace:               cfg.K8sNamespace,
		ProductionNamespace:               cfg.ProductionNamespace,
		WorkerPodNamespace:                cfg.WorkerPodNamespace,
		AIRepairNamespace:                 resolveAIRepairNamespace(cfg),
		AIRepairServiceAccount:            cfg.AIRepairServiceAccount,
		AIModelGPT54Label:                 cfg.AIModelGPT54Label,
		AIModelGPT53CodexLabel:            cfg.AIModelGPT53CodexLabel,
		AIModelGPT53CodexSparkLabel:       cfg.AIModelGPT53CodexSparkLabel,
		AIModelGPT52CodexLabel:            cfg.AIModelGPT52CodexLabel,
		AIModelGPT52Label:                 cfg.AIModelGPT52Label,
		AIModelGPT51CodexMaxLabel:         cfg.AIModelGPT51CodexMaxLabel,
		AIModelGPT51CodexMiniLabel:        cfg.AIModelGPT51CodexMiniLabel,
		AIReasoningLowLabel:               cfg.AIReasoningLowLabel,
		AIReasoningMediumLabel:            cfg.AIReasoningMediumLabel,
		AIReasoningHighLabel:              cfg.AIReasoningHighLabel,
		AIReasoningExtraHighLabel:         cfg.AIReasoningExtraHighLabel,
	}, worker.Dependencies{
		Runs:                  runs,
		Events:                events,
		Feedback:              feedback,
		Launcher:              launcher,
		RuntimePreparer:       controlPlane,
		MCPTokenIssuer:        controlPlane,
		RunStatus:             controlPlane,
		Interactions:          controlPlane,
		GitHubRateLimits:      controlPlane,
		MissionControl:        controlPlane,
		InteractionDispatcher: interactionDispatcher,
		JobImageChecker:       jobImageChecker,
		Logger:                logger,
		SystemSettings:        workerSystemSettings,
	})

	if resolveWorkerMode(cfg.Mode) == workerModeNamespaceCleanupOnce {
		cleanupCtx, cancelCleanup := context.WithTimeout(appCtx, tickTimeout)
		defer cancelCleanup()
		return service.RunNamespaceCleanupOnce(cleanupCtx)
	}

	workerStartedAt := time.Now().UTC()
	initialHeartbeatCtx, cancelInitialHeartbeat := context.WithTimeout(ctx, 5*time.Second)
	if err := workerHeartbeat(initialHeartbeatCtx, workerInstances, workerHeartbeatParams{
		WorkerID:  cfg.WorkerID,
		Namespace: cfg.WorkerPodNamespace,
		PodName:   cfg.WorkerPodName,
		StartedAt: workerStartedAt,
		TTL:       workerInstanceTTL,
		Now:       workerStartedAt,
	}); err != nil {
		cancelInitialHeartbeat()
		return fmt.Errorf("register worker heartbeat: %w", err)
	}
	cancelInitialHeartbeat()

	go runWorkerHeartbeatLoop(ctx, logger, workerInstances, workerHeartbeatLoopParams{
		WorkerID:  cfg.WorkerID,
		Namespace: cfg.WorkerPodNamespace,
		PodName:   cfg.WorkerPodName,
		StartedAt: workerStartedAt,
		Interval:  workerHeartbeatInterval,
		TTL:       workerInstanceTTL,
	})

	httpMux := http.NewServeMux()
	httpMux.Handle("/metrics", promhttp.Handler())
	httpMux.HandleFunc("/health/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	httpMux.HandleFunc("/health/livez", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("alive"))
	})
	httpMux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	httpMux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("alive"))
	})

	httpServer := &http.Server{Addr: cfg.HTTPAddr, Handler: httpMux}
	httpErrCh := make(chan error, 1)
	go func() {
		logger.Info("worker http started", "addr", cfg.HTTPAddr)
		httpErrCh <- httpServer.ListenAndServe()
	}()

	logger.Info("worker started", "worker_id", cfg.WorkerID, "poll_interval", pollInterval.String())

	if err := service.Tick(ctx); err != nil {
		logger.Error("initial worker tick failed", "err", err)
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case err := <-httpErrCh:
			if err == nil || err == http.ErrServerClosed {
				return nil
			}
			return fmt.Errorf("worker http server failed: %w", err)
		case <-ctx.Done():
			shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				cancelShutdown()
				return fmt.Errorf("shutdown worker http server: %w", err)
			}
			cancelShutdown()
			stopCtx, cancelStop := context.WithTimeout(context.Background(), 5*time.Second)
			if err := service.ReleaseOwnedRunLeasesOnShutdown(stopCtx); err != nil {
				logger.Warn("release owned run leases on shutdown failed", "worker_id", cfg.WorkerID, "err", err)
			}
			if err := markWorkerStopped(stopCtx, workerInstances, cfg.WorkerID, time.Now().UTC()); err != nil {
				logger.Warn("mark worker stopped failed", "worker_id", cfg.WorkerID, "err", err)
			}
			cancelStop()
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
