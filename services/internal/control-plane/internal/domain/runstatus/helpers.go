package runstatus

import (
	"encoding/json"
	"fmt"
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
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
	return string(webhookdomain.NormalizeTriggerKind(value))
}

func resolveUpsertTriggerKind(explicit string, fallback string) string {
	trimmedExplicit := strings.TrimSpace(explicit)
	if trimmedExplicit != "" {
		return normalizeTriggerKind(trimmedExplicit)
	}
	return normalizeTriggerKind(fallback)
}

func normalizeTriggerSource(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case triggerSourcePullRequestReview:
		return triggerSourcePullRequestReview
	default:
		return triggerSourceIssueLabel
	}
}

func normalizeCommentTargetKind(value string) commentTargetKind {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(commentTargetKindIssue):
		return commentTargetKindIssue
	case string(commentTargetKindPullRequest):
		return commentTargetKindPullRequest
	default:
		return ""
	}
}

func normalizeRuntimeMode(value string, triggerKind string) string {
	if strings.EqualFold(strings.TrimSpace(value), runtimeModeFullEnv) {
		return runtimeModeFullEnv
	}
	if webhookdomain.IsKnownTriggerKind(webhookdomain.NormalizeTriggerKind(triggerKind)) {
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
		return 7
	case PhaseFinished:
		return 6
	case PhaseAuthResolved:
		return 5
	case PhaseAuthRequired:
		return 4
	case PhaseStarted:
		return 3
	case PhasePreparingRuntime:
		return 2
	case PhaseCreated:
		return 1
	default:
		return 0
	}
}

func mergeState(base commentState, update commentState) commentState {
	if phaseOrder(update.Phase) >= phaseOrder(base.Phase) {
		base.Phase = update.Phase
	}
	if strings.TrimSpace(update.RepositoryFullName) != "" {
		base.RepositoryFullName = strings.TrimSpace(update.RepositoryFullName)
	}
	if update.IssueNumber > 0 {
		base.IssueNumber = update.IssueNumber
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
	if strings.TrimSpace(update.SlotURL) != "" {
		base.SlotURL = strings.TrimSpace(update.SlotURL)
	}
	if strings.TrimSpace(update.IssueURL) != "" {
		base.IssueURL = strings.TrimSpace(update.IssueURL)
	}
	if strings.TrimSpace(update.PullRequestURL) != "" {
		base.PullRequestURL = strings.TrimSpace(update.PullRequestURL)
	}
	if strings.TrimSpace(update.TriggerKind) != "" {
		base.TriggerKind = normalizeTriggerKind(update.TriggerKind)
	}
	if strings.TrimSpace(update.PromptLocale) != "" {
		base.PromptLocale = normalizeLocale(update.PromptLocale, localeEN)
	}
	if strings.TrimSpace(update.Model) != "" {
		base.Model = strings.TrimSpace(update.Model)
	}
	if strings.TrimSpace(update.ReasoningEffort) != "" {
		base.ReasoningEffort = strings.TrimSpace(update.ReasoningEffort)
	}
	if strings.TrimSpace(update.RunStatus) != "" {
		base.RunStatus = strings.TrimSpace(update.RunStatus)
	}
	if strings.TrimSpace(update.CodexAuthVerificationURL) != "" {
		base.CodexAuthVerificationURL = strings.TrimSpace(update.CodexAuthVerificationURL)
	}
	if strings.TrimSpace(update.CodexAuthUserCode) != "" {
		base.CodexAuthUserCode = strings.TrimSpace(update.CodexAuthUserCode)
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

func resolveCommentTarget(payload querytypes.RunPayload) (commentTargetKind, int, error) {
	issueNumber := int64(0)
	if payload.Issue != nil {
		issueNumber = payload.Issue.Number
	}
	pullRequestNumber := int64(0)
	if payload.PullRequest != nil {
		pullRequestNumber = payload.PullRequest.Number
	}
	triggerSource := triggerSourceIssueLabel
	if payload.Trigger != nil {
		triggerSource = normalizeTriggerSource(payload.Trigger.Source)
	}

	if triggerSource == triggerSourcePullRequestReview {
		switch {
		case pullRequestNumber > 0:
			return commentTargetKindPullRequest, int(pullRequestNumber), nil
		case issueNumber > 0:
			return commentTargetKindPullRequest, int(issueNumber), nil
		default:
			return "", 0, fmt.Errorf("%w", errRunCommentTargetMissing)
		}
	}

	switch {
	case issueNumber > 0:
		return commentTargetKindIssue, int(issueNumber), nil
	case pullRequestNumber > 0:
		return commentTargetKindPullRequest, int(pullRequestNumber), nil
	default:
		return "", 0, fmt.Errorf("%w", errRunCommentTargetMissing)
	}
}
