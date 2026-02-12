package mcp

import (
	"context"
	"fmt"
	"strings"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
)

const (
	promptContextVersion = "v1"
	defaultIssueLimit    = 100
	maxIssueLimit        = 500
	defaultBranchLimit   = 100
	maxBranchLimit       = 200
	defaultK8sLimit      = 200
	maxK8sLimit          = 500
	defaultTailLines     = int64(200)
	maxTailLines         = int64(2000)
	defaultBaseBranch    = "main"
)

func (s *Service) PromptContext(ctx context.Context, session SessionContext) (PromptContextResult, error) {
	tool, err := s.toolCapability(ToolPromptContextGet)
	if err != nil {
		return PromptContextResult{}, err
	}

	runCtx, err := s.resolveRunContext(ctx, session, false)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return PromptContextResult{}, err
	}

	s.auditToolCalled(ctx, runCtx.Session, tool)

	issueCtx := buildPromptIssueContext(runCtx.Payload.Issue)
	result := PromptContextResult{
		Status: ToolExecutionStatusOK,
		Context: PromptContext{
			Version: promptContextVersion,
			Run: PromptRunContext{
				RunID:         runCtx.Session.RunID,
				CorrelationID: runCtx.Session.CorrelationID,
				ProjectID:     runCtx.Session.ProjectID,
				Namespace:     runCtx.Session.Namespace,
				RuntimeMode:   runCtx.Session.RuntimeMode,
			},
			Repository: PromptRepositoryContext{
				Provider:     runCtx.Repository.Provider,
				Owner:        runCtx.Repository.Owner,
				Name:         runCtx.Repository.Name,
				FullName:     runCtx.Repository.Owner + "/" + runCtx.Repository.Name,
				ServicesYAML: runCtx.Repository.ServicesYAMLPath,
			},
			Issue: issueCtx,
			Environment: PromptEnvironmentContext{
				ServiceName: s.cfg.ServerName,
				MCPBaseURL:  s.cfg.InternalMCPBaseURL,
			},
			Services: buildPromptServices(s.cfg.PublicBaseURL, s.cfg.InternalMCPBaseURL),
			MCP: PromptMCPContext{
				ServerName: s.cfg.ServerName,
				Tools:      s.ToolCatalog(),
			},
		},
	}

	s.auditPromptContextAssembled(ctx, runCtx)
	s.auditToolSucceeded(ctx, runCtx.Session, tool)
	return result, nil
}

func (s *Service) GitHubIssueGet(ctx context.Context, session SessionContext, input GitHubIssueGetInput) (GitHubIssueGetResult, error) {
	issue, err := githubIssueScopedRead(ctx, s, session, ToolGitHubIssueGet, input.IssueNumber, "github issue get", func(ctx context.Context, runCtx resolvedRunContext, issueNumber int) (GitHubIssue, error) {
		return s.github.GetIssue(ctx, GitHubGetIssueParams{
			Token:       runCtx.Token,
			Owner:       runCtx.Repository.Owner,
			Repository:  runCtx.Repository.Name,
			IssueNumber: issueNumber,
		})
	})
	if err != nil {
		return GitHubIssueGetResult{}, err
	}

	return GitHubIssueGetResult{
		Status: ToolExecutionStatusOK,
		Issue:  issue,
	}, nil
}

func (s *Service) GitHubPullRequestGet(ctx context.Context, session SessionContext, input GitHubPullRequestGetInput) (GitHubPullRequestGetResult, error) {
	tool, err := s.toolCapability(ToolGitHubPullRequestGet)
	if err != nil {
		return GitHubPullRequestGetResult{}, err
	}

	runCtx, err := s.resolveRunContext(ctx, session, true)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return GitHubPullRequestGetResult{}, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	if input.PullRequestNumber <= 0 {
		err := fmt.Errorf("pull_request_number is required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubPullRequestGetResult{}, err
	}

	pr, err := s.github.GetPullRequest(ctx, GitHubGetPullRequestParams{
		Token:             runCtx.Token,
		Owner:             runCtx.Repository.Owner,
		Repository:        runCtx.Repository.Name,
		PullRequestNumber: input.PullRequestNumber,
	})
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubPullRequestGetResult{}, fmt.Errorf("github pull request get: %w", err)
	}

	s.auditToolSucceeded(ctx, runCtx.Session, tool)
	return GitHubPullRequestGetResult{
		Status:      ToolExecutionStatusOK,
		PullRequest: pr,
	}, nil
}

