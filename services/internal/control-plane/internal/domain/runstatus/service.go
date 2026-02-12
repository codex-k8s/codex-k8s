package runstatus

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

// NewService creates run-status domain service.
func NewService(cfg Config, deps Dependencies) (*Service, error) {
	cfg.PublicBaseURL = strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/")
	cfg.TokenSigningKey = strings.TrimSpace(cfg.TokenSigningKey)
	cfg.DefaultLocale = normalizeLocale(cfg.DefaultLocale, localeEN)

	if cfg.PublicBaseURL == "" {
		return nil, fmt.Errorf("public base url is required")
	}
	if cfg.TokenSigningKey == "" {
		return nil, fmt.Errorf("token signing key is required")
	}
	if deps.Runs == nil {
		return nil, fmt.Errorf("runs repository is required")
	}
	if deps.Repos == nil {
		return nil, fmt.Errorf("repository config repository is required")
	}
	if deps.TokenCrypt == nil {
		return nil, fmt.Errorf("token crypt service is required")
	}
	if deps.GitHub == nil {
		return nil, fmt.Errorf("github client is required")
	}
	if deps.Kubernetes == nil {
		return nil, fmt.Errorf("kubernetes client is required")
	}

	return &Service{
		cfg:        cfg,
		runs:       deps.Runs,
		repos:      deps.Repos,
		tokenCrypt: deps.TokenCrypt,
		github:     deps.GitHub,
		kubernetes: deps.Kubernetes,
		flowEvents: deps.FlowEvents,
	}, nil
}

// UpsertRunStatusComment creates or updates one run status comment in the linked issue.
func (s *Service) UpsertRunStatusComment(ctx context.Context, params UpsertCommentParams) (UpsertCommentResult, error) {
	runID := strings.TrimSpace(params.RunID)
	if runID == "" {
		return UpsertCommentResult{}, fmt.Errorf("run_id is required")
	}
	if strings.TrimSpace(string(params.Phase)) == "" {
		return UpsertCommentResult{}, fmt.Errorf("phase is required")
	}

	runCtx, err := s.loadRunContext(ctx, runID)
	if err != nil {
		return UpsertCommentResult{}, err
	}

	currentState := commentState{
		RunID:          runID,
		Phase:          params.Phase,
		JobName:        strings.TrimSpace(params.JobName),
		JobNamespace:   strings.TrimSpace(params.JobNamespace),
		RuntimeMode:    normalizeRuntimeMode(params.RuntimeMode, params.TriggerKind),
		Namespace:      strings.TrimSpace(params.Namespace),
		TriggerKind:    normalizeTriggerKind(params.TriggerKind),
		PromptLocale:   normalizeLocale(params.PromptLocale, s.cfg.DefaultLocale),
		RunStatus:      strings.TrimSpace(params.RunStatus),
		Deleted:        params.Deleted,
		AlreadyDeleted: params.AlreadyDeleted,
	}

	comments, err := s.github.ListIssueComments(ctx, mcpdomain.GitHubListIssueCommentsParams{
		Token:       runCtx.repoToken,
		Owner:       runCtx.repoOwner,
		Repository:  runCtx.repoName,
		IssueNumber: runCtx.issueNumber,
		Limit:       200,
	})
	if err != nil {
		return UpsertCommentResult{}, fmt.Errorf("list issue comments: %w", err)
	}

	existingCommentID := int64(0)
	for _, comment := range comments {
		if !commentContainsRunID(comment.Body, runID) {
			continue
		}
		existingCommentID = comment.ID
		existingState, ok := extractStateMarker(comment.Body)
		if ok {
			currentState = mergeState(existingState, currentState)
		}
		break
	}

	deleteURL := ""
	if currentState.RuntimeMode == runtimeModeFullEnv && strings.TrimSpace(currentState.Namespace) != "" {
		token, err := s.signDeleteToken(deleteTokenPayload{
			RunID:     currentState.RunID,
			Namespace: currentState.Namespace,
		})
		if err != nil {
			return UpsertCommentResult{}, err
		}
		deleteURL = s.cfg.PublicBaseURL + deleteNamespacePath + token
	}

	body, err := renderCommentBody(currentState, deleteURL)
	if err != nil {
		return UpsertCommentResult{}, err
	}

	var savedComment mcpdomain.GitHubIssueComment
	if existingCommentID > 0 {
		savedComment, err = s.github.EditIssueComment(ctx, mcpdomain.GitHubEditIssueCommentParams{
			Token:      runCtx.repoToken,
			Owner:      runCtx.repoOwner,
			Repository: runCtx.repoName,
			CommentID:  existingCommentID,
			Body:       body,
		})
		if err != nil {
			return UpsertCommentResult{}, fmt.Errorf("edit run status issue comment: %w", err)
		}
	} else {
		savedComment, err = s.github.CreateIssueComment(ctx, mcpdomain.GitHubCreateIssueCommentParams{
			Token:       runCtx.repoToken,
			Owner:       runCtx.repoOwner,
			Repository:  runCtx.repoName,
			IssueNumber: runCtx.issueNumber,
			Body:        body,
		})
		if err != nil {
			return UpsertCommentResult{}, fmt.Errorf("create run status issue comment: %w", err)
		}
	}

	s.insertFlowEvent(ctx, runCtx.run.CorrelationID, floweventdomain.EventTypeRunStatusCommentUpserted, runStatusCommentUpsertedPayload{
		RunID:              runID,
		IssueNumber:        runCtx.issueNumber,
		RepositoryFullName: runCtx.payload.Repository.FullName,
		CommentID:          savedComment.ID,
		CommentURL:         savedComment.URL,
		Phase:              currentState.Phase,
	})

	return UpsertCommentResult{
		CommentID:          savedComment.ID,
		CommentURL:         savedComment.URL,
		DeleteNamespaceURL: deleteURL,
	}, nil
}

