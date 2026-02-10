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

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	"github.com/codex-k8s/codex-k8s/libs/go/crypto/tokencrypt"
	"github.com/codex-k8s/codex-k8s/libs/go/postgres"
	repoprovider "github.com/codex-k8s/codex-k8s/libs/go/repo/provider"
	githubprovider "github.com/codex-k8s/codex-k8s/libs/go/repo/provider/github"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/staff"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/webhook"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/flowevent"
	learningfeedbackrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/learningfeedback"
	projectrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/project"
	projectmemberrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/projectmember"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/repocfg"
	staffrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/staffrun"
	userrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/user"
	grpctransport "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/transport/grpc"
)

// Run starts control-plane servers and blocks until shutdown or fatal error.
func Run() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := postgres.Open(context.Background(), postgres.OpenParams{
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

	agentRuns := agentrunrepo.NewRepository(db)
	flowEvents := floweventrepo.NewRepository(db)

	users := userrepo.NewRepository(db)
	projects := projectrepo.NewRepository(db)
	members := projectmemberrepo.NewRepository(db)
	runs := staffrunrepo.NewRepository(db)
	repos := repocfgrepo.NewRepository(db)
	feedback := learningfeedbackrepo.NewRepository(db)

	tokenCrypto, err := tokencrypt.NewService(cfg.TokenEncryptionKey)
	if err != nil {
		return fmt.Errorf("init token encryption: %w", err)
	}
	githubRepoProvider := githubprovider.NewProvider(nil)

	learningDefault, err := cfg.LearningModeDefaultBool()
	if err != nil {
		return err
	}

	webhookService := webhook.NewService(agentRuns, flowEvents, repos, projects, users, members, learningDefault)

	webhookURL := strings.TrimSpace(cfg.GitHubWebhookURL)
	if webhookURL == "" {
		webhookURL = strings.TrimRight(cfg.PublicBaseURL, "/") + "/api/v1/webhooks/github"
	}
	events := splitCSV(cfg.GitHubWebhookEvents)
	staffService := staff.NewService(staff.Config{
		LearningModeDefault: learningDefault,
		WebhookSpec: repoprovider.WebhookSpec{
			URL:    webhookURL,
			Secret: cfg.GitHubWebhookSecret,
			Events: events,
		},
	}, users, projects, members, repos, feedback, runs, tokenCrypto, githubRepoProvider)

	// Ensure bootstrap users exist so that the first login can be matched by email.
	if _, err := users.EnsureOwner(context.Background(), cfg.BootstrapOwnerEmail); err != nil {
		return fmt.Errorf("ensure bootstrap owner user: %w", err)
	}
	if err := ensureBootstrapAllowedUsers(context.Background(), users, cfg.BootstrapOwnerEmail, cfg.BootstrapAllowedEmails, logger); err != nil {
		return fmt.Errorf("ensure bootstrap allowed users: %w", err)
	}
	if err := ensureBootstrapPlatformAdmins(context.Background(), users, cfg.BootstrapOwnerEmail, cfg.BootstrapPlatformAdminEmails, logger); err != nil {
		return fmt.Errorf("ensure bootstrap platform admins: %w", err)
	}

	grpcLis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen grpc %s: %w", cfg.GRPCAddr, err)
	}
	defer func() { _ = grpcLis.Close() }()

	grpcServer := grpc.NewServer()
	controlplanev1.RegisterControlPlaneServiceServer(grpcServer, grpctransport.NewServer(grpctransport.Dependencies{
		Webhook: webhookService,
		Staff:   staffService,
		Users:   users,
		Logger:  logger,
	}))

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer stop()

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
	case <-ctx.Done():
		logger.Info("shutting down control-plane")

		grpcServer.GracefulStop()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
