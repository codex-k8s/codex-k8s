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
		{name: "revise by explicit value", value: promptTemplateKindRevise, triggerKind: "dev", want: promptTemplateKindRevise},
		{name: "revise by legacy explicit value", value: promptTemplateKindReviewOld, triggerKind: "dev", want: promptTemplateKindRevise},
		{name: "revise by revise trigger", value: "", triggerKind: "dev_revise", want: promptTemplateKindRevise},
		{name: "work by self-improve trigger", value: "", triggerKind: "self_improve", want: promptTemplateKindWork},
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
