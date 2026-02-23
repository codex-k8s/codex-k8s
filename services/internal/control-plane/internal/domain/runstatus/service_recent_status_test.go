package runstatus

import "testing"

func TestParseRunAgentStatusPayload_Object(t *testing.T) {
	t.Parallel()

	statusText, agentKey := parseRunAgentStatusPayload([]byte(`{"status_text":"Пишу websocket transport","agent_key":"dev"}`))
	if statusText != "Пишу websocket transport" {
		t.Fatalf("unexpected status_text: %q", statusText)
	}
	if agentKey != "dev" {
		t.Fatalf("unexpected agent_key: %q", agentKey)
	}
}

func TestParseRunAgentStatusPayload_DoubleEncodedJSON(t *testing.T) {
	t.Parallel()

	statusText, agentKey := parseRunAgentStatusPayload([]byte(`"{\"status_text\":\"Running tests\",\"agent_key\":\"qa\"}"`))
	if statusText != "Running tests" {
		t.Fatalf("unexpected status_text: %q", statusText)
	}
	if agentKey != "qa" {
		t.Fatalf("unexpected agent_key: %q", agentKey)
	}
}

func TestParseRunAgentStatusPayload_Invalid(t *testing.T) {
	t.Parallel()

	statusText, agentKey := parseRunAgentStatusPayload([]byte(`not-json`))
	if statusText != "" || agentKey != "" {
		t.Fatalf("expected empty parse result, got status=%q agent=%q", statusText, agentKey)
	}
}
