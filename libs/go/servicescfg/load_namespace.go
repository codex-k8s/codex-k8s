package servicescfg

import (
	"fmt"
	"strings"
)

func applyNamespaceTemplate(stack *Stack, opts LoadOptions, ctx *ResolvedContext) error {
	if strings.TrimSpace(opts.Namespace) != "" {
		return nil
	}
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}

	envCfg, err := ResolveEnvironment(stack, ctx.Env)
	if err != nil {
		return err
	}
	if strings.TrimSpace(envCfg.NamespaceTemplate) == "" {
		return nil
	}

	nsRaw, err := renderTemplate("namespace", []byte(envCfg.NamespaceTemplate), *ctx)
	if err != nil {
		return fmt.Errorf("render namespace template: %w", err)
	}
	if ns := strings.TrimSpace(string(nsRaw)); ns != "" {
		ctx.Namespace = ns
	}
	return nil
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
