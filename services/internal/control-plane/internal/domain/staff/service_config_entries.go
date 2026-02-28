package staff

import (
	"context"
	"fmt"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	configentryrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/configentry"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func (s *Service) ListConfigEntries(ctx context.Context, principal Principal, scope string, projectID string, repositoryID string, limit int) ([]configentryrepo.ConfigEntry, error) {
	scopeEnum := enumtypes.ConfigEntryScope(strings.TrimSpace(scope))
	if scopeEnum == "" {
		return nil, errs.Validation{Field: "scope", Msg: "is required"}
	}

	switch scopeEnum {
	case enumtypes.ConfigEntryScopePlatform:
		if !principal.IsPlatformAdmin {
			return nil, errs.Forbidden{Msg: "platform admin required"}
		}
		projectID = ""
		repositoryID = ""
	case enumtypes.ConfigEntryScopeProject:
		if projectID == "" {
			return nil, errs.Validation{Field: "project_id", Msg: "is required"}
		}
		repositoryID = ""
		if !principal.IsPlatformAdmin {
			_, ok, err := s.members.GetRole(ctx, projectID, principal.UserID)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, errs.Forbidden{Msg: "project access required"}
			}
		}
	case enumtypes.ConfigEntryScopeRepository:
		if repositoryID == "" {
			return nil, errs.Validation{Field: "repository_id", Msg: "is required"}
		}
		projectID = ""
		if !principal.IsPlatformAdmin {
			repo, ok, err := s.repos.GetByID(ctx, repositoryID)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, errs.Validation{Field: "repository_id", Msg: "not found"}
			}
			_, okRole, err := s.members.GetRole(ctx, repo.ProjectID, principal.UserID)
			if err != nil {
				return nil, err
			}
			if !okRole {
				return nil, errs.Forbidden{Msg: "project access required"}
			}
		}
	default:
		return nil, errs.Validation{Field: "scope", Msg: fmt.Sprintf("unsupported scope %q", scopeEnum)}
	}

	items, err := s.configEntries.List(ctx, configentryrepo.ListFilter{
		Scope:        scopeEnum,
		ProjectID:    projectID,
		RepositoryID: repositoryID,
		Limit:        limit,
	})
	if err != nil {
		return nil, err
	}

	if scopeEnum == enumtypes.ConfigEntryScopePlatform && principal.IsPlatformAdmin {
		// Keep platform config view in sync with what is actually mounted in Kubernetes.
		// Import is create-if-missing (safe) and does not overwrite existing keys.
		if err := s.importPlatformConfigEntriesFromKubernetes(ctx, principal.UserID); err != nil {
			return nil, err
		}
		items, err = s.configEntries.List(ctx, configentryrepo.ListFilter{
			Scope:        scopeEnum,
			ProjectID:    projectID,
			RepositoryID: repositoryID,
			Limit:        limit,
		})
		if err != nil {
			return nil, err
		}
	}

	return items, nil
}

