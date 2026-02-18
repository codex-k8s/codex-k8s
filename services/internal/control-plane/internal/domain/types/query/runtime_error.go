package query

import "encoding/json"

// RuntimeErrorListState defines runtime error list scope.
type RuntimeErrorListState string

const (
	RuntimeErrorListStateActive RuntimeErrorListState = "active"
	RuntimeErrorListStateViewed RuntimeErrorListState = "viewed"
	RuntimeErrorListStateAll    RuntimeErrorListState = "all"
)

// RuntimeErrorRecordParams describes one runtime error journal write.
type RuntimeErrorRecordParams struct {
	Source        string
	Level         string
	Message       string
	DetailsJSON   json.RawMessage
	StackTrace    string
	CorrelationID string
	RunID         string
	ProjectID     string
	Namespace     string
	JobName       string
}

// RuntimeErrorListFilter defines runtime error list filters.
type RuntimeErrorListFilter struct {
	Limit         int
	State         RuntimeErrorListState
	Level         string
	Source        string
	RunID         string
	CorrelationID string
}

// RuntimeErrorMarkViewedParams marks one runtime error as viewed.
type RuntimeErrorMarkViewedParams struct {
	ID       string
	ViewerID string
}
