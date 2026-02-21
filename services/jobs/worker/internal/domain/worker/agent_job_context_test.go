package worker

import (
	"encoding/json"
	"testing"
)

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
			name:         "gpt-5.3-codex-spark",
			labels:       []string{"[ai-model-gpt-5.3-codex-spark]"},
			defaultModel: "gpt-5.2-codex",
			wantModel:    "gpt-5.3-codex-spark",
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

func TestResolveRunAgentContext_FallbackFromGPT53SparkWithoutAuth(t *testing.T) {
	t.Parallel()

	runPayload := json.RawMessage(`{
		"repository":{"full_name":"codex-k8s/codex-k8s"},
		"issue":{"number":13},
		"agent":{"key":"dev","name":"AI Developer"},
		"trigger":{"kind":"dev","label":"run:dev"},
		"raw_payload":{"issue":{"labels":[{"name":"[ai-model-gpt-5.3-codex-spark]"}]}}
	}`)

	got, err := resolveRunAgentContext(runPayload, runAgentDefaults{
		DefaultModel:           modelGPT52Codex,
		DefaultReasoningEffort: reasoningEffortExtraHigh,
		DefaultLocale:          "ru",
		AllowGPT53:             false,
	})
	if err != nil {
		t.Fatalf("resolveRunAgentContext() error = %v", err)
	}
	if got.Model != modelGPT52Codex {
		t.Fatalf("Model = %q, want %q", got.Model, modelGPT52Codex)
	}
	if got.ModelSource != modelSourceFallback {
		t.Fatalf("ModelSource = %q, want %q", got.ModelSource, modelSourceFallback)
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

func TestResolveRunAgentContext_UsesPullRequestHintsForRevise(t *testing.T) {
	t.Parallel()

	runPayload := json.RawMessage(`{
		"repository":{"full_name":"codex-k8s/codex-k8s"},
		"agent":{"key":"dev","name":"AI Developer"},
		"trigger":{"kind":"dev_revise","label":"run:dev:revise"},
		"raw_payload":{
			"pull_request":{
				"number":200,
				"head":{"ref":"codex/issue-13"},
				"labels":[{"name":"[ai-model-gpt-5.2-codex]"}]
			}
		}
	}`)

	got, err := resolveRunAgentContext(runPayload, runAgentDefaults{
		DefaultModel:           modelGPT52Codex,
		DefaultReasoningEffort: reasoningEffortExtraHigh,
		DefaultLocale:          "ru",
		AllowGPT53:             true,
	})
	if err != nil {
		t.Fatalf("resolveRunAgentContext() error = %v", err)
	}
	if got.IssueNumber != 200 {
		t.Fatalf("IssueNumber = %d, want 200", got.IssueNumber)
	}
	if got.TargetBranch != "codex/issue-13" {
		t.Fatalf("TargetBranch = %q, want codex/issue-13", got.TargetBranch)
	}
	if got.ExistingPRNumber != 200 {
		t.Fatalf("ExistingPRNumber = %d, want 200", got.ExistingPRNumber)
	}
	if got.PromptTemplateKind != promptTemplateKindRevise {
		t.Fatalf("PromptTemplateKind = %q, want %q", got.PromptTemplateKind, promptTemplateKindRevise)
	}
}

func TestResolveRunAgentContext_ReviewTemplateKinds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		runPayload      json.RawMessage
		wantTriggerKind string
	}{
		{
			name: "stage revise uses revise template",
			runPayload: json.RawMessage(`{
				"repository":{"full_name":"codex-k8s/codex-k8s"},
				"issue":{"number":201},
				"agent":{"key":"dev","name":"AI Developer"},
				"trigger":{"kind":"vision_revise","label":"run:vision:revise"},
				"raw_payload":{"issue":{"labels":[{"name":"run:vision:revise"}]}}
			}`),
			wantTriggerKind: "vision_revise",
		},
		{
			name: "self improve uses work template",
			runPayload: json.RawMessage(`{
				"repository":{"full_name":"codex-k8s/codex-k8s"},
				"issue":{"number":202},
				"agent":{"key":"km","name":"AI Knowledge Manager"},
				"trigger":{"kind":"self_improve","label":"run:self-improve"},
				"raw_payload":{"issue":{"labels":[{"name":"run:self-improve"}]}}
			}`),
			wantTriggerKind: "self_improve",
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := resolveRunAgentContext(testCase.runPayload, runAgentDefaults{
				DefaultModel:           modelGPT52Codex,
				DefaultReasoningEffort: reasoningEffortExtraHigh,
				DefaultLocale:          "ru",
				AllowGPT53:             true,
			})
			if err != nil {
				t.Fatalf("resolveRunAgentContext() error = %v", err)
			}
			if got.TriggerKind != testCase.wantTriggerKind {
				t.Fatalf("TriggerKind = %q, want %q", got.TriggerKind, testCase.wantTriggerKind)
			}
			wantTemplateKind := promptTemplateKindWork
			if testCase.wantTriggerKind == "vision_revise" {
				wantTemplateKind = promptTemplateKindRevise
			}
			if got.PromptTemplateKind != wantTemplateKind {
				t.Fatalf("PromptTemplateKind = %q, want %q", got.PromptTemplateKind, wantTemplateKind)
			}
		})
	}
}

