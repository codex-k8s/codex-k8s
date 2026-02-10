package staff

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/errs"
	learningfeedbackrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/learningfeedback"
	projectrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/project"
	projectmemberrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/projectmember"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/repocfg"
	staffrunrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/staffrun"
	userrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/user"

	"github.com/google/uuid"

	"github.com/codex-k8s/codex-k8s/libs/go/crypto/tokencrypt"
	"github.com/codex-k8s/codex-k8s/libs/go/repo/provider"
)

// Config defines staff service behavior.
type Config struct {
	// LearningModeDefault is the default for newly created projects.
	LearningModeDefault bool

	// WebhookSpec is used when attaching repositories to projects.
	WebhookSpec provider.WebhookSpec
}

// Service exposes staff-only read/write operations protected by JWT + RBAC.
type Service struct {
	cfg      Config
	users    userrepo.Repository
	projects projectrepo.Repository
	members  projectmemberrepo.Repository
	repos    repocfgrepo.Repository
	feedback learningfeedbackrepo.Repository
	runs     staffrunrepo.Repository

	tokencrypt *tokencrypt.Service
	github     provider.RepositoryProvider
}

// NewService constructs staff service.
func NewService(
	cfg Config,
	users userrepo.Repository,
	projects projectrepo.Repository,
	members projectmemberrepo.Repository,
	repos repocfgrepo.Repository,
	feedback learningfeedbackrepo.Repository,
	runs staffrunrepo.Repository,
	tokencrypt *tokencrypt.Service,
	github provider.RepositoryProvider,
) *Service {
	return &Service{
		cfg:        cfg,
		users:      users,
		projects:   projects,
		members:    members,
		repos:      repos,
		feedback:   feedback,
		runs:       runs,
		tokencrypt: tokencrypt,
		github:     github,
	}
}

// ListProjects returns projects visible to the principal.
func (s *Service) ListProjects(ctx context.Context, principal Principal, limit int) ([]any, error) {
	if principal.IsPlatformAdmin {
		items, err := s.projects.ListAll(ctx, limit)
		if err != nil {
			return nil, err
		}
		out := make([]any, 0, len(items))
		for _, p := range items {
			out = append(out, map[string]any{
				"id":   p.ID,
				"slug": p.Slug,
				"name": p.Name,
				"role": "admin",
			})
		}
		return out, nil
	}

	items, err := s.projects.ListForUser(ctx, principal.UserID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]any, 0, len(items))
	for _, p := range items {
		out = append(out, map[string]any{
			"id":   p.ID,
			"slug": p.Slug,
			"name": p.Name,
			"role": p.Role,
		})
	}
	return out, nil
}

// ListRuns returns runs visible to the principal.
func (s *Service) ListRuns(ctx context.Context, principal Principal, limit int) ([]staffrunrepo.Run, error) {
	if principal.IsPlatformAdmin {
		return s.runs.ListAll(ctx, limit)
	}
	return s.runs.ListForUser(ctx, principal.UserID, limit)
}

// ListRunFlowEvents returns flow events for a run id, enforcing project RBAC.
func (s *Service) ListRunFlowEvents(ctx context.Context, principal Principal, runID string, limit int) ([]staffrunrepo.FlowEvent, error) {
	if runID == "" {
		return nil, errs.Validation{Field: "run_id", Msg: "is required"}
	}

	correlationID, projectID, ok, err := s.runs.GetCorrelationByRunID(ctx, runID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errs.Validation{Field: "run_id", Msg: "not found"}
	}

	if !principal.IsPlatformAdmin {
		if projectID == "" {
			return nil, errs.Forbidden{Msg: "run is not assigned to a project"}
		}
		_, hasRole, err := s.members.GetRole(ctx, projectID, principal.UserID)
		if err != nil {
			return nil, err
		}
		if !hasRole {
			return nil, errs.Forbidden{Msg: "project access required"}
		}
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

	settings := map[string]any{
		"learning_mode_default": s.cfg.LearningModeDefault,
	}
	settingsJSON, err := json.Marshal(settings)
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
func (s *Service) UpsertProjectRepository(ctx context.Context, principal Principal, projectID string, providerID string, owner string, name string, token string, servicesYAMLPath string) (repocfgrepo.RepositoryBinding, error) {
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

	role := "admin"
	if !principal.IsPlatformAdmin {
		r, ok, err := s.members.GetRole(ctx, projectID, principal.UserID)
		if err != nil {
			return repocfgrepo.RepositoryBinding{}, err
		}
		if !ok {
			return repocfgrepo.RepositoryBinding{}, errs.Forbidden{Msg: "project access required"}
		}
		role = r
	}
	if role != "admin" && role != "read_write" {
		return repocfgrepo.RepositoryBinding{}, errs.Forbidden{Msg: "project write access required"}
	}

	if servicesYAMLPath = strings.TrimSpace(servicesYAMLPath); servicesYAMLPath == "" {
		servicesYAMLPath = "services.yaml"
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
			Provider:         string(info.Provider),
			ExternalID:       info.ExternalID,
			Owner:            info.Owner,
			Name:             info.Name,
			TokenEncrypted:   enc,
			ServicesYAMLPath: servicesYAMLPath,
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

	if err := s.repos.Delete(ctx, projectID, repositoryID); err != nil {
		if err == sql.ErrNoRows {
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
		if err == sql.ErrNoRows {
			return errs.Validation{Field: "user_id", Msg: "member not found"}
		}
		return err
	}
	return nil
}

// ListRunLearningFeedback returns feedback entries for a run id.
func (s *Service) ListRunLearningFeedback(ctx context.Context, principal Principal, runID string, limit int) ([]learningfeedbackrepo.Feedback, error) {
	if runID == "" {
		return nil, errs.Validation{Field: "run_id", Msg: "is required"}
	}

	_, projectID, ok, err := s.runs.GetCorrelationByRunID(ctx, runID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errs.Validation{Field: "run_id", Msg: "not found"}
	}

	if !principal.IsPlatformAdmin {
		if projectID == "" {
			return nil, errs.Forbidden{Msg: "run is not assigned to a project"}
		}
		_, hasRole, err := s.members.GetRole(ctx, projectID, principal.UserID)
		if err != nil {
			return nil, err
		}
		if !hasRole {
			return nil, errs.Forbidden{Msg: "project access required"}
		}
	}

	return s.feedback.ListForRun(ctx, runID, limit)
}

// Principal is an authenticated staff identity.
type Principal struct {
	UserID          string
	Email           string
	GitHubLogin     string
	IsPlatformAdmin bool
}
