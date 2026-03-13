package missioncontrol

import (
	"fmt"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

// DuplicateIntentError indicates that one existing command already owns the requested business_intent_key.
type DuplicateIntentError struct {
	ProjectID         string
	BusinessIntentKey string
	ExistingCommand   entitytypes.MissionControlCommand
}

func (e DuplicateIntentError) Error() string {
	return fmt.Sprintf(
		"mission control duplicate intent for %s/%s (existing command %s)",
		e.ProjectID,
		e.BusinessIntentKey,
		e.ExistingCommand.ID,
	)
}
