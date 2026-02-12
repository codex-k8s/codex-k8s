package mcp

import "testing"

func TestFilterIssueCommentsByAuthor(t *testing.T) {
	comments := []GitHubIssueComment{
		{ID: 1, User: "platform-bot", Body: "system comment"},
		{ID: 2, User: "owner-user", Body: "human comment"},
		{ID: 3, User: "Platform-Bot", Body: "another system comment"},
	}

	filtered := filterIssueCommentsByAuthor(comments, "platform-bot")
	if len(filtered) != 1 {
		t.Fatalf("expected exactly one visible comment, got %d", len(filtered))
	}
	if filtered[0].ID != 2 {
		t.Fatalf("expected comment id 2 to remain, got %d", filtered[0].ID)
	}
}

func TestFilterIssueCommentsByAuthorNoopWhenExcludedIsEmpty(t *testing.T) {
	comments := []GitHubIssueComment{
		{ID: 1, User: "platform-bot", Body: "system comment"},
		{ID: 2, User: "owner-user", Body: "human comment"},
	}

	filtered := filterIssueCommentsByAuthor(comments, "")
	if len(filtered) != len(comments) {
		t.Fatalf("expected noop filter, got %d items", len(filtered))
	}
}
