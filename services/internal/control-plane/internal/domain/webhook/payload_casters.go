package webhook

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

type runPayloadInput struct {
	Command           IngestCommand
	Envelope          githubWebhookEnvelope
	ProjectID         string
	RepositoryID      string
	ServicesYAMLPath  string
	HasBinding        bool
	LearningMode      bool
	Trigger           *issueRunTrigger
	Agent             runAgentProfile
	RuntimeMode       agentdomain.RuntimeMode
	RuntimeSource     string
	RuntimeTargetEnv  string
	RuntimeNamespace  string
	RuntimeBuildRef   string
	RuntimeDeployOnly bool
}

type eventPayloadInput struct {
	Command  IngestCommand
	Envelope githubWebhookEnvelope
	Inserted bool
	RunID    string
	Trigger  *issueRunTrigger
}

type ignoredEventPayloadInput struct {
	Command           IngestCommand
	Envelope          githubWebhookEnvelope
	Reason            string
	RunKind           webhookdomain.TriggerKind
	HasBinding        bool
	ConflictingLabels []string
}

type ignoredWebhookParams struct {
	Reason            string
	RunKind           webhookdomain.TriggerKind
	HasBinding        bool
	ConflictingLabels []string
}

func buildRunPayload(input runPayloadInput) (json.RawMessage, error) {
	payload := githubRunPayload{
		Source:     "github",
		DeliveryID: input.Command.DeliveryID,
		EventType:  input.Command.EventType,
		ReceivedAt: input.Command.ReceivedAt.UTC().Format(time.RFC3339Nano),
		Repository: githubRunRepositoryPayload{
			ID:       input.Envelope.Repository.ID,
			FullName: input.Envelope.Repository.FullName,
			Name:     input.Envelope.Repository.Name,
			Private:  input.Envelope.Repository.Private,
			Fork:     input.Envelope.Repository.Fork,
		},
		Installation: githubInstallationPayload{
			ID: input.Envelope.Installation.ID,
		},
		Sender: githubActorPayload{
			ID:    input.Envelope.Sender.ID,
			Login: input.Envelope.Sender.Login,
		},
		Action:        input.Envelope.Action,
		RawPayload:    json.RawMessage(input.Command.Payload),
		CorrelationID: input.Command.CorrelationID,
		Project: githubRunProjectPayload{
			ID:              input.ProjectID,
			RepositoryID:    input.RepositoryID,
			ServicesYAML:    input.ServicesYAMLPath,
			BindingResolved: input.HasBinding,
		},
		LearningMode: input.LearningMode,
		Agent: githubRunAgentPayload{
			ID:   input.Agent.ID,
			Key:  input.Agent.Key,
			Name: input.Agent.Name,
		},
		Runtime: githubRunRuntimePayload{
			Mode:       strings.TrimSpace(string(input.RuntimeMode)),
			Source:     strings.TrimSpace(input.RuntimeSource),
			TargetEnv:  strings.TrimSpace(input.RuntimeTargetEnv),
			Namespace:  strings.TrimSpace(input.RuntimeNamespace),
			BuildRef:   strings.TrimSpace(input.RuntimeBuildRef),
			DeployOnly: input.RuntimeDeployOnly,
		},
	}

	if input.Envelope.Issue.Number > 0 {
		payload.Issue = &githubRunIssuePayload{
			ID:      input.Envelope.Issue.ID,
			Number:  input.Envelope.Issue.Number,
			Title:   input.Envelope.Issue.Title,
			HTMLURL: input.Envelope.Issue.HTMLURL,
			State:   input.Envelope.Issue.State,
			User: githubActorPayload{
				ID:    input.Envelope.Issue.User.ID,
				Login: input.Envelope.Issue.User.Login,
			},
		}
		if input.Envelope.Issue.PullRequest != nil {
			payload.Issue.PullRequest = &githubPullRequestPayload{
				URL:     input.Envelope.Issue.PullRequest.URL,
				HTMLURL: input.Envelope.Issue.PullRequest.HTMLURL,
			}
		}
	} else if input.Envelope.PullRequest.Number > 0 {
		payload.Issue = &githubRunIssuePayload{
			ID:      input.Envelope.PullRequest.ID,
			Number:  input.Envelope.PullRequest.Number,
			Title:   input.Envelope.PullRequest.Title,
			HTMLURL: input.Envelope.PullRequest.HTMLURL,
			State:   input.Envelope.PullRequest.State,
			User: githubActorPayload{
				ID:    input.Envelope.PullRequest.User.ID,
				Login: input.Envelope.PullRequest.User.Login,
			},
			PullRequest: &githubPullRequestPayload{
				HTMLURL: input.Envelope.PullRequest.HTMLURL,
			},
		}
	}

	if input.Envelope.PullRequest.Number > 0 {
		payload.PullRequest = &githubRunPRPayload{
			ID:      input.Envelope.PullRequest.ID,
			Number:  input.Envelope.PullRequest.Number,
			Title:   input.Envelope.PullRequest.Title,
			HTMLURL: input.Envelope.PullRequest.HTMLURL,
			State:   input.Envelope.PullRequest.State,
			Head: githubRunPRRef{
				Ref: input.Envelope.PullRequest.Head.Ref,
			},
			Base: githubRunPRRef{
				Ref: input.Envelope.PullRequest.Base.Ref,
			},
			User: githubActorPayload{
				ID:    input.Envelope.PullRequest.User.ID,
				Login: input.Envelope.PullRequest.User.Login,
			},
		}
	}

	if input.Trigger != nil {
		payload.Trigger = &githubIssueTriggerPayload{
			Source: input.Trigger.Source,
			Label:  input.Trigger.Label,
			Kind:   input.Trigger.Kind,
		}
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal run payload: %w", err)
	}
	return raw, nil
}

