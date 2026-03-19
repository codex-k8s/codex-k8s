package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

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

func TestResolveMissionControlWorkspaceArg_UsesRootLimitQueryParam(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/staff/mission-control/workspace?root_limit=17&view_mode=graph&state_preset=working", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	arg, err := resolveMissionControlWorkspaceArg(ctx)
	if err != nil {
		t.Fatalf("resolveMissionControlWorkspaceArg() error = %v", err)
	}
	if got, want := arg.rootLimit, int32(17); got != want {
		t.Fatalf("rootLimit = %d, want %d", got, want)
	}
	if got, want := arg.viewMode, "graph"; got != want {
		t.Fatalf("viewMode = %q, want %q", got, want)
	}
	if got, want := arg.statePreset, "working"; got != want {
		t.Fatalf("statePreset = %q, want %q", got, want)
	}
}
