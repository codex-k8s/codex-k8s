package http

import (
	"testing"

	"github.com/codex-k8s/kodex/services/external/api-gateway/internal/transport/http/models"
)

func TestBuildRunWaitRealtimeMessages_EmitsWaitEntered(t *testing.T) {
	t.Parallel()

	messages := buildRunWaitRealtimeMessages(
		runRealtimeSnapshot{Run: models.Run{}},
		runRealtimeSnapshot{Run: models.Run{WaitProjection: testRunWaitProjection("wait-1", nil)}},
	)

	if len(messages) != 1 {
		t.Fatalf("messages length = %d, want 1", len(messages))
	}
	if got, want := messages[0].Type, models.RunRealtimeMessageTypeWaitEntered; got != want {
		t.Fatalf("message type = %q, want %q", got, want)
	}
	if messages[0].WaitProjection == nil {
		t.Fatal("wait_projection is nil")
	}
	if got, want := messages[0].WaitProjection.DominantWait.WaitID, "wait-1"; got != want {
		t.Fatalf("dominant wait id = %q, want %q", got, want)
	}
}

func TestBuildRunWaitRealtimeMessages_EmitsManualActionRequired(t *testing.T) {
	t.Parallel()

	nextManualAction := &models.GitHubRateLimitManualAction{
		Kind:            "resume_agent_session",
		Summary:         "Resume runner session",
		DetailsMarkdown: "Retry resume once operator checks rate-limit window.",
	}
	messages := buildRunWaitRealtimeMessages(
		runRealtimeSnapshot{Run: models.Run{WaitProjection: testRunWaitProjection("wait-1", nil)}},
		runRealtimeSnapshot{
			Run: models.Run{WaitProjection: testRunWaitProjection("wait-1", nextManualAction)},
			Events: []models.FlowEvent{{
				EventType:   "run.wait.paused",
				CreatedAt:   "2026-03-14T18:00:00Z",
				PayloadJSON: `{"wait_id":"wait-1","event_key":"github_rate_limit.manual_action_required"}`,
			}},
		},
	)

	if len(messages) != 1 {
		t.Fatalf("messages length = %d, want 1", len(messages))
	}
	if got, want := messages[0].Type, models.RunRealtimeMessageTypeWaitManualActionRequired; got != want {
		t.Fatalf("message type = %q, want %q", got, want)
	}
	if messages[0].WaitManualAction == nil {
		t.Fatal("wait_manual_action is nil")
	}
	if got, want := messages[0].WaitManualAction.ManualAction.Kind, "resume_agent_session"; got != want {
		t.Fatalf("manual action kind = %q, want %q", got, want)
	}
	if got, want := messages[0].WaitManualAction.UpdatedAt, "2026-03-14T18:00:00Z"; got != want {
		t.Fatalf("updated_at = %q, want %q", got, want)
	}
}

func TestBuildRunWaitRealtimeMessages_EmitsWaitResolvedFromFlowEvent(t *testing.T) {
	t.Parallel()

	messages := buildRunWaitRealtimeMessages(
		runRealtimeSnapshot{Run: models.Run{WaitProjection: testRunWaitProjection("wait-1", nil)}},
		runRealtimeSnapshot{
			Run: models.Run{},
			Events: []models.FlowEvent{{
				EventType:   "run.wait.resumed",
				CreatedAt:   "2026-03-14T19:10:00Z",
				PayloadJSON: `{"wait_id":"wait-1","contour_kind":"platform_pat","resolution_kind":"manually_resolved"}`,
			}},
		},
	)

	if len(messages) != 1 {
		t.Fatalf("messages length = %d, want 1", len(messages))
	}
	if got, want := messages[0].Type, models.RunRealtimeMessageTypeWaitResolved; got != want {
		t.Fatalf("message type = %q, want %q", got, want)
	}
	if messages[0].WaitResolution == nil {
		t.Fatal("wait_resolution is nil")
	}
	if got, want := messages[0].WaitResolution.ResolutionKind, "manually_resolved"; got != want {
		t.Fatalf("resolution_kind = %q, want %q", got, want)
	}
	if got, want := messages[0].WaitResolution.ResolvedAt, "2026-03-14T19:10:00Z"; got != want {
		t.Fatalf("resolved_at = %q, want %q", got, want)
	}
}

