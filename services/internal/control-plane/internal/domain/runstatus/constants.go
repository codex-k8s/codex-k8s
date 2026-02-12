package runstatus

const (
	localeRU = "ru"
	localeEN = "en"
)

const (
	commentMarkerPrefix = "<!-- codex-k8s:run-status "
	commentMarkerSuffix = " -->"
)

const (
	deleteNamespacePath = "/api/v1/runs/namespace/cleanup/"
)

const (
	triggerKindDev       = "dev"
	triggerKindDevRevise = "dev_revise"
)

const (
	runtimeModeFullEnv = "full-env"
	runtimeModeCode    = "code-only"
)

const (
	runStatusSucceeded = "succeeded"
	runStatusFailed    = "failed"
)
