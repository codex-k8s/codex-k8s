package systemsettings

import (
	"context"
	"fmt"
	"sync"

	sharedsystemsettings "github.com/codex-k8s/codex-k8s/libs/go/systemsettings"
)

type repository interface {
	GetBoolean(ctx context.Context, key string) (bool, bool, error)
}

// Service caches worker-visible runtime settings and refreshes them from PostgreSQL.
type Service struct {
	repo repository

	mu                         sync.RWMutex
	githubRateLimitWaitEnabled bool
}

func NewService(repo repository) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("worker system settings repository is required")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) RefreshCache(ctx context.Context) error {
	value, found, err := s.repo.GetBoolean(ctx, sharedsystemsettings.GitHubRateLimitWaitEnabledKey)
	if err != nil {
		return err
	}

	s.mu.Lock()
	if found {
		s.githubRateLimitWaitEnabled = value
	} else {
		s.githubRateLimitWaitEnabled = false
	}
	s.mu.Unlock()
	return nil
}

func (s *Service) GitHubRateLimitWaitEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.githubRateLimitWaitEnabled
}
