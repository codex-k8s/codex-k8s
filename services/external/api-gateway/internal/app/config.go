package app

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config defines environment-backed runtime settings for api-gateway.
type Config struct {
	// HTTPAddr is the bind address for the HTTP server.
	HTTPAddr string `env:"KODEX_HTTP_ADDR" envDefault:":8080"`

	// ControlPlaneGRPCTarget is the control-plane gRPC target host:port, e.g. kodex-control-plane:9090.
	ControlPlaneGRPCTarget string `env:"KODEX_CONTROL_PLANE_GRPC_TARGET,required,notEmpty"`

	// ViteDevUpstream enables staff UI in "vite dev server" mode (dev/production).
	// When set, api-gateway will reverse-proxy non-API paths to this upstream, e.g. http://kodex-web-console:5173.
	ViteDevUpstream string `env:"KODEX_VITE_DEV_UPSTREAM"`

	// OpenAPISpecPath points to OpenAPI source file used by request validation middleware.
	// If empty, api-gateway tries default candidates.
	OpenAPISpecPath string `env:"KODEX_OPENAPI_SPEC_PATH"`
	// OpenAPIValidationEnabled toggles OpenAPI request validation middleware.
	OpenAPIValidationEnabled bool `env:"KODEX_OPENAPI_VALIDATION_ENABLED" envDefault:"true"`

	// PublicBaseURL is a public service base URL, e.g. https://platform.kodex.works.
	// Used for OAuth redirect/callback URL generation.
	PublicBaseURL string `env:"KODEX_PUBLIC_BASE_URL,required,notEmpty"`

	// GitHubOAuthClientID is GitHub OAuth App client id.
	GitHubOAuthClientID string `env:"KODEX_GITHUB_OAUTH_CLIENT_ID,required,notEmpty"`
	// GitHubOAuthClientSecret is GitHub OAuth App client secret.
	GitHubOAuthClientSecret string `env:"KODEX_GITHUB_OAUTH_CLIENT_SECRET,required,notEmpty"`

	// JWTSigningKey is the HMAC key for staff JWT tokens.
	JWTSigningKey string `env:"KODEX_JWT_SIGNING_KEY,required,notEmpty"`
	// JWTTTL is the short-lived JWT TTL duration, e.g. 15m.
	JWTTTL string `env:"KODEX_JWT_TTL" envDefault:"15m"`
	// CookieSecure controls Secure attribute for auth cookies (should be true under HTTPS).
	CookieSecure bool `env:"KODEX_COOKIE_SECURE" envDefault:"false"`

	// GitHubWebhookSecret is used to validate X-Hub-Signature-256.
	GitHubWebhookSecret string `env:"KODEX_GITHUB_WEBHOOK_SECRET,required,notEmpty"`
	// MCPCallbackToken is shared token for external approver/executor callback contracts.
	// If empty, callback endpoints work without token auth (network perimeter restrictions are expected).
	MCPCallbackToken string `env:"KODEX_MCP_CALLBACK_TOKEN"`
	// WebhookMaxBodyBytes limits accepted webhook payload size.
	WebhookMaxBodyBytes int64 `env:"KODEX_WEBHOOK_MAX_BODY_BYTES" envDefault:"1048576"`
}

// LoadConfig parses and validates configuration from environment variables.
func LoadConfig() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse app config from environment: %w", err)
	}

	return cfg, nil
}
