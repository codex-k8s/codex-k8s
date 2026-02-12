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
	runManagementPathPrefix = "/runs/"
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
