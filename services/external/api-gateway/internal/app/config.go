package app

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config defines environment-backed runtime settings for api-gateway.
type Config struct {
	// HTTPAddr is the bind address for the HTTP server.
	HTTPAddr string `env:"CODEXK8S_HTTP_ADDR" envDefault:":8080"`

	// GitHubWebhookSecret is used to validate X-Hub-Signature-256.
	GitHubWebhookSecret string `env:"CODEXK8S_GITHUB_WEBHOOK_SECRET,required,notEmpty"`
	// WebhookMaxBodyBytes limits accepted webhook payload size.
	WebhookMaxBodyBytes int64 `env:"CODEXK8S_WEBHOOK_MAX_BODY_BYTES" envDefault:"1048576"`

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
