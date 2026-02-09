package staff

import (
	"context"
	"fmt"

	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/errs"
	projectrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/project"
	projectmemberrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/projectmember"
	staffrunrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/staffrun"
	userrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/user"
)

// Service exposes staff-only read/write operations protected by JWT + RBAC.
type Service struct {
	users    userrepo.Repository
	projects projectrepo.Repository
	members  projectmemberrepo.Repository
	runs     staffrunrepo.Repository
}

// NewService constructs staff service.
func NewService(
	users userrepo.Repository,
	projects projectrepo.Repository,
	members projectmemberrepo.Repository,
	runs staffrunrepo.Repository,
) *Service {
	return &Service{
		users:    users,
		projects: projects,
		members:  members,
		runs:     runs,
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

// Principal is an authenticated staff identity.
type Principal struct {
	UserID          string
	Email           string
	GitHubLogin     string
	IsPlatformAdmin bool
}
