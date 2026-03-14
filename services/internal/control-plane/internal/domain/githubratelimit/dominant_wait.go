package githubratelimit

import (
	"slices"
	"time"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// ElectDominantWait picks the single wait that should drive run-level linkage.
func ElectDominantWait(candidates []entitytypes.GitHubRateLimitWait) (entitytypes.GitHubRateLimitWait, bool) {
	open := make([]entitytypes.GitHubRateLimitWait, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.State.IsOpen() {
			open = append(open, candidate)
		}
	}
	if len(open) == 0 {
		return entitytypes.GitHubRateLimitWait{}, false
	}

	slices.SortStableFunc(open, compareDominantWait)
	return open[0], true
}

func compareDominantWait(left entitytypes.GitHubRateLimitWait, right entitytypes.GitHubRateLimitWait) int {
	leftRank := dominantStateRank(left.State)
	rightRank := dominantStateRank(right.State)
	if leftRank != rightRank {
		return leftRank - rightRank
	}
	if cmp := compareOptionalTimeDesc(left.ResumeNotBefore, right.ResumeNotBefore); cmp != 0 {
		return cmp
	}
	if cmp := right.UpdatedAt.Compare(left.UpdatedAt); cmp != 0 {
		return cmp
	}
	if cmp := right.CreatedAt.Compare(left.CreatedAt); cmp != 0 {
		return cmp
	}
	return compareStringAsc(left.ID, right.ID)
}

func dominantStateRank(state enumtypes.GitHubRateLimitWaitState) int {
	if state == enumtypes.GitHubRateLimitWaitStateManualActionRequired {
		return 0
	}
	return 1
}

func compareOptionalTimeDesc(left *time.Time, right *time.Time) int {
	switch {
	case left != nil && right == nil:
		return -1
	case left == nil && right != nil:
		return 1
	case left == nil && right == nil:
		return 0
	default:
		return right.Compare(*left)
	}
}

func compareStringAsc(left string, right string) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}
