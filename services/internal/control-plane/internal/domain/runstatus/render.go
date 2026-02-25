package runstatus

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
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
	StatusText string
	AgentKey   string
	ReportedAt string
}

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
	ReviseActionURL          string
	NextStageActionLabel     string
	NextStageActionURL       string
	AlternativeActionLabel   string
	AlternativeActionURL     string
	ActionCardLaunchProfile  string
	ActionCardStagePath      string
	ActionCardPrimaryAction  string
	ActionCardFallbackAction string
	ActionCardGuardrailNote  string
	RecentAgentStatuses      []recentAgentStatus

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
	ShowActionCardContract bool

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

func renderCommentBody(state commentState, managementURL string, publicBaseURL string, recentStatuses []recentAgentStatus) (string, error) {
	marker, err := renderStateMarker(state)
	if err != nil {
		return "", err
	}

	ctx := buildCommentTemplateContext(state, strings.TrimSpace(managementURL), strings.TrimSpace(publicBaseURL), marker, recentStatuses)
	templateName := resolveCommentTemplateName(normalizeLocale(state.PromptLocale, localeEN))
	var out bytes.Buffer
	if err := commentTemplates.ExecuteTemplate(&out, templateName, ctx); err != nil {
		return "", fmt.Errorf("render run status template %s: %w", templateName, err)
	}
	return strings.TrimSpace(out.String()) + "\n", nil
}

