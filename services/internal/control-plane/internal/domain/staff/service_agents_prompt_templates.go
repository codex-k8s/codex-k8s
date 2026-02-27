package staff

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

const (
	promptTemplateSeedModeDryRun = "dry_run"
	promptTemplateSeedModeApply  = "apply"
)

var promptSeedFilePattern = regexp.MustCompile(`^role-([a-z0-9][a-z0-9_-]*)-(work|revise)(?:_([a-zA-Z0-9-]+))?\.md$`)

// ListAgents returns agents visible to principal (global + member projects).
func (s *Service) ListAgents(ctx context.Context, principal Principal, limit int) ([]entitytypes.Agent, error) {
	filter := querytypes.AgentListFilter{Limit: limit}
	if principal.IsPlatformAdmin {
		filter.IncludeAllProjects = true
		return s.agents.List(ctx, filter)
	}

	projectIDs, err := s.listProjectIDsForUser(ctx, principal.UserID)
	if err != nil {
		return nil, err
	}
	filter.ProjectIDs = projectIDs
	return s.agents.List(ctx, filter)
}

// GetAgent returns one agent after RBAC validation.
func (s *Service) GetAgent(ctx context.Context, principal Principal, agentID string) (entitytypes.Agent, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return entitytypes.Agent{}, errs.Validation{Field: "agent_id", Msg: "is required"}
	}

	item, ok, err := s.agents.GetByID(ctx, agentID)
	if err != nil {
		return entitytypes.Agent{}, err
	}
	if !ok {
		return entitytypes.Agent{}, errs.Validation{Field: "agent_id", Msg: "not found"}
	}
	if err := s.requireAgentAccess(ctx, principal, item); err != nil {
		return entitytypes.Agent{}, err
	}
	return item, nil
}

// UpdateAgentSettings updates agent settings with optimistic concurrency.
func (s *Service) UpdateAgentSettings(ctx context.Context, principal Principal, params querytypes.AgentUpdateSettingsParams) (entitytypes.Agent, error) {
	if !principal.IsPlatformAdmin {
		return entitytypes.Agent{}, errs.Forbidden{Msg: "platform admin required"}
	}

	params.AgentID = strings.TrimSpace(params.AgentID)
	if params.AgentID == "" {
		return entitytypes.Agent{}, errs.Validation{Field: "agent_id", Msg: "is required"}
	}
	if params.ExpectedVersion <= 0 {
		return entitytypes.Agent{}, errs.Validation{Field: "expected_version", Msg: "must be a positive integer"}
	}
	if err := validateAgentSettings(params.Settings); err != nil {
		return entitytypes.Agent{}, err
	}

	params.UpdatedByUserID = principal.UserID
	item, ok, err := s.agents.UpdateSettings(ctx, params)
	if err != nil {
		return entitytypes.Agent{}, err
	}
	if ok {
		return item, nil
	}

	_, exists, err := s.agents.GetByID(ctx, params.AgentID)
	if err != nil {
		return entitytypes.Agent{}, err
	}
	if !exists {
		return entitytypes.Agent{}, errs.Validation{Field: "agent_id", Msg: "not found"}
	}
	return entitytypes.Agent{}, errs.Conflict{Msg: "settings version conflict"}
}

// ListPromptTemplateKeys returns prompt template key index.
func (s *Service) ListPromptTemplateKeys(ctx context.Context, principal Principal, filter querytypes.PromptTemplateKeyListFilter) ([]entitytypes.PromptTemplateKeyItem, error) {
	normalizedScope := strings.ToLower(strings.TrimSpace(filter.Scope))
	filter.Scope = normalizedScope

	if principal.IsPlatformAdmin {
		return s.promptTemplates.ListKeys(ctx, filter)
	}

	switch normalizedScope {
	case "", string(enumtypes.PromptTemplateScopeGlobal):
		globalFilter := filter
		globalFilter.Scope = string(enumtypes.PromptTemplateScopeGlobal)
		globalItems, err := s.promptTemplates.ListKeys(ctx, globalFilter)
		if err != nil {
			return nil, err
		}
		if normalizedScope == string(enumtypes.PromptTemplateScopeGlobal) {
			return globalItems, nil
		}

		projectItems, err := s.listPromptTemplateProjectKeysForUser(ctx, principal, filter)
		if err != nil {
			return nil, err
		}
		return mergePromptTemplateKeyItems(globalItems, projectItems, normalizeListLimit(filter.Limit)), nil
	case string(enumtypes.PromptTemplateScopeProject):
		return s.listPromptTemplateProjectKeysForUser(ctx, principal, filter)
	default:
		return nil, errs.Validation{Field: "scope", Msg: "must be one of: global, project"}
	}
}

