package staff

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	docsetdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/docset"
	configentryrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/configentry"
	learningfeedbackrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/learningfeedback"
	projectrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/project"
	projectmemberrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projectmember"
	projecttokenrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projecttoken"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	staffrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/staffrun"
	userrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/user"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"

	"github.com/google/uuid"

	"github.com/codex-k8s/codex-k8s/libs/go/repo/provider"
	"github.com/jackc/pgx/v5"
)

func (s *Service) resolveRunAccess(ctx context.Context, principal Principal, runID string) (correlationID string, projectID string, err error) {
	if runID == "" {
		return "", "", errs.Validation{Field: "run_id", Msg: "is required"}
	}

	correlationID, projectID, ok, err := s.runs.GetCorrelationByRunID(ctx, runID)
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", errs.Validation{Field: "run_id", Msg: "not found"}
	}

	if !principal.IsPlatformAdmin {
		if projectID == "" {
			return "", "", errs.Forbidden{Msg: "run is not assigned to a project"}
		}
		_, hasRole, err := s.members.GetRole(ctx, projectID, principal.UserID)
		if err != nil {
			return "", "", err
		}
		if !hasRole {
			return "", "", errs.Forbidden{Msg: "project access required"}
		}
	}

	return correlationID, projectID, nil
}

// ListProjects returns projects visible to the principal.
func (s *Service) ListProjects(ctx context.Context, principal Principal, limit int) ([]ProjectView, error) {
	if principal.IsPlatformAdmin {
		items, err := s.projects.ListAll(ctx, limit)
		if err != nil {
			return nil, err
		}
		out := make([]ProjectView, 0, len(items))
		for _, p := range items {
			out = append(out, ProjectView{
				ID:   p.ID,
				Slug: p.Slug,
				Name: p.Name,
				Role: "admin",
			})
		}
		return out, nil
	}

	items, err := s.projects.ListForUser(ctx, principal.UserID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]ProjectView, 0, len(items))
	for _, p := range items {
		out = append(out, ProjectView{
			ID:   p.ID,
			Slug: p.Slug,
			Name: p.Name,
			Role: p.Role,
		})
	}
	return out, nil
}

// GetProject returns a single project visible to the principal.
func (s *Service) GetProject(ctx context.Context, principal Principal, projectID string) (projectrepo.Project, error) {
	if projectID == "" {
		return projectrepo.Project{}, errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if !principal.IsPlatformAdmin {
		_, ok, err := s.members.GetRole(ctx, projectID, principal.UserID)
		if err != nil {
			return projectrepo.Project{}, err
		}
		if !ok {
			return projectrepo.Project{}, errs.Forbidden{Msg: "project access required"}
		}
	}
	p, ok, err := s.projects.GetByID(ctx, projectID)
	if err != nil {
		return projectrepo.Project{}, err
	}
	if !ok {
		return projectrepo.Project{}, errs.Validation{Field: "project_id", Msg: "not found"}
	}
	return p, nil
}

// ListRuns returns runs visible to the principal.
func (s *Service) ListRuns(ctx context.Context, principal Principal, limit int) ([]staffrunrepo.Run, error) {
	if principal.IsPlatformAdmin {
		return s.runs.ListAll(ctx, limit)
	}
	return s.runs.ListForUser(ctx, principal.UserID, limit)
}

// GetRun returns a single run record visible to the principal.
func (s *Service) GetRun(ctx context.Context, principal Principal, runID string) (staffrunrepo.Run, error) {
	if runID == "" {
		return staffrunrepo.Run{}, errs.Validation{Field: "run_id", Msg: "is required"}
	}

	r, ok, err := s.runs.GetByID(ctx, runID)
	if err != nil {
		return staffrunrepo.Run{}, err
	}
	if !ok {
		return staffrunrepo.Run{}, errs.Validation{Field: "run_id", Msg: "not found"}
	}

	if !principal.IsPlatformAdmin {
		if r.ProjectID == "" {
			return staffrunrepo.Run{}, errs.Forbidden{Msg: "run is not assigned to a project"}
		}
		_, hasRole, err := s.members.GetRole(ctx, r.ProjectID, principal.UserID)
		if err != nil {
			return staffrunrepo.Run{}, err
		}
		if !hasRole {
			return staffrunrepo.Run{}, errs.Forbidden{Msg: "project access required"}
		}
	}

	if s.runStatus != nil {
		runtimeState, runtimeErr := s.runStatus.GetRunRuntimeState(ctx, r.ID)
		if runtimeErr == nil {
			r.JobName = runtimeState.JobName
			r.JobNamespace = runtimeState.JobNamespace
			r.Namespace = runtimeState.Namespace
			r.JobExists = runtimeState.JobExists
			r.NamespaceExists = runtimeState.NamespaceExists
		}
	}

	return r, nil
}

// ListRunFlowEvents returns flow events for a run id, enforcing project RBAC.
func (s *Service) ListRunFlowEvents(ctx context.Context, principal Principal, runID string, limit int) ([]staffrunrepo.FlowEvent, error) {
	correlationID, _, err := s.resolveRunAccess(ctx, principal, runID)
	if err != nil {
		return nil, err
	}

	return s.runs.ListEventsByCorrelation(ctx, correlationID, limit)
}

// ListUsers returns all allowed users (platform admin only).
func (s *Service) ListUsers(ctx context.Context, principal Principal, limit int) ([]userrepo.User, error) {
	if !principal.IsPlatformAdmin {
		return nil, errs.Forbidden{Msg: "platform admin required"}
	}
	return s.users.List(ctx, limit)
}

// CreateAllowedUser creates/updates an allowed user record (platform admin only).
func (s *Service) CreateAllowedUser(ctx context.Context, principal Principal, email string, isPlatformAdmin bool) (userrepo.User, error) {
	if !principal.IsPlatformAdmin {
		return userrepo.User{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if email == "" {
		return userrepo.User{}, errs.Validation{Field: "email", Msg: "is required"}
	}
	return s.users.CreateAllowedUser(ctx, email, isPlatformAdmin)
}

// DeleteUser removes a staff user record (RBAC enforced).
func (s *Service) DeleteUser(ctx context.Context, principal Principal, userID string) error {
	if userID == "" {
		return errs.Validation{Field: "user_id", Msg: "is required"}
	}
	if !principal.IsPlatformAdmin {
		return errs.Forbidden{Msg: "platform admin required"}
	}
	if principal.UserID == userID {
		return errs.Forbidden{Msg: "cannot delete self"}
	}

	target, ok, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if !ok {
		return errs.Validation{Field: "user_id", Msg: "not found"}
	}

	if principal.IsPlatformOwner {
		// Owner can delete anyone except themselves (checked above).
	} else {
		// Platform admin cannot delete other admins/owner.
		if target.IsPlatformOwner || target.IsPlatformAdmin {
			return errs.Forbidden{Msg: "cannot delete platform admin"}
		}
	}

	if err := s.users.DeleteByID(ctx, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.Validation{Field: "user_id", Msg: "not found"}
		}
		return err
	}
	return nil
}

// ListProjectMembers returns members for a project (platform admin only in MVP).
func (s *Service) ListProjectMembers(ctx context.Context, principal Principal, projectID string, limit int) ([]projectmemberrepo.Member, error) {
	if !principal.IsPlatformAdmin {
		return nil, errs.Forbidden{Msg: "platform admin required"}
	}
	if projectID == "" {
		return nil, errs.Validation{Field: "project_id", Msg: "is required"}
	}
	return s.members.List(ctx, projectID, limit)
}

// UpsertProjectMemberByEmail sets a role for a user in a project by email (platform owner only).
func (s *Service) UpsertProjectMemberByEmail(ctx context.Context, principal Principal, projectID string, email string, role string) error {
	if !principal.IsPlatformOwner {
		return errs.Forbidden{Msg: "platform owner required"}
	}
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return errs.Validation{Field: "email", Msg: "is required"}
	}
	switch role {
	case "read", "read_write", "admin":
	default:
		return errs.Validation{Field: "role", Msg: fmt.Sprintf("invalid role %q", role)}
	}

	u, ok, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if !ok {
		return errs.Validation{Field: "email", Msg: "not found"}
	}

	return s.members.Upsert(ctx, projectID, u.ID, role)
}

// DeleteProjectMember removes a user from a project (platform owner only).
func (s *Service) DeleteProjectMember(ctx context.Context, principal Principal, projectID string, userID string) error {
	if !principal.IsPlatformOwner {
		return errs.Forbidden{Msg: "platform owner required"}
	}
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if userID == "" {
		return errs.Validation{Field: "user_id", Msg: "is required"}
	}

	u, ok, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if ok && u.IsPlatformOwner {
		return errs.Forbidden{Msg: "cannot remove platform owner from project"}
	}

	if err := s.members.Delete(ctx, projectID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.Validation{Field: "user_id", Msg: "member not found"}
		}
		return err
	}
	return nil
}

