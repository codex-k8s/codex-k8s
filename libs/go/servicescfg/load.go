package servicescfg

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type rawHeader struct {
	Metadata Metadata `yaml:"metadata"`
	Spec     struct {
		Project      string                 `yaml:"project"`
		Versions     map[string]string      `yaml:"versions"`
		Environments map[string]Environment `yaml:"environments"`
	} `yaml:"spec"`
}

// Load parses, renders and validates services.yaml contract.
func Load(path string, opts LoadOptions) (LoadResult, error) {
	var zero LoadResult
	if strings.TrimSpace(path) == "" {
		return zero, fmt.Errorf("config path is required")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return zero, fmt.Errorf("resolve config path: %w", err)
	}

	rootMap, err := loadMergedMap(absPath, nil)
	if err != nil {
		return zero, err
	}
	rawMerged, err := yaml.Marshal(rootMap)
	if err != nil {
		return zero, fmt.Errorf("marshal merged config: %w", err)
	}

	ctx, err := buildContext(rawMerged, opts)
	if err != nil {
		return zero, err
	}
	rendered, err := renderTemplate(absPath, rawMerged, ctx)
	if err != nil {
		return zero, err
	}
	if err := validateRenderedSchema(rendered); err != nil {
		return zero, err
	}

	var stack Stack
	if err := yaml.Unmarshal(rendered, &stack); err != nil {
		return zero, fmt.Errorf("parse rendered services.yaml: %w", err)
	}
	normalizeRootDefaults(&stack, ctx)
	if err := applyServiceComponents(&stack); err != nil {
		return zero, err
	}
	if err := normalizeAndValidate(&stack, ctx.Env); err != nil {
		return zero, err
	}

	if strings.TrimSpace(opts.Namespace) == "" {
		envCfg, err := ResolveEnvironment(&stack, ctx.Env)
		if err != nil {
			return zero, err
		}
		if strings.TrimSpace(envCfg.NamespaceTemplate) != "" {
			nsRaw, err := renderTemplate("namespace", []byte(envCfg.NamespaceTemplate), ctx)
			if err != nil {
				return zero, fmt.Errorf("render namespace template: %w", err)
			}
			if ns := strings.TrimSpace(string(nsRaw)); ns != "" {
				ctx.Namespace = ns
			}
		}
	}

	return LoadResult{
		Stack:   &stack,
		Context: ctx,
		RawYAML: rendered,
	}, nil
}

// LoadFromYAML parses, renders and validates services.yaml contract from in-memory YAML bytes.
//
// This loader does not resolve `spec.imports` because it has no filesystem context.
// It is intended for preflight checks and other read-only operations.
func LoadFromYAML(raw []byte, opts LoadOptions) (LoadResult, error) {
	var zero LoadResult
	if len(raw) == 0 {
		return zero, fmt.Errorf("yaml bytes are empty")
	}

	ctx, err := buildContext(raw, opts)
	if err != nil {
		return zero, err
	}
	rendered, err := renderTemplate("services.yaml", raw, ctx)
	if err != nil {
		return zero, err
	}
	if err := validateRenderedSchema(rendered); err != nil {
		return zero, err
	}

	var stack Stack
	if err := yaml.Unmarshal(rendered, &stack); err != nil {
		return zero, fmt.Errorf("parse rendered services.yaml: %w", err)
	}
	normalizeRootDefaults(&stack, ctx)
	if err := applyServiceComponents(&stack); err != nil {
		return zero, err
	}
	if err := normalizeAndValidate(&stack, ctx.Env); err != nil {
		return zero, err
	}

	if strings.TrimSpace(opts.Namespace) == "" {
		envCfg, err := ResolveEnvironment(&stack, ctx.Env)
		if err != nil {
			return zero, err
		}
		if strings.TrimSpace(envCfg.NamespaceTemplate) != "" {
			nsRaw, err := renderTemplate("namespace", []byte(envCfg.NamespaceTemplate), ctx)
			if err != nil {
				return zero, fmt.Errorf("render namespace template: %w", err)
			}
			if ns := strings.TrimSpace(string(nsRaw)); ns != "" {
				ctx.Namespace = ns
			}
		}
	}

	return LoadResult{
		Stack:   &stack,
		Context: ctx,
		RawYAML: rendered,
	}, nil
}

