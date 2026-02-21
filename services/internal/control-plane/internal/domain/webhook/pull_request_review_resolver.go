package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	runstatusdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runstatus"
)

type pullRequestReviewStageCandidate struct {
	stage       string
	runLabel    string
	reviseLabel string
	kind        string
}

type pullRequestReviewResolveResult struct {
	trigger         issueRunTrigger
	ok              bool
	ignoreReason    string
	conflicting     []string
	suggestedLabels []string
	resolverSource  string
	resolvedIssue   int64
}

type stageLabelResolution struct {
	candidate       pullRequestReviewStageCandidate
	found           bool
	conflicting     []string
	suggestedLabels []string
}

func (s *Service) pullRequestReviewStageCatalog() []pullRequestReviewStageCandidate {
	return []pullRequestReviewStageCandidate{
		{stage: "intake", runLabel: strings.TrimSpace(s.triggerLabels.RunIntake), reviseLabel: strings.TrimSpace(s.triggerLabels.RunIntakeRevise), kind: "intake_revise"},
		{stage: "vision", runLabel: strings.TrimSpace(s.triggerLabels.RunVision), reviseLabel: strings.TrimSpace(s.triggerLabels.RunVisionRevise), kind: "vision_revise"},
		{stage: "prd", runLabel: strings.TrimSpace(s.triggerLabels.RunPRD), reviseLabel: strings.TrimSpace(s.triggerLabels.RunPRDRevise), kind: "prd_revise"},
		{stage: "arch", runLabel: strings.TrimSpace(s.triggerLabels.RunArch), reviseLabel: strings.TrimSpace(s.triggerLabels.RunArchRevise), kind: "arch_revise"},
		{stage: "design", runLabel: strings.TrimSpace(s.triggerLabels.RunDesign), reviseLabel: strings.TrimSpace(s.triggerLabels.RunDesignRevise), kind: "design_revise"},
		{stage: "plan", runLabel: strings.TrimSpace(s.triggerLabels.RunPlan), reviseLabel: strings.TrimSpace(s.triggerLabels.RunPlanRevise), kind: "plan_revise"},
		{stage: "dev", runLabel: strings.TrimSpace(s.triggerLabels.RunDev), reviseLabel: strings.TrimSpace(s.triggerLabels.RunDevRevise), kind: "dev_revise"},
	}
}

func resolveStageFromLabelRecords(labels []githubLabelRecord, catalog []pullRequestReviewStageCandidate) stageLabelResolution {
	matched := make([]pullRequestReviewStageCandidate, 0, 1)
	for _, stage := range catalog {
		if !hasAnyLabel(labels, stage.runLabel, stage.reviseLabel) {
			continue
		}
		matched = append(matched, stage)
	}

	if len(matched) == 0 {
		return stageLabelResolution{}
	}
	if len(matched) > 1 {
		labels := make([]string, 0, len(matched))
		for _, stage := range matched {
			labels = append(labels, stage.reviseLabel)
		}
		return stageLabelResolution{
			conflicting:     normalizeWebhookLabels(labels),
			suggestedLabels: normalizeWebhookLabels(labels),
		}
	}
	return stageLabelResolution{candidate: matched[0], found: true}
}

func supportedStageLabelHints(catalog []pullRequestReviewStageCandidate) []string {
	labels := make([]string, 0, len(catalog)*2)
	for _, item := range catalog {
		if strings.TrimSpace(item.runLabel) != "" {
			labels = append(labels, item.runLabel)
		}
		if strings.TrimSpace(item.reviseLabel) != "" {
			labels = append(labels, item.reviseLabel)
		}
	}
	return normalizeWebhookLabels(labels)
}

