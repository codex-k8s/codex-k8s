package webhook

import (
	"encoding/json"
	"time"
)

// IngestCommand is a normalized webhook command accepted by the domain service.
type IngestCommand struct {
	// CorrelationID is a deduplication key shared across flow records.
	CorrelationID string
	// EventType is a provider event name from request headers.
	EventType string
	// DeliveryID is a provider delivery identifier from webhook headers.
	DeliveryID string
	// ReceivedAt is the time when ingress received the webhook.
	ReceivedAt time.Time
	// Payload is raw webhook JSON body.
	Payload json.RawMessage
}

// IngestResult is a transport-facing outcome of webhook ingestion.
type IngestResult struct {
	// CorrelationID mirrors the request correlation id.
	CorrelationID string `json:"correlation_id"`
	// RunID references an agent run linked to this webhook.
	RunID string `json:"run_id,omitempty"`
	// Status is accepted, duplicate, or ignored.
	Status string `json:"status"`
	// Duplicate is true when webhook was already processed before.
	Duplicate bool `json:"duplicate"`
}

// TriggerLabels defines active issue labels that create development runs.
type TriggerLabels struct {
	RunDev       string
	RunDevRevise string
}

func defaultTriggerLabels() TriggerLabels {
	return TriggerLabels{
		RunDev:       "run:dev",
		RunDevRevise: "run:dev:revise",
	}
}

type issueRunTrigger struct {
	Label string
	Kind  string
}

type githubEnvelope struct {
	Action       string `json:"action"`
	Installation struct {
		ID int64 `json:"id"`
	} `json:"installation"`
	Repository struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
		Name     string `json:"name"`
		Private  bool   `json:"private"`
	} `json:"repository"`
	Issue struct {
		ID      int64  `json:"id"`
		Number  int64  `json:"number"`
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
		State   string `json:"state"`
		User    struct {
			Login string `json:"login"`
			ID    int64  `json:"id"`
		} `json:"user"`
		PullRequest *struct {
			URL     string `json:"url"`
			HTMLURL string `json:"html_url"`
		} `json:"pull_request"`
	} `json:"issue"`
	Label struct {
		Name string `json:"name"`
	} `json:"label"`
	Sender struct {
		Login string `json:"login"`
		ID    int64  `json:"id"`
	} `json:"sender"`
}
