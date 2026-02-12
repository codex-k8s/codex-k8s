package entity

import "time"

// LearningFeedback is a persisted learning-mode explanation bound to an agent run.
type LearningFeedback struct {
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
