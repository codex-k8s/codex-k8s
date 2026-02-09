package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/errs"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/projectmember"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/staffrun"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/user"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/staff"
	"github.com/labstack/echo/v5"
)

type staffService interface {
	ListProjects(ctx context.Context, principal staff.Principal, limit int) ([]any, error)
	ListRuns(ctx context.Context, principal staff.Principal, limit int) ([]staffrun.Run, error)
	ListRunFlowEvents(ctx context.Context, principal staff.Principal, runID string, limit int) ([]staffrun.FlowEvent, error)

	ListUsers(ctx context.Context, principal staff.Principal, limit int) ([]user.User, error)
	CreateAllowedUser(ctx context.Context, principal staff.Principal, email string, isPlatformAdmin bool) (user.User, error)

	ListProjectMembers(ctx context.Context, principal staff.Principal, projectID string, limit int) ([]projectmember.Member, error)
	UpsertProjectMember(ctx context.Context, principal staff.Principal, projectID string, userID string, role string) error
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
	return c.JSON(http.StatusOK, map[string]any{"items": items})
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
	return c.JSON(http.StatusOK, map[string]any{"items": items})
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
	return c.JSON(http.StatusOK, map[string]any{"items": items})
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
	return c.JSON(http.StatusCreated, u)
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
	return c.JSON(http.StatusOK, map[string]any{"items": items})
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

