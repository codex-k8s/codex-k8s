package query

// LearningFeedbackInsertParams defines inputs for creating a feedback record.
type LearningFeedbackInsertParams struct {
	RunID        string
	RepositoryID string
	PRNumber     *int
	FilePath     *string
	Line         *int
	Kind         string
	Explanation  string
}