func (s *Service) importPlatformConfigEntriesFromKubernetes(ctx context.Context, userID string) error {
	if s.k8s == nil {
		return fmt.Errorf("failed_precondition: kubernetes client is not configured")
	}
	if s.tokencrypt == nil {
		return fmt.Errorf("failed_precondition: token crypt service is not configured")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	namespace := getOptionalEnv("CODEXK8S_PLATFORM_NAMESPACE")
	if namespace == "" {
		namespace = getOptionalEnv("CODEXK8S_PRODUCTION_NAMESPACE")
	}
	if namespace == "" {
		return fmt.Errorf("platform namespace is not configured")
	}

	// Build a fast lookup for existing platform keys to ensure import is create-if-missing.
	existing, err := s.configEntries.List(ctx, configentryrepo.ListFilter{
		Scope: enumtypes.ConfigEntryScopePlatform,
		Limit: 5000,
	})
	if err != nil {
		return err
	}
	existingKeys := make(map[string]struct{}, len(existing))
	for _, item := range existing {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			continue
		}
		existingKeys[key] = struct{}{}
	}

	// Import all codex-k8s-* secrets/configmaps from the platform namespace.
	const managedPrefix = "codex-k8s-"

	secretNames, err := s.k8s.ListSecretNames(ctx, namespace)
	if err != nil {
		return err
	}
	for _, secretName := range secretNames {
		secretName = strings.TrimSpace(secretName)
		if !strings.HasPrefix(secretName, managedPrefix) {
			continue
		}
		data, ok, err := s.k8s.GetSecretData(ctx, namespace, secretName)
		if err != nil {
			return err
		}
		if !ok || len(data) == 0 {
			continue
		}
		for key, raw := range data {
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			// Kubernetes secrets can contain optional/empty keys (for example placeholder values).
			// We do not import empty secret values into DB to avoid failing encryption and to keep config
			// governance aligned with config.env rules (skip empty values).
			if strings.TrimSpace(string(raw)) == "" {
				continue
			}
			if _, exists := existingKeys[key]; exists {
				continue
			}
			enc, err := s.tokencrypt.EncryptString(string(raw))
			if err != nil {
				return fmt.Errorf("encrypt imported secret %s: %w", key, err)
			}
			if _, err := s.configEntries.Upsert(ctx, configentryrepo.UpsertParams{
				Scope:           enumtypes.ConfigEntryScopePlatform,
				Kind:            enumtypes.ConfigEntryKindSecret,
				Key:             key,
				ValueEncrypted:  enc,
				SyncTargets:     []string{syncTargetK8sSecretPrefix + namespace + "/" + secretName},
				Mutability:      enumtypes.ConfigEntryMutabilityStartupRequired,
				IsDangerous:     false,
				CreatedByUserID: userID,
				UpdatedByUserID: userID,
			}); err != nil {
				return err
			}
			existingKeys[key] = struct{}{}
		}
	}

	configMapNames, err := s.k8s.ListConfigMapNames(ctx, namespace)
	if err != nil {
		return err
	}
	for _, configMapName := range configMapNames {
		configMapName = strings.TrimSpace(configMapName)
		if !strings.HasPrefix(configMapName, managedPrefix) {
			continue
		}
		data, ok, err := s.k8s.GetConfigMapData(ctx, namespace, configMapName)
		if err != nil {
			return err
		}
		if !ok || len(data) == 0 {
			continue
		}
		for key, value := range data {
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			if strings.TrimSpace(value) == "" {
				continue
			}
			if _, exists := existingKeys[key]; exists {
				continue
			}
			if _, err := s.configEntries.Upsert(ctx, configentryrepo.UpsertParams{
				Scope:           enumtypes.ConfigEntryScopePlatform,
				Kind:            enumtypes.ConfigEntryKindVariable,
				Key:             key,
				ValuePlain:      strings.TrimSpace(value),
				SyncTargets:     []string{syncTargetK8sConfigMapPrefix + namespace + "/" + configMapName},
				Mutability:      enumtypes.ConfigEntryMutabilityStartupRequired,
				IsDangerous:     false,
				CreatedByUserID: userID,
				UpdatedByUserID: userID,
			}); err != nil {
				return err
			}
			existingKeys[key] = struct{}{}
		}
	}

	return nil
}