func buildEventPayload(input eventPayloadInput) (json.RawMessage, error) {
	payload := buildBaseFlowEventPayload(input.Command, input.Envelope)
	payload.Inserted = &input.Inserted
	payload.RunID = input.RunID
	if input.Trigger != nil {
		payload.Label = input.Trigger.Label
		payload.RunKind = input.Trigger.Kind
	}
	if input.Envelope.Issue.Number > 0 {
		payload.IssueNumber = input.Envelope.Issue.Number
	} else if input.Envelope.PullRequest.Number > 0 {
		payload.IssueNumber = input.Envelope.PullRequest.Number
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal flow event payload: %w", err)
	}
	return raw, nil
}

func buildReceivedEventPayload(cmd IngestCommand, envelope githubWebhookEnvelope) (json.RawMessage, error) {
	payload := buildBaseFlowEventPayload(cmd, envelope)
	if envelope.Issue.Number > 0 {
		payload.IssueNumber = envelope.Issue.Number
	} else if envelope.PullRequest.Number > 0 {
		payload.IssueNumber = envelope.PullRequest.Number
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal flow event payload: %w", err)
	}
	return raw, nil
}

func buildIgnoredEventPayload(input ignoredEventPayloadInput) (json.RawMessage, error) {
	payload := buildBaseFlowEventPayload(input.Command, input.Envelope)
	payload.Reason = input.Reason
	payload.BindingResolved = &input.HasBinding
	payload.ConflictingLabels = input.ConflictingLabels

	if strings.TrimSpace(input.Envelope.Label.Name) != "" {
		payload.Label = input.Envelope.Label.Name
	}
	if strings.TrimSpace(string(input.RunKind)) != "" {
		payload.RunKind = input.RunKind
	}
	if input.Envelope.Issue.Number > 0 {
		payload.Issue = &githubIgnoredIssuePayload{
			ID:      input.Envelope.Issue.ID,
			Number:  input.Envelope.Issue.Number,
			Title:   input.Envelope.Issue.Title,
			HTMLURL: input.Envelope.Issue.HTMLURL,
			State:   input.Envelope.Issue.State,
		}
	} else if input.Envelope.PullRequest.Number > 0 {
		payload.Issue = &githubIgnoredIssuePayload{
			ID:      input.Envelope.PullRequest.ID,
			Number:  input.Envelope.PullRequest.Number,
			Title:   input.Envelope.PullRequest.Title,
			HTMLURL: input.Envelope.PullRequest.HTMLURL,
			State:   input.Envelope.PullRequest.State,
		}
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal ignored event payload: %w", err)
	}
	return raw, nil
}

func buildBaseFlowEventPayload(cmd IngestCommand, envelope githubWebhookEnvelope) githubFlowEventPayload {
	return githubFlowEventPayload{
		Source:        "github",
		DeliveryID:    cmd.DeliveryID,
		EventType:     cmd.EventType,
		Action:        envelope.Action,
		CorrelationID: cmd.CorrelationID,
		Sender: githubActorPayload{
			ID:    envelope.Sender.ID,
			Login: envelope.Sender.Login,
		},
		Repository: githubFlowRepositoryPayload{
			ID:       envelope.Repository.ID,
			FullName: envelope.Repository.FullName,
			Name:     envelope.Repository.Name,
		},
	}
}
