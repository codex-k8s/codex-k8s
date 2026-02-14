package runtimedeploy

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
)

const (
	defaultServicesConfigPath = "services.yaml"
	defaultRepositoryRoot     = "."
	defaultRolloutTimeout     = 20 * time.Minute
	defaultKanikoTimeout      = 30 * time.Minute
	defaultWaitPollInterval   = 2 * time.Second
	defaultFieldManager       = "codex-k8s-control-plane"
)

var placeholderPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)
var imageTagSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

// Service prepares runtime environments from services.yaml contract.
type Service struct {
	cfg   Config
	k8s   KubernetesClient
	tasks runtimedeploytaskrepo.Repository
}

// NewService creates runtime deployment service.
func NewService(cfg Config, deps Dependencies) (*Service, error) {
	if deps.Kubernetes == nil {
		return nil, fmt.Errorf("kubernetes client is required")
	}
	if deps.Tasks == nil {
		return nil, fmt.Errorf("runtime deploy task repository is required")
	}
	cfg.ServicesConfigPath = strings.TrimSpace(cfg.ServicesConfigPath)
	if cfg.ServicesConfigPath == "" {
		cfg.ServicesConfigPath = defaultServicesConfigPath
	}
	cfg.RepositoryRoot = strings.TrimSpace(cfg.RepositoryRoot)
	if cfg.RepositoryRoot == "" {
		cfg.RepositoryRoot = defaultRepositoryRoot
	}
	cfg.KanikoFieldManager = strings.TrimSpace(cfg.KanikoFieldManager)
	if cfg.KanikoFieldManager == "" {
		cfg.KanikoFieldManager = defaultFieldManager
	}
	if cfg.RolloutTimeout <= 0 {
		cfg.RolloutTimeout = defaultRolloutTimeout
	}
	if cfg.KanikoTimeout <= 0 {
		cfg.KanikoTimeout = defaultKanikoTimeout
	}
	if cfg.WaitPollInterval <= 0 {
		cfg.WaitPollInterval = defaultWaitPollInterval
	}
	return &Service{
		cfg:   cfg,
		k8s:   deps.Kubernetes,
		tasks: deps.Tasks,
	}, nil
}
