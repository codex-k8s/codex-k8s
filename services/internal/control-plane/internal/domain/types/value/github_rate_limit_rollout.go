package value

import sharedsystemsettings "github.com/codex-k8s/codex-k8s/libs/go/systemsettings"

// GitHubRateLimitCoreSettingKey gates persisted rate-limit wait creation and worker orchestration.
const GitHubRateLimitCoreSettingKey = sharedsystemsettings.GitHubRateLimitWaitEnabledKey

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
