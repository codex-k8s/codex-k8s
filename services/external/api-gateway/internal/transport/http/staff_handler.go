package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/controlplane"
)

// staffHandler implements staff/private JSON endpoints protected by JWT.
type staffHandler struct {
	cp *controlplane.Client
}

func newStaffHandler(cp *controlplane.Client) *staffHandler {
	return &staffHandler{cp: cp}
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

func requirePrincipal(c *echo.Context) (*controlplanev1.Principal, error) {
	p, ok := getPrincipal(c)
	if !ok || p == nil || strings.TrimSpace(p.UserId) == "" {
		return nil, errs.Unauthorized{Msg: "not authenticated"}
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

func toRFC3339Nano(ts *timestamppb.Timestamp) any {
	if ts == nil {
		return nil
	}
	return ts.AsTime().UTC().Format(time.RFC3339Nano)
}

func (h *staffHandler) deleteWith1Param(c *echo.Context, paramName string, fn func(ctx context.Context, principal *controlplanev1.Principal, id string) error) error {
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
	fn func(ctx context.Context, principal *controlplanev1.Principal, id1 string, id2 string) error,
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
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	limit, err := parseLimit(c, 200)
	if err != nil {
		return err
	}
	resp, err := h.cp.Service().ListProjects(c.Request().Context(), &controlplanev1.ListProjectsRequest{Principal: p, Limit: int32(limit)})
	if err != nil {
		return controlplane.ToDomainError(err)
	}
	out := make([]any, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		out = append(out, map[string]any{
			"id":   it.GetId(),
			"slug": it.GetSlug(),
			"name": it.GetName(),
			"role": it.GetRole(),
		})
	}
	return c.JSON(http.StatusOK, map[string]any{"items": out})
}

func (h *staffHandler) GetProject(c *echo.Context) error {
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	projectID, err := requirePathParam(c, "project_id")
	if err != nil {
		return err
	}
	item, err := h.cp.Service().GetProject(c.Request().Context(), &controlplanev1.GetProjectRequest{Principal: p, ProjectId: projectID})
	if err != nil {
		return controlplane.ToDomainError(err)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"id":   item.GetId(),
		"slug": item.GetSlug(),
		"name": item.GetName(),
	})
}

type upsertProjectRequest struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func (h *staffHandler) UpsertProject(c *echo.Context) error {
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	var req upsertProjectRequest
	if err := c.Bind(&req); err != nil {
		return errs.Validation{Field: "body", Msg: "invalid JSON"}
	}
	item, err := h.cp.Service().UpsertProject(c.Request().Context(), &controlplanev1.UpsertProjectRequest{Principal: p, Slug: req.Slug, Name: req.Name})
	if err != nil {
		return controlplane.ToDomainError(err)
	}
	return c.JSON(http.StatusCreated, map[string]any{
		"id":   item.GetId(),
		"slug": item.GetSlug(),
		"name": item.GetName(),
	})
}

func (h *staffHandler) DeleteProject(c *echo.Context) error {
	return h.deleteWith1Param(c, "project_id", h.deleteProject)
}

func (h *staffHandler) ListRuns(c *echo.Context) error {
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	limit, err := parseLimit(c, 200)
	if err != nil {
		return err
	}
	resp, err := h.cp.Service().ListRuns(c.Request().Context(), &controlplanev1.ListRunsRequest{Principal: p, Limit: int32(limit)})
	if err != nil {
		return controlplane.ToDomainError(err)
	}
	out := make([]any, 0, len(resp.GetItems()))
	for _, r := range resp.GetItems() {
		out = append(out, runToJSON(r))
	}
	return c.JSON(http.StatusOK, map[string]any{"items": out})
}

func (h *staffHandler) GetRun(c *echo.Context) error {
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	runID, err := requirePathParam(c, "run_id")
	if err != nil {
		return err
	}
	r, err := h.cp.Service().GetRun(c.Request().Context(), &controlplanev1.GetRunRequest{Principal: p, RunId: runID})
	if err != nil {
		return controlplane.ToDomainError(err)
	}
	return c.JSON(http.StatusOK, runToJSON(r))
}

func (h *staffHandler) ListRunEvents(c *echo.Context) error {
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	runID, err := requirePathParam(c, "run_id")
	if err != nil {
		return err
	}
	limit, err := parseLimit(c, 500)
	if err != nil {
		return err
	}
	resp, err := h.cp.Service().ListRunEvents(c.Request().Context(), &controlplanev1.ListRunEventsRequest{Principal: p, RunId: runID, Limit: int32(limit)})
	if err != nil {
		return controlplane.ToDomainError(err)
	}
	out := make([]any, 0, len(resp.GetItems()))
	for _, e := range resp.GetItems() {
		out = append(out, map[string]any{
			"correlation_id": e.GetCorrelationId(),
			"event_type":     e.GetEventType(),
			"created_at":     toRFC3339Nano(e.GetCreatedAt()),
			"payload_json":   e.GetPayloadJson(),
		})
	}
	return c.JSON(http.StatusOK, map[string]any{"items": out})
}

func (h *staffHandler) ListRunLearningFeedback(c *echo.Context) error {
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	runID, err := requirePathParam(c, "run_id")
	if err != nil {
		return err
	}
	limit, err := parseLimit(c, 200)
	if err != nil {
		return err
	}
	resp, err := h.cp.Service().ListRunLearningFeedback(c.Request().Context(), &controlplanev1.ListRunLearningFeedbackRequest{Principal: p, RunId: runID, Limit: int32(limit)})
	if err != nil {
		return controlplane.ToDomainError(err)
	}
	out := make([]any, 0, len(resp.GetItems()))
	for _, f := range resp.GetItems() {
		out = append(out, map[string]any{
			"id":            f.GetId(),
			"run_id":        f.GetRunId(),
			"repository_id": f.GetRepositoryId(),
			"pr_number":     f.GetPrNumber(),
			"file_path":     f.GetFilePath(),
			"line":          f.GetLine(),
			"kind":          f.GetKind(),
			"explanation":   f.GetExplanation(),
			"created_at":    toRFC3339Nano(f.GetCreatedAt()),
		})
	}
	return c.JSON(http.StatusOK, map[string]any{"items": out})
}

func (h *staffHandler) ListUsers(c *echo.Context) error {
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	limit, err := parseLimit(c, 200)
	if err != nil {
		return err
	}
	resp, err := h.cp.Service().ListUsers(c.Request().Context(), &controlplanev1.ListUsersRequest{Principal: p, Limit: int32(limit)})
	if err != nil {
		return controlplane.ToDomainError(err)
	}
	out := make([]any, 0, len(resp.GetItems()))
	for _, u := range resp.GetItems() {
		out = append(out, map[string]any{
			"id":                u.GetId(),
			"email":             u.GetEmail(),
			"github_user_id":    u.GetGithubUserId(),
			"github_login":      u.GetGithubLogin(),
			"is_platform_admin": u.GetIsPlatformAdmin(),
			"is_platform_owner": u.GetIsPlatformOwner(),
		})
	}
	return c.JSON(http.StatusOK, map[string]any{"items": out})
}

func (h *staffHandler) DeleteUser(c *echo.Context) error {
	return h.deleteWith1Param(c, "user_id", h.deleteUser)
}

type createUserRequest struct {
	Email           string `json:"email"`
	IsPlatformAdmin bool   `json:"is_platform_admin"`
}

func (h *staffHandler) CreateUser(c *echo.Context) error {
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	var req createUserRequest
	if err := c.Bind(&req); err != nil {
		return errs.Validation{Field: "body", Msg: "invalid JSON"}
	}
	u, err := h.cp.Service().CreateUser(c.Request().Context(), &controlplanev1.CreateUserRequest{Principal: p, Email: req.Email, IsPlatformAdmin: req.IsPlatformAdmin})
	if err != nil {
		return controlplane.ToDomainError(err)
	}
	return c.JSON(http.StatusCreated, map[string]any{
		"id":                u.GetId(),
		"email":             u.GetEmail(),
		"github_user_id":    u.GetGithubUserId(),
		"github_login":      u.GetGithubLogin(),
		"is_platform_admin": u.GetIsPlatformAdmin(),
		"is_platform_owner": u.GetIsPlatformOwner(),
	})
}

func (h *staffHandler) ListProjectMembers(c *echo.Context) error {
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	projectID, err := requirePathParam(c, "project_id")
	if err != nil {
		return err
	}
	limit, err := parseLimit(c, 200)
	if err != nil {
		return err
	}
	resp, err := h.cp.Service().ListProjectMembers(c.Request().Context(), &controlplanev1.ListProjectMembersRequest{Principal: p, ProjectId: projectID, Limit: int32(limit)})
	if err != nil {
		return controlplane.ToDomainError(err)
	}
	out := make([]any, 0, len(resp.GetItems()))
	for _, m := range resp.GetItems() {
		var learningOverride any = nil
		if m.GetLearningModeOverride() != nil {
			learningOverride = m.GetLearningModeOverride().Value
		}
		out = append(out, map[string]any{
			"project_id":             m.GetProjectId(),
			"user_id":                m.GetUserId(),
			"email":                  m.GetEmail(),
			"role":                   m.GetRole(),
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
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	projectID, err := requirePathParam(c, "project_id")
	if err != nil {
		return err
	}
	var req upsertMemberRequest
	if err := c.Bind(&req); err != nil {
		return errs.Validation{Field: "body", Msg: "invalid JSON"}
	}

	email := strings.TrimSpace(req.Email)
	userID := strings.TrimSpace(req.UserID)
	if email != "" && userID != "" {
		return errs.Validation{Field: "user_id", Msg: "either user_id or email must be set"}
	}
	if email == "" && userID == "" {
		return errs.Validation{Field: "user_id", Msg: "is required"}
	}

	if _, err := h.cp.Service().UpsertProjectMember(c.Request().Context(), &controlplanev1.UpsertProjectMemberRequest{
		Principal: p,
		ProjectId: projectID,
		UserId:    userID,
		Email:     email,
		Role:      req.Role,
	}); err != nil {
		return controlplane.ToDomainError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *staffHandler) DeleteProjectMember(c *echo.Context) error {
	return h.deleteWith2Params(c, "project_id", "user_id", h.deleteProjectMember)
}

type setLearningModeRequest struct {
	// Enabled can be true/false or null to inherit project default.
	Enabled *bool `json:"enabled"`
}

func (h *staffHandler) SetProjectMemberLearningModeOverride(c *echo.Context) error {
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	projectID, err := requirePathParam(c, "project_id")
	if err != nil {
		return err
	}
	userID, err := requirePathParam(c, "user_id")
	if err != nil {
		return err
	}
	var req setLearningModeRequest
	if err := c.Bind(&req); err != nil {
		return errs.Validation{Field: "body", Msg: "invalid JSON"}
	}
	var enabled *wrapperspb.BoolValue
	if req.Enabled != nil {
		enabled = wrapperspb.Bool(*req.Enabled)
	}
	if _, err := h.cp.Service().SetProjectMemberLearningModeOverride(c.Request().Context(), &controlplanev1.SetProjectMemberLearningModeOverrideRequest{
		Principal: p,
		ProjectId: projectID,
		UserId:    userID,
		Enabled:   enabled,
	}); err != nil {
		return controlplane.ToDomainError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *staffHandler) ListProjectRepositories(c *echo.Context) error {
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	projectID, err := requirePathParam(c, "project_id")
	if err != nil {
		return err
	}
	limit, err := parseLimit(c, 200)
	if err != nil {
		return err
	}
	resp, err := h.cp.Service().ListProjectRepositories(c.Request().Context(), &controlplanev1.ListProjectRepositoriesRequest{Principal: p, ProjectId: projectID, Limit: int32(limit)})
	if err != nil {
		return controlplane.ToDomainError(err)
	}
	out := make([]any, 0, len(resp.GetItems()))
	for _, r := range resp.GetItems() {
		out = append(out, map[string]any{
			"id":                 r.GetId(),
			"project_id":         r.GetProjectId(),
			"provider":           r.GetProvider(),
			"external_id":        r.GetExternalId(),
			"owner":              r.GetOwner(),
			"name":               r.GetName(),
			"services_yaml_path": r.GetServicesYamlPath(),
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
	p, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	projectID, err := requirePathParam(c, "project_id")
	if err != nil {
		return err
	}
	var req upsertProjectRepositoryRequest
	if err := c.Bind(&req); err != nil {
		return errs.Validation{Field: "body", Msg: "invalid JSON"}
	}
	item, err := h.cp.Service().UpsertProjectRepository(c.Request().Context(), &controlplanev1.UpsertProjectRepositoryRequest{
		Principal:        p,
		ProjectId:        projectID,
		Provider:         req.Provider,
		Owner:            req.Owner,
		Name:             req.Name,
		Token:            req.Token,
		ServicesYamlPath: req.ServicesYAMLPath,
	})
	if err != nil {
		return controlplane.ToDomainError(err)
	}
	return c.JSON(http.StatusCreated, map[string]any{
		"id":                 item.GetId(),
		"project_id":         item.GetProjectId(),
		"provider":           item.GetProvider(),
		"external_id":        item.GetExternalId(),
		"owner":              item.GetOwner(),
		"name":               item.GetName(),
		"services_yaml_path": item.GetServicesYamlPath(),
	})
}

func (h *staffHandler) DeleteProjectRepository(c *echo.Context) error {
	return h.deleteWith2Params(c, "project_id", "repository_id", h.deleteProjectRepository)
}

func runToJSON(r *controlplanev1.Run) map[string]any {
	var projectID any = nil
	if strings.TrimSpace(r.GetProjectId()) != "" {
		projectID = r.GetProjectId()
	}
	return map[string]any{
		"id":             r.GetId(),
		"correlation_id": r.GetCorrelationId(),
		"project_id":     projectID,
		"project_slug":   r.GetProjectSlug(),
		"project_name":   r.GetProjectName(),
		"status":         r.GetStatus(),
		"created_at":     toRFC3339Nano(r.GetCreatedAt()),
		"started_at":     toRFC3339Nano(r.GetStartedAt()),
		"finished_at":    toRFC3339Nano(r.GetFinishedAt()),
	}
}

func (h *staffHandler) deleteProject(ctx context.Context, principal *controlplanev1.Principal, id string) error {
	req := &controlplanev1.DeleteProjectRequest{Principal: principal, ProjectId: id}
	_, err := h.cp.Service().DeleteProject(ctx, req)
	return controlplane.ToDomainError(err)
}

func (h *staffHandler) deleteUser(ctx context.Context, principal *controlplanev1.Principal, id string) error {
	if _, err := h.cp.Service().DeleteUser(ctx, &controlplanev1.DeleteUserRequest{Principal: principal, UserId: id}); err != nil {
		return controlplane.ToDomainError(err)
	}
	return nil
}

func (h *staffHandler) deleteProjectMember(ctx context.Context, principal *controlplanev1.Principal, projectID string, userID string) error {
	svc := h.cp.Service()
	_, err := svc.DeleteProjectMember(ctx, &controlplanev1.DeleteProjectMemberRequest{Principal: principal, ProjectId: projectID, UserId: userID})
	return controlplane.ToDomainError(err)
}

func (h *staffHandler) deleteProjectRepository(ctx context.Context, principal *controlplanev1.Principal, projectID string, repositoryID string) error {
	req := controlplanev1.DeleteProjectRepositoryRequest{Principal: principal, ProjectId: projectID, RepositoryId: repositoryID}
	_, err := h.cp.Service().DeleteProjectRepository(ctx, &req)
	if err != nil {
		return controlplane.ToDomainError(err)
	}
	return nil
}
