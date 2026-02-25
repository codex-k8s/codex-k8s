package webhook

import (
	"encoding/json"
	"time"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

// IngestCommand is a normalized webhook command accepted by the domain service.
type IngestCommand struct {
	// CorrelationID is a deduplication key shared across flow records.
	CorrelationID string
	// EventType is a provider event name from request headers.
	EventType string
	// DeliveryID is a provider delivery identifier from webhook headers.
	DeliveryID string
	// ReceivedAt is the time when ingress received the webhook.
	ReceivedAt time.Time
	// Payload is raw webhook JSON body.
	Payload json.RawMessage
}

// IngestResult is a transport-facing outcome of webhook ingestion.
type IngestResult struct {
	// CorrelationID mirrors the request correlation id.
	CorrelationID string `json:"correlation_id"`
	// RunID references an agent run linked to this webhook.
	RunID string `json:"run_id,omitempty"`
	// Status is accepted, duplicate, or ignored.
	Status webhookdomain.IngestStatus `json:"status"`
	// Duplicate is true when webhook was already processed before.
	Duplicate bool `json:"duplicate"`
}

// TriggerLabels defines active run:* labels that create stage runs.
type TriggerLabels struct {
	RunIntake, RunIntakeRevise string
	RunVision, RunVisionRevise string
	RunPRD, RunPRDRevise       string
	RunArch, RunArchRevise     string
	RunDesign, RunDesignRevise string
	RunPlan, RunPlanRevise     string
	RunDev, RunDevRevise       string
	RunDocAudit, RunAIRepair   string
	RunQA                      string
	RunRelease, RunPostDeploy  string
	RunOps, RunSelfImprove     string
	RunRethink                 string
	NeedReviewer               string
}

func defaultTriggerLabels() TriggerLabels {
	return TriggerLabels{
		RunIntake:       webhookdomain.DefaultRunIntakeLabel,
		RunIntakeRevise: webhookdomain.DefaultRunIntakeReviseLabel,
		RunVision:       webhookdomain.DefaultRunVisionLabel,
		RunVisionRevise: webhookdomain.DefaultRunVisionReviseLabel,
		RunPRD:          webhookdomain.DefaultRunPRDLabel,
		RunPRDRevise:    webhookdomain.DefaultRunPRDReviseLabel,
		RunArch:         webhookdomain.DefaultRunArchLabel,
		RunArchRevise:   webhookdomain.DefaultRunArchReviseLabel,
		RunDesign:       webhookdomain.DefaultRunDesignLabel,
		RunDesignRevise: webhookdomain.DefaultRunDesignReviseLabel,
		RunPlan:         webhookdomain.DefaultRunPlanLabel,
		RunPlanRevise:   webhookdomain.DefaultRunPlanReviseLabel,
		RunDev:          webhookdomain.DefaultRunDevLabel,
		RunDevRevise:    webhookdomain.DefaultRunDevReviseLabel,
		RunDocAudit:     webhookdomain.DefaultRunDocAuditLabel,
		RunAIRepair:     webhookdomain.DefaultRunAIRepairLabel,
		RunQA:           webhookdomain.DefaultRunQALabel,
		RunRelease:      webhookdomain.DefaultRunReleaseLabel,
		RunPostDeploy:   webhookdomain.DefaultRunPostDeployLabel,
		RunOps:          webhookdomain.DefaultRunOpsLabel,
		RunSelfImprove:  webhookdomain.DefaultRunSelfImproveLabel,
		RunRethink:      webhookdomain.DefaultRunRethinkLabel,
		NeedReviewer:    webhookdomain.DefaultNeedReviewerLabel,
	}
}

type issueRunTrigger struct {
	Source string
	Label  string
	Kind   webhookdomain.TriggerKind
}

type triggerConflictResult struct {
	ConflictingLabels []string
	IgnoreReason      string
	SuggestedLabels   []string
	ResolverSource    string
	ResolvedIssue     int64
}

type pullRequestReviewResolutionMeta struct {
	ReceivedChangesRequested bool
	ResolverSource           string
	ResolvedIssueNumber      int64
	SuggestedLabels          []string
}

type runAgentProfile struct {
	ID   string
	Key  string
	Name string
}

// githubWebhookEnvelope is a local transport DTO for fields used by the domain.
// It is intentionally minimal and independent from provider SDK structs.
type githubWebhookEnvelope struct {
	Action       string                   `json:"action"`
	Ref          string                   `json:"ref"`
	Before       string                   `json:"before"`
	After        string                   `json:"after"`
	Deleted      bool                     `json:"deleted"`
	Commits      []githubPushCommitRecord `json:"commits"`
	Installation githubInstallationRecord `json:"installation"`
	Repository   githubRepositoryRecord   `json:"repository"`
	Issue        githubIssueRecord        `json:"issue"`
	PullRequest  githubPullRequestRecord  `json:"pull_request"`
	Review       githubReviewRecord       `json:"review"`
	Label        githubLabelRecord        `json:"label"`
	Sender       githubActorRecord        `json:"sender"`
}

type githubInstallationRecord struct {
	ID int64 `json:"id"`
}

type githubRepositoryRecord struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
	Name     string `json:"name"`
	Private  bool   `json:"private"`
	Fork     bool   `json:"fork"`
}

type githubIssueRecord struct {
	ID          int64                 `json:"id"`
	Number      int64                 `json:"number"`
	Title       string                `json:"title"`
	HTMLURL     string                `json:"html_url"`
	State       string                `json:"state"`
	Labels      []githubLabelRecord   `json:"labels"`
	User        githubActorRecord     `json:"user"`
	PullRequest *githubPullRequestRef `json:"pull_request"`
}

type githubLabelRecord struct {
	Name string `json:"name"`
}

type githubActorRecord struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
}

type githubPullRequestRef struct {
	URL     string `json:"url"`
	HTMLURL string `json:"html_url"`
}

type githubPullRequestRecord struct {
	ID      int64                 `json:"id"`
	Number  int64                 `json:"number"`
	Title   string                `json:"title"`
	HTMLURL string                `json:"html_url"`
	State   string                `json:"state"`
	Labels  []githubLabelRecord   `json:"labels"`
	Head    githubPullRequestHead `json:"head"`
	User    githubActorRecord     `json:"user"`
	Base    githubPullRequestBase `json:"base"`
}

type githubPullRequestHead struct {
	Ref string `json:"ref"`
}

type githubPullRequestBase struct {
	Ref string `json:"ref"`
}

type githubReviewRecord struct {
	State string `json:"state"`
}

type githubPushCommitRecord struct {
	ID       string   `json:"id"`
	Added    []string `json:"added"`
	Modified []string `json:"modified"`
	Removed  []string `json:"removed"`
}
