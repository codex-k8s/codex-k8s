package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config defines environment-backed runtime settings for control-plane.
type Config struct {
	// GRPCAddr is the bind address for the gRPC server.
	GRPCAddr string `env:"CODEXK8S_CONTROL_PLANE_GRPC_ADDR" envDefault:":9090"`
	// HTTPAddr is the bind address for the HTTP health/metrics server.
	HTTPAddr string `env:"CODEXK8S_CONTROL_PLANE_HTTP_ADDR" envDefault:":8081"`

	// PublicBaseURL is used to build default webhook URL when CODEXK8S_GITHUB_WEBHOOK_URL is empty.
	PublicBaseURL string `env:"CODEXK8S_PUBLIC_BASE_URL,required,notEmpty"`

	// BootstrapOwnerEmail is the first allowed email for staff access (platform admin).
	BootstrapOwnerEmail          string   `env:"CODEXK8S_BOOTSTRAP_OWNER_EMAIL,required,notEmpty"`
	BootstrapAllowedEmails       []string `env:"CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS"`
	BootstrapPlatformAdminEmails []string `env:"CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS"`

	// LearningModeDefault controls the default for newly created projects.
	// Empty string means "false".
	LearningModeDefault string `env:"CODEXK8S_LEARNING_MODE_DEFAULT" envDefault:"false"`

	// GitHubWebhookSecret is used when attaching repository hooks (staff operations).
	GitHubWebhookSecret string   `env:"CODEXK8S_GITHUB_WEBHOOK_SECRET,required,notEmpty"`
	GitHubWebhookURL    string   `env:"CODEXK8S_GITHUB_WEBHOOK_URL"`
	GitHubWebhookEvents []string `env:"CODEXK8S_GITHUB_WEBHOOK_EVENTS" envDefault:"push,pull_request,issues,issue_comment,pull_request_review,pull_request_review_comment"`

	// TokenEncryptionKey is used to encrypt/decrypt repository tokens stored in DB.
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

func (c Config) LearningModeDefaultBool() (bool, error) {
	if strings.TrimSpace(c.LearningModeDefault) == "" {
		return false, nil
	}
	v, err := strconv.ParseBool(c.LearningModeDefault)
	if err != nil {
		return false, fmt.Errorf("parse CODEXK8S_LEARNING_MODE_DEFAULT=%q: %w", c.LearningModeDefault, err)
	}
	return v, nil
}

// LoadConfig parses and validates configuration from environment variables.
func LoadConfig() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse app config from environment: %w", err)
	}
	return cfg, nil
}
