package runstatus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	floweventdomain "github.com/codex-k8s/kodex/libs/go/domain/flowevent"
	"github.com/codex-k8s/kodex/libs/go/errs"
	"github.com/codex-k8s/kodex/libs/go/k8s/joblauncher"
	agentsessionrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/agentsession"
	floweventrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/flowevent"
	runtimedeploydomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/runtimedeploy"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
)

// CancelRun stops active runtime artifacts for one run, marks the run canceled,
// and prevents later resume paths from recreating continuation work.
func (s *Service) CancelRun(ctx context.Context, params CancelRunParams) (CancelRunResult, error) {
	runID := strings.TrimSpace(params.RunID)
	if runID == "" {
		return CancelRunResult{}, errs.Validation{Field: "run_id", Msg: "is required"}
	}

	runCtx, err := s.loadRunContext(ctx, runID)
	if err != nil {
		if errors.Is(err, errRunNotFound) {
			return CancelRunResult{}, errs.Validation{Field: "run_id", Msg: "not found"}
		}
		return CancelRunResult{}, err
	}

	previousStatus := normalizeRunStatus(runCtx.run.Status)
	alreadyTerminal := isRunTerminalStatus(previousStatus)
	if alreadyTerminal && previousStatus != runStatusCanceled {
		return CancelRunResult{
			RunID:           runID,
			PreviousStatus:  previousStatus,
			CurrentStatus:   previousStatus,
			AlreadyTerminal: true,
		}, nil
	}

	runtimeState, err := s.buildRunRuntimeStateForCancel(ctx, runCtx)
	if err != nil {
		return CancelRunResult{}, err
	}

	result := CancelRunResult{
		RunID:           runID,
		PreviousStatus:  previousStatus,
		CurrentStatus:   runStatusCanceled,
		AlreadyTerminal: alreadyTerminal,
	}

	result.RuntimeDeployCancelRequested, err = s.requestRunRuntimeDeployCancel(ctx, params, runID)
	if err != nil {
		return CancelRunResult{}, err
	}

	result.JobStopped, err = s.stopRunJobIfPresent(ctx, runtimeState)
	if err != nil {
		return CancelRunResult{}, err
	}

	if !alreadyTerminal {
		canceled, err := s.runs.CancelActiveByID(ctx, runID)
		if err != nil {
			return CancelRunResult{}, err
		}
		if !canceled {
			refreshedRun, found, lookupErr := s.runs.GetByID(ctx, runID)
			if lookupErr != nil {
				return CancelRunResult{}, fmt.Errorf("reload run after cancel: %w", lookupErr)
			}
			if !found {
				return CancelRunResult{}, errs.Validation{Field: "run_id", Msg: "not found"}
			}
			result.CurrentStatus = normalizeRunStatus(refreshedRun.Status)
			result.AlreadyTerminal = isRunTerminalStatus(result.CurrentStatus)
			if result.CurrentStatus != runStatusCanceled {
				return CancelRunResult{}, errs.Conflict{Msg: "run state changed during cancellation"}
			}
		}
	}

	if err := s.clearRunSessionWaitState(ctx, runID); err != nil {
		return CancelRunResult{}, err
	}

	result.CanceledGitHubWaits, err = s.cancelOpenGitHubRateLimitWaits(ctx, runID)
	if err != nil {
		return CancelRunResult{}, err
	}

	commentResult, err := s.upsertCanceledRunStatusComment(ctx, runCtx, runtimeState)
	if err != nil {
		return CancelRunResult{}, err
	}
	result.CommentURL = commentResult.CommentURL

	if !alreadyTerminal {
		s.insertRunCanceledFlowEvent(ctx, runCtx.run.CorrelationID, params, result, commentResult.CommentID)
	}

	return result, nil
}

func (s *Service) requestRunRuntimeDeployCancel(ctx context.Context, params CancelRunParams, runID string) (bool, error) {
	if s.runtimeDeploy == nil {
		return false, nil
	}

	actor := runtimedeploydomain.TaskActionActor{
		UserID:      strings.TrimSpace(params.RequestedByID),
		Email:       strings.TrimSpace(params.RequestedByEmail),
		GitHubLogin: strings.TrimSpace(params.RequestedByGitHub),
	}
	if actor.UserID == "" && actor.Email == "" && actor.GitHubLogin == "" {
		actor.UserID = string(floweventdomain.ActorIDControlPlane)
	}

	actionResult, err := s.runtimeDeploy.RequestTaskAction(ctx, runtimedeploydomain.TaskActionParams{
		RunID:  runID,
		Action: runtimedeploydomain.TaskActionCancel,
		Reason: strings.TrimSpace(params.Reason),
		Actor:  actor,
	})
	if err != nil {
		var notFound errs.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, fmt.Errorf("request runtime deploy cancel: %w", err)
	}

	return !actionResult.AlreadyTerminal, nil
}