// UpsertProjectMember sets a role for a user in a project (platform admin only).
func (s *Service) UpsertProjectMember(ctx context.Context, principal Principal, projectID string, userID string, role string) error {
	if !principal.IsPlatformAdmin {
		return errs.Forbidden{Msg: "platform admin required"}
	}
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if userID == "" {
		return errs.Validation{Field: "user_id", Msg: "is required"}
	}
	switch role {
	case "read", "read_write", "admin":
	default:
		return errs.Validation{Field: "role", Msg: fmt.Sprintf("invalid role %q", role)}
	}
	return s.members.Upsert(ctx, projectID, userID, role)
}

// UpsertProject creates or updates a project (platform admin only).
func (s *Service) UpsertProject(ctx context.Context, principal Principal, slug string, name string) (projectrepo.Project, error) {
	if !principal.IsPlatformAdmin {
		return projectrepo.Project{}, errs.Forbidden{Msg: "platform admin required"}
	}
	slug = strings.TrimSpace(slug)
	name = strings.TrimSpace(name)
	if slug == "" {
		return projectrepo.Project{}, errs.Validation{Field: "slug", Msg: "is required"}
	}
	if name == "" {
		return projectrepo.Project{}, errs.Validation{Field: "name", Msg: "is required"}
	}

	settingsJSON, err := json.Marshal(querytypes.ProjectSettings{
		LearningModeDefault: s.cfg.LearningModeDefault,
	})
	if err != nil {
		return projectrepo.Project{}, fmt.Errorf("marshal project settings: %w", err)
	}

	return s.projects.Upsert(ctx, projectrepo.UpsertParams{
		ID:           uuid.NewString(),
		Slug:         slug,
		Name:         name,
		SettingsJSON: settingsJSON,
	})
}

// DeleteProject deletes a project and all its related data (platform owner only).
func (s *Service) DeleteProject(ctx context.Context, principal Principal, projectID string) error {
	if !principal.IsPlatformOwner {
		return errs.Forbidden{Msg: "platform owner required"}
	}
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if _, ok := s.cfg.ProtectedProjectIDs[projectID]; ok {
		return errs.Forbidden{Msg: "cannot delete platform project"}
	}

	// Best-effort webhook cleanup before removing bindings.
	bindings, err := s.repos.ListForProject(ctx, projectID, 500)
	if err != nil {
		return err
	}
	for _, b := range bindings {
		if s.github == nil {
			continue
		}
		if provider.Provider(b.Provider) != provider.ProviderGitHub {
			continue
		}
		enc, ok, err := s.repos.GetTokenEncrypted(ctx, b.ID)
		if err != nil || !ok {
			continue
		}
		token, err := s.tokencrypt.DecryptString(enc)
		if err != nil || strings.TrimSpace(token) == "" {
			continue
		}
		_ = s.github.DeleteWebhook(ctx, token, b.Owner, b.Name, s.cfg.WebhookSpec.URL)
	}

	// Flow events are not FK-linked, so remove them explicitly.
	if err := s.runs.DeleteFlowEventsByProjectID(ctx, projectID); err != nil {
		return err
	}

	// The rest is cascaded via FK constraints (projects -> repositories/project_members/slots/agent_runs -> learning_feedback).
	if err := s.projects.DeleteByID(ctx, projectID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.Validation{Field: "project_id", Msg: "not found"}
		}
		return err
	}
	return nil
}

