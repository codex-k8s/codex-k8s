package githubratelimit

import "time"

import valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"

func defaultNowUTC() time.Time {
	return time.Now().UTC()
}

func (s *Service) rolloutState() valuetypes.GitHubRateLimitRolloutState {
	if s != nil && s.rollout != nil {
		return s.rollout.CurrentGitHubRateLimitRolloutState()
	}
	return s.cfg.RolloutState
}

func (s *Service) capabilities() (valuetypes.GitHubRateLimitRolloutCapabilities, error) {
	return ResolveRolloutCapabilities(s.rolloutState())
}
