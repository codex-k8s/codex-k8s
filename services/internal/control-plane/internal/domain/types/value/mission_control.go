package value

// MissionControlWarmupSummary captures aggregate verification evidence for warmup/backfill.
type MissionControlWarmupSummary struct {
	ProjectID                    string
	EntityCount                  int64
	RelationCount                int64
	TimelineEntryCount           int64
	CommandCount                 int64
	MaxProjectionVersion         int64
	RunEntityCount               int64
	LegacyAgentCount             int64
	ContinuityGapCount           int64
	OpenContinuityGapCount       int64
	BlockingGapCount             int64
	MissingPullRequestGapCount   int64
	MissingFollowUpIssueGapCount int64
	WatermarkCount               int64
}

// MissionControlRolloutState captures rollout readiness and enablement gates.
type MissionControlRolloutState struct {
	SchemaReady bool
	DomainReady bool
}

// MissionControlRolloutCapabilities describes what the current rollout state permits.
type MissionControlRolloutCapabilities struct {
	CanRunWarmup         bool
	CanServeSnapshot     bool
	CanOpenRealtime      bool
	CanSubmitCoreCommand bool
	CanUseVoicePath      bool
}
