package dbmodel

// RequestStateMetricRow mirrors aggregated interaction request state counts.
type RequestStateMetricRow struct {
	InteractionKind string `db:"interaction_kind"`
	State           string `db:"state"`
	Total           int64  `db:"total"`
}

// PendingDispatchBacklogMetricRow mirrors aggregated pending dispatch backlog counts.
type PendingDispatchBacklogMetricRow struct {
	InteractionKind string `db:"interaction_kind"`
	QueueKind       string `db:"queue_kind"`
	Total           int64  `db:"total"`
}

// OverdueWaitMetricRow mirrors aggregated overdue wait counts.
type OverdueWaitMetricRow struct {
	InteractionKind string `db:"interaction_kind"`
	Total           int64  `db:"total"`
}

// CallbackEventMetricRow mirrors aggregated callback evidence counts.
type CallbackEventMetricRow struct {
	CallbackKind   string `db:"callback_kind"`
	Classification string `db:"classification"`
	Total          int64  `db:"total"`
}

// DispatchAttemptMetricRow mirrors aggregated delivery attempt counts.
type DispatchAttemptMetricRow struct {
	InteractionKind string `db:"interaction_kind"`
	AdapterKind     string `db:"adapter_kind"`
	Status          string `db:"status"`
	Total           int64  `db:"total"`
}