func (s *Service) UpsertConfigEntry(ctx context.Context, principal Principal, params querytypes.ConfigEntryUpsertParams, dangerousConfirmed bool) (configentryrepo.ConfigEntry, error) {
	params.Scope = enumtypes.ConfigEntryScope(strings.TrimSpace(string(params.Scope)))
	params.Kind = enumtypes.ConfigEntryKind(strings.TrimSpace(string(params.Kind)))
	params.Mutability = enumtypes.ConfigEntryMutability(strings.TrimSpace(string(params.Mutability)))
	params.Key = strings.TrimSpace(params.Key)
	if params.Scope == "" {
		return configentryrepo.ConfigEntry{}, errs.Validation{Field: "scope", Msg: "is required"}
	}
	if params.Kind == "" {
		return configentryrepo.ConfigEntry{}, errs.Validation{Field: "kind", Msg: "is required"}
	}
	if params.Key == "" {
		return configentryrepo.ConfigEntry{}, errs.Validation{Field: "key", Msg: "is required"}
	}

	// Normalize irrelevant scope refs early (affects dangerous-key exists check).
	switch params.Scope {
	case enumtypes.ConfigEntryScopePlatform:
		params.ProjectID = ""
		params.RepositoryID = ""
	case enumtypes.ConfigEntryScopeProject:
		params.RepositoryID = ""
	case enumtypes.ConfigEntryScopeRepository:
		params.ProjectID = ""
	}

	if params.IsDangerous && !dangerousConfirmed {
		exists, err := s.configEntries.Exists(ctx, params.Scope, params.ProjectID, params.RepositoryID, params.Key)
		if err != nil {
			return configentryrepo.ConfigEntry{}, err
		}
		if exists {
			return configentryrepo.ConfigEntry{}, errs.Validation{Field: "dangerous_confirmed", Msg: "is required for updates to dangerous keys"}
		}
	}

	switch params.Scope {
	case enumtypes.ConfigEntryScopePlatform:
		if !principal.IsPlatformAdmin {
			return configentryrepo.ConfigEntry{}, errs.Forbidden{Msg: "platform admin required"}
		}
	case enumtypes.ConfigEntryScopeProject:
		if params.ProjectID == "" {
			return configentryrepo.ConfigEntry{}, errs.Validation{Field: "project_id", Msg: "is required"}
		}
		if !principal.IsPlatformAdmin {
			role, ok, err := s.members.GetRole(ctx, params.ProjectID, principal.UserID)
			if err != nil {
				return configentryrepo.ConfigEntry{}, err
			}
			if !ok {
				return configentryrepo.ConfigEntry{}, errs.Forbidden{Msg: "project access required"}
			}
			if role != "admin" && role != "read_write" {
				return configentryrepo.ConfigEntry{}, errs.Forbidden{Msg: "project write access required"}
			}
		}
	case enumtypes.ConfigEntryScopeRepository:
		if params.RepositoryID == "" {
			return configentryrepo.ConfigEntry{}, errs.Validation{Field: "repository_id", Msg: "is required"}
		}
		if !principal.IsPlatformAdmin {
			repo, ok, err := s.repos.GetByID(ctx, params.RepositoryID)
			if err != nil {
				return configentryrepo.ConfigEntry{}, err
			}
			if !ok {
				return configentryrepo.ConfigEntry{}, errs.Validation{Field: "repository_id", Msg: "not found"}
			}
			role, okRole, err := s.members.GetRole(ctx, repo.ProjectID, principal.UserID)
			if err != nil {
				return configentryrepo.ConfigEntry{}, err
			}
			if !okRole {
				return configentryrepo.ConfigEntry{}, errs.Forbidden{Msg: "project access required"}
			}
			if role != "admin" && role != "read_write" {
				return configentryrepo.ConfigEntry{}, errs.Forbidden{Msg: "project write access required"}
			}
		}
	default:
		return configentryrepo.ConfigEntry{}, errs.Validation{Field: "scope", Msg: fmt.Sprintf("unsupported scope %q", params.Scope)}
	}

	switch params.Kind {
	case enumtypes.ConfigEntryKindVariable:
		params.ValuePlain = strings.TrimSpace(params.ValuePlain)
		params.ValueEncrypted = nil
	case enumtypes.ConfigEntryKindSecret:
		params.ValuePlain = ""
		if len(params.ValueEncrypted) == 0 {
			return configentryrepo.ConfigEntry{}, errs.Validation{Field: "value_secret", Msg: "is required"}
		}
	default:
		return configentryrepo.ConfigEntry{}, errs.Validation{Field: "kind", Msg: fmt.Sprintf("unsupported kind %q", params.Kind)}
	}

	params.UpdatedByUserID = principal.UserID
	if params.CreatedByUserID == "" {
		params.CreatedByUserID = principal.UserID
	}

	item, err := s.configEntries.Upsert(ctx, params)
	if err != nil {
		return configentryrepo.ConfigEntry{}, err
	}
	if err := s.syncConfigEntryTargets(ctx, params); err != nil {
		return configentryrepo.ConfigEntry{}, err
	}
	return item, nil
}

