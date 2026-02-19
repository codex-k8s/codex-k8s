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

// runAccessBypassArg keeps query/path input for public run bypass endpoints.
type runAccessBypassArg struct {
	runID       string
	accessKey   string
	namespace   string
	targetEnv   string
	runtimeMode string
}

// runtimeDeployListArg keeps filters for runtime deploy tasks list endpoint.
type runtimeDeployListArg struct {
	limit     int32
	status    string
	targetEnv string
}

// runtimeErrorsListArg keeps filters for runtime errors list endpoint.
type runtimeErrorsListArg struct {
	limit         int32
	state         string
	level         string
	source        string
	runID         string
	correlationID string
}

// registryImagesListArg keeps list filters for registry images endpoint.
type registryImagesListArg struct {
	repository        string
	limitRepositories int32
	limitTags         int32
}
