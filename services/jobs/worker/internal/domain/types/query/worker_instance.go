package query

import "time"

// WorkerInstanceHeartbeatParams describes one worker liveness heartbeat upsert.
type WorkerInstanceHeartbeatParams struct {
	WorkerID    string
	Namespace   string
	PodName     string
	StartedAt   time.Time
	HeartbeatAt time.Time
	ExpiresAt   time.Time
}

// WorkerInstanceStopParams describes one worker stop marker update.
type WorkerInstanceStopParams struct {
	WorkerID  string
	StoppedAt time.Time
}
