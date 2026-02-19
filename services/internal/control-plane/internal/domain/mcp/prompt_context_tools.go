package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func (s *Service) PromptContext(ctx context.Context, session SessionContext) (PromptContextResult, error) {
	tool, err := s.toolCapability(ToolPromptContextGet)
	if err != nil {
		return PromptContextResult{}, err
	}

	runCtx, err := s.resolveRunContext(ctx, session, false)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return PromptContextResult{}, err
	}
	allowedTools := s.allowedToolsForRunContext(runCtx)

	s.auditToolCalled(ctx, runCtx.Session, tool)

	issueCtx := buildPromptIssueContext(runCtx.Payload.Issue)
	roleCtx := buildPromptRoleContext(runCtx)
	runtimeCtx, docs := s.buildPromptRuntimeContext(runCtx, roleCtx.AgentKey)
	result := PromptContextResult{
		Status: ToolExecutionStatusOK,
		Context: PromptContext{
			Version: promptContextVersion,
			Run: PromptRunContext{
				RunID:         runCtx.Session.RunID,
				CorrelationID: runCtx.Session.CorrelationID,
				ProjectID:     runCtx.Session.ProjectID,
				Namespace:     runCtx.Session.Namespace,
				RuntimeMode:   runCtx.Session.RuntimeMode,
			},
			Repository: PromptRepositoryContext{
				Provider:     runCtx.Repository.Provider,
				Owner:        runCtx.Repository.Owner,
				Name:         runCtx.Repository.Name,
				FullName:     runCtx.Repository.Owner + "/" + runCtx.Repository.Name,
				ServicesYAML: runCtx.Repository.ServicesYAMLPath,
			},
			Issue: issueCtx,
			Role:  roleCtx,
			Docs:  docs,
			Environment: PromptEnvironmentContext{
				ServiceName: s.cfg.ServerName,
				MCPBaseURL:  s.cfg.InternalMCPBaseURL,
			},
			Runtime:  runtimeCtx,
			Services: buildPromptServices(s.cfg.PublicBaseURL, s.cfg.InternalMCPBaseURL),
			MCP: PromptMCPContext{
				ServerName: s.cfg.ServerName,
				Tools:      allowedTools,
			},
		},
	}

	s.auditPromptContextAssembled(ctx, runCtx)
	s.auditToolSucceeded(ctx, runCtx.Session, tool)
	return result, nil
}

func buildPromptIssueContext(issue *querytypes.RunPayloadIssue) *PromptIssueContext {
	if issue == nil {
		return nil
	}
	if issue.Number <= 0 {
		return nil
	}
	return &PromptIssueContext{
		Number: issue.Number,
		Title:  strings.TrimSpace(issue.Title),
		State:  strings.TrimSpace(issue.State),
		URL:    strings.TrimSpace(issue.HTMLURL),
	}
}

func buildPromptServices(publicBaseURL string, internalMCPBaseURL string) []PromptServiceContext {
	items := []PromptServiceContext{
		{Name: "control-plane-grpc", Endpoint: "codex-k8s-control-plane:9090", Kind: "grpc"},
		{Name: "control-plane-mcp", Endpoint: strings.TrimSpace(internalMCPBaseURL), Kind: "mcp-http"},
		{Name: "worker", Endpoint: "codex-k8s-worker", Kind: "job-orchestrator"},
	}
	if strings.TrimSpace(publicBaseURL) != "" {
		items = append(items, PromptServiceContext{
			Name:     "api-gateway",
			Endpoint: strings.TrimSpace(publicBaseURL),
			Kind:     "http",
		})
	}
	return items
}

func (s *Service) buildPromptRuntimeContext(runCtx resolvedRunContext, roleKey string) (PromptRuntimeContext, []PromptProjectDocRef) {
	targetEnv := resolvePromptTargetEnv(runCtx, s.cfg.ServicesConfigEnv)
	servicesPath := strings.TrimSpace(runCtx.Repository.ServicesYAMLPath)
	if servicesPath == "" {
		servicesPath = "services.yaml"
	}

	result := PromptRuntimeContext{
		TargetEnv:    targetEnv,
		ServicesYAML: servicesPath,
		Hints: PromptRuntimeHints{
			RuntimeMode:    runCtx.Session.RuntimeMode,
			RepositoryRoot: strings.TrimSpace(s.cfg.RepositoryRoot),
		},
	}

	configPath, err := s.resolvePromptServicesConfigPath(runCtx, servicesPath)
	if err != nil {
		result.Hints.InventoryError = err.Error()
		return result, nil
	}

	loadResult, err := servicescfg.Load(configPath, servicescfg.LoadOptions{
		Env:       targetEnv,
		Namespace: strings.TrimSpace(runCtx.Session.Namespace),
	})
	if err != nil {
		result.InventorySource = configPath
		result.Hints.InventoryError = err.Error()
		return result, nil
	}

	result.InventorySource = configPath
	result.Inventory = buildPromptRuntimeInventory(loadResult.Stack)
	result.Hints.InventoryLoaded = true
	return result, buildPromptProjectDocs(loadResult.Stack, roleKey)
}

