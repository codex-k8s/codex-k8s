package servicescfg

import (
	"fmt"
	"strings"
)

const (
	// APIVersionV1Alpha1 identifies current typed services.yaml contract version.
	APIVersionV1Alpha1 = "codex-k8s.dev/v1alpha1"
	// KindServiceStack is the root object kind for services config.
	KindServiceStack = "ServiceStack"
)

// RuntimeMode controls runtime execution profile for webhook-triggered runs.
type RuntimeMode string

const (
	RuntimeModeFullEnv  RuntimeMode = "full-env"
	RuntimeModeCodeOnly RuntimeMode = "code-only"
)

// NormalizeRuntimeMode validates and normalizes runtime mode values.
func NormalizeRuntimeMode(value RuntimeMode) (RuntimeMode, error) {
	v := RuntimeMode(strings.TrimSpace(strings.ToLower(string(value))))
	if v == "" {
		return "", nil
	}
	switch v {
	case RuntimeModeFullEnv, RuntimeModeCodeOnly:
		return v, nil
	default:
		return "", fmt.Errorf("unsupported runtime mode %q", value)
	}
}

// CodeUpdateStrategy controls how code updates become effective in non-prod runtime.
type CodeUpdateStrategy string

const (
	CodeUpdateStrategyHotReload CodeUpdateStrategy = "hot-reload"
	CodeUpdateStrategyRebuild   CodeUpdateStrategy = "rebuild"
	CodeUpdateStrategyRestart   CodeUpdateStrategy = "restart"
)

// NormalizeCodeUpdateStrategy validates and normalizes strategy values.
func NormalizeCodeUpdateStrategy(value CodeUpdateStrategy) (CodeUpdateStrategy, error) {
	v := CodeUpdateStrategy(strings.TrimSpace(strings.ToLower(string(value))))
	if v == "" {
		return CodeUpdateStrategyRebuild, nil
	}
	switch v {
	case CodeUpdateStrategyHotReload, CodeUpdateStrategyRebuild, CodeUpdateStrategyRestart:
		return v, nil
	default:
		return "", fmt.Errorf("unsupported codeUpdateStrategy %q", value)
	}
}

// Stack is a typed root contract for services.yaml.
type Stack struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

// Metadata contains high-level stack identity.
type Metadata struct {
	Name string `yaml:"name"`
}

// Spec contains deployable stack definition.
type Spec struct {
	Project        string                 `yaml:"project,omitempty"`
	Imports        []ImportRef            `yaml:"imports,omitempty"`
	Components     []Component            `yaml:"components,omitempty"`
	Environments   map[string]Environment `yaml:"environments,omitempty"`
	WebhookRuntime WebhookRuntime         `yaml:"webhookRuntime,omitempty"`
	Images         map[string]Image       `yaml:"images,omitempty"`
	Infrastructure []InfrastructureItem   `yaml:"infrastructure,omitempty"`
	Services       []Service              `yaml:"services,omitempty"`
	Orchestration  Orchestration          `yaml:"orchestration,omitempty"`
}

// ImportRef points to reusable services.yaml fragment.
type ImportRef struct {
	Path string `yaml:"path"`
}

// Component declares reusable defaults.
type Component struct {
	Name            string           `yaml:"name"`
	ServiceDefaults *ServiceDefaults `yaml:"serviceDefaults,omitempty"`
}

// ServiceDefaults describes reusable service defaults.
type ServiceDefaults struct {
	CodeUpdateStrategy CodeUpdateStrategy `yaml:"codeUpdateStrategy,omitempty"`
	DeployGroup        string             `yaml:"deployGroup,omitempty"`
	DependsOn          []string           `yaml:"dependsOn,omitempty"`
}

// Environment configures environment-level defaults and namespace strategy.
type Environment struct {
	From              string `yaml:"from,omitempty"`
	NamespaceTemplate string `yaml:"namespaceTemplate,omitempty"`
	ImagePullPolicy   string `yaml:"imagePullPolicy,omitempty"`
}

// WebhookRuntime configures trigger->runtime mode mapping for webhook orchestration.
type WebhookRuntime struct {
	DefaultMode  RuntimeMode            `yaml:"defaultMode,omitempty"`
	TriggerModes map[string]RuntimeMode `yaml:"triggerModes,omitempty"`
}

// Image describes a stack image entry.
type Image struct {
	Type        string            `yaml:"type,omitempty"`
	From        string            `yaml:"from,omitempty"`
	Local       string            `yaml:"local,omitempty"`
	Repository  string            `yaml:"repository,omitempty"`
	TagTemplate string            `yaml:"tagTemplate,omitempty"`
	Dockerfile  string            `yaml:"dockerfile,omitempty"`
	Context     string            `yaml:"context,omitempty"`
	BuildArgs   map[string]string `yaml:"buildArgs,omitempty"`
}

// InfrastructureItem groups infra manifests and dependencies.
type InfrastructureItem struct {
	Name      string        `yaml:"name"`
	DependsOn []string      `yaml:"dependsOn,omitempty"`
	Manifests []ManifestRef `yaml:"manifests,omitempty"`
	When      string        `yaml:"when,omitempty"`
}

// ManifestRef points to one YAML manifest.
type ManifestRef struct {
	Path string `yaml:"path"`
}

// Service describes one deployable service.
type Service struct {
	Name               string             `yaml:"name"`
	Use                []string           `yaml:"use,omitempty"`
	CodeUpdateStrategy CodeUpdateStrategy `yaml:"codeUpdateStrategy,omitempty"`
	DeployGroup        string             `yaml:"deployGroup,omitempty"`
	DependsOn          []string           `yaml:"dependsOn,omitempty"`
	Manifests          []ManifestRef      `yaml:"manifests,omitempty"`
	When               string             `yaml:"when,omitempty"`
	Image              ServiceImage       `yaml:"image,omitempty"`
}

// ServiceImage defines how service image reference is built.
type ServiceImage struct {
	Repository  string `yaml:"repository,omitempty"`
	TagTemplate string `yaml:"tagTemplate,omitempty"`
}

// Orchestration defines global rollout and cleanup policy.
type Orchestration struct {
	DeployOrder   []string      `yaml:"deployOrder,omitempty"`
	CleanupPolicy CleanupPolicy `yaml:"cleanupPolicy,omitempty"`
}

// CleanupPolicy defines cleanup behaviors for runtime environments.
type CleanupPolicy struct {
	FullEnvIdleTTL string `yaml:"fullEnvIdleTTL,omitempty"`
}

// ResolvedContext is final context used for template rendering.
type ResolvedContext struct {
	Env       string
	Namespace string
	Project   string
	Slot      int
	Vars      map[string]string
}

// LoadOptions controls services.yaml rendering behavior.
type LoadOptions struct {
	Env       string
	Namespace string
	Slot      int
	Vars      map[string]string
}

// LoadResult returns typed stack plus effective render context.
type LoadResult struct {
	Stack   *Stack
	Context ResolvedContext
	RawYAML []byte
}
