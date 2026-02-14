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
	NamespaceEnvVar       string        `yaml:"namespace_env_var,omitempty"`
	WaitRolloutEnvVar     string        `yaml:"wait_rollout_env_var,omitempty"`
	RolloutTimeoutEnvVar  string        `yaml:"rollout_timeout_env_var,omitempty"`
	ApplyNamespaceEnvVar  string        `yaml:"apply_namespace_env_var,omitempty"`
	NetworkPolicyScript   string        `yaml:"network_policy_script,omitempty"`
	PostgresSecretName    string        `yaml:"postgres_secret_name,omitempty"`
	RuntimeSecretName     string        `yaml:"runtime_secret_name,omitempty"`
	OAuthSecretName       string        `yaml:"oauth_secret_name,omitempty"`
	MigrationsConfigMap   string        `yaml:"migrations_configmap,omitempty"`
	MigrationsDirectory   string        `yaml:"migrations_directory,omitempty"`
	ResourcesConfigMap    string        `yaml:"resources_configmap,omitempty"`
	LabelCatalogConfigMap string        `yaml:"label_catalog_configmap,omitempty"`
	ManifestPhases        []DeployPhase `yaml:"manifest_phases,omitempty"`
}

// DeployPhase is one ordered deploy phase.
type DeployPhase struct {
	Name              string       `yaml:"name"`
	EnabledWhenEnv    string       `yaml:"enabled_when_env,omitempty"`
	EnabledWhenEquals string       `yaml:"enabled_when_equals,omitempty"`
	PreDelete         []string     `yaml:"pre_delete,omitempty"`
	Manifests         []string     `yaml:"manifests,omitempty"`
	RolloutRestart    []string     `yaml:"rollout_restart,omitempty"`
	WaitFor           []WaitTarget `yaml:"wait_for,omitempty"`
}

// WaitTarget describes one wait rule after phase apply/restart.
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
