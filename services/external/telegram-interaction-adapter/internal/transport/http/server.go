package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/codex-k8s/codex-k8s/services/external/telegram-interaction-adapter/internal/transport/http/generated"
)

const defaultMaxBodyBytes int64 = 1 << 20

// ServerConfig defines runtime options for telegram-interaction-adapter HTTP transport.
type ServerConfig struct {
	HTTPAddr     string
	MaxBodyBytes int64
	Service      adapterService
	Logger       *slog.Logger
}

// Server wraps the HTTP server lifecycle for telegram-interaction-adapter.
type Server struct {
	server *http.Server
	addr   string
	logger *slog.Logger
}

// NewServer builds the HTTP router and middleware stack.
func NewServer(cfg ServerConfig) (*Server, error) {
	if cfg.Service == nil {
		return nil, fmt.Errorf("telegram adapter service is required")
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = defaultMaxBodyBytes
	}

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /readyz", readyHandler)
	mux.HandleFunc("GET /healthz", liveHandler)
	mux.HandleFunc("GET /health/readyz", readyHandler)
	mux.HandleFunc("GET /health/livez", liveHandler)

	handler := generated.HandlerFromMux(
		newHandler(cfg.Service, cfg.MaxBodyBytes, cfg.Logger),
		mux,
	)

	return &Server{
		server: &http.Server{
			Addr:    cfg.HTTPAddr,
			Handler: handler,
		},
		addr:   cfg.HTTPAddr,
		logger: cfg.Logger,
	}, nil
}

// Start begins serving HTTP traffic.
func (s *Server) Start() error {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("start telegram adapter http server: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown telegram adapter http server: %w", err)
	}
	return nil
}

func readyHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("ok"))
}

func liveHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("alive"))
}
