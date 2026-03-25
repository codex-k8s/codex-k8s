package mcp

import (
	"encoding/json"

	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func parseRunPayload(raw json.RawMessage) (querytypes.RunPayload, error) {
	return querytypes.DecodeRunPayload(raw)
}