func TestResolveRunAgentContext_ConfigLabelsPullRequestOverrideIssue(t *testing.T) {
	t.Parallel()

	runPayload := json.RawMessage(`{
		"repository":{"full_name":"codex-k8s/codex-k8s"},
		"agent":{"key":"dev","name":"AI Developer"},
		"trigger":{"kind":"dev_revise","label":"run:dev:revise"},
		"raw_payload":{
			"issue":{
				"number":20,
				"labels":[
					{"name":"[ai-model-gpt-5.1-codex-mini]"},
					{"name":"[ai-reasoning-low]"}
				]
			},
			"pull_request":{
				"number":200,
				"head":{"ref":"codex/issue-20"},
				"labels":[
					{"name":"[ai-model-gpt-5.2-codex]"},
					{"name":"[ai-reasoning-high]"}
				]
			}
		}
	}`)

	got, err := resolveRunAgentContext(runPayload, runAgentDefaults{
		DefaultModel:           modelGPT52Codex,
		DefaultReasoningEffort: reasoningEffortExtraHigh,
		DefaultLocale:          "ru",
		AllowGPT53:             true,
	})
	if err != nil {
		t.Fatalf("resolveRunAgentContext() error = %v", err)
	}
	if got.Model != modelGPT52Codex {
		t.Fatalf("Model = %q, want %q", got.Model, modelGPT52Codex)
	}
	if got.ModelSource != modelSourcePullRequestLabel {
		t.Fatalf("ModelSource = %q, want %q", got.ModelSource, modelSourcePullRequestLabel)
	}
	if got.ReasoningEffort != "high" {
		t.Fatalf("ReasoningEffort = %q, want high", got.ReasoningEffort)
	}
	if got.ReasoningSource != modelSourcePullRequestLabel {
		t.Fatalf("ReasoningSource = %q, want %q", got.ReasoningSource, modelSourcePullRequestLabel)
	}
}

func TestResolveRunAgentContext_ConflictingPullRequestLabelsFail(t *testing.T) {
	t.Parallel()

	runPayload := json.RawMessage(`{
		"repository":{"full_name":"codex-k8s/codex-k8s"},
		"agent":{"key":"dev","name":"AI Developer"},
		"trigger":{"kind":"dev_revise","label":"run:dev:revise"},
		"raw_payload":{
			"issue":{
				"number":20,
				"labels":[
					{"name":"[ai-model-gpt-5.2-codex]"}
				]
			},
			"pull_request":{
				"number":200,
				"head":{"ref":"codex/issue-20"},
				"labels":[
					{"name":"[ai-model-gpt-5.2-codex]"},
					{"name":"[ai-model-gpt-5.1-codex-mini]"}
				]
			}
		}
	}`)

	_, err := resolveRunAgentContext(runPayload, runAgentDefaults{
		DefaultModel:           modelGPT52Codex,
		DefaultReasoningEffort: reasoningEffortExtraHigh,
		DefaultLocale:          "ru",
		AllowGPT53:             true,
	})
	if err == nil {
		t.Fatal("expected conflict error for multiple pull_request ai-model labels")
	}
}

func TestResolveRunAgentContext_ReviewDrivenReviseIssueLabelsOverridePullRequest(t *testing.T) {
	t.Parallel()

	runPayload := json.RawMessage(`{
		"repository":{"full_name":"codex-k8s/codex-k8s"},
		"agent":{"key":"dev","name":"AI Developer"},
		"trigger":{"source":"pull_request_review","kind":"dev_revise","label":"run:dev:revise"},
		"raw_payload":{
			"issue":{
				"number":20,
				"labels":[
					{"name":"[ai-model-gpt-5.1-codex-mini]"},
					{"name":"[ai-reasoning-low]"}
				]
			},
			"pull_request":{
				"number":200,
				"head":{"ref":"codex/issue-20"},
				"labels":[
					{"name":"[ai-model-gpt-5.2-codex]"},
					{"name":"[ai-reasoning-high]"}
				]
			}
		}
	}`)

	got, err := resolveRunAgentContext(runPayload, runAgentDefaults{
		DefaultModel:           modelGPT52Codex,
		DefaultReasoningEffort: reasoningEffortExtraHigh,
		DefaultLocale:          "ru",
		AllowGPT53:             true,
	})
	if err != nil {
		t.Fatalf("resolveRunAgentContext() error = %v", err)
	}
	if got.Model != modelGPT51CodexMini {
		t.Fatalf("Model = %q, want %q", got.Model, modelGPT51CodexMini)
	}
	if got.ModelSource != modelSourceIssueLabel {
		t.Fatalf("ModelSource = %q, want %q", got.ModelSource, modelSourceIssueLabel)
	}
	if got.ReasoningEffort != reasoningEffortLow {
		t.Fatalf("ReasoningEffort = %q, want %q", got.ReasoningEffort, reasoningEffortLow)
	}
	if got.ReasoningSource != modelSourceIssueLabel {
		t.Fatalf("ReasoningSource = %q, want %q", got.ReasoningSource, modelSourceIssueLabel)
	}
}