func (s *Service) DeleteConfigEntry(ctx context.Context, principal Principal, configEntryID string) error {
	configEntryID = strings.TrimSpace(configEntryID)
	if configEntryID == "" {
		return errs.Validation{Field: "config_entry_id", Msg: "is required"}
	}

	item, ok, err := s.configEntries.GetByID(ctx, configEntryID)
	if err != nil {
		return err
	}
	if !ok {
		return errs.Validation{Field: "config_entry_id", Msg: "not found"}
	}

	switch enumtypes.ConfigEntryScope(strings.TrimSpace(string(item.Scope))) {
	case enumtypes.ConfigEntryScopePlatform:
		if !principal.IsPlatformAdmin {
			return errs.Forbidden{Msg: "platform admin required"}
		}
	case enumtypes.ConfigEntryScopeProject:
		projectID := strings.TrimSpace(item.ProjectID)
		if projectID == "" {
			return errs.Validation{Field: "config_entry_id", Msg: "project_id is empty"}
		}
		if !principal.IsPlatformAdmin {
			role, okRole, err := s.members.GetRole(ctx, projectID, principal.UserID)
			if err != nil {
				return err
			}
			if !okRole {
				return errs.Forbidden{Msg: "project access required"}
			}
			if role != "admin" && role != "read_write" {
				return errs.Forbidden{Msg: "project write access required"}
			}
		}
	case enumtypes.ConfigEntryScopeRepository:
		repositoryID := strings.TrimSpace(item.RepositoryID)
		if repositoryID == "" {
			return errs.Validation{Field: "config_entry_id", Msg: "repository_id is empty"}
		}
		if !principal.IsPlatformAdmin {
			repo, ok, err := s.repos.GetByID(ctx, repositoryID)
			if err != nil {
				return err
			}
			if !ok {
				return errs.Validation{Field: "config_entry_id", Msg: "repository binding not found"}
			}
			role, okRole, err := s.members.GetRole(ctx, repo.ProjectID, principal.UserID)
			if err != nil {
				return err
			}
			if !okRole {
				return errs.Forbidden{Msg: "project access required"}
			}
			if role != "admin" && role != "read_write" {
				return errs.Forbidden{Msg: "project write access required"}
			}
		}
	default:
		return errs.Validation{Field: "config_entry_id", Msg: fmt.Sprintf("unsupported scope %q", item.Scope)}
	}

	return s.configEntries.Delete(ctx, configEntryID)
}

const (
	syncTargetGitHubEnvSecretPrefix = "github_env_secret:"
	syncTargetGitHubEnvVarPrefix    = "github_env_var:"
	syncTargetK8sSecretPrefix       = "k8s_secret:"
	syncTargetK8sConfigMapPrefix    = "k8s_configmap:"
)

