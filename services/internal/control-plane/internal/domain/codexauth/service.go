package codexauth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	defaultKubernetesSecretName = "kodex-codex-auth"
	defaultKubernetesSecretKey  = "auth.json"
)

type Config struct {
	PlatformNamespace string

	KubernetesSecretName string
	KubernetesSecretKey  string
}

type Kubernetes interface {
	GetSecretData(ctx context.Context, namespace string, name string) (map[string][]byte, bool, error)
	UpsertSecret(ctx context.Context, namespace string, secretName string, data map[string][]byte) error
}

type Service struct {
	cfg Config
	k8s Kubernetes
}

func NewService(cfg Config, k8s Kubernetes) (*Service, error) {
	cfg.PlatformNamespace = strings.TrimSpace(cfg.PlatformNamespace)
	if cfg.PlatformNamespace == "" {
		return nil, fmt.Errorf("platform namespace is required")
	}

	if strings.TrimSpace(cfg.KubernetesSecretName) == "" {
		cfg.KubernetesSecretName = defaultKubernetesSecretName
	}
	if strings.TrimSpace(cfg.KubernetesSecretKey) == "" {
		cfg.KubernetesSecretKey = defaultKubernetesSecretKey
	}

	if k8s == nil {
		return nil, fmt.Errorf("kubernetes client is required")
	}

	return &Service{cfg: cfg, k8s: k8s}, nil
}

func (s *Service) Get(ctx context.Context) ([]byte, bool, error) {
	if s == nil {
		return nil, false, fmt.Errorf("codex auth service is nil")
	}
	data, found, err := s.k8s.GetSecretData(ctx, s.cfg.PlatformNamespace, s.cfg.KubernetesSecretName)
	if err != nil {
		return nil, false, fmt.Errorf("get kubernetes secret %s/%s: %w", s.cfg.PlatformNamespace, s.cfg.KubernetesSecretName, err)
	}
	if !found || len(data) == 0 {
		return nil, false, nil
	}
	key := strings.TrimSpace(s.cfg.KubernetesSecretKey)
	raw := data[key]
	if len(raw) == 0 {
		return nil, false, nil
	}
	out := append([]byte(nil), raw...)
	return out, true, nil
}

func (s *Service) Upsert(ctx context.Context, authJSON []byte) error {
	if s == nil {
		return fmt.Errorf("codex auth service is nil")
	}
	authJSON = []byte(strings.TrimSpace(string(authJSON)))
	if len(authJSON) == 0 {
		return fmt.Errorf("auth json is required")
	}
	if !json.Valid(authJSON) {
		return fmt.Errorf("auth json is invalid")
	}

	secretData := map[string][]byte{
		strings.TrimSpace(s.cfg.KubernetesSecretKey): authJSON,
	}
	if err := s.k8s.UpsertSecret(ctx, s.cfg.PlatformNamespace, s.cfg.KubernetesSecretName, secretData); err != nil {
		return fmt.Errorf("upsert kubernetes secret %s/%s: %w", s.cfg.PlatformNamespace, s.cfg.KubernetesSecretName, err)
	}
	return nil
}