// Render renders final contract to YAML bytes.
func Render(path string, opts LoadOptions) ([]byte, ResolvedContext, error) {
	result, err := Load(path, opts)
	if err != nil {
		return nil, ResolvedContext{}, err
	}
	out, err := yaml.Marshal(result.Stack)
	if err != nil {
		return nil, ResolvedContext{}, fmt.Errorf("marshal rendered stack: %w", err)
	}
	return out, result.Context, nil
}

// ResolveEnvironment returns final env config with inheritance resolved.
func ResolveEnvironment(stack *Stack, envName string) (Environment, error) {
	if stack == nil {
		return Environment{}, fmt.Errorf("stack is nil")
	}
	envName = strings.TrimSpace(envName)
	if envName == "" {
		return Environment{}, fmt.Errorf("environment is required")
	}
	if len(stack.Spec.Environments) == 0 {
		return Environment{}, fmt.Errorf("environments section is empty")
	}

	visited := make(map[string]struct{})
	var resolve func(name string) (Environment, error)
	resolve = func(name string) (Environment, error) {
		if _, ok := visited[name]; ok {
			return Environment{}, fmt.Errorf("environment inheritance cycle detected at %q", name)
		}
		visited[name] = struct{}{}

		current, ok := stack.Spec.Environments[name]
		if !ok {
			return Environment{}, fmt.Errorf("environment %q not defined", name)
		}
		parent := strings.TrimSpace(current.From)
		if parent == "" {
			return current, nil
		}

		base, err := resolve(parent)
		if err != nil {
			return Environment{}, err
		}
		merged := base
		if strings.TrimSpace(current.NamespaceTemplate) != "" {
			merged.NamespaceTemplate = current.NamespaceTemplate
		}
		if strings.TrimSpace(current.DomainTemplate) != "" {
			merged.DomainTemplate = current.DomainTemplate
		}
		if strings.TrimSpace(current.ImagePullPolicy) != "" {
			merged.ImagePullPolicy = current.ImagePullPolicy
		}
		merged.From = current.From
		return merged, nil
	}

	return resolve(envName)
}

func normalizeRootDefaults(stack *Stack, ctx ResolvedContext) {
	if stack.APIVersion == "" {
		stack.APIVersion = APIVersionV1Alpha1
	}
	if stack.Kind == "" {
		stack.Kind = KindServiceStack
	}
	if strings.TrimSpace(stack.Metadata.Name) == "" {
		stack.Metadata.Name = ctx.Project
	}
	if strings.TrimSpace(stack.Spec.Project) == "" {
		stack.Spec.Project = ctx.Project
	}
}

