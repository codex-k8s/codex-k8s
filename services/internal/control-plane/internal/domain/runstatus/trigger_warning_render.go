package runstatus

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
)

const (
	triggerWarningTemplateNameRU = "trigger_warning_ru.md.tmpl"
	triggerWarningTemplateNameEN = "trigger_warning_en.md.tmpl"
)

//go:embed templates/trigger_warning_*.md.tmpl
var triggerWarningTemplatesFS embed.FS

var triggerWarningTemplates = template.Must(template.New("runstatus-trigger-warning").ParseFS(triggerWarningTemplatesFS, "templates/trigger_warning_*.md.tmpl"))

type triggerWarningRenderParams struct {
	Locale            string
	ThreadKind        string
	ReasonCode        string
	ConflictingLabels []string
}

type triggerWarningTemplateContext struct {
	ReasonCode        string
	ReasonText        string
	HintText          string
	IsPullRequest     bool
	ConflictingLabels []string
}

func renderTriggerWarningCommentBody(params triggerWarningRenderParams) (string, error) {
	locale := normalizeLocale(params.Locale, localeEN)
	reasonCode := strings.TrimSpace(params.ReasonCode)
	if reasonCode == "" {
		return "", fmt.Errorf("reason code is required")
	}
	threadKind := normalizeCommentTargetKind(params.ThreadKind)
	if threadKind == "" {
		return "", fmt.Errorf("thread kind is required")
	}

	reasonText, hintText := resolveTriggerWarningTexts(locale, reasonCode)
	templateName := triggerWarningTemplateNameEN
	if locale == localeRU {
		templateName = triggerWarningTemplateNameRU
	}

	var out bytes.Buffer
	if err := triggerWarningTemplates.ExecuteTemplate(&out, templateName, triggerWarningTemplateContext{
		ReasonCode:        reasonCode,
		ReasonText:        reasonText,
		HintText:          hintText,
		IsPullRequest:     threadKind == commentTargetKindPullRequest,
		ConflictingLabels: normalizeConflictLabels(params.ConflictingLabels),
	}); err != nil {
		return "", fmt.Errorf("render trigger warning template %s: %w", templateName, err)
	}
	return strings.TrimSpace(out.String()) + "\n", nil
}

func resolveTriggerWarningTexts(locale string, reasonCode string) (string, string) {
	switch strings.TrimSpace(reasonCode) {
	case "pull_request_review_missing_stage_label":
		if locale == localeRU {
			return "в PR review пришло changes_requested, но на PR нет stage-лейбла run:*.", "Поставьте на PR ровно один stage-лейбл run:<stage> (или run:<stage>:revise) и отправьте review снова."
		}
		return "changes_requested was received for PR review, but PR has no run:* stage label.", "Set exactly one run:<stage> (or run:<stage>:revise) label on PR and submit review again."
	case "pull_request_review_stage_label_conflict":
		if locale == localeRU {
			return "в PR одновременно несколько stage-лейблов run:*, из-за чего невозможно выбрать единый revise-этап.", "Оставьте на PR только один stage-лейбл и повторите review."
		}
		return "PR contains multiple run:* stage labels, so revise stage cannot be selected unambiguously.", "Keep only one stage label on PR and submit review again."
	case "repository_not_bound_for_issue_label":
		if locale == localeRU {
			return "репозиторий не привязан к проекту платформы.", "Привяжите репозиторий к проекту в staff console и повторите запуск."
		}
		return "repository is not bound to a platform project.", "Bind repository to a project in staff console and retry."
	default:
		if strings.HasPrefix(reasonCode, "sender_") {
			if locale == localeRU {
				return "инициатор события не имеет прав запускать trigger-лейблы.", "Проверьте роль пользователя (owner/admin/read_write) и доступ в проект."
			}
			return "event sender is not permitted to trigger run labels.", "Verify sender role (owner/admin/read_write) and project membership."
		}
		if locale == localeRU {
			return "платформа не создала run для этого webhook-события.", "Проверьте лейблы и ограничения policy в документации."
		}
		return "platform did not create a run for this webhook event.", "Check labels and trigger policy constraints in documentation."
	}
}
