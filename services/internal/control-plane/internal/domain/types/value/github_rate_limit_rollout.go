package value

// GitHubRateLimitCoreFeatureFlagEnv gates persisted rate-limit wait creation and worker orchestration.
const GitHubRateLimitCoreFeatureFlagEnv = "CODEXK8S_GITHUB_RATE_LIMIT_WAIT_ENABLED"

// GitHubRateLimitUIFeatureFlagEnv gates transport and UI exposure of wait projections.
const GitHubRateLimitUIFeatureFlagEnv = "CODEXK8S_GITHUB_RATE_LIMIT_WAIT_UI_ENABLED"

// GitHubRateLimitRolloutState captures rollout sequencing for the S12 capability.
type GitHubRateLimitRolloutState struct {
	CoreFeatureEnabled bool
	SchemaReady        bool
	DomainReady        bool
	WorkerReady        bool
	RunnerReady        bool
	TransportReady     bool
	UIFeatureEnabled   bool
}

// GitHubRateLimitRolloutCapabilities describes which paths are allowed under current rollout gates.
type GitHubRateLimitRolloutCapabilities struct {
	CanPersistWaits   bool
	CanRunWorkerSweep bool
	CanAcceptSignals  bool
	CanServeTransport bool
	CanExposeUI       bool
}
