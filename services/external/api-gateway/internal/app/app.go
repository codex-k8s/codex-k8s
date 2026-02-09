package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/codex-k8s/codex-k8s/libs/go/crypto/tokencrypt"
	repoprovider "github.com/codex-k8s/codex-k8s/libs/go/repo/provider"
	githubprovider "github.com/codex-k8s/codex-k8s/libs/go/repo/provider/github"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/auth"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/staff"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/webhook"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/repository/postgres/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/repository/postgres/flowevent"
	learningfeedbackrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/repository/postgres/learningfeedback"
	projectrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/repository/postgres/project"
	projectmemberrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/repository/postgres/projectmember"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/repository/postgres/repocfg"
	staffrunrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/repository/postgres/staffrun"
	userrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/repository/postgres/user"
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

	users := userrepo.NewRepository(db)
	projects := projectrepo.NewRepository(db)
	members := projectmemberrepo.NewRepository(db)
	runs := staffrunrepo.NewRepository(db)
	repos := repocfgrepo.NewRepository(db)
	feedback := learningfeedbackrepo.NewRepository(db)

	webhookService := webhook.NewService(agentRuns, flowEvents, repos, projects, users, members)

	tokenCrypto, err := tokencrypt.NewService(cfg.TokenEncryptionKey)
	if err != nil {
		return fmt.Errorf("init token encryption: %w", err)
	}
	githubRepoProvider := githubprovider.NewProvider(nil)

	learningDefault := false
	if strings.TrimSpace(cfg.LearningModeDefault) != "" {
		v, err := strconv.ParseBool(cfg.LearningModeDefault)
		if err != nil {
			return fmt.Errorf("parse CODEXK8S_LEARNING_MODE_DEFAULT=%q: %w", cfg.LearningModeDefault, err)
		}
		learningDefault = v
	}

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

	// Ensure the bootstrap owner exists so that the first GitHub OAuth login can be matched by email.
	if _, err := users.EnsureOwner(context.Background(), cfg.BootstrapOwnerEmail); err != nil {
		return fmt.Errorf("ensure bootstrap owner user: %w", err)
	}
	// Optionally pre-provision additional staff emails into DB to avoid "email is not allowed"
	// on first login.
	if err := ensureBootstrapAllowedUsers(context.Background(), users, cfg.BootstrapOwnerEmail, cfg.BootstrapAllowedEmails, logger); err != nil {
		return fmt.Errorf("ensure bootstrap allowed users: %w", err)
	}
	// Optionally pre-provision additional platform admins ("owners") into DB.
	if err := ensureBootstrapPlatformAdmins(context.Background(), users, cfg.BootstrapOwnerEmail, cfg.BootstrapPlatformAdminEmails, logger); err != nil {
		return fmt.Errorf("ensure bootstrap platform admins: %w", err)
	}

	jwtTTL, err := time.ParseDuration(cfg.JWTTTL)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_JWT_TTL=%q: %w", cfg.JWTTTL, err)
	}
	authService, err := auth.NewService(auth.Config{
		PublicBaseURL:           cfg.PublicBaseURL,
		GitHubOAuthClientID:     cfg.GitHubOAuthClientID,
		GitHubOAuthClientSecret: cfg.GitHubOAuthClientSecret,
		JWTSigningKey:           []byte(cfg.JWTSigningKey),
		JWTTTL:                  jwtTTL,
		CookieSecure:            cfg.CookieSecure,
	}, users)
	if err != nil {
		return fmt.Errorf("init auth service: %w", err)
	}

	server := httptransport.NewServer(httptransport.ServerConfig{
		HTTPAddr:            cfg.HTTPAddr,
		GitHubWebhookSecret: cfg.GitHubWebhookSecret,
		MaxBodyBytes:        cfg.WebhookMaxBodyBytes,
		CookieSecure:        cfg.CookieSecure,
		StaticDir:           "/app/web",
		ViteDevUpstream:     cfg.ViteDevUpstream,
	}, webhookService, authService, users, staffService, logger)

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
