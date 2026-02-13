package runstatus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	"github.com/codex-k8s/codex-k8s/libs/go/k8s/joblauncher"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

// NewService creates run-status domain service.
func NewService(cfg Config, deps Dependencies) (*Service, error) {
	cfg.PublicBaseURL = strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/")
	cfg.DefaultLocale = normalizeLocale(cfg.DefaultLocale, localeEN)

	if cfg.PublicBaseURL == "" {
		return nil, fmt.Errorf("public base url is required")
	}
	if deps.Runs == nil {
		return nil, fmt.Errorf("runs repository is required")
	}
	if deps.Platform == nil {
		return nil, fmt.Errorf("platform token repository is required")
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
		platform:   deps.Platform,
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

	comments, err := s.listRunIssueComments(ctx, runCtx)
	if err != nil {
		return UpsertCommentResult{}, err
	}

	existingCommentID := int64(0)
	if existingComment, existingState, found := findRunStatusComment(comments, runID); found {
		existingCommentID = existingComment.ID
		currentState = mergeState(existingState, currentState)
	}

	body, err := renderCommentBody(currentState, s.buildRunManagementURL(runID))
	if err != nil {
		return UpsertCommentResult{}, err
	}

	var savedComment mcpdomain.GitHubIssueComment
	if existingCommentID > 0 {
		savedComment, err = s.github.EditIssueComment(ctx, mcpdomain.GitHubEditIssueCommentParams{
			Token:      runCtx.githubToken,
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
			Token:       runCtx.githubToken,
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
		CommentID:  savedComment.ID,
		CommentURL: savedComment.URL,
	}, nil
}

// DeleteRunNamespace deletes one managed run namespace and updates the run status comment.
func (s *Service) DeleteRunNamespace(ctx context.Context, params DeleteNamespaceParams) (DeleteNamespaceResult, error) {
	runID := strings.TrimSpace(params.RunID)
	if runID == "" {
		return DeleteNamespaceResult{}, errs.Validation{Field: "run_id", Msg: "is required"}
	}

	runCtx, err := s.loadRunContext(ctx, runID)
	if err != nil {
		if errors.Is(err, errRunNotFound) {
			return DeleteNamespaceResult{}, errs.Validation{Field: "run_id", Msg: "not found"}
		}
		return DeleteNamespaceResult{}, err
	}

	comments, err := s.listRunIssueComments(ctx, runCtx)
	if err != nil {
		return DeleteNamespaceResult{}, err
	}

	_, state, _ := findRunStatusComment(comments, runID)
	namespace := strings.TrimSpace(state.Namespace)
	if namespace == "" {
		fallbackNamespace, fallbackFound, fallbackErr := s.kubernetes.FindManagedRunNamespaceByRunID(ctx, runID)
		if fallbackErr != nil {
			return DeleteNamespaceResult{}, fmt.Errorf("find managed run namespace by run id: %w", fallbackErr)
		}
		if fallbackFound {
			namespace = strings.TrimSpace(fallbackNamespace)
		}
	}
	if namespace == "" {
		return DeleteNamespaceResult{}, errs.Validation{Field: "run_id", Msg: errRunNamespaceMissing.Error()}
	}

	deleted, err := s.kubernetes.DeleteManagedRunNamespace(ctx, namespace)
	if err != nil {
		return DeleteNamespaceResult{}, fmt.Errorf("delete managed run namespace: %w", err)
	}

	jobName := strings.TrimSpace(state.JobName)
	if jobName == "" {
		jobName = joblauncher.BuildRunJobName(runID)
	}
	jobNamespace := strings.TrimSpace(state.JobNamespace)
	if jobNamespace == "" {
		jobNamespace = namespace
	}
	runtimeMode := strings.TrimSpace(state.RuntimeMode)
	if runtimeMode == "" {
		runtimeMode = runtimeModeFullEnv
	}
	promptLocale := strings.TrimSpace(state.PromptLocale)
	if promptLocale == "" {
		promptLocale = s.cfg.DefaultLocale
	}

	upsertResult, upsertErr := s.UpsertRunStatusComment(ctx, UpsertCommentParams{
		RunID:          runID,
		Phase:          PhaseNamespaceDeleted,
		JobName:        jobName,
		JobNamespace:   jobNamespace,
		RuntimeMode:    runtimeMode,
		Namespace:      namespace,
		TriggerKind:    runCtx.triggerKind,
		PromptLocale:   promptLocale,
		RunStatus:      strings.TrimSpace(runCtx.run.Status),
		Deleted:        deleted,
		AlreadyDeleted: !deleted,
	})
	if upsertErr != nil {
		return DeleteNamespaceResult{}, upsertErr
	}

	requestedByType := normalizeRequestedByType(params.RequestedByType)
	requestedByID := strings.TrimSpace(params.RequestedByID)
	eventType := floweventdomain.EventTypeRunNamespaceDeleteByStaff
	if requestedByType == RequestedByTypeSystem {
		eventType = floweventdomain.EventTypeRunNamespaceDeleteBySystem
	}
	s.insertFlowEvent(ctx, runCtx.run.CorrelationID, eventType, runNamespaceDeleteByStaffPayload{
		RunID:              runID,
		Namespace:          namespace,
		Deleted:            deleted,
		AlreadyDeleted:     !deleted,
		RunStatusURL:       upsertResult.CommentURL,
		RunStatusCommentID: upsertResult.CommentID,
		RequestedByType:    string(requestedByType),
		RequestedByID:      requestedByID,
	})

	return DeleteNamespaceResult{
		RunID:          runID,
		Namespace:      namespace,
		Deleted:        deleted,
		AlreadyDeleted: !deleted,
		CommentURL:     upsertResult.CommentURL,
	}, nil
}

// GetRunRuntimeState returns run runtime details used by staff UI details page.
func (s *Service) GetRunRuntimeState(ctx context.Context, runID string) (RuntimeState, error) {
	trimmedRunID := strings.TrimSpace(runID)
	if trimmedRunID == "" {
		return RuntimeState{}, errs.Validation{Field: "run_id", Msg: "is required"}
	}

	runCtx, err := s.loadRunContext(ctx, trimmedRunID)
	if err != nil {
		if errors.Is(err, errRunNotFound) {
			return RuntimeState{}, errs.Validation{Field: "run_id", Msg: "not found"}
		}
		return RuntimeState{}, err
	}

	comments, err := s.listRunIssueComments(ctx, runCtx)
	if err != nil {
		return RuntimeState{}, err
	}

	_, state, found := findRunStatusComment(comments, trimmedRunID)
	result := RuntimeState{
		HasStatusComment: found,
		JobName:          strings.TrimSpace(state.JobName),
		JobNamespace:     strings.TrimSpace(state.JobNamespace),
		Namespace:        strings.TrimSpace(state.Namespace),
	}

	if result.Namespace == "" {
		namespace, namespaceFound, findErr := s.kubernetes.FindManagedRunNamespaceByRunID(ctx, trimmedRunID)
		if findErr != nil {
			return RuntimeState{}, fmt.Errorf("find managed run namespace by run id: %w", findErr)
		}
		if namespaceFound {
			result.Namespace = strings.TrimSpace(namespace)
		}
	}

	if result.Namespace != "" {
		exists, err := s.kubernetes.NamespaceExists(ctx, result.Namespace)
		if err != nil {
			return RuntimeState{}, fmt.Errorf("check namespace exists %s: %w", result.Namespace, err)
		}
		result.NamespaceExists = exists
	}

	if result.JobName == "" {
		result.JobName = joblauncher.BuildRunJobName(trimmedRunID)
	}
	if result.JobNamespace == "" {
		result.JobNamespace = result.Namespace
	}
	if result.JobName != "" && result.JobNamespace != "" {
		exists, err := s.kubernetes.JobExists(ctx, result.JobNamespace, result.JobName)
		if err != nil {
			return RuntimeState{}, fmt.Errorf("check job exists %s/%s: %w", result.JobNamespace, result.JobName, err)
		}
		result.JobExists = exists
	}

	return result, nil
}

// CleanupNamespacesByIssue removes preserved run namespaces when issue/PR gets closed.
func (s *Service) CleanupNamespacesByIssue(ctx context.Context, params CleanupByIssueParams) (CleanupByIssueResult, error) {
	return s.cleanupNamespacesByRepositoryReference(ctx, params.RepositoryFullName, params.IssueNumber, "issue_number", params.RequestedByID, s.runs.ListRunIDsByRepositoryIssue, "repository/issue")
}

// CleanupNamespacesByPullRequest removes preserved run namespaces when PR is closed/merged.
func (s *Service) CleanupNamespacesByPullRequest(ctx context.Context, params CleanupByPullRequestParams) (CleanupByIssueResult, error) {
	return s.cleanupNamespacesByRepositoryReference(ctx, params.RepositoryFullName, params.PRNumber, "pr_number", params.RequestedByID, s.runs.ListRunIDsByRepositoryPullRequest, "repository/pull request")
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
	token, err := s.loadBotToken(ctx)
	if err != nil {
		return runContext{}, err
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
		githubToken: token,
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

func (s *Service) listRunIssueComments(ctx context.Context, runCtx runContext) ([]mcpdomain.GitHubIssueComment, error) {
	comments, err := s.github.ListIssueComments(ctx, mcpdomain.GitHubListIssueCommentsParams{
		Token:       runCtx.githubToken,
		Owner:       runCtx.repoOwner,
		Repository:  runCtx.repoName,
		IssueNumber: runCtx.issueNumber,
		Limit:       200,
	})
	if err != nil {
		return nil, fmt.Errorf("list issue comments: %w", err)
	}
	return comments, nil
}

func (s *Service) loadBotToken(ctx context.Context) (string, error) {
	item, found, err := s.platform.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("load platform github tokens: %w", err)
	}
	if !found || len(item.BotTokenEncrypted) == 0 {
		return "", errRunBotTokenMissing
	}
	token, err := s.tokenCrypt.DecryptString(item.BotTokenEncrypted)
	if err != nil {
		return "", errRunBotTokenDecrypt
	}
	if strings.TrimSpace(token) == "" {
		return "", errRunBotTokenMissing
	}
	return token, nil
}

func (s *Service) buildRunManagementURL(runID string) string {
	id := strings.TrimSpace(runID)
	if s.cfg.PublicBaseURL == "" || id == "" {
		return ""
	}
	return s.cfg.PublicBaseURL + runManagementPathPrefix + id
}

func findRunStatusComment(comments []mcpdomain.GitHubIssueComment, runID string) (mcpdomain.GitHubIssueComment, commentState, bool) {
	for _, comment := range comments {
		if !commentContainsRunID(comment.Body, runID) {
			continue
		}
		state, ok := extractStateMarker(comment.Body)
		if !ok {
			return mcpdomain.GitHubIssueComment{}, commentState{}, false
		}
		return comment, state, true
	}
	return mcpdomain.GitHubIssueComment{}, commentState{}, false
}

func isIgnorableCleanupError(err error) bool {
	if err == nil {
		return false
	}

	var validationErr errs.Validation
	if errors.As(err, &validationErr) {
		msg := strings.ToLower(strings.TrimSpace(validationErr.Msg))
		if strings.Contains(msg, errRunStatusCommentNotFound.Error()) {
			return true
		}
		if strings.Contains(msg, errRunNamespaceMissing.Error()) {
			return true
		}
	}
	return false
}

func (s *Service) cleanupNamespacesByRunIDs(ctx context.Context, runIDs []string, requestedByID string) CleanupByIssueResult {
	result := CleanupByIssueResult{
		MatchedRuns: len(runIDs),
	}
	trimmedRequestedByID := strings.TrimSpace(requestedByID)
	if trimmedRequestedByID == "" {
		trimmedRequestedByID = "system"
	}

	for _, runID := range runIDs {
		deleteResult, err := s.DeleteRunNamespace(ctx, DeleteNamespaceParams{
			RunID:           runID,
			RequestedByType: RequestedByTypeSystem,
			RequestedByID:   trimmedRequestedByID,
		})
		if err != nil {
			if isIgnorableCleanupError(err) {
				result.SkippedRuns++
				continue
			}
			result.FailedRuns++
			continue
		}
		if deleteResult.Deleted {
			result.CleanedNamespaces++
		} else {
			result.AlreadyDeletedCount++
		}
	}

	return result
}

func (s *Service) cleanupNamespacesByRepositoryReference(ctx context.Context, repositoryFullName string, referenceNumber int64, referenceField string, requestedByID string, listFn func(context.Context, string, int64, int) ([]string, error), errorSuffix string) (CleanupByIssueResult, error) {
	normalizedRepositoryFullName := strings.TrimSpace(repositoryFullName)
	if normalizedRepositoryFullName == "" {
		return CleanupByIssueResult{}, errs.Validation{Field: "repository_full_name", Msg: "is required"}
	}
	if referenceNumber <= 0 {
		return CleanupByIssueResult{}, errs.Validation{Field: referenceField, Msg: "must be positive"}
	}

	runIDs, err := listFn(ctx, normalizedRepositoryFullName, referenceNumber, 200)
	if err != nil {
		return CleanupByIssueResult{}, fmt.Errorf("list runs by %s: %w", errorSuffix, err)
	}

	return s.cleanupNamespacesByRunIDs(ctx, runIDs, requestedByID), nil
}
