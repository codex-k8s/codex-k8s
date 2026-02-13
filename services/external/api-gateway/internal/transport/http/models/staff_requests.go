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
