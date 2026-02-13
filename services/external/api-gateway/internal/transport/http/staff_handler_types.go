package http

// pathLimit keeps parsed path id and list limit together for staff list handlers.
type pathLimit struct {
	id    string
	limit int
}

// idLimitArg is a helper DTO for gRPC request builders with id+limit.
type idLimitArg struct {
	id    string
	limit int32
}

// runListFilterArg keeps query filters for jobs/waits staff endpoints.
type runListFilterArg struct {
	limit       int32
	triggerKind string
	status      string
	agentKey    string
	waitState   string
}

// runLogsArg keeps path+query input for run logs endpoint.
type runLogsArg struct {
	runID     string
	tailLines int32
}
