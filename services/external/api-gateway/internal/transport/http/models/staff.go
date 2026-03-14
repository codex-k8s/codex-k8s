package models

type Project struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type Run struct {
	ID              string             `json:"id"`
	CorrelationID   string             `json:"correlation_id"`
	ProjectID       *string            `json:"project_id"`
	ProjectSlug     string             `json:"project_slug"`
	ProjectName     string             `json:"project_name"`
	IssueNumber     *int32             `json:"issue_number"`
	IssueURL        *string            `json:"issue_url"`
	PRNumber        *int32             `json:"pr_number"`
	PRURL           *string            `json:"pr_url"`
	TriggerKind     *string            `json:"trigger_kind"`
	TriggerLabel    *string            `json:"trigger_label"`
	AgentKey        *string            `json:"agent_key"`
	JobName         *string            `json:"job_name"`
	JobNamespace    *string            `json:"job_namespace"`
	Namespace       *string            `json:"namespace"`
	JobExists       bool               `json:"job_exists"`
	NamespaceExists bool               `json:"namespace_exists"`
	WaitState       *string            `json:"wait_state"`
	WaitReason      *string            `json:"wait_reason"`
	WaitSince       *string            `json:"wait_since"`
	LastHeartbeatAt *string            `json:"last_heartbeat_at"`
	Status          string             `json:"status"`
	CreatedAt       string             `json:"created_at"`
	StartedAt       *string            `json:"started_at"`
	FinishedAt      *string            `json:"finished_at"`
	WaitProjection  *RunWaitProjection `json:"wait_projection,omitempty"`
}

type RunWaitProjection struct {
	WaitState          string                    `json:"wait_state"`
	WaitReason         string                    `json:"wait_reason"`
	DominantWait       GitHubRateLimitWaitItem   `json:"dominant_wait"`
	RelatedWaits       []GitHubRateLimitWaitItem `json:"related_waits"`
	CommentMirrorState string                    `json:"comment_mirror_state"`
}

type GitHubRateLimitWaitItem struct {
	WaitID          string                       `json:"wait_id"`
	ContourKind     string                       `json:"contour_kind"`
	LimitKind       string                       `json:"limit_kind"`
	OperationClass  string                       `json:"operation_class"`
	State           string                       `json:"state"`
	Confidence      string                       `json:"confidence"`
	EnteredAt       string                       `json:"entered_at"`
	ResumeNotBefore *string                      `json:"resume_not_before,omitempty"`
	AttemptsUsed    int32                        `json:"attempts_used"`
	MaxAttempts     int32                        `json:"max_attempts"`
	RecoveryHint    GitHubRateLimitRecoveryHint  `json:"recovery_hint"`
	ManualAction    *GitHubRateLimitManualAction `json:"manual_action,omitempty"`
}

type GitHubRateLimitRecoveryHint struct {
	HintKind        string  `json:"hint_kind"`
	ResumeNotBefore *string `json:"resume_not_before,omitempty"`
	SourceHeaders   string  `json:"source_headers"`
	DetailsMarkdown string  `json:"details_markdown"`
}

type GitHubRateLimitManualAction struct {
	Kind               string  `json:"kind"`
	Summary            string  `json:"summary"`
	DetailsMarkdown    string  `json:"details_markdown"`
	SuggestedNotBefore *string `json:"suggested_not_before,omitempty"`
}

type RunWaitResolution struct {
	WaitID         string `json:"wait_id"`
	ContourKind    string `json:"contour_kind"`
	ResolutionKind string `json:"resolution_kind"`
	ResolvedAt     string `json:"resolved_at"`
}

type RunWaitManualActionEvent struct {
	WaitID       string                      `json:"wait_id"`
	ManualAction GitHubRateLimitManualAction `json:"manual_action"`
	UpdatedAt    string                      `json:"updated_at"`
}

type RunLogs struct {
	RunID        string   `json:"run_id"`
	Status       string   `json:"status"`
	UpdatedAt    *string  `json:"updated_at"`
	SnapshotJSON string   `json:"snapshot_json"`
	TailLines    []string `json:"tail_lines"`
}

type RuntimeDeployTaskLog struct {
	Stage     string  `json:"stage"`
	Level     string  `json:"level"`
	Message   string  `json:"message"`
	CreatedAt *string `json:"created_at"`
}

type RuntimeDeployTaskListItem struct {
	RunID              string  `json:"run_id"`
	Status             string  `json:"status"`
	RepositoryFullName string  `json:"repository_full_name"`
	TargetEnv          string  `json:"target_env"`
	ResultTargetEnv    *string `json:"result_target_env"`
	Namespace          string  `json:"namespace"`
	ResultNamespace    *string `json:"result_namespace"`
	RuntimeMode        string  `json:"runtime_mode"`
	BuildRef           string  `json:"build_ref"`
	CreatedAt          *string `json:"created_at"`
	UpdatedAt          *string `json:"updated_at"`
}