func (s *Service) resolvePullRequestReviewReviseTrigger(ctx context.Context, projectID string, envelope githubWebhookEnvelope) (pullRequestReviewResolveResult, error) {
	catalog := s.pullRequestReviewStageCatalog()
	repoFullName := strings.TrimSpace(envelope.Repository.FullName)
	prNumber := envelope.PullRequest.Number

	prStage := resolveStageFromLabelRecords(envelope.PullRequest.Labels, catalog)
	if len(prStage.conflicting) > 0 {
		return pullRequestReviewResolveResult{
			ignoreReason:    string(runstatusdomain.TriggerWarningReasonPullRequestReviewStageAmbiguous),
			conflicting:     prStage.conflicting,
			suggestedLabels: prStage.suggestedLabels,
			resolverSource:  "pr_labels",
		}, nil
	}
	if prStage.found {
		return pullRequestReviewResolveResult{
			trigger: issueRunTrigger{
				Source: "pull_request_review",
				Label:  strings.TrimSpace(prStage.candidate.reviseLabel),
				Kind:   webhookdomain.NormalizeTriggerKind(prStage.candidate.kind),
			},
			ok:             true,
			resolverSource: "pr_labels",
		}, nil
	}

	resolvedIssueNumber, ambiguousIssueNumbers, err := s.resolveLinkedIssueNumberFromHistory(ctx, projectID, repoFullName, prNumber)
	if err != nil {
		return pullRequestReviewResolveResult{}, fmt.Errorf("resolve linked issue number from history: %w", err)
	}
	if len(ambiguousIssueNumbers) > 1 {
		tokens := make([]string, 0, len(ambiguousIssueNumbers))
		for _, item := range ambiguousIssueNumbers {
			tokens = append(tokens, "issue#"+strconv.FormatInt(item, 10))
		}
		return pullRequestReviewResolveResult{
			ignoreReason:    string(runstatusdomain.TriggerWarningReasonPullRequestReviewStageAmbiguous),
			conflicting:     normalizeWebhookLabels(tokens),
			suggestedLabels: supportedStageLabelHints(catalog),
			resolverSource:  "run_context",
		}, nil
	}

	issueLabels := envelope.Issue.Labels
	if len(issueLabels) == 0 && resolvedIssueNumber > 0 {
		issueLabels, err = s.loadIssueLabelsFromRunHistory(ctx, projectID, repoFullName, resolvedIssueNumber, prNumber)
		if err != nil {
			return pullRequestReviewResolveResult{}, fmt.Errorf("load issue labels from history: %w", err)
		}
	}
	issueStage := resolveStageFromLabelRecords(issueLabels, catalog)
	if len(issueStage.conflicting) > 0 {
		return pullRequestReviewResolveResult{
			ignoreReason:    string(runstatusdomain.TriggerWarningReasonPullRequestReviewStageAmbiguous),
			conflicting:     issueStage.conflicting,
			suggestedLabels: issueStage.suggestedLabels,
			resolverSource:  "issue_labels",
			resolvedIssue:   resolvedIssueNumber,
		}, nil
	}
	if issueStage.found {
		return pullRequestReviewResolveResult{
			trigger: issueRunTrigger{
				Source: "pull_request_review",
				Label:  strings.TrimSpace(issueStage.candidate.reviseLabel),
				Kind:   webhookdomain.NormalizeTriggerKind(issueStage.candidate.kind),
			},
			ok:             true,
			resolverSource: "issue_labels",
			resolvedIssue:  resolvedIssueNumber,
		}, nil
	}

	if ctxStage, ok := s.resolveStageFromRunHistory(ctx, projectID, repoFullName, 0, prNumber, catalog); ok {
		return pullRequestReviewResolveResult{
			trigger: issueRunTrigger{
				Source: "pull_request_review",
				Label:  strings.TrimSpace(ctxStage.reviseLabel),
				Kind:   webhookdomain.NormalizeTriggerKind(ctxStage.kind),
			},
			ok:             true,
			resolverSource: "last_run_context",
			resolvedIssue:  resolvedIssueNumber,
		}, nil
	}

	if resolvedIssueNumber > 0 {
		if historyStage, ok := s.resolveStageFromRunHistory(ctx, projectID, repoFullName, resolvedIssueNumber, 0, catalog); ok {
			return pullRequestReviewResolveResult{
				trigger: issueRunTrigger{
					Source: "pull_request_review",
					Label:  strings.TrimSpace(historyStage.reviseLabel),
					Kind:   webhookdomain.NormalizeTriggerKind(historyStage.kind),
				},
				ok:             true,
				resolverSource: "issue_history",
				resolvedIssue:  resolvedIssueNumber,
			}, nil
		}
	}

	return pullRequestReviewResolveResult{
		ignoreReason:    string(runstatusdomain.TriggerWarningReasonPullRequestReviewStageNotResolved),
		suggestedLabels: supportedStageLabelHints(catalog),
		resolverSource:  "stage_not_resolved",
		resolvedIssue:   resolvedIssueNumber,
	}, nil
}

