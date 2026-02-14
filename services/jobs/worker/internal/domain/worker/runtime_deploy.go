package worker

import "context"

// PrepareRunEnvironmentParams describes one runtime environment preparation request.
type PrepareRunEnvironmentParams struct {
	RunID              string
	RuntimeMode        string
	Namespace          string
	TargetEnv          string
	SlotNo             int
	RepositoryFullName string
	ServicesYAMLPath   string
	BuildRef           string
	DeployOnly         bool
}

// PrepareRunEnvironmentResult describes resolved runtime target after preparation.
type PrepareRunEnvironmentResult struct {
	Namespace string
	TargetEnv string
}

// RuntimeEnvironmentPreparer prepares runtime namespace stack before agent job launch.
type RuntimeEnvironmentPreparer interface {
	PrepareRunEnvironment(ctx context.Context, params PrepareRunEnvironmentParams) (PrepareRunEnvironmentResult, error)
}
