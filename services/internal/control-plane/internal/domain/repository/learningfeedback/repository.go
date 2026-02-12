package learningfeedback

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	Feedback     = entitytypes.LearningFeedback
	InsertParams = querytypes.LearningFeedbackInsertParams
)

// Repository persists learning feedback records.
type Repository interface {
	// ListForRun returns feedback entries for a run.
	ListForRun(ctx context.Context, runID string, limit int) ([]Feedback, error)
	// Insert creates a new feedback record.
	Insert(ctx context.Context, params InsertParams) (int64, error)
}
