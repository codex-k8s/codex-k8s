package runner

import "testing"

func TestResolveRunWriteScopePolicy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		triggerKind   string
		agentKey      string
		wantMode      runWriteScopeMode
		wantRequirePR bool
	}{
		{
			name:          "reviewer is comment only",
			triggerKind:   "dev",
			agentKey:      "reviewer",
			wantMode:      runWriteScopeModeNoRepoChanges,
			wantRequirePR: true,
		},
		{
			name:          "design is markdown only",
			triggerKind:   "design",
			agentKey:      "sa",
			wantMode:      runWriteScopeModeMarkdownOnly,
			wantRequirePR: false,
		},
		{
			name:          "qa is markdown only",
			triggerKind:   "qa",
			agentKey:      "qa",
			wantMode:      runWriteScopeModeMarkdownOnly,
			wantRequirePR: false,
		},
		{
			name:          "ops is markdown only",
			triggerKind:   "ops",
			agentKey:      "sre",
			wantMode:      runWriteScopeModeMarkdownOnly,
			wantRequirePR: false,
		},
		{
			name:          "self improve has restricted scope",
			triggerKind:   "self_improve",
			agentKey:      "km",
			wantMode:      runWriteScopeModeSelfImproveOnly,
			wantRequirePR: false,
		},
		{
			name:          "dev keeps full scope",
			triggerKind:   "dev",
			agentKey:      "dev",
			wantMode:      runWriteScopeModeAny,
			wantRequirePR: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := resolveRunWriteScopePolicy(testCase.triggerKind, testCase.agentKey)
			if got.Mode != testCase.wantMode {
				t.Fatalf("resolveRunWriteScopePolicy().Mode = %q, want %q", got.Mode, testCase.wantMode)
			}
			if got.RequireExistingPR != testCase.wantRequirePR {
				t.Fatalf("resolveRunWriteScopePolicy().RequireExistingPR = %v, want %v", got.RequireExistingPR, testCase.wantRequirePR)
			}
		})
	}
}

func TestIsSelfImproveAllowedPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		path string
		want bool
	}{
		{path: "docs/product/labels_and_trigger_policy.md", want: true},
		{path: "services/jobs/agent-runner/internal/runner/promptseeds/design-work.md", want: true},
		{path: "services/jobs/agent-runner/internal/runner/templates/prompt_envelope.tmpl", want: true},
		{path: "services/jobs/agent-runner/Dockerfile", want: true},
		{path: "services/internal/control-plane/internal/domain/webhook/service.go", want: false},
		{path: "services/jobs/agent-runner/scripts/bootstrap_tools.sh", want: false},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.path, func(t *testing.T) {
			t.Parallel()

			got := isSelfImproveAllowedPath(testCase.path)
			if got != testCase.want {
				t.Fatalf("isSelfImproveAllowedPath(%q) = %v, want %v", testCase.path, got, testCase.want)
			}
		})
	}
}
