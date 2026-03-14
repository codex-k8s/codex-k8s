package missioncontrolworker

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

type projectionRunContext struct {
	TriggerLabel      string
	RuntimeMode       string
	WaitReason        string
	LastHeartbeatAt   *time.Time
	IssueTitle        string
	IssueState        string
	IssueOwner        string
	IssueLabels       []string
	PullRequestTitle  string
	PullRequestState  string
	PullRequestAuthor string
	PullRequestLabels []string
	PullRequestHead   string
	PullRequestBase   string
	StageLabel        string
}

type projectionStoredRunPayload struct {
	RawPayload  json.RawMessage                `json:"raw_payload"`
	Issue       *projectionStoredIssuePayload  `json:"issue,omitempty"`
	PullRequest *projectionStoredPullRequest   `json:"pull_request,omitempty"`
	Trigger     *projectionStoredTrigger       `json:"trigger,omitempty"`
	Runtime     projectionStoredRuntimePayload `json:"runtime"`
}

type projectionStoredIssuePayload struct {
	Title string          `json:"title"`
	State string          `json:"state"`
	User  projectionActor `json:"user"`
}

type projectionStoredPullRequest struct {
	Title string          `json:"title"`
	State string          `json:"state"`
	Head  projectionRef   `json:"head"`
	Base  projectionRef   `json:"base"`
	User  projectionActor `json:"user"`
}

type projectionStoredTrigger struct {
	Label string `json:"label"`
}

type projectionStoredRuntimePayload struct {
	Mode string `json:"mode"`
}

type projectionRawWebhookEvent struct {
	Issue       *projectionRawIssuePayload       `json:"issue,omitempty"`
	PullRequest *projectionRawPullRequestPayload `json:"pull_request,omitempty"`
}

type projectionRawIssuePayload struct {
	Title  string            `json:"title"`
	State  string            `json:"state"`
	Labels []projectionLabel `json:"labels"`
	User   projectionActor   `json:"user"`
}

type projectionRawPullRequestPayload struct {
	Title  string            `json:"title"`
	State  string            `json:"state"`
	Labels []projectionLabel `json:"labels"`
	Head   projectionRef     `json:"head"`
	Base   projectionRef     `json:"base"`
	User   projectionActor   `json:"user"`
}

type projectionLabel struct {
	Name string `json:"name"`
}

type projectionActor struct {
	Login string `json:"login"`
}

type projectionRef struct {
	Ref string `json:"ref"`
}

func (s *Service) loadProjectionRunContext(ctx context.Context, runID string) (projectionRunContext, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return projectionRunContext{}, nil
	}

	out := projectionRunContext{}
	run, found, err := s.agentRuns.GetByID(ctx, runID)
	if err != nil {
		return projectionRunContext{}, err
	}
	if found {
		out = decodeProjectionRunContext(run.RunPayload)
	}

	staffRun, found, err := s.staffRuns.GetByID(ctx, runID)
	if err != nil {
		return projectionRunContext{}, err
	}
	if found {
		out.WaitReason = strings.TrimSpace(staffRun.WaitReason)
		out.LastHeartbeatAt = cloneProjectionTime(staffRun.LastHeartbeatAt)
	}

	out.StageLabel = resolveProjectionStageLabel(out.IssueLabels, out.PullRequestLabels, out.TriggerLabel)
	return out, nil
}

func decodeProjectionRunContext(raw json.RawMessage) projectionRunContext {
	if len(raw) == 0 {
		return projectionRunContext{}
	}

	var stored projectionStoredRunPayload
	if err := json.Unmarshal(raw, &stored); err != nil {
		return projectionRunContext{}
	}

	out := projectionRunContext{
		RuntimeMode: strings.TrimSpace(stored.Runtime.Mode),
	}
	if stored.Trigger != nil {
		out.TriggerLabel = strings.TrimSpace(stored.Trigger.Label)
	}
	if stored.Issue != nil {
		out.IssueTitle = strings.TrimSpace(stored.Issue.Title)
		out.IssueState = strings.TrimSpace(stored.Issue.State)
		out.IssueOwner = strings.TrimSpace(stored.Issue.User.Login)
	}
	if stored.PullRequest != nil {
		out.PullRequestTitle = strings.TrimSpace(stored.PullRequest.Title)
		out.PullRequestState = strings.TrimSpace(stored.PullRequest.State)
		out.PullRequestAuthor = strings.TrimSpace(stored.PullRequest.User.Login)
		out.PullRequestHead = strings.TrimSpace(stored.PullRequest.Head.Ref)
		out.PullRequestBase = strings.TrimSpace(stored.PullRequest.Base.Ref)
	}
	if len(stored.RawPayload) == 0 {
		return out
	}

	var event projectionRawWebhookEvent
	if err := json.Unmarshal(stored.RawPayload, &event); err != nil {
		return out
	}
	if event.Issue != nil {
		if out.IssueTitle == "" {
			out.IssueTitle = strings.TrimSpace(event.Issue.Title)
		}
		if out.IssueState == "" {
			out.IssueState = strings.TrimSpace(event.Issue.State)
		}
		if out.IssueOwner == "" {
			out.IssueOwner = strings.TrimSpace(event.Issue.User.Login)
		}
		out.IssueLabels = normalizeProjectionLabels(event.Issue.Labels)
	}
	if event.PullRequest != nil {
		if out.PullRequestTitle == "" {
			out.PullRequestTitle = strings.TrimSpace(event.PullRequest.Title)
		}
		if out.PullRequestState == "" {
			out.PullRequestState = strings.TrimSpace(event.PullRequest.State)
		}
		if out.PullRequestAuthor == "" {
			out.PullRequestAuthor = strings.TrimSpace(event.PullRequest.User.Login)
		}
		if out.PullRequestHead == "" {
			out.PullRequestHead = strings.TrimSpace(event.PullRequest.Head.Ref)
		}
		if out.PullRequestBase == "" {
			out.PullRequestBase = strings.TrimSpace(event.PullRequest.Base.Ref)
		}
		out.PullRequestLabels = normalizeProjectionLabels(event.PullRequest.Labels)
	}
	return out
}

func normalizeProjectionLabels(labels []projectionLabel) []string {
	if len(labels) == 0 {
		return nil
	}
	out := make([]string, 0, len(labels))
	seen := make(map[string]struct{}, len(labels))
	for _, item := range labels {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func resolveProjectionStageLabel(issueLabels []string, pullRequestLabels []string, triggerLabel string) string {
	if label := firstProjectionRunLabel(issueLabels); label != "" {
		return label
	}
	if label := firstProjectionRunLabel(pullRequestLabels); label != "" {
		return label
	}
	triggerLabel = strings.TrimSpace(triggerLabel)
	if strings.HasPrefix(triggerLabel, "run:") {
		return triggerLabel
	}
	return ""
}

func firstProjectionRunLabel(labels []string) string {
	for _, raw := range labels {
		label := strings.TrimSpace(raw)
		if strings.HasPrefix(label, "run:") {
			return label
		}
	}
	return ""
}

func cloneProjectionTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	resolved := value.UTC()
	return &resolved
}
