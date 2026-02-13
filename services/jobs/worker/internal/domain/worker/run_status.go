package worker

import "context"

// RunStatusPhase identifies one run status transition mirrored to issue comment.
type RunStatusPhase string

const (
	RunStatusPhaseStarted          RunStatusPhase = "started"
	RunStatusPhaseFinished         RunStatusPhase = "finished"
	RunStatusPhaseNamespaceDeleted RunStatusPhase = "namespace_deleted"
)

// RunStatusCommentParams describes one run status comment upsert request.
type RunStatusCommentParams struct {
	RunID           string
	Phase           RunStatusPhase
	JobName         string
	JobNamespace    string
	RuntimeMode     string
	Namespace       string
	TriggerKind     string
	PromptLocale    string
	Model           string
	ReasoningEffort string
	RunStatus       string
	Deleted         bool
	AlreadyDeleted  bool
}

// RunStatusCommentResult is a worker-side run status update response.
type RunStatusCommentResult struct {
	CommentID  int64
	CommentURL string
}

// RunStatusNotifier updates one per-run status comment through control-plane.
type RunStatusNotifier interface {
	UpsertRunStatusComment(ctx context.Context, params RunStatusCommentParams) (RunStatusCommentResult, error)
}
