package models

type IngestGitHubWebhookResponse struct {
	CorrelationID string `json:"correlation_id"`
	RunID         string `json:"run_id"`
	Status        string `json:"status"`
	Duplicate     bool   `json:"duplicate"`
}
