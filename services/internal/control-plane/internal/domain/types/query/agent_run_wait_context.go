package query

import (
	"time"

	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// AgentRunSetWaitContextParams updates typed wait linkage stored in agent_runs.
type AgentRunSetWaitContextParams struct {
	RunID          string
	WaitReason     enumtypes.AgentRunWaitReason
	WaitTargetKind enumtypes.AgentRunWaitTargetKind
	WaitTargetRef  string
	WaitDeadlineAt *time.Time
}

// AgentRunClearWaitContextParams clears wait linkage only when current wait matches expected values.
type AgentRunClearWaitContextParams struct {
	RunID          string
	WaitReason     enumtypes.AgentRunWaitReason
	WaitTargetKind enumtypes.AgentRunWaitTargetKind
	WaitTargetRef  string
}
