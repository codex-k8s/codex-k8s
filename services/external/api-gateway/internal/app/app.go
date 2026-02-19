package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/postgres"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/auth"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/controlplane"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/realtime"
	httptransport "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http"
)

// Run starts api-gateway and blocks until shutdown or fatal server error.
func Run() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	appCtx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dialCtx, cancel := context.WithTimeout(appCtx, 30*time.Second)
	defer cancel()
	cp, err := controlplane.Dial(dialCtx, cfg.ControlPlaneGRPCTarget)
	if err != nil {
		return err
	}
	defer func() { _ = cp.Close() }()

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
	}, cp)
	if err != nil {
		return fmt.Errorf("init auth service: %w", err)
	}

	var realtimeBackplane *realtime.Backplane
	var realtimeRepo *realtime.Repository
	var realtimeDBPoolCloser interface{ Close() }
	if cfg.RealtimeBackplaneEnabled {
		realtimeDBPool, poolErr := postgres.OpenPGXPool(appCtx, postgres.OpenParams{
			Host:        cfg.DBHost,
			Port:        cfg.DBPort,
			DBName:      cfg.DBName,
			User:        cfg.DBUser,
			Password:    cfg.DBPassword,
			SSLMode:     cfg.DBSSLMode,
			PingTimeout: 5 * time.Second,
		})
		if poolErr != nil {
			return fmt.Errorf("open realtime postgres pool: %w", poolErr)
		}
		realtimeDBPoolCloser = realtimeDBPool
		realtimeRepo = realtime.NewRepository(realtimeDBPool)

		cleanupInterval, cleanupErr := time.ParseDuration(cfg.RealtimeCleanupInterval)
		if cleanupErr != nil {
			return fmt.Errorf("parse CODEXK8S_REALTIME_CLEANUP_INTERVAL=%q: %w", cfg.RealtimeCleanupInterval, cleanupErr)
		}
		retention, retentionErr := time.ParseDuration(cfg.RealtimeRetention)
		if retentionErr != nil {
			return fmt.Errorf("parse CODEXK8S_REALTIME_RETENTION=%q: %w", cfg.RealtimeRetention, retentionErr)
		}

		realtimeBackplane = realtime.NewBackplane(realtime.Config{
			DSN:             postgres.BuildDSN(postgres.OpenParams{Host: cfg.DBHost, Port: cfg.DBPort, DBName: cfg.DBName, User: cfg.DBUser, Password: cfg.DBPassword, SSLMode: cfg.DBSSLMode}),
			Channel:         cfg.RealtimeChannel,
			CleanupInterval: cleanupInterval,
			Retention:       retention,
		}, realtimeRepo, logger)
		realtimeBackplane.Start(appCtx)
		defer realtimeBackplane.Stop()
	}
	if realtimeDBPoolCloser != nil {
		defer realtimeDBPoolCloser.Close()
	}

	server, err := httptransport.NewServer(appCtx, httptransport.ServerConfig{
		HTTPAddr:                 cfg.HTTPAddr,
		GitHubWebhookSecret:      cfg.GitHubWebhookSecret,
		MCPCallbackToken:         strings.TrimSpace(cfg.MCPCallbackToken),
		MaxBodyBytes:             cfg.WebhookMaxBodyBytes,
		CookieSecure:             cfg.CookieSecure,
		StaticDir:                "/app/web",
		ViteDevUpstream:          cfg.ViteDevUpstream,
		OpenAPISpecPath:          cfg.OpenAPISpecPath,
		OpenAPIValidationEnabled: cfg.OpenAPIValidationEnabled,
	}, cp, authService, logger, realtimeBackplane, realtimeRepo)
	if err != nil {
		return fmt.Errorf("init http server: %w", err)
	}

	ctx, stop := signal.NotifyContext(appCtx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("api-gateway started", "addr", cfg.HTTPAddr, "control_plane_grpc_target", cfg.ControlPlaneGRPCTarget)
		serverErr <- server.Start()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(appCtx, 15*time.Second)
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
