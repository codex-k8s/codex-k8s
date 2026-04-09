package githubratelimit

import (
	"testing"
	"time"

	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

func TestElectDominantWait_PrefersManualActionRequired(t *testing.T) {
	now := time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC)
	later := now.Add(15 * time.Minute)

	selected, found := ElectDominantWait([]entitytypes.GitHubRateLimitWait{
		{
			ID:              "resume",
			State:           enumtypes.GitHubRateLimitWaitStateOpen,
			ResumeNotBefore: &later,
			UpdatedAt:       now,
			CreatedAt:       now,
		},
		{
			ID:        "manual",
			State:     enumtypes.GitHubRateLimitWaitStateManualActionRequired,
			UpdatedAt: now.Add(-time.Minute),
			CreatedAt: now.Add(-time.Minute),
		},
	})
	if !found {
		t.Fatalf("expected dominant wait to be selected")
	}
	if selected.ID != "manual" {
		t.Fatalf("dominant wait id = %q, want %q", selected.ID, "manual")
	}
}

func TestElectDominantWait_PrefersLaterResumeDeadline(t *testing.T) {
	now := time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC)
	earlier := now.Add(2 * time.Minute)
	later := now.Add(5 * time.Minute)

	selected, found := ElectDominantWait([]entitytypes.GitHubRateLimitWait{
		{
			ID:              "earlier",
			State:           enumtypes.GitHubRateLimitWaitStateOpen,
			ResumeNotBefore: &earlier,
			UpdatedAt:       now.Add(time.Minute),
			CreatedAt:       now,
		},
		{
			ID:              "later",
			State:           enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled,
			ResumeNotBefore: &later,
			UpdatedAt:       now,
			CreatedAt:       now,
		},
	})
	if !found {
		t.Fatalf("expected dominant wait to be selected")
	}
	if selected.ID != "later" {
		t.Fatalf("dominant wait id = %q, want %q", selected.ID, "later")
	}
}

func TestElectDominantWait_FallsBackToLatestUpdate(t *testing.T) {
	now := time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC)

	selected, found := ElectDominantWait([]entitytypes.GitHubRateLimitWait{
		{
			ID:        "older",
			State:     enumtypes.GitHubRateLimitWaitStateOpen,
			UpdatedAt: now,
			CreatedAt: now,
		},
		{
			ID:        "newer",
			State:     enumtypes.GitHubRateLimitWaitStateAutoResumeInProgress,
			UpdatedAt: now.Add(time.Minute),
			CreatedAt: now,
		},
	})
	if !found {
		t.Fatalf("expected dominant wait to be selected")
	}
	if selected.ID != "newer" {
		t.Fatalf("dominant wait id = %q, want %q", selected.ID, "newer")
	}
}

func TestElectDominantWait_ReturnsFalseForNoOpenWaits(t *testing.T) {
	selected, found := ElectDominantWait([]entitytypes.GitHubRateLimitWait{
		{ID: "resolved", State: enumtypes.GitHubRateLimitWaitStateResolved},
		{ID: "cancelled", State: enumtypes.GitHubRateLimitWaitStateCancelled},
	})
	if found {
		t.Fatalf("expected no dominant wait, got %q", selected.ID)
	}
}