func (s *Service) syncConfigEntryTargets(ctx context.Context, params querytypes.ConfigEntryUpsertParams) error {
	if len(params.SyncTargets) == 0 {
		return nil
	}

	kind := enumtypes.ConfigEntryKind(strings.TrimSpace(string(params.Kind)))
	mutability := enumtypes.ConfigEntryMutability(strings.TrimSpace(string(params.Mutability)))
	if mutability == "" {
		mutability = enumtypes.ConfigEntryMutabilityStartupRequired
	}
	key := strings.TrimSpace(params.Key)
	if key == "" {
		return nil
	}

	value := ""
	switch kind {
	case enumtypes.ConfigEntryKindVariable:
		value = params.ValuePlain
	case enumtypes.ConfigEntryKindSecret:
		if len(params.ValueEncrypted) == 0 {
			return nil
		}
		plain, err := s.tokencrypt.DecryptString(params.ValueEncrypted)
		if err != nil {
			return fmt.Errorf("decrypt config entry %s: %w", key, err)
		}
		value = plain
	default:
		return nil
	}
	if strings.TrimSpace(value) == "" {
		// Empty values are persisted, but we avoid syncing them to external systems by default.
		return nil
	}

	for _, rawTarget := range params.SyncTargets {
		target := strings.TrimSpace(rawTarget)
		if target == "" {
			continue
		}

		switch {
		case strings.HasPrefix(target, syncTargetGitHubEnvSecretPrefix):
			envName := strings.TrimSpace(strings.TrimPrefix(target, syncTargetGitHubEnvSecretPrefix))
			if err := s.syncGitHubEnvironmentValue(ctx, params, envName, "secret", key, value, mutability); err != nil {
				return err
			}
		case strings.HasPrefix(target, syncTargetGitHubEnvVarPrefix):
			envName := strings.TrimSpace(strings.TrimPrefix(target, syncTargetGitHubEnvVarPrefix))
			if err := s.syncGitHubEnvironmentValue(ctx, params, envName, "variable", key, value, mutability); err != nil {
				return err
			}
		case strings.HasPrefix(target, syncTargetK8sSecretPrefix):
			spec := strings.TrimSpace(strings.TrimPrefix(target, syncTargetK8sSecretPrefix))
			ns, name, err := parseNamespaceNameSpec(spec)
			if err != nil {
				return errs.Validation{Field: "sync_targets", Msg: err.Error()}
			}
			if err := s.syncKubernetesSecret(ctx, ns, name, key, value, mutability); err != nil {
				return err
			}
		case strings.HasPrefix(target, syncTargetK8sConfigMapPrefix):
			if kind != enumtypes.ConfigEntryKindVariable {
				return errs.Validation{Field: "sync_targets", Msg: "k8s configmap sync target requires kind=variable"}
			}
			spec := strings.TrimSpace(strings.TrimPrefix(target, syncTargetK8sConfigMapPrefix))
			ns, name, err := parseNamespaceNameSpec(spec)
			if err != nil {
				return errs.Validation{Field: "sync_targets", Msg: err.Error()}
			}
			if err := s.syncKubernetesConfigMap(ctx, ns, name, key, value, mutability); err != nil {
				return err
			}
		default:
			return errs.Validation{Field: "sync_targets", Msg: fmt.Sprintf("unsupported sync target %q", target)}
		}
	}

	return nil
}

func (s *Service) syncGitHubEnvironmentValue(
	ctx context.Context,
	params querytypes.ConfigEntryUpsertParams,
	envName string,
	targetKind string, // secret|variable
	key string,
	value string,
	mutability enumtypes.ConfigEntryMutability,
) error {
	if s.githubMgmt == nil {
		return fmt.Errorf("failed_precondition: github management client is not configured")
	}
	envName = strings.TrimSpace(envName)
	if envName == "" {
		return errs.Validation{Field: "sync_targets", Msg: "github environment name is required"}
	}

	repos, err := s.resolveGitHubReposForConfigSync(ctx, params)
	if err != nil {
		return err
	}
	for _, repo := range repos {
		platformToken, _, _, _, tokenErr := s.resolveEffectiveGitHubTokens(ctx, params.ProjectID, repo.ID)
		if params.Scope == enumtypes.ConfigEntryScopePlatform {
			platformToken, tokenErr = s.resolvePlatformManagementToken(ctx)
		}
		if tokenErr != nil {
			return tokenErr
		}

		if err := s.githubMgmt.EnsureEnvironment(ctx, platformToken, repo.Owner, repo.Name, envName); err != nil {
			return err
		}

		switch targetKind {
		case "secret":
			if mutability == enumtypes.ConfigEntryMutabilityStartupRequired {
				names, err := s.githubMgmt.ListEnvSecretNames(ctx, platformToken, repo.Owner, repo.Name, envName)
				if err != nil {
					return err
				}
				if _, exists := names[key]; exists {
					continue
				}
			}
			if err := s.githubMgmt.UpsertEnvSecret(ctx, platformToken, repo.Owner, repo.Name, envName, key, value); err != nil {
				return err
			}
		case "variable":
			existing, err := s.githubMgmt.ListEnvVariableValues(ctx, platformToken, repo.Owner, repo.Name, envName)
			if err != nil {
				return err
			}
			if current, ok := existing[key]; ok {
				if mutability == enumtypes.ConfigEntryMutabilityStartupRequired {
					continue
				}
				if strings.TrimSpace(current) == strings.TrimSpace(value) {
					continue
				}
			}
			if err := s.githubMgmt.UpsertEnvVariable(ctx, platformToken, repo.Owner, repo.Name, envName, key, value); err != nil {
				return err
			}
		default:
			return errs.Validation{Field: "sync_targets", Msg: fmt.Sprintf("unsupported github target kind %q", targetKind)}
		}
	}
	return nil
}

