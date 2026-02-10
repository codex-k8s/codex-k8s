package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"
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
	GetProject(ctx context.Context, principal staff.Principal, projectID string) (projectrepo.Project, error)
	DeleteProject(ctx context.Context, principal staff.Principal, projectID string) error

	ListRuns(ctx context.Context, principal staff.Principal, limit int) ([]staffrun.Run, error)
	GetRun(ctx context.Context, principal staff.Principal, runID string) (staffrun.Run, error)
	ListRunFlowEvents(ctx context.Context, principal staff.Principal, runID string, limit int) ([]staffrun.FlowEvent, error)
	ListRunLearningFeedback(ctx context.Context, principal staff.Principal, runID string, limit int) ([]learningfeedbackrepo.Feedback, error)

	ListUsers(ctx context.Context, principal staff.Principal, limit int) ([]user.User, error)
	CreateAllowedUser(ctx context.Context, principal staff.Principal, email string, isPlatformAdmin bool) (user.User, error)
	DeleteUser(ctx context.Context, principal staff.Principal, userID string) error

	ListProjectMembers(ctx context.Context, principal staff.Principal, projectID string, limit int) ([]projectmember.Member, error)
	UpsertProjectMember(ctx context.Context, principal staff.Principal, projectID string, userID string, role string) error
	UpsertProjectMemberByEmail(ctx context.Context, principal staff.Principal, projectID string, email string, role string) error
	DeleteProjectMember(ctx context.Context, principal staff.Principal, projectID string, userID string) error
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

func requirePrincipal(c *echo.Context) (staff.Principal, error) {
	p, ok := getPrincipal(c)
	if !ok {
		return staff.Principal{}, errs.Unauthorized{Msg: "not authenticated"}
	}
	return p, nil
}

func requirePathParam(c *echo.Context, name string) (string, error) {
	v := strings.TrimSpace(c.Param(name))
	if v == "" {
		return "", errs.Validation{Field: name, Msg: "is required"}
	}
	return v, nil
}

func (h *staffHandler) deleteWith1Param(c *echo.Context, paramName string, fn func(ctx context.Context, principal staff.Principal, id string) error) error {
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	id, err := requirePathParam(c, paramName)
	if err != nil {
		return err
	}
	if err := fn(c.Request().Context(), p, id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *staffHandler) deleteWith2Params(
	c *echo.Context,
	param1 string,
	param2 string,
	fn func(ctx context.Context, principal staff.Principal, id1 string, id2 string) error,
) error {
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	id1, err := requirePathParam(c, param1)
	if err != nil {
		return err
	}
	id2, err := requirePathParam(c, param2)
	if err != nil {
		return err
	}
	if err := fn(c.Request().Context(), p, id1, id2); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
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

func (h *staffHandler) GetProject(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	projectID := c.Param("project_id")
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}

	item, err := h.svc.GetProject(c.Request().Context(), p, projectID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{
		"id":   item.ID,
		"slug": item.Slug,
		"name": item.Name,
	})
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

func (h *staffHandler) DeleteProject(c *echo.Context) error {
	return h.deleteWith1Param(c, "project_id", h.svc.DeleteProject)
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
		createdAt := r.CreatedAt.UTC().Format(time.RFC3339Nano)
		var startedAt any = nil
		if r.StartedAt != nil {
			startedAt = r.StartedAt.UTC().Format(time.RFC3339Nano)
		}
		var finishedAt any = nil
		if r.FinishedAt != nil {
			finishedAt = r.FinishedAt.UTC().Format(time.RFC3339Nano)
		}
		var projectID any = nil
		if r.ProjectID != "" {
			projectID = r.ProjectID
		}
		out = append(out, map[string]any{
			"id":             r.ID,
			"correlation_id": r.CorrelationID,
			"project_id":     projectID,
			"project_slug":   r.ProjectSlug,
			"project_name":   r.ProjectName,
			"status":         r.Status,
			"created_at":     createdAt,
			"started_at":     startedAt,
			"finished_at":    finishedAt,
		})
	}
	return c.JSON(http.StatusOK, map[string]any{"items": out})
}

func (h *staffHandler) GetRun(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	runID := c.Param("run_id")
	if runID == "" {
		return errs.Validation{Field: "run_id", Msg: "is required"}
	}

	r, err := h.svc.GetRun(c.Request().Context(), p, runID)
	if err != nil {
		return err
	}

	createdAt := r.CreatedAt.UTC().Format(time.RFC3339Nano)
	var startedAt any = nil
	if r.StartedAt != nil {
		startedAt = r.StartedAt.UTC().Format(time.RFC3339Nano)
	}
	var finishedAt any = nil
	if r.FinishedAt != nil {
		finishedAt = r.FinishedAt.UTC().Format(time.RFC3339Nano)
	}
	var projectID any = nil
	if r.ProjectID != "" {
		projectID = r.ProjectID
	}

	return c.JSON(http.StatusOK, map[string]any{
		"id":             r.ID,
		"correlation_id": r.CorrelationID,
		"project_id":     projectID,
		"project_slug":   r.ProjectSlug,
		"project_name":   r.ProjectName,
		"status":         r.Status,
		"created_at":     createdAt,
		"started_at":     startedAt,
		"finished_at":    finishedAt,
	})
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
		createdAt := e.CreatedAt.UTC().Format(time.RFC3339Nano)
		out = append(out, map[string]any{
			"correlation_id": e.CorrelationID,
			"event_type":     e.EventType,
			"created_at":     createdAt,
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
			"is_platform_owner": u.IsPlatformOwner,
		})
	}
	return c.JSON(http.StatusOK, map[string]any{"items": out})
}

func (h *staffHandler) DeleteUser(c *echo.Context) error {
	return h.deleteWith1Param(c, "user_id", h.svc.DeleteUser)
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
		"is_platform_owner": u.IsPlatformOwner,
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
		var learningOverride any = nil
		if m.LearningModeOverride != nil {
			learningOverride = *m.LearningModeOverride
		}
		out = append(out, map[string]any{
			"project_id":             m.ProjectID,
			"user_id":                m.UserID,
			"email":                  m.Email,
			"role":                   m.Role,
			"learning_mode_override": learningOverride,
		})
	}
	return c.JSON(http.StatusOK, map[string]any{"items": out})
}

type upsertMemberRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
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
	if strings.TrimSpace(req.Email) != "" && strings.TrimSpace(req.UserID) != "" {
		return errs.Validation{Field: "user_id", Msg: "either user_id or email must be set"}
	}
	if strings.TrimSpace(req.Email) != "" {
		if err := h.svc.UpsertProjectMemberByEmail(c.Request().Context(), p, projectID, req.Email, req.Role); err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	}
	if strings.TrimSpace(req.UserID) == "" {
		return errs.Validation{Field: "user_id", Msg: "is required"}
	}
	if err := h.svc.UpsertProjectMember(c.Request().Context(), p, projectID, req.UserID, req.Role); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *staffHandler) DeleteProjectMember(c *echo.Context) error {
	return h.deleteWith2Params(c, "project_id", "user_id", h.svc.DeleteProjectMember)
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
	return h.deleteWith2Params(c, "project_id", "repository_id", h.svc.DeleteProjectRepository)
}
