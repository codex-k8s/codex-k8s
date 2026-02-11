package learningfeedback

import (
	"context"
	"time"
)

// Feedback is a persisted learning-mode explanation bound to an agent run.
type Feedback struct {
	ID           int64
	RunID        string
	RepositoryID string
	PRNumber     int
	FilePath     string
	Line         int
	Kind         string
	Explanation  string
	CreatedAt    time.Time
}

// InsertParams defines inputs for creating a feedback record.
type InsertParams struct {
	RunID        string
	RepositoryID string
	PRNumber     *int
	FilePath     *string
	Line         *int
	Kind         string
	Explanation  string
}

// Repository persists learning feedback records.
type Repository interface {
	// ListForRun returns feedback entries for a run.
	ListForRun(ctx context.Context, runID string, limit int) ([]Feedback, error)
	// Insert creates a new feedback record.
	Insert(ctx context.Context, params InsertParams) (int64, error)
}
