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