func (s *Service) GitHubIssueCommentsList(ctx context.Context, session SessionContext, input GitHubIssueCommentsListInput) (GitHubIssueCommentsListResult, error) {
	tool, err := s.toolCapability(ToolGitHubIssueComments)
	if err != nil {
		return GitHubIssueCommentsListResult{}, err
	}

	runCtx, err := s.resolveRunContext(ctx, session, true)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return GitHubIssueCommentsListResult{}, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	issueNumber, err := resolveIssueNumber(input.IssueNumber, runCtx.Payload)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubIssueCommentsListResult{}, err
	}

	comments, err := s.github.ListIssueComments(ctx, GitHubListIssueCommentsParams{
		Token:       runCtx.Token,
		Owner:       runCtx.Repository.Owner,
		Repository:  runCtx.Repository.Name,
		IssueNumber: issueNumber,
		Limit:       clampLimit(input.Limit, defaultIssueLimit, maxIssueLimit),
	})
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubIssueCommentsListResult{}, fmt.Errorf("github issue comments list: %w", err)
	}

	s.auditToolSucceeded(ctx, runCtx.Session, tool)
	return GitHubIssueCommentsListResult{
		Status:   ToolExecutionStatusOK,
		Comments: comments,
	}, nil
}

func (s *Service) GitHubLabelsList(ctx context.Context, session SessionContext, input GitHubLabelsListInput) (GitHubLabelsListResult, error) {
	labels, err := githubIssueScopedRead(ctx, s, session, ToolGitHubLabelsList, input.IssueNumber, "github labels list", func(ctx context.Context, runCtx resolvedRunContext, issueNumber int) ([]GitHubLabel, error) {
		return s.github.ListIssueLabels(ctx, GitHubListIssueLabelsParams{
			Token:       runCtx.Token,
			Owner:       runCtx.Repository.Owner,
			Repository:  runCtx.Repository.Name,
			IssueNumber: issueNumber,
		})
	})
	if err != nil {
		return GitHubLabelsListResult{}, err
	}

	return GitHubLabelsListResult{
		Status: ToolExecutionStatusOK,
		Labels: labels,
	}, nil
}

func (s *Service) GitHubBranchesList(ctx context.Context, session SessionContext, input GitHubBranchesListInput) (GitHubBranchesListResult, error) {
	tool, err := s.toolCapability(ToolGitHubBranchesList)
	if err != nil {
		return GitHubBranchesListResult{}, err
	}

	runCtx, err := s.resolveRunContext(ctx, session, true)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return GitHubBranchesListResult{}, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	branches, err := s.github.ListBranches(ctx, GitHubListBranchesParams{
		Token:      runCtx.Token,
		Owner:      runCtx.Repository.Owner,
		Repository: runCtx.Repository.Name,
		Limit:      clampLimit(input.Limit, defaultBranchLimit, maxBranchLimit),
	})
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubBranchesListResult{}, fmt.Errorf("github branches list: %w", err)
	}

	s.auditToolSucceeded(ctx, runCtx.Session, tool)
	return GitHubBranchesListResult{
		Status:   ToolExecutionStatusOK,
		Branches: branches,
	}, nil
}

