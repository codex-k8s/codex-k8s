package agentcallback

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/codex-k8s/codex-k8s/libs/go/mcp/userinteraction"
)

func extractInteractionResumePayload(runPayload json.RawMessage) (json.RawMessage, bool, error) {
	if len(runPayload) == 0 {
		return nil, false, nil
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(runPayload, &envelope); err != nil {
		return nil, false, fmt.Errorf("decode run payload for interaction resume payload: %w", err)
	}

	raw, found := envelope[userinteraction.ResumePayloadRunPayloadFieldName]
	if !found {
		return nil, false, nil
	}

	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, false, nil
	}
	if !json.Valid(trimmed) {
		return nil, false, fmt.Errorf("interaction resume payload is not valid JSON")
	}
	if len(trimmed) > userinteraction.ResumePayloadMaxBytes {
		return nil, false, fmt.Errorf("interaction resume payload exceeds %d bytes", userinteraction.ResumePayloadMaxBytes)
	}

	var compact bytes.Buffer
	if err := json.Compact(&compact, trimmed); err != nil {
		return nil, false, fmt.Errorf("compact interaction resume payload: %w", err)
	}
	if compact.Len() > userinteraction.ResumePayloadMaxBytes {
		return nil, false, fmt.Errorf("interaction resume payload exceeds %d bytes", userinteraction.ResumePayloadMaxBytes)
	}
	return append(json.RawMessage(nil), compact.Bytes()...), true, nil
}
