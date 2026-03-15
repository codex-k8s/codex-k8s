package runner

import (
	"strings"
	"testing"
)

func TestRenderTaskTemplate_DevSeedsRequireRunStatusReport(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		triggerKind  string
		templateKind string
		locale       string
	}{
		{
			name:         "dev work ru",
			triggerKind:  "dev",
			templateKind: promptTemplateKindWork,
			locale:       "ru",
		},
		{
			name:         "dev work en",
			triggerKind:  "dev",
			templateKind: promptTemplateKindWork,
			locale:       "en",
		},
		{
			name:         "dev revise ru",
			triggerKind:  "dev_revise",
			templateKind: promptTemplateKindRevise,
			locale:       "ru",
		},
		{
			name:         "dev revise en",
			triggerKind:  "dev_revise",
			templateKind: promptTemplateKindRevise,
			locale:       "en",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			service := &Service{
				cfg: Config{
					AgentKey: "dev",
					PromptConfig: PromptConfig{
						TriggerKind:          testCase.triggerKind,
						AgentBaseBranch:      "main",
						PromptTemplateKind:   testCase.templateKind,
						PromptTemplateLocale: testCase.locale,
					},
				},
			}

			body, err := service.renderTaskTemplate(testCase.templateKind, t.TempDir())
			if err != nil {
				t.Fatalf("renderTaskTemplate() error = %v", err)
			}

			normalizedBody := strings.ToLower(body)
			if !strings.Contains(normalizedBody, "run_status_report") {
				t.Fatalf("rendered template must mention run_status_report, got: %q", body)
			}
			if !strings.Contains(normalizedBody, "user.decision.request") {
				t.Fatalf("rendered template must mention user.decision.request, got: %q", body)
			}
			if !strings.Contains(normalizedBody, "3-4") {
				t.Fatalf("rendered template must mention 3-4 cadence, got: %q", body)
			}
		})
	}
}

func TestRenderTaskTemplate_DiscussionRequiresCommentRefresh(t *testing.T) {
	t.Parallel()

	service := &Service{
		cfg: Config{
			AgentKey: "dev",
			PromptConfig: PromptConfig{
				TriggerKind:          "dev",
				PromptTemplateKind:   promptTemplateKindDiscussion,
				PromptTemplateLocale: promptLocaleRU,
			},
		},
	}

	body, err := service.renderTaskTemplate(promptTemplateKindDiscussion, t.TempDir())
	if err != nil {
		t.Fatalf("renderTaskTemplate() error = %v", err)
	}

	if !strings.Contains(body, "перед КАЖДЫМ ответом") {
		t.Fatalf("discussion prompt must require comment refresh before each reply, got: %q", body)
	}
	if !strings.Contains(body, "не оставляйте без ответа") {
		t.Fatalf("discussion prompt must require answering all new human comments, got: %q", body)
	}
}
