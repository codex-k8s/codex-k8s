package missioncontrol

import (
	"fmt"
	"time"

	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	missioncontrolrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/missioncontrol"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

// Config controls Mission Control domain service rollout gates.
type Config struct {
	RolloutState         valuetypes.MissionControlRolloutState
	DefaultTimelineLimit int
}

// Dependencies contains required collaborators for Mission Control use-cases.
type Dependencies struct {
	Repository missioncontrolrepo.Repository
	FlowEvents floweventrepo.Repository
}

// Service implements Mission Control owner-owned domain use-cases.
type Service struct {
	cfg        Config
	repository missioncontrolrepo.Repository
	flowEvents floweventrepo.Repository
	now        func() time.Time
}

// NewService constructs Mission Control domain service.
func NewService(cfg Config, deps Dependencies) (*Service, error) {
	if deps.Repository == nil {
		return nil, fmt.Errorf("mission control repository is required")
	}
	if err := ValidateRolloutState(cfg.RolloutState); err != nil {
		return nil, err
	}
	if cfg.DefaultTimelineLimit <= 0 {
		cfg.DefaultTimelineLimit = 50
	}
	return &Service{
		cfg:        cfg,
		repository: deps.Repository,
		flowEvents: deps.FlowEvents,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}, nil
}
