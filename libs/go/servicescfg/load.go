package servicescfg

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type rawHeader struct {
	Metadata Metadata `yaml:"metadata"`
	Spec     struct {
		Project      string                 `yaml:"project"`
		Versions     map[string]VersionSpec `yaml:"versions"`
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
	if err := applyNamespaceTemplate(&stack, opts, &ctx); err != nil {
		return zero, err
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
	if err := applyNamespaceTemplate(&stack, opts, &ctx); err != nil {
		return zero, err
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
