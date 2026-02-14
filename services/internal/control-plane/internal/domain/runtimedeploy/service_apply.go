package runtimedeploy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

func (s *Service) applyInfrastructure(ctx context.Context, stack *servicescfg.Stack, namespace string, vars map[string]string) (map[string]struct{}, error) {
	enabled := make(map[string]servicescfg.InfrastructureItem, len(stack.Spec.Infrastructure))
	for _, item := range stack.Spec.Infrastructure {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return nil, fmt.Errorf("infrastructure item name is required")
		}
		include, err := evaluateWhen(item.When)
		if err != nil {
			return nil, fmt.Errorf("infrastructure %q when expression: %w", name, err)
		}
		if !include {
			continue
		}
		enabled[name] = item
	}
	order, err := topoSortInfrastructure(enabled)
	if err != nil {
		return nil, err
	}

	applied := make(map[string]struct{}, len(enabled))
	for _, name := range order {
		item := enabled[name]
		if err := s.applyUnit(ctx, name, item.Manifests, namespace, vars); err != nil {
			return nil, err
		}
		applied[name] = struct{}{}
	}
	return applied, nil
}

func (s *Service) applyServices(ctx context.Context, stack *servicescfg.Stack, namespace string, vars map[string]string, applied map[string]struct{}) error {
	enabledByName := make(map[string]servicescfg.Service, len(stack.Spec.Services))
	groupToNames := make(map[string][]string)
	for _, service := range stack.Spec.Services {
		name := strings.TrimSpace(service.Name)
		if name == "" {
			return fmt.Errorf("service name is required")
		}
		include, err := evaluateWhen(service.When)
		if err != nil {
			return fmt.Errorf("service %q when expression: %w", name, err)
		}
		if !include {
			continue
		}
		enabledByName[name] = service
		group := strings.TrimSpace(service.DeployGroup)
		groupToNames[group] = append(groupToNames[group], name)
	}

	if len(enabledByName) == 0 {
		return nil
	}

	groupOrder := buildServiceGroupOrder(stack.Spec.Orchestration.DeployOrder, groupToNames)
	for _, group := range groupOrder {
		names := append([]string(nil), groupToNames[group]...)
		sort.Strings(names)
		for len(names) > 0 {
			progress := false
			for idx := 0; idx < len(names); idx++ {
				name := names[idx]
				service := enabledByName[name]
				if !dependenciesSatisfied(service.DependsOn, applied) {
					continue
				}
				if err := s.applyUnit(ctx, name, service.Manifests, namespace, vars); err != nil {
					return err
				}
				applied[name] = struct{}{}
				names = append(names[:idx], names[idx+1:]...)
				progress = true
				break
			}
			if !progress {
				return fmt.Errorf("service dependency deadlock in group %q: unresolved %s", group, strings.Join(names, ", "))
			}
		}
	}

	return nil
}

func (s *Service) applyUnit(ctx context.Context, unitName string, manifests []servicescfg.ManifestRef, namespace string, vars map[string]string) error {
	for _, manifest := range manifests {
		path := strings.TrimSpace(manifest.Path)
		if path == "" {
			continue
		}
		fullPath := path
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(s.cfg.RepositoryRoot, path)
		}
		raw, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("read manifest %s for %s: %w", fullPath, unitName, err)
		}
		rendered := renderPlaceholders(string(raw), vars)

		refs, err := parseManifestRefs([]byte(rendered), namespace)
		if err != nil {
			return fmt.Errorf("parse manifest refs %s for %s: %w", fullPath, unitName, err)
		}
		for _, ref := range refs {
			if strings.EqualFold(ref.Kind, "Job") && strings.TrimSpace(ref.Name) != "" {
				jobNamespace := strings.TrimSpace(ref.Namespace)
				if jobNamespace == "" {
					jobNamespace = namespace
				}
				if jobNamespace != "" {
					if err := s.k8s.DeleteJobIfExists(ctx, jobNamespace, ref.Name); err != nil {
						return fmt.Errorf("delete previous job %s/%s before apply: %w", jobNamespace, ref.Name, err)
					}
				}
			}
		}

		appliedRefs, err := s.k8s.ApplyManifest(ctx, []byte(rendered), namespace, s.cfg.KanikoFieldManager)
		if err != nil {
			return fmt.Errorf("apply manifest %s for %s: %w", fullPath, unitName, err)
		}
		for _, ref := range appliedRefs {
			if err := s.waitAppliedResource(ctx, ref, namespace); err != nil {
				return fmt.Errorf("wait applied resource %s/%s for %s: %w", ref.Kind, ref.Name, unitName, err)
			}
		}
	}
	return nil
}

