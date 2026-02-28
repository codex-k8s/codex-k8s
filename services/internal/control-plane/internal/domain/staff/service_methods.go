package staff

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	learningfeedbackrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/learningfeedback"
	projectrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/project"
	projectmemberrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projectmember"
	projecttokenrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projecttoken"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	staffrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/staffrun"
	userrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/user"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"

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
