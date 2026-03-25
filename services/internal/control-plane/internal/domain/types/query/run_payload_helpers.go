package query

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DecodeRunPayload decodes normalized run payload persisted in agent_runs.
func DecodeRunPayload(raw json.RawMessage) (RunPayload, error) {
	return DecodeRunPayloadInto[RunPayload](raw, "decode run payload")
}

// DecodeRunPayloadInto decodes run payload into the requested typed envelope.
func DecodeRunPayloadInto[T any](raw json.RawMessage, errorContext string) (T, error) {
	var zero T
	if len(raw) == 0 {
		return zero, fmt.Errorf("run payload is empty")
	}
	var payload T
	if err := json.Unmarshal(raw, &payload); err != nil {
		context := strings.TrimSpace(errorContext)
		if context == "" {
			context = "decode run payload"
		}
		return zero, fmt.Errorf("%s: %w", context, err)
	}
	return payload, nil
}
