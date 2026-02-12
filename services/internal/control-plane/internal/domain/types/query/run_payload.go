package query

// RunPayload represents normalized run payload persisted in agent_runs.
type RunPayload struct {
	Project    RunPayloadProject    `json:"project"`
	Repository RunPayloadRepository `json:"repository"`
	Agent      *RunPayloadAgent     `json:"agent,omitempty"`
	Issue      *RunPayloadIssue     `json:"issue,omitempty"`
	Trigger    *RunPayloadTrigger   `json:"trigger,omitempty"`
}

// RunPayloadProject is project section of run payload.
type RunPayloadProject struct {
	ID           string `json:"id"`
	RepositoryID string `json:"repository_id"`
	ServicesYAML string `json:"services_yaml"`
}

// RunPayloadRepository is repository section of run payload.
type RunPayloadRepository struct {
	FullName string `json:"full_name"`
	Name     string `json:"name"`
}

// RunPayloadAgent is agent context section of run payload.
type RunPayloadAgent struct {
	ID   string `json:"id,omitempty"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// RunPayloadIssue is issue section of run payload.
type RunPayloadIssue struct {
	Number  int64  `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	HTMLURL string `json:"html_url"`
}

// RunPayloadTrigger is trigger section of run payload.
type RunPayloadTrigger struct {
	Label string `json:"label"`
	Kind  string `json:"kind"`
}
