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

// EvaluateRuntimeReuseParams describes one reusable-namespace fast-path decision request.
type EvaluateRuntimeReuseParams struct {
	RunID              string
	ProjectID          string
	IssueNumber        int64
	AgentKey           string
	RuntimeMode        string
	Namespace          string
	TargetEnv          string
	SlotNo             int
	RepositoryFullName string
	ServicesYAMLPath   string
	BuildRef           string
	DeployOnly         bool
}

// EvaluateRuntimeReuseResult describes whether worker can skip runtime deploy/build.
type EvaluateRuntimeReuseResult struct {
	Reusable          bool
	Namespace         string
	TargetEnv         string
	EffectiveBuildRef string
	FingerprintHash   string
	Reason            string
}

// RuntimeEnvironmentPreparer prepares runtime namespace stack before agent job launch.
type RuntimeEnvironmentPreparer interface {
	EvaluateRuntimeReuse(ctx context.Context, params EvaluateRuntimeReuseParams) (EvaluateRuntimeReuseResult, error)
	PrepareRunEnvironment(ctx context.Context, params PrepareRunEnvironmentParams) (PrepareRunEnvironmentResult, error)
}