// ListProjectRepositories returns repository bindings for a project.
func (s *Service) ListProjectRepositories(ctx context.Context, principal Principal, projectID string, limit int) ([]repocfgrepo.RepositoryBinding, error) {
	if projectID == "" {
		return nil, errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if !principal.IsPlatformAdmin {
		_, ok, err := s.members.GetRole(ctx, projectID, principal.UserID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errs.Forbidden{Msg: "project access required"}
		}
	}
	return s.repos.ListForProject(ctx, projectID, limit)
}

// UpsertProjectRepository attaches a GitHub repository to a project (requires write role).
func (s *Service) UpsertProjectRepository(
	ctx context.Context,
	principal Principal,
	projectID string,
	providerID string,
	owner string,
	name string,
	token string,
	servicesYAMLPath string,
	alias string,
	role string,
	defaultRef string,
	docsRootPath string,
) (repocfgrepo.RepositoryBinding, error) {
	if projectID == "" {
		return repocfgrepo.RepositoryBinding{}, errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if providerID == "" {
		return repocfgrepo.RepositoryBinding{}, errs.Validation{Field: "provider", Msg: "is required"}
	}
	owner = strings.TrimSpace(owner)
	name = strings.TrimSpace(name)
	if owner == "" {
		return repocfgrepo.RepositoryBinding{}, errs.Validation{Field: "owner", Msg: "is required"}
	}
	if name == "" {
		return repocfgrepo.RepositoryBinding{}, errs.Validation{Field: "name", Msg: "is required"}
	}
	if strings.TrimSpace(token) == "" {
		return repocfgrepo.RepositoryBinding{}, errs.Validation{Field: "token", Msg: "is required"}
	}

	memberRole := "admin"
	if !principal.IsPlatformAdmin {
		r, ok, err := s.members.GetRole(ctx, projectID, principal.UserID)
		if err != nil {
			return repocfgrepo.RepositoryBinding{}, err
		}
		if !ok {
			return repocfgrepo.RepositoryBinding{}, errs.Forbidden{Msg: "project access required"}
		}
		memberRole = r
	}
	if memberRole != "admin" && memberRole != "read_write" {
		return repocfgrepo.RepositoryBinding{}, errs.Forbidden{Msg: "project write access required"}
	}

	if servicesYAMLPath = strings.TrimSpace(servicesYAMLPath); servicesYAMLPath == "" {
		servicesYAMLPath = "services.yaml"
	}
	servicesPathNormalized, err := normalizeRepositoryRelativePath(servicesYAMLPath)
	if err != nil {
		return repocfgrepo.RepositoryBinding{}, errs.Validation{Field: "services_yaml_path", Msg: err.Error()}
	}
	aliasNormalized, roleNormalized, defaultRefNormalized, docsRootNormalized, err := normalizeRepositoryTopology(owner, name, alias, role, defaultRef, docsRootPath)
	if err != nil {
		return repocfgrepo.RepositoryBinding{}, errs.Validation{Field: "repository_topology", Msg: err.Error()}
	}

	switch provider.Provider(providerID) {
	case provider.ProviderGitHub:
		if s.github == nil {
			return repocfgrepo.RepositoryBinding{}, errs.Conflict{Msg: "github provider is not configured"}
		}

		info, err := s.github.ValidateRepository(ctx, token, owner, name)
		if err != nil {
			return repocfgrepo.RepositoryBinding{}, err
		}
		if err := s.github.EnsureWebhook(ctx, token, owner, name, s.cfg.WebhookSpec); err != nil {
			return repocfgrepo.RepositoryBinding{}, err
		}

		enc, err := s.tokencrypt.EncryptString(token)
		if err != nil {
			return repocfgrepo.RepositoryBinding{}, fmt.Errorf("encrypt repo token: %w", err)
		}

		return s.repos.Upsert(ctx, repocfgrepo.UpsertParams{
			ProjectID:        projectID,
			Alias:            aliasNormalized,
			Role:             roleNormalized,
			DefaultRef:       defaultRefNormalized,
			Provider:         string(info.Provider),
			ExternalID:       info.ExternalID,
			Owner:            info.Owner,
			Name:             info.Name,
			TokenEncrypted:   enc,
			ServicesYAMLPath: servicesPathNormalized,
			DocsRootPath:     docsRootNormalized,
		})
	default:
		return repocfgrepo.RepositoryBinding{}, errs.Validation{Field: "provider", Msg: fmt.Sprintf("unsupported provider %q", providerID)}
	}
}

// DeleteProjectRepository removes a repository binding from a project.
func (s *Service) DeleteProjectRepository(ctx context.Context, principal Principal, projectID string, repositoryID string) error {
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if repositoryID == "" {
		return errs.Validation{Field: "repository_id", Msg: "is required"}
	}
	if _, ok := s.cfg.ProtectedRepositoryIDs[repositoryID]; ok {
		return errs.Forbidden{Msg: "cannot delete platform repository binding"}
	}
	if platformRepo := getOptionalEnv("CODEXK8S_GITHUB_REPO"); platformRepo != "" {
		platformOwner, platformName, ok := strings.Cut(platformRepo, "/")
		platformOwner = strings.TrimSpace(platformOwner)
		platformName = strings.TrimSpace(platformName)
		if ok && platformOwner != "" && platformName != "" {
			binding, found, err := s.repos.GetByID(ctx, repositoryID)
			if err != nil {
				return err
			}
			if found && strings.EqualFold(binding.Owner, platformOwner) && strings.EqualFold(binding.Name, platformName) {
				return errs.Forbidden{Msg: "cannot delete platform repository binding"}
			}
		}
	}

	role := "admin"
	if !principal.IsPlatformAdmin {
		r, ok, err := s.members.GetRole(ctx, projectID, principal.UserID)
		if err != nil {
			return err
		}
		if !ok {
			return errs.Forbidden{Msg: "project access required"}
		}
		role = r
	}
	if role != "admin" && role != "read_write" {
		return errs.Forbidden{Msg: "project write access required"}
	}

	// Best-effort: attempt to delete the webhook from the provider repo.
	// Errors are intentionally ignored (revoked token, missing permissions, already deleted, etc).
	if s.github != nil {
		bindings, err := s.repos.ListForProject(ctx, projectID, 500)
		if err == nil {
			var b *repocfgrepo.RepositoryBinding
			for i := range bindings {
				if bindings[i].ID == repositoryID {
					b = &bindings[i]
					break
				}
			}
			if b != nil && provider.Provider(b.Provider) == provider.ProviderGitHub {
				enc, ok, err := s.repos.GetTokenEncrypted(ctx, repositoryID)
				if err == nil && ok {
					token, err := s.tokencrypt.DecryptString(enc)
					if err == nil && strings.TrimSpace(token) != "" {
						_ = s.github.DeleteWebhook(ctx, token, b.Owner, b.Name, s.cfg.WebhookSpec.URL)
					}
				}
			}
		}
	}

	if err := s.repos.Delete(ctx, projectID, repositoryID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.Validation{Field: "repository_id", Msg: "not found"}
		}
		return err
	}
	return nil
}

// SetProjectMemberLearningModeOverride sets per-member learning mode override (platform admin only).
func (s *Service) SetProjectMemberLearningModeOverride(ctx context.Context, principal Principal, projectID string, userID string, enabled *bool) error {
	if !principal.IsPlatformAdmin {
		return errs.Forbidden{Msg: "platform admin required"}
	}
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if userID == "" {
		return errs.Validation{Field: "user_id", Msg: "is required"}
	}
	if err := s.members.SetLearningModeOverride(ctx, projectID, userID, enabled); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.Validation{Field: "user_id", Msg: "member not found"}
		}
		return err
	}
	return nil
}