func normalizeAndValidate(stack *Stack, env string) error {
	if stack == nil {
		return fmt.Errorf("stack is nil")
	}
	if strings.TrimSpace(stack.APIVersion) == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if strings.TrimSpace(stack.Kind) == "" {
		return fmt.Errorf("kind is required")
	}
	if strings.TrimSpace(stack.Metadata.Name) == "" && strings.TrimSpace(stack.Spec.Project) == "" {
		return fmt.Errorf("metadata.name or spec.project is required")
	}
	if strings.TrimSpace(env) == "" {
		return fmt.Errorf("environment is required")
	}
	if _, ok := stack.Spec.Environments[env]; !ok {
		return fmt.Errorf("environment %q not found in spec.environments", env)
	}

	defaultRuntimeMode, err := NormalizeRuntimeMode(stack.Spec.WebhookRuntime.DefaultMode)
	if err != nil {
		return fmt.Errorf("spec.webhookRuntime.defaultMode: %w", err)
	}
	if defaultRuntimeMode == "" {
		defaultRuntimeMode = RuntimeModeFullEnv
	}
	stack.Spec.WebhookRuntime.DefaultMode = defaultRuntimeMode

	if len(stack.Spec.WebhookRuntime.TriggerModes) > 0 {
		normalizedTriggerModes := make(map[string]RuntimeMode, len(stack.Spec.WebhookRuntime.TriggerModes))
		for rawTrigger, rawMode := range stack.Spec.WebhookRuntime.TriggerModes {
			triggerKey := normalizeTriggerModeKey(rawTrigger)
			if triggerKey == "" {
				return fmt.Errorf("spec.webhookRuntime.triggerModes contains empty trigger key")
			}
			mode, modeErr := NormalizeRuntimeMode(rawMode)
			if modeErr != nil {
				return fmt.Errorf("spec.webhookRuntime.triggerModes[%q]: %w", rawTrigger, modeErr)
			}
			if mode == "" {
				mode = defaultRuntimeMode
			}
			normalizedTriggerModes[triggerKey] = mode
		}
		stack.Spec.WebhookRuntime.TriggerModes = normalizedTriggerModes
	}
	if err := validateSecretResolution(stack.Spec.SecretResolution); err != nil {
		return err
	}

	seenServices := make(map[string]struct{})
	for i := range stack.Spec.Services {
		svc := &stack.Spec.Services[i]
		name := strings.TrimSpace(svc.Name)
		if name == "" {
			return fmt.Errorf("service[%d].name is required", i)
		}
		if _, ok := seenServices[name]; ok {
			return fmt.Errorf("duplicate service name %q", name)
		}
		seenServices[name] = struct{}{}

		strategy, err := NormalizeCodeUpdateStrategy(svc.CodeUpdateStrategy)
		if err != nil {
			return fmt.Errorf("service %q: %w", name, err)
		}
		svc.CodeUpdateStrategy = strategy
		scope, err := NormalizeServiceScope(svc.Scope)
		if err != nil {
			return fmt.Errorf("service %q: %w", name, err)
		}
		svc.Scope = scope
	}

	projectName := strings.TrimSpace(stack.Spec.Project)
	if projectName == "" {
		projectName = strings.TrimSpace(stack.Metadata.Name)
	}
	if projectName == "codex-k8s" {
		production, ok := stack.Spec.Environments["production"]
		if !ok {
			return fmt.Errorf("codex-k8s requires production environment")
		}
		namespaceTemplate := strings.TrimSpace(production.NamespaceTemplate)
		expectedTemplate := "{{ .Project }}-prod"
		expectedResolved := fmt.Sprintf("%s-prod", projectName)
		if namespaceTemplate != expectedTemplate && namespaceTemplate != expectedResolved {
			return fmt.Errorf("codex-k8s requires production namespace template {{ .Project }}-prod")
		}
	}

	return nil
}

func applyServiceComponents(stack *Stack) error {
	if stack == nil {
		return fmt.Errorf("stack is nil")
	}
	if len(stack.Spec.Services) == 0 {
		return nil
	}

	components := make(map[string]Component, len(stack.Spec.Components))
	for _, component := range stack.Spec.Components {
		name := strings.TrimSpace(component.Name)
		if name == "" {
			return fmt.Errorf("component name is required")
		}
		if _, exists := components[name]; exists {
			return fmt.Errorf("duplicate component %q", name)
		}
		components[name] = component
	}

	for idx := range stack.Spec.Services {
		svc := &stack.Spec.Services[idx]
		for _, ref := range svc.Use {
			componentName := strings.TrimSpace(ref)
			component, ok := components[componentName]
			if !ok {
				return fmt.Errorf("service %q references unknown component %q", svc.Name, componentName)
			}
			if component.ServiceDefaults == nil {
				continue
			}
			if svc.CodeUpdateStrategy == "" && component.ServiceDefaults.CodeUpdateStrategy != "" {
				svc.CodeUpdateStrategy = component.ServiceDefaults.CodeUpdateStrategy
			}
			if strings.TrimSpace(svc.DeployGroup) == "" && strings.TrimSpace(component.ServiceDefaults.DeployGroup) != "" {
				svc.DeployGroup = component.ServiceDefaults.DeployGroup
			}
			if len(svc.DependsOn) == 0 && len(component.ServiceDefaults.DependsOn) > 0 {
				svc.DependsOn = append([]string(nil), component.ServiceDefaults.DependsOn...)
			}
		}
	}

	return nil
}

