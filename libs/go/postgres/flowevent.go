package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type sqlExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// InsertFlowEvent inserts a row into flow_events using the provided SQL query.
func InsertFlowEvent(
	ctx context.Context,
	db sqlExecer,
	query string,
	correlationID string,
	actorType string,
	actorID string,
	eventType string,
	payload []byte,
	createdAt time.Time,
) error {
	_, err := db.ExecContext(ctx, query, correlationID, actorType, actorID, eventType, payload, createdAt.UTC())
	if err != nil {
		return fmt.Errorf("insert flow event: %w", err)
	}
	return nil
}

