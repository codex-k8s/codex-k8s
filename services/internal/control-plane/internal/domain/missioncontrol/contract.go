package missioncontrol

import (
	"context"

	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

type (
	WarmupRequest = querytypes.MissionControlWarmupRequest
	WarmupSummary = valuetypes.MissionControlWarmupSummary
)

// WarmupExecutor defines the owner-controlled warmup/backfill entry-point that worker wave #371 must execute.
type WarmupExecutor interface {
	RunWarmup(ctx context.Context, params WarmupRequest) (WarmupSummary, error)
}