func resolvePromptTargetEnv(runCtx resolvedRunContext, fallback string) string {
	fallback = strings.TrimSpace(fallback)
	if fallback == "" {
		fallback = "production"
	}

	triggerKind := ""
	if runCtx.Payload.Trigger != nil {
		triggerKind = strings.ToLower(strings.TrimSpace(runCtx.Payload.Trigger.Kind))
	}
	switch triggerKind {
	case "dev", "debug", "qa":
		return "ai"
	}

	if runCtx.Session.RuntimeMode == agentdomain.RuntimeModeFullEnv && strings.TrimSpace(runCtx.Session.Namespace) != "" {
		ns := strings.ToLower(strings.TrimSpace(runCtx.Session.Namespace))
		if strings.Contains(ns, "-dev-") || strings.HasSuffix(ns, "-dev") {
			return "ai"
		}
	}

	return fallback
}

func buildPromptRuntimeInventory(stack *servicescfg.Stack) []PromptRuntimeServiceContext {
	if stack == nil {
		return nil
	}
	services := make([]PromptRuntimeServiceContext, 0, len(stack.Spec.Services))
	for _, service := range stack.Spec.Services {
		name := strings.TrimSpace(service.Name)
		if name == "" {
			continue
		}

		strategy, err := servicescfg.NormalizeCodeUpdateStrategy(service.CodeUpdateStrategy)
		if err != nil || strategy == "" {
			strategy = servicescfg.CodeUpdateStrategyRebuild
		}

		services = append(services, PromptRuntimeServiceContext{
			Name:               name,
			DeployGroup:        strings.TrimSpace(service.DeployGroup),
			CodeUpdateStrategy: string(strategy),
			DependsOn:          trimAndFilter(service.DependsOn),
			ManifestPaths:      manifestPaths(service.Manifests),
		})
	}
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})
	return services
}