// ListRunLearningFeedback returns feedback entries for a run id.
func (s *Service) ListRunLearningFeedback(ctx context.Context, principal Principal, runID string, limit int) ([]learningfeedbackrepo.Feedback, error) {
	if _, _, err := s.resolveRunAccess(ctx, principal, runID); err != nil {
		return nil, err
	}

	return s.feedback.ListForRun(ctx, runID, limit)
}

func (s *Service) GetProjectGitHubTokens(ctx context.Context, principal Principal, projectID string) (projecttokenrepo.ProjectGitHubTokens, bool, error) {
	if projectID == "" {
		return projecttokenrepo.ProjectGitHubTokens{}, false, errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if !principal.IsPlatformAdmin {
		_, roleOK, err := s.members.GetRole(ctx, projectID, principal.UserID)
		if err != nil {
			return projecttokenrepo.ProjectGitHubTokens{}, false, err
		}
		if !roleOK {
			return projecttokenrepo.ProjectGitHubTokens{}, false, errs.Forbidden{Msg: "project access required"}
		}
	}
	return s.projectTokens.GetByProjectID(ctx, projectID)
}

func (s *Service) UpsertProjectGitHubTokens(ctx context.Context, principal Principal, projectID string, platformTokenRaw *string, botTokenRaw *string, botUsername *string, botEmail *string) error {
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}

	role := "admin"
	if !principal.IsPlatformAdmin {
		r, ok, err := s.members.GetRole(ctx, projectID, principal.UserID)
		if err != nil {
			return err
		}
		if !ok {
			return errs.Forbidden{Msg: "project access required"}
		}
		role = r
	}
	if role != "admin" && role != "read_write" {
		return errs.Forbidden{Msg: "project write access required"}
	}

	var platformEnc []byte
	var botEnc []byte
	if platformTokenRaw != nil {
		raw := strings.TrimSpace(*platformTokenRaw)
		if raw != "" {
			enc, err := s.tokencrypt.EncryptString(raw)
			if err != nil {
				return fmt.Errorf("encrypt project platform token: %w", err)
			}
			platformEnc = enc
		}
	}
	if botTokenRaw != nil {
		raw := strings.TrimSpace(*botTokenRaw)
		if raw != "" {
			enc, err := s.tokencrypt.EncryptString(raw)
			if err != nil {
				return fmt.Errorf("encrypt project bot token: %w", err)
			}
			botEnc = enc
		}
	}

	username := ""
	if botUsername != nil {
		username = strings.TrimSpace(*botUsername)
	}
	email := ""
	if botEmail != nil {
		email = strings.TrimSpace(*botEmail)
	}

	return s.projectTokens.Upsert(ctx, projecttokenrepo.UpsertParams{
		ProjectID:              projectID,
		PlatformTokenEncrypted: platformEnc,
		BotTokenEncrypted:      botEnc,
		BotUsername:            username,
		BotEmail:               email,
	})
}

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

