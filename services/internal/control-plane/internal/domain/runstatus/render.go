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
	commentTemplatePathRU = "templates/comment_ru.md.tmpl"
	commentTemplatePathEN = "templates/comment_en.md.tmpl"
)

//go:embed templates/comment_*.md.tmpl
var commentTemplatesFS embed.FS

var commentTemplates = template.Must(template.New("runstatus-comments").ParseFS(commentTemplatesFS, "templates/comment_*.md.tmpl"))

type commentTemplateContext struct {
	RunID        string
	TriggerKind  string
	RuntimeMode  string
	JobName      string
	JobNamespace string
	Namespace    string
	RunStatus    string

	ManagementURL string
	StateMarker   string

	ShowTriggerKind     bool
	ShowRuntimeMode     bool
	ShowJobRef          bool
	ShowNamespace       bool
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
	templatePath := resolveCommentTemplatePath(normalizeLocale(state.PromptLocale, localeEN))
	var out bytes.Buffer
	if err := commentTemplates.ExecuteTemplate(&out, templatePath, ctx); err != nil {
		return "", fmt.Errorf("render run status template %s: %w", templatePath, err)
	}
	return strings.TrimSpace(out.String()) + "\n", nil
}

func buildCommentTemplateContext(state commentState, managementURL string, marker string) commentTemplateContext {
	trimmedTriggerKind := strings.TrimSpace(state.TriggerKind)
	trimmedRuntimeMode := strings.TrimSpace(state.RuntimeMode)
	trimmedJobName := strings.TrimSpace(state.JobName)
	trimmedJobNamespace := strings.TrimSpace(state.JobNamespace)
	trimmedNamespace := strings.TrimSpace(state.Namespace)
	normalizedRunStatus := strings.ToLower(strings.TrimSpace(state.RunStatus))

	return commentTemplateContext{
		RunID:        strings.TrimSpace(state.RunID),
		TriggerKind:  normalizeTriggerKind(trimmedTriggerKind),
		RuntimeMode:  trimmedRuntimeMode,
		JobName:      trimmedJobName,
		JobNamespace: trimmedJobNamespace,
		Namespace:    trimmedNamespace,
		RunStatus:    strings.TrimSpace(state.RunStatus),

		ManagementURL: managementURL,
		StateMarker:   marker,

		ShowTriggerKind:     trimmedTriggerKind != "",
		ShowRuntimeMode:     trimmedRuntimeMode != "",
		ShowJobRef:          trimmedJobName != "" && trimmedJobNamespace != "",
		ShowNamespace:       trimmedNamespace != "",
		ShowFinished:        phaseOrder(state.Phase) >= phaseOrder(PhaseFinished),
		ShowNamespaceAction: trimmedNamespace != "" && phaseOrder(state.Phase) >= phaseOrder(PhaseNamespaceDeleted),

		IsRunSucceeded: normalizedRunStatus == runStatusSucceeded,
		IsRunFailed:    normalizedRunStatus == runStatusFailed,
		Deleted:        state.Deleted,
		AlreadyDeleted: state.AlreadyDeleted,
	}
}

func resolveCommentTemplatePath(locale string) string {
	if locale == localeRU {
		return commentTemplatePathRU
	}
	return commentTemplatePathEN
}

func renderStateMarker(state commentState) (string, error) {
	raw, err := json.Marshal(state)
	if err != nil {
		return "", fmt.Errorf("marshal run status marker: %w", err)
	}
	return commentMarkerPrefix + string(raw) + commentMarkerSuffix, nil
}
