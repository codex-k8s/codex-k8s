package runtimedeploy

import "testing"

func TestNormalizeRuntimeBuildRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "plain branch", raw: "main", want: "main"},
		{name: "heads prefix", raw: "refs/heads/codex/dev", want: "codex/dev"},
		{name: "origin prefix", raw: "origin/codex/dev", want: "codex/dev"},
		{name: "quoted", raw: "'codex/dev'", want: "codex/dev"},
		{name: "branch option payload", raw: "-b codex/dev", want: "codex/dev"},
		{name: "checkout command payload", raw: "git checkout --detach codex/dev", want: "codex/dev"},
		{name: "invalid option only", raw: "--detach", want: ""},
		{name: "invalid with shell chars", raw: "main;rm -rf /", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := normalizeRuntimeBuildRef(tt.raw); got != tt.want {
				t.Fatalf("normalizeRuntimeBuildRef(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestResolveRuntimeBuildRef(t *testing.T) {
	t.Parallel()

	if got, want := resolveRuntimeBuildRef("-b codex/dev"), "codex/dev"; got != want {
		t.Fatalf("resolveRuntimeBuildRef single candidate = %q, want %q", got, want)
	}

	if got, want := resolveRuntimeBuildRef("--detach", "refs/heads/main"), "main"; got != want {
		t.Fatalf("resolveRuntimeBuildRef fallback candidate = %q, want %q", got, want)
	}

	if got, want := resolveRuntimeBuildRef("", " ", "--detach"), "main"; got != want {
		t.Fatalf("resolveRuntimeBuildRef default = %q, want %q", got, want)
	}
}

func TestNormalizePrepareParamsBuildRef(t *testing.T) {
	t.Parallel()

	params := normalizePrepareParams(PrepareParams{
		RunID:    "run-1",
		BuildRef: "git checkout -b codex/feature-205",
	})
	if got, want := params.BuildRef, "codex/feature-205"; got != want {
		t.Fatalf("normalizePrepareParams BuildRef = %q, want %q", got, want)
	}
}
