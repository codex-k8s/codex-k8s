package entity

import "time"

// ProjectDatabase stores ownership mapping between project and managed database.
type ProjectDatabase struct {
	ProjectID    string
	Environment  string
	DatabaseName string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
