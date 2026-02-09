package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/webhook"
)

// ServerConfig defines HTTP transport runtime options.
type ServerConfig struct {
	// HTTPAddr is a bind address for the Echo server.
	HTTPAddr string
	// GitHubWebhookSecret is used by webhook handler to verify signatures.
	GitHubWebhookSecret string
	// MaxBodyBytes sets webhook body size limit.
	MaxBodyBytes int64
}

type webhookIngress interface {
	IngestGitHubWebhook(ctx context.Context, cmd webhook.IngestCommand) (webhook.IngestResult, error)
}

// Server is an HTTP transport wrapper around Echo.
type Server struct {
	echo   *echo.Echo
	server *http.Server
	addr   string
	logger *slog.Logger
}

// NewServer builds and configures HTTP routes and middleware.
func NewServer(cfg ServerConfig, webhookService webhookIngress, logger *slog.Logger) *Server {
	e := echo.New()
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = newHTTPErrorHandler(logger)

	h := newWebhookHandler(cfg, webhookService)

	e.GET("/readyz", readyHandler)
	e.GET("/healthz", liveHandler)
	e.GET("/health/readyz", readyHandler)
	e.GET("/health/livez", liveHandler)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.POST("/api/v1/webhooks/github", h.IngestGitHubWebhook)

	httpServer := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: e,
	}

	return &Server{
		echo:   e,
		server: httpServer,
		addr:   cfg.HTTPAddr,
		logger: logger,
	}
}

// Start runs the HTTP server until shutdown or fatal error.
func (s *Server) Start() error {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("echo start: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("echo shutdown: %w", err)
	}
	return nil
}

func readyHandler(c *echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

func liveHandler(c *echo.Context) error {
	return c.String(http.StatusOK, "alive")
}
