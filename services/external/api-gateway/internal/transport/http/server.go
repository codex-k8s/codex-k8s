package http

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/controlplane"
)

// ServerConfig defines HTTP transport runtime options.
type ServerConfig struct {
	// HTTPAddr is a bind address for the Echo server.
	HTTPAddr string
	// GitHubWebhookSecret is used by webhook handler to verify signatures.
	GitHubWebhookSecret string
	// MCPCallbackToken is shared token for external approver/executor callback contracts.
	// Empty value means callback auth is disabled on HTTP level.
	MCPCallbackToken string
	// MaxBodyBytes sets webhook body size limit.
	MaxBodyBytes int64
	// CookieSecure controls Secure attribute for auth cookies.
	CookieSecure bool
	// StaticDir is a directory with built staff UI (Vue) assets.
	StaticDir string
	// ViteDevUpstream enables staff UI in "vite dev server" mode.
	// When set, non-API GET/HEAD requests are reverse-proxied to this upstream.
	ViteDevUpstream string
	// OpenAPISpecPath is an optional path to OpenAPI specification.
	OpenAPISpecPath string
	// OpenAPIValidationEnabled enables request validation middleware.
	OpenAPIValidationEnabled bool
}

// Server is an HTTP transport wrapper around Echo.
type Server struct {
	echo   *echo.Echo
	server *http.Server
	addr   string
	logger *slog.Logger
}

// NewServer builds and configures HTTP routes and middleware.
func NewServer(initCtx context.Context, cfg ServerConfig, cp *controlplane.Client, auth authService, logger *slog.Logger) (*Server, error) {
	if initCtx == nil {
		return nil, fmt.Errorf("init context is nil")
	}

	e := echo.New()
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = newHTTPErrorHandler(logger)

	if cfg.OpenAPIValidationEnabled {
		validator, err := newOpenAPIRequestValidator(initCtx, cfg.OpenAPISpecPath)
		if err != nil {
			return nil, fmt.Errorf("init openapi validator: %w", err)
		}
		e.Use(validator.middleware())
		logger.Info("openapi request validation enabled", "spec_path", validator.specPath)
	}

	h := newWebhookHandler(cfg, cp)
	mcpH := newMCPCallbackHandler(cfg, cp)
	authH := newAuthHandler(auth, cfg.CookieSecure)
	staffH := newStaffHandler(cp)

	staffAuthMw := requireStaffAuth(auth, cp.ResolveStaffByEmail)

	e.GET("/readyz", readyHandler)
	e.GET("/healthz", liveHandler)
	e.GET("/health/readyz", readyHandler)
	e.GET("/health/livez", liveHandler)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.POST("/api/v1/webhooks/github", h.IngestGitHubWebhook)
	e.POST("/api/v1/mcp/approver/callback", mcpH.CallbackApprover)
	e.POST("/api/v1/mcp/executor/callback", mcpH.CallbackExecutor)

	e.GET("/api/v1/auth/github/login", authH.LoginGitHub)
	e.GET("/api/v1/auth/github/callback", authH.CallbackGitHub)
	e.POST("/api/v1/auth/logout", authH.Logout, staffAuthMw)
	e.GET("/api/v1/auth/me", authH.Me, staffAuthMw)

	staffGroup := e.Group("/api/v1/staff", staffAuthMw)
	staffGroup.GET("/projects", staffH.ListProjects)
	staffGroup.POST("/projects", staffH.UpsertProject)
	staffGroup.GET("/projects/:project_id", staffH.GetProject)
	staffGroup.DELETE("/projects/:project_id", staffH.DeleteProject)
	staffGroup.GET("/runs", staffH.ListRuns)
	staffGroup.GET("/runs/jobs", staffH.ListRunJobs)
	staffGroup.GET("/runs/waits", staffH.ListRunWaits)
	staffGroup.GET("/approvals", staffH.ListPendingApprovals)
	staffGroup.POST("/approvals/:approval_request_id/decision", staffH.ResolveApprovalDecision)
	staffGroup.GET("/runs/:run_id", staffH.GetRun)
	staffGroup.DELETE("/runs/:run_id/namespace", staffH.DeleteRunNamespace)
	staffGroup.GET("/runs/:run_id/events", staffH.ListRunEvents)
	staffGroup.GET("/runs/:run_id/logs", staffH.GetRunLogs)
	staffGroup.GET("/runs/:run_id/learning-feedback", staffH.ListRunLearningFeedback)
	staffGroup.GET("/runtime-deploy/tasks", staffH.ListRuntimeDeployTasks)
	staffGroup.GET("/runtime-deploy/tasks/:run_id", staffH.GetRuntimeDeployTask)
	staffGroup.GET("/runtime-errors", staffH.ListRuntimeErrors)
	staffGroup.POST("/runtime-errors/:runtime_error_id/viewed", staffH.MarkRuntimeErrorViewed)
	staffGroup.GET("/runtime-deploy/images", staffH.ListRegistryImages)
	staffGroup.DELETE("/runtime-deploy/images", staffH.DeleteRegistryImageTag)
	staffGroup.POST("/runtime-deploy/images/cleanup", staffH.CleanupRegistryImages)
	staffGroup.GET("/users", staffH.ListUsers)
	staffGroup.POST("/users", staffH.CreateUser)
	staffGroup.DELETE("/users/:user_id", staffH.DeleteUser)
	staffGroup.GET("/projects/:project_id/members", staffH.ListProjectMembers)
	staffGroup.POST("/projects/:project_id/members", staffH.UpsertProjectMember)
	staffGroup.DELETE("/projects/:project_id/members/:user_id", staffH.DeleteProjectMember)
	staffGroup.PUT("/projects/:project_id/members/:user_id/learning-mode", staffH.SetProjectMemberLearningModeOverride)
	staffGroup.GET("/projects/:project_id/repositories", staffH.ListProjectRepositories)
	staffGroup.POST("/projects/:project_id/repositories", staffH.UpsertProjectRepository)
	staffGroup.DELETE("/projects/:project_id/repositories/:repository_id", staffH.DeleteProjectRepository)
	staffGroup.PUT("/projects/:project_id/repositories/:repository_id/bot-params", staffH.UpsertRepositoryBotParams)
	staffGroup.POST("/projects/:project_id/repositories/:repository_id/preflight", staffH.RunRepositoryPreflight)
	staffGroup.GET("/projects/:project_id/github-tokens", staffH.GetProjectGitHubTokens)
	staffGroup.PUT("/projects/:project_id/github-tokens", staffH.UpsertProjectGitHubTokens)
	staffGroup.GET("/config-entries", staffH.ListConfigEntries)
	staffGroup.POST("/config-entries", staffH.UpsertConfigEntry)
	staffGroup.DELETE("/config-entries/:config_entry_id", staffH.DeleteConfigEntry)
	staffGroup.GET("/docset/groups", staffH.ListDocsetGroups)
	staffGroup.POST("/projects/:project_id/docset/import", staffH.ImportDocset)
	staffGroup.POST("/projects/:project_id/docset/sync", staffH.SyncDocset)

	registerStaffUI(e, cfg.StaticDir, cfg.ViteDevUpstream)

	httpServer := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: e,
	}

	return &Server{
		echo:   e,
		server: httpServer,
		addr:   cfg.HTTPAddr,
		logger: logger,
	}, nil
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

