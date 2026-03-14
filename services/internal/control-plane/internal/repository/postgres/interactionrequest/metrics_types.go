package interactionrequest

import "time"

// MetricsSnapshot groups persisted interaction observability counters and gauges.
type MetricsSnapshot struct {
	CollectedAt            time.Time
	RequestStateTotals     []RequestStateTotal
	PendingDispatchBacklog []PendingDispatchBacklog
	OverdueWaitTotals      []OverdueWaitTotal
	CallbackEventTotals    []CallbackEventTotal
	DispatchAttemptTotals  []DispatchAttemptTotal
}

// RequestStateTotal stores one current interaction_requests aggregate count.
type RequestStateTotal struct {
	InteractionKind string
	State           string
	Total           int64
}

// PendingDispatchBacklog stores one current pending dispatch backlog count.
type PendingDispatchBacklog struct {
	InteractionKind string
	QueueKind       string
	Total           int64
}

// OverdueWaitTotal stores count of decision interactions past deadline without terminal outcome.
type OverdueWaitTotal struct {
	InteractionKind string
	Total           int64
}

// CallbackEventTotal stores one persisted callback evidence counter.
type CallbackEventTotal struct {
	CallbackKind   string
	Classification string
	Total          int64
}

// DispatchAttemptTotal stores one persisted delivery attempt counter.
type DispatchAttemptTotal struct {
	InteractionKind string
	AdapterKind     string
	Status          string
	Total           int64
}
