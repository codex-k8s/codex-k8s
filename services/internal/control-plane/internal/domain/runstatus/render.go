package runstatus

import (
	"encoding/json"
	"fmt"
	"strings"
)

type localizedCommentCopy struct {
	Title                string
	AgentStartedText     string
	TriggerLabel         string
	TimelineTitle        string
	ManagementLinkFormat string
	StartedText          string
	FinishedDefault      string
	FinishedSuccess      string
	FinishedFailed       string
	NamespaceDeleted     string
	NamespaceAlreadyGone string
	NamespacePending     string
}

func renderCommentBody(state commentState, managementURL string) (string, error) {
	copy := resolveLocalizedCommentCopy(normalizeLocale(state.PromptLocale, localeEN))
	var b strings.Builder

	b.WriteString(copy.Title)
	b.WriteString("\n")
	b.WriteString(copy.AgentStartedText)
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("- Run ID: `%s`\n", state.RunID))
	if strings.TrimSpace(state.TriggerKind) != "" {
		b.WriteString(fmt.Sprintf("- %s: `%s`\n", copy.TriggerLabel, normalizeTriggerKind(state.TriggerKind)))
	}
	if strings.TrimSpace(state.JobNamespace) != "" && strings.TrimSpace(state.JobName) != "" {
		b.WriteString(fmt.Sprintf("- Job: `%s/%s`\n", state.JobNamespace, state.JobName))
	}
	if strings.TrimSpace(state.Namespace) != "" {
		b.WriteString(fmt.Sprintf("- Namespace: `%s`\n", state.Namespace))
	}

	if strings.TrimSpace(managementURL) != "" {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(copy.ManagementLinkFormat, managementURL))
		b.WriteString("\n")
	}

	b.WriteString("\n### ")
	b.WriteString(copy.TimelineTitle)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("- ✅ %s\n", copy.StartedText))
	if phaseOrder(state.Phase) >= phaseOrder(PhaseFinished) {
		b.WriteString(fmt.Sprintf("- %s %s\n", finishedEmoji(state), finishedLabel(state, copy)))
	}
	if strings.TrimSpace(state.Namespace) != "" && phaseOrder(state.Phase) >= phaseOrder(PhaseNamespaceDeleted) {
		b.WriteString(fmt.Sprintf("- 🗑️ %s\n", namespaceLabel(state, copy)))
	}

	marker, err := renderStateMarker(state)
	if err != nil {
		return "", err
	}
	b.WriteString("\n")
	b.WriteString(marker)
	b.WriteString("\n")
	return b.String(), nil
}

func resolveLocalizedCommentCopy(locale string) localizedCommentCopy {
	if locale == localeRU {
		return localizedCommentCopy{
			Title:                "Статус агентного запуска",
			AgentStartedText:     "✅ Агент запущен",
			TriggerLabel:         "Режим запуска",
			TimelineTitle:        "Таймлайн",
			ManagementLinkFormat: "🚦 Ран запущен: управление -> 🔗 [Ссылка на запуск](%s)",
			StartedText:          "Запуск создан и выполняется",
			FinishedDefault:      "Задача завершена",
			FinishedSuccess:      "Задача завершена успешно",
			FinishedFailed:       "Задача завершена с ошибкой",
			NamespaceDeleted:     "Namespace удален",
			NamespaceAlreadyGone: "Namespace уже был удален ранее",
			NamespacePending:     "Namespace ожидает удаления",
		}
	}

	return localizedCommentCopy{
		Title:                "Agent Run Status",
		AgentStartedText:     "✅ Agent started",
		TriggerLabel:         "Trigger mode",
		TimelineTitle:        "Timeline",
		ManagementLinkFormat: "🚦 Run started: manage -> 🔗 [Run link](%s)",
		StartedText:          "Run was created and is running",
		FinishedDefault:      "Run finished",
		FinishedSuccess:      "Run finished successfully",
		FinishedFailed:       "Run finished with errors",
		NamespaceDeleted:     "Namespace deleted",
		NamespaceAlreadyGone: "Namespace was already deleted",
		NamespacePending:     "Namespace is waiting for cleanup",
	}
}

func finishedLabel(state commentState, copy localizedCommentCopy) string {
	switch strings.ToLower(strings.TrimSpace(state.RunStatus)) {
	case runStatusSucceeded:
		return copy.FinishedSuccess
	case runStatusFailed:
		return copy.FinishedFailed
	default:
		return copy.FinishedDefault
	}
}

func namespaceLabel(state commentState, copy localizedCommentCopy) string {
	if state.AlreadyDeleted {
		return copy.NamespaceAlreadyGone
	}
	if state.Deleted {
		return copy.NamespaceDeleted
	}
	return copy.NamespacePending
}

func finishedEmoji(state commentState) string {
	switch strings.ToLower(strings.TrimSpace(state.RunStatus)) {
	case runStatusSucceeded:
		return "👌"
	case runStatusFailed:
		return "⚠️"
	default:
		return "✅"
	}
}

func renderStateMarker(state commentState) (string, error) {
	raw, err := json.Marshal(state)
	if err != nil {
		return "", fmt.Errorf("marshal run status marker: %w", err)
	}
	return commentMarkerPrefix + string(raw) + commentMarkerSuffix, nil
}