func (s *Service) stopRunJobIfPresent(ctx context.Context, runtimeState RuntimeState) (bool, error) {
	jobName := strings.TrimSpace(runtimeState.JobName)
	jobNamespace := strings.TrimSpace(runtimeState.JobNamespace)
	if !runtimeState.JobExists || jobName == "" || jobNamespace == "" {
		return false, nil
	}
	if err := s.kubernetes.DeleteJobIfExists(ctx, jobNamespace, jobName); err != nil {
		return false, fmt.Errorf("delete run job %s/%s: %w", jobNamespace, jobName, err)
	}
	return true, nil
}

func (s *Service) clearRunSessionWaitState(ctx context.Context, runID string) error {
	if s.sessions == nil {
		return nil
	}
	_, err := s.sessions.SetWaitStateByRunID(ctx, agentsessionrepo.SetWaitStateParams{
		RunID:                runID,
		WaitState:            string(enumtypes.AgentSessionWaitStateNone),
		TimeoutGuardDisabled: false,
		LastHeartbeatAt:      nil,
	})
	if err != nil {
		return fmt.Errorf("clear run wait state: %w", err)
	}
	return nil
}

func (s *Service) cancelOpenGitHubRateLimitWaits(ctx context.Context, runID string) (int, error) {
	if s.githubRateLimitWaits == nil {
		return 0, nil
	}

	waits, err := s.githubRateLimitWaits.ListByRunID(ctx, runID)
	if err != nil {
		return 0, fmt.Errorf("list github rate-limit waits by run: %w", err)
	}

	now := nowUTC()
	canceled := 0
	for _, wait := range waits {
		if !wait.State.IsOpen() {
			continue
		}
		_, found, updateErr := s.githubRateLimitWaits.Update(ctx, querytypes.GitHubRateLimitWaitUpdateParams{
			ID:                     wait.ID,
			SignalOrigin:           enumtypes.GitHubRateLimitSignalOriginControlPlane,
			OperationClass:         wait.OperationClass,
			State:                  enumtypes.GitHubRateLimitWaitStateCancelled,
			LimitKind:              wait.LimitKind,
			Confidence:             wait.Confidence,
			RecoveryHintKind:       wait.RecoveryHintKind,
			SignalID:               wait.SignalID,
			RequestFingerprint:     wait.RequestFingerprint,
			CorrelationID:          wait.CorrelationID,
			ResumeActionKind:       wait.ResumeActionKind,
			ResumePayloadJSON:      wait.ResumePayloadJSON,
			ManualActionKind:       "",
			AutoResumeAttemptsUsed: wait.AutoResumeAttemptsUsed,
			MaxAutoResumeAttempts:  wait.MaxAutoResumeAttempts,
			ResumeNotBefore:        wait.ResumeNotBefore,
			LastResumeAttemptAt:    wait.LastResumeAttemptAt,
			LastSignalAt:           now,
			ResolvedAt:             &now,
		})
		if updateErr != nil {
			return 0, fmt.Errorf("cancel github rate-limit wait %s: %w", wait.ID, updateErr)
		}
		if found {
			canceled++
		}
	}

	if _, err := s.githubRateLimitWaits.RefreshRunProjection(ctx, runID); err != nil {
		return 0, fmt.Errorf("refresh github rate-limit projection after cancel: %w", err)
	}
	return canceled, nil
}

func (s *Service) upsertCanceledRunStatusComment(ctx context.Context, runCtx runContext, runtimeState RuntimeState) (UpsertCommentResult, error) {
	state, err := s.loadRunCommentState(ctx, runCtx, runCtx.run.ID)
	if err != nil {
		return UpsertCommentResult{}, err
	}

	triggerLabel := ""
	if runCtx.payload.Trigger != nil {
		triggerLabel = strings.TrimSpace(runCtx.payload.Trigger.Label)
	}
	discussionMode := runCtx.payload.DiscussionMode || isDiscussionTriggerLabel(triggerLabel)

	runtimeMode := strings.TrimSpace(state.RuntimeMode)
	if runtimeMode == "" && runCtx.payload.Runtime != nil {
		runtimeMode = strings.TrimSpace(runCtx.payload.Runtime.Mode)
	}
	runtimeMode = normalizeRuntimeMode(runtimeMode, resolveCommentTriggerKind(runCtx.triggerKind, triggerLabel, discussionMode))

	namespace := strings.TrimSpace(runtimeState.Namespace)
	if namespace == "" {
		namespace = strings.TrimSpace(state.Namespace)
	}
	if namespace == "" && runCtx.payload.Runtime != nil {
		namespace = strings.TrimSpace(runCtx.payload.Runtime.Namespace)
	}

	jobName := strings.TrimSpace(runtimeState.JobName)
	if jobName == "" {
		jobName = strings.TrimSpace(state.JobName)
	}
	if jobName == "" {
		jobName = joblauncher.BuildRunJobName(runCtx.run.ID)
	}

	jobNamespace := strings.TrimSpace(runtimeState.JobNamespace)
	if jobNamespace == "" {
		jobNamespace = strings.TrimSpace(state.JobNamespace)
	}
	if jobNamespace == "" {
		jobNamespace = namespace
	}

	promptLocale := strings.TrimSpace(state.PromptLocale)
	if promptLocale == "" {
		promptLocale = s.cfg.DefaultLocale
	}

	return s.UpsertRunStatusComment(ctx, UpsertCommentParams{
		RunID:        runCtx.run.ID,
		Phase:        PhaseFinished,
		JobName:      jobName,
		JobNamespace: jobNamespace,
		RuntimeMode:  runtimeMode,
		Namespace:    namespace,
		TriggerKind:  runCtx.triggerKind,
		PromptLocale: promptLocale,
		RunStatus:    runStatusCanceled,
	})
}

