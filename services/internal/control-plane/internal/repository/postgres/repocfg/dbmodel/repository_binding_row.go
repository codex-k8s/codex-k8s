package dbmodel

// RepositoryBindingRow mirrors repository binding row selected from PostgreSQL.
type RepositoryBindingRow struct {
	ID                 string `db:"id"`
	ProjectID          string `db:"project_id"`
	Alias              string `db:"alias"`
	Role               string `db:"role"`
	DefaultRef         string `db:"default_ref"`
	Provider           string `db:"provider"`
	ExternalID         int64  `db:"external_id"`
	Owner              string `db:"owner"`
	Name               string `db:"name"`
	ServicesYAMLPath   string `db:"services_yaml_path"`
	DocsRootPath       string `db:"docs_root_path"`
	BotUsername        string `db:"bot_username"`
	BotEmail           string `db:"bot_email"`
	PreflightUpdatedAt string `db:"preflight_updated_at"`
}

// RepositoryBindingLookupRow mirrors provider lookup projection.
type RepositoryBindingLookupRow struct {
	ProjectID        string `db:"project_id"`
	RepositoryID     string `db:"repository_id"`
	ServicesYAMLPath string `db:"services_yaml_path"`
	DefaultRef       string `db:"default_ref"`
}
