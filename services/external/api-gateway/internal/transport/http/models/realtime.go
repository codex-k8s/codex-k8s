package models

// RunRealtimeMessageType is one envelope message discriminator for run realtime stream.
type RunRealtimeMessageType string

const (
	RunRealtimeMessageTypeSnapshot                 RunRealtimeMessageType = "snapshot"
	RunRealtimeMessageTypeRun                      RunRealtimeMessageType = "run"
	RunRealtimeMessageTypeEvents                   RunRealtimeMessageType = "events"
	RunRealtimeMessageTypeLogs                     RunRealtimeMessageType = "logs"
	RunRealtimeMessageTypeWaitEntered              RunRealtimeMessageType = "wait_entered"
	RunRealtimeMessageTypeWaitUpdated              RunRealtimeMessageType = "wait_updated"
	RunRealtimeMessageTypeWaitResolved             RunRealtimeMessageType = "wait_resolved"
	RunRealtimeMessageTypeWaitManualActionRequired RunRealtimeMessageType = "wait_manual_action_required"
	RunRealtimeMessageTypeError                    RunRealtimeMessageType = "error"
)

// RunRealtimeMessage is a typed websocket message for staff run realtime updates.
type RunRealtimeMessage struct {
	Type             RunRealtimeMessageType    `json:"type"`
	Run              *Run                      `json:"run,omitempty"`
	Events           []FlowEvent               `json:"events,omitempty"`
	Logs             *RunLogs                  `json:"logs,omitempty"`
	WaitProjection   *RunWaitProjection        `json:"wait_projection,omitempty"`
	WaitResolution   *RunWaitResolution        `json:"wait_resolution,omitempty"`
	WaitManualAction *RunWaitManualActionEvent `json:"wait_manual_action,omitempty"`
	Message          *string                   `json:"message,omitempty"`
	SentAt           string                    `json:"sent_at"`
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

// SystemSettingsRealtimeMessage is a typed websocket message for platform settings updates.
type SystemSettingsRealtimeMessage struct {
	Type    ListRealtimeMessageType `json:"type"`
	Items   []SystemSetting         `json:"items,omitempty"`
	Message *string                 `json:"message,omitempty"`
	SentAt  string                  `json:"sent_at"`
}