func (s *Service) resolveGitHubReposForConfigSync(ctx context.Context, params querytypes.ConfigEntryUpsertParams) ([]repocfgrepo.RepositoryBinding, error) {
	scope := enumtypes.ConfigEntryScope(strings.TrimSpace(string(params.Scope)))
	switch scope {
	case enumtypes.ConfigEntryScopePlatform:
		fullName := getOptionalEnv("CODEXK8S_GITHUB_REPO")
		owner, name, err := parseGitHubFullName(fullName)
		if err != nil {
			return nil, err
		}
		return []repocfgrepo.RepositoryBinding{{Owner: owner, Name: name}}, nil
	case enumtypes.ConfigEntryScopeProject:
		if params.ProjectID == "" {
			return nil, errs.Validation{Field: "project_id", Msg: "is required"}
		}
		return s.repos.ListForProject(ctx, params.ProjectID, 1000)
	case enumtypes.ConfigEntryScopeRepository:
		if params.RepositoryID == "" {
			return nil, errs.Validation{Field: "repository_id", Msg: "is required"}
		}
		repo, ok, err := s.repos.GetByID(ctx, params.RepositoryID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errs.Validation{Field: "repository_id", Msg: "not found"}
		}
		return []repocfgrepo.RepositoryBinding{repo}, nil
	default:
		return nil, errs.Validation{Field: "scope", Msg: fmt.Sprintf("unsupported scope %q", scope)}
	}
}

func (s *Service) syncKubernetesSecret(ctx context.Context, namespace string, secretName string, key string, value string, mutability enumtypes.ConfigEntryMutability) error {
	if s.k8s == nil {
		return fmt.Errorf("failed_precondition: kubernetes client is not configured")
	}
	namespace = strings.TrimSpace(namespace)
	secretName = strings.TrimSpace(secretName)
	key = strings.TrimSpace(key)
	if namespace == "" || secretName == "" || key == "" {
		return errs.Validation{Field: "sync_targets", Msg: "k8s secret namespace/name/key are required"}
	}

	existing, ok, err := s.k8s.GetSecretData(ctx, namespace, secretName)
	if err != nil {
		return err
	}
	if !ok {
		existing = map[string][]byte{}
	}
	if _, exists := existing[key]; exists && mutability == enumtypes.ConfigEntryMutabilityStartupRequired {
		return nil
	}

	merged := make(map[string][]byte, len(existing)+1)
	for k, v := range existing {
		merged[k] = append([]byte(nil), v...)
	}
	merged[key] = []byte(value)
	return s.k8s.UpsertSecret(ctx, namespace, secretName, merged)
}

func (s *Service) syncKubernetesConfigMap(ctx context.Context, namespace string, configMapName string, key string, value string, mutability enumtypes.ConfigEntryMutability) error {
	if s.k8s == nil {
		return fmt.Errorf("failed_precondition: kubernetes client is not configured")
	}
	namespace = strings.TrimSpace(namespace)
	configMapName = strings.TrimSpace(configMapName)
	key = strings.TrimSpace(key)
	if namespace == "" || configMapName == "" || key == "" {
		return errs.Validation{Field: "sync_targets", Msg: "k8s configmap namespace/name/key are required"}
	}

	existing, ok, err := s.k8s.GetConfigMapData(ctx, namespace, configMapName)
	if err != nil {
		return err
	}
	if !ok {
		existing = map[string]string{}
	}
	if _, exists := existing[key]; exists && mutability == enumtypes.ConfigEntryMutabilityStartupRequired {
		return nil
	}

	merged := make(map[string]string, len(existing)+1)
	for k, v := range existing {
		merged[k] = v
	}
	merged[key] = value
	return s.k8s.UpsertConfigMap(ctx, namespace, configMapName, merged)
}