type RuntimeDeployTaskActionResponse struct {
	RunID           string `json:"run_id"`
	Action          string `json:"action"`
	PreviousStatus  string `json:"previous_status"`
	CurrentStatus   string `json:"current_status"`
	AlreadyTerminal bool   `json:"already_terminal"`
}

type RuntimeDeployTask struct {
	RunID                string                 `json:"run_id"`
	RuntimeMode          string                 `json:"runtime_mode"`
	Namespace            string                 `json:"namespace"`
	TargetEnv            string                 `json:"target_env"`
	SlotNo               int32                  `json:"slot_no"`
	RepositoryFullName   string                 `json:"repository_full_name"`
	ServicesYAMLPath     string                 `json:"services_yaml_path"`
	BuildRef             string                 `json:"build_ref"`
	DeployOnly           bool                   `json:"deploy_only"`
	Status               string                 `json:"status"`
	LeaseOwner           *string                `json:"lease_owner"`
	LeaseUntil           *string                `json:"lease_until"`
	Attempts             int32                  `json:"attempts"`
	LastError            *string                `json:"last_error"`
	ResultNamespace      *string                `json:"result_namespace"`
	ResultTargetEnv      *string                `json:"result_target_env"`
	CancelRequestedAt    *string                `json:"cancel_requested_at"`
	CancelRequestedBy    *string                `json:"cancel_requested_by"`
	CancelReason         *string                `json:"cancel_reason"`
	StopRequestedAt      *string                `json:"stop_requested_at"`
	StopRequestedBy      *string                `json:"stop_requested_by"`
	StopReason           *string                `json:"stop_reason"`
	TerminalStatusSource *string                `json:"terminal_status_source"`
	TerminalEventSeq     int64                  `json:"terminal_event_seq"`
	CreatedAt            *string                `json:"created_at"`
	UpdatedAt            *string                `json:"updated_at"`
	StartedAt            *string                `json:"started_at"`
	FinishedAt           *string                `json:"finished_at"`
	Logs                 []RuntimeDeployTaskLog `json:"logs"`
}

type RuntimeError struct {
	ID            string  `json:"id"`
	Source        string  `json:"source"`
	Level         string  `json:"level"`
	Message       string  `json:"message"`
	DetailsJSON   string  `json:"details_json"`
	StackTrace    *string `json:"stack_trace"`
	CorrelationID *string `json:"correlation_id"`
	RunID         *string `json:"run_id"`
	ProjectID     *string `json:"project_id"`
	Namespace     *string `json:"namespace"`
	JobName       *string `json:"job_name"`
	ViewedAt      *string `json:"viewed_at"`
	ViewedBy      *string `json:"viewed_by"`
	CreatedAt     string  `json:"created_at"`
}

type RegistryImageTag struct {
	Tag             string  `json:"tag"`
	Digest          string  `json:"digest"`
	CreatedAt       *string `json:"created_at"`
	ConfigSizeBytes int64   `json:"config_size_bytes"`
}

type RegistryImageRepository struct {
	Repository string             `json:"repository"`
	TagCount   int32              `json:"tag_count"`
	Tags       []RegistryImageTag `json:"tags"`
}

type RegistryImageDeleteResult struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Digest     string `json:"digest"`
	Deleted    bool   `json:"deleted"`
}

type CleanupRegistryImagesResponse struct {
	RepositoriesScanned int32                       `json:"repositories_scanned"`
	TagsDeleted         int32                       `json:"tags_deleted"`
	TagsSkipped         int32                       `json:"tags_skipped"`
	Deleted             []RegistryImageDeleteResult `json:"deleted"`
	Skipped             []RegistryImageDeleteResult `json:"skipped"`
}

type ApprovalRequest struct {
	ID            int64   `json:"id"`
	CorrelationID string  `json:"correlation_id"`
	RunID         *string `json:"run_id"`
	ProjectID     *string `json:"project_id"`
	ProjectSlug   *string `json:"project_slug"`
	ProjectName   *string `json:"project_name"`
	IssueNumber   *int32  `json:"issue_number"`
	PRNumber      *int32  `json:"pr_number"`
	TriggerLabel  *string `json:"trigger_label"`
	ToolName      string  `json:"tool_name"`
	Action        string  `json:"action"`
	ApprovalMode  string  `json:"approval_mode"`
	RequestedBy   string  `json:"requested_by"`
	CreatedAt     string  `json:"created_at"`
}

