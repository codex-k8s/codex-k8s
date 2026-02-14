package servicescfg

// Config is typed declarative configuration for bootstrap/deploy/runtime flows.
type Config struct {
	Project      string                     `yaml:"project"`
	Templates    TemplatesConfig            `yaml:"templates,omitempty"`
	Bootstrap    BootstrapConfig            `yaml:"bootstrap,omitempty"`
	Deploy       DeployConfig               `yaml:"deploy,omitempty"`
	Runtime      RuntimeConfig              `yaml:"runtime,omitempty"`
	Environments map[string]EnvironmentSpec `yaml:"environments,omitempty"`
}

// TemplatesConfig configures shared template behavior.
type TemplatesConfig struct {
	Partials []string `yaml:"partials,omitempty"`
}

// BootstrapConfig configures install/validate/reconcile bootstrap UX.
type BootstrapConfig struct {
	EnvFilePath          string   `yaml:"env_file_path,omitempty"`
	InstallScriptPath    string   `yaml:"install_script_path,omitempty"`
	RequiredEnvInstall   []string `yaml:"required_env_install,omitempty"`
	RequiredEnvValidate  []string `yaml:"required_env_validate,omitempty"`
	RequiredEnvReconcile []string `yaml:"required_env_reconcile,omitempty"`
	SecretEnvKeys        []string `yaml:"secret_env_keys,omitempty"`
}

// DeployConfig describes declarative deploy behavior.
type DeployConfig struct {
	Environments map[string]DeployEnvironment `yaml:"environments,omitempty"`
}

// DeployEnvironment describes deploy orchestration for a named environment.
type DeployEnvironment struct {
	Namespace            DeployNamespace `yaml:"namespace,omitempty"`
	WaitRolloutEnvVar    string          `yaml:"wait_rollout_env_var,omitempty"`
	RolloutTimeoutEnvVar string          `yaml:"rollout_timeout_env_var,omitempty"`
	ManifestPhases       []DeployPhase   `yaml:"manifest_phases,omitempty"`
}

// DeployNamespace configures namespace resolution for reconcile.
type DeployNamespace struct {
	EnvVar  string `yaml:"env_var,omitempty"`
	Default string `yaml:"default,omitempty"`
	Pattern string `yaml:"pattern,omitempty"`
}

// DeployPhase is one ordered deploy phase.
type DeployPhase struct {
	Name              string         `yaml:"name"`
	EnabledWhenEnv    string         `yaml:"enabled_when_env,omitempty"`
	EnabledWhenEquals string         `yaml:"enabled_when_equals,omitempty"`
	Actions           []DeployAction `yaml:"actions,omitempty"`
}

// DeployAction is a generic action for declarative deploy plan execution.
// Supported types:
// - set_defaults
// - set_from_env
// - import_secret
// - generate_hex
// - assert_env
// - upsert_secret_from_env
// - upsert_configmap_from_env
// - upsert_configmap_from_dir
// - run_script
// - apply_manifests
// - delete_resources
// - rollout_restart
// - wait_targets
type DeployAction struct {
	Type      string            `yaml:"type"`
	Name      string            `yaml:"name,omitempty"`
	Path      string            `yaml:"path,omitempty"`
	Directory string            `yaml:"directory,omitempty"`
	Paths     []string          `yaml:"paths,omitempty"`
	Resources []string          `yaml:"resources,omitempty"`
	Keys      []string          `yaml:"keys,omitempty"`
	Defaults  map[string]string `yaml:"defaults,omitempty"`
	Assign    []EnvAssign       `yaml:"assign,omitempty"`
	Generate  []EnvGenerate     `yaml:"generate,omitempty"`
	Mappings  []EnvDataMapping  `yaml:"mappings,omitempty"`
	Values    map[string]string `yaml:"values,omitempty"`
	WaitFor   []WaitTarget      `yaml:"wait_for,omitempty"`
}

// EnvAssign copies value from Source env key to Target env key when Target is empty.
type EnvAssign struct {
	Target string `yaml:"target"`
	Source string `yaml:"source"`
}

// EnvGenerate describes generation policy for env values.
type EnvGenerate struct {
	Key                     string `yaml:"key"`
	HexBytes                int    `yaml:"hex_bytes,omitempty"`
	IfEmpty                 bool   `yaml:"if_empty,omitempty"`
	RegenerateIfLengthNotIn []int  `yaml:"regenerate_if_length_not_in,omitempty"`
}

// EnvDataMapping maps an env value to specific output key in secret/configmap.
type EnvDataMapping struct {
	Key      string `yaml:"key"`
	Env      string `yaml:"env,omitempty"`
	Required bool   `yaml:"required,omitempty"`
}

// WaitTarget describes one wait rule.
type WaitTarget struct {
	Type       string `yaml:"type"` // rollout|job-complete
	Resource   string `yaml:"resource"`
	Optional   bool   `yaml:"optional,omitempty"`
	TimeoutEnv string `yaml:"timeout_env,omitempty"`
}

// RuntimeConfig configures runtime parity defaults.
type RuntimeConfig struct {
	NonProd RuntimeProfile `yaml:"non_prod,omitempty"`
	Prod    RuntimeProfile `yaml:"prod,omitempty"`
}

// RuntimeProfile describes runtime expectations per environment class.
type RuntimeProfile struct {
	GoHotReload       bool `yaml:"go_hot_reload,omitempty"`
	FrontendHotReload bool `yaml:"frontend_hot_reload,omitempty"`
}

// EnvironmentSpec keeps compatibility with existing environments block.
type EnvironmentSpec struct {
	From               string   `yaml:"from,omitempty"`
	ImagePullPolicy    string   `yaml:"image_pull_policy,omitempty"`
	SlotBootstrapInfra []string `yaml:"slot_bootstrap_infra,omitempty"`
}
