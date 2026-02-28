package servicescfg

import (
	"fmt"
	"strings"
)

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
