package http

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	// CookieSecure controls Secure attribute for auth cookies.
	CookieSecure bool
	// StaticDir is a directory with built staff UI (Vue) assets.
	StaticDir string
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
func NewServer(cfg ServerConfig, webhookService webhookIngress, auth authService, staffSvc staffService, logger *slog.Logger) *Server {
	e := echo.New()
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = newHTTPErrorHandler(logger)

	h := newWebhookHandler(cfg, webhookService)
	authH := newAuthHandler(auth, cfg.CookieSecure)
	staffH := newStaffHandler(staffSvc)

	e.GET("/readyz", readyHandler)
	e.GET("/healthz", liveHandler)
	e.GET("/health/readyz", readyHandler)
	e.GET("/health/livez", liveHandler)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.POST("/api/v1/webhooks/github", h.IngestGitHubWebhook)

	e.GET("/api/v1/auth/github/login", authH.LoginGitHub)
	e.GET("/api/v1/auth/github/callback", authH.CallbackGitHub)
	e.POST("/api/v1/auth/logout", authH.Logout, requireStaffAuth(auth))
	e.GET("/api/v1/auth/me", authH.Me, requireStaffAuth(auth))

	staffGroup := e.Group("/api/v1/staff", requireStaffAuth(auth))
	staffGroup.GET("/projects", staffH.ListProjects)
	staffGroup.GET("/runs", staffH.ListRuns)
	staffGroup.GET("/runs/:run_id/events", staffH.ListRunEvents)
	staffGroup.GET("/users", staffH.ListUsers)
	staffGroup.POST("/users", staffH.CreateUser)
	staffGroup.GET("/projects/:project_id/members", staffH.ListProjectMembers)
	staffGroup.POST("/projects/:project_id/members", staffH.UpsertProjectMember)

	registerStaticUI(e, cfg.StaticDir)

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

func registerStaticUI(e *echo.Echo, staticDir string) {
	if staticDir == "" {
		staticDir = "/app/web"
	}
	if st, err := os.Stat(staticDir); err != nil || !st.IsDir() {
		// In local/dev runs the folder may be missing; keep API endpoints working.
		return
	}

	webFS := os.DirFS(staticDir)
	assetsFS := os.DirFS(filepath.Join(staticDir, "assets"))

	// Serve known static assets directly.
	e.GET("/assets/*", func(c *echo.Context) error {
		assetPath := c.Param("*")
		if assetPath == "" {
			return echo.ErrNotFound
		}
		// Avoid path traversal outside staticDir/assets.
		clean := filepath.Clean(assetPath)
		if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") || strings.Contains(clean, string(filepath.Separator)+".."+string(filepath.Separator)) {
			return echo.ErrNotFound
		}
		return c.FileFS(clean, assetsFS)
	})

	// SPA fallback: for any GET not under /api, serve index.html.
	serveSPA := func(c *echo.Context) error {
		reqPath := c.Request().URL.Path
		if strings.HasPrefix(reqPath, "/api/") || strings.HasPrefix(reqPath, "/metrics") || strings.HasPrefix(reqPath, "/health") || strings.HasPrefix(reqPath, "/readyz") || strings.HasPrefix(reqPath, "/healthz") {
			return echo.ErrNotFound
		}
		if strings.HasPrefix(reqPath, "/assets/") {
			return echo.ErrNotFound
		}
		// Attempt to serve an existing file first.
		clean := filepath.Clean(strings.TrimPrefix(reqPath, "/"))
		if clean != "." && clean != "/" {
			if _, err := fs.Stat(webFS, clean); err == nil {
				return c.FileFS(clean, webFS)
			}
		}
		c.Response().Header().Set("Cache-Control", "no-store")
		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		return c.FileFS("index.html", webFS)
	}

	// Echo wildcard routes (`/*`) do not match the bare root `/`, so register both.
	e.GET("/", serveSPA)
	e.GET("/*", serveSPA)
}
