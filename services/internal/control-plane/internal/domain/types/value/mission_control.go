package value

// MissionControlCoreFeatureFlagEnv gates worker warmup and core write-side Mission Control flows.
const MissionControlCoreFeatureFlagEnv = "CODEXK8S_MISSION_CONTROL_ENABLED"

// MissionControlVoiceFeatureFlagEnv gates optional voice-only Mission Control paths.
const MissionControlVoiceFeatureFlagEnv = "CODEXK8S_MISSION_CONTROL_VOICE_ENABLED"

// MissionControlWarmupSummary captures aggregate verification evidence for warmup/backfill.
type MissionControlWarmupSummary struct {
	ProjectID            string
	EntityCount          int64
	RelationCount        int64
	TimelineEntryCount   int64
	CommandCount         int64
	MaxProjectionVersion int64
}

// MissionControlRolloutState captures rollout readiness and enablement gates.
type MissionControlRolloutState struct {
	CoreFeatureEnabled  bool
	VoiceFeatureEnabled bool
	SchemaReady         bool
	DomainReady         bool
	WarmupVerified      bool
	WritePathEnabled    bool
}

// MissionControlRolloutCapabilities describes what the current rollout state permits.
type MissionControlRolloutCapabilities struct {
	CanRunWarmup         bool
	CanServeSnapshot     bool
	CanOpenRealtime      bool
	CanSubmitCoreCommand bool
	CanUseVoicePath      bool
}
