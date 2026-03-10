package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCodexExec_NonResumePassesExecFlagsAndReadsOutputFile(t *testing.T) {
	t.Setenv("PATH", buildFakeCodexPath(t))

	repoDir := t.TempDir()
	schemaFile := filepath.Join(t.TempDir(), "schema.json")
	if err := os.WriteFile(schemaFile, []byte(`{"type":"object"}`), 0o644); err != nil {
		t.Fatalf("write schema file: %v", err)
	}

	output, stderr, err := runCodexExec(context.Background(), codexExecParams{
		RepoDir:          repoDir,
		OutputSchemaFile: schemaFile,
		Prompt:           "prompt-non-resume",
	})
	if err != nil {
		t.Fatalf("runCodexExec() error = %v, stderr=%s", err, stderr)
	}

	args := readFakeCodexArgs(t)
	assertContainsArg(t, args, "exec")
	assertContainsSequence(t, args, []string{"--cd", repoDir})
	assertContainsSequence(t, args, []string{"--output-schema", schemaFile})
	assertContainsArg(t, args, "prompt-non-resume")
	assertContainsArg(t, args, "--output-last-message")
	assertNotContainsArg(t, args, "resume")

	got := strings.TrimSpace(string(output))
	want := `{"summary":"ok","branch":"main","pr_number":1,"pr_url":"https://example.test/pr/1","session_id":"sess-1","model":"gpt-5.4","reasoning_effort":"xhigh"}`
	if got != want {
		t.Fatalf("runCodexExec() output = %q, want %q", got, want)
	}
}

func TestRunCodexExec_ResumeUsesSessionIDWithoutUnsupportedFlags(t *testing.T) {
	t.Setenv("PATH", buildFakeCodexPath(t))

	repoDir := t.TempDir()
	output, stderr, err := runCodexExec(context.Background(), codexExecParams{
		RepoDir:          repoDir,
		Resume:           true,
		ResumeSessionID:  "session-123",
		OutputSchemaFile: filepath.Join(t.TempDir(), "schema.json"),
		Prompt:           "prompt-resume",
	})
	if err != nil {
		t.Fatalf("runCodexExec() error = %v, stderr=%s", err, stderr)
	}

	args := readFakeCodexArgs(t)
	assertContainsSequence(t, args, []string{"exec", "resume"})
	assertContainsSequence(t, args, []string{"--output-last-message"})
	assertContainsArg(t, args, "session-123")
	assertContainsArg(t, args, "prompt-resume")
	assertNotContainsArg(t, args, "--cd")
	assertNotContainsArg(t, args, "--output-schema")
	assertNotContainsArg(t, args, "--last")

	got := strings.TrimSpace(string(output))
	want := `{"summary":"ok","branch":"main","pr_number":1,"pr_url":"https://example.test/pr/1","session_id":"sess-1","model":"gpt-5.4","reasoning_effort":"xhigh"}`
	if got != want {
		t.Fatalf("runCodexExec() output = %q, want %q", got, want)
	}
}

func TestRunCodexExec_ResumeFallsBackToLastWhenSessionIDMissing(t *testing.T) {
	t.Setenv("PATH", buildFakeCodexPath(t))

	repoDir := t.TempDir()
	if _, stderr, err := runCodexExec(context.Background(), codexExecParams{
		RepoDir: repoDir,
		Resume:  true,
		Prompt:  "prompt-resume-last",
	}); err != nil {
		t.Fatalf("runCodexExec() error = %v, stderr=%s", err, stderr)
	}

	args := readFakeCodexArgs(t)
	assertContainsSequence(t, args, []string{"exec", "resume"})
	assertContainsArg(t, args, "--last")
	assertContainsArg(t, args, "prompt-resume-last")
}

func buildFakeCodexPath(t *testing.T) string {
	t.Helper()

	binDir := t.TempDir()
	argsFile := filepath.Join(binDir, "args.txt")
	mockOutputFile := filepath.Join(binDir, "mock-output.json")
	scriptPath := filepath.Join(binDir, "codex")
	script := `#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$@" > "$FAKE_CODEX_ARGS_FILE"
output_file=""
prev=""
for arg in "$@"; do
  if [[ "$prev" == "--output-last-message" ]]; then
    output_file="$arg"
  fi
  prev="$arg"
done
if [[ -n "$output_file" ]]; then
  cat "$FAKE_CODEX_OUTPUT_FILE" > "$output_file"
fi
printf 'stdout-fallback'
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake codex script: %v", err)
	}
	if err := os.WriteFile(mockOutputFile, []byte(`{"summary":"ok","branch":"main","pr_number":1,"pr_url":"https://example.test/pr/1","session_id":"sess-1","model":"gpt-5.4","reasoning_effort":"xhigh"}`), 0o644); err != nil {
		t.Fatalf("write fake codex output: %v", err)
	}

	t.Setenv("FAKE_CODEX_ARGS_FILE", argsFile)
	t.Setenv("FAKE_CODEX_OUTPUT_FILE", mockOutputFile)
	return binDir + string(os.PathListSeparator) + os.Getenv("PATH")
}

func readFakeCodexArgs(t *testing.T) []string {
	t.Helper()

	argsPath := os.Getenv("FAKE_CODEX_ARGS_FILE")
	bytes, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("read fake codex args: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(bytes)), "\n")
	return lines
}

func assertContainsArg(t *testing.T, args []string, want string) {
	t.Helper()
	for _, arg := range args {
		if arg == want {
			return
		}
	}
	t.Fatalf("args %q do not contain %q", args, want)
}

func assertNotContainsArg(t *testing.T, args []string, want string) {
	t.Helper()
	for _, arg := range args {
		if arg == want {
			t.Fatalf("args %q unexpectedly contain %q", args, want)
		}
	}
}

func assertContainsSequence(t *testing.T, args []string, want []string) {
	t.Helper()
	if len(want) == 0 {
		return
	}
	for idx := 0; idx <= len(args)-len(want); idx++ {
		match := true
		for offset := range want {
			if args[idx+offset] != want[offset] {
				match = false
				break
			}
		}
		if match {
			return
		}
	}
	t.Fatalf("args %q do not contain sequence %q", args, want)
}
