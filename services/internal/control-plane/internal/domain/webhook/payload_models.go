package webhook

import (
	"encoding/json"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

type githubRunPayload struct {
	Source        string                     `json:"source"`
	DeliveryID    string                     `json:"delivery_id"`
	EventType     string                     `json:"event_type"`
	ReceivedAt    string                     `json:"received_at"`
	Repository    githubRunRepositoryPayload `json:"repository"`
	Installation  githubInstallationPayload  `json:"installation"`
	Sender        githubActorPayload         `json:"sender"`
	Action        string                     `json:"action"`
	RawPayload    json.RawMessage            `json:"raw_payload"`
	CorrelationID string                     `json:"correlation_id"`
	Project       githubRunProjectPayload    `json:"project"`
	LearningMode  bool                       `json:"learning_mode"`
	Agent         githubRunAgentPayload      `json:"agent"`
	Issue         *githubRunIssuePayload     `json:"issue,omitempty"`
	Trigger       *githubIssueTriggerPayload `json:"trigger,omitempty"`
}

type githubRunRepositoryPayload struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
	Name     string `json:"name"`
	Private  bool   `json:"private"`
}

type githubInstallationPayload struct {
	ID int64 `json:"id"`
}

type githubActorPayload struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}

type githubRunProjectPayload struct {
	ID              string `json:"id"`
	RepositoryID    string `json:"repository_id"`
	ServicesYAML    string `json:"services_yaml"`
	BindingResolved bool   `json:"binding_resolved"`
}

type githubRunAgentPayload struct {
	ID   string `json:"id,omitempty"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type githubRunIssuePayload struct {
	ID          int64                     `json:"id"`
	Number      int64                     `json:"number"`
	Title       string                    `json:"title"`
	HTMLURL     string                    `json:"html_url"`
	State       string                    `json:"state"`
	User        githubActorPayload        `json:"user"`
	PullRequest *githubPullRequestPayload `json:"pull_request,omitempty"`
}

type githubPullRequestPayload struct {
	URL     string `json:"url"`
	HTMLURL string `json:"html_url"`
}

type githubIssueTriggerPayload struct {
	Source string                    `json:"source"`
	Label  string                    `json:"label"`
	Kind   webhookdomain.TriggerKind `json:"kind"`
}

type githubFlowEventPayload struct {
	Source          string                      `json:"source"`
	DeliveryID      string                      `json:"delivery_id"`
	EventType       string                      `json:"event_type"`
	Action          string                      `json:"action"`
	CorrelationID   string                      `json:"correlation_id"`
	Sender          githubActorPayload          `json:"sender"`
	Repository      githubFlowRepositoryPayload `json:"repository"`
	Inserted        *bool                       `json:"inserted,omitempty"`
	RunID           string                      `json:"run_id,omitempty"`
	Label           string                      `json:"label,omitempty"`
	RunKind         webhookdomain.TriggerKind   `json:"run_kind,omitempty"`
	IssueNumber     int64                       `json:"issue_number,omitempty"`
	Reason          string                      `json:"reason,omitempty"`
	BindingResolved *bool                       `json:"binding_resolved,omitempty"`
	Issue           *githubIgnoredIssuePayload  `json:"issue,omitempty"`
}

type githubFlowRepositoryPayload struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
	Name     string `json:"name"`
}

type githubIgnoredIssuePayload struct {
	ID      int64  `json:"id"`
	Number  int64  `json:"number"`
	Title   string `json:"title"`
	HTMLURL string `json:"html_url"`
	State   string `json:"state"`
}