func TestBuildRunWaitRealtimeMessages_EmitsWaitResolvedThenUpdatedOnDominantWaitSwitch(t *testing.T) {
	t.Parallel()

	previousProjection := testRunWaitProjection("wait-1", nil)
	previousProjection.RelatedWaits = []models.GitHubRateLimitWaitItem{
		testRunWaitItem("wait-2", nil),
	}

	nextProjection := testRunWaitProjection("wait-2", nil)

	messages := buildRunWaitRealtimeMessages(
		runRealtimeSnapshot{Run: models.Run{WaitProjection: previousProjection}},
		runRealtimeSnapshot{
			Run: models.Run{WaitProjection: nextProjection},
			Events: []models.FlowEvent{{
				EventType:   "run.wait.resumed",
				CreatedAt:   "2026-03-14T19:15:00Z",
				PayloadJSON: `{"wait_id":"wait-1","contour_kind":"platform_pat","resolution_kind":"auto_resumed"}`,
			}},
		},
	)

	if len(messages) != 2 {
		t.Fatalf("messages length = %d, want 2", len(messages))
	}
	if got, want := messages[0].Type, models.RunRealtimeMessageTypeWaitResolved; got != want {
		t.Fatalf("first message type = %q, want %q", got, want)
	}
	if messages[0].WaitResolution == nil {
		t.Fatal("first wait_resolution is nil")
	}
	if got, want := messages[0].WaitResolution.WaitID, "wait-1"; got != want {
		t.Fatalf("resolved wait_id = %q, want %q", got, want)
	}
	if got, want := messages[1].Type, models.RunRealtimeMessageTypeWaitUpdated; got != want {
		t.Fatalf("second message type = %q, want %q", got, want)
	}
	if messages[1].WaitProjection == nil {
		t.Fatal("second wait_projection is nil")
	}
	if got, want := messages[1].WaitProjection.DominantWait.WaitID, "wait-2"; got != want {
		t.Fatalf("updated dominant wait id = %q, want %q", got, want)
	}
}

func TestBuildRunWaitResolution_PrefersMatchingWaitIDFromFlowEvent(t *testing.T) {
	t.Parallel()

	resolution := buildRunWaitResolution(*testRunWaitProjection("wait-2", nil), []models.FlowEvent{
		{
			EventType:   "run.wait.resumed",
			CreatedAt:   "2026-03-14T19:10:00Z",
			PayloadJSON: `{"wait_id":"wait-1","contour_kind":"agent_bot_token","resolution_kind":"auto_resumed"}`,
		},
		{
			EventType:   "run.wait.resumed",
			CreatedAt:   "2026-03-14T19:12:00Z",
			PayloadJSON: `{"wait_id":"wait-2","contour_kind":"platform_pat","resolution_kind":"manually_resolved"}`,
		},
	})

	if got, want := resolution.WaitID, "wait-2"; got != want {
		t.Fatalf("wait_id = %q, want %q", got, want)
	}
	if got, want := resolution.ResolutionKind, "manually_resolved"; got != want {
		t.Fatalf("resolution_kind = %q, want %q", got, want)
	}
	if got, want := resolution.ResolvedAt, "2026-03-14T19:12:00Z"; got != want {
		t.Fatalf("resolved_at = %q, want %q", got, want)
	}
}

func testRunWaitProjection(waitID string, manualAction *models.GitHubRateLimitManualAction) *models.RunWaitProjection {
	return &models.RunWaitProjection{
		WaitState:    "waiting_backpressure",
		WaitReason:   "github_rate_limit",
		DominantWait: testRunWaitItem(waitID, manualAction),
		RelatedWaits: []models.GitHubRateLimitWaitItem{},
	}
}

func testRunWaitItem(waitID string, manualAction *models.GitHubRateLimitManualAction) models.GitHubRateLimitWaitItem {
	return models.GitHubRateLimitWaitItem{
		WaitID:       waitID,
		ContourKind:  "platform_pat",
		LimitKind:    "secondary",
		State:        "open",
		AttemptsUsed: 1,
		MaxAttempts:  3,
		ManualAction: manualAction,
	}
}