func (s *Service) UpsertRepositoryBotParams(ctx context.Context, principal Principal, projectID string, repositoryID string, botTokenRaw *string, botUsername *string, botEmail *string) error {
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if repositoryID == "" {
		return errs.Validation{Field: "repository_id", Msg: "is required"}
	}

	role := "admin"
	if !principal.IsPlatformAdmin {
		r, ok, err := s.members.GetRole(ctx, projectID, principal.UserID)
		if err != nil {
			return err
		}
		if !ok {
			return errs.Forbidden{Msg: "project access required"}
		}
		role = r
	}
	if role != "admin" && role != "read_write" {
		return errs.Forbidden{Msg: "project write access required"}
	}

	var enc []byte
	if botTokenRaw != nil {
		raw := strings.TrimSpace(*botTokenRaw)
		if raw != "" {
			encrypted, err := s.tokencrypt.EncryptString(raw)
			if err != nil {
				return fmt.Errorf("encrypt repository bot token: %w", err)
			}
			enc = encrypted
		}
	} else {
		// Preserve existing bot token if caller does not provide one.
		current, ok, err := s.repos.GetBotTokenEncrypted(ctx, repositoryID)
		if err != nil {
			return err
		}
		if ok {
			enc = current
		}
	}

	username := ""
	if botUsername != nil {
		username = strings.TrimSpace(*botUsername)
	}
	email := ""
	if botEmail != nil {
		email = strings.TrimSpace(*botEmail)
	}

	return s.repos.UpsertBotParams(ctx, repocfgrepo.RepositoryBotParamsUpsertParams{
		RepositoryID:      repositoryID,
		BotTokenEncrypted: enc,
		BotUsername:       username,
		BotEmail:          email,
	})
}

