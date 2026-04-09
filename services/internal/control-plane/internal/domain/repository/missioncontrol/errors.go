package missioncontrol

import (
	"fmt"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

// ProjectionVersionConflict indicates that an entity projection update used a stale version.
type ProjectionVersionConflict struct {
	ProjectID                 string
	EntityKind                enumtypes.MissionControlEntityKind
	EntityExternalKey         string
	ExpectedProjectionVersion int64
	ActualProjectionVersion   int64
}

func (e ProjectionVersionConflict) Error() string {
	return fmt.Sprintf(
		"mission control projection version conflict for %s/%s/%s: expected %d, actual %d",
		e.ProjectID,
		e.EntityKind,
		e.EntityExternalKey,
		e.ExpectedProjectionVersion,
		e.ActualProjectionVersion,
	)
}

// DuplicateBusinessIntent indicates that one semantic command already exists for the same project/business intent key.
type DuplicateBusinessIntent struct {
	ProjectID         string
	BusinessIntentKey string
}

func (e DuplicateBusinessIntent) Error() string {
	return fmt.Sprintf(
		"mission control duplicate business intent for %s/%s",
		e.ProjectID,
		e.BusinessIntentKey,
	)
}
