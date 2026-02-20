package entity

import "time"

// RuntimeDeployTaskStatus describes runtime deployment task lifecycle.
type RuntimeDeployTaskStatus string

const (
	RuntimeDeployTaskStatusPending   RuntimeDeployTaskStatus = "pending"
	RuntimeDeployTaskStatusRunning   RuntimeDeployTaskStatus = "running"
	RuntimeDeployTaskStatusSucceeded RuntimeDeployTaskStatus = "succeeded"
	RuntimeDeployTaskStatusFailed    RuntimeDeployTaskStatus = "failed"
	RuntimeDeployTaskStatusCanceled  RuntimeDeployTaskStatus = "canceled"
)

// RuntimeDeployTask stores desired and actual runtime deployment state for one run.
type RuntimeDeployTask struct {
	RunID              string
	RuntimeMode        string
	Namespace          string
	TargetEnv          string
	SlotNo             int
	RepositoryFullName string
	ServicesYAMLPath   string
	BuildRef           string
	DeployOnly         bool
	Status             RuntimeDeployTaskStatus
	LeaseOwner         string
	LeaseUntil         time.Time
	Attempts           int
	LastError          string
	ResultNamespace    string
	ResultTargetEnv    string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	StartedAt          time.Time
	FinishedAt         time.Time
	Logs               []RuntimeDeployTaskLogEntry
}

// RuntimeDeployTaskLogEntry stores one build/deploy task log line.
type RuntimeDeployTaskLogEntry struct {
	Stage     string
	Level     string
	Message   string
	CreatedAt time.Time
}

// IsTerminal returns true when task reached a terminal state.
func (s RuntimeDeployTaskStatus) IsTerminal() bool {
	return s == RuntimeDeployTaskStatusSucceeded || s == RuntimeDeployTaskStatusFailed || s == RuntimeDeployTaskStatusCanceled
}
