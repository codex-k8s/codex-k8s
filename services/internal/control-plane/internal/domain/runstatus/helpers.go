package runstatus

import (
	"encoding/json"
	"strings"
)

func normalizeLocale(value string, fallback string) string {
	locale := strings.ToLower(strings.TrimSpace(value))
	if locale == "" {
		locale = strings.ToLower(strings.TrimSpace(fallback))
	}
	if strings.HasPrefix(locale, localeRU) {
		return localeRU
	}
	return localeEN
}

func normalizeTriggerKind(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case triggerKindDevRevise:
		return triggerKindDevRevise
	default:
		return triggerKindDev
	}
}

func normalizeRuntimeMode(value string, triggerKind string) string {
	if strings.EqualFold(strings.TrimSpace(value), runtimeModeFullEnv) {
		return runtimeModeFullEnv
	}
	if normalizeTriggerKind(triggerKind) == triggerKindDevRevise || normalizeTriggerKind(triggerKind) == triggerKindDev {
		return runtimeModeFullEnv
	}
	return runtimeModeCode
}

func normalizeRequestedByType(value RequestedByType) RequestedByType {
	switch value {
	case RequestedByTypeStaffUser:
		return RequestedByTypeStaffUser
	default:
		return RequestedByTypeSystem
	}
}

func phaseOrder(phase Phase) int {
	switch phase {
	case PhaseNamespaceDeleted:
		return 3
	case PhaseFinished:
		return 2
	default:
		return 1
	}
}

func mergeState(base commentState, update commentState) commentState {
	if phaseOrder(update.Phase) >= phaseOrder(base.Phase) {
		base.Phase = update.Phase
	}
	if strings.TrimSpace(update.JobName) != "" {
		base.JobName = strings.TrimSpace(update.JobName)
	}
	if strings.TrimSpace(update.JobNamespace) != "" {
		base.JobNamespace = strings.TrimSpace(update.JobNamespace)
	}
	if strings.TrimSpace(update.RuntimeMode) != "" {
		base.RuntimeMode = strings.TrimSpace(update.RuntimeMode)
	}
	if strings.TrimSpace(update.Namespace) != "" {
		base.Namespace = strings.TrimSpace(update.Namespace)
	}
	if strings.TrimSpace(update.TriggerKind) != "" {
		base.TriggerKind = normalizeTriggerKind(update.TriggerKind)
	}
	if strings.TrimSpace(update.PromptLocale) != "" {
		base.PromptLocale = normalizeLocale(update.PromptLocale, localeEN)
	}
	if strings.TrimSpace(update.RunStatus) != "" {
		base.RunStatus = strings.TrimSpace(update.RunStatus)
	}
	if update.Deleted {
		base.Deleted = true
	}
	if update.AlreadyDeleted {
		base.AlreadyDeleted = true
	}
	return base
}

func extractStateMarker(body string) (commentState, bool) {
	lines := strings.Split(strings.TrimSpace(body), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, commentMarkerPrefix) || !strings.HasSuffix(line, commentMarkerSuffix) {
			continue
		}
		rawJSON := strings.TrimSuffix(strings.TrimPrefix(line, commentMarkerPrefix), commentMarkerSuffix)
		var state commentState
		if err := json.Unmarshal([]byte(rawJSON), &state); err != nil {
			return commentState{}, false
		}
		if strings.TrimSpace(state.RunID) == "" {
			return commentState{}, false
		}
		return state, true
	}
	return commentState{}, false
}

func commentContainsRunID(body string, runID string) bool {
	state, ok := extractStateMarker(body)
	if !ok {
		return false
	}
	return strings.TrimSpace(state.RunID) == strings.TrimSpace(runID)
}
