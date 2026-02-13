package runstatus

import (
	"bytes"
	"embed"
	"fmt"
	"slices"
	"strings"
	"text/template"
)

const (
	triggerConflictTemplateNameRU = "trigger_conflict_ru.md.tmpl"
	triggerConflictTemplateNameEN = "trigger_conflict_en.md.tmpl"
)

//go:embed templates/trigger_conflict_*.md.tmpl
var triggerConflictTemplatesFS embed.FS

var triggerConflictTemplates = template.Must(template.New("runstatus-trigger-conflict").ParseFS(triggerConflictTemplatesFS, "templates/trigger_conflict_*.md.tmpl"))

type triggerConflictTemplateContext struct {
	TriggerLabel      string
	ConflictingLabels []string
}

func renderTriggerLabelConflictCommentBody(locale string, triggerLabel string, conflictingLabels []string) (string, error) {
	normalizedLabels := normalizeConflictLabels(conflictingLabels)
	if len(normalizedLabels) == 0 {
		return "", fmt.Errorf("conflicting labels are required")
	}
	normalizedTrigger := strings.TrimSpace(triggerLabel)
	if normalizedTrigger == "" {
		normalizedTrigger = normalizedLabels[0]
	}

	templateName := triggerConflictTemplateNameEN
	if normalizeLocale(locale, localeEN) == localeRU {
		templateName = triggerConflictTemplateNameRU
	}
	var out bytes.Buffer
	if err := triggerConflictTemplates.ExecuteTemplate(&out, templateName, triggerConflictTemplateContext{
		TriggerLabel:      normalizedTrigger,
		ConflictingLabels: normalizedLabels,
	}); err != nil {
		return "", fmt.Errorf("render trigger conflict template %s: %w", templateName, err)
	}
	return strings.TrimSpace(out.String()) + "\n", nil
}

func normalizeConflictLabels(labels []string) []string {
	result := make([]string, 0, len(labels))
	for _, raw := range labels {
		label := strings.TrimSpace(raw)
		if label == "" {
			continue
		}
		if !slices.Contains(result, label) {
			result = append(result, label)
		}
	}
	slices.Sort(result)
	return result
}
