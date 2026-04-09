package http

import "github.com/codex-k8s/kodex/services/external/api-gateway/internal/transport/http/models"

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
	runID           string
	tailLines       int32
	includeSnapshot bool
}

// runActionArg keeps path+body input for run-level action endpoints.
type runActionArg struct {
	runID string
	body  models.RunActionRequest
}

// runListPageArg keeps query input for one runs list page.
type runListPageArg struct {
	page     int32
	pageSize int32
}

type systemSettingUpdateArg struct {
	key  string
	body models.UpdateSystemSettingBooleanRequest
}

// runEventsArg keeps path+query input for run events endpoint.
type runEventsArg struct {
	runID          string
	limit          int32
	includePayload bool
}

// runRealtimeArg keeps path+query input for run realtime websocket endpoint.
type runRealtimeArg struct {
	runID       string
	eventsLimit int32
	includeLogs bool
	tailLines   int32
}

// runtimeDeployListArg keeps filters for runtime deploy tasks list endpoint.
type runtimeDeployListArg struct {
	page      int32
	pageSize  int32
	status    string
	targetEnv string
}

// runtimeDeployActionArg keeps path+body input for cancel/stop endpoints.
type runtimeDeployActionArg struct {
	runID string
	body  models.RuntimeDeployTaskActionRequest
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
