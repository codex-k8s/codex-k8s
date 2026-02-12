package worker

import "testing"

func TestResolveModelFromLabels(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		labels       []string
		defaultModel string
		wantModel    string
		wantSource   string
	}{
		{
			name:         "default",
			labels:       nil,
			defaultModel: "gpt-5.2-codex",
			wantModel:    "gpt-5.2-codex",
			wantSource:   modelSourceDefault,
		},
		{
			name:         "gpt-5.3-codex",
			labels:       []string{"[ai-model-gpt-5.3-codex]"},
			defaultModel: "gpt-5.2-codex",
			wantModel:    "gpt-5.3-codex",
			wantSource:   modelSourceIssueLabel,
		},
		{
			name:         "gpt-5.2-codex",
			labels:       []string{"[ai-model-gpt-5.2-codex]"},
			defaultModel: "gpt-5.2-codex",
			wantModel:    "gpt-5.2-codex",
			wantSource:   modelSourceIssueLabel,
		},
		{
			name:         "gpt-5.1-codex-max",
			labels:       []string{"[ai-model-gpt-5.1-codex-max]"},
			defaultModel: "gpt-5.2-codex",
			wantModel:    "gpt-5.1-codex-max",
			wantSource:   modelSourceIssueLabel,
		},
		{
			name:         "gpt-5.2",
			labels:       []string{"[ai-model-gpt-5.2]"},
			defaultModel: "gpt-5.2-codex",
			wantModel:    "gpt-5.2",
			wantSource:   modelSourceIssueLabel,
		},
		{
			name:         "gpt-5.1-codex-mini",
			labels:       []string{"[ai-model-gpt-5.1-codex-mini]"},
			defaultModel: "gpt-5.2-codex",
			wantModel:    "gpt-5.1-codex-mini",
			wantSource:   modelSourceIssueLabel,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			gotModel, gotSource, err := resolveModelFromLabels(testCase.labels, testCase.defaultModel)
			if err != nil {
				t.Fatalf("resolveModelFromLabels() error = %v", err)
			}
			if gotModel != testCase.wantModel {
				t.Fatalf("resolveModelFromLabels() model = %q, want %q", gotModel, testCase.wantModel)
			}
			if gotSource != testCase.wantSource {
				t.Fatalf("resolveModelFromLabels() source = %q, want %q", gotSource, testCase.wantSource)
			}
		})
	}
}

func TestResolveModelFromLabels_ConflictingLabels(t *testing.T) {
	t.Parallel()

	_, _, err := resolveModelFromLabels([]string{
		"[ai-model-gpt-5.2-codex]",
		"[ai-model-gpt-5.1-codex-mini]",
	}, "gpt-5.2-codex")
	if err == nil {
		t.Fatal("expected conflict error for multiple ai-model labels")
	}
}
