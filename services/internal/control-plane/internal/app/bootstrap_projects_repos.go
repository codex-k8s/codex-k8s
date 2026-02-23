package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/codex-k8s/codex-k8s/libs/go/crypto/tokencrypt"
	"github.com/codex-k8s/codex-k8s/libs/go/repo/provider"
	platformtokenrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/platformtoken"
	projectrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/project"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type seedBootstrapProjectsAndRepositoriesParams struct {
	GitHubRepo             string
	FirstProjectGitHubRepo string
	LearningModeDefault    bool
	GitHubPAT              string

	TokenCrypt     *tokencrypt.Service
	PlatformTokens platformtokenrepo.Repository
	Projects       projectrepo.Repository
	Repos          repocfgrepo.Repository
	GitHub         provider.RepositoryProvider
	Logger         *slog.Logger
}

type seedBootstrapProjectsAndRepositoriesResult struct {
	ProtectedProjectIDs    map[string]struct{}
	ProtectedRepositoryIDs map[string]struct{}
}

func seedBootstrapProjectsAndRepositories(ctx context.Context, params seedBootstrapProjectsAndRepositoriesParams) (seedBootstrapProjectsAndRepositoriesResult, error) {
	if params.Projects == nil {
		return seedBootstrapProjectsAndRepositoriesResult{}, fmt.Errorf("projects repository is required")
	}
	if params.Repos == nil {
		return seedBootstrapProjectsAndRepositoriesResult{}, fmt.Errorf("repositories repository is required")
	}
	if params.TokenCrypt == nil {
		return seedBootstrapProjectsAndRepositoriesResult{}, fmt.Errorf("token crypt service is required")
	}

	logger := params.Logger
	if logger == nil {
		logger = slog.Default()
	}

	platformBinding, err := ensureBootstrapGitHubRepositoryBinding(ctx, ensureBootstrapGitHubRepositoryBindingParams{
		RepoFullName:        params.GitHubRepo,
		LearningModeDefault: params.LearningModeDefault,
		TokenRaw:            resolveBootstrapGitHubToken(ctx, params.GitHubPAT, params.PlatformTokens, params.TokenCrypt),
		TokenCrypt:          params.TokenCrypt,
		Projects:            params.Projects,
		Repos:               params.Repos,
		GitHub:              params.GitHub,
		DefaultServicesYAML: "services.yaml",
		Logger:              logger,
		LogKey:              "platform_repo",
	})
	if err != nil {
		return seedBootstrapProjectsAndRepositoriesResult{}, fmt.Errorf("seed platform repository binding: %w", err)
	}

	protectedProjects := map[string]struct{}{platformBinding.ProjectID: {}}
	protectedRepos := map[string]struct{}{platformBinding.RepositoryID: {}}

	firstRepo := strings.TrimSpace(params.FirstProjectGitHubRepo)
	if firstRepo != "" && !strings.EqualFold(firstRepo, strings.TrimSpace(params.GitHubRepo)) {
		if _, err := ensureBootstrapGitHubRepositoryBinding(ctx, ensureBootstrapGitHubRepositoryBindingParams{
			RepoFullName:        firstRepo,
			LearningModeDefault: params.LearningModeDefault,
			TokenRaw:            resolveBootstrapGitHubToken(ctx, params.GitHubPAT, params.PlatformTokens, params.TokenCrypt),
			TokenCrypt:          params.TokenCrypt,
			Projects:            params.Projects,
			Repos:               params.Repos,
			GitHub:              params.GitHub,
			DefaultServicesYAML: "services.yaml",
			Logger:              logger,
			LogKey:              "first_project_repo",
		}); err != nil {
			return seedBootstrapProjectsAndRepositoriesResult{}, fmt.Errorf("seed first project repository binding: %w", err)
		}
	}

	return seedBootstrapProjectsAndRepositoriesResult{
		ProtectedProjectIDs:    protectedProjects,
		ProtectedRepositoryIDs: protectedRepos,
	}, nil
}

func resolveBootstrapGitHubToken(ctx context.Context, envToken string, platformTokens platformtokenrepo.Repository, tokenCrypt *tokencrypt.Service) string {
	if strings.TrimSpace(envToken) != "" {
		return strings.TrimSpace(envToken)
	}
	if platformTokens == nil || tokenCrypt == nil {
		return ""
	}

	stored, ok, err := platformTokens.Get(ctx)
	if err != nil || !ok {
		return ""
	}
	if len(stored.PlatformTokenEncrypted) == 0 {
		return ""
	}
	raw, err := tokenCrypt.DecryptString(stored.PlatformTokenEncrypted)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(raw)
}

type ensureBootstrapGitHubRepositoryBindingParams struct {
	RepoFullName        string
	LearningModeDefault bool
	TokenRaw            string

	TokenCrypt          *tokencrypt.Service
	Projects            projectrepo.Repository
	Repos               repocfgrepo.Repository
	GitHub              provider.RepositoryProvider
	DefaultServicesYAML string

	Logger *slog.Logger
	LogKey string
}

type ensuredBootstrapGitHubRepositoryBinding struct {
	ProjectID    string
	RepositoryID string
}

