package value

import "time"

// AgentSessionSnapshotState describes persisted snapshot version metadata.
type AgentSessionSnapshotState struct {
	SnapshotVersion   int64
	SnapshotChecksum  string
	SnapshotUpdatedAt time.Time
}
