package enum

// MissionControlEntityKind identifies the active-set entity contour.
type MissionControlEntityKind string

const (
	MissionControlEntityKindWorkItem    MissionControlEntityKind = "work_item"
	MissionControlEntityKindDiscussion  MissionControlEntityKind = "discussion"
	MissionControlEntityKindRun         MissionControlEntityKind = "run"
	MissionControlEntityKindPullRequest MissionControlEntityKind = "pull_request"
	MissionControlEntityKindAgent       MissionControlEntityKind = "agent"
)

// MissionControlProviderKind identifies the upstream source for one entity.
type MissionControlProviderKind string

const (
	MissionControlProviderKindGitHub   MissionControlProviderKind = "github"
	MissionControlProviderKindPlatform MissionControlProviderKind = "platform"
)

// MissionControlActiveState describes the board/list placement of one entity.
type MissionControlActiveState string

const (
	MissionControlActiveStateWorking               MissionControlActiveState = "working"
	MissionControlActiveStateWaiting               MissionControlActiveState = "waiting"
	MissionControlActiveStateBlocked               MissionControlActiveState = "blocked"
	MissionControlActiveStateReview                MissionControlActiveState = "review"
	MissionControlActiveStateRecentCriticalUpdates MissionControlActiveState = "recent_critical_updates"
	MissionControlActiveStateArchived              MissionControlActiveState = "archived"
)

// MissionControlSyncStatus captures the projection freshness state for one entity.
type MissionControlSyncStatus string

const (
	MissionControlSyncStatusSynced      MissionControlSyncStatus = "synced"
	MissionControlSyncStatusPendingSync MissionControlSyncStatus = "pending_sync"
	MissionControlSyncStatusFailed      MissionControlSyncStatus = "failed"
	MissionControlSyncStatusDegraded    MissionControlSyncStatus = "degraded"
)

// MissionControlContinuityStatus captures the persisted completeness summary for one entity.
type MissionControlContinuityStatus string

const (
	MissionControlContinuityStatusComplete             MissionControlContinuityStatus = "complete"
	MissionControlContinuityStatusMissingRun           MissionControlContinuityStatus = "missing_run"
	MissionControlContinuityStatusMissingPullRequest   MissionControlContinuityStatus = "missing_pull_request"
	MissionControlContinuityStatusMissingFollowUpIssue MissionControlContinuityStatus = "missing_follow_up_issue"
	MissionControlContinuityStatusStaleProvider        MissionControlContinuityStatus = "stale_provider"
	MissionControlContinuityStatusOutOfScope           MissionControlContinuityStatus = "out_of_scope"
)

// MissionControlCoverageClass captures whether the entity belongs to open-primary scope or bounded closed context.
type MissionControlCoverageClass string

const (
	MissionControlCoverageClassOpenPrimary         MissionControlCoverageClass = "open_primary"
	MissionControlCoverageClassRecentClosedContext MissionControlCoverageClass = "recent_closed_context"
	MissionControlCoverageClassOutOfScope          MissionControlCoverageClass = "out_of_scope"
)

// MissionControlRelationKind describes one typed graph edge.
type MissionControlRelationKind string

const (
	MissionControlRelationKindLinkedTo            MissionControlRelationKind = "linked_to"
	MissionControlRelationKindBlocks              MissionControlRelationKind = "blocks"
	MissionControlRelationKindBlockedBy           MissionControlRelationKind = "blocked_by"
	MissionControlRelationKindFormalizedFrom      MissionControlRelationKind = "formalized_from"
	MissionControlRelationKindSpawnedRun          MissionControlRelationKind = "spawned_run"
	MissionControlRelationKindProducedPullRequest MissionControlRelationKind = "produced_pull_request"
	MissionControlRelationKindContinuesWith       MissionControlRelationKind = "continues_with"
	MissionControlRelationKindRelatedTo           MissionControlRelationKind = "related_to"
	MissionControlRelationKindOwnedBy             MissionControlRelationKind = "owned_by"
	MissionControlRelationKindAssignedTo          MissionControlRelationKind = "assigned_to"
	MissionControlRelationKindTrackedBy           MissionControlRelationKind = "tracked_by_command"
)