func (s *Service) resolveLinkedIssueNumberFromHistory(ctx context.Context, projectID string, repositoryFullName string, prNumber int64) (int64, []int64, error) {
	if s.agentRuns == nil || strings.TrimSpace(projectID) == "" || strings.TrimSpace(repositoryFullName) == "" || prNumber <= 0 {
		return 0, nil, nil
	}
	items, err := s.agentRuns.SearchRecentByProjectIssueOrPullRequest(ctx, projectID, repositoryFullName, 0, prNumber, 20)
	if err != nil {
		return 0, nil, err
	}
	unique := make([]int64, 0, 2)
	for _, item := range items {
		if item.IssueNumber <= 0 {
			continue
		}
		if !slices.Contains(unique, item.IssueNumber) {
			unique = append(unique, item.IssueNumber)
		}
	}
	sort.Slice(unique, func(i int, j int) bool { return unique[i] < unique[j] })
	if len(unique) == 1 {
		return unique[0], unique, nil
	}
	return 0, unique, nil
}

func (s *Service) resolveStageFromRunHistory(ctx context.Context, projectID string, repositoryFullName string, issueNumber int64, prNumber int64, catalog []pullRequestReviewStageCandidate) (pullRequestReviewStageCandidate, bool) {
	if s.agentRuns == nil || strings.TrimSpace(projectID) == "" || strings.TrimSpace(repositoryFullName) == "" {
		return pullRequestReviewStageCandidate{}, false
	}
	if issueNumber <= 0 && prNumber <= 0 {
		return pullRequestReviewStageCandidate{}, false
	}
	items, err := s.agentRuns.SearchRecentByProjectIssueOrPullRequest(ctx, projectID, repositoryFullName, issueNumber, prNumber, 50)
	if err != nil {
		return pullRequestReviewStageCandidate{}, false
	}
	for _, item := range items {
		stage, ok := resolveStageFromTriggerKind(item.TriggerKind, catalog)
		if !ok {
			continue
		}
		return stage, true
	}
	return pullRequestReviewStageCandidate{}, false
}

func resolveStageFromTriggerKind(triggerKind string, catalog []pullRequestReviewStageCandidate) (pullRequestReviewStageCandidate, bool) {
	normalized := string(webhookdomain.NormalizeTriggerKind(triggerKind))
	if normalized == "" {
		return pullRequestReviewStageCandidate{}, false
	}
	for _, item := range catalog {
		if normalized == item.kind || normalized == strings.TrimSuffix(item.kind, "_revise") {
			return item, true
		}
	}
	return pullRequestReviewStageCandidate{}, false
}

