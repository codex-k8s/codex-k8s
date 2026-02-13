package postgres

import (
	"context"
	"fmt"
	"time"
)

// InsertFlowEvent inserts a row into flow_events using the provided SQL query.
func InsertFlowEvent(
	ctx context.Context,
	db execer,
	query string,
	correlationID string,
	actorType string,
	actorID string,
	eventType string,
	payload []byte,
	createdAt time.Time,
) error {
	_, err := db.Exec(ctx, query, correlationID, actorType, actorID, eventType, payload, createdAt.UTC())
	if err != nil {
		return fmt.Errorf("insert flow event: %w", err)
	}
	return nil
}
