package mcp

import (
	"context"
	"fmt"
	"strings"
)

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
