package runtimedeploy

import (
	"errors"
	"fmt"
	"strings"
)

// ErrTaskCanceled is returned when runtime deploy task was explicitly canceled.
var ErrTaskCanceled = errors.New("runtime deploy task canceled")

// TaskCanceledError carries cancel reason for one runtime deploy task.
type TaskCanceledError struct {
	RunID  string
	Reason string
}

func (e TaskCanceledError) Error() string {
	runID := strings.TrimSpace(e.RunID)
	reason := strings.TrimSpace(e.Reason)
	if runID == "" {
		if reason == "" {
			return ErrTaskCanceled.Error()
		}
		return fmt.Sprintf("%s: %s", ErrTaskCanceled.Error(), reason)
	}
	if reason == "" {
		return fmt.Sprintf("%s for run_id=%s", ErrTaskCanceled.Error(), runID)
	}
	return fmt.Sprintf("%s for run_id=%s: %s", ErrTaskCanceled.Error(), runID, reason)
}

func (e TaskCanceledError) Unwrap() error {
	return ErrTaskCanceled
}
