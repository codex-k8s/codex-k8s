package agentcallback

import (
	"bytes"
	"encoding/json"
	"fmt"

	sharedgithubratelimit "github.com/codex-k8s/kodex/libs/go/domain/githubratelimit"
	"github.com/codex-k8s/kodex/libs/go/mcp/userinteraction"
)

func extractInteractionResumePayload(runPayload json.RawMessage) (json.RawMessage, bool, error) {
	return extractRunPayloadJSONField(
		runPayload,
		userinteraction.ResumePayloadRunPayloadFieldName,
		userinteraction.ResumePayloadMaxBytes,
		"interaction resume payload",
	)
}

func extractGitHubRateLimitResumePayload(runPayload json.RawMessage) (json.RawMessage, bool, error) {
	return extractRunPayloadJSONField(
		runPayload,
		sharedgithubratelimit.ResumePayloadRunPayloadFieldName,
		sharedgithubratelimit.ResumePayloadMaxBytes,
		"github rate-limit resume payload",
	)
}

func extractRunPayloadJSONField(runPayload json.RawMessage, fieldName string, maxBytes int, payloadLabel string) (json.RawMessage, bool, error) {
	if len(runPayload) == 0 {
		return nil, false, nil
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(runPayload, &envelope); err != nil {
		return nil, false, fmt.Errorf("decode run payload for %s: %w", payloadLabel, err)
	}

	raw, found := envelope[fieldName]
	if !found {
		return nil, false, nil
	}

	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, false, nil
	}
	if !json.Valid(trimmed) {
		return nil, false, fmt.Errorf("%s is not valid JSON", payloadLabel)
	}
	if len(trimmed) > maxBytes {
		return nil, false, fmt.Errorf("%s exceeds %d bytes", payloadLabel, maxBytes)
	}

	var compact bytes.Buffer
	if err := json.Compact(&compact, trimmed); err != nil {
		return nil, false, fmt.Errorf("compact %s: %w", payloadLabel, err)
	}
	if compact.Len() > maxBytes {
		return nil, false, fmt.Errorf("%s exceeds %d bytes", payloadLabel, maxBytes)
	}
	return append(json.RawMessage(nil), compact.Bytes()...), true, nil
}
