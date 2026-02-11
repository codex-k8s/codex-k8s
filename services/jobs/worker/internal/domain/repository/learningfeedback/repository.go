package learningfeedback

import "context"

// Kind identifies learning feedback category.
type Kind string

const (
	// KindInline is feedback shown inline for a run.
	KindInline Kind = "inline"
	// KindPostPR is feedback published after PR completion.
	KindPostPR Kind = "post_pr"
)

// InsertParams defines inputs for inserting learning mode feedback.
type InsertParams struct {
	// RunID references agent_runs.id.
	RunID string
	// Kind is a feedback kind: inline or post_pr.
	Kind Kind
	// Explanation is a learning-oriented explanation text (why/tradeoffs/alternatives).
	Explanation string
}

// Repository persists learning feedback records.
type Repository interface {
	// Insert creates a new learning_feedback record.
	Insert(ctx context.Context, params InsertParams) error
}
