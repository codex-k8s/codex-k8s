package app

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config defines environment-backed runtime settings for api-gateway.
type Config struct {
	// HTTPAddr is the bind address for the HTTP server.
	HTTPAddr string `env:"CODEXK8S_HTTP_ADDR" envDefault:":8080"`

	// ControlPlaneGRPCTarget is the control-plane gRPC target host:port, e.g. codex-k8s-control-plane:9090.
	ControlPlaneGRPCTarget string `env:"CODEXK8S_CONTROL_PLANE_GRPC_TARGET,required,notEmpty"`

	// ViteDevUpstream enables staff UI in "vite dev server" mode (dev/production).
	// When set, api-gateway will reverse-proxy non-API paths to this upstream, e.g. http://codex-k8s-web-console:5173.
	ViteDevUpstream string `env:"CODEXK8S_VITE_DEV_UPSTREAM"`

	// OpenAPISpecPath points to OpenAPI source file used by request validation middleware.
	// If empty, api-gateway tries default candidates.
	OpenAPISpecPath string `env:"CODEXK8S_OPENAPI_SPEC_PATH"`
	// OpenAPIValidationEnabled toggles OpenAPI request validation middleware.
	OpenAPIValidationEnabled bool `env:"CODEXK8S_OPENAPI_VALIDATION_ENABLED" envDefault:"true"`

	// PublicBaseURL is a public service base URL, e.g. https://platform.codex-k8s.dev.
	// Used for OAuth redirect/callback URL generation.
	PublicBaseURL string `env:"CODEXK8S_PUBLIC_BASE_URL,required,notEmpty"`

	// GitHubOAuthClientID is GitHub OAuth App client id.
	GitHubOAuthClientID string `env:"CODEXK8S_GITHUB_OAUTH_CLIENT_ID,required,notEmpty"`
	// GitHubOAuthClientSecret is GitHub OAuth App client secret.
	GitHubOAuthClientSecret string `env:"CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET,required,notEmpty"`

	// JWTSigningKey is the HMAC key for staff JWT tokens.
	JWTSigningKey string `env:"CODEXK8S_JWT_SIGNING_KEY,required,notEmpty"`
	// JWTTTL is the short-lived JWT TTL duration, e.g. 15m.
	JWTTTL string `env:"CODEXK8S_JWT_TTL" envDefault:"15m"`
	// CookieSecure controls Secure attribute for auth cookies (should be true under HTTPS).
	CookieSecure bool `env:"CODEXK8S_COOKIE_SECURE" envDefault:"false"`

	// GitHubWebhookSecret is used to validate X-Hub-Signature-256.
	GitHubWebhookSecret string `env:"CODEXK8S_GITHUB_WEBHOOK_SECRET,required,notEmpty"`
	// MCPCallbackToken is shared token for external approver/executor callback contracts.
	// If empty, callback endpoints work without token auth (network perimeter restrictions are expected).
	MCPCallbackToken string `env:"CODEXK8S_MCP_CALLBACK_TOKEN"`
	// WebhookMaxBodyBytes limits accepted webhook payload size.
	WebhookMaxBodyBytes int64 `env:"CODEXK8S_WEBHOOK_MAX_BODY_BYTES" envDefault:"1048576"`

	// DBHost is PostgreSQL host for realtime event backplane.
	DBHost string `env:"CODEXK8S_DB_HOST"`
	// DBPort is PostgreSQL port for realtime event backplane.
	DBPort int `env:"CODEXK8S_DB_PORT" envDefault:"5432"`
	// DBName is PostgreSQL database name for realtime event backplane.
	DBName string `env:"CODEXK8S_DB_NAME"`
	// DBUser is PostgreSQL username for realtime event backplane.
	DBUser string `env:"CODEXK8S_DB_USER"`
	// DBPassword is PostgreSQL password for realtime event backplane.
	DBPassword string `env:"CODEXK8S_DB_PASSWORD"`
	// DBSSLMode is PostgreSQL SSL mode for realtime event backplane.
	DBSSLMode string `env:"CODEXK8S_DB_SSLMODE" envDefault:"disable"`

	// RealtimeBackplaneEnabled enables LISTEN/NOTIFY websocket backplane.
	RealtimeBackplaneEnabled bool `env:"CODEXK8S_REALTIME_BACKPLANE_ENABLED" envDefault:"true"`
	// RealtimeChannel is PostgreSQL LISTEN/NOTIFY channel name.
	RealtimeChannel string `env:"CODEXK8S_REALTIME_CHANNEL" envDefault:"codex_realtime"`
	// RealtimeCleanupInterval controls cleanup loop frequency for realtime_events table.
	RealtimeCleanupInterval string `env:"CODEXK8S_REALTIME_CLEANUP_INTERVAL" envDefault:"10m"`
	// RealtimeRetention controls how long realtime_events rows are retained.
	RealtimeRetention string `env:"CODEXK8S_REALTIME_RETENTION" envDefault:"72h"`
}

// LoadConfig parses and validates configuration from environment variables.
func LoadConfig() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse app config from environment: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if !c.RealtimeBackplaneEnabled {
		return nil
	}
	if strings.TrimSpace(c.DBHost) == "" {
		return fmt.Errorf("parse app config from environment: env: required environment variable %q is not set", "CODEXK8S_DB_HOST")
	}
	if strings.TrimSpace(c.DBName) == "" {
		return fmt.Errorf("parse app config from environment: env: required environment variable %q is not set", "CODEXK8S_DB_NAME")
	}
	if strings.TrimSpace(c.DBUser) == "" {
		return fmt.Errorf("parse app config from environment: env: required environment variable %q is not set", "CODEXK8S_DB_USER")
	}
	if strings.TrimSpace(c.DBPassword) == "" {
		return fmt.Errorf("parse app config from environment: env: required environment variable %q is not set", "CODEXK8S_DB_PASSWORD")
	}
	return nil
}
