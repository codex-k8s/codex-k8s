package models

// UpsertProjectRequest is a typed payload for project create/update.
type UpsertProjectRequest struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// CreateUserRequest is a typed payload for creating an allowlisted user.
type CreateUserRequest struct {
	Email           string `json:"email"`
	IsPlatformAdmin bool   `json:"is_platform_admin"`
}

// UpsertProjectMemberRequest is a typed payload for project membership upsert.
type UpsertProjectMemberRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// SetProjectMemberLearningModeRequest is a typed payload for learning mode override.
type SetProjectMemberLearningModeRequest struct {
	// Enabled can be true/false or null to inherit project default.
	Enabled *bool `json:"enabled"`
}

// UpsertProjectRepositoryRequest is a typed payload for repository binding upsert.
type UpsertProjectRepositoryRequest struct {
	Provider         string `json:"provider"`
	Owner            string `json:"owner"`
	Name             string `json:"name"`
	Token            string `json:"token"`
	ServicesYAMLPath string `json:"services_yaml_path"`
}

// ResolveApprovalDecisionRequest is a typed payload for pending approval resolution.
type ResolveApprovalDecisionRequest struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// DeleteRegistryImageTagRequest identifies one registry tag for deletion.
type DeleteRegistryImageTagRequest struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}

// CleanupRegistryImagesRequest configures bulk registry image cleanup.
type CleanupRegistryImagesRequest struct {
	RepositoryPrefix  string `json:"repository_prefix"`
	LimitRepositories int32  `json:"limit_repositories"`
	KeepTags          int32  `json:"keep_tags"`
	DryRun            bool   `json:"dry_run"`
}

type UpsertRepositoryBotParamsRequest struct {
	BotToken    *string `json:"bot_token"`
	BotUsername *string `json:"bot_username"`
	BotEmail    *string `json:"bot_email"`
}

type UpsertProjectGitHubTokensRequest struct {
	PlatformToken *string `json:"platform_token"`
	BotToken      *string `json:"bot_token"`
	BotUsername   *string `json:"bot_username"`
	BotEmail      *string `json:"bot_email"`
}

type TransitionIssueStageLabelRequest struct {
	RepositoryFullName string `json:"repository_full_name"`
	IssueNumber        int32  `json:"issue_number"`
	TargetLabel        string `json:"target_label"`
}

type UpsertConfigEntryRequest struct {
	Scope              string   `json:"scope"`
	Kind               string   `json:"kind"`
	ProjectID          *string  `json:"project_id"`
	RepositoryID       *string  `json:"repository_id"`
	Key                string   `json:"key"`
	ValuePlain         *string  `json:"value_plain"`
	ValueSecret        *string  `json:"value_secret"`
	SyncTargets        []string `json:"sync_targets"`
	Mutability         string   `json:"mutability"`
	IsDangerous        bool     `json:"is_dangerous"`
	DangerousConfirmed bool     `json:"dangerous_confirmed"`
}

type ImportDocsetRequest struct {
	RepositoryID string   `json:"repository_id"`
	DocsetRef    string   `json:"docset_ref"`
	Locale       string   `json:"locale"`
	GroupIDs     []string `json:"group_ids"`
}

type SyncDocsetRequest struct {
	RepositoryID string `json:"repository_id"`
	DocsetRef    string `json:"docset_ref"`
}
