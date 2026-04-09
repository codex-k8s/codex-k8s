package value

import (
	"time"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

// GitHubRateLimitProjectionRefreshResult describes dominant-wait linkage recomputed for one run.
type GitHubRateLimitProjectionRefreshResult struct {
	RunID          string
	OpenWaitCount  int
	DominantWaitID string
	WaitDeadlineAt *time.Time
	SyncState      enumtypes.GitHubRateLimitProjectionSyncState
}
