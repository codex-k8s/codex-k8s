package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/errs"
	learningfeedbackrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/learningfeedback"
	projectrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/project"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/projectmember"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/repocfg"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/staffrun"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/user"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/staff"
	"github.com/labstack/echo/v5"
)

type staffService interface {
	ListProjects(ctx context.Context, principal staff.Principal, limit int) ([]any, error)
	UpsertProject(ctx context.Context, principal staff.Principal, slug string, name string) (projectrepo.Project, error)

	ListRuns(ctx context.Context, principal staff.Principal, limit int) ([]staffrun.Run, error)
	ListRunFlowEvents(ctx context.Context, principal staff.Principal, runID string, limit int) ([]staffrun.FlowEvent, error)
	ListRunLearningFeedback(ctx context.Context, principal staff.Principal, runID string, limit int) ([]learningfeedbackrepo.Feedback, error)

	ListUsers(ctx context.Context, principal staff.Principal, limit int) ([]user.User, error)
	CreateAllowedUser(ctx context.Context, principal staff.Principal, email string, isPlatformAdmin bool) (user.User, error)

	ListProjectMembers(ctx context.Context, principal staff.Principal, projectID string, limit int) ([]projectmember.Member, error)
	UpsertProjectMember(ctx context.Context, principal staff.Principal, projectID string, userID string, role string) error
	SetProjectMemberLearningModeOverride(ctx context.Context, principal staff.Principal, projectID string, userID string, enabled *bool) error

	ListProjectRepositories(ctx context.Context, principal staff.Principal, projectID string, limit int) ([]repocfgrepo.RepositoryBinding, error)
	UpsertProjectRepository(ctx context.Context, principal staff.Principal, projectID string, provider string, owner string, name string, token string, servicesYAMLPath string) (repocfgrepo.RepositoryBinding, error)
	DeleteProjectRepository(ctx context.Context, principal staff.Principal, projectID string, repositoryID string) error
}

// staffHandler implements staff/private JSON endpoints protected by JWT.
type staffHandler struct {
	svc staffService
}

func newStaffHandler(svc staffService) *staffHandler {
	return &staffHandler{svc: svc}
}

func parseLimit(c *echo.Context, def int) (int, error) {
	limitStr := c.QueryParam("limit")
	if limitStr == "" {
		return def, nil
	}
	n, err := strconv.Atoi(limitStr)
	if err != nil || n <= 0 {
		return 0, errs.Validation{Field: "limit", Msg: "must be a positive integer"}
	}
	if n > 1000 {
		n = 1000
	}
	return n, nil
}

func (h *staffHandler) ListProjects(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	limit, err := parseLimit(c, 200)
	if err != nil {
		return err
	}
	items, err := h.svc.ListProjects(c.Request().Context(), p, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

type upsertProjectRequest struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func (h *staffHandler) UpsertProject(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	var req upsertProjectRequest
	if err := c.Bind(&req); err != nil {
		return errs.Validation{Field: "body", Msg: "invalid JSON"}
	}
	item, err := h.svc.UpsertProject(c.Request().Context(), p, req.Slug, req.Name)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]any{
		"id":   item.ID,
		"slug": item.Slug,
		"name": item.Name,
	})
}

func (h *staffHandler) ListRuns(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	limit, err := parseLimit(c, 200)
	if err != nil {
		return err
	}
	items, err := h.svc.ListRuns(c.Request().Context(), p, limit)
	if err != nil {
		return err
	}
	out := make([]any, 0, len(items))
	for _, r := range items {
		out = append(out, map[string]any{
			"id":             r.ID,
			"correlation_id": r.CorrelationID,
			"project_id":     r.ProjectID,
			"status":         r.Status,
			"created_at":     r.CreatedAt,
			"started_at":     r.StartedAt,
			"finished_at":    r.FinishedAt,
		})
	}
	return c.JSON(http.StatusOK, map[string]any{"items": out})
}

func (h *staffHandler) ListRunEvents(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	runID := c.Param("run_id")
	if runID == "" {
		return errs.Validation{Field: "run_id", Msg: "is required"}
	}
	limit, err := parseLimit(c, 500)
	if err != nil {
		return err
	}
	items, err := h.svc.ListRunFlowEvents(c.Request().Context(), p, runID, limit)
	if err != nil {
		return err
	}
	out := make([]any, 0, len(items))
	for _, e := range items {
		out = append(out, map[string]any{
			"correlation_id": e.CorrelationID,
			"event_type":     e.EventType,
			"created_at":     e.CreatedAt,
			"payload_json":   string(e.PayloadJSON),
		})
	}
	return c.JSON(http.StatusOK, map[string]any{"items": out})
}

func (h *staffHandler) ListRunLearningFeedback(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	runID := c.Param("run_id")
	if runID == "" {
		return errs.Validation{Field: "run_id", Msg: "is required"}
	}
	limit, err := parseLimit(c, 200)
	if err != nil {
		return err
	}
	items, err := h.svc.ListRunLearningFeedback(c.Request().Context(), p, runID, limit)
	if err != nil {
		return err
	}
	out := make([]any, 0, len(items))
	for _, f := range items {
		createdAt := f.CreatedAt.UTC().Format(time.RFC3339Nano)
		out = append(out, map[string]any{
			"id":            f.ID,
			"run_id":        f.RunID,
			"repository_id": f.RepositoryID,
			"pr_number":     f.PRNumber,
			"file_path":     f.FilePath,
			"line":          f.Line,
			"kind":          f.Kind,
			"explanation":   f.Explanation,
			"created_at":    createdAt,
		})
	}
	return c.JSON(http.StatusOK, map[string]any{"items": out})
}

