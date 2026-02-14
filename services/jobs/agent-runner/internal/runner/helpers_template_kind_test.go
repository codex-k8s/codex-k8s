package runner

import "testing"

func TestNormalizeTemplateKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		value       string
		triggerKind string
		want        string
	}{
		{name: "work by default", value: "", triggerKind: "dev", want: promptTemplateKindWork},
		{name: "review by explicit value", value: promptTemplateKindReview, triggerKind: "dev", want: promptTemplateKindReview},
		{name: "review by revise trigger", value: "", triggerKind: "dev_revise", want: promptTemplateKindReview},
		{name: "review by self-improve trigger", value: "", triggerKind: "self_improve", want: promptTemplateKindReview},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeTemplateKind(testCase.value, testCase.triggerKind)
			if got != testCase.want {
				t.Fatalf("normalizeTemplateKind(%q, %q) = %q, want %q", testCase.value, testCase.triggerKind, got, testCase.want)
			}
		})
	}
}
