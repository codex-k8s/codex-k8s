package entity

import "encoding/json"

// AgentRun stores the subset of agent_runs required by runtime services.
type AgentRun struct {
	ID            string
	CorrelationID string
	ProjectID     string
	Status        string
	RunPayload    json.RawMessage
}
