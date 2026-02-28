package servicescfg

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

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

	defaultNamespaceTTLRaw := strings.TrimSpace(stack.Spec.WebhookRuntime.DefaultNamespaceTTL)
	if defaultNamespaceTTLRaw == "" {
		defaultNamespaceTTLRaw = "24h"
	}
	defaultNamespaceTTL, err := time.ParseDuration(defaultNamespaceTTLRaw)
	if err != nil {
		return fmt.Errorf("spec.webhookRuntime.defaultNamespaceTTL: parse duration %q: %w", defaultNamespaceTTLRaw, err)
	}
	if defaultNamespaceTTL <= 0 {
		return fmt.Errorf("spec.webhookRuntime.defaultNamespaceTTL must be > 0")
	}
	stack.Spec.WebhookRuntime.DefaultNamespaceTTL = defaultNamespaceTTL.String()

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
	if len(stack.Spec.WebhookRuntime.NamespaceTTLByRole) > 0 {
		normalizedNamespaceTTLByRole := make(map[string]string, len(stack.Spec.WebhookRuntime.NamespaceTTLByRole))
		for rawRole, rawTTL := range stack.Spec.WebhookRuntime.NamespaceTTLByRole {
			roleKey := strings.ToLower(strings.TrimSpace(rawRole))
			if roleKey == "" {
				return fmt.Errorf("spec.webhookRuntime.namespaceTTLByRole contains empty role key")
			}
			if _, exists := normalizedNamespaceTTLByRole[roleKey]; exists {
				return fmt.Errorf("spec.webhookRuntime.namespaceTTLByRole contains duplicate role key %q", roleKey)
			}

			ttlRaw := strings.TrimSpace(rawTTL)
			if ttlRaw == "" {
				return fmt.Errorf("spec.webhookRuntime.namespaceTTLByRole[%q] is empty", rawRole)
			}

			ttl, ttlErr := time.ParseDuration(ttlRaw)
			if ttlErr != nil {
				return fmt.Errorf("spec.webhookRuntime.namespaceTTLByRole[%q]: parse duration %q: %w", rawRole, ttlRaw, ttlErr)
			}
			if ttl <= 0 {
				return fmt.Errorf("spec.webhookRuntime.namespaceTTLByRole[%q] must be > 0", rawRole)
			}
			normalizedNamespaceTTLByRole[roleKey] = ttl.String()
		}
		stack.Spec.WebhookRuntime.NamespaceTTLByRole = normalizedNamespaceTTLByRole
	}

	normalizedVersions, err := normalizeVersions(stack.Spec.Versions)
	if err != nil {
		return err
	}
	stack.Spec.Versions = normalizedVersions

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
	if err := normalizeAndValidateProjectDocs(stack); err != nil {
		return err
	}
	if err := normalizeAndValidateRoleDocTemplates(stack); err != nil {
		return err
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

func normalizeAndValidateProjectDocs(stack *Stack) error {
	if len(stack.Spec.ProjectDocs) == 0 {
		return nil
	}

	seenByRepositoryPath := make(map[string]struct{}, len(stack.Spec.ProjectDocs))
	for i := range stack.Spec.ProjectDocs {
		item := &stack.Spec.ProjectDocs[i]
		repository := normalizeProjectDocRepository(item.Repository)
		if repository != "" && !isValidRepositoryAlias(repository) {
			return fmt.Errorf("spec.projectDocs[%d].repository must match [a-z0-9][a-z0-9._-]*", i)
		}
		item.Repository = repository

		normalizedPath, err := normalizeRepositoryRelativePath(item.Path)
		if err != nil {
			return fmt.Errorf("spec.projectDocs[%d].path %w", i, err)
		}
		lowerPath := strings.ToLower(normalizedPath)
		key := lowerPath
		if repository != "" {
			key = repository + "::" + lowerPath
		}
		if _, exists := seenByRepositoryPath[key]; exists {
			if repository == "" {
				return fmt.Errorf("duplicate spec.projectDocs path %q", normalizedPath)
			}
			return fmt.Errorf("duplicate spec.projectDocs entry %q/%q", repository, normalizedPath)
		}
		seenByRepositoryPath[key] = struct{}{}
		item.Path = normalizedPath

		item.Description = strings.TrimSpace(item.Description)
		if len(item.Roles) > 0 {
			seenRoles := make(map[string]struct{}, len(item.Roles))
			roles := make([]string, 0, len(item.Roles))
			for _, rawRole := range item.Roles {
				role := strings.ToLower(strings.TrimSpace(rawRole))
				if role == "" {
					continue
				}
				if _, exists := seenRoles[role]; exists {
					continue
				}
				seenRoles[role] = struct{}{}
				roles = append(roles, role)
			}
			sort.Strings(roles)
			item.Roles = roles
		}
	}
	return nil
}

func normalizeAndValidateRoleDocTemplates(stack *Stack) error {
	if len(stack.Spec.RoleDocTemplates) == 0 {
		return nil
	}

	normalizedByRole := make(map[string][]RoleDocTemplateRef, len(stack.Spec.RoleDocTemplates))
	for rawRole, rawTemplates := range stack.Spec.RoleDocTemplates {
		roleKey := normalizeRoleKey(rawRole)
		if roleKey == "" {
			return fmt.Errorf("spec.roleDocTemplates contains empty role key")
		}
		if !isValidRoleKey(roleKey) {
			return fmt.Errorf("spec.roleDocTemplates[%q] role key must match [a-z0-9][a-z0-9._-]*", rawRole)
		}
		if _, exists := normalizedByRole[roleKey]; exists {
			return fmt.Errorf("spec.roleDocTemplates contains duplicate role key %q", roleKey)
		}
		if len(rawTemplates) == 0 {
			return fmt.Errorf("spec.roleDocTemplates[%q] must contain at least one template", rawRole)
		}

		seenByRepositoryPath := make(map[string]struct{}, len(rawTemplates))
		templates := make([]RoleDocTemplateRef, 0, len(rawTemplates))
		for idx := range rawTemplates {
			item := rawTemplates[idx]
			repository := normalizeProjectDocRepository(item.Repository)
			if repository != "" && !isValidRepositoryAlias(repository) {
				return fmt.Errorf("spec.roleDocTemplates[%q][%d].repository must match [a-z0-9][a-z0-9._-]*", rawRole, idx)
			}

			normalizedPath, err := normalizeRepositoryRelativePath(item.Path)
			if err != nil {
				return fmt.Errorf("spec.roleDocTemplates[%q][%d].path %w", rawRole, idx, err)
			}
			lowerPath := strings.ToLower(normalizedPath)
			key := lowerPath
			if repository != "" {
				key = repository + "::" + lowerPath
			}
			if _, exists := seenByRepositoryPath[key]; exists {
				if repository == "" {
					return fmt.Errorf("duplicate spec.roleDocTemplates entry for role %q path %q", roleKey, normalizedPath)
				}
				return fmt.Errorf("duplicate spec.roleDocTemplates entry for role %q %q/%q", roleKey, repository, normalizedPath)
			}
			seenByRepositoryPath[key] = struct{}{}

			templates = append(templates, RoleDocTemplateRef{
				Repository:  repository,
				Path:        normalizedPath,
				Description: strings.TrimSpace(item.Description),
			})
		}

		sort.SliceStable(templates, func(i, j int) bool {
			left := templates[i]
			right := templates[j]
			if left.Repository != right.Repository {
				return left.Repository < right.Repository
			}
			return left.Path < right.Path
		})
		normalizedByRole[roleKey] = templates
	}

	stack.Spec.RoleDocTemplates = normalizedByRole
	return nil
}

func normalizeProjectDocRepository(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeRoleKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isValidRoleKey(value string) bool {
	return isValidRepositoryAlias(value)
}

func normalizeRepositoryRelativePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("is required")
	}
	normalizedPath := filepath.ToSlash(filepath.Clean(path))
	if normalizedPath == "." || normalizedPath == ".." || normalizedPath == "/" ||
		strings.HasPrefix(normalizedPath, "/") || strings.HasPrefix(normalizedPath, "../") {
		return "", fmt.Errorf("must be repository-relative: %q", path)
	}
	return normalizedPath, nil
}

func isValidRepositoryAlias(value string) bool {
	if value == "" {
		return false
	}
	for i, r := range value {
		isLetter := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		isAllowedSpecial := r == '.' || r == '_' || r == '-'
		if !(isLetter || isDigit || isAllowedSpecial) {
			return false
		}
		if i == 0 && !isLetter && !isDigit {
			return false
		}
	}
	return true
}