// MissionControlRelationSourceKind describes where one relation originated.
type MissionControlRelationSourceKind string

const (
	MissionControlRelationSourceKindPlatform       MissionControlRelationSourceKind = "platform"
	MissionControlRelationSourceKindProvider       MissionControlRelationSourceKind = "provider"
	MissionControlRelationSourceKindCommand        MissionControlRelationSourceKind = "command"
	MissionControlRelationSourceKindVoiceCandidate MissionControlRelationSourceKind = "voice_candidate"
)

// MissionControlTimelineSourceKind describes one timeline entry source.
type MissionControlTimelineSourceKind string

const (
	MissionControlTimelineSourceKindProvider       MissionControlTimelineSourceKind = "provider"
	MissionControlTimelineSourceKindPlatform       MissionControlTimelineSourceKind = "platform"
	MissionControlTimelineSourceKindCommand        MissionControlTimelineSourceKind = "command"
	MissionControlTimelineSourceKindVoiceCandidate MissionControlTimelineSourceKind = "voice_candidate"
)

// MissionControlCommandKind identifies one provider-safe inline command.
type MissionControlCommandKind string

const (
	MissionControlCommandKindDiscussionCreate    MissionControlCommandKind = "discussion.create"
	MissionControlCommandKindWorkItemCreate      MissionControlCommandKind = "work_item.create"
	MissionControlCommandKindDiscussionFormalize MissionControlCommandKind = "discussion.formalize"
	MissionControlCommandKindStageNextStep       MissionControlCommandKind = "stage.next_step.execute"
	MissionControlCommandKindRetrySync           MissionControlCommandKind = "command.retry_sync"
)

// MissionControlCommandStatus captures one command lifecycle state.
type MissionControlCommandStatus string

const (
	MissionControlCommandStatusAccepted        MissionControlCommandStatus = "accepted"
	MissionControlCommandStatusPendingApproval MissionControlCommandStatus = "pending_approval"
	MissionControlCommandStatusQueued          MissionControlCommandStatus = "queued"
	MissionControlCommandStatusPendingSync     MissionControlCommandStatus = "pending_sync"
	MissionControlCommandStatusReconciled      MissionControlCommandStatus = "reconciled"
	MissionControlCommandStatusFailed          MissionControlCommandStatus = "failed"
	MissionControlCommandStatusBlocked         MissionControlCommandStatus = "blocked"
	MissionControlCommandStatusCancelled       MissionControlCommandStatus = "cancelled"
)

// MissionControlCommandFailureReason captures one typed command failure class.
type MissionControlCommandFailureReason string

const (
	MissionControlCommandFailureReasonProviderError   MissionControlCommandFailureReason = "provider_error"
	MissionControlCommandFailureReasonPolicyDenied    MissionControlCommandFailureReason = "policy_denied"
	MissionControlCommandFailureReasonProjectionStale MissionControlCommandFailureReason = "projection_stale"
	MissionControlCommandFailureReasonDuplicateIntent MissionControlCommandFailureReason = "duplicate_intent"
	MissionControlCommandFailureReasonTimeout         MissionControlCommandFailureReason = "timeout"
	MissionControlCommandFailureReasonApprovalDenied  MissionControlCommandFailureReason = "approval_denied"
	MissionControlCommandFailureReasonApprovalExpired MissionControlCommandFailureReason = "approval_expired"
	MissionControlCommandFailureReasonUnknown         MissionControlCommandFailureReason = "unknown"
)

// MissionControlApprovalState captures approval sub-state for one command.
type MissionControlApprovalState string

const (
	MissionControlApprovalStateNotRequired MissionControlApprovalState = "not_required"
	MissionControlApprovalStatePending     MissionControlApprovalState = "pending"
	MissionControlApprovalStateApproved    MissionControlApprovalState = "approved"
	MissionControlApprovalStateDenied      MissionControlApprovalState = "denied"
	MissionControlApprovalStateExpired     MissionControlApprovalState = "expired"
)

// MissionControlApprovalRequirement describes whether one command must pause for owner review.
type MissionControlApprovalRequirement string