func (s *Service) GitHubBranchEnsure(ctx context.Context, session SessionContext, input GitHubBranchEnsureInput) (GitHubBranchEnsureResult, error) {
	tool, err := s.toolCapability(ToolGitHubBranchEnsure)
	if err != nil {
		return GitHubBranchEnsureResult{}, err
	}

	runCtx, err := s.resolveRunContext(ctx, session, true)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return GitHubBranchEnsureResult{}, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	branchName := strings.TrimSpace(input.BranchName)
	if branchName == "" {
		err := fmt.Errorf("branch_name is required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubBranchEnsureResult{}, err
	}
	baseBranch := strings.TrimSpace(input.BaseBranch)
	if baseBranch == "" {
		baseBranch = defaultBaseBranch
	}

	branch, err := s.github.EnsureBranch(ctx, GitHubEnsureBranchParams{
		Token:      runCtx.Token,
		Owner:      runCtx.Repository.Owner,
		Repository: runCtx.Repository.Name,
		BranchName: branchName,
		BaseBranch: baseBranch,
		BaseSHA:    strings.TrimSpace(input.BaseSHA),
		Force:      input.Force,
	})
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubBranchEnsureResult{}, fmt.Errorf("github ensure branch: %w", err)
	}

	s.auditToolSucceeded(ctx, runCtx.Session, tool)
	return GitHubBranchEnsureResult{
		Status: ToolExecutionStatusOK,
		Branch: branch,
	}, nil
}

func (s *Service) GitHubPullRequestUpsert(ctx context.Context, session SessionContext, input GitHubPullRequestUpsertInput) (GitHubPullRequestUpsertResult, error) {
	tool, err := s.toolCapability(ToolGitHubPullRequestUpsert)
	if err != nil {
		return GitHubPullRequestUpsertResult{}, err
	}

	runCtx, err := s.resolveRunContext(ctx, session, true)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return GitHubPullRequestUpsertResult{}, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	title := strings.TrimSpace(input.Title)
	headBranch := strings.TrimSpace(input.HeadBranch)
	if title == "" {
		err := fmt.Errorf("title is required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubPullRequestUpsertResult{}, err
	}
	if headBranch == "" {
		err := fmt.Errorf("head_branch is required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubPullRequestUpsertResult{}, err
	}
	baseBranch := strings.TrimSpace(input.BaseBranch)
	if baseBranch == "" {
		baseBranch = defaultBaseBranch
	}

	pr, err := s.github.UpsertPullRequest(ctx, GitHubUpsertPullRequestParams{
		Token:             runCtx.Token,
		Owner:             runCtx.Repository.Owner,
		Repository:        runCtx.Repository.Name,
		PullRequestNumber: input.PullRequestNumber,
		Title:             title,
		Body:              input.Body,
		HeadBranch:        headBranch,
		BaseBranch:        baseBranch,
		Draft:             input.Draft,
	})
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubPullRequestUpsertResult{}, fmt.Errorf("github pull request upsert: %w", err)
	}

	s.auditToolSucceeded(ctx, runCtx.Session, tool)
	return GitHubPullRequestUpsertResult{
		Status:      ToolExecutionStatusOK,
		PullRequest: pr,
	}, nil
}

func (s *Service) GitHubIssueCommentCreate(ctx context.Context, session SessionContext, input GitHubIssueCommentCreateInput) (GitHubIssueCommentCreateResult, error) {
	tool, err := s.toolCapability(ToolGitHubIssueCommentCreate)
	if err != nil {
		return GitHubIssueCommentCreateResult{}, err
	}

	runCtx, err := s.resolveRunContext(ctx, session, true)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return GitHubIssueCommentCreateResult{}, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	body := strings.TrimSpace(input.Body)
	if body == "" {
		err := fmt.Errorf("body is required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubIssueCommentCreateResult{}, err
	}
	issueNumber, err := resolveIssueNumber(input.IssueNumber, runCtx.Payload)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubIssueCommentCreateResult{}, err
	}

	comment, err := s.github.CreateIssueComment(ctx, GitHubCreateIssueCommentParams{
		Token:       runCtx.Token,
		Owner:       runCtx.Repository.Owner,
		Repository:  runCtx.Repository.Name,
		IssueNumber: issueNumber,
		Body:        body,
	})
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubIssueCommentCreateResult{}, fmt.Errorf("github issue comment create: %w", err)
	}

	s.auditToolSucceeded(ctx, runCtx.Session, tool)
	return GitHubIssueCommentCreateResult{
		Status:  ToolExecutionStatusOK,
		Comment: comment,
	}, nil
}

