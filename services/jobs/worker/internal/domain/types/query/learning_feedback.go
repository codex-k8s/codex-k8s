package query

import enumtypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/enum"

// LearningFeedbackInsertParams defines inputs for inserting learning mode feedback.
type LearningFeedbackInsertParams struct {
	// RunID references agent_runs.id.
	RunID string
	// Kind is a feedback kind: inline or post_pr.
	Kind enumtypes.LearningFeedbackKind
	// Explanation is a learning-oriented explanation text (why/tradeoffs/alternatives).
	Explanation string
}
