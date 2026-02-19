package runaccess

import "encoding/json"

const payloadMarshalFailedError = "payload_marshal_failed"

type keyLifecyclePayload struct {
	RunID       string `json:"run_id"`
	ProjectID   string `json:"project_id,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	RuntimeMode string `json:"runtime_mode,omitempty"`
	TargetEnv   string `json:"target_env,omitempty"`
	Status      string `json:"status,omitempty"`
	ExpiresAt   string `json:"expires_at,omitempty"`
	IssuedBy    string `json:"issued_by,omitempty"`
	RevokedBy   string `json:"revoked_by,omitempty"`
	Reason      string `json:"reason,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

type payloadMarshalError struct {
	Error string `json:"error"`
}

func encodeKeyLifecyclePayload(payload keyLifecyclePayload) json.RawMessage {
	bytes, err := json.Marshal(payload)
	if err == nil {
		return json.RawMessage(bytes)
	}
	fallback, fallbackErr := json.Marshal(payloadMarshalError{Error: payloadMarshalFailedError})
	if fallbackErr != nil {
		return json.RawMessage(`{"error":"payload_marshal_failed"}`)
	}
	return json.RawMessage(fallback)
}