func (s *Service) GitHubLabelsAdd(ctx context.Context, session SessionContext, input GitHubLabelsAddInput) (GitHubLabelsMutationResult, error) {
	return s.githubLabelsMutate(ctx, session, ToolGitHubLabelsAdd, input.IssueNumber, input.Labels, "github add labels", s.github.AddLabels)
}

func (s *Service) GitHubLabelsRemove(ctx context.Context, session SessionContext, input GitHubLabelsRemoveInput) (GitHubLabelsMutationResult, error) {
	return s.githubLabelsMutate(ctx, session, ToolGitHubLabelsRemove, input.IssueNumber, input.Labels, "github remove labels", s.github.RemoveLabels)
}

func (s *Service) KubernetesPodsList(ctx context.Context, session SessionContext, input KubernetesPodsListInput) (KubernetesPodsListResult, error) {
	return newKubernetesPodsListResult(
		kubernetesList(ctx, s, session, ToolKubernetesPodsList, input.Limit, "kubernetes pods list", s.kubernetes.ListPods),
	)
}

func (s *Service) KubernetesEventsList(ctx context.Context, session SessionContext, input KubernetesEventsListInput) (KubernetesEventsListResult, error) {
	events, err := kubernetesList(ctx, s, session, ToolKubernetesEventsList, input.Limit, "kubernetes events list", s.kubernetes.ListEvents)
	if err != nil {
		return KubernetesEventsListResult{}, err
	}

	return KubernetesEventsListResult{Status: ToolExecutionStatusOK, Events: events}, nil
}

func newKubernetesPodsListResult(pods []KubernetesPod, err error) (KubernetesPodsListResult, error) {
	if err != nil {
		return KubernetesPodsListResult{}, err
	}

	return KubernetesPodsListResult{Status: ToolExecutionStatusOK, Pods: pods}, nil
}

func (s *Service) KubernetesPodLogsGet(ctx context.Context, session SessionContext, input KubernetesPodLogsGetInput) (KubernetesPodLogsGetResult, error) {
	tool, err := s.toolCapability(ToolKubernetesPodLogsGet)
	if err != nil {
		return KubernetesPodLogsGetResult{}, err
	}
	if err := requireRuntimeNamespace(session); err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return KubernetesPodLogsGetResult{}, err
	}
	podName := strings.TrimSpace(input.Pod)
	if podName == "" {
		err := fmt.Errorf("pod is required")
		s.auditToolFailed(ctx, session, tool, err)
		return KubernetesPodLogsGetResult{}, err
	}

	s.auditToolCalled(ctx, session, tool)
	logs, err := s.kubernetes.GetPodLogs(ctx, session.Namespace, podName, strings.TrimSpace(input.Container), clampTail(input.TailLines))
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return KubernetesPodLogsGetResult{}, fmt.Errorf("kubernetes pod logs get: %w", err)
	}
	s.auditToolSucceeded(ctx, session, tool)
	return KubernetesPodLogsGetResult{
		Status: ToolExecutionStatusOK,
		Logs:   logs,
	}, nil
}

func (s *Service) KubernetesPodExec(ctx context.Context, session SessionContext, input KubernetesPodExecInput) (KubernetesPodExecToolResult, error) {
	tool, err := s.toolCapability(ToolKubernetesPodExec)
	if err != nil {
		return KubernetesPodExecToolResult{}, err
	}
	if err := requireRuntimeNamespace(session); err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return KubernetesPodExecToolResult{}, err
	}
	podName := strings.TrimSpace(input.Pod)
	if podName == "" {
		err := fmt.Errorf("pod is required")
		s.auditToolFailed(ctx, session, tool, err)
		return KubernetesPodExecToolResult{}, err
	}
	if len(input.Command) == 0 {
		err := fmt.Errorf("command is required")
		s.auditToolFailed(ctx, session, tool, err)
		return KubernetesPodExecToolResult{}, err
	}

	s.auditToolCalled(ctx, session, tool)
	execResult, err := s.kubernetes.ExecPod(ctx, session.Namespace, podName, strings.TrimSpace(input.Container), input.Command)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return KubernetesPodExecToolResult{}, fmt.Errorf("kubernetes pod exec: %w", err)
	}
	s.auditToolSucceeded(ctx, session, tool)
	return KubernetesPodExecToolResult{
		Status: ToolExecutionStatusOK,
		Exec:   execResult,
	}, nil
}

