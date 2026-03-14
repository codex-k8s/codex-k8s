package worker

import (
	"context"
	"time"
)

type noopMissionControlClient struct{}

func (noopMissionControlClient) ListMissionControlWarmupProjects(context.Context, int) ([]MissionControlWarmupProject, error) {
	return nil, nil
}

func (noopMissionControlClient) RunMissionControlWarmup(context.Context, string, string, string, bool) (MissionControlWarmupResult, error) {
	return MissionControlWarmupResult{}, nil
}

func (noopMissionControlClient) ClaimMissionControlPendingCommands(context.Context, string, time.Duration, int) ([]MissionControlPendingCommand, error) {
	return nil, nil
}

func (noopMissionControlClient) QueueMissionControlCommand(context.Context, MissionControlQueueCommandParams) (MissionControlCommandState, error) {
	return MissionControlCommandState{}, nil
}

func (noopMissionControlClient) MarkMissionControlCommandPendingSync(context.Context, MissionControlPendingSyncParams) (MissionControlCommandState, error) {
	return MissionControlCommandState{}, nil
}

func (noopMissionControlClient) MarkMissionControlCommandReconciled(context.Context, MissionControlReconciledParams) (MissionControlCommandState, error) {
	return MissionControlCommandState{}, nil
}

func (noopMissionControlClient) MarkMissionControlCommandFailed(context.Context, MissionControlFailedParams) (MissionControlCommandState, error) {
	return MissionControlCommandState{}, nil
}

func (noopMissionControlClient) ExecuteNextStepAction(context.Context, NextStepExecuteParams) error {
	return nil
}
