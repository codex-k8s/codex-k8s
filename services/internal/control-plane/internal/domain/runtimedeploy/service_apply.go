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

	"github.com/codex-k8s/codex-k8s/libs/go/manifesttpl"
	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

func (s *Service) applyInfrastructure(ctx context.Context, repositoryRoot string, stack *servicescfg.Stack, namespace string, vars map[string]string, runID string) (map[string]struct{}, error) {
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
		if err := s.applyUnit(ctx, repositoryRoot, name, item.Manifests, namespace, vars, runID); err != nil {
			return nil, err
		}
		applied[name] = struct{}{}
	}
	return applied, nil
}

func (s *Service) applyServices(ctx context.Context, repositoryRoot string, stack *servicescfg.Stack, namespace string, vars map[string]string, applied map[string]struct{}, runID string) error {
	enabledByName := make(map[string]servicescfg.Service, len(stack.Spec.Services))
	groupToNames := make(map[string][]string)
	for _, service := range stack.Spec.Services {
		name := strings.TrimSpace(service.Name)
		if name == "" {
			return fmt.Errorf("service name is required")
		}
		includeScope, skipReason := shouldApplyServiceScope(service.Scope, namespace, vars)
		if !includeScope {
			if strings.TrimSpace(skipReason) != "" {
				s.appendTaskLogBestEffort(ctx, runID, "apply", "info", "Skip unit "+name+": "+skipReason)
			}
			continue
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
				if !dependenciesSatisfied(service.DependsOn, applied, enabledByName) {
					continue
				}
				if err := s.applyUnit(ctx, repositoryRoot, name, service.Manifests, namespace, vars, runID); err != nil {
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

func (s *Service) applyUnit(ctx context.Context, repositoryRoot string, unitName string, manifests []servicescfg.ManifestRef, namespace string, vars map[string]string, runID string) error {
	s.appendTaskLogBestEffort(ctx, runID, "apply", "info", "Apply unit "+unitName+" started")
	repoRoot := strings.TrimSpace(repositoryRoot)
	if repoRoot == "" {
		repoRoot = s.cfg.RepositoryRoot
	}
	for _, manifest := range manifests {
		path := strings.TrimSpace(manifest.Path)
		if path == "" {
			continue
		}
		fullPath := path
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(repoRoot, path)
		}
		raw, err := os.ReadFile(fullPath)
		if err != nil {
			s.appendTaskLogBestEffort(ctx, runID, "apply", "error", "Read manifest failed for "+unitName+": "+fullPath)
			return fmt.Errorf("read manifest %s for %s: %w", fullPath, unitName, err)
		}
		renderedRaw, err := manifesttpl.Render(fullPath, raw, vars)
		if err != nil {
			s.appendTaskLogBestEffort(ctx, runID, "apply", "error", "Render manifest failed for "+unitName+": "+fullPath)
			return fmt.Errorf("render manifest template %s for %s: %w", fullPath, unitName, err)
		}
		rendered := string(renderedRaw)

		refs, err := parseManifestRefs([]byte(rendered), namespace)
		if err != nil {
			s.appendTaskLogBestEffort(ctx, runID, "apply", "error", "Parse manifest refs failed for "+unitName+": "+fullPath)
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
						s.appendTaskLogBestEffort(ctx, runID, "apply", "error", "Delete existing job failed for "+unitName+": "+ref.Name)
						return fmt.Errorf("delete previous job %s/%s before apply: %w", jobNamespace, ref.Name, err)
					}
				}
			}
		}

		appliedRefs, err := s.k8s.ApplyManifest(ctx, []byte(rendered), namespace, s.cfg.KanikoFieldManager)
		if err != nil {
			s.appendTaskLogBestEffort(ctx, runID, "apply", "error", "Apply manifest failed for "+unitName+": "+fullPath)
			return fmt.Errorf("apply manifest %s for %s: %w", fullPath, unitName, err)
		}
		for _, ref := range appliedRefs {
			if err := s.waitAppliedResource(ctx, ref, namespace); err != nil {
				s.appendTaskLogBestEffort(ctx, runID, "apply", "error", "Wait resource failed for "+unitName+": "+ref.Kind+"/"+ref.Name)
				return fmt.Errorf("wait applied resource %s/%s for %s: %w", ref.Kind, ref.Name, unitName, err)
			}
		}
	}
	s.appendTaskLogBestEffort(ctx, runID, "apply", "info", "Apply unit "+unitName+" finished")
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

func shouldApplyServiceScope(scope servicescfg.ServiceScope, targetNamespace string, vars map[string]string) (bool, string) {
	if scope == "" || scope == servicescfg.ServiceScopeEnvironment {
		return true, ""
	}
	if scope != servicescfg.ServiceScopeInfrastructureSingleton {
		return true, ""
	}

	envName := strings.ToLower(strings.TrimSpace(vars["CODEXK8S_ENV"]))
	if envName == "" {
		envName = strings.ToLower(strings.TrimSpace(vars["CODEXK8S_SERVICES_CONFIG_ENV"]))
	}
	if envName != "" && envName != "production" && envName != "prod" {
		return false, "service scope=infrastructure-singleton applies only in production environment"
	}

	platformNamespace := strings.TrimSpace(vars["CODEXK8S_PLATFORM_NAMESPACE"])
	if platformNamespace == "" {
		platformNamespace = strings.TrimSpace(vars["CODEXK8S_PRODUCTION_NAMESPACE"])
	}
	namespace := strings.TrimSpace(targetNamespace)
	if platformNamespace != "" && namespace != "" && namespace != platformNamespace {
		return false, "service scope=infrastructure-singleton applies only in platform namespace " + platformNamespace
	}

	return true, ""
}

func dependenciesSatisfied(dependsOn []string, applied map[string]struct{}, enabledServices map[string]servicescfg.Service) bool {
	for _, dependency := range dependsOn {
		name := strings.TrimSpace(dependency)
		if name == "" {
			continue
		}
		if _, ok := applied[name]; !ok {
			if _, enabled := enabledServices[name]; !enabled {
				continue
			}
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
