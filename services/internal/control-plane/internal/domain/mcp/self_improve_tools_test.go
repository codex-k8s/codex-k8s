package mcp

import (
	"testing"

	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func TestNormalizeSelfImproveLimit(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		value int
		want  int
	}{
		{name: "default when empty", value: 0, want: defaultSelfImproveRunsLimit},
		{name: "max when too high", value: 999, want: maxSelfImproveRunsLimit},
		{name: "keep valid", value: 10, want: 10},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := normalizeSelfImproveLimit(testCase.value); got != testCase.want {
				t.Fatalf("normalizeSelfImproveLimit(%d) = %d, want %d", testCase.value, got, testCase.want)
			}
		})
	}
}

func TestSanitizePathSegment(t *testing.T) {
	t.Parallel()

	const valueAllowed = "abc-123_run.1"
	if got := sanitizePathSegment(valueAllowed); got != valueAllowed {
		t.Fatalf("sanitizePathSegment(%q) = %q, want %q", valueAllowed, got, valueAllowed)
	}

	const valueWithSeparators = "../a/b/c"
	const wantWithoutSeparators = "a-b-c"
	if got := sanitizePathSegment(valueWithSeparators); got != wantWithoutSeparators {
		t.Fatalf("sanitizePathSegment(%q) = %q, want %q", valueWithSeparators, got, wantWithoutSeparators)
	}

	const valueOnlySpecial = "!!!"
	const fallbackValue = "run"
	if got := sanitizePathSegment(valueOnlySpecial); got != fallbackValue {
		t.Fatalf("sanitizePathSegment(%q) = %q, want %q", valueOnlySpecial, got, fallbackValue)
	}
}

func TestResolveSelfImproveRepositoryFullName(t *testing.T) {
	t.Parallel()

	runCtx := resolvedRunContext{
		Payload: querytypes.RunPayload{
			Repository: querytypes.RunPayloadRepository{
				FullName: "codex-k8s/codex-k8s",
			},
		},
	}

	if got := resolveSelfImproveRepositoryFullName("example/repo", runCtx); got != "example/repo" {
		t.Fatalf("explicit repository must win, got %q", got)
	}

	if got := resolveSelfImproveRepositoryFullName("", runCtx); got != "codex-k8s/codex-k8s" {
		t.Fatalf("payload repository fallback mismatch, got %q", got)
	}
}