func (s *Service) loadIssueLabelsFromRunHistory(ctx context.Context, projectID string, repositoryFullName string, issueNumber int64, prNumber int64) ([]githubLabelRecord, error) {
	if s.agentRuns == nil || strings.TrimSpace(projectID) == "" || strings.TrimSpace(repositoryFullName) == "" || issueNumber <= 0 {
		return nil, nil
	}
	items, err := s.agentRuns.SearchRecentByProjectIssueOrPullRequest(ctx, projectID, repositoryFullName, issueNumber, prNumber, 50)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		runItem, found, runErr := s.agentRuns.GetByID(ctx, strings.TrimSpace(item.RunID))
		if runErr != nil || !found {
			continue
		}
		labels := extractIssueLabelsFromNormalizedRunPayload(runItem.RunPayload)
		if len(labels) > 0 {
			return labels, nil
		}
	}
	return nil, nil
}

func (s *Service) loadProfileHintsFromRunHistory(ctx context.Context, projectID string, repositoryFullName string, issueNumber int64, prNumber int64) (*githubRunProfileHints, error) {
	if s.agentRuns == nil || strings.TrimSpace(projectID) == "" || strings.TrimSpace(repositoryFullName) == "" {
		return nil, nil
	}
	if issueNumber <= 0 && prNumber <= 0 {
		return nil, nil
	}
	items, err := s.agentRuns.SearchRecentByProjectIssueOrPullRequest(ctx, projectID, repositoryFullName, issueNumber, prNumber, 50)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		runItem, found, runErr := s.agentRuns.GetByID(ctx, strings.TrimSpace(item.RunID))
		if runErr != nil || !found {
			continue
		}
		issueLabels, prLabels := extractIssueAndPullRequestLabelsFromNormalizedRunPayload(runItem.RunPayload)
		if len(issueLabels) == 0 && len(prLabels) == 0 {
			continue
		}
		return &githubRunProfileHints{
			LastRunIssueLabels:       normalizeWebhookLabels(issueLabels),
			LastRunPullRequestLabels: normalizeWebhookLabels(prLabels),
		}, nil
	}
	return nil, nil
}

type runRawPayloadEnvelope struct {
	RawPayload json.RawMessage `json:"raw_payload"`
}

type rawIssueLabelsEvent struct {
	Issue       *rawIssueLabelsItem       `json:"issue"`
	PullRequest *rawPullRequestLabelsItem `json:"pull_request"`
}

type rawIssueLabelsItem struct {
	Labels []githubLabelRecord `json:"labels"`
}

type rawPullRequestLabelsItem struct {
	Labels []githubLabelRecord `json:"labels"`
}

func extractIssueAndPullRequestLabelsFromNormalizedRunPayload(runPayload json.RawMessage) ([]string, []string) {
	if len(runPayload) == 0 {
		return nil, nil
	}
	var envelope runRawPayloadEnvelope
	if err := json.Unmarshal(runPayload, &envelope); err != nil {
		return nil, nil
	}
	if len(envelope.RawPayload) == 0 {
		return nil, nil
	}
	var event rawIssueLabelsEvent
	if err := json.Unmarshal(envelope.RawPayload, &event); err != nil {
		return nil, nil
	}
	issueLabels := make([]string, 0, 8)
	pullRequestLabels := make([]string, 0, 8)
	if event.Issue != nil {
		for _, item := range event.Issue.Labels {
			name := strings.TrimSpace(item.Name)
			if name == "" {
				continue
			}
			issueLabels = append(issueLabels, name)
		}
	}
	if event.PullRequest != nil {
		for _, item := range event.PullRequest.Labels {
			name := strings.TrimSpace(item.Name)
			if name == "" {
				continue
			}
			pullRequestLabels = append(pullRequestLabels, name)
		}
	}
	return issueLabels, pullRequestLabels
}

func extractIssueLabelsFromNormalizedRunPayload(runPayload json.RawMessage) []githubLabelRecord {
	issueLabels, _ := extractIssueAndPullRequestLabelsFromNormalizedRunPayload(runPayload)
	labels := make([]githubLabelRecord, 0, len(issueLabels))
	for _, raw := range issueLabels {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		labels = append(labels, githubLabelRecord{Name: name})
	}
	return labels
}
