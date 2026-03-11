package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// ComputeSessionSnapshotChecksum returns a stable checksum for persisted session snapshots.
func ComputeSessionSnapshotChecksum(sessionJSON []byte, codexSessionJSON []byte) (string, error) {
	canonicalCodex, err := canonicalJSONValue(codexSessionJSON, nil)
	if err != nil {
		return "", err
	}
	canonicalSession, err := canonicalJSONValue(sessionJSON, map[string]any{})
	if err != nil {
		return "", err
	}

	payload := map[string]any{
		"codex_cli_session_json": canonicalCodex,
		"session_json":           canonicalSession,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func canonicalJSONValue(raw []byte, fallback any) (any, error) {
	if len(raw) == 0 {
		return fallback, nil
	}

	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}