func buildContext(raw []byte, opts LoadOptions) (ResolvedContext, error) {
	var header rawHeader
	if err := yaml.Unmarshal(raw, &header); err != nil {
		return ResolvedContext{}, fmt.Errorf("parse header: %w", err)
	}

	project := strings.TrimSpace(header.Spec.Project)
	if project == "" {
		project = strings.TrimSpace(header.Metadata.Name)
	}
	if project == "" {
		return ResolvedContext{}, fmt.Errorf("project is required (spec.project or metadata.name)")
	}

	ctx := ResolvedContext{
		Env:       strings.TrimSpace(opts.Env),
		Namespace: strings.TrimSpace(opts.Namespace),
		Project:   project,
		Slot:      opts.Slot,
		Vars:      cloneStringMap(opts.Vars),
		Versions:  cloneStringMap(header.Spec.Versions),
	}
	if ctx.Env == "" {
		ctx.Env = "production"
	}
	if ctx.Vars == nil {
		ctx.Vars = make(map[string]string)
	}
	if ctx.Versions == nil {
		ctx.Versions = make(map[string]string)
	}

	if ctx.Namespace != "" {
		return ctx, nil
	}

	envCfg, err := resolveEnvironmentFromMap(header.Spec.Environments, ctx.Env)
	if err != nil {
		return ResolvedContext{}, err
	}
	if strings.TrimSpace(envCfg.NamespaceTemplate) != "" {
		nsRaw, err := renderTemplate("namespace", []byte(envCfg.NamespaceTemplate), ctx)
		if err != nil {
			return ResolvedContext{}, fmt.Errorf("render namespace template: %w", err)
		}
		if ns := strings.TrimSpace(string(nsRaw)); ns != "" {
			ctx.Namespace = ns
		}
	}
	if ctx.Namespace == "" {
		switch ctx.Env {
		case "ai":
			if ctx.Slot > 0 {
				ctx.Namespace = fmt.Sprintf("%s-dev-%d", ctx.Project, ctx.Slot)
			}
		case "ai-repair":
			ctx.Namespace = fmt.Sprintf("%s-production", ctx.Project)
		default:
			ctx.Namespace = fmt.Sprintf("%s-%s", ctx.Project, ctx.Env)
		}
	}

	return ctx, nil
}

func resolveEnvironmentFromMap(environments map[string]Environment, envName string) (Environment, error) {
	if len(environments) == 0 {
		return Environment{}, fmt.Errorf("environments section is empty")
	}
	if strings.TrimSpace(envName) == "" {
		return Environment{}, fmt.Errorf("environment is required")
	}

	visited := make(map[string]struct{})
	var resolve func(name string) (Environment, error)
	resolve = func(name string) (Environment, error) {
		if _, seen := visited[name]; seen {
			return Environment{}, fmt.Errorf("environment inheritance cycle detected at %q", name)
		}
		visited[name] = struct{}{}

		current, ok := environments[name]
		if !ok {
			return Environment{}, fmt.Errorf("environment %q not found", name)
		}
		parent := strings.TrimSpace(current.From)
		if parent == "" {
			return current, nil
		}

		base, err := resolve(parent)
		if err != nil {
			return Environment{}, err
		}
		merged := base
		if strings.TrimSpace(current.NamespaceTemplate) != "" {
			merged.NamespaceTemplate = current.NamespaceTemplate
		}
		if strings.TrimSpace(current.DomainTemplate) != "" {
			merged.DomainTemplate = current.DomainTemplate
		}
		if strings.TrimSpace(current.ImagePullPolicy) != "" {
			merged.ImagePullPolicy = current.ImagePullPolicy
		}
		merged.From = current.From
		return merged, nil
	}

	return resolve(envName)
}

func loadMergedMap(path string, trail []string) (map[string]any, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve import path %q: %w", path, err)
	}
	for _, entry := range trail {
		if entry == absPath {
			chain := append(append([]string(nil), trail...), absPath)
			return nil, fmt.Errorf("imports cycle detected: %s", strings.Join(chain, " -> "))
		}
	}

	raw, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read services file %q: %w", absPath, err)
	}
	decoded, err := decodeYAMLToMap(raw)
	if err != nil {
		return nil, fmt.Errorf("decode services file %q: %w", absPath, err)
	}
	importPaths, err := extractImportPaths(decoded)
	if err != nil {
		return nil, fmt.Errorf("parse imports in %q: %w", absPath, err)
	}
	clearImports(decoded)

	merged := make(map[string]any)
	baseDir := filepath.Dir(absPath)
	for _, importPath := range importPaths {
		expanded, err := expandImportPaths(baseDir, importPath)
		if err != nil {
			return nil, fmt.Errorf("resolve import %q from %q: %w", importPath, absPath, err)
		}
		for _, candidate := range expanded {
			child, err := loadMergedMap(candidate, append(trail, absPath))
			if err != nil {
				return nil, err
			}
			merged = deepMergeMaps(merged, child)
		}
	}

	merged = deepMergeMaps(merged, decoded)
	return merged, nil
}

