package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	"github.com/codex-k8s/codex-k8s/libs/go/crypto/tokencrypt"
	"github.com/codex-k8s/codex-k8s/libs/go/postgres"
	"github.com/codex-k8s/codex-k8s/libs/go/registry"
	repoprovider "github.com/codex-k8s/codex-k8s/libs/go/repo/provider"
	githubprovider "github.com/codex-k8s/codex-k8s/libs/go/repo/provider/github"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	githubclient "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/clients/github"
	githubmgmtclient "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/clients/githubmgmt"
	kubernetesclient "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/clients/kubernetes"
	postgresadminclient "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/clients/postgresadmin"
	agentcallbackdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/agentcallback"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	registryimagesdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/registryimages"
	runstatusdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runstatus"
	runtimedeploydomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runtimedeploy"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/staff"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/webhook"
	agentrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/agent"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/agentrun"
	agentsessionrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/agentsession"
	configentryrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/configentry"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/flowevent"
	learningfeedbackrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/learningfeedback"
	mcpactionrequestrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/mcpactionrequest"
	platformtokenrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/platformtoken"
	projectrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/project"
	projectdatabaserepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/projectdatabase"
	projectmemberrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/projectmember"
	projecttokenrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/projecttoken"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/repocfg"
	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/runtimedeploytask"
	staffrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/staffrun"
	userrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/user"
	grpctransport "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/transport/grpc"
	mcptransport "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/transport/mcp"
)

