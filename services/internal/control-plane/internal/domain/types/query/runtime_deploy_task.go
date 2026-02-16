package query

// RuntimeDeployTaskUpsertDesiredParams describes one desired runtime deployment state update.
type RuntimeDeployTaskUpsertDesiredParams struct {
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

// RuntimeDeployTaskClaimParams describes one runtime deploy lease claim request.
type RuntimeDeployTaskClaimParams struct {
	LeaseOwner string
	LeaseTTL   string
}

// RuntimeDeployTaskMarkSucceededParams describes one successful deployment completion.
type RuntimeDeployTaskMarkSucceededParams struct {
	RunID           string
	LeaseOwner      string
	ResultNamespace string
	ResultTargetEnv string
}

// RuntimeDeployTaskMarkFailedParams describes one failed deployment completion.
type RuntimeDeployTaskMarkFailedParams struct {
	RunID      string
	LeaseOwner string
	LastError  string
}

// RuntimeDeployTaskRenewLeaseParams describes one active lease extension.
type RuntimeDeployTaskRenewLeaseParams struct {
	RunID      string
	LeaseOwner string
	LeaseTTL   string
}

// RuntimeDeployTaskListFilter describes optional task list filters.
type RuntimeDeployTaskListFilter struct {
	Limit     int
	Status    string
	TargetEnv string
}

// RuntimeDeployTaskAppendLogParams appends one log line for a task.
type RuntimeDeployTaskAppendLogParams struct {
	RunID    string
	Stage    string
	Level    string
	Message  string
	MaxLines int
}
