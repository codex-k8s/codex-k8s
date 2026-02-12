package worker

import "context"

type noopRunStatusNotifier struct{}

func (noopRunStatusNotifier) UpsertRunStatusComment(_ context.Context, _ RunStatusCommentParams) (RunStatusCommentResult, error) {
	return RunStatusCommentResult{}, nil
}
