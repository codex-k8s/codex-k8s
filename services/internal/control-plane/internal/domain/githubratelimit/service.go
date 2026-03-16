package githubratelimit

import (
	"fmt"
)

// NewService constructs canonical GitHub rate-limit domain service.
func NewService(cfg Config, deps Dependencies) (*Service, error) {
	if deps.Runs == nil {
		return nil, fmt.Errorf("agent run repository is required")
	}
	if deps.Waits == nil {
		return nil, fmt.Errorf("github rate-limit wait repository is required")
	}

	return &Service{
		cfg:        cfg,
		runs:       deps.Runs,
		waits:      deps.Waits,
		flowEvents: deps.FlowEvents,
		runStatus:  deps.RunStatusRetry,
		platform:   deps.PlatformReplay,
		rollout:    deps.RolloutState,
		now:        defaultNowUTC,
	}, nil
}
