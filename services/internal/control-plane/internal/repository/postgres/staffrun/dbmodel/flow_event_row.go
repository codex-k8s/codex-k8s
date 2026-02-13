package dbmodel

import "time"

// FlowEventRow mirrors one flow event selected for staff run details.
type FlowEventRow struct {
	CorrelationID string    `db:"correlation_id"`
	EventType     string    `db:"event_type"`
	CreatedAt     time.Time `db:"created_at"`
	PayloadText   string    `db:"payload"`
}