type ResolveApprovalDecisionResponse struct {
	ID            int64   `json:"id"`
	CorrelationID string  `json:"correlation_id"`
	RunID         *string `json:"run_id"`
	ToolName      string  `json:"tool_name"`
	Action        string  `json:"action"`
	ApprovalState string  `json:"approval_state"`
}

type FlowEvent struct {
	CorrelationID string `json:"correlation_id"`
	EventType     string `json:"event_type"`
	CreatedAt     string `json:"created_at"`
	PayloadJSON   string `json:"payload_json"`
}

type LearningFeedback struct {
	ID           int64   `json:"id"`
	RunID        string  `json:"run_id"`
	RepositoryID *string `json:"repository_id"`
	PRNumber     *int32  `json:"pr_number"`
	FilePath     *string `json:"file_path"`
	Line         *int32  `json:"line"`
	Kind         string  `json:"kind"`
	Explanation  string  `json:"explanation"`
	CreatedAt    string  `json:"created_at"`
}

type User struct {
	ID              string  `json:"id"`
	Email           string  `json:"email"`
	GitHubUserID    *int64  `json:"github_user_id"`
	GitHubLogin     *string `json:"github_login"`
	IsPlatformAdmin bool    `json:"is_platform_admin"`
	IsPlatformOwner bool    `json:"is_platform_owner"`
}

type ProjectMember struct {
	ProjectID            string `json:"project_id"`
	UserID               string `json:"user_id"`
	Email                string `json:"email"`
	Role                 string `json:"role"`
	LearningModeOverride *bool  `json:"learning_mode_override"`
}

type RepositoryBinding struct {
	ID                 string  `json:"id"`
	ProjectID          string  `json:"project_id"`
	Alias              string  `json:"alias"`
	Role               string  `json:"role"`
	DefaultRef         string  `json:"default_ref"`
	Provider           string  `json:"provider"`
	ExternalID         int64   `json:"external_id"`
	Owner              string  `json:"owner"`
	Name               string  `json:"name"`
	ServicesYAMLPath   string  `json:"services_yaml_path"`
	DocsRootPath       *string `json:"docs_root_path"`
	BotUsername        *string `json:"bot_username"`
	BotEmail           *string `json:"bot_email"`
	PreflightUpdatedAt *string `json:"preflight_updated_at"`
}

type ProjectGitHubTokens struct {
	ProjectID        string  `json:"project_id"`
	HasPlatformToken bool    `json:"has_platform_token"`
	HasBotToken      bool    `json:"has_bot_token"`
	BotUsername      *string `json:"bot_username"`
	BotEmail         *string `json:"bot_email"`
}

type NextStepActionResponse struct {
	RepositoryFullName string   `json:"repository_full_name"`
	ThreadKind         string   `json:"thread_kind"`
	ThreadNumber       int32    `json:"thread_number"`
	ThreadURL          *string  `json:"thread_url"`
	RemovedLabels      []string `json:"removed_labels"`
	AddedLabels        []string `json:"added_labels"`
	FinalLabels        []string `json:"final_labels"`
}

type PreflightCheckResult struct {
	Name    string  `json:"name"`
	Status  string  `json:"status"`
	Details *string `json:"details"`
}

type RunRepositoryPreflightResponse struct {
	RepositoryID string                 `json:"repository_id"`
	Status       string                 `json:"status"`
	Checks       []PreflightCheckResult `json:"checks"`
	ReportJSON   string                 `json:"report_json"`
	FinishedAt   string                 `json:"finished_at"`
}

type ConfigEntry struct {
	ID           string   `json:"id"`
	Scope        string   `json:"scope"`
	Kind         string   `json:"kind"`
	ProjectID    *string  `json:"project_id"`
	RepositoryID *string  `json:"repository_id"`
	Key          string   `json:"key"`
	Value        *string  `json:"value"`
	SyncTargets  []string `json:"sync_targets"`
	Mutability   string   `json:"mutability"`
	IsDangerous  bool     `json:"is_dangerous"`
	UpdatedAt    *string  `json:"updated_at"`
}

type DocsetGroup struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	DefaultSelected bool   `json:"default_selected"`
}

type DocsetGroupItemsResponse struct {
	Groups []DocsetGroup `json:"groups"`
}

type ImportDocsetResponse struct {
	RepositoryFullName string `json:"repository_full_name"`
	PRNumber           int32  `json:"pr_number"`
	PRURL              string `json:"pr_url"`
	Branch             string `json:"branch"`
	FilesTotal         int32  `json:"files_total"`
}

type SyncDocsetResponse struct {
	RepositoryFullName string `json:"repository_full_name"`
	PRNumber           int32  `json:"pr_number"`
	PRURL              string `json:"pr_url"`
	Branch             string `json:"branch"`
	FilesUpdated       int32  `json:"files_updated"`
	FilesDrift         int32  `json:"files_drift"`
}
