package runtimedeploy

import (
	"context"
	"fmt"
	"time"
)

// CleanupTaskLogs clears logs payloads for runtime deploy tasks older than cutoff.
func (s *Service) CleanupTaskLogs(ctx context.Context, updatedBefore time.Time) (int64, error) {
	if s == nil || s.tasks == nil {
		return 0, fmt.Errorf("runtime deploy task repository is not configured")
	}
	cutoff := updatedBefore.UTC()
	if cutoff.IsZero() {
		return 0, fmt.Errorf("updated_before is required")
	}
	return s.tasks.CleanupTaskLogsUpdatedBefore(ctx, cutoff)
}