func (s *Service) buildRunRuntimeStateForCancel(ctx context.Context, runCtx runContext) (RuntimeState, error) {
	state, err := s.loadRunCommentState(ctx, runCtx, runCtx.run.ID)
	if err != nil {
		state = commentState{}
	}

	result := RuntimeState{
		HasStatusComment: strings.TrimSpace(state.RunID) != "",
		JobName:          strings.TrimSpace(state.JobName),
		JobNamespace:     strings.TrimSpace(state.JobNamespace),
		Namespace:        strings.TrimSpace(state.Namespace),
	}

	if result.Namespace == "" && runCtx.payload.Runtime != nil {
		result.Namespace = strings.TrimSpace(runCtx.payload.Runtime.Namespace)
	}
	if result.Namespace == "" {
		namespace, found, findErr := s.kubernetes.FindManagedRunNamespaceByRunID(ctx, runCtx.run.ID)
		if findErr != nil {
			return RuntimeState{}, fmt.Errorf("find managed run namespace by run id: %w", findErr)
		}
		if found {
			result.Namespace = strings.TrimSpace(namespace)
		}
	}

	if result.Namespace != "" {
		namespaceExists, namespaceErr := s.kubernetes.NamespaceExists(ctx, result.Namespace)
		if namespaceErr != nil {
			return RuntimeState{}, fmt.Errorf("check namespace exists %s: %w", result.Namespace, namespaceErr)
		}
		result.NamespaceExists = namespaceExists
	}

	if result.JobName == "" {
		result.JobName = joblauncher.BuildRunJobName(runCtx.run.ID)
	}
	if result.JobNamespace == "" {
		result.JobNamespace = result.Namespace
	}
	if err := s.populateRuntimeJobState(ctx, &result); err != nil {
		return RuntimeState{}, err
	}
	return result, nil
}

func (s *Service) loadRunCommentState(ctx context.Context, runCtx runContext, runID string) (commentState, error) {
	if !runCtx.hasCommentTarget() {
		return commentState{}, nil
	}
	comments, err := s.listRunIssueComments(ctx, runCtx)
	if err != nil {
		return commentState{}, err
	}
	_, state, found := findRunStatusComment(comments, runID)
	if !found {
		return commentState{}, nil
	}
	return state, nil
}

func (s *Service) insertRunCanceledFlowEvent(
	ctx context.Context,
	correlationID string,
	params CancelRunParams,
	result CancelRunResult,
	commentID int64,
) {
	if s.flowEvents == nil || strings.TrimSpace(correlationID) == "" {
		return
	}

	actorType := floweventdomain.ActorTypeSystem
	actorID := floweventdomain.ActorIDControlPlane
	requestedByType := normalizeRequestedByType(params.RequestedByType)
	if requestedByType == RequestedByTypeStaffUser {
		actorType = floweventdomain.ActorTypeHuman
		switch {
		case strings.TrimSpace(params.RequestedByEmail) != "":
			actorID = floweventdomain.ActorID(strings.TrimSpace(params.RequestedByEmail))
		case strings.TrimSpace(params.RequestedByGitHub) != "":
			actorID = floweventdomain.ActorID(strings.TrimSpace(params.RequestedByGitHub))
		case strings.TrimSpace(params.RequestedByID) != "":
			actorID = floweventdomain.ActorID(strings.TrimSpace(params.RequestedByID))
		}
	}

	rawPayload, err := json.Marshal(runCanceledPayload{
		RunID:                        result.RunID,
		PreviousStatus:               result.PreviousStatus,
		CurrentStatus:                result.CurrentStatus,
		AlreadyTerminal:              result.AlreadyTerminal,
		RuntimeDeployCancelRequested: result.RuntimeDeployCancelRequested,
		JobStopped:                   result.JobStopped,
		CanceledGitHubWaits:          result.CanceledGitHubWaits,
		RunStatusCommentID:           commentID,
		RunStatusURL:                 result.CommentURL,
		RequestedByType:              string(requestedByType),
		RequestedByID:                strings.TrimSpace(params.RequestedByID),
		RequestedByEmail:             strings.TrimSpace(params.RequestedByEmail),
		RequestedByGitHub:            strings.TrimSpace(params.RequestedByGitHub),
	})
	if err != nil {
		return
	}

	_ = s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: correlationID,
		ActorType:     actorType,
		ActorID:       actorID,
		EventType:     floweventdomain.EventTypeRunCanceled,
		Payload:       rawPayload,
		CreatedAt:     nowUTC(),
	})
}

func isRunTerminalStatus(status string) bool {
	switch normalizeRunStatus(status) {
	case runStatusSucceeded, runStatusFailed, runStatusCanceled:
		return true
	default:
		return false
	}
}