func registerStaffUI(e *echo.Echo, staticDir string, viteDevUpstream string) {
	if viteDevUpstream != "" {
		registerViteDevUI(e, viteDevUpstream)
		return
	}

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

func registerViteDevUI(e *echo.Echo, upstream string) {
	u, err := url.Parse(upstream)
	if err != nil || u.Scheme == "" || u.Host == "" {
		// Invalid config: keep API endpoints working.
		return
	}

	rp := httputil.NewSingleHostReverseProxy(u)
	origDirector := rp.Director
	rp.Director = func(r *http.Request) {
		// Preserve original Host for Vite's host allowlist check.
		// In production/dev we allow the public domain (e.g. platform.codex-k8s.dev),
		// and the UI is accessed through the gateway reverse proxy.
		origHost := r.Header.Get("X-Forwarded-Host")
		if origHost == "" {
			origHost = r.Host
		}
		if host, _, err := net.SplitHostPort(origHost); err == nil && host != "" {
			origHost = host
		}
		origDirector(r)
		if origHost != "" {
			r.Host = origHost
			if r.Header.Get("X-Forwarded-Host") == "" {
				r.Header.Set("X-Forwarded-Host", origHost)
			}
		}
	}
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		// Surface upstream failures as 502, but don't crash api-gateway.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"code":"bad_gateway","message":"ui upstream unavailable"}`))
	}

	proxy := echo.WrapHandler(rp)
	serve := func(c *echo.Context) error {
		p := c.Request().URL.Path
		if strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/metrics") || strings.HasPrefix(p, "/health") || strings.HasPrefix(p, "/readyz") || strings.HasPrefix(p, "/healthz") {
			return echo.ErrNotFound
		}
		return proxy(c)
	}

	// Support Vite module fetches and HMR websocket on arbitrary paths.
	e.GET("/", serve)
	e.HEAD("/", serve)
	e.GET("/*", serve)
	e.HEAD("/*", serve)
}
