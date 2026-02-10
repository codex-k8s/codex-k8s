package app

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
)

// Config defines environment-backed runtime settings for worker service.
type Config struct {
	// WorkerID identifies current worker instance in logs and events.
	WorkerID string `env:"CODEXK8S_WORKER_ID"`
	// PollInterval controls tick interval for run-loop.
	PollInterval string `env:"CODEXK8S_WORKER_POLL_INTERVAL" envDefault:"5s"`
	// ClaimLimit controls how many pending runs worker claims per tick.
	ClaimLimit int `env:"CODEXK8S_WORKER_CLAIM_LIMIT" envDefault:"2"`
	// RunningCheckLimit controls how many running runs are reconciled per tick.
	RunningCheckLimit int `env:"CODEXK8S_WORKER_RUNNING_CHECK_LIMIT" envDefault:"200"`
	// SlotsPerProject defines initial slot pool size per project.
	SlotsPerProject int `env:"CODEXK8S_WORKER_SLOTS_PER_PROJECT" envDefault:"2"`
	// SlotLeaseTTL controls for how long slot is leased before expiration.
	SlotLeaseTTL string `env:"CODEXK8S_WORKER_SLOT_LEASE_TTL" envDefault:"10m"`

	// LearningModeDefault controls default project learning-mode when worker auto-creates projects.
	// Keep empty value to disable by default; set to "true" to enable by default.
	LearningModeDefault string `env:"CODEXK8S_LEARNING_MODE_DEFAULT"`

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

	// KubeconfigPath is optional kubeconfig path for local development.
	KubeconfigPath string `env:"CODEXK8S_KUBECONFIG"`
	// K8sNamespace is a namespace for worker-created Jobs.
	K8sNamespace string `env:"CODEXK8S_WORKER_K8S_NAMESPACE" envDefault:"codex-k8s-ai-staging"`
	// JobImage is a container image used for spawned run Jobs.
	JobImage string `env:"CODEXK8S_WORKER_JOB_IMAGE" envDefault:"busybox:1.36"`
	// JobCommand is a shell command executed by run Jobs.
	JobCommand string `env:"CODEXK8S_WORKER_JOB_COMMAND" envDefault:"echo codex-k8s run && sleep 2"`
	// JobTTLSeconds controls ttlSecondsAfterFinished for run Jobs.
	JobTTLSeconds int32 `env:"CODEXK8S_WORKER_JOB_TTL_SECONDS" envDefault:"600"`
	// JobBackoffLimit controls Job retry attempts.
	JobBackoffLimit int32 `env:"CODEXK8S_WORKER_JOB_BACKOFF_LIMIT" envDefault:"0"`
	// JobActiveDeadlineSeconds controls max run duration before termination.
	JobActiveDeadlineSeconds int64 `env:"CODEXK8S_WORKER_JOB_ACTIVE_DEADLINE_SECONDS" envDefault:"900"`
}

// LoadConfig parses and validates worker configuration from environment.
func LoadConfig() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse worker config from environment: %w", err)
	}

	if cfg.WorkerID == "" {
		hostname, hostErr := os.Hostname()
		if hostErr != nil || hostname == "" {
			cfg.WorkerID = "worker"
		} else {
			cfg.WorkerID = hostname
		}
	}

	return cfg, nil
}