func expandImportPaths(baseDir string, pattern string) ([]string, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nil, fmt.Errorf("empty import path")
	}

	resolved := pattern
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(baseDir, pattern)
	}
	matches, err := filepath.Glob(resolved)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no files matched import %q", pattern)
	}

	sort.Strings(matches)
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		stat, err := os.Stat(match)
		if err != nil {
			return nil, err
		}
		if stat.IsDir() {
			continue
		}
		out = append(out, match)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("import %q resolved only directories", pattern)
	}
	return out, nil
}

func decodeYAMLToMap(raw []byte) (map[string]any, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	merged := make(map[string]any)

	for {
		var doc map[string]any
		if err := decoder.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if len(doc) == 0 {
			continue
		}
		merged = deepMergeMaps(merged, doc)
	}
	return merged, nil
}

func extractImportPaths(doc map[string]any) ([]string, error) {
	spec, ok := doc["spec"].(map[string]any)
	if !ok {
		return nil, nil
	}
	rawImports, ok := spec["imports"]
	if !ok {
		return nil, nil
	}
	items, ok := rawImports.([]any)
	if !ok {
		return nil, fmt.Errorf("spec.imports must be a list")
	}

	var paths []string
	for _, item := range items {
		switch typed := item.(type) {
		case string:
			if strings.TrimSpace(typed) == "" {
				return nil, fmt.Errorf("spec.imports contains empty path")
			}
			paths = append(paths, typed)
		case map[string]any:
			path, _ := typed["path"].(string)
			path = strings.TrimSpace(path)
			if path == "" {
				return nil, fmt.Errorf("spec.imports item must define path")
			}
			paths = append(paths, path)
		default:
			return nil, fmt.Errorf("spec.imports item has unsupported type %T", item)
		}
	}
	return paths, nil
}

func clearImports(doc map[string]any) {
	spec, ok := doc["spec"].(map[string]any)
	if !ok {
		return
	}
	delete(spec, "imports")
	if len(spec) == 0 {
		delete(doc, "spec")
	}
}

func deepMergeMaps(base map[string]any, overlay map[string]any) map[string]any {
	out := cloneMapAny(base)
	for key, value := range overlay {
		existing, ok := out[key]
		if !ok {
			out[key] = cloneAny(value)
			continue
		}
		out[key] = deepMergeValue(existing, value)
	}
	return out
}

func deepMergeValue(base any, overlay any) any {
	baseMap, baseIsMap := base.(map[string]any)
	overlayMap, overlayIsMap := overlay.(map[string]any)
	if baseIsMap && overlayIsMap {
		return deepMergeMaps(baseMap, overlayMap)
	}
	return cloneAny(overlay)
}

func cloneAny(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneMapAny(typed)
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, cloneAny(item))
		}
		return out
	default:
		return typed
	}
}

func cloneMapAny(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = cloneAny(value)
	}
	return out
}

func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func renderTemplate(name string, raw []byte, ctx ResolvedContext) ([]byte, error) {
	funcs := template.FuncMap{
		"default": func(value string, def string) string {
			if strings.TrimSpace(value) == "" {
				return def
			}
			return value
		},
		"envOr": func(key string, def string) string {
			if value, ok := ctx.Vars[key]; ok && value != "" {
				return value
			}
			if value := os.Getenv(key); value != "" {
				return value
			}
			return def
		},
		"ternary": func(cond bool, left any, right any) any {
			if cond {
				return left
			}
			return right
		},
		"join":       strings.Join,
		"trimPrefix": strings.TrimPrefix,
		"toLower":    strings.ToLower,
	}

	tmpl, err := template.New(name).Funcs(funcs).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse template %q: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return nil, fmt.Errorf("execute template %q: %w", name, err)
	}
	return buf.Bytes(), nil
}
