package runner

import "testing"

func TestDeriveDiscussionIssueStateWithoutAgentReply(t *testing.T) {
	issue := discussionIssueAPIResponse{
		State: "open",
		Labels: []discussionIssueLabelItem{
			{Name: "mode:discussion"},
		},
	}
	comments := []discussionIssueCommentResponse{
		{ID: 11, Body: "first", User: discussionIssueCommentUser{Login: "owner", Type: "User"}},
		{ID: 12, Body: "second", User: discussionIssueCommentUser{Login: "reviewer", Type: "User"}},
	}

	state := deriveDiscussionIssueState(issue, comments, "codex-bot")

	if !state.HasDiscussionLabel {
		t.Fatal("expected discussion label to be detected")
	}
	if state.HasAgentReply {
		t.Fatal("did not expect agent reply")
	}
	if state.HasHumanAfterAgentReply {
		t.Fatal("did not expect human-after-agent flag without agent reply")
	}
	if got, want := state.MaxHumanCommentID, int64(12); got != want {
		t.Fatalf("expected max human comment id %d, got %d", want, got)
	}
	if got, want := len(state.PendingHumanComments), 2; got != want {
		t.Fatalf("expected %d pending comments, got %d", want, got)
	}
	if state.PendingHumanComments[0].ID != 11 || state.PendingHumanComments[1].ID != 12 {
		t.Fatalf("unexpected pending comment ids: %#v", state.PendingHumanComments)
	}
}

func TestDeriveDiscussionIssueStateTracksCommentsAfterAgentReply(t *testing.T) {
	issue := discussionIssueAPIResponse{
		State: "open",
		Labels: []discussionIssueLabelItem{
			{Name: "mode:discussion"},
			{Name: "run:dev"},
		},
	}
	comments := []discussionIssueCommentResponse{
		{ID: 10, Body: "first", User: discussionIssueCommentUser{Login: "owner", Type: "User"}},
		{ID: 20, Body: "reply", User: discussionIssueCommentUser{Login: "codex-bot", Type: "Bot"}},
		{ID: 30, Body: "status <!-- codex-k8s:run-status abc -->", User: discussionIssueCommentUser{Login: "codex-bot", Type: "Bot"}},
		{ID: 40, Body: "follow-up", User: discussionIssueCommentUser{Login: "owner", Type: "User"}},
		{ID: 50, Body: "another", User: discussionIssueCommentUser{Login: "qa-user", Type: "User"}},
	}

	state := deriveDiscussionIssueState(issue, comments, "codex-bot")

	if !state.HasAgentReply {
		t.Fatal("expected agent reply to be detected")
	}
	if !state.HasHumanAfterAgentReply {
		t.Fatal("expected human-after-agent flag to be set")
	}
	if !state.HasRunLabel {
		t.Fatal("expected run label to be detected")
	}
	if got, want := len(state.PendingHumanComments), 2; got != want {
		t.Fatalf("expected %d pending comments, got %d", want, got)
	}
	if state.PendingHumanComments[0].ID != 40 || state.PendingHumanComments[1].ID != 50 {
		t.Fatalf("unexpected pending comment ids: %#v", state.PendingHumanComments)
	}
}

func TestProcessedDiscussionCommentIDs(t *testing.T) {
	previous := []discussionPendingHumanComment{
		{ID: 10},
		{ID: 30},
		{ID: 20},
	}
	current := discussionIssueState{
		PendingHumanComments: []discussionPendingHumanComment{
			{ID: 30},
			{ID: 40},
		},
	}

	processed := processedDiscussionCommentIDs(previous, current)

	if got, want := len(processed), 2; got != want {
		t.Fatalf("expected %d processed ids, got %d", want, got)
	}
	if processed[0] != 10 || processed[1] != 20 {
		t.Fatalf("unexpected processed ids: %#v", processed)
	}
}