func (s *Service) KubernetesManifestApply(ctx context.Context, session SessionContext, _ KubernetesManifestApplyInput) (ApprovalRequiredResult, error) {
	return s.kubernetesApprovalOnly(ctx, session, ToolKubernetesManifestApply, "approval is required by policy before manifest apply")
}

func (s *Service) KubernetesManifestDelete(ctx context.Context, session SessionContext, _ KubernetesManifestDeleteInput) (ApprovalRequiredResult, error) {
	return s.kubernetesApprovalOnly(ctx, session, ToolKubernetesManifestDelete, "approval is required by policy before manifest delete")
}

func (s *Service) resolveGitHubIssueRunContext(ctx context.Context, session SessionContext, tool ToolCapability, explicitIssue int) (resolvedRunContext, int, error) {
	runCtx, err := s.resolveRunContext(ctx, session, true)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return resolvedRunContext{}, 0, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	issueNumber, err := resolveIssueNumber(explicitIssue, runCtx.Payload)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return resolvedRunContext{}, 0, err
	}
	return runCtx, issueNumber, nil
}

func githubIssueScopedRead[T any](
	ctx context.Context,
	svc *Service,
	session SessionContext,
	toolName ToolName,
	explicitIssue int,
	errorPrefix string,
	readFn func(context.Context, resolvedRunContext, int) (T, error),
) (T, error) {
	var zero T

	tool, err := svc.toolCapability(toolName)
	if err != nil {
		return zero, err
	}

	runCtx, issueNumber, err := svc.resolveGitHubIssueRunContext(ctx, session, tool, explicitIssue)
	if err != nil {
		return zero, err
	}

	value, err := readFn(ctx, runCtx, issueNumber)
	if err != nil {
		svc.auditToolFailed(ctx, runCtx.Session, tool, err)
		return zero, fmt.Errorf("%s: %w", errorPrefix, err)
	}

	svc.auditToolSucceeded(ctx, runCtx.Session, tool)
	return value, nil
}

func (s *Service) githubLabelsMutate(
	ctx context.Context,
	session SessionContext,
	toolName ToolName,
	issueNumber int,
	labels []string,
	errorPrefix string,
	mutate func(context.Context, GitHubMutateLabelsParams) ([]GitHubLabel, error),
) (GitHubLabelsMutationResult, error) {
	tool, err := s.toolCapability(toolName)
	if err != nil {
		return GitHubLabelsMutationResult{}, err
	}

	runCtx, resolvedIssueNumber, err := s.resolveGitHubIssueRunContext(ctx, session, tool, issueNumber)
	if err != nil {
		return GitHubLabelsMutationResult{}, err
	}
	if tool.Approval == ToolApprovalRequired {
		message := "approval is required by policy before labels mutation"
		s.auditToolApprovalPending(ctx, runCtx.Session, tool, message)
		return GitHubLabelsMutationResult{
			Status:  ToolExecutionStatusApprovalRequired,
			Message: message,
		}, nil
	}

	normalizedLabels := normalizeLabels(labels)
	if len(normalizedLabels) == 0 {
		err := fmt.Errorf("labels are required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubLabelsMutationResult{}, err
	}

	mutated, err := mutate(ctx, GitHubMutateLabelsParams{
		Token:       runCtx.Token,
		Owner:       runCtx.Repository.Owner,
		Repository:  runCtx.Repository.Name,
		IssueNumber: resolvedIssueNumber,
		Labels:      normalizedLabels,
	})
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubLabelsMutationResult{}, fmt.Errorf("%s: %w", errorPrefix, err)
	}

	s.auditToolSucceeded(ctx, runCtx.Session, tool)
	return GitHubLabelsMutationResult{
		Status: ToolExecutionStatusOK,
		Labels: mutated,
	}, nil
}