// Run starts control-plane servers and blocks until shutdown or fatal error.
func Run() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	appCtx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	runCtx, stop := signal.NotifyContext(appCtx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer stop()

	// DB readiness is handled by initContainer in deployment; control-plane starts fail-fast.
	pgxPool, err := postgres.OpenPGXPool(runCtx, postgres.OpenParams{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		DBName:   cfg.DBName,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		return fmt.Errorf("open postgres pgx pool: %w", err)
	}
	defer pgxPool.Close()

	agentRuns := agentrunrepo.NewRepository(pgxPool)
	agents := agentrepo.NewRepository(pgxPool)
	flowEvents := floweventrepo.NewRepository(pgxPool)

	users := userrepo.NewRepository(pgxPool)
	projects := projectrepo.NewRepository(pgxPool)
	members := projectmemberrepo.NewRepository(pgxPool)
	runs := staffrunrepo.NewRepository(pgxPool)
	repos := repocfgrepo.NewRepository(pgxPool)
	feedback := learningfeedbackrepo.NewRepository(pgxPool)
	agentSessions := agentsessionrepo.NewRepository(pgxPool)
	platformTokens := platformtokenrepo.NewRepository(pgxPool)
	projectTokens := projecttokenrepo.NewRepository(pgxPool)
	configEntries := configentryrepo.NewRepository(pgxPool)
	mcpActionRequests := mcpactionrequestrepo.NewRepository(pgxPool)
	projectDatabases := projectdatabaserepo.NewRepository(pgxPool)
	runtimeDeployTasks := runtimedeploytaskrepo.NewRepository(pgxPool)

	tokenCrypto, err := tokencrypt.NewService(cfg.TokenEncryptionKey)
	if err != nil {
		return fmt.Errorf("init token encryption: %w", err)
	}
	if err := syncGitHubTokens(runCtx, syncGitHubTokensParams{
		PlatformTokenRaw: strings.TrimSpace(cfg.GitHubPAT),
		BotTokenRaw:      strings.TrimSpace(cfg.GitBotToken),
		PlatformTokens:   platformTokens,
		Repos:            repos,
		TokenCrypt:       tokenCrypto,
		Logger:           logger,
	}); err != nil {
		return fmt.Errorf("sync github tokens: %w", err)
	}
	k8sClient, err := kubernetesclient.NewClient(cfg.KubeconfigPath)
	if err != nil {
		return fmt.Errorf("init kubernetes mcp client: %w", err)
	}
	postgresAdminClient, err := postgresadminclient.NewClient(runCtx, postgresadminclient.Config{
		Host:         cfg.ProjectDBAdminHost,
		Port:         cfg.ProjectDBAdminPort,
		User:         cfg.ProjectDBAdminUser,
		Password:     cfg.ProjectDBAdminPassword,
		SSLMode:      cfg.ProjectDBAdminSSLMode,
		AdminDBName:  cfg.ProjectDBAdminDatabase,
		ProtectedDBs: []string{cfg.DBName},
	})
	if err != nil {
		return fmt.Errorf("init postgres admin client: %w", err)
	}
	defer postgresAdminClient.Close()
	githubMCPClient := githubclient.NewClient(nil)
	githubMgmtClient := githubmgmtclient.NewClient(nil)
	githubRepoProvider := githubprovider.NewProvider(nil)

	mcpTokenTTL, err := time.ParseDuration(cfg.MCPTokenTTL)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_MCP_TOKEN_TTL=%q: %w", cfg.MCPTokenTTL, err)
	}
	mcpSigningKey := strings.TrimSpace(cfg.MCPTokenSigningKey)
	if mcpSigningKey == "" {
		mcpSigningKey = cfg.TokenEncryptionKey
	}
	mcpService, err := mcpdomain.NewService(mcpdomain.Config{
		TokenSigningKey:              mcpSigningKey,
		PublicBaseURL:                cfg.PublicBaseURL,
		InternalMCPBaseURL:           cfg.ControlPlaneMCPBaseURL,
		DefaultTokenTTL:              mcpTokenTTL,
		DatabaseLifecycleAllowedEnvs: cfg.ProjectDBLifecycleAllowedEnvs,
	}, mcpdomain.Dependencies{
		Runs:             agentRuns,
		FlowEvents:       flowEvents,
		Repos:            repos,
		Platform:         platformTokens,
		Actions:          mcpActionRequests,
		Sessions:         agentSessions,
		ProjectDatabases: projectDatabases,
		TokenCrypt:       tokenCrypto,
		GitHub:           githubMCPClient,
		Kubernetes:       k8sClient,
		Database:         postgresAdminClient,
	})
	if err != nil {
		return fmt.Errorf("init mcp domain service: %w", err)
	}
	runStatusService, err := runstatusdomain.NewService(runstatusdomain.Config{
		PublicBaseURL: cfg.PublicBaseURL,
		DefaultLocale: "ru",
	}, runstatusdomain.Dependencies{
		Runs:       agentRuns,
		Platform:   platformTokens,
		TokenCrypt: tokenCrypto,
		GitHub:     githubMCPClient,
		Kubernetes: k8sClient,
		FlowEvents: flowEvents,
	})
	if err != nil {
		return fmt.Errorf("init runstatus domain service: %w", err)
	}
	runtimeDeployRolloutTimeout, err := time.ParseDuration(cfg.RuntimeDeployRolloutTimeout)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_RUNTIME_DEPLOY_ROLLOUT_TIMEOUT=%q: %w", cfg.RuntimeDeployRolloutTimeout, err)
	}
	runtimeDeployKanikoTimeout, err := time.ParseDuration(cfg.RuntimeDeployKanikoTimeout)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_RUNTIME_DEPLOY_KANIKO_TIMEOUT=%q: %w", cfg.RuntimeDeployKanikoTimeout, err)
	}
	runtimeDeployWaitPollInterval, err := time.ParseDuration(cfg.RuntimeDeployWaitPollInterval)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_RUNTIME_DEPLOY_WAIT_POLL_INTERVAL=%q: %w", cfg.RuntimeDeployWaitPollInterval, err)
	}
	registryHTTPTimeout, err := time.ParseDuration(cfg.RegistryHTTPTimeout)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_REGISTRY_HTTP_TIMEOUT=%q: %w", cfg.RegistryHTTPTimeout, err)
	}
	runtimeDeployReconcileInterval, err := time.ParseDuration(cfg.RuntimeDeployReconcileInterval)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_RUNTIME_DEPLOY_RECONCILE_INTERVAL=%q: %w", cfg.RuntimeDeployReconcileInterval, err)
	}
	runtimeDeployLeaseTTL, err := time.ParseDuration(cfg.RuntimeDeployLeaseTTL)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_RUNTIME_DEPLOY_LEASE_TTL=%q: %w", cfg.RuntimeDeployLeaseTTL, err)
	}
	runtimeDeployWorkerID := strings.TrimSpace(cfg.RuntimeDeployWorkerID)
	if runtimeDeployWorkerID == "" {
		hostname, hostErr := os.Hostname()
		if hostErr != nil || strings.TrimSpace(hostname) == "" {
			runtimeDeployWorkerID = "runtime-deploy-control-plane"
		} else {
			runtimeDeployWorkerID = "runtime-deploy-" + strings.TrimSpace(hostname)
		}
	}
	registryScheme := strings.TrimSpace(cfg.InternalRegistryScheme)
	if registryScheme == "" {
		registryScheme = "http"
	}
	registryBaseURL := registryScheme + "://" + strings.TrimSpace(cfg.InternalRegistryHost)
	registryClient, err := registry.NewClient(registryBaseURL, registryHTTPTimeout)
	if err != nil {
		return fmt.Errorf("init registry client: %w", err)
	}
	registryImagesService, err := registryimagesdomain.NewService(registryimagesdomain.Config{
		DefaultCleanupKeepTags: cfg.RegistryCleanupKeepTags,
	}, registryClient)
	if err != nil {
		return fmt.Errorf("init registry images service: %w", err)
	}
	runtimeDeployService, err := runtimedeploydomain.NewService(runtimedeploydomain.Config{
		ServicesConfigPath:      cfg.ServicesConfigPath,
		RepositoryRoot:          cfg.RepositoryRoot,
		RolloutTimeout:          runtimeDeployRolloutTimeout,
		KanikoTimeout:           runtimeDeployKanikoTimeout,
		WaitPollInterval:        runtimeDeployWaitPollInterval,
		KanikoFieldManager:      cfg.RuntimeDeployFieldManager,
		GitHubPAT:               strings.TrimSpace(cfg.GitHubPAT),
		RegistryCleanupKeepTags: cfg.RegistryCleanupKeepTags,
		KanikoJobLogTailLines:   200,
	}, runtimedeploydomain.Dependencies{
		Kubernetes: newRuntimeDeployKubernetesAdapter(k8sClient),
		Tasks:      runtimeDeployTasks,
		Registry:   registryClient,
		Logger:     logger,
	})
	if err != nil {
		return fmt.Errorf("init runtime deploy domain service: %w", err)
	}

	learningDefault, err := cfg.LearningModeDefaultBool()
	if err != nil {
		return err
	}
	webhookRuntimeModePolicy := loadWebhookRuntimeModePolicy(cfg, logger)

	webhookService := webhook.NewService(webhook.Config{
		AgentRuns:           agentRuns,
		Agents:              agents,
		FlowEvents:          flowEvents,
		Repos:               repos,
		Projects:            projects,
		Users:               users,
		Members:             members,
		RunStatus:           runStatusService,
		LearningModeDefault: learningDefault,
		TriggerLabels:       buildWebhookTriggerLabels(cfg),
		RuntimeModePolicy:   webhookRuntimeModePolicy,
		PlatformNamespace:   strings.TrimSpace(cfg.PlatformNamespace),
	})

	webhookURL := strings.TrimSpace(cfg.GitHubWebhookURL)
	if webhookURL == "" {
		webhookURL = strings.TrimRight(cfg.PublicBaseURL, "/") + "/api/v1/webhooks/github"
	}
	events := make([]string, 0, len(cfg.GitHubWebhookEvents))
	for _, event := range cfg.GitHubWebhookEvents {
		event = strings.TrimSpace(event)
		if event == "" {
			continue
		}
		events = append(events, event)
	}

	bootstrapSeed, err := seedBootstrapProjectsAndRepositories(runCtx, seedBootstrapProjectsAndRepositoriesParams{
		GitHubRepo:             strings.TrimSpace(cfg.GitHubRepo),
		FirstProjectGitHubRepo: strings.TrimSpace(cfg.FirstProjectGitHubRepo),
		LearningModeDefault:    learningDefault,
		GitHubPAT:              strings.TrimSpace(cfg.GitHubPAT),
		TokenCrypt:             tokenCrypto,
		PlatformTokens:         platformTokens,
		Projects:               projects,
		Repos:                  repos,
		GitHub:                 githubRepoProvider,
		Logger:                 logger,
	})
	if err != nil {
		return fmt.Errorf("seed bootstrap projects and repositories: %w", err)
	}
	staffService := staff.NewService(staff.Config{
		LearningModeDefault: learningDefault,
		WebhookSpec: repoprovider.WebhookSpec{
			URL:    webhookURL,
			Secret: cfg.GitHubWebhookSecret,
			Events: events,
		},
		ProtectedProjectIDs:    bootstrapSeed.ProtectedProjectIDs,
		ProtectedRepositoryIDs: bootstrapSeed.ProtectedRepositoryIDs,
	}, users, projects, members, repos, projectTokens, configEntries, feedback, runs, runtimeDeployTasks, registryImagesService, k8sClient, tokenCrypto, platformTokens, githubRepoProvider, githubMgmtClient, runStatusService)

	// Ensure bootstrap users exist so that the first login can be matched by email.
	if _, err := users.EnsureOwner(runCtx, cfg.BootstrapOwnerEmail); err != nil {
		return fmt.Errorf("ensure bootstrap owner user: %w", err)
	}
	if err := ensureBootstrapAllowedUsers(runCtx, users, cfg.BootstrapOwnerEmail, cfg.BootstrapAllowedEmails, logger); err != nil {
		return fmt.Errorf("ensure bootstrap allowed users: %w", err)
	}
	if err := ensureBootstrapPlatformAdmins(runCtx, users, cfg.BootstrapOwnerEmail, cfg.BootstrapPlatformAdminEmails, logger); err != nil {
		return fmt.Errorf("ensure bootstrap platform admins: %w", err)
	}

	grpcLis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen grpc %s: %w", cfg.GRPCAddr, err)
	}
	defer func() { _ = grpcLis.Close() }()

	agentCallbackService := agentcallbackdomain.NewService(agentSessions, flowEvents, agentRuns)
	if err := startRunAgentLogsCleanupLoop(runCtx, agentCallbackService, logger, cfg.RunAgentLogsRetentionDays); err != nil {
		return fmt.Errorf("start run agent logs cleanup loop: %w", err)
	}
	if err := startRuntimeDeployReconcilerLoop(runCtx, runtimeDeployService, logger, runtimeDeployWorkerID, runtimeDeployReconcileInterval, runtimeDeployLeaseTTL); err != nil {
		return fmt.Errorf("start runtime deploy reconciler loop: %w", err)
	}

	grpcServer := grpc.NewServer()
	controlplanev1.RegisterControlPlaneServiceServer(grpcServer, grpctransport.NewServer(grpctransport.Dependencies{
		Webhook:        webhookService,
		Staff:          staffService,
		Users:          users,
		AgentCallbacks: agentCallbackService,
		RunStatus:      runStatusService,
		RuntimeDeploy:  runtimeDeployService,
		MCP:            mcpService,
		Logger:         logger,
	}))

	mcpHandler := mcptransport.NewHandler(mcpService, logger)
	httpMux := http.NewServeMux()
	httpMux.Handle("/mcp", mcpHandler)
	httpMux.Handle("/mcp/", mcpHandler)
	httpMux.Handle("/metrics", promhttp.Handler())
	httpMux.HandleFunc("/health/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	httpMux.HandleFunc("/health/livez", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("alive"))
	})
	// Backward compatibility for existing probes patterns.
	httpMux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	httpMux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("alive"))
	})

	httpServer := &http.Server{Addr: cfg.HTTPAddr, Handler: httpMux}

	errCh := make(chan error, 2)
	go func() {
		logger.Info("control-plane grpc started", "addr", cfg.GRPCAddr)
		errCh <- grpcServer.Serve(grpcLis)
	}()
	go func() {
		logger.Info("control-plane http started", "addr", cfg.HTTPAddr)
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-runCtx.Done():
		logger.Info("shutting down control-plane")

		grpcServer.GracefulStop()

		shutdownCtx, cancel := context.WithTimeout(appCtx, 15*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown control-plane http: %w", err)
		}
		return nil
	case err := <-errCh:
		if err == nil {
			return nil
		}
		if err == http.ErrServerClosed {
			return nil
		}
		return fmt.Errorf("control-plane server failed: %w", err)
	}
}
