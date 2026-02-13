package dbmodel

import "time"

// ProjectDatabaseRow mirrors one project_databases row.
type ProjectDatabaseRow struct {
	ProjectID    string    `db:"project_id"`
	Environment  string    `db:"environment"`
	DatabaseName string    `db:"database_name"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
