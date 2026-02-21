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
	RunID                    string
	TriggerKind              string
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
	ReviseActionLabel        string
	NextStageActionLabel     string

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
	ShowActionCards        bool

	CreatedReached              bool
	PreparingRuntimeReached     bool
	RuntimePreparationCompleted bool
	StartedReached              bool
	AgentStarted                bool
	AuthRequested               bool
	AuthResolvedReached         bool

	IsRunSucceeded bool
	IsRunFailed    bool
	Deleted        bool
	AlreadyDeleted bool

	NeedsCodexAuth               bool
	ShowCodexAuthVerificationURL bool
	ShowCodexAuthUserCode        bool
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
	trimmedSlotURL := strings.TrimSpace(state.SlotURL)
	trimmedIssueURL := strings.TrimSpace(state.IssueURL)
	trimmedPullRequestURL := strings.TrimSpace(state.PullRequestURL)
	trimmedModel := strings.TrimSpace(state.Model)
	trimmedReasoningEffort := strings.TrimSpace(state.ReasoningEffort)
	normalizedRunStatus := strings.ToLower(strings.TrimSpace(state.RunStatus))
	normalizedRuntimeMode := strings.ToLower(strings.TrimSpace(state.RuntimeMode))
	phaseLevel := phaseOrder(state.Phase)
	reviseLabel, nextStageLabel := resolveStageActionLabels(trimmedTriggerKind)

	return commentTemplateContext{
		RunID:                    strings.TrimSpace(state.RunID),
		TriggerKind:              normalizeTriggerKind(trimmedTriggerKind),
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
		ReviseActionLabel:        reviseLabel,
		NextStageActionLabel:     nextStageLabel,

		ManagementURL: managementURL,
		StateMarker:   marker,

		ShowTriggerKind:        trimmedTriggerKind != "",
		ShowRuntimeMode:        trimmedRuntimeMode != "",
		ShowJobRef:             trimmedJobName != "" && trimmedJobNamespace != "",
		ShowNamespace:          trimmedNamespace != "",
		ShowSlotURL:            trimmedSlotURL != "",
		ShowIssueURL:           trimmedIssueURL != "",
		ShowPullRequestURL:     trimmedPullRequestURL != "",
		ShowModel:              trimmedModel != "",
		ShowReasoningEffort:    trimmedReasoningEffort != "",
		ShowFinished:           phaseLevel >= phaseOrder(PhaseFinished),
		ShowNamespaceAction:    trimmedNamespace != "" && phaseLevel >= phaseOrder(PhaseNamespaceDeleted),
		ShowRuntimePreparation: normalizedRuntimeMode == runtimeModeFullEnv,
		ShowActionCards:        reviseLabel != "" || nextStageLabel != "",

		CreatedReached:              phaseLevel >= phaseOrder(PhaseCreated),
		PreparingRuntimeReached:     phaseLevel >= phaseOrder(PhasePreparingRuntime),
		RuntimePreparationCompleted: phaseLevel >= phaseOrder(PhaseStarted),
		StartedReached:              phaseLevel >= phaseOrder(PhaseStarted),
		AgentStarted:                phaseLevel >= phaseOrder(PhaseStarted),
		AuthRequested:               phaseLevel >= phaseOrder(PhaseAuthRequired),
		AuthResolvedReached:         phaseLevel >= phaseOrder(PhaseAuthResolved),

		IsRunSucceeded: normalizedRunStatus == runStatusSucceeded,
		IsRunFailed:    normalizedRunStatus == runStatusFailed,
		Deleted:        state.Deleted,
		AlreadyDeleted: state.AlreadyDeleted,

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

func resolveStageActionLabels(triggerKind string) (reviseLabel string, nextStageLabel string) {
	switch normalizeTriggerKind(triggerKind) {
	case "intake", "intake_revise":
		return "run:intake:revise", "run:vision"
	case "vision", "vision_revise":
		return "run:vision:revise", "run:prd"
	case "prd", "prd_revise":
		return "run:prd:revise", "run:arch"
	case "arch", "arch_revise":
		return "run:arch:revise", "run:design"
	case "design", "design_revise":
		return "run:design:revise", "run:plan"
	case "plan", "plan_revise":
		return "run:plan:revise", "run:dev"
	case "dev", "dev_revise":
		return "run:dev:revise", "run:qa"
	case "qa":
		return "", "run:release"
	case "release":
		return "", "run:postdeploy"
	case "postdeploy":
		return "", "run:ops"
	default:
		return "", ""
	}
}

func renderStateMarker(state commentState) (string, error) {
	raw, err := json.Marshal(state)
	if err != nil {
		return "", fmt.Errorf("marshal run status marker: %w", err)
	}
	return commentMarkerPrefix + string(raw) + commentMarkerSuffix, nil
}