func (s *Service) RunRepositoryPreflight(ctx context.Context, principal Principal, projectID string, repositoryID string) (valuetypes.GitHubPreflightReport, error) {
	if !principal.IsPlatformAdmin {
		return valuetypes.GitHubPreflightReport{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if projectID == "" {
		return valuetypes.GitHubPreflightReport{}, errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if repositoryID == "" {
		return valuetypes.GitHubPreflightReport{}, errs.Validation{Field: "repository_id", Msg: "is required"}
	}

	repo, ok, err := s.repos.GetByID(ctx, repositoryID)
	if err != nil {
		return valuetypes.GitHubPreflightReport{}, err
	}
	if !ok {
		return valuetypes.GitHubPreflightReport{}, errs.Validation{Field: "repository_id", Msg: "not found"}
	}

	lockToken := uuid.NewString()
	acquiredToken, acquired, err := s.repos.AcquirePreflightLock(ctx, repocfgrepo.RepositoryPreflightLockAcquireParams{
		RepositoryID:   repositoryID,
		LockToken:      lockToken,
		LockedByUserID: principal.UserID,
		LockedUntilUTC: time.Now().UTC().Add(10 * time.Minute),
	})
	if err != nil {
		return valuetypes.GitHubPreflightReport{}, err
	}
	if !acquired {
		return valuetypes.GitHubPreflightReport{}, errs.Conflict{Msg: "repository preflight is already running"}
	}
	lockToken = acquiredToken
	defer func() {
		_ = s.repos.ReleasePreflightLock(ctx, repositoryID, lockToken)
	}()

	platformToken, botToken, platformScope, botScope, err := s.resolveEffectiveGitHubTokens(ctx, projectID, repositoryID)
	if err != nil {
		return valuetypes.GitHubPreflightReport{}, err
	}

	expectedHost, expectedIPs := resolveExpectedIngressIPs(s.cfg.WebhookSpec.URL)

	report := valuetypes.GitHubPreflightReport{
		Status: "running",
		TokenScopes: valuetypes.GitHubPreflightTokenScopes{
			Platform: platformScope,
			Bot:      botScope,
		},
		Checks:     make([]valuetypes.GitHubPreflightCheck, 0, 32),
		Artifacts:  make([]valuetypes.GitHubPreflightArtifact, 0),
		FinishedAt: time.Time{},
	}
	report.Checks = append(report.Checks,
		valuetypes.GitHubPreflightCheck{Name: "github:tokens:platform_scope", Status: "ok", Details: platformScope},
		valuetypes.GitHubPreflightCheck{Name: "github:tokens:bot_scope", Status: "ok", Details: botScope},
	)

	hasFailures := false

	dnsCandidates := make([]dnsCandidate, 0, 8)

	// Always validate that the platform webhook host resolves (best-effort expected ingress).
	if expectedHost != "" {
		dnsCandidates = append(dnsCandidates, dnsCandidate{CheckName: "dns:platform:webhook_host", Domain: expectedHost})
	}

	if s.githubMgmt == nil {
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:preflight", Status: "failed", Details: "github management client is not configured"})
		hasFailures = true
	} else {
		baseBranch, branchErr := s.githubMgmt.GetDefaultBranch(ctx, platformToken, repo.Owner, repo.Name)
		if branchErr != nil {
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:default_branch", Status: "failed", Details: branchErr.Error()})
			hasFailures = true
		} else {
			servicesPath := strings.TrimSpace(repo.ServicesYAMLPath)
			if servicesPath == "" {
				servicesPath = "services.yaml"
			}

			servicesYAML, found, getErr := s.githubMgmt.GetFile(ctx, platformToken, repo.Owner, repo.Name, servicesPath, baseBranch)
			if getErr != nil {
				report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:get", Status: "failed", Details: getErr.Error()})
				hasFailures = true
			} else if !found {
				report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:get", Status: "failed", Details: fmt.Sprintf("%s not found on %s", servicesPath, baseBranch)})
				hasFailures = true
			} else {
				report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:get", Status: "ok"})

				envNames, parseErr := listServicesYAMLEnvironments(servicesYAML)
				if parseErr != nil {
					report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:parse", Status: "failed", Details: parseErr.Error()})
					hasFailures = true
				} else {
					report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:parse", Status: "ok"})

					vars := envVarsMap()

					for _, item := range []struct {
						Env  string
						Slot int
					}{
						{Env: "production", Slot: 0},
						{Env: "ai", Slot: 1},
					} {
						if _, ok := envNames[item.Env]; !ok {
							report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:env:" + item.Env, Status: "skipped", Details: "environment not defined"})
							continue
						}

						domain, source, ns, err := resolveServicesYAMLDomain(servicesYAML, item.Env, item.Slot, vars)
						if err != nil {
							report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:domain:" + item.Env, Status: "failed", Details: err.Error()})
							hasFailures = true
							continue
						}
						if strings.TrimSpace(domain) == "" {
							report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:domain:" + item.Env, Status: "failed", Details: "resolved domain is empty"})
							hasFailures = true
							continue
						}

						report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{
							Name:    "services_yaml:domain:" + item.Env,
							Status:  "ok",
							Details: fmt.Sprintf("source=%s namespace=%s domain=%s", source, ns, domain),
						})
						dnsCandidates = append(dnsCandidates, dnsCandidate{
							CheckName: "dns:services_yaml:" + item.Env + ":" + domain,
							Domain:    domain,
						})
					}
				}
			}
		}
	}

	for _, candidate := range dnsCandidates {
		check := runDNSCheck(candidate.CheckName, candidate.Domain, expectedIPs)
		if check.Status != "ok" {
			hasFailures = true
		}
		report.Checks = append(report.Checks, check)
	}

	if s.githubMgmt != nil {
		ghReport, ghErr := s.githubMgmt.Preflight(ctx, valuetypes.GitHubPreflightParams{
			PlatformToken: platformToken,
			BotToken:      botToken,
			Owner:         repo.Owner,
			Repository:    repo.Name,
			WebhookURL:    s.cfg.WebhookSpec.URL,
			WebhookSecret: s.cfg.WebhookSpec.Secret,
		})
		if ghErr != nil {
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:preflight", Status: "failed", Details: ghErr.Error()})
			hasFailures = true
		} else {
			report.Checks = append(report.Checks, ghReport.Checks...)
			report.Artifacts = append(report.Artifacts, ghReport.Artifacts...)
			if strings.TrimSpace(ghReport.Status) != "ok" {
				hasFailures = true
			}
		}
	}

	report.FinishedAt = time.Now().UTC()
	if hasFailures {
		report.Status = "failed"
	} else {
		report.Status = "ok"
	}

	encoded, _ := json.Marshal(report)
	_ = s.repos.UpsertPreflightReport(ctx, repocfgrepo.RepositoryPreflightReportUpsertParams{
		RepositoryID: repositoryID,
		ReportJSON:   encoded,
	})

	return report, nil
}

func (s *Service) ListDocsetGroups(ctx context.Context, principal Principal, docsetRef string, locale string) ([]querytypes.DocsetGroup, error) {
	if !principal.IsPlatformAdmin {
		return nil, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.githubMgmt == nil {
		return nil, fmt.Errorf("failed_precondition: github management client is not configured")
	}
	token, err := s.resolvePlatformManagementToken(ctx)
	if err != nil {
		return nil, err
	}

	docsetRef = strings.TrimSpace(docsetRef)
	if docsetRef == "" {
		docsetRef = "main"
	}
	locale = strings.ToLower(strings.TrimSpace(locale))
	if locale == "" {
		locale = "ru"
	}

	manifestBlob, ok, err := s.githubMgmt.GetFile(ctx, token, "codex-k8s", "agent-knowledge-base", "docset.manifest.json", docsetRef)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("docset.manifest.json not found at ref %q", docsetRef)
	}
	manifest, err := docsetdomain.ParseManifest(manifestBlob)
	if err != nil {
		return nil, err
	}

	out := make([]querytypes.DocsetGroup, 0, len(manifest.Groups))
	for _, g := range manifest.Groups {
		out = append(out, querytypes.DocsetGroup{
			ID:              g.ID,
			Title:           g.Title.ForLocale(locale),
			Description:     g.Description.ForLocale(locale),
			DefaultSelected: g.DefaultSelected,
		})
	}
	return out, nil
}

func (s *Service) ImportDocset(ctx context.Context, principal Principal, projectID string, repositoryID string, docsetRef string, locale string, groupIDs []string) (querytypes.DocsetImportResult, error) {
	if !principal.IsPlatformAdmin {
		return querytypes.DocsetImportResult{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.githubMgmt == nil {
		return querytypes.DocsetImportResult{}, fmt.Errorf("failed_precondition: github management client is not configured")
	}
	projectID = strings.TrimSpace(projectID)
	repositoryID = strings.TrimSpace(repositoryID)
	if projectID == "" {
		return querytypes.DocsetImportResult{}, errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if repositoryID == "" {
		return querytypes.DocsetImportResult{}, errs.Validation{Field: "repository_id", Msg: "is required"}
	}
	docsetRef = strings.TrimSpace(docsetRef)
	if docsetRef == "" {
		docsetRef = "main"
	}
	locale = strings.ToLower(strings.TrimSpace(locale))
	if locale == "" {
		locale = "ru"
	}

	targetRepo, ok, err := s.repos.GetByID(ctx, repositoryID)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}
	if !ok {
		return querytypes.DocsetImportResult{}, errs.Validation{Field: "repository_id", Msg: "not found"}
	}

	token, _, _, _, err := s.resolveEffectiveGitHubTokens(ctx, projectID, repositoryID)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}

	manifestBlob, ok, err := s.githubMgmt.GetFile(ctx, token, "codex-k8s", "agent-knowledge-base", "docset.manifest.json", docsetRef)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}
	if !ok {
		return querytypes.DocsetImportResult{}, fmt.Errorf("docset.manifest.json not found at ref %q", docsetRef)
	}
	manifest, err := docsetdomain.ParseManifest(manifestBlob)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}

	plan, selectedGroups, err := docsetdomain.BuildImportPlan(manifest, locale, groupIDs)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}

	files := make(map[string][]byte, len(plan.Files)+1)
	lockFiles := make([]valuetypes.DocsetLockFile, 0, len(plan.Files))
	for _, f := range plan.Files {
		blob, ok, err := s.githubMgmt.GetFile(ctx, token, "codex-k8s", "agent-knowledge-base", f.SrcPath, docsetRef)
		if err != nil {
			return querytypes.DocsetImportResult{}, err
		}
		if !ok {
			return querytypes.DocsetImportResult{}, fmt.Errorf("docset source file %q not found at ref %q", f.SrcPath, docsetRef)
		}
		if f.ExpectedSHA256 != "" {
			if got := docsetdomain.SHA256Hex(blob); got != f.ExpectedSHA256 {
				return querytypes.DocsetImportResult{}, fmt.Errorf("sha256 mismatch for %s: got %s want %s", f.SrcPath, got, f.ExpectedSHA256)
			}
		}
		files[f.DstPath] = blob
		lockFiles = append(lockFiles, valuetypes.DocsetLockFile{
			Path:       f.DstPath,
			SHA256:     docsetdomain.SHA256Hex(blob),
			SourcePath: f.SrcPath,
		})
	}

	lock := docsetdomain.NewLock(manifest.Docset.ID, docsetRef, locale, selectedGroups, lockFiles)
	lockBlob, err := docsetdomain.MarshalLock(lock)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}
	files["docs/.docset-lock.json"] = lockBlob

	baseBranch, err := s.githubMgmt.GetDefaultBranch(ctx, token, targetRepo.Owner, targetRepo.Name)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}
	branch := fmt.Sprintf("codex-k8s-docset-import/%s", time.Now().UTC().Format("20060102-150405"))
	title := fmt.Sprintf("chore(docs): import docset %s (%s)", manifest.Docset.ID, docsetRef)
	body := fmt.Sprintf("Docset import\n\n- docset: %s\n- ref: %s\n- locale: %s\n- groups: %s\n- files: %d\n", manifest.Docset.ID, docsetRef, locale, strings.Join(selectedGroups, ", "), len(plan.Files))
	prNumber, prURL, err := s.githubMgmt.CreatePullRequestWithFiles(ctx, token, targetRepo.Owner, targetRepo.Name, baseBranch, branch, title, body, files)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}

	return querytypes.DocsetImportResult{
		RepositoryFullName: targetRepo.Owner + "/" + targetRepo.Name,
		PRNumber:           prNumber,
		PRURL:              prURL,
		Branch:             branch,
		FilesTotal:         len(plan.Files),
	}, nil
}

