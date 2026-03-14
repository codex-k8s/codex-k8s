package http

import "testing"

func TestMissionControlResumeToken_RoundTrip(t *testing.T) {
	t.Parallel()

	token, err := encodeMissionControlResumeToken(missionControlResumeTokenPayload{
		SnapshotID:   "snapshot-1",
		ViewMode:     "board",
		ActiveFilter: "working",
		Search:       "owner",
		Cursor:       "cursor-1",
		Limit:        25,
		IssuedAt:     "2026-03-14T00:00:00Z",
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
	if payload.ViewMode != "board" || payload.ActiveFilter != "working" {
		t.Fatalf("expected scope to round-trip, got view_mode=%q active_filter=%q", payload.ViewMode, payload.ActiveFilter)
	}
	if payload.Limit != 25 {
		t.Fatalf("expected limit 25, got %d", payload.Limit)
	}
}

func TestMissionControlDashboardArgFromResumeToken_DefaultsLimit(t *testing.T) {
	t.Parallel()

	arg := missionControlDashboardArgFromResumeToken(missionControlResumeTokenPayload{
		ViewMode:     "list",
		ActiveFilter: "blocked",
		Search:       "sync",
	})

	if arg.limit != missionControlDefaultLimit {
		t.Fatalf("expected default limit %d, got %d", missionControlDefaultLimit, arg.limit)
	}
	if arg.viewMode != "list" || arg.activeFilter != "blocked" || arg.search != "sync" {
		t.Fatalf("expected scope fields to be preserved, got %+v", arg)
	}
}
