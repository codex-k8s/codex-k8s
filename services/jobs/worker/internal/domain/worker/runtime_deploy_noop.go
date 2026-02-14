package worker

import "context"

type noopRuntimeEnvironmentPreparer struct{}

func (noopRuntimeEnvironmentPreparer) PrepareRunEnvironment(_ context.Context, _ PrepareRunEnvironmentParams) (PrepareRunEnvironmentResult, error) {
	return PrepareRunEnvironmentResult{}, nil
}
