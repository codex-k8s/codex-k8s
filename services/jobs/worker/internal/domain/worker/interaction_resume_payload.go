package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const interactionResumePayloadFieldName = "interaction_resume_payload"

func extractInteractionResumePayloadJSON(runPayload json.RawMessage) (string, error) {
	if len(runPayload) == 0 {
		return "", nil
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(runPayload, &envelope); err != nil {
		return "", fmt.Errorf("decode run payload for interaction resume payload: %w", err)
	}

	raw, found := envelope[interactionResumePayloadFieldName]
	if !found {
		return "", nil
	}

	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return "", nil
	}
	if !json.Valid(trimmed) {
		return "", fmt.Errorf("interaction resume payload is not valid JSON")
	}

	var compact bytes.Buffer
	if err := json.Compact(&compact, trimmed); err != nil {
		return "", fmt.Errorf("compact interaction resume payload: %w", err)
	}
	return compact.String(), nil
}
