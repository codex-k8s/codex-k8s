package runner

import (
	"slices"
	"testing"
)

func TestNormalizeStringList(t *testing.T) {
	t.Parallel()

	got := normalizeStringList([]string{" gh ", "", "kubectl", "GH", "kubectl"})
	want := []string{"gh", "kubectl"}
	if !slices.Equal(got, want) {
		t.Fatalf("normalizeStringList() = %#v, want %#v", got, want)
	}
}

func TestDetectToolGaps(t *testing.T) {
	t.Parallel()

	report := codexReport{
		ToolGaps: []string{"protoc"},
	}
	codexOutput := `
failed to run tool: "golangci-lint": executable file not found
bash: line 1: jq: command not found
`
	gitPushOutput := `
missing required command: kubectl
`

	got := detectToolGaps(report, codexOutput, gitPushOutput)
	want := []string{"protoc", "golangci-lint", "jq", "kubectl"}
	if !slices.Equal(got, want) {
		t.Fatalf("detectToolGaps() = %#v, want %#v", got, want)
	}
}
