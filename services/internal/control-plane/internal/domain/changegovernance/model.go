package changegovernance

import (
	"fmt"

	domainrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/changegovernance"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

// Config controls rollout gates for change-governance foundation.
type Config struct {
	RolloutState valuetypes.ChangeGovernanceRolloutState
}

// Dependencies contains persistence collaborators required by the service.
type Dependencies struct {
	Repository   domainrepo.Repository
	RolloutState rolloutStateProvider
}

type rolloutStateProvider interface {
	CurrentChangeGovernanceRolloutState() valuetypes.ChangeGovernanceRolloutState
}

// Service owns canonical change-governance semantics in control-plane.
type Service struct {
	cfg     Config
	repo    domainrepo.Repository
	rollout rolloutStateProvider
}

// NewService constructs change-governance foundation service.
func NewService(cfg Config, deps Dependencies) (*Service, error) {
	if deps.Repository == nil {
		return nil, fmt.Errorf("change governance repository is required")
	}
	return &Service{
		cfg:     cfg,
		repo:    deps.Repository,
		rollout: deps.RolloutState,
	}, nil
}
