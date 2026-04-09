package value

import (
	"time"

	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

// MissionControlWorkspaceProjectionSummary captures owner-owned continuity refresh evidence and rollout gates.
type MissionControlWorkspaceProjectionSummary struct {
	ProjectID                    string    `json:"project_id"`
	EntityCount                  int       `json:"entity_count"`
	RootCount                    int       `json:"root_count"`
	NodeCount                    int       `json:"node_count"`
	OpenGapCount                 int       `json:"open_gap_count"`
	BlockingGapCount             int       `json:"blocking_gap_count"`
	WarningGapCount              int       `json:"warning_gap_count"`
	MissingPullRequestGapCount   int       `json:"missing_pull_request_gap_count"`
	MissingFollowUpIssueGapCount int       `json:"missing_follow_up_issue_gap_count"`
	WatermarkCount               int       `json:"watermark_count"`
	ReadyForReconcile            bool      `json:"ready_for_reconcile"`
	GatingReason                 string    `json:"gating_reason,omitempty"`
	ObservedAt                   time.Time `json:"observed_at"`
}

// MissionControlWorkspaceSummary captures aggregate counters for one snapshot.
type MissionControlWorkspaceSummary struct {
	RootCount                  int `json:"root_count"`
	NodeCount                  int `json:"node_count"`
	BlockingGapCount           int `json:"blocking_gap_count"`
	WarningGapCount            int `json:"warning_gap_count"`
	RecentClosedContextCount   int `json:"recent_closed_context_count"`
	WorkingCount               int `json:"working_count"`
	WaitingCount               int `json:"waiting_count"`
	BlockedCount               int `json:"blocked_count"`
	ReviewCount                int `json:"review_count"`
	RecentCriticalUpdatesCount int `json:"recent_critical_updates_count"`
	SecondaryDimmedNodeCount   int `json:"secondary_dimmed_node_count"`
}

// MissionControlWorkspaceRootGroup groups graph nodes under one workspace root.
type MissionControlWorkspaceRootGroup struct {
	RootNodeRef      MissionControlEntityRef   `json:"root_node_ref"`
	RootTitle        string                    `json:"root_title"`
	NodeRefs         []MissionControlEntityRef `json:"node_refs,omitempty"`
	HasBlockingGap   bool                      `json:"has_blocking_gap"`
	LatestActivityAt *time.Time                `json:"latest_activity_at,omitempty"`
}

// MissionControlWorkspaceNode represents one typed graph workspace node.
type MissionControlWorkspaceNode struct {
	NodeRef           MissionControlEntityRef                         `json:"node_ref"`
	Title             string                                          `json:"title"`
	VisibilityTier    enumtypes.MissionControlWorkspaceVisibilityTier `json:"visibility_tier"`
	ActiveState       enumtypes.MissionControlActiveState             `json:"active_state"`
	ContinuityStatus  enumtypes.MissionControlContinuityStatus        `json:"continuity_status"`
	CoverageClass     enumtypes.MissionControlCoverageClass           `json:"coverage_class"`
	ProviderReference *MissionControlProviderReferenceView            `json:"provider_reference,omitempty"`
	RootNodePublicID  string                                          `json:"root_node_public_id"`
	ColumnIndex       int32                                           `json:"column_index"`
	LastActivityAt    *time.Time                                      `json:"last_activity_at,omitempty"`
	HasBlockingGap    bool                                            `json:"has_blocking_gap"`
	Badges            []string                                        `json:"badges,omitempty"`
	ProjectionVersion int64                                           `json:"projection_version"`
}

// MissionControlWorkspaceEdge represents one typed graph workspace edge.
type MissionControlWorkspaceEdge struct {
	RelationKind   enumtypes.MissionControlRelationKind            `json:"relation_kind"`
	SourceNodeRef  MissionControlEntityRef                         `json:"source_node_ref"`
	TargetNodeRef  MissionControlEntityRef                         `json:"target_node_ref"`
	VisibilityTier enumtypes.MissionControlWorkspaceVisibilityTier `json:"visibility_tier"`
	SourceOfTruth  enumtypes.MissionControlRelationSourceKind      `json:"source_of_truth"`
	IsPrimaryPath  bool                                            `json:"is_primary_path"`
}

// MissionControlWorkspaceSnapshot captures one read-only workspace projection slice.
type MissionControlWorkspaceSnapshot struct {
	Summary             MissionControlWorkspaceSummary                 `json:"summary"`
	WorkspaceWatermarks []entitytypes.MissionControlWorkspaceWatermark `json:"workspace_watermarks,omitempty"`
	RootGroups          []MissionControlWorkspaceRootGroup             `json:"root_groups,omitempty"`
	Nodes               []MissionControlWorkspaceNode                  `json:"nodes,omitempty"`
	Edges               []MissionControlWorkspaceEdge                  `json:"edges,omitempty"`
}

// MissionControlLaunchPreviewLabelDiff captures read-only label mutation preview.
type MissionControlLaunchPreviewLabelDiff struct {
	RemovedLabels []string `json:"removed_labels,omitempty"`
	AddedLabels   []string `json:"added_labels,omitempty"`
	FinalLabels   []string `json:"final_labels,omitempty"`
}

// MissionControlLaunchPreviewContinuityEffect captures the continuity effect of one preview.
type MissionControlLaunchPreviewContinuityEffect struct {
	ResolvedGapIDs    []int64                   `json:"resolved_gap_ids,omitempty"`
	RemainingGapIDs   []int64                   `json:"remaining_gap_ids,omitempty"`
	ResultingNodeRefs []MissionControlEntityRef `json:"resulting_node_refs,omitempty"`
	ProviderRedirects []string                  `json:"provider_redirects,omitempty"`
}

// MissionControlLaunchPreview captures one deterministic read-only preview.
type MissionControlLaunchPreview struct {
	PreviewID           string                                      `json:"preview_id"`
	ApprovalRequirement enumtypes.MissionControlApprovalRequirement `json:"approval_requirement"`
	LabelDiff           MissionControlLaunchPreviewLabelDiff        `json:"label_diff"`
	ContinuityEffect    MissionControlLaunchPreviewContinuityEffect `json:"continuity_effect"`
	BlockingReason      string                                      `json:"blocking_reason,omitempty"`
}

// MissionControlContinuityGapView exposes one typed continuity gap with public node refs.
type MissionControlContinuityGapView struct {
	GapID              int64                               `json:"gap_id"`
	GapKind            enumtypes.MissionControlGapKind     `json:"gap_kind"`
	Severity           enumtypes.MissionControlGapSeverity `json:"severity"`
	Status             enumtypes.MissionControlGapStatus   `json:"status"`
	SubjectNodeRef     MissionControlEntityRef             `json:"subject_node_ref"`
	ExpectedNodeKind   enumtypes.MissionControlEntityKind  `json:"expected_node_kind"`
	ExpectedStageLabel string                              `json:"expected_stage_label,omitempty"`
	DetectedAt         time.Time                           `json:"detected_at"`
	ResolvedAt         *time.Time                          `json:"resolved_at,omitempty"`
	ResolutionHint     string                              `json:"resolution_hint,omitempty"`
}

// MissionControlStageNextStepTemplate exposes one read-only command template for preview_next_stage.
type MissionControlStageNextStepTemplate struct {
	ThreadKind          string                                      `json:"thread_kind"`
	ThreadNumber        int                                         `json:"thread_number"`
	TargetLabel         string                                      `json:"target_label"`
	RemovedLabels       []string                                    `json:"removed_labels,omitempty"`
	DisplayVariant      string                                      `json:"display_variant,omitempty"`
	ApprovalRequirement enumtypes.MissionControlApprovalRequirement `json:"approval_requirement"`
	ExpectedGapIDs      []int64                                     `json:"expected_gap_ids,omitempty"`
}

// MissionControlLaunchSurface exposes one launch/inspect affordance for workspace nodes.
type MissionControlLaunchSurface struct {
	ActionKind          string                                      `json:"action_kind"`
	Presentation        string                                      `json:"presentation"`
	ApprovalRequirement enumtypes.MissionControlApprovalRequirement `json:"approval_requirement"`
	BlockedReason       string                                      `json:"blocked_reason,omitempty"`
	CommandTemplate     *MissionControlStageNextStepTemplate        `json:"command_template,omitempty"`
}
