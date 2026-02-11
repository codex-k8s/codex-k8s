package entity

// ProjectView is a typed projection for staff project list responses.
type ProjectView struct {
	ID   string
	Slug string
	Name string
	Role string
}
