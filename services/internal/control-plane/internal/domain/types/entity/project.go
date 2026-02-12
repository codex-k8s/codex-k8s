package entity

// Project is a persisted project catalog entry.
type Project struct {
	ID   string
	Slug string
	Name string
}

// ProjectWithRole extends project with an effective role for a user.
type ProjectWithRole struct {
	Project
	Role string
}