func (h *staffHandler) ListUsers(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	limit, err := parseLimit(c, 200)
	if err != nil {
		return err
	}
	items, err := h.svc.ListUsers(c.Request().Context(), p, limit)
	if err != nil {
		return err
	}
	out := make([]any, 0, len(items))
	for _, u := range items {
		out = append(out, map[string]any{
			"id":                u.ID,
			"email":             u.Email,
			"github_user_id":    u.GitHubUserID,
			"github_login":      u.GitHubLogin,
			"is_platform_admin": u.IsPlatformAdmin,
		})
	}
	return c.JSON(http.StatusOK, map[string]any{"items": out})
}

type createUserRequest struct {
	Email           string `json:"email"`
	IsPlatformAdmin bool   `json:"is_platform_admin"`
}

func (h *staffHandler) CreateUser(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	var req createUserRequest
	if err := c.Bind(&req); err != nil {
		return errs.Validation{Field: "body", Msg: "invalid JSON"}
	}
	u, err := h.svc.CreateAllowedUser(c.Request().Context(), p, req.Email, req.IsPlatformAdmin)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]any{
		"id":                u.ID,
		"email":             u.Email,
		"github_user_id":    u.GitHubUserID,
		"github_login":      u.GitHubLogin,
		"is_platform_admin": u.IsPlatformAdmin,
	})
}

func (h *staffHandler) ListProjectMembers(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	projectID := c.Param("project_id")
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	limit, err := parseLimit(c, 200)
	if err != nil {
		return err
	}
	items, err := h.svc.ListProjectMembers(c.Request().Context(), p, projectID, limit)
	if err != nil {
		return err
	}
	out := make([]any, 0, len(items))
	for _, m := range items {
		out = append(out, map[string]any{
			"project_id": m.ProjectID,
			"user_id":    m.UserID,
			"email":      m.Email,
			"role":       m.Role,
		})
	}
	return c.JSON(http.StatusOK, map[string]any{"items": out})
}

type upsertMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func (h *staffHandler) UpsertProjectMember(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	projectID := c.Param("project_id")
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	var req upsertMemberRequest
	if err := c.Bind(&req); err != nil {
		return errs.Validation{Field: "body", Msg: "invalid JSON"}
	}
	if err := h.svc.UpsertProjectMember(c.Request().Context(), p, projectID, req.UserID, req.Role); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

type setLearningModeRequest struct {
	// Enabled can be true/false or null to inherit project default.
	Enabled *bool `json:"enabled"`
}

func (h *staffHandler) SetProjectMemberLearningModeOverride(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	projectID := c.Param("project_id")
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	userID := c.Param("user_id")
	if userID == "" {
		return errs.Validation{Field: "user_id", Msg: "is required"}
	}
	var req setLearningModeRequest
	if err := c.Bind(&req); err != nil {
		return errs.Validation{Field: "body", Msg: "invalid JSON"}
	}
	if err := h.svc.SetProjectMemberLearningModeOverride(c.Request().Context(), p, projectID, userID, req.Enabled); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *staffHandler) ListProjectRepositories(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	projectID := c.Param("project_id")
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	limit, err := parseLimit(c, 200)
	if err != nil {
		return err
	}
	items, err := h.svc.ListProjectRepositories(c.Request().Context(), p, projectID, limit)
	if err != nil {
		return err
	}
	out := make([]any, 0, len(items))
	for _, r := range items {
		out = append(out, map[string]any{
			"id":                 r.ID,
			"project_id":         r.ProjectID,
			"provider":           r.Provider,
			"external_id":        r.ExternalID,
			"owner":              r.Owner,
			"name":               r.Name,
			"services_yaml_path": r.ServicesYAMLPath,
		})
	}
	return c.JSON(http.StatusOK, map[string]any{"items": out})
}

type upsertProjectRepositoryRequest struct {
	Provider         string `json:"provider"`
	Owner            string `json:"owner"`
	Name             string `json:"name"`
	Token            string `json:"token"`
	ServicesYAMLPath string `json:"services_yaml_path"`
}

func (h *staffHandler) UpsertProjectRepository(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	projectID := c.Param("project_id")
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	var req upsertProjectRepositoryRequest
	if err := c.Bind(&req); err != nil {
		return errs.Validation{Field: "body", Msg: "invalid JSON"}
	}
	item, err := h.svc.UpsertProjectRepository(
		c.Request().Context(),
		p,
		projectID,
		req.Provider,
		req.Owner,
		req.Name,
		req.Token,
		req.ServicesYAMLPath,
	)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]any{
		"id":                 item.ID,
		"project_id":         item.ProjectID,
		"provider":           item.Provider,
		"external_id":        item.ExternalID,
		"owner":              item.Owner,
		"name":               item.Name,
		"services_yaml_path": item.ServicesYAMLPath,
	})
}

func (h *staffHandler) DeleteProjectRepository(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	projectID := c.Param("project_id")
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	repositoryID := c.Param("repository_id")
	if repositoryID == "" {
		return errs.Validation{Field: "repository_id", Msg: "is required"}
	}
	if err := h.svc.DeleteProjectRepository(c.Request().Context(), p, projectID, repositoryID); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
