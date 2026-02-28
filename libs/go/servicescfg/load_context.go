package servicescfg

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

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
		Versions:  versionValues(header.Spec.Versions),
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
