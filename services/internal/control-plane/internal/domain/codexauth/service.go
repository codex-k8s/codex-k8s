package codexauth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

const (
	defaultKubernetesSecretName = "codex-k8s-codex-auth"
	defaultKubernetesSecretKey  = "auth.json"

	defaultGitHubSecretName = "CODEXK8S_CODEX_AUTH_JSON"
)

type Config struct {
	PlatformNamespace string

	KubernetesSecretName string
	KubernetesSecretKey  string

	GitHubRepo   string
	GitHubPAT    string
	GitHubSecret string
}

type Kubernetes interface {
	GetSecretData(ctx context.Context, namespace string, name string) (map[string][]byte, bool, error)
	UpsertSecret(ctx context.Context, namespace string, secretName string, data map[string][]byte) error
}

type GitHubMgmt interface {
	EnsureEnvironment(ctx context.Context, token string, owner string, repo string, envName string) error
	UpsertEnvSecret(ctx context.Context, token string, owner string, repo string, envName string, key string, value string) error
}

type Service struct {
	cfg    Config
	k8s    Kubernetes
	github GitHubMgmt
}

func NewService(cfg Config, k8s Kubernetes, github GitHubMgmt) (*Service, error) {
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
	if strings.TrimSpace(cfg.GitHubSecret) == "" {
		cfg.GitHubSecret = defaultGitHubSecretName
	}

	if k8s == nil {
		return nil, fmt.Errorf("kubernetes client is required")
	}

	return &Service{cfg: cfg, k8s: k8s, github: github}, nil
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

	if err := s.syncGitHubSecrets(ctx, authJSON); err != nil {
		return err
	}

	return nil
}

func (s *Service) syncGitHubSecrets(ctx context.Context, authJSON []byte) error {
	if s.github == nil {
		return nil
	}
	token := strings.TrimSpace(s.cfg.GitHubPAT)
	repoFullName := strings.TrimSpace(s.cfg.GitHubRepo)
	if token == "" || repoFullName == "" {
		return nil
	}

	owner, repo, ok := strings.Cut(repoFullName, "/")
	if !ok || strings.TrimSpace(owner) == "" || strings.TrimSpace(repo) == "" {
		return fmt.Errorf("invalid CODEXK8S_GITHUB_REPO %q", repoFullName)
	}

	secretName := strings.TrimSpace(s.cfg.GitHubSecret)
	if secretName == "" {
		secretName = defaultGitHubSecretName
	}

	envs := []string{"production", "ai"}

	var wg sync.WaitGroup
	errCh := make(chan error, len(envs))
	for _, envName := range envs {
		envName := envName
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.github.EnsureEnvironment(ctx, token, owner, repo, envName); err != nil {
				errCh <- fmt.Errorf("ensure github environment %s: %w", envName, err)
				return
			}
			if err := s.github.UpsertEnvSecret(ctx, token, owner, repo, envName, secretName, string(authJSON)); err != nil {
				errCh <- fmt.Errorf("upsert github env secret %s:%s: %w", envName, secretName, err)
				return
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

