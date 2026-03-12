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

type recentAgentStatus struct {
	StatusText  string
	ReportedAt  string
	TimeLabel   string
	RepeatCount int
}

type commentTemplateNextStepAction struct {
	ActionKind     string
	DisplayVariant string
	TargetLabel    string
	URL            string
}

type commentTemplateContext struct {
	RunID                    string
	TriggerKindDisplay       string
	WorkloadKind             string
	RuntimeMode              string
	JobName                  string
	JobNamespace             string
	Namespace                string
	SlotURL                  string
	IssueURL                 string
	PullRequestURL           string
	Model                    string
	ReasoningEffort          string
	RunStatus                string
	CodexAuthVerificationURL string
	CodexAuthUserCode        string
	RecentAgentStatuses      []recentAgentStatus
	NextStepActions          []commentTemplateNextStepAction

	ManagementURL string
	StateMarker   string

	ShowTriggerKind        bool
	ShowRuntimeMode        bool
	ShowJobRef             bool
	ShowNamespace          bool
	ShowSlotURL            bool
	ShowIssueURL           bool
	ShowPullRequestURL     bool
	ShowModel              bool
	ShowReasoningEffort    bool
	ShowFinished           bool
	ShowNamespaceAction    bool
	ShowRuntimePreparation bool
	ShowNextStepActions    bool

	CreatedReached              bool
	PreparingRuntimeReached     bool
	RuntimePreparationActive    bool
	RuntimePreparationCompleted bool
	StartedReached              bool
	AuthRequested               bool
	AuthResolvedReached         bool
	ReadyReached                bool

	IsRunSucceeded bool
	IsRunFailed    bool
	Deleted        bool
	AlreadyDeleted bool

	NeedsCodexAuth               bool
	ShowCodexAuthVerificationURL bool
	ShowCodexAuthUserCode        bool
}

func renderCommentBody(state commentState, managementURL string, nextStepActions []nextStepCommentAction, recentStatuses []recentAgentStatus) (string, error) {
	marker, err := renderStateMarker(state)
	if err != nil {
		return "", err
	}

	ctx := buildCommentTemplateContext(state, strings.TrimSpace(managementURL), marker, nextStepActions, recentStatuses)
	templateName := resolveCommentTemplateName(normalizeLocale(state.PromptLocale, localeEN))
	var out bytes.Buffer
	if err := commentTemplates.ExecuteTemplate(&out, templateName, ctx); err != nil {
		return "", fmt.Errorf("render run status template %s: %w", templateName, err)
	}
	return strings.TrimSpace(out.String()) + "\n", nil
}

