package learningfeedback

import "context"

// InsertParams defines inputs for inserting learning mode feedback.
type InsertParams struct {
	// RunID references agent_runs.id.
	RunID string
	// Kind is a feedback kind: inline or post_pr.
	Kind string
	// Explanation is a learning-oriented explanation text (why/tradeoffs/alternatives).
	Explanation string
}

// Repository persists learning feedback records.
type Repository interface {
	// Insert creates a new learning_feedback record.
	Insert(ctx context.Context, params InsertParams) error
}