func TestResolveRunAgentContext_ReviewDrivenReviseUsesLastRunProfileHints(t *testing.T) {
	t.Parallel()

	runPayload := json.RawMessage(`{
		"repository":{"full_name":"codex-k8s/codex-k8s"},
		"agent":{"key":"dev","name":"AI Developer"},
		"trigger":{"source":"pull_request_review","kind":"dev_revise","label":"run:dev:revise"},
		"profile_hints":{
			"last_run_issue_labels":[
				"[ai-model-gpt-5.1-codex-mini]",
				"[ai-reasoning-low]"
			]
		},
		"raw_payload":{
			"pull_request":{
				"number":200,
				"head":{"ref":"codex/issue-20"},
				"labels":[]
			},
			"issue":{
				"number":20,
				"labels":[]
			}
		}
	}`)

	got, err := resolveRunAgentContext(runPayload, runAgentDefaults{
		DefaultModel:           modelGPT52Codex,
		DefaultReasoningEffort: reasoningEffortExtraHigh,
		DefaultLocale:          "ru",
		AllowGPT53:             true,
	})
	if err != nil {
		t.Fatalf("resolveRunAgentContext() error = %v", err)
	}
	if got.Model != modelGPT51CodexMini {
		t.Fatalf("Model = %q, want %q", got.Model, modelGPT51CodexMini)
	}
	if got.ModelSource != modelSourceLastRunContext {
		t.Fatalf("ModelSource = %q, want %q", got.ModelSource, modelSourceLastRunContext)
	}
	if got.ReasoningEffort != reasoningEffortLow {
		t.Fatalf("ReasoningEffort = %q, want %q", got.ReasoningEffort, reasoningEffortLow)
	}
	if got.ReasoningSource != modelSourceLastRunContext {
		t.Fatalf("ReasoningSource = %q, want %q", got.ReasoningSource, modelSourceLastRunContext)
	}
}

func TestResolveRunAgentContext_ReasoningExtraHighLabel(t *testing.T) {
	t.Parallel()

	runPayload := json.RawMessage(`{
		"repository":{"full_name":"codex-k8s/codex-k8s"},
		"agent":{"key":"dev","name":"AI Developer"},
		"trigger":{"kind":"dev","label":"run:dev"},
		"raw_payload":{
			"issue":{
				"number":21,
				"labels":[{"name":"[ai-reasoning-extra-high]"}]
			}
		}
	}`)

	got, err := resolveRunAgentContext(runPayload, runAgentDefaults{
		DefaultModel:           modelGPT52Codex,
		DefaultReasoningEffort: reasoningEffortHigh,
		DefaultLocale:          "ru",
		AllowGPT53:             true,
	})
	if err != nil {
		t.Fatalf("resolveRunAgentContext() error = %v", err)
	}
	if got.ReasoningEffort != reasoningEffortExtraHigh {
		t.Fatalf("ReasoningEffort = %q, want %q", got.ReasoningEffort, reasoningEffortExtraHigh)
	}
	if got.ReasoningSource != modelSourceIssueLabel {
		t.Fatalf("ReasoningSource = %q, want %q", got.ReasoningSource, modelSourceIssueLabel)
	}
}

func TestResolveRunAgentContext_CustomReasoningLabelFromCatalog(t *testing.T) {
	t.Parallel()

	runPayload := json.RawMessage(`{
		"repository":{"full_name":"codex-k8s/codex-k8s"},
		"agent":{"key":"dev","name":"AI Developer"},
		"trigger":{"kind":"dev","label":"run:dev"},
		"raw_payload":{
			"issue":{
				"number":22,
				"labels":[{"name":"[team-reasoning-ultra]"}]
			}
		}
	}`)

	got, err := resolveRunAgentContext(runPayload, runAgentDefaults{
		DefaultModel:           modelGPT52Codex,
		DefaultReasoningEffort: reasoningEffortHigh,
		DefaultLocale:          "ru",
		AllowGPT53:             true,
		LabelCatalog: runAgentLabelCatalog{
			AIReasoningExtraHighLabel: "[team-reasoning-ultra]",
		},
	})
	if err != nil {
		t.Fatalf("resolveRunAgentContext() error = %v", err)
	}
	if got.ReasoningEffort != reasoningEffortExtraHigh {
		t.Fatalf("ReasoningEffort = %q, want %q", got.ReasoningEffort, reasoningEffortExtraHigh)
	}
	if got.ReasoningSource != modelSourceIssueLabel {
		t.Fatalf("ReasoningSource = %q, want %q", got.ReasoningSource, modelSourceIssueLabel)
	}
}
