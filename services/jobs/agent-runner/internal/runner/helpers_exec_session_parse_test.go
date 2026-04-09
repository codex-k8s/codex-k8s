package runner

import "testing"

func TestParseCodexReportOutput_FlatSchema(t *testing.T) {
	raw := []byte(`{"summary":"ok","branch":"main","pr_number":324,"pr_url":"https://example.test/pr/324","session_id":"sess-1","model":"gpt-5.4","reasoning_effort":"high"}`)

	report, payload, err := parseCodexReportOutput(raw)
	if err != nil {
		t.Fatalf("parseCodexReportOutput() error = %v", err)
	}
	if string(payload) != string(raw) {
		t.Fatalf("payload = %s, want %s", string(payload), string(raw))
	}
	if report.Summary != "ok" {
		t.Fatalf("summary = %q, want ok", report.Summary)
	}
	if report.PRNumber != 324 {
		t.Fatalf("pr_number = %d, want 324", report.PRNumber)
	}
	if report.PRURL != "https://example.test/pr/324" {
		t.Fatalf("pr_url = %q", report.PRURL)
	}
}

func TestParseCodexReportOutput_CompletedEnvelope(t *testing.T) {
	raw := []byte(`{
		"status":"completed",
		"branch":"codex/issue-320",
		"commit":"77a833f",
		"pr":{"number":324,"url":"https://github.com/codex-k8s/kodex/pull/324"},
		"issue":{"number":320,"follow_up_issue":{"number":325,"url":"https://github.com/codex-k8s/kodex/issues/325"}},
		"summary":[
			"Удалил устаревшую проверку.",
			"Синхронизировал delivery-доки."
		],
		"checks":[
			{"name":"git diff","result":"passed","details":"git diff --check"}
		],
		"acceptance":{
			"review":"Проверьте PR",
			"approve":"Approve PR",
			"merge":"Merge PR"
		}
	}`)

	report, _, err := parseCodexReportOutput(raw)
	if err != nil {
		t.Fatalf("parseCodexReportOutput() error = %v", err)
	}
	if report.Branch != "codex/issue-320" {
		t.Fatalf("branch = %q, want codex/issue-320", report.Branch)
	}
	if report.PRNumber != 324 {
		t.Fatalf("pr_number = %d, want 324", report.PRNumber)
	}
	if report.PRURL != "https://github.com/codex-k8s/kodex/pull/324" {
		t.Fatalf("pr_url = %q", report.PRURL)
	}
	wantSummary := "Удалил устаревшую проверку.\nСинхронизировал delivery-доки."
	if report.Summary != wantSummary {
		t.Fatalf("summary = %q, want %q", report.Summary, wantSummary)
	}
}

func TestParseCodexReportOutput_LastJSONLineFallback(t *testing.T) {
	raw := []byte("noise before\n{\"status\":\"completed\",\"branch\":\"codex/issue-320\",\"pr\":{\"number\":324,\"url\":\"https://example.test/pr/324\"},\"summary\":[\"done\"]}")

	report, payload, err := parseCodexReportOutput(raw)
	if err != nil {
		t.Fatalf("parseCodexReportOutput() error = %v", err)
	}
	if string(payload) != "{\"status\":\"completed\",\"branch\":\"codex/issue-320\",\"pr\":{\"number\":324,\"url\":\"https://example.test/pr/324\"},\"summary\":[\"done\"]}" {
		t.Fatalf("unexpected payload: %s", string(payload))
	}
	if report.Summary != "done" {
		t.Fatalf("summary = %q, want done", report.Summary)
	}
	if report.PRNumber != 324 {
		t.Fatalf("pr_number = %d, want 324", report.PRNumber)
	}
}

func TestParseCodexReportOutput_StatusOKEnvelopeWithNumericPR(t *testing.T) {
	raw := []byte(`{
		"status":"ok",
		"issue":347,
		"pr":354,
		"branch":"codex/issue-347",
		"commit":"3be32de49ac5c7aac3a3581091e613fd085d28e2",
		"summary":[
			"Исправлен dev websocket proxy.",
			"PR обновлен и review threads закрыты."
		],
		"checks":[
			{"command":"git diff --check","result":"passed"}
		],
		"next_action":"Откройте PR #354."
	}`)

	report, _, err := parseCodexReportOutput(raw)
	if err != nil {
		t.Fatalf("parseCodexReportOutput() error = %v", err)
	}
	if report.Branch != "codex/issue-347" {
		t.Fatalf("branch = %q, want codex/issue-347", report.Branch)
	}
	if report.PRNumber != 354 {
		t.Fatalf("pr_number = %d, want 354", report.PRNumber)
	}
	wantSummary := "Исправлен dev websocket proxy.\nPR обновлен и review threads закрыты."
	if report.Summary != wantSummary {
		t.Fatalf("summary = %q, want %q", report.Summary, wantSummary)
	}
}
