package learningfeedback

import (
	"context"

	enumtypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/query"
)

type (
	Kind         = enumtypes.LearningFeedbackKind
	InsertParams = querytypes.LearningFeedbackInsertParams
)

const (
	KindInline = enumtypes.LearningFeedbackKindInline
	KindPostPR = enumtypes.LearningFeedbackKindPostPR
)

// Repository persists learning feedback records.
type Repository interface {
	// Insert creates a new learning_feedback record.
	Insert(ctx context.Context, params InsertParams) error
}
