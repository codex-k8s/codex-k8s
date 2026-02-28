package runner

import (
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
)

const (
	maxPromptDocs         = 16
	maxRoleDocTemplates   = 12
	promptDocsDefaultRank = 2
)

type promptDocSortable interface {
	promptDocRepository() string
	promptDocPath() string
}

func (doc promptProjectDocTemplateData) promptDocRepository() string {
	return doc.Repository
}

func (doc promptProjectDocTemplateData) promptDocPath() string {
	return doc.Path
}

func (doc promptRoleDocTemplateData) promptDocRepository() string {
	return doc.Repository
}

func (doc promptRoleDocTemplateData) promptDocPath() string {
	return doc.Path
}

func loadProjectDocsForPrompt(repoDir string, roleKey string, triggerKind string, runtimeMode string) ([]promptProjectDocTemplateData, int, bool) {
	stack, ok := loadPromptServicesStackForPrompt(repoDir, triggerKind, runtimeMode)
	if !ok || len(stack.Spec.ProjectDocs) == 0 {
		return nil, 0, false
	}

	normalizedRole := strings.ToLower(strings.TrimSpace(roleKey))
	items := make([]promptProjectDocTemplateData, 0, len(stack.Spec.ProjectDocs))
	for _, doc := range stack.Spec.ProjectDocs {
		path := strings.TrimSpace(doc.Path)
		if path == "" {
			continue
		}
		roles := normalizeRoles(doc.Roles)
		if len(roles) > 0 && normalizedRole != "" && !slices.Contains(roles, normalizedRole) {
			continue
		}

		items = append(items, promptProjectDocTemplateData{
			Repository:  strings.ToLower(strings.TrimSpace(doc.Repository)),
			Path:        path,
			Description: strings.TrimSpace(doc.Description),
			Optional:    doc.Optional,
		})
	}
	if len(items) == 0 {
		return nil, 0, false
	}

	sortPromptDocItems(items)
	items = dedupePromptDocItems(items)
	return trimPromptDocItems(items, maxPromptDocs)
}

func loadRoleDocTemplatesForPrompt(repoDir string, roleKey string, triggerKind string, runtimeMode string) ([]promptRoleDocTemplateData, int, bool) {
	stack, ok := loadPromptServicesStackForPrompt(repoDir, triggerKind, runtimeMode)
	if !ok || len(stack.Spec.RoleDocTemplates) == 0 {
		return nil, 0, false
	}

	normalizedRole := strings.ToLower(strings.TrimSpace(roleKey))
	if normalizedRole == "" {
		return nil, 0, false
	}

	rawTemplates, exists := stack.Spec.RoleDocTemplates[normalizedRole]
	if !exists || len(rawTemplates) == 0 {
		return nil, 0, false
	}

	items := make([]promptRoleDocTemplateData, 0, len(rawTemplates))
	for _, templateRef := range rawTemplates {
		path := strings.TrimSpace(templateRef.Path)
		if path == "" {
			continue
		}
		templateName := filepath.Base(path)
		if templateName == "" || templateName == "." || templateName == "/" {
			templateName = path
		}

		items = append(items, promptRoleDocTemplateData{
			Repository:   strings.ToLower(strings.TrimSpace(templateRef.Repository)),
			Path:         path,
			TemplateName: templateName,
			Description:  strings.TrimSpace(templateRef.Description),
		})
	}
	if len(items) == 0 {
		return nil, 0, false
	}

	sortPromptDocItems(items)
	items = dedupePromptDocItems(items)
	return trimPromptDocItems(items, maxRoleDocTemplates)
}

func loadPromptServicesStackForPrompt(repoDir string, triggerKind string, runtimeMode string) (*servicescfg.Stack, bool) {
	servicesPath := filepath.Join(strings.TrimSpace(repoDir), "services.yaml")
	raw, err := os.ReadFile(servicesPath)
	if err != nil {
		return nil, false
	}

	env := resolvePromptDocsEnv(triggerKind, runtimeMode)
	loadResult, err := servicescfg.LoadFromYAML(raw, servicescfg.LoadOptions{Env: env})
	if err != nil || loadResult.Stack == nil {
		return nil, false
	}
	return loadResult.Stack, true
}

func resolvePromptDocsEnv(triggerKind string, runtimeMode string) string {
	if strings.EqualFold(strings.TrimSpace(runtimeMode), runtimeModeFullEnv) {
		normalizedTrigger := webhookdomain.NormalizeTriggerKind(triggerKind)
		switch normalizedTrigger {
		case webhookdomain.TriggerKindDev,
			webhookdomain.TriggerKindDevRevise,
			webhookdomain.TriggerKindQA,
			webhookdomain.TriggerKindOps:
			return "ai"
		}
	}
	return "production"
}

func promptDocsRepositoryPriority(repository string) int {
	repository = strings.ToLower(strings.TrimSpace(repository))
	switch {
	case repository == "":
		return promptDocsDefaultRank
	case strings.Contains(repository, "policy"), strings.Contains(repository, "docs"):
		return 0
	case strings.Contains(repository, "orchestrator"):
		return 1
	default:
		return promptDocsDefaultRank
	}
}

func normalizeRoles(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		role := strings.ToLower(strings.TrimSpace(value))
		if role == "" {
			continue
		}
		if _, exists := seen[role]; exists {
			continue
		}
		seen[role] = struct{}{}
		normalized = append(normalized, role)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func sortPromptDocItems[T promptDocSortable](items []T) {
	sort.SliceStable(items, func(i, j int) bool {
		left := items[i]
		right := items[j]
		leftPriority := promptDocsRepositoryPriority(left.promptDocRepository())
		rightPriority := promptDocsRepositoryPriority(right.promptDocRepository())
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		if left.promptDocRepository() != right.promptDocRepository() {
			return left.promptDocRepository() < right.promptDocRepository()
		}
		return left.promptDocPath() < right.promptDocPath()
	})
}

func dedupePromptDocItems[T promptDocSortable](items []T) []T {
	deduped := make([]T, 0, len(items))
	seenByPath := make(map[string]struct{}, len(items))
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.promptDocPath()))
		if key == "" {
			continue
		}
		if _, exists := seenByPath[key]; exists {
			continue
		}
		seenByPath[key] = struct{}{}
		deduped = append(deduped, item)
	}
	if len(deduped) == 0 {
		return nil
	}
	return deduped
}

func trimPromptDocItems[T any](items []T, maxItems int) ([]T, int, bool) {
	total := len(items)
	if total == 0 {
		return nil, 0, false
	}
	if maxItems <= 0 || total <= maxItems {
		return items, total, false
	}
	return items[:maxItems], total, true
}