func (s *Service) SyncDocset(ctx context.Context, principal Principal, projectID string, repositoryID string, docsetRef string) (querytypes.DocsetSyncResult, error) {
	if !principal.IsPlatformAdmin {
		return querytypes.DocsetSyncResult{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.githubMgmt == nil {
		return querytypes.DocsetSyncResult{}, fmt.Errorf("failed_precondition: github management client is not configured")
	}
	projectID = strings.TrimSpace(projectID)
	repositoryID = strings.TrimSpace(repositoryID)
	if projectID == "" {
		return querytypes.DocsetSyncResult{}, errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if repositoryID == "" {
		return querytypes.DocsetSyncResult{}, errs.Validation{Field: "repository_id", Msg: "is required"}
	}
	docsetRef = strings.TrimSpace(docsetRef)
	if docsetRef == "" {
		return querytypes.DocsetSyncResult{}, errs.Validation{Field: "docset_ref", Msg: "is required"}
	}

	targetRepo, ok, err := s.repos.GetByID(ctx, repositoryID)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}
	if !ok {
		return querytypes.DocsetSyncResult{}, errs.Validation{Field: "repository_id", Msg: "not found"}
	}

	token, _, _, _, err := s.resolveEffectiveGitHubTokens(ctx, projectID, repositoryID)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}
	baseBranch, err := s.githubMgmt.GetDefaultBranch(ctx, token, targetRepo.Owner, targetRepo.Name)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}

	lockBlob, ok, err := s.githubMgmt.GetFile(ctx, token, targetRepo.Owner, targetRepo.Name, "docs/.docset-lock.json", baseBranch)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}
	if !ok {
		return querytypes.DocsetSyncResult{}, fmt.Errorf("docset lock not found: docs/.docset-lock.json (run import first)")
	}
	lock, err := docsetdomain.ParseLock(lockBlob)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}
	locale := strings.ToLower(strings.TrimSpace(lock.Docset.Locale))
	if locale == "" {
		locale = "ru"
	}

	manifestBlob, ok, err := s.githubMgmt.GetFile(ctx, token, "codex-k8s", "agent-knowledge-base", "docset.manifest.json", docsetRef)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}
	if !ok {
		return querytypes.DocsetSyncResult{}, fmt.Errorf("docset.manifest.json not found at ref %q", docsetRef)
	}
	manifest, err := docsetdomain.ParseManifest(manifestBlob)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}

	currentSHA := make(map[string]string, len(lock.Files))
	for _, f := range lock.Files {
		blob, ok, err := s.githubMgmt.GetFile(ctx, token, targetRepo.Owner, targetRepo.Name, f.Path, baseBranch)
		if err != nil {
			return querytypes.DocsetSyncResult{}, err
		}
		if !ok {
			currentSHA[f.Path] = ""
			continue
		}
		currentSHA[f.Path] = docsetdomain.SHA256Hex(blob)
	}

	plan, err := docsetdomain.BuildSafeSyncPlan(lock, manifest, locale, currentSHA)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}

	files := make(map[string][]byte, len(plan.Updates)+1)
	updatedLockFiles := make([]valuetypes.DocsetLockFile, 0, len(plan.Updates))
	for _, f := range plan.Updates {
		blob, ok, err := s.githubMgmt.GetFile(ctx, token, "codex-k8s", "agent-knowledge-base", f.SrcPath, docsetRef)
		if err != nil {
			return querytypes.DocsetSyncResult{}, err
		}
		if !ok {
			return querytypes.DocsetSyncResult{}, fmt.Errorf("docset source file %q not found at ref %q", f.SrcPath, docsetRef)
		}
		if f.ExpectedSHA256 != "" {
			if got := docsetdomain.SHA256Hex(blob); got != f.ExpectedSHA256 {
				return querytypes.DocsetSyncResult{}, fmt.Errorf("sha256 mismatch for %s: got %s want %s", f.SrcPath, got, f.ExpectedSHA256)
			}
		}
		files[f.DstPath] = blob
		updatedLockFiles = append(updatedLockFiles, valuetypes.DocsetLockFile{
			Path:       f.DstPath,
			SHA256:     docsetdomain.SHA256Hex(blob),
			SourcePath: f.SrcPath,
		})
	}

	nextLock, err := docsetdomain.UpdateLockForSync(lock, docsetRef, updatedLockFiles)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}
	lockOut, err := docsetdomain.MarshalLock(nextLock)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}
	files["docs/.docset-lock.json"] = lockOut

	branch := fmt.Sprintf("codex-k8s-docset-sync/%s", time.Now().UTC().Format("20060102-150405"))
	title := fmt.Sprintf("chore(docs): sync docset %s (%s)", manifest.Docset.ID, docsetRef)
	body := fmt.Sprintf("Docset sync\n\n- docset: %s\n- ref: %s\n- locale: %s\n- updated: %d\n- drift: %d\n", manifest.Docset.ID, docsetRef, locale, len(plan.Updates), len(plan.Drift))
	prNumber, prURL, err := s.githubMgmt.CreatePullRequestWithFiles(ctx, token, targetRepo.Owner, targetRepo.Name, baseBranch, branch, title, body, files)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}

	return querytypes.DocsetSyncResult{
		RepositoryFullName: targetRepo.Owner + "/" + targetRepo.Name,
		PRNumber:           prNumber,
		PRURL:              prURL,
		Branch:             branch,
		FilesUpdated:       len(plan.Updates),
		FilesDrift:         len(plan.Drift),
	}, nil
}

