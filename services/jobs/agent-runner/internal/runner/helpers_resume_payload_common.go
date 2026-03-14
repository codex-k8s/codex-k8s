package runner

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func buildDeterministicResumePromptBlock[T any](
	locale string,
	rawPayload string,
	resume bool,
	restoredSessionMessage string,
	parse func(string) (T, error),
	ruHeading string,
	ruDescription string,
	enHeading string,
	enDescription string,
) (string, error) {
	if strings.TrimSpace(rawPayload) == "" {
		return "", nil
	}
	if !resume {
		return "", fmt.Errorf("%s", restoredSessionMessage)
	}

	payload, err := parse(rawPayload)
	if err != nil {
		return "", err
	}

	prettyPayload, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal deterministic resume payload: %w", err)
	}

	if normalizePromptLocale(locale) == promptLocaleRU {
		return strings.TrimSpace(fmt.Sprintf("%s\n%s\n\n```json\n%s\n```", ruHeading, ruDescription, prettyPayload)), nil
	}
	return strings.TrimSpace(fmt.Sprintf("%s\n%s\n\n```json\n%s\n```", enHeading, enDescription, prettyPayload)), nil
}

func parseDeterministicResumePayload[T any](
	rawPayload string,
	maxBytes int,
	payloadLabel string,
	normalize func(T) T,
	validate func(T) error,
) (T, error) {
	var zero T

	trimmed := strings.TrimSpace(rawPayload)
	if trimmed == "" {
		return zero, fmt.Errorf("%s is empty", payloadLabel)
	}
	if len([]byte(trimmed)) > maxBytes {
		return zero, fmt.Errorf("%s exceeds %d bytes", payloadLabel, maxBytes)
	}

	var payload T
	decoder := json.NewDecoder(strings.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return zero, fmt.Errorf("decode %s: %w", payloadLabel, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return zero, fmt.Errorf("%s contains trailing data", payloadLabel)
	}

	payload = normalize(payload)
	if err := validate(payload); err != nil {
		return zero, err
	}
	return payload, nil
}

func trimStringFields(fields ...*string) {
	for _, field := range fields {
		if field == nil {
			continue
		}
		*field = strings.TrimSpace(*field)
	}
}