func kubernetesList[T any](
	ctx context.Context,
	svc *Service,
	session SessionContext,
	toolName ToolName,
	limit int,
	errorPrefix string,
	listFn func(context.Context, string, int) ([]T, error),
) ([]T, error) {
	tool, err := svc.toolCapability(toolName)
	if err != nil {
		return nil, err
	}
	if err := requireRuntimeNamespace(session); err != nil {
		svc.auditToolFailed(ctx, session, tool, err)
		return nil, err
	}

	svc.auditToolCalled(ctx, session, tool)
	items, err := listFn(ctx, session.Namespace, clampLimit(limit, defaultK8sLimit, maxK8sLimit))
	if err != nil {
		svc.auditToolFailed(ctx, session, tool, err)
		return nil, fmt.Errorf("%s: %w", errorPrefix, err)
	}
	svc.auditToolSucceeded(ctx, session, tool)
	return items, nil
}

func (s *Service) kubernetesApprovalOnly(ctx context.Context, session SessionContext, toolName ToolName, message string) (ApprovalRequiredResult, error) {
	tool, err := s.toolCapability(toolName)
	if err != nil {
		return ApprovalRequiredResult{}, err
	}
	if err := requireRuntimeNamespace(session); err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return ApprovalRequiredResult{}, err
	}

	s.auditToolCalled(ctx, session, tool)
	s.auditToolApprovalPending(ctx, session, tool, message)
	return ApprovalRequiredResult{
		Status:  ToolExecutionStatusApprovalRequired,
		Tool:    tool.Name,
		Message: message,
	}, nil
}

func (s *Service) resolveRunContext(ctx context.Context, session SessionContext, requireRepoToken bool) (resolvedRunContext, error) {
	runID := strings.TrimSpace(session.RunID)
	if runID == "" {
		return resolvedRunContext{}, fmt.Errorf("run_id is required")
	}

	run, ok, err := s.runs.GetByID(ctx, runID)
	if err != nil {
		return resolvedRunContext{}, fmt.Errorf("get run: %w", err)
	}
	if !ok {
		return resolvedRunContext{}, fmt.Errorf("run not found")
	}
	if !isRunActive(run.Status) {
		return resolvedRunContext{}, fmt.Errorf("run status %q is not active", run.Status)
	}

	payload, err := parseRunPayload(run.RunPayload)
	if err != nil {
		return resolvedRunContext{}, err
	}

	repositoryID := strings.TrimSpace(payload.Project.RepositoryID)
	if repositoryID == "" {
		return resolvedRunContext{}, fmt.Errorf("run payload missing repository_id")
	}

	repository, ok, err := s.repos.GetByID(ctx, repositoryID)
	if err != nil {
		return resolvedRunContext{}, fmt.Errorf("get repository binding: %w", err)
	}
	if !ok {
		return resolvedRunContext{}, fmt.Errorf("repository binding not found")
	}

	owner := strings.TrimSpace(repository.Owner)
	repoName := strings.TrimSpace(repository.Name)
	if owner == "" || repoName == "" {
		fallbackOwner, fallbackName := splitRepoFullName(payload.Repository.FullName)
		if owner == "" {
			owner = fallbackOwner
		}
		if repoName == "" {
			repoName = fallbackName
		}
	}
	if owner == "" || repoName == "" {
		return resolvedRunContext{}, fmt.Errorf("repository owner/name are required")
	}
	repository.Owner = owner
	repository.Name = repoName
	if strings.TrimSpace(repository.ServicesYAMLPath) == "" {
		repository.ServicesYAMLPath = "services.yaml"
	}

	token := ""
	if requireRepoToken {
		encrypted, ok, err := s.repos.GetTokenEncrypted(ctx, repository.ID)
		if err != nil {
			return resolvedRunContext{}, fmt.Errorf("get repository token: %w", err)
		}
		if !ok || len(encrypted) == 0 {
			return resolvedRunContext{}, fmt.Errorf("repository token is not configured")
		}
		token, err = s.tokenCrypt.DecryptString(encrypted)
		if err != nil {
			return resolvedRunContext{}, fmt.Errorf("decrypt repository token: %w", err)
		}
	}

	sessionContext := session
	if sessionContext.CorrelationID == "" {
		sessionContext.CorrelationID = run.CorrelationID
	}
	if sessionContext.ProjectID == "" {
		switch {
		case run.ProjectID != "":
			sessionContext.ProjectID = run.ProjectID
		case repository.ProjectID != "":
			sessionContext.ProjectID = repository.ProjectID
		case payload.Project.ID != "":
			sessionContext.ProjectID = payload.Project.ID
		}
	}
	if sessionContext.RuntimeMode == "" {
		triggerKind := ""
		if payload.Trigger != nil {
			triggerKind = payload.Trigger.Kind
		}
		sessionContext.RuntimeMode = parseRuntimeMode(triggerKind)
	}
	sessionContext.RuntimeMode = normalizeRuntimeMode(sessionContext.RuntimeMode)

	return resolvedRunContext{
		Session:    sessionContext,
		Run:        run,
		Repository: repository,
		Token:      token,
		Payload:    payload,
	}, nil
}