func ensureBootstrapGitHubRepositoryBinding(ctx context.Context, params ensureBootstrapGitHubRepositoryBindingParams) (ensuredBootstrapGitHubRepositoryBinding, error) {
	if params.Projects == nil {
		return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("projects repository is required")
	}
	if params.Repos == nil {
		return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("repositories repository is required")
	}
	if params.TokenCrypt == nil {
		return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("token crypt service is required")
	}

	repoFullName := strings.TrimSpace(params.RepoFullName)
	if repoFullName == "" {
		return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("repo full name is required")
	}
	owner, name, ok := strings.Cut(repoFullName, "/")
	if !ok || strings.TrimSpace(owner) == "" || strings.TrimSpace(name) == "" {
		return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("repo full name must be in owner/name form, got %q", repoFullName)
	}
	owner = strings.TrimSpace(owner)
	name = strings.TrimSpace(name)

	logger := params.Logger
	if logger == nil {
		logger = slog.Default()
	}
	logKey := strings.TrimSpace(params.LogKey)
	if logKey == "" {
		logKey = "repo"
	}

	if existing, found, err := params.Repos.FindByProviderOwnerName(ctx, string(provider.ProviderGitHub), owner, name); err != nil {
		return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("find repository binding by owner/name: %w", err)
	} else if found {
		logger.Info("bootstrap repository binding present", logKey, repoFullName, "project_id", existing.ProjectID, "repository_id", existing.RepositoryID)
		return ensuredBootstrapGitHubRepositoryBinding{ProjectID: existing.ProjectID, RepositoryID: existing.RepositoryID}, nil
	}

	if params.GitHub == nil {
		return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("github provider is required to seed repository binding for %s", repoFullName)
	}
	token := strings.TrimSpace(params.TokenRaw)
	if token == "" {
		return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("github token is required to seed repository binding for %s (set CODEXK8S_GITHUB_PAT)", repoFullName)
	}

	info, err := params.GitHub.ValidateRepository(ctx, token, owner, name)
	if err != nil {
		return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("validate repository %s: %w", repoFullName, err)
	}

	servicesYAMLPath := strings.TrimSpace(params.DefaultServicesYAML)
	if servicesYAMLPath == "" {
		servicesYAMLPath = "services.yaml"
	}

	if byExternalID, found, err := params.Repos.FindByProviderExternalID(ctx, string(info.Provider), info.ExternalID); err != nil {
		return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("find repository binding by external id: %w", err)
	} else if found {
		enc, err := params.TokenCrypt.EncryptString(token)
		if err != nil {
			return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("encrypt repository token: %w", err)
		}

		updated, err := params.Repos.Upsert(ctx, repocfgrepo.UpsertParams{
			ProjectID:        byExternalID.ProjectID,
			Alias:            strings.ToLower(strings.TrimSpace(info.Owner + "-" + info.Name)),
			Role:             "service",
			DefaultRef:       "main",
			Provider:         string(info.Provider),
			ExternalID:       info.ExternalID,
			Owner:            info.Owner,
			Name:             info.Name,
			TokenEncrypted:   enc,
			ServicesYAMLPath: byExternalID.ServicesYAMLPath,
			DocsRootPath:     "",
		})
		if err != nil {
			return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("upsert existing repository binding: %w", err)
		}
		logger.Info("bootstrap repository binding ensured", logKey, info.FullName, "project_id", updated.ProjectID, "repository_id", updated.ID)
		return ensuredBootstrapGitHubRepositoryBinding{ProjectID: updated.ProjectID, RepositoryID: updated.ID}, nil
	}

	settingsJSON, err := json.Marshal(querytypes.ProjectSettings{
		LearningModeDefault: params.LearningModeDefault,
	})
	if err != nil {
		return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("marshal project settings: %w", err)
	}

	project, err := params.Projects.Upsert(ctx, projectrepo.UpsertParams{
		ID:           uuid.NewString(),
		Slug:         info.FullName,
		Name:         info.Name,
		SettingsJSON: settingsJSON,
	})
	if err != nil {
		return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("upsert project for %s: %w", info.FullName, err)
	}

	enc, err := params.TokenCrypt.EncryptString(token)
	if err != nil {
		return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("encrypt repository token: %w", err)
	}

	binding, err := params.Repos.Upsert(ctx, repocfgrepo.UpsertParams{
		ProjectID:        project.ID,
		Alias:            strings.ToLower(strings.TrimSpace(info.Owner + "-" + info.Name)),
		Role:             "service",
		DefaultRef:       "main",
		Provider:         string(info.Provider),
		ExternalID:       info.ExternalID,
		Owner:            info.Owner,
		Name:             info.Name,
		TokenEncrypted:   enc,
		ServicesYAMLPath: servicesYAMLPath,
		DocsRootPath:     "",
	})
	if err != nil {
		return ensuredBootstrapGitHubRepositoryBinding{}, fmt.Errorf("upsert repository binding for %s: %w", info.FullName, err)
	}

	logger.Info("bootstrap repository binding created", logKey, info.FullName, "project_id", project.ID, "repository_id", binding.ID)
	return ensuredBootstrapGitHubRepositoryBinding{ProjectID: project.ID, RepositoryID: binding.ID}, nil
}