// ListPromptTemplateVersions returns lifecycle versions for one template key.
func (s *Service) ListPromptTemplateVersions(ctx context.Context, principal Principal, templateKey string, limit int) ([]entitytypes.PromptTemplateVersion, error) {
	key, err := parsePromptTemplateKey(templateKey)
	if err != nil {
		return nil, err
	}
	if err := s.requirePromptTemplateAccess(ctx, principal, key); err != nil {
		return nil, err
	}

	return s.promptTemplates.ListVersions(ctx, querytypes.PromptTemplateVersionListFilter{Key: key, Limit: limit})
}

// CreatePromptTemplateVersion creates next draft version for one key.
func (s *Service) CreatePromptTemplateVersion(ctx context.Context, principal Principal, params querytypes.PromptTemplateVersionCreateParams) (entitytypes.PromptTemplateVersion, error) {
	if !principal.IsPlatformAdmin {
		return entitytypes.PromptTemplateVersion{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if strings.TrimSpace(params.BodyMarkdown) == "" {
		return entitytypes.PromptTemplateVersion{}, errs.Validation{Field: "body_markdown", Msg: "is required"}
	}
	if params.ExpectedVersion < 0 {
		return entitytypes.PromptTemplateVersion{}, errs.Validation{Field: "expected_version", Msg: "must be zero or positive integer"}
	}
	if err := validatePromptTemplateSource(params.Source); err != nil {
		return entitytypes.PromptTemplateVersion{}, err
	}

	params.UpdatedByUserID = principal.UserID
	return s.promptTemplates.CreateVersion(ctx, params)
}

// ActivatePromptTemplateVersion marks selected version as active.
func (s *Service) ActivatePromptTemplateVersion(ctx context.Context, principal Principal, params querytypes.PromptTemplateVersionActivateParams) (entitytypes.PromptTemplateVersion, error) {
	if !principal.IsPlatformAdmin {
		return entitytypes.PromptTemplateVersion{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if params.Version <= 0 {
		return entitytypes.PromptTemplateVersion{}, errs.Validation{Field: "version", Msg: "must be a positive integer"}
	}
	if params.ExpectedVersion <= 0 {
		return entitytypes.PromptTemplateVersion{}, errs.Validation{Field: "expected_version", Msg: "must be a positive integer"}
	}
	if strings.TrimSpace(params.ChangeReason) == "" {
		return entitytypes.PromptTemplateVersion{}, errs.Validation{Field: "change_reason", Msg: "is required"}
	}

	params.UpdatedByUserID = principal.UserID
	return s.promptTemplates.ActivateVersion(ctx, params)
}

// PreviewPromptTemplate resolves selected version or active version with global fallback.
func (s *Service) PreviewPromptTemplate(ctx context.Context, principal Principal, lookup querytypes.PromptTemplatePreviewLookup) (entitytypes.PromptTemplateVersion, error) {
	if err := s.requirePromptTemplateAccess(ctx, principal, lookup.Key); err != nil {
		return entitytypes.PromptTemplateVersion{}, err
	}

	if lookup.Version > 0 {
		item, found, err := s.promptTemplates.GetVersion(ctx, querytypes.PromptTemplateVersionLookup(lookup))
		if err != nil {
			return entitytypes.PromptTemplateVersion{}, err
		}
		if found {
			return item, nil
		}
		return entitytypes.PromptTemplateVersion{}, errs.Validation{Field: "version", Msg: "not found"}
	}

	item, found, err := s.promptTemplates.GetActiveVersion(ctx, lookup)
	if err != nil {
		return entitytypes.PromptTemplateVersion{}, err
	}
	if found {
		return item, nil
	}
	if lookup.Key.Scope != enumtypes.PromptTemplateScopeProject {
		return entitytypes.PromptTemplateVersion{}, errs.Validation{Field: "template_key", Msg: "active version not found"}
	}

	globalKey := lookup.Key
	globalKey.Scope = enumtypes.PromptTemplateScopeGlobal
	globalKey.ScopeID = ""
	globalItem, globalFound, err := s.promptTemplates.GetActiveVersion(ctx, querytypes.PromptTemplatePreviewLookup{Key: globalKey})
	if err != nil {
		return entitytypes.PromptTemplateVersion{}, err
	}
	if !globalFound {
		return entitytypes.PromptTemplateVersion{}, errs.Validation{Field: "template_key", Msg: "active version not found"}
	}
	return globalItem, nil
}

// DiffPromptTemplateVersions loads two explicit versions for diff view.
func (s *Service) DiffPromptTemplateVersions(ctx context.Context, principal Principal, key querytypes.PromptTemplateKey, fromVersion int, toVersion int) (entitytypes.PromptTemplateVersion, entitytypes.PromptTemplateVersion, error) {
	if err := s.requirePromptTemplateAccess(ctx, principal, key); err != nil {
		return entitytypes.PromptTemplateVersion{}, entitytypes.PromptTemplateVersion{}, err
	}
	if fromVersion <= 0 {
		return entitytypes.PromptTemplateVersion{}, entitytypes.PromptTemplateVersion{}, errs.Validation{Field: "from_version", Msg: "must be a positive integer"}
	}
	if toVersion <= 0 {
		return entitytypes.PromptTemplateVersion{}, entitytypes.PromptTemplateVersion{}, errs.Validation{Field: "to_version", Msg: "must be a positive integer"}
	}

	fromItem, found, err := s.promptTemplates.GetVersion(ctx, querytypes.PromptTemplateVersionLookup{Key: key, Version: fromVersion})
	if err != nil {
		return entitytypes.PromptTemplateVersion{}, entitytypes.PromptTemplateVersion{}, err
	}
	if !found {
		return entitytypes.PromptTemplateVersion{}, entitytypes.PromptTemplateVersion{}, errs.Validation{Field: "from_version", Msg: "not found"}
	}

	toItem, found, err := s.promptTemplates.GetVersion(ctx, querytypes.PromptTemplateVersionLookup{Key: key, Version: toVersion})
	if err != nil {
		return entitytypes.PromptTemplateVersion{}, entitytypes.PromptTemplateVersion{}, err
	}
	if !found {
		return entitytypes.PromptTemplateVersion{}, entitytypes.PromptTemplateVersion{}, errs.Validation{Field: "to_version", Msg: "not found"}
	}
	return fromItem, toItem, nil
}

// ListPromptTemplateAuditEvents returns prompt-template scoped audit trail.
func (s *Service) ListPromptTemplateAuditEvents(ctx context.Context, principal Principal, filter querytypes.PromptTemplateAuditListFilter) ([]entitytypes.PromptTemplateAuditEvent, error) {
	if !principal.IsPlatformAdmin {
		projectID := strings.TrimSpace(filter.ProjectID)
		if projectID == "" {
			return nil, errs.Forbidden{Msg: "project_id is required for non-admin user"}
		}
		if err := s.requireProjectAccess(ctx, principal, projectID); err != nil {
			return nil, err
		}
	}
	return s.promptTemplates.ListAuditEvents(ctx, filter)
}

// SyncPromptTemplateSeeds loads role-aware prompt seeds from repository and creates missing DB rows.
func (s *Service) SyncPromptTemplateSeeds(ctx context.Context, principal Principal, params querytypes.PromptTemplateSeedSyncParams) (entitytypes.PromptTemplateSeedSyncResult, error) {
	if !principal.IsPlatformAdmin {
		return entitytypes.PromptTemplateSeedSyncResult{}, errs.Forbidden{Msg: "platform admin required"}
	}

	mode := strings.ToLower(strings.TrimSpace(params.Mode))
	if mode != promptTemplateSeedModeDryRun && mode != promptTemplateSeedModeApply {
		return entitytypes.PromptTemplateSeedSyncResult{}, errs.Validation{Field: "mode", Msg: "must be one of: dry_run, apply"}
	}
	scope := strings.ToLower(strings.TrimSpace(params.Scope))
	if scope != "" && scope != string(enumtypes.PromptTemplateScopeGlobal) {
		return entitytypes.PromptTemplateSeedSyncResult{}, errs.Validation{Field: "scope", Msg: "only global scope is supported"}
	}
	if strings.TrimSpace(s.cfg.PromptSeedsDir) == "" {
		return entitytypes.PromptTemplateSeedSyncResult{}, errs.Validation{Field: "prompt_seeds_dir", Msg: "is not configured"}
	}

	localeFilter := make(map[string]struct{}, len(params.IncludeLocales))
	for _, locale := range params.IncludeLocales {
		value := strings.TrimSpace(locale)
		if value == "" {
			continue
		}
		localeFilter[value] = struct{}{}
	}

	seeds, err := loadPromptSeedFiles(s.cfg.PromptSeedsDir, localeFilter)
	if err != nil {
		return entitytypes.PromptTemplateSeedSyncResult{}, err
	}

	result := entitytypes.PromptTemplateSeedSyncResult{Items: make([]entitytypes.PromptTemplateSeedSyncItem, 0, len(seeds))}
	for _, seed := range seeds {
		item := entitytypes.PromptTemplateSeedSyncItem{TemplateKey: seed.TemplateKey, Action: "skipped"}
		if mode == promptTemplateSeedModeDryRun {
			_, found, lookupErr := s.promptTemplates.GetActiveVersion(ctx, querytypes.PromptTemplatePreviewLookup{Key: seed.Key})
			if lookupErr != nil {
				return entitytypes.PromptTemplateSeedSyncResult{}, lookupErr
			}
			if found {
				item.Reason = "already_exists"
				result.SkippedCount++
			} else {
				item.Action = "created"
				item.Checksum = seed.Checksum
				result.CreatedCount++
			}
			result.Items = append(result.Items, item)
			continue
		}

		createdVersion, created, createErr := s.promptTemplates.CreateSeedIfMissing(ctx, querytypes.PromptTemplateSeedCreateParams{
			Key:             seed.Key,
			BodyMarkdown:    seed.Body,
			UpdatedByUserID: principal.UserID,
		})
		if createErr != nil {
			return entitytypes.PromptTemplateSeedSyncResult{}, createErr
		}
		if created {
			item.Action = "created"
			item.Checksum = createdVersion.Checksum
			result.CreatedCount++
		} else {
			item.Reason = "already_exists"
			result.SkippedCount++
		}
		result.Items = append(result.Items, item)
	}

	sort.Slice(result.Items, func(i, j int) bool { return result.Items[i].TemplateKey < result.Items[j].TemplateKey })
	return result, nil
}

func validateAgentSettings(settings entitytypes.AgentSettings) error {
	mode := strings.TrimSpace(settings.RuntimeMode)
	if mode != "full-env" && mode != "code-only" {
		return errs.Validation{Field: "settings.runtime_mode", Msg: "must be one of: full-env, code-only"}
	}
	if settings.TimeoutSeconds <= 0 {
		return errs.Validation{Field: "settings.timeout_seconds", Msg: "must be a positive integer"}
	}
	if settings.MaxRetryCount < 0 {
		return errs.Validation{Field: "settings.max_retry_count", Msg: "must be zero or positive integer"}
	}
	locale := strings.TrimSpace(settings.PromptLocale)
	if locale == "" {
		return errs.Validation{Field: "settings.prompt_locale", Msg: "is required"}
	}
	return nil
}

func validatePromptTemplateSource(source enumtypes.PromptTemplateSource) error {
	switch strings.TrimSpace(string(source)) {
	case "":
		return nil
	case string(enumtypes.PromptTemplateSourceProjectOverride), string(enumtypes.PromptTemplateSourceGlobalOverride), string(enumtypes.PromptTemplateSourceRepoSeed):
		return nil
	default:
		return errs.Validation{Field: "source", Msg: "must be one of: project_override, global_override, repo_seed"}
	}
}

func parsePromptTemplateKey(raw string) (querytypes.PromptTemplateKey, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return querytypes.PromptTemplateKey{}, errs.Validation{Field: "template_key", Msg: "is required"}
	}

	parts := strings.Split(value, "/")
	if len(parts) < 4 {
		return querytypes.PromptTemplateKey{}, errs.Validation{Field: "template_key", Msg: "invalid format"}
	}

	switch parts[0] {
	case string(enumtypes.PromptTemplateScopeGlobal):
		if len(parts) != 4 {
			return querytypes.PromptTemplateKey{}, errs.Validation{Field: "template_key", Msg: "invalid global key format"}
		}
		kind := enumtypes.PromptTemplateKind(strings.TrimSpace(parts[2]))
		if kind != enumtypes.PromptTemplateKindWork && kind != enumtypes.PromptTemplateKindRevise {
			return querytypes.PromptTemplateKey{}, errs.Validation{Field: "template_key", Msg: "invalid template kind"}
		}
		return querytypes.PromptTemplateKey{Scope: enumtypes.PromptTemplateScopeGlobal, Role: strings.TrimSpace(parts[1]), Kind: kind, Locale: strings.TrimSpace(parts[3])}, nil
	case string(enumtypes.PromptTemplateScopeProject):
		if len(parts) != 5 {
			return querytypes.PromptTemplateKey{}, errs.Validation{Field: "template_key", Msg: "invalid project key format"}
		}
		kind := enumtypes.PromptTemplateKind(strings.TrimSpace(parts[3]))
		if kind != enumtypes.PromptTemplateKindWork && kind != enumtypes.PromptTemplateKindRevise {
			return querytypes.PromptTemplateKey{}, errs.Validation{Field: "template_key", Msg: "invalid template kind"}
		}
		projectID := strings.TrimSpace(parts[1])
		if projectID == "" {
			return querytypes.PromptTemplateKey{}, errs.Validation{Field: "template_key", Msg: "project_id is required for project scope"}
		}
		return querytypes.PromptTemplateKey{Scope: enumtypes.PromptTemplateScopeProject, ScopeID: projectID, Role: strings.TrimSpace(parts[2]), Kind: kind, Locale: strings.TrimSpace(parts[4])}, nil
	default:
		return querytypes.PromptTemplateKey{}, errs.Validation{Field: "template_key", Msg: "scope must be global or project"}
	}
}

// ParsePromptTemplateKey parses public template key format into structured key.
func ParsePromptTemplateKey(raw string) (querytypes.PromptTemplateKey, error) {
	return parsePromptTemplateKey(raw)
}

func mergePromptTemplateKeyItems(left []entitytypes.PromptTemplateKeyItem, right []entitytypes.PromptTemplateKeyItem, limit int) []entitytypes.PromptTemplateKeyItem {
	byKey := make(map[string]entitytypes.PromptTemplateKeyItem, len(left)+len(right))
	for _, item := range left {
		byKey[item.TemplateKey] = item
	}
	for _, item := range right {
		byKey[item.TemplateKey] = item
	}
	out := make([]entitytypes.PromptTemplateKeyItem, 0, len(byKey))
	for _, item := range byKey {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func normalizeListLimit(limit int) int {
	if limit <= 0 {
		return 200
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}

func (s *Service) listProjectIDsForUser(ctx context.Context, userID string) ([]string, error) {
	projects, err := s.projects.ListForUser(ctx, userID, 1000)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(projects))
	for _, project := range projects {
		projectID := strings.TrimSpace(project.ID)
		if projectID == "" {
			continue
		}
		out = append(out, projectID)
	}
	return out, nil
}

func (s *Service) requireProjectAccess(ctx context.Context, principal Principal, projectID string) error {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if principal.IsPlatformAdmin {
		return nil
	}
	_, hasRole, err := s.members.GetRole(ctx, projectID, principal.UserID)
	if err != nil {
		return err
	}
	if !hasRole {
		return errs.Forbidden{Msg: "project access required"}
	}
	return nil
}

func (s *Service) requireAgentAccess(ctx context.Context, principal Principal, agent entitytypes.Agent) error {
	if principal.IsPlatformAdmin || strings.TrimSpace(agent.ProjectID) == "" {
		return nil
	}
	return s.requireProjectAccess(ctx, principal, agent.ProjectID)
}

func (s *Service) requirePromptTemplateAccess(ctx context.Context, principal Principal, key querytypes.PromptTemplateKey) error {
	if principal.IsPlatformAdmin || key.Scope != enumtypes.PromptTemplateScopeProject {
		return nil
	}
	return s.requireProjectAccess(ctx, principal, key.ScopeID)
}

func (s *Service) listPromptTemplateProjectKeysForUser(ctx context.Context, principal Principal, filter querytypes.PromptTemplateKeyListFilter) ([]entitytypes.PromptTemplateKeyItem, error) {
	limit := normalizeListLimit(filter.Limit)
	projectID := strings.TrimSpace(filter.ProjectID)
	if projectID != "" {
		if err := s.requireProjectAccess(ctx, principal, projectID); err != nil {
			return nil, err
		}
		projectFilter := filter
		projectFilter.Scope = string(enumtypes.PromptTemplateScopeProject)
		projectFilter.ProjectID = projectID
		return s.promptTemplates.ListKeys(ctx, projectFilter)
	}

	projectIDs, err := s.listProjectIDsForUser(ctx, principal.UserID)
	if err != nil {
		return nil, err
	}
	aggregated := make([]entitytypes.PromptTemplateKeyItem, 0, len(projectIDs))
	for _, currentProjectID := range projectIDs {
		projectFilter := filter
		projectFilter.Scope = string(enumtypes.PromptTemplateScopeProject)
		projectFilter.ProjectID = currentProjectID
		items, queryErr := s.promptTemplates.ListKeys(ctx, projectFilter)
		if queryErr != nil {
			return nil, queryErr
		}
		aggregated = append(aggregated, items...)
	}
	if len(aggregated) == 0 {
		return []entitytypes.PromptTemplateKeyItem{}, nil
	}
	return mergePromptTemplateKeyItems(nil, aggregated, limit), nil
}

type promptSeedFile struct {
	Key         querytypes.PromptTemplateKey
	TemplateKey string
	Body        string
	Checksum    string
}

func loadPromptSeedFiles(root string, localeFilter map[string]struct{}) ([]promptSeedFile, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errs.Validation{Field: "prompt_seeds_dir", Msg: "is required"}
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, errs.Validation{Field: "prompt_seeds_dir", Msg: fmt.Sprintf("cannot read directory: %v", err)}
	}

	out := make([]promptSeedFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSpace(entry.Name())
		matches := promptSeedFilePattern.FindStringSubmatch(name)
		if len(matches) != 4 {
			continue
		}
		role := strings.TrimSpace(matches[1])
		kind := strings.TrimSpace(matches[2])
		locale := strings.TrimSpace(matches[3])
		if locale == "" {
			locale = "en"
		}
		if len(localeFilter) > 0 {
			if _, ok := localeFilter[locale]; !ok {
				continue
			}
		}

		path := filepath.Join(root, name)
		bodyRaw, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, errs.Validation{Field: "prompt_seeds_dir", Msg: fmt.Sprintf("cannot read seed file %s: %v", name, readErr)}
		}
		body := string(bodyRaw)
		key := querytypes.PromptTemplateKey{Scope: enumtypes.PromptTemplateScopeGlobal, Role: role, Kind: enumtypes.PromptTemplateKind(kind), Locale: locale}
		templateKey := fmt.Sprintf("global/%s/%s/%s", role, kind, locale)
		out = append(out, promptSeedFile{Key: key, TemplateKey: templateKey, Body: body, Checksum: checksumSHA256(body)})
	}
	return out, nil
}

func checksumSHA256(body string) string {
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])
}