func (s *Service) toolCapability(name ToolName) (ToolCapability, error) {
	tool, ok := toolCapabilityByName(s.toolCatalog, name)
	if !ok {
		return ToolCapability{}, fmt.Errorf("tool %q is not registered", name)
	}
	return tool, nil
}

func buildPromptIssueContext(issue *runPayloadIssue) *PromptIssueContext {
	if issue == nil {
		return nil
	}
	if issue.Number <= 0 {
		return nil
	}
	return &PromptIssueContext{
		Number: issue.Number,
		Title:  strings.TrimSpace(issue.Title),
		State:  strings.TrimSpace(issue.State),
		URL:    strings.TrimSpace(issue.HTMLURL),
	}
}

func buildPromptServices(publicBaseURL string, internalMCPBaseURL string) []PromptServiceContext {
	items := []PromptServiceContext{
		{Name: "control-plane-grpc", Endpoint: "codex-k8s-control-plane:9090", Kind: "grpc"},
		{Name: "control-plane-mcp", Endpoint: strings.TrimSpace(internalMCPBaseURL), Kind: "mcp-http"},
		{Name: "worker", Endpoint: "codex-k8s-worker", Kind: "job-orchestrator"},
	}
	if strings.TrimSpace(publicBaseURL) != "" {
		items = append(items, PromptServiceContext{
			Name:     "api-gateway",
			Endpoint: strings.TrimSpace(publicBaseURL),
			Kind:     "http",
		})
	}
	return items
}

func requireRuntimeNamespace(session SessionContext) error {
	if session.RuntimeMode != agentdomain.RuntimeModeFullEnv {
		return fmt.Errorf("runtime_mode %q does not allow Kubernetes tools", session.RuntimeMode)
	}
	if strings.TrimSpace(session.Namespace) == "" {
		return fmt.Errorf("namespace is required for Kubernetes tools")
	}
	return nil
}

func resolveIssueNumber(explicit int, payload runPayload) (int, error) {
	if explicit > 0 {
		return explicit, nil
	}
	if payload.Issue != nil && payload.Issue.Number > 0 {
		return int(payload.Issue.Number), nil
	}
	return 0, fmt.Errorf("issue_number is required")
}

func clampTail(value int64) int64 {
	if value <= 0 {
		return defaultTailLines
	}
	if value > maxTailLines {
		return maxTailLines
	}
	return value
}

func normalizeLabels(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, value := range in {
		label := strings.TrimSpace(value)
		if label == "" {
			continue
		}
		key := strings.ToLower(label)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, label)
	}
	return out
}
