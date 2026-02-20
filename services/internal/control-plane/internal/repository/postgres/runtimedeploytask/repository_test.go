package runtimedeploytask

import (
	"testing"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

func TestParseRuntimeDeployStatus_Canceled(t *testing.T) {
	t.Parallel()

	got, err := parseRuntimeDeployStatus("canceled")
	if err != nil {
		t.Fatalf("parseRuntimeDeployStatus() error = %v", err)
	}
	if got != entitytypes.RuntimeDeployTaskStatusCanceled {
		t.Fatalf("unexpected status: got %q want %q", got, entitytypes.RuntimeDeployTaskStatusCanceled)
	}
}
