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
	JobName         string
	JobNamespace    string
	Namespace       string
	JobExists       bool
	NamespaceExists bool
	WaitState       string
	WaitReason      string
	Status          string
	CreatedAt       time.Time
	StartedAt       *time.Time
	FinishedAt      *time.Time
}

// StaffFlowEvent is a staff-visible flow event.
type StaffFlowEvent struct {
	CorrelationID string
	EventType     string
	CreatedAt     time.Time
	PayloadJSON   []byte
}