func buildCommentTemplateContext(state commentState, managementURL string, publicBaseURL string, marker string, recentStatuses []recentAgentStatus) commentTemplateContext {
	trimmedTriggerKind := strings.TrimSpace(state.TriggerKind)
	trimmedRepositoryFullName := strings.TrimSpace(state.RepositoryFullName)
	trimmedRuntimeMode := strings.TrimSpace(state.RuntimeMode)
	trimmedJobName := strings.TrimSpace(state.JobName)
	trimmedJobNamespace := strings.TrimSpace(state.JobNamespace)
	trimmedNamespace := strings.TrimSpace(state.Namespace)
	trimmedSlotURL := strings.TrimSpace(state.SlotURL)
	trimmedIssueURL := strings.TrimSpace(state.IssueURL)
	trimmedPullRequestURL := strings.TrimSpace(state.PullRequestURL)
	trimmedModel := strings.TrimSpace(state.Model)
	trimmedReasoningEffort := strings.TrimSpace(state.ReasoningEffort)
	trimmedActionCardLaunchProfile := strings.TrimSpace(state.LaunchProfile)
	trimmedActionCardStagePath := strings.TrimSpace(state.StagePath)
	trimmedActionCardPrimaryAction := strings.TrimSpace(state.PrimaryAction)
	trimmedActionCardFallbackAction := strings.TrimSpace(state.FallbackAction)
	trimmedActionCardGuardrail := strings.TrimSpace(state.GuardrailNote)
	trimmedReviseActionLabel := strings.TrimSpace(state.ReviseActionLabel)
	trimmedNextStageActionLabel := strings.TrimSpace(state.NextStageActionLabel)
	trimmedAlternativeActionLabel := strings.TrimSpace(state.AlternativeActionLabel)
	normalizedRunStatus := strings.ToLower(strings.TrimSpace(state.RunStatus))
	normalizedRuntimeMode := strings.ToLower(strings.TrimSpace(state.RuntimeMode))
	phaseLevel := phaseOrder(state.Phase)
	reviseLabel := trimmedReviseActionLabel
	nextStageLabel := trimmedNextStageActionLabel
	alternativeLabel := trimmedAlternativeActionLabel
	if reviseLabel == "" && nextStageLabel == "" && alternativeLabel == "" {
		reviseLabel, nextStageLabel, alternativeLabel = resolveStageActionLabels(trimmedTriggerKind)
	}
	if isBlockedActionCard(trimmedActionCardGuardrail) {
		reviseLabel = ""
		nextStageLabel = ""
		alternativeLabel = ""
	}
	reviseActionURL := buildStageTransitionActionURL(publicBaseURL, trimmedRepositoryFullName, state.IssueNumber, reviseLabel, trimmedIssueURL)
	nextStageActionURL := strings.TrimSpace(state.PrimaryAction)
	if nextStageActionURL == "" {
		nextStageActionURL = buildStageTransitionActionURL(publicBaseURL, trimmedRepositoryFullName, state.IssueNumber, nextStageLabel, trimmedIssueURL)
	}
	alternativeActionURL := buildStageTransitionActionURL(publicBaseURL, trimmedRepositoryFullName, state.IssueNumber, alternativeLabel, trimmedIssueURL)

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
		ReviseActionURL:          reviseActionURL,
		NextStageActionLabel:     nextStageLabel,
		NextStageActionURL:       nextStageActionURL,
		AlternativeActionLabel:   alternativeLabel,
		AlternativeActionURL:     alternativeActionURL,
		ActionCardLaunchProfile:  trimmedActionCardLaunchProfile,
		ActionCardStagePath:      trimmedActionCardStagePath,
		ActionCardPrimaryAction:  trimmedActionCardPrimaryAction,
		ActionCardFallbackAction: trimmedActionCardFallbackAction,
		ActionCardGuardrailNote:  trimmedActionCardGuardrail,
		RecentAgentStatuses:      recentStatuses,

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
		ShowActionCards:        reviseLabel != "" || nextStageLabel != "" || alternativeLabel != "",
		ShowActionCardContract: trimmedActionCardLaunchProfile != "" || trimmedActionCardStagePath != "" || trimmedActionCardPrimaryAction != "" || trimmedActionCardFallbackAction != "" || trimmedActionCardGuardrail != "",

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

func resolveStageActionLabels(triggerKind string) (reviseLabel string, nextStageLabel string, alternativeLabel string) {
	switch normalizeTriggerKind(triggerKind) {
	case "intake", "intake_revise":
		return "run:intake:revise", "run:vision", ""
	case "vision", "vision_revise":
		return "run:vision:revise", "run:prd", ""
	case "prd", "prd_revise":
		return "run:prd:revise", "run:arch", ""
	case "arch", "arch_revise":
		return "run:arch:revise", "run:design", ""
	case "design", "design_revise":
		return "run:design:revise", "run:plan", "run:dev"
	case "plan", "plan_revise":
		return "run:plan:revise", "run:dev", ""
	case "dev", "dev_revise":
		return "run:dev:revise", "run:qa", ""
	case "qa":
		return "", "run:release", ""
	case "release":
		return "", "run:postdeploy", ""
	case "postdeploy":
		return "", "run:ops", ""
	default:
		return "", "", ""
	}
}

func buildStageTransitionActionURL(publicBaseURL string, repositoryFullName string, issueNumber int, targetLabel string, issueURL string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	repositoryFullName = strings.TrimSpace(repositoryFullName)
	targetLabel = strings.TrimSpace(targetLabel)
	if baseURL == "" || repositoryFullName == "" || issueNumber <= 0 || targetLabel == "" {
		return ""
	}

	values := url.Values{}
	values.Set("repo", repositoryFullName)
	values.Set("issue", strconv.Itoa(issueNumber))
	values.Set("target", targetLabel)
	if trimmedIssueURL := strings.TrimSpace(issueURL); trimmedIssueURL != "" {
		values.Set("issue_url", trimmedIssueURL)
	}

	return baseURL + "/governance/labels-stages?" + values.Encode()
}

func renderStateMarker(state commentState) (string, error) {
	raw, err := json.Marshal(state)
	if err != nil {
		return "", fmt.Errorf("marshal run status marker: %w", err)
	}
	return commentMarkerPrefix + string(raw) + commentMarkerSuffix, nil
}
