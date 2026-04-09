package agentsession

import (
	"encoding/json"
	"testing"
	"time"

	domainrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/agentsession"
)

func TestBuildUpsertRecord_PreservesExistingSnapshotPayloads(t *testing.T) {
	t.Parallel()

	existing := domainrepo.Session{
		RunID:              "run-1",
		CorrelationID:      "corr-1",
		RepositoryFullName: "codex-k8s/kodex",
		AgentKey:           "dev",
		BranchName:         "codex/issue-259",
		SessionID:          "session-existing",
		CodexSessionPath:   "/tmp/session.json",
		CodexSessionJSON:   json.RawMessage(`{"cursor":"abc"}`),
		StartedAt:          time.Date(2026, time.March, 11, 14, 0, 0, 0, time.UTC),
		SnapshotVersion:    3,
		SnapshotChecksum:   "old",
		SnapshotUpdatedAt:  time.Date(2026, time.March, 11, 14, 1, 0, 0, time.UTC),
	}

	record, err := buildUpsertRecord(domainrepo.UpsertParams{
		RunID:                   "run-1",
		CorrelationID:           "corr-1",
		RepositoryFullName:      "codex-k8s/kodex",
		AgentKey:                "dev",
		BranchName:              "codex/issue-259",
		Status:                  "running",
		SessionJSON:             json.RawMessage(`{"status":"running"}`),
		ExpectedSnapshotVersion: 3,
	}, &existing)
	if err != nil {
		t.Fatalf("buildUpsertRecord() error = %v", err)
	}

	if record.SessionID != "session-existing" {
		t.Fatalf("expected preserved session id, got %q", record.SessionID)
	}
	if record.CodexSessionPath != "/tmp/session.json" {
		t.Fatalf("expected preserved session path, got %q", record.CodexSessionPath)
	}
	if string(record.CodexSessionJSON) != `{"cursor":"abc"}` {
		t.Fatalf("expected preserved codex session json, got %s", string(record.CodexSessionJSON))
	}
	if record.SnapshotChecksum == "" {
		t.Fatal("expected non-empty checksum")
	}
}

func TestIsIdempotentReplay(t *testing.T) {
	t.Parallel()

	if !isIdempotentReplay(domainrepo.Session{
		SnapshotVersion:  4,
		SnapshotChecksum: "abc",
	}, 3, "abc") {
		t.Fatal("expected idempotent replay to be detected")
	}
	if isIdempotentReplay(domainrepo.Session{
		SnapshotVersion:  4,
		SnapshotChecksum: "def",
	}, 3, "abc") {
		t.Fatal("unexpected idempotent replay for different checksum")
	}
}
