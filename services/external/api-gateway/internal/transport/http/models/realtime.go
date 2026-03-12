package models

// RunRealtimeMessageType is one envelope message discriminator for run realtime stream.
type RunRealtimeMessageType string

const (
	RunRealtimeMessageTypeSnapshot RunRealtimeMessageType = "snapshot"
	RunRealtimeMessageTypeRun      RunRealtimeMessageType = "run"
	RunRealtimeMessageTypeEvents   RunRealtimeMessageType = "events"
	RunRealtimeMessageTypeLogs     RunRealtimeMessageType = "logs"
	RunRealtimeMessageTypeError    RunRealtimeMessageType = "error"
)

// RunRealtimeMessage is a typed websocket message for staff run realtime updates.
type RunRealtimeMessage struct {
	Type    RunRealtimeMessageType `json:"type"`
	Run     *Run                   `json:"run,omitempty"`
	Events  []FlowEvent            `json:"events,omitempty"`
	Logs    *RunLogs               `json:"logs,omitempty"`
	Message *string                `json:"message,omitempty"`
	SentAt  string                 `json:"sent_at"`
}

// ListRealtimeMessageType is one envelope discriminator for paginated staff list streams.
type ListRealtimeMessageType string

const (
	ListRealtimeMessageTypeSnapshot ListRealtimeMessageType = "snapshot"
	ListRealtimeMessageTypeError    ListRealtimeMessageType = "error"
)

// RunsRealtimeMessage is a typed websocket message for the paginated runs list.
type RunsRealtimeMessage struct {
	Type                  ListRealtimeMessageType `json:"type"`
	Items                 []Run                   `json:"items,omitempty"`
	Pagination            *Pagination             `json:"pagination,omitempty"`
	WaitQueueCount        *int                    `json:"wait_queue_count,omitempty"`
	PendingApprovalsCount *int                    `json:"pending_approvals_count,omitempty"`
	Message               *string                 `json:"message,omitempty"`
	SentAt                string                  `json:"sent_at"`
}

// RuntimeDeployTasksRealtimeMessage is a typed websocket message for the paginated runtime deploy tasks list.
type RuntimeDeployTasksRealtimeMessage struct {
	Type       ListRealtimeMessageType     `json:"type"`
	Items      []RuntimeDeployTaskListItem `json:"items,omitempty"`
	Pagination *Pagination                 `json:"pagination,omitempty"`
	Message    *string                     `json:"message,omitempty"`
	SentAt     string                      `json:"sent_at"`
}
