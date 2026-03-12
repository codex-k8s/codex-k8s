package runstatus

import "testing"

func TestParseRunAgentStatusPayload_Object(t *testing.T) {
	t.Parallel()

	statusText := parseRunAgentStatusPayload([]byte(`{"status_text":"Пишу websocket transport","agent_key":"dev"}`))
	if statusText != "Пишу websocket transport" {
		t.Fatalf("unexpected status_text: %q", statusText)
	}
}

func TestParseRunAgentStatusPayload_DoubleEncodedJSON(t *testing.T) {
	t.Parallel()

	statusText := parseRunAgentStatusPayload([]byte(`"{\"status_text\":\"Running tests\",\"agent_key\":\"qa\"}"`))
	if statusText != "Running tests" {
		t.Fatalf("unexpected status_text: %q", statusText)
	}
}

func TestParseRunAgentStatusPayload_Invalid(t *testing.T) {
	t.Parallel()

	statusText := parseRunAgentStatusPayload([]byte(`not-json`))
	if statusText != "" {
		t.Fatalf("expected empty parse result, got status=%q", statusText)
	}
}
