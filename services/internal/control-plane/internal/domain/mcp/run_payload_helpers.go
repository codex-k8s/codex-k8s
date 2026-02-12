package mcp

import (
	"encoding/json"
	"fmt"

	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func parseRunPayload(raw json.RawMessage) (querytypes.RunPayload, error) {
	if len(raw) == 0 {
		return querytypes.RunPayload{}, fmt.Errorf("run payload is empty")
	}
	var payload querytypes.RunPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return querytypes.RunPayload{}, fmt.Errorf("decode run payload: %w", err)
	}
	return payload, nil
}