const (
	MissionControlApprovalRequirementNone        MissionControlApprovalRequirement = "none"
	MissionControlApprovalRequirementOwnerReview MissionControlApprovalRequirement = "owner_review"
)

// MissionControlGapKind identifies one persisted continuity-gap category.
type MissionControlGapKind string

const (
	MissionControlGapKindMissingRun           MissionControlGapKind = "missing_run"
	MissionControlGapKindMissingPullRequest   MissionControlGapKind = "missing_pull_request"
	MissionControlGapKindMissingFollowUpIssue MissionControlGapKind = "missing_follow_up_issue"
	MissionControlGapKindProviderOutOfScope   MissionControlGapKind = "provider_out_of_scope"
	MissionControlGapKindProviderStale        MissionControlGapKind = "provider_stale"
	MissionControlGapKindOrphanNode           MissionControlGapKind = "orphan_node"
)

// MissionControlGapSeverity identifies the urgency of one continuity gap.
type MissionControlGapSeverity string

const (
	MissionControlGapSeverityBlocking MissionControlGapSeverity = "blocking"
	MissionControlGapSeverityWarning  MissionControlGapSeverity = "warning"
	MissionControlGapSeverityInfo     MissionControlGapSeverity = "info"
)

// MissionControlGapStatus captures whether a continuity gap is still open.
type MissionControlGapStatus string

const (
	MissionControlGapStatusOpen     MissionControlGapStatus = "open"
	MissionControlGapStatusResolved MissionControlGapStatus = "resolved"
	MissionControlGapStatusDeferred MissionControlGapStatus = "deferred"
)

// MissionControlWorkspaceWatermarkKind identifies one persisted workspace watermark lane.
type MissionControlWorkspaceWatermarkKind string

const (
	MissionControlWorkspaceWatermarkKindProviderFreshness MissionControlWorkspaceWatermarkKind = "provider_freshness"
	MissionControlWorkspaceWatermarkKindProviderCoverage  MissionControlWorkspaceWatermarkKind = "provider_coverage"
	MissionControlWorkspaceWatermarkKindGraphProjection   MissionControlWorkspaceWatermarkKind = "graph_projection"
	MissionControlWorkspaceWatermarkKindLaunchPolicy      MissionControlWorkspaceWatermarkKind = "launch_policy"
)

// MissionControlWorkspaceWatermarkStatus captures the effective health of one watermark.
type MissionControlWorkspaceWatermarkStatus string

const (
	MissionControlWorkspaceWatermarkStatusFresh      MissionControlWorkspaceWatermarkStatus = "fresh"
	MissionControlWorkspaceWatermarkStatusStale      MissionControlWorkspaceWatermarkStatus = "stale"
	MissionControlWorkspaceWatermarkStatusDegraded   MissionControlWorkspaceWatermarkStatus = "degraded"
	MissionControlWorkspaceWatermarkStatusOutOfScope MissionControlWorkspaceWatermarkStatus = "out_of_scope"
)

// MissionControlWorkspaceStatePreset identifies one fixed workspace active-state preset.
type MissionControlWorkspaceStatePreset string

const (
	MissionControlWorkspaceStatePresetWorking               MissionControlWorkspaceStatePreset = "working"
	MissionControlWorkspaceStatePresetWaiting               MissionControlWorkspaceStatePreset = "waiting"
	MissionControlWorkspaceStatePresetBlocked               MissionControlWorkspaceStatePreset = "blocked"
	MissionControlWorkspaceStatePresetReview                MissionControlWorkspaceStatePreset = "review"
	MissionControlWorkspaceStatePresetRecentCriticalUpdates MissionControlWorkspaceStatePreset = "recent_critical_updates"
	MissionControlWorkspaceStatePresetAllActive             MissionControlWorkspaceStatePreset = "all_active"
)

// MissionControlWorkspaceVisibilityTier identifies how one node/edge is rendered in workspace projection.
type MissionControlWorkspaceVisibilityTier string

const (
	MissionControlWorkspaceVisibilityTierPrimary         MissionControlWorkspaceVisibilityTier = "primary"
	MissionControlWorkspaceVisibilityTierSecondaryDimmed MissionControlWorkspaceVisibilityTier = "secondary_dimmed"
)
