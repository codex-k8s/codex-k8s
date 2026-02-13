package query

// StaffRunListFilter defines list filters for staff runtime views.
type StaffRunListFilter struct {
	Limit       int
	TriggerKind string
	Status      string
	AgentKey    string
	WaitState   string
}