func manifestPaths(items []servicescfg.ManifestRef) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		path := strings.TrimSpace(item.Path)
		if path == "" {
			continue
		}
		out = append(out, path)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func buildPromptProjectDocs(stack *servicescfg.Stack, roleKey string) []PromptProjectDocRef {
	if stack == nil || len(stack.Spec.ProjectDocs) == 0 {
		return nil
	}

	normalizedRole := strings.ToLower(strings.TrimSpace(roleKey))
	out := make([]PromptProjectDocRef, 0, len(stack.Spec.ProjectDocs))
	for _, item := range stack.Spec.ProjectDocs {
		path := strings.TrimSpace(item.Path)
		if path == "" {
			continue
		}
		roles := trimAndFilter(item.Roles)
		if len(roles) > 0 && normalizedRole != "" && !slices.Contains(roles, normalizedRole) {
			continue
		}
		out = append(out, PromptProjectDocRef{
			Path:        path,
			Description: strings.TrimSpace(item.Description),
			Roles:       roles,
			Optional:    item.Optional,
		})
	}
	if len(out) == 0 {
		return nil
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	return out
}

func buildPromptRoleContext(runCtx resolvedRunContext) PromptRoleContext {
	key := ""
	if runCtx.Payload.Agent != nil {
		key = strings.ToLower(strings.TrimSpace(runCtx.Payload.Agent.Key))
	}
	if key == "" {
		key = "dev"
	}

	capabilitiesByRole := map[string]PromptRoleContext{
		"dev": {
			AgentKey:    "dev",
			DisplayName: "Developer",
			Capabilities: []PromptRoleCapability{
				{Area: "code", Scope: "implement"},
				{Area: "tests", Scope: "run-and-fix"},
				{Area: "docs", Scope: "update-when-behavior-changes"},
			},
		},
		"pm": {
			AgentKey:    "pm",
			DisplayName: "Product Manager",
			Capabilities: []PromptRoleCapability{
				{Area: "requirements", Scope: "refine-and-structure"},
				{Area: "delivery", Scope: "prioritize-and-scope"},
				{Area: "traceability", Scope: "keep-links-consistent"},
			},
		},
		"sa": {
			AgentKey:    "sa",
			DisplayName: "Solution Architect",
			Capabilities: []PromptRoleCapability{
				{Area: "architecture", Scope: "define-boundaries"},
				{Area: "contracts", Scope: "enforce-typed-interfaces"},
				{Area: "risks", Scope: "analyze-tradeoffs"},
			},
		},
		"em": {
			AgentKey:    "em",
			DisplayName: "Engineering Manager",
			Capabilities: []PromptRoleCapability{
				{Area: "planning", Scope: "decompose-and-sequence"},
				{Area: "quality-gates", Scope: "enforce-checklists"},
				{Area: "delivery", Scope: "coordinate-handovers"},
			},
		},
		"reviewer": {
			AgentKey:    "reviewer",
			DisplayName: "Reviewer",
			Capabilities: []PromptRoleCapability{
				{Area: "code-review", Scope: "find-bugs-risks-regressions"},
				{Area: "tests", Scope: "identify-coverage-gaps"},
				{Area: "docs", Scope: "verify-consistency"},
			},
		},
		"qa": {
			AgentKey:    "qa",
			DisplayName: "QA",
			Capabilities: []PromptRoleCapability{
				{Area: "test-design", Scope: "derive-cases-and-edges"},
				{Area: "verification", Scope: "reproduce-and-validate"},
				{Area: "regression", Scope: "protect-critical-flows"},
			},
		},
		"sre": {
			AgentKey:    "sre",
			DisplayName: "SRE",
			Capabilities: []PromptRoleCapability{
				{Area: "operations", Scope: "diagnose-runtime-issues"},
				{Area: "kubernetes", Scope: "inspect-and-remediate"},
				{Area: "reliability", Scope: "stabilize-deploy-flow"},
			},
		},
		"km": {
			AgentKey:    "km",
			DisplayName: "Knowledge Manager",
			Capabilities: []PromptRoleCapability{
				{Area: "docset", Scope: "maintain-quality"},
				{Area: "prompts", Scope: "improve-role-templates"},
				{Area: "self-improve", Scope: "collect-evidence-and-propose-fixes"},
			},
		},
	}

	if roleCtx, ok := capabilitiesByRole[key]; ok {
		return roleCtx
	}
	return PromptRoleContext{
		AgentKey:    key,
		DisplayName: strings.ToUpper(key),
		Capabilities: []PromptRoleCapability{
			{Area: "generic", Scope: "follow-task-contract"},
		},
	}
}

func trimAndFilter(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (s *Service) resolvePromptServicesConfigPath(runCtx resolvedRunContext, servicesPath string) (string, error) {
	servicesPath = strings.TrimSpace(servicesPath)
	if servicesPath == "" {
		servicesPath = "services.yaml"
	}

	candidates := make([]string, 0, 8)
	if filepath.IsAbs(servicesPath) {
		candidates = append(candidates, servicesPath)
	}

	repositoryRoot := strings.TrimSpace(s.cfg.RepositoryRoot)
	if repositoryRoot != "" {
		candidates = append(candidates, filepath.Join(repositoryRoot, servicesPath))
		if owner, name := strings.TrimSpace(runCtx.Repository.Owner), strings.TrimSpace(runCtx.Repository.Name); owner != "" && name != "" {
			cacheBase := filepath.Join(repositoryRoot, "github", owner, name)
			for _, ref := range promptRepoRefCandidates(runCtx) {
				candidates = append(candidates, filepath.Join(cacheBase, sanitizePromptRepoRef(ref), servicesPath))
			}
			if entries, err := os.ReadDir(cacheBase); err == nil {
				names := make([]string, 0, len(entries))
				for _, entry := range entries {
					if !entry.IsDir() {
						continue
					}
					names = append(names, entry.Name())
				}
				sort.Strings(names)
				for i := len(names) - 1; i >= 0; i-- {
					candidates = append(candidates, filepath.Join(cacheBase, names[i], servicesPath))
				}
			}
		}
	}

	checked := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		clean := filepath.Clean(strings.TrimSpace(candidate))
		if clean == "." || clean == "" {
			continue
		}
		if _, exists := seen[clean]; exists {
			continue
		}
		seen[clean] = struct{}{}
		checked = append(checked, clean)
		if stat, err := os.Stat(clean); err == nil && !stat.IsDir() {
			return clean, nil
		}
	}

	if len(checked) == 0 {
		return "", fmt.Errorf("services.yaml path candidates are empty")
	}
	return "", fmt.Errorf("services.yaml not found, checked: %s", strings.Join(checked, ", "))
}

func promptRepoRefCandidates(runCtx resolvedRunContext) []string {
	candidates := []string{"main"}
	if runCtx.Payload.PullRequest != nil {
		if head := strings.TrimSpace(runCtx.Payload.PullRequest.HeadRef); head != "" {
			candidates = append([]string{head}, candidates...)
		}
		if base := strings.TrimSpace(runCtx.Payload.PullRequest.BaseRef); base != "" {
			candidates = append(candidates, base)
		}
	}
	if runCtx.Payload.Trigger != nil {
		if trigger := strings.TrimSpace(runCtx.Payload.Trigger.Kind); trigger != "" {
			candidates = append(candidates, trigger)
		}
	}

	seen := make(map[string]struct{}, len(candidates))
	out := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		trimmed := strings.TrimSpace(candidate)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func sanitizePromptRepoRef(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return ""
	}
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.ReplaceAll(normalized, ".", "-")
	replacer := strings.NewReplacer("/", "-", "\\", "-", " ", "-")
	normalized = replacer.Replace(normalized)
	for strings.Contains(normalized, "--") {
		normalized = strings.ReplaceAll(normalized, "--", "-")
	}
	return strings.Trim(normalized, "-")
}
