package enum

// LearningFeedbackKind identifies learning feedback category.
type LearningFeedbackKind string

const (
	// LearningFeedbackKindInline is feedback shown inline for a run.
	LearningFeedbackKindInline LearningFeedbackKind = "inline"
	// LearningFeedbackKindPostPR is feedback published after PR completion.
	LearningFeedbackKindPostPR LearningFeedbackKind = "post_pr"
)