func (s *Service) waitAppliedResource(ctx context.Context, ref AppliedResourceRef, fallbackNamespace string) error {
	namespace := strings.TrimSpace(ref.Namespace)
	if namespace == "" {
		namespace = strings.TrimSpace(fallbackNamespace)
	}

	switch strings.ToLower(strings.TrimSpace(ref.Kind)) {
	case "deployment":
		if namespace == "" {
			return nil
		}
		return s.k8s.WaitForDeploymentReady(ctx, namespace, ref.Name, s.cfg.RolloutTimeout)
	case "statefulset":
		if namespace == "" {
			return nil
		}
		return s.k8s.WaitForStatefulSetReady(ctx, namespace, ref.Name, s.cfg.RolloutTimeout)
	case "daemonset":
		if namespace == "" {
			return nil
		}
		return s.k8s.WaitForDaemonSetReady(ctx, namespace, ref.Name, s.cfg.RolloutTimeout)
	case "job":
		if namespace == "" {
			return nil
		}
		return s.k8s.WaitForJobComplete(ctx, namespace, ref.Name, s.cfg.RolloutTimeout)
	default:
		return nil
	}
}

func parseManifestRefs(manifest []byte, namespaceOverride string) ([]AppliedResourceRef, error) {
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 4096)
	overrideNamespace := strings.TrimSpace(namespaceOverride)
	out := make([]AppliedResourceRef, 0, 8)
	for {
		var objectMap map[string]any
		if err := decoder.Decode(&objectMap); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if len(objectMap) == 0 {
			continue
		}
		obj := &unstructured.Unstructured{Object: objectMap}
		name := strings.TrimSpace(obj.GetName())
		if name == "" {
			continue
		}
		namespace := strings.TrimSpace(obj.GetNamespace())
		if namespace == "" {
			namespace = overrideNamespace
		}
		out = append(out, AppliedResourceRef{
			APIVersion: obj.GetAPIVersion(),
			Kind:       obj.GetKind(),
			Namespace:  namespace,
			Name:       name,
		})
	}
	return out, nil
}

func renderPlaceholders(input string, vars map[string]string) string {
	return placeholderPattern.ReplaceAllStringFunc(input, func(token string) string {
		matches := placeholderPattern.FindStringSubmatch(token)
		if len(matches) != 2 {
			return token
		}
		key := matches[1]
		if value, ok := vars[key]; ok {
			return value
		}
		if value, ok := os.LookupEnv(key); ok {
			return value
		}
		return ""
	})
}

func evaluateWhen(value string) (bool, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return true, nil
	}
	parsed, err := strconv.ParseBool(strings.ToLower(trimmed))
	if err != nil {
		return false, err
	}
	return parsed, nil
}

func dependenciesSatisfied(dependsOn []string, applied map[string]struct{}) bool {
	for _, dependency := range dependsOn {
		name := strings.TrimSpace(dependency)
		if name == "" {
			continue
		}
		if _, ok := applied[name]; !ok {
			return false
		}
	}
	return true
}

func topoSortInfrastructure(items map[string]servicescfg.InfrastructureItem) ([]string, error) {
	perm := make(map[string]struct{}, len(items))
	temp := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))

	var visit func(name string) error
	visit = func(name string) error {
		if _, ok := perm[name]; ok {
			return nil
		}
		if _, ok := temp[name]; ok {
			return fmt.Errorf("infrastructure dependency cycle detected at %q", name)
		}
		temp[name] = struct{}{}
		item := items[name]
		for _, dependency := range item.DependsOn {
			depName := strings.TrimSpace(dependency)
			if depName == "" {
				continue
			}
			if _, exists := items[depName]; !exists {
				continue
			}
			if err := visit(depName); err != nil {
				return err
			}
		}
		delete(temp, name)
		perm[name] = struct{}{}
		out = append(out, name)
		return nil
	}

	names := make([]string, 0, len(items))
	for name := range items {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := visit(name); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func buildServiceGroupOrder(deployOrder []string, groupToNames map[string][]string) []string {
	seen := make(map[string]struct{}, len(groupToNames))
	out := make([]string, 0, len(groupToNames))

	for _, group := range deployOrder {
		trimmed := strings.TrimSpace(group)
		if _, ok := groupToNames[trimmed]; !ok {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}

	rest := make([]string, 0, len(groupToNames))
	for group := range groupToNames {
		if _, ok := seen[group]; ok {
			continue
		}
		rest = append(rest, group)
	}
	sort.Strings(rest)
	out = append(out, rest...)
	return out
}