func (s *Service) resolvePlatformManagementToken(ctx context.Context) (string, error) {
	if s.platformTokens == nil {
		return "", fmt.Errorf("failed_precondition: platform tokens repository is not configured")
	}
	item, ok, err := s.platformTokens.Get(ctx)
	if err != nil {
		return "", err
	}
	if !ok || len(item.PlatformTokenEncrypted) == 0 {
		return "", fmt.Errorf("failed_precondition: platform token is not configured")
	}
	raw, err := s.tokencrypt.DecryptString(item.PlatformTokenEncrypted)
	if err != nil {
		return "", fmt.Errorf("decrypt platform token: %w", err)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("failed_precondition: platform token is empty after decrypt")
	}
	return raw, nil
}

func (s *Service) resolveEffectiveGitHubTokens(ctx context.Context, projectID string, repositoryID string) (platformToken string, botToken string, platformScope string, botScope string, err error) {
	repoPlatformEnc, _, encErr := s.repos.GetTokenEncrypted(ctx, repositoryID)
	if encErr != nil {
		return "", "", "", "", encErr
	}
	if len(repoPlatformEnc) > 0 {
		raw, decErr := s.tokencrypt.DecryptString(repoPlatformEnc)
		if decErr == nil && strings.TrimSpace(raw) != "" {
			platformToken = strings.TrimSpace(raw)
			platformScope = "repository"
		}
	}

	repoBotEnc, _, botErr := s.repos.GetBotTokenEncrypted(ctx, repositoryID)
	if botErr != nil {
		return "", "", "", "", botErr
	}
	if len(repoBotEnc) > 0 {
		raw, decErr := s.tokencrypt.DecryptString(repoBotEnc)
		if decErr == nil && strings.TrimSpace(raw) != "" {
			botToken = strings.TrimSpace(raw)
			botScope = "repository"
		}
	}

	if (platformToken == "" || botToken == "") && s.projectTokens != nil && projectID != "" {
		projPlatformEnc, projBotEnc, _, _, ok, projErr := s.projectTokens.GetEncryptedByProjectID(ctx, projectID)
		if projErr != nil {
			return "", "", "", "", projErr
		}
		if ok {
			if platformToken == "" && len(projPlatformEnc) > 0 {
				raw, decErr := s.tokencrypt.DecryptString(projPlatformEnc)
				if decErr == nil && strings.TrimSpace(raw) != "" {
					platformToken = strings.TrimSpace(raw)
					platformScope = "project"
				}
			}
			if botToken == "" && len(projBotEnc) > 0 {
				raw, decErr := s.tokencrypt.DecryptString(projBotEnc)
				if decErr == nil && strings.TrimSpace(raw) != "" {
					botToken = strings.TrimSpace(raw)
					botScope = "project"
				}
			}
		}
	}

	if (platformToken == "" || botToken == "") && s.platformTokens != nil {
		item, ok, tokErr := s.platformTokens.Get(ctx)
		if tokErr != nil {
			return "", "", "", "", tokErr
		}
		if ok {
			if platformToken == "" && len(item.PlatformTokenEncrypted) > 0 {
				raw, decErr := s.tokencrypt.DecryptString(item.PlatformTokenEncrypted)
				if decErr == nil && strings.TrimSpace(raw) != "" {
					platformToken = strings.TrimSpace(raw)
					platformScope = "platform"
				}
			}
			if botToken == "" && len(item.BotTokenEncrypted) > 0 {
				raw, decErr := s.tokencrypt.DecryptString(item.BotTokenEncrypted)
				if decErr == nil && strings.TrimSpace(raw) != "" {
					botToken = strings.TrimSpace(raw)
					botScope = "platform"
				}
			}
		}
	}

	if platformToken == "" {
		return "", "", "", "", fmt.Errorf("failed_precondition: effective platform token is not configured (repo/project/platform fallback empty)")
	}
	if botToken == "" {
		return "", "", "", "", fmt.Errorf("failed_precondition: effective bot token is not configured (repo/project/platform fallback empty)")
	}
	return platformToken, botToken, strings.TrimSpace(platformScope), strings.TrimSpace(botScope), nil
}
