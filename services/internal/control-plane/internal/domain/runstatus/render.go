package runstatus

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

const (
	commentTemplateNameRU = "comment_ru.md.tmpl"
	commentTemplateNameEN = "comment_en.md.tmpl"
)

//go:embed templates/comment_*.md.tmpl
var commentTemplatesFS embed.FS

var commentTemplates = template.Must(template.New("runstatus-comments").ParseFS(commentTemplatesFS, "templates/comment_*.md.tmpl"))

type commentTemplateContext struct {
	RunID           string
	TriggerKind     string
	RuntimeMode     string
	JobName         string
	JobNamespace    string
	Namespace       string
	Model           string
	ReasoningEffort string
	RunStatus       string

	ManagementURL string
	StateMarker   string

	ShowTriggerKind     bool
	ShowRuntimeMode     bool
	ShowJobRef          bool
	ShowNamespace       bool
	ShowModel           bool
	ShowReasoningEffort bool
	ShowFinished        bool
	ShowNamespaceAction bool

	IsRunSucceeded bool
	IsRunFailed    bool
	Deleted        bool
	AlreadyDeleted bool
}

func renderCommentBody(state commentState, managementURL string) (string, error) {
	marker, err := renderStateMarker(state)
	if err != nil {
		return "", err
	}

	ctx := buildCommentTemplateContext(state, strings.TrimSpace(managementURL), marker)
	templateName := resolveCommentTemplateName(normalizeLocale(state.PromptLocale, localeEN))
	var out bytes.Buffer
	if err := commentTemplates.ExecuteTemplate(&out, templateName, ctx); err != nil {
		return "", fmt.Errorf("render run status template %s: %w", templateName, err)
	}
	return strings.TrimSpace(out.String()) + "\n", nil
}

func buildCommentTemplateContext(state commentState, managementURL string, marker string) commentTemplateContext {
	trimmedTriggerKind := strings.TrimSpace(state.TriggerKind)
	trimmedRuntimeMode := strings.TrimSpace(state.RuntimeMode)
	trimmedJobName := strings.TrimSpace(state.JobName)
	trimmedJobNamespace := strings.TrimSpace(state.JobNamespace)
	trimmedNamespace := strings.TrimSpace(state.Namespace)
	trimmedModel := strings.TrimSpace(state.Model)
	trimmedReasoningEffort := strings.TrimSpace(state.ReasoningEffort)
	normalizedRunStatus := strings.ToLower(strings.TrimSpace(state.RunStatus))

	return commentTemplateContext{
		RunID:           strings.TrimSpace(state.RunID),
		TriggerKind:     normalizeTriggerKind(trimmedTriggerKind),
		RuntimeMode:     trimmedRuntimeMode,
		JobName:         trimmedJobName,
		JobNamespace:    trimmedJobNamespace,
		Namespace:       trimmedNamespace,
		Model:           trimmedModel,
		ReasoningEffort: trimmedReasoningEffort,
		RunStatus:       strings.TrimSpace(state.RunStatus),

		ManagementURL: managementURL,
		StateMarker:   marker,

		ShowTriggerKind:     trimmedTriggerKind != "",
		ShowRuntimeMode:     trimmedRuntimeMode != "",
		ShowJobRef:          trimmedJobName != "" && trimmedJobNamespace != "",
		ShowNamespace:       trimmedNamespace != "",
		ShowModel:           trimmedModel != "",
		ShowReasoningEffort: trimmedReasoningEffort != "",
		ShowFinished:        phaseOrder(state.Phase) >= phaseOrder(PhaseFinished),
		ShowNamespaceAction: trimmedNamespace != "" && phaseOrder(state.Phase) >= phaseOrder(PhaseNamespaceDeleted),

		IsRunSucceeded: normalizedRunStatus == runStatusSucceeded,
		IsRunFailed:    normalizedRunStatus == runStatusFailed,
		Deleted:        state.Deleted,
		AlreadyDeleted: state.AlreadyDeleted,
	}
}

func resolveCommentTemplateName(locale string) string {
	if locale == localeRU {
		return commentTemplateNameRU
	}
	return commentTemplateNameEN
}

func renderStateMarker(state commentState) (string, error) {
	raw, err := json.Marshal(state)
	if err != nil {
		return "", fmt.Errorf("marshal run status marker: %w", err)
	}
	return commentMarkerPrefix + string(raw) + commentMarkerSuffix, nil
}