func buildCommentTemplateContext(state commentState, managementURL string, marker string, nextStepActions []nextStepCommentAction, recentStatuses []recentAgentStatus) commentTemplateContext {
	trimmedTriggerKind := strings.TrimSpace(state.TriggerKind)
	trimmedRuntimeMode := strings.TrimSpace(state.RuntimeMode)
	trimmedJobName := strings.TrimSpace(state.JobName)
	trimmedJobNamespace := strings.TrimSpace(state.JobNamespace)
	trimmedNamespace := strings.TrimSpace(state.Namespace)
	trimmedSlotURL := strings.TrimSpace(state.SlotURL)
	trimmedIssueURL := strings.TrimSpace(state.IssueURL)
	trimmedPullRequestURL := strings.TrimSpace(state.PullRequestURL)
	trimmedModel := strings.TrimSpace(state.Model)
	trimmedReasoningEffort := strings.TrimSpace(state.ReasoningEffort)
	normalizedRunStatus := strings.ToLower(strings.TrimSpace(state.RunStatus))
	normalizedRuntimeMode := strings.ToLower(strings.TrimSpace(state.RuntimeMode))
	phaseLevel := phaseOrder(state.Phase)
	hasAuthRequested := state.AuthRequested

	formattedRecentStatuses := make([]recentAgentStatus, 0, len(recentStatuses))
	for _, item := range recentStatuses {
		item.TimeLabel = formatRecentStatusTimeLabel(item.ReportedAt, state.PromptLocale, nowUTC())
		formattedRecentStatuses = append(formattedRecentStatuses, item)
	}

	templateActions := make([]commentTemplateNextStepAction, 0, len(nextStepActions))
	for _, item := range nextStepActions {
		templateActions = append(templateActions, commentTemplateNextStepAction{
			ActionKind:     strings.TrimSpace(item.ActionKind),
			DisplayVariant: strings.TrimSpace(item.DisplayVariant),
			TargetLabel:    strings.TrimSpace(item.TargetLabel),
			URL:            strings.TrimSpace(item.URL),
		})
	}

	return commentTemplateContext{
		RunID:                    strings.TrimSpace(state.RunID),
		TriggerKindDisplay:       resolveTriggerKindDisplay(trimmedTriggerKind, state.TriggerLabel, state.DiscussionMode),
		WorkloadKind:             resolveWorkloadKind(trimmedTriggerKind, state.TriggerLabel, state.DiscussionMode),
		RuntimeMode:              trimmedRuntimeMode,
		JobName:                  trimmedJobName,
		JobNamespace:             trimmedJobNamespace,
		Namespace:                trimmedNamespace,
		SlotURL:                  trimmedSlotURL,
		IssueURL:                 trimmedIssueURL,
		PullRequestURL:           trimmedPullRequestURL,
		Model:                    trimmedModel,
		ReasoningEffort:          trimmedReasoningEffort,
		RunStatus:                strings.TrimSpace(state.RunStatus),
		CodexAuthVerificationURL: strings.TrimSpace(state.CodexAuthVerificationURL),
		CodexAuthUserCode:        strings.TrimSpace(state.CodexAuthUserCode),
		RecentAgentStatuses:      formattedRecentStatuses,
		NextStepActions:          templateActions,
		ManagementURL:            managementURL,
		StateMarker:              marker,
		ShowTriggerKind:          resolveTriggerKindDisplay(trimmedTriggerKind, state.TriggerLabel, state.DiscussionMode) != "",
		ShowRuntimeMode:          trimmedRuntimeMode != "",
		ShowJobRef:               trimmedJobName != "" && trimmedJobNamespace != "",
		ShowNamespace:            trimmedNamespace != "",
		ShowSlotURL:              trimmedSlotURL != "",
		ShowIssueURL:             trimmedIssueURL != "",
		ShowPullRequestURL:       trimmedPullRequestURL != "",
		ShowModel:                trimmedModel != "",
		ShowReasoningEffort:      trimmedReasoningEffort != "",
		ShowFinished:             phaseLevel >= phaseOrder(PhaseFinished),
		ShowNamespaceAction:      trimmedNamespace != "" && phaseLevel >= phaseOrder(PhaseNamespaceDeleted),
		ShowRuntimePreparation:   normalizedRuntimeMode == runtimeModeFullEnv,
		ShowNextStepActions:      len(templateActions) > 0,
		CreatedReached:           phaseLevel >= phaseOrder(PhaseCreated),
		PreparingRuntimeReached:  phaseLevel >= phaseOrder(PhasePreparingRuntime),
		RuntimePreparationActive: normalizedRuntimeMode == runtimeModeFullEnv &&
			phaseLevel >= phaseOrder(PhasePreparingRuntime) &&
			phaseLevel < phaseOrder(PhaseStarted),
		RuntimePreparationCompleted:  phaseLevel >= phaseOrder(PhaseStarted),
		StartedReached:               phaseLevel >= phaseOrder(PhaseStarted),
		AuthRequested:                hasAuthRequested,
		AuthResolvedReached:          hasAuthRequested && phaseLevel >= phaseOrder(PhaseAuthResolved),
		ReadyReached:                 phaseLevel >= phaseOrder(PhaseReady),
		IsRunSucceeded:               normalizedRunStatus == runStatusSucceeded,
		IsRunFailed:                  normalizedRunStatus == runStatusFailed,
		Deleted:                      state.Deleted,
		AlreadyDeleted:               state.AlreadyDeleted,
		NeedsCodexAuth:               state.Phase == PhaseAuthRequired,
		ShowCodexAuthVerificationURL: strings.TrimSpace(state.CodexAuthVerificationURL) != "",
		ShowCodexAuthUserCode:        strings.TrimSpace(state.CodexAuthUserCode) != "",
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
