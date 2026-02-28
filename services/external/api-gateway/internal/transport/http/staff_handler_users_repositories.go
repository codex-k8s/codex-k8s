package http

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

func (h *staffHandler) ListUsers(c *echo.Context) error {
	return listByLimitResp(c, 200, h.listUsersCall, casters.Users)
}

func (h *staffHandler) DeleteUser(c *echo.Context) error {
	return h.deleteWith1Param(c, "user_id", h.deleteUser)
}

func (h *staffHandler) CreateUser(c *echo.Context) error {
	return createByBodyResp(c, http.StatusCreated, h.createUserCall, casters.User)
}

func (h *staffHandler) ListProjectMembers(c *echo.Context) error {
	return listByPathLimitResp(c, "project_id", 200, h.listProjectMembersCall, casters.ProjectMembers)
}

func (h *staffHandler) UpsertProjectMember(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePath("project_id"), func(principal *controlplanev1.Principal, projectID string) error {
		var req models.UpsertProjectMemberRequest
		if err := bindBody(c, &req); err != nil {
			return err
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
			Principal: principal,
			ProjectId: projectID,
			UserId:    optionalStringPtr(userID),
			Email:     optionalStringPtr(email),
			Role:      req.Role,
		}); err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})
}

func (h *staffHandler) DeleteProjectMember(c *echo.Context) error {
	return h.deleteWith2Params(c, "project_id", "user_id", h.deleteProjectMember)
}

func (h *staffHandler) SetProjectMemberLearningModeOverride(c *echo.Context) error {
	return withPrincipalAndTwoPaths(c, "project_id", "user_id", func(principal *controlplanev1.Principal, projectID string, userID string) error {
		var req models.SetProjectMemberLearningModeRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		var enabled *wrapperspb.BoolValue
		if req.Enabled != nil {
			enabled = wrapperspb.Bool(*req.Enabled)
		}
		if _, err := h.cp.Service().SetProjectMemberLearningModeOverride(c.Request().Context(), &controlplanev1.SetProjectMemberLearningModeOverrideRequest{
			Principal: principal,
			ProjectId: projectID,
			UserId:    userID,
			Enabled:   enabled,
		}); err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})
}

func (h *staffHandler) ListProjectRepositories(c *echo.Context) error {
	return listByPathLimitResp(c, "project_id", 200, h.listProjectRepositoriesCall, casters.RepositoryBindings)
}

func (h *staffHandler) UpsertProjectRepository(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePath("project_id"), func(principal *controlplanev1.Principal, projectID string) error {
		var req models.UpsertProjectRepositoryRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		item, err := h.cp.Service().UpsertProjectRepository(c.Request().Context(), &controlplanev1.UpsertProjectRepositoryRequest{
			Principal:        principal,
			ProjectId:        projectID,
			Provider:         req.Provider,
			Owner:            req.Owner,
			Name:             req.Name,
			Token:            req.Token,
			ServicesYamlPath: req.ServicesYAMLPath,
			Alias:            req.Alias,
			Role:             req.Role,
			DefaultRef:       req.DefaultRef,
			DocsRootPath:     optionalStringPtr(req.DocsRootPath),
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, casters.RepositoryBinding(item))
	})
}

func (h *staffHandler) DeleteProjectRepository(c *echo.Context) error {
	return h.deleteWith2Params(c, "project_id", "repository_id", h.deleteProjectRepository)
}

func (h *staffHandler) UpsertRepositoryBotParams(c *echo.Context) error {
	return withPrincipalAndTwoPaths(c, "project_id", "repository_id", func(principal *controlplanev1.Principal, projectID string, repositoryID string) error {
		var req models.UpsertRepositoryBotParamsRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		if _, err := h.cp.Service().UpsertRepositoryBotParams(c.Request().Context(), &controlplanev1.UpsertRepositoryBotParamsRequest{
			Principal:    principal,
			ProjectId:    projectID,
			RepositoryId: repositoryID,
			BotToken:     req.BotToken,
			BotUsername:  req.BotUsername,
			BotEmail:     req.BotEmail,
		}); err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})
}

func (h *staffHandler) RunRepositoryPreflight(c *echo.Context) error {
	return withPrincipalAndTwoPaths(c, "project_id", "repository_id", func(principal *controlplanev1.Principal, projectID string, repositoryID string) error {
		resp, err := h.cp.Service().RunRepositoryPreflight(c.Request().Context(), &controlplanev1.RunRepositoryPreflightRequest{
			Principal:    principal,
			ProjectId:    projectID,
			RepositoryId: repositoryID,
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.RunRepositoryPreflightResponse(resp))
	})
}

func (h *staffHandler) GetProjectGitHubTokens(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePath("project_id"), func(principal *controlplanev1.Principal, projectID string) error {
		item, err := h.cp.Service().GetProjectGitHubTokens(c.Request().Context(), &controlplanev1.GetProjectGitHubTokensRequest{
			Principal: principal,
			ProjectId: projectID,
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.ProjectGitHubTokens(item))
	})
}

func (h *staffHandler) UpsertProjectGitHubTokens(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePath("project_id"), func(principal *controlplanev1.Principal, projectID string) error {
		var req models.UpsertProjectGitHubTokensRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		if _, err := h.cp.Service().UpsertProjectGitHubTokens(c.Request().Context(), &controlplanev1.UpsertProjectGitHubTokensRequest{
			Principal:     principal,
			ProjectId:     projectID,
			PlatformToken: req.PlatformToken,
			BotToken:      req.BotToken,
			BotUsername:   req.BotUsername,
			BotEmail:      req.BotEmail,
		}); err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})
}
