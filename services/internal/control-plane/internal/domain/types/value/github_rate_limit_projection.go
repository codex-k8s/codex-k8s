package value

import "time"

// GitHubRateLimitProjectionRefreshResult describes dominant-wait linkage recomputed for one run.
type GitHubRateLimitProjectionRefreshResult struct {
	RunID          string
	OpenWaitCount  int
	DominantWaitID string
	WaitDeadlineAt *time.Time
}
