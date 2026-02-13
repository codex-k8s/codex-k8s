package query

// ProjectDatabaseUpsertParams describes one project_databases upsert.
type ProjectDatabaseUpsertParams struct {
	ProjectID    string
	Environment  string
	DatabaseName string
}
