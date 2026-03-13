package worker

import "context"

type noopRuntimeEnvironmentPreparer struct{}

func (noopRuntimeEnvironmentPreparer) EvaluateRuntimeReuse(_ context.Context, _ EvaluateRuntimeReuseParams) (EvaluateRuntimeReuseResult, error) {
	return EvaluateRuntimeReuseResult{}, nil
}

func (noopRuntimeEnvironmentPreparer) PrepareRunEnvironment(_ context.Context, _ PrepareRunEnvironmentParams) (PrepareRunEnvironmentResult, error) {
	return PrepareRunEnvironmentResult{}, nil
}