// DeleteRunNamespaceByToken verifies one signed token and deletes run namespace.
func (s *Service) DeleteRunNamespaceByToken(ctx context.Context, rawToken string) (DeleteByTokenResult, error) {
	payload, err := s.verifyDeleteToken(rawToken)
	if err != nil {
		return DeleteByTokenResult{}, err
	}

	runCtx, err := s.loadRunContext(ctx, payload.RunID)
	if err != nil {
		return DeleteByTokenResult{}, err
	}

	deleted, err := s.kubernetes.DeleteManagedRunNamespace(ctx, payload.Namespace)
	if err != nil {
		return DeleteByTokenResult{}, fmt.Errorf("delete managed run namespace: %w", err)
	}

	upsertResult, upsertErr := s.UpsertRunStatusComment(ctx, UpsertCommentParams{
		RunID:          payload.RunID,
		Phase:          PhaseNamespaceDeleted,
		JobName:        "",
		JobNamespace:   "",
		RuntimeMode:    runtimeModeFullEnv,
		Namespace:      payload.Namespace,
		TriggerKind:    runCtx.triggerKind,
		PromptLocale:   s.cfg.DefaultLocale,
		RunStatus:      strings.TrimSpace(runCtx.run.Status),
		Deleted:        deleted,
		AlreadyDeleted: !deleted,
	})
	if upsertErr != nil {
		return DeleteByTokenResult{}, upsertErr
	}

	s.insertFlowEvent(ctx, runCtx.run.CorrelationID, floweventdomain.EventTypeRunNamespaceDeleteByToken, runNamespaceDeleteByTokenPayload{
		RunID:              payload.RunID,
		Namespace:          payload.Namespace,
		Deleted:            deleted,
		AlreadyDeleted:     !deleted,
		RunStatusURL:       upsertResult.CommentURL,
		RunStatusCommentID: upsertResult.CommentID,
	})

	return DeleteByTokenResult{
		RunID:          payload.RunID,
		Namespace:      payload.Namespace,
		Deleted:        deleted,
		AlreadyDeleted: !deleted,
		CommentURL:     upsertResult.CommentURL,
	}, nil
}

func (s *Service) loadRunContext(ctx context.Context, runID string) (runContext, error) {
	run, ok, err := s.runs.GetByID(ctx, runID)
	if err != nil {
		return runContext{}, fmt.Errorf("get run by id: %w", err)
	}
	if !ok {
		return runContext{}, errRunNotFound
	}
	if len(run.RunPayload) == 0 {
		return runContext{}, errRunPayloadEmpty
	}

	var payload querytypes.RunPayload
	if err := json.Unmarshal(run.RunPayload, &payload); err != nil {
		return runContext{}, errRunPayloadDecode
	}
	issueNumber := 0
	if payload.Issue != nil {
		issueNumber = int(payload.Issue.Number)
	}
	if issueNumber <= 0 {
		return runContext{}, errRunIssueNumberMissing
	}

	repoOwner := ""
	repoName := ""
	fullName := strings.TrimSpace(payload.Repository.FullName)
	owner, name, ok := strings.Cut(fullName, "/")
	if ok {
		repoOwner = strings.TrimSpace(owner)
		repoName = strings.TrimSpace(name)
	}
	if repoOwner == "" || repoName == "" {
		return runContext{}, errRunRepoNameMissing
	}
	repositoryID := strings.TrimSpace(payload.Project.RepositoryID)
	if repositoryID == "" {
		return runContext{}, errRunRepoBindingRequired
	}

	tokenEncrypted, found, err := s.repos.GetTokenEncrypted(ctx, repositoryID)
	if err != nil {
		return runContext{}, fmt.Errorf("load repository token: %w", err)
	}
	if !found || len(tokenEncrypted) == 0 {
		return runContext{}, errRunRepoTokenMissing
	}
	token, err := s.tokenCrypt.DecryptString(tokenEncrypted)
	if err != nil {
		return runContext{}, errRunRepoTokenDecrypt
	}

	triggerKind := triggerKindDev
	if payload.Trigger != nil {
		triggerKind = normalizeTriggerKind(payload.Trigger.Kind)
	}

	return runContext{
		run:         run,
		payload:     payload,
		issueNumber: issueNumber,
		repoOwner:   repoOwner,
		repoName:    repoName,
		repoToken:   token,
		triggerKind: triggerKind,
	}, nil
}

func (s *Service) insertFlowEvent(ctx context.Context, correlationID string, eventType floweventdomain.EventType, payload any) {
	if s.flowEvents == nil || strings.TrimSpace(correlationID) == "" {
		return
	}
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_ = s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: correlationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorIDControlPlane,
		EventType:     eventType,
		Payload:       rawPayload,
		CreatedAt:     nowUTC(),
	})
}
