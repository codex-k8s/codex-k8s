package http

import "testing"

func TestMissionControlResumeToken_RoundTrip(t *testing.T) {
	t.Parallel()

	token, err := encodeMissionControlResumeToken(missionControlResumeTokenPayload{
		SnapshotID:  "snapshot-1",
		ViewMode:    "graph",
		StatePreset: "working",
		Search:      "owner",
		Cursor:      "cursor-1",
		RootLimit:   25,
		IssuedAt:    "2026-03-14T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("encodeMissionControlResumeToken() error = %v", err)
	}

	payload, err := decodeMissionControlResumeToken(token)
	if err != nil {
		t.Fatalf("decodeMissionControlResumeToken() error = %v", err)
	}
	if payload.SnapshotID != "snapshot-1" {
		t.Fatalf("expected snapshot_id %q, got %q", "snapshot-1", payload.SnapshotID)
	}
	if payload.ViewMode != "graph" || payload.StatePreset != "working" {
		t.Fatalf("expected scope to round-trip, got view_mode=%q state_preset=%q", payload.ViewMode, payload.StatePreset)
	}
	if payload.RootLimit != 25 {
		t.Fatalf("expected root limit 25, got %d", payload.RootLimit)
	}
}

func TestMissionControlWorkspaceArgFromResumeToken_DefaultsLimit(t *testing.T) {
	t.Parallel()

	arg := missionControlWorkspaceArgFromResumeToken(missionControlResumeTokenPayload{
		ViewMode:    "list",
		StatePreset: "blocked",
		Search:      "sync",
	})

	if arg.rootLimit != missionControlDefaultLimit {
		t.Fatalf("expected default limit %d, got %d", missionControlDefaultLimit, arg.rootLimit)
	}
	if arg.viewMode != "list" || arg.statePreset != "blocked" || arg.search != "sync" {
		t.Fatalf("expected scope fields to be preserved, got %+v", arg)
	}
}
