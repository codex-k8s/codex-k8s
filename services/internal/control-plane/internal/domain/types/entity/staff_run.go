package entity

import "time"

// StaffRun is a staff-visible run record.
type StaffRun struct {
	ID              string
	CorrelationID   string
	ProjectID       string
	ProjectSlug     string
	ProjectName     string
	IssueNumber     int
	IssueURL        string
	PRNumber        int
	PRURL           string
	TriggerKind     string
	TriggerLabel    string
	AgentKey        string
	JobName         string
	JobNamespace    string
	Namespace       string
	JobExists       bool
	NamespaceExists bool
	WaitState       string
	WaitReason      string
	WaitSince       *time.Time
	LastHeartbeatAt *time.Time
	Status          string
	CreatedAt       time.Time
	StartedAt       *time.Time
	FinishedAt      *time.Time
}

// StaffRunLogs is a staff-visible logs snapshot for one run.
type StaffRunLogs struct {
	RunID        string
	Status       string
	UpdatedAt    *time.Time
	SnapshotJSON []byte
	TailLines    []string
}

// StaffFlowEvent is a staff-visible flow event.
type StaffFlowEvent struct {
	CorrelationID string
	EventType     string
	CreatedAt     time.Time
	PayloadJSON   []byte
}
