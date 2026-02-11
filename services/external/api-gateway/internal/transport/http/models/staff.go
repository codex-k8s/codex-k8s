package models

type Project struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type Run struct {
	ID            string  `json:"id"`
	CorrelationID string  `json:"correlation_id"`
	ProjectID     *string `json:"project_id"`
	ProjectSlug   string  `json:"project_slug"`
	ProjectName   string  `json:"project_name"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	StartedAt     *string `json:"started_at"`
	FinishedAt    *string `json:"finished_at"`
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
	ID               string `json:"id"`
	ProjectID        string `json:"project_id"`
	Provider         string `json:"provider"`
	ExternalID       int64  `json:"external_id"`
	Owner            string `json:"owner"`
	Name             string `json:"name"`
	ServicesYAMLPath string `json:"services_yaml_path"`
}
