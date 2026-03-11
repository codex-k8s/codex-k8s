package agentsession

import "fmt"

// SnapshotVersionConflict indicates that snapshot CAS version is stale.
type SnapshotVersionConflict struct {
	ExpectedSnapshotVersion int64
	ActualSnapshotVersion   int64
}

func (e SnapshotVersionConflict) Error() string {
	return fmt.Sprintf(
		"agent session snapshot version conflict: expected %d, actual %d",
		e.ExpectedSnapshotVersion,
		e.ActualSnapshotVersion,
	)
}
