package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/controlplane"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/auth"
	httptransport "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http"
)

// Run starts api-gateway and blocks until shutdown or fatal server error.
func Run() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dialCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	server := httptransport.NewServer(httptransport.ServerConfig{
		HTTPAddr:            cfg.HTTPAddr,
		GitHubWebhookSecret: cfg.GitHubWebhookSecret,
		MaxBodyBytes:        cfg.WebhookMaxBodyBytes,
		CookieSecure:        cfg.CookieSecure,
		StaticDir:           "/app/web",
		ViteDevUpstream:     cfg.ViteDevUpstream,
	}, cp, authService, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("api-gateway started", "addr", cfg.HTTPAddr, "control_plane_grpc_target", cfg.ControlPlaneGRPCTarget)
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
