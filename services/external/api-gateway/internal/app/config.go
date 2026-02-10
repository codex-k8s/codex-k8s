package app

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config defines environment-backed runtime settings for api-gateway.
type Config struct {
	// HTTPAddr is the bind address for the HTTP server.
	HTTPAddr string `env:"CODEXK8S_HTTP_ADDR" envDefault:":8080"`

	// ViteDevUpstream enables staff UI in "vite dev server" mode (dev/staging).
	// When set, api-gateway will reverse-proxy non-API paths to this upstream, e.g. http://codex-k8s-web-console:5173.
	ViteDevUpstream string `env:"CODEXK8S_VITE_DEV_UPSTREAM"`

	// PublicBaseURL is a public service base URL, e.g. https://staging.codex-k8s.dev.
	// Used for OAuth redirect/callback URL generation.
	PublicBaseURL string `env:"CODEXK8S_PUBLIC_BASE_URL,required,notEmpty"`

	// BootstrapOwnerEmail is the first allowed email for GitHub OAuth login (platform admin).
	BootstrapOwnerEmail string `env:"CODEXK8S_BOOTSTRAP_OWNER_EMAIL,required,notEmpty"`

	// BootstrapAllowedEmails is an optional comma-separated list of additional staff emails
	// that should be allowed to login (pre-provisioned into DB on startup).
	//
	// Example: "dev1@example.com,dev2@example.com"
	BootstrapAllowedEmails string `env:"CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS"`

	// BootstrapPlatformAdminEmails is an optional comma-separated list of additional platform admins (owners).
	// These emails will be upserted into `users` with `is_platform_admin=true` on startup.
	//
	// Example: "owner2@example.com"
	BootstrapPlatformAdminEmails string `env:"CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS"`

	// LearningModeDefault controls the default for newly created projects.
	// Set to "true" to enable by default.
	// Default is "false".
	LearningModeDefault string `env:"CODEXK8S_LEARNING_MODE_DEFAULT" envDefault:"false"`

	// GitHubOAuthClientID is GitHub OAuth App client id.
	GitHubOAuthClientID string `env:"CODEXK8S_GITHUB_OAUTH_CLIENT_ID,required,notEmpty"`
	// GitHubOAuthClientSecret is GitHub OAuth App client secret.
	GitHubOAuthClientSecret string `env:"CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET,required,notEmpty"`

	// JWTSigningKey is the HMAC key for staff JWT tokens.
	JWTSigningKey string `env:"CODEXK8S_JWT_SIGNING_KEY,required,notEmpty"`
	// JWTTTL is the short-lived JWT TTL duration, e.g. 15m.
	JWTTTL string `env:"CODEXK8S_JWT_TTL" envDefault:"15m"`
	// CookieSecure controls Secure attribute for auth cookies (should be true under HTTPS).
	CookieSecure bool `env:"CODEXK8S_COOKIE_SECURE" envDefault:"true"`

	// GitHubWebhookSecret is used to validate X-Hub-Signature-256.
	GitHubWebhookSecret string `env:"CODEXK8S_GITHUB_WEBHOOK_SECRET,required,notEmpty"`
	// GitHubWebhookURL overrides the public webhook callback URL used when configuring repository hooks.
	// If empty, api-gateway uses PublicBaseURL + "/api/v1/webhooks/github".
	GitHubWebhookURL string `env:"CODEXK8S_GITHUB_WEBHOOK_URL"`
	// GitHubWebhookEvents is a comma-separated list of GitHub webhook events to subscribe to.
	GitHubWebhookEvents string `env:"CODEXK8S_GITHUB_WEBHOOK_EVENTS" envDefault:"push,pull_request,issues,issue_comment,pull_request_review,pull_request_review_comment"`
	// WebhookMaxBodyBytes limits accepted webhook payload size.
	WebhookMaxBodyBytes int64 `env:"CODEXK8S_WEBHOOK_MAX_BODY_BYTES" envDefault:"1048576"`

	// TokenEncryptionKey is used to encrypt/decrypt repository and agent tokens stored in DB.
	// Must be a hex-encoded 32-byte key (64 hex chars).
	TokenEncryptionKey string `env:"CODEXK8S_TOKEN_ENCRYPTION_KEY,required,notEmpty"`

	// DBHost is the PostgreSQL host.
	DBHost string `env:"CODEXK8S_DB_HOST,required,notEmpty"`
	// DBPort is the PostgreSQL port.
	DBPort int `env:"CODEXK8S_DB_PORT" envDefault:"5432"`
	// DBName is the PostgreSQL database name.
	DBName string `env:"CODEXK8S_DB_NAME,required,notEmpty"`
	// DBUser is the PostgreSQL username.
	DBUser string `env:"CODEXK8S_DB_USER,required,notEmpty"`
	// DBPassword is the PostgreSQL password.
	DBPassword string `env:"CODEXK8S_DB_PASSWORD,required,notEmpty"`
	// DBSSLMode is the PostgreSQL SSL mode.
	DBSSLMode string `env:"CODEXK8S_DB_SSLMODE" envDefault:"disable"`
}

// LoadConfig parses and validates configuration from environment variables.
func LoadConfig() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse app config from environment: %w", err)
	}

	return cfg, nil
}
