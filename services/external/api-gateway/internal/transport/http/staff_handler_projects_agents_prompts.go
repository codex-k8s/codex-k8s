package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

func (h *staffHandler) ListProjects(c *echo.Context) error {
	return listByLimitResp(c, 200, h.listProjectsCall, casters.Projects)
}

func (h *staffHandler) GetProject(c *echo.Context) error {
	return getByPathResp(c, "project_id", h.getProjectCall, casters.Project)
}

func (h *staffHandler) UpsertProject(c *echo.Context) error {
	return createByBodyResp(c, http.StatusCreated, h.upsertProjectCall, casters.Project)
}

func (h *staffHandler) DeleteProject(c *echo.Context) error {
	return h.deleteWith1Param(c, "project_id", h.deleteProject)
}

func (h *staffHandler) ListAgents(c *echo.Context) error {
	return listByLimitResp(c, 200, h.listAgentsCall, casters.Agents)
}

func (h *staffHandler) GetAgent(c *echo.Context) error {
	return getByPathResp(c, "agent_id", h.getAgentCall, casters.Agent)
}

func (h *staffHandler) UpdateAgentSettings(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePath("agent_id"), func(principal *controlplanev1.Principal, agentID string) error {
		var req models.UpdateAgentSettingsRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		item, err := h.cp.Service().UpdateAgentSettings(c.Request().Context(), &controlplanev1.UpdateAgentSettingsRequest{
			Principal:       principal,
			AgentId:         agentID,
			ExpectedVersion: req.ExpectedVersion,
			Settings: &controlplanev1.AgentSettings{
				RuntimeMode:       req.Settings.RuntimeMode,
				TimeoutSeconds:    req.Settings.TimeoutSeconds,
				MaxRetryCount:     req.Settings.MaxRetryCount,
				PromptLocale:      req.Settings.PromptLocale,
				ApprovalsRequired: req.Settings.ApprovalsRequired,
			},
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.Agent(item))
	})
}

func (h *staffHandler) ListPromptTemplateKeys(c *echo.Context) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		limit, err := parseLimit(c, 200)
		if err != nil {
			return err
		}
		resp, err := h.cp.Service().ListPromptTemplateKeys(c.Request().Context(), &controlplanev1.ListPromptTemplateKeysRequest{
			Principal: principal,
			Limit:     int32(limit),
			Scope:     optionalStringPtr(c.QueryParam("scope")),
			ProjectId: optionalStringPtr(c.QueryParam("project_id")),
			Role:      optionalStringPtr(c.QueryParam("role")),
			Kind:      optionalStringPtr(c.QueryParam("kind")),
			Locale:    optionalStringPtr(c.QueryParam("locale")),
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, models.ItemsResponse[models.PromptTemplateKey]{Items: casters.PromptTemplateKeys(resp.GetItems())})
	})
}

func (h *staffHandler) ListPromptTemplateVersions(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePathUnescaped("template_key"), func(principal *controlplanev1.Principal, templateKey string) error {
		limit, err := parseLimit(c, 200)
		if err != nil {
			return err
		}
		resp, err := h.cp.Service().ListPromptTemplateVersions(c.Request().Context(), &controlplanev1.ListPromptTemplateVersionsRequest{
			Principal:   principal,
			TemplateKey: templateKey,
			Limit:       int32(limit),
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, models.ItemsResponse[models.PromptTemplateVersion]{Items: casters.PromptTemplateVersions(resp.GetItems())})
	})
}

func (h *staffHandler) CreatePromptTemplateVersion(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePathUnescaped("template_key"), func(principal *controlplanev1.Principal, templateKey string) error {
		var req models.CreatePromptTemplateVersionRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		item, err := h.createPromptTemplateVersionCall(c.Request().Context(), principal, createPromptTemplateVersionArg{
			templateKey: templateKey,
			req:         req,
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, casters.PromptTemplateVersion(item))
	})
}

func (h *staffHandler) ActivatePromptTemplateVersion(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePathUnescaped("template_key"), func(principal *controlplanev1.Principal, templateKey string) error {
		versionRaw, err := requirePathParam(c, "version")
		if err != nil {
			return err
		}
		version, convErr := strconv.Atoi(versionRaw)
		if convErr != nil || version <= 0 {
			return errs.Validation{Field: "version", Msg: "must be a positive integer"}
		}
		var req models.ActivatePromptTemplateVersionRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		item, err := h.cp.Service().ActivatePromptTemplateVersion(c.Request().Context(), &controlplanev1.ActivatePromptTemplateVersionRequest{
			Principal:       principal,
			TemplateKey:     templateKey,
			Version:         int32(version),
			ExpectedVersion: req.ExpectedVersion,
			ChangeReason:    req.ChangeReason,
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.PromptTemplateVersion(item))
	})
}

func (h *staffHandler) SyncPromptTemplateSeeds(c *echo.Context) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		var req models.PromptTemplateSeedSyncRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		item, err := h.cp.Service().SyncPromptTemplateSeeds(c.Request().Context(), &controlplanev1.PromptTemplateSeedSyncRequest{
			Principal:      principal,
			Mode:           req.Mode,
			Scope:          req.Scope,
			ProjectId:      req.ProjectID,
			IncludeLocales: req.IncludeLocales,
			ForceOverwrite: req.ForceOverwrite,
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.PromptTemplateSeedSyncResponse(item))
	})
}

func (h *staffHandler) PreviewPromptTemplate(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePathUnescaped("template_key"), func(principal *controlplanev1.Principal, templateKey string) error {
		var req models.PreviewPromptTemplateRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		item, err := h.cp.Service().PreviewPromptTemplate(c.Request().Context(), &controlplanev1.PreviewPromptTemplateRequest{
			Principal:   principal,
			TemplateKey: templateKey,
			ProjectId:   req.ProjectID,
			Version:     req.Version,
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.PreviewPromptTemplateResponse(item))
	})
}

func (h *staffHandler) DiffPromptTemplateVersions(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePathUnescaped("template_key"), func(principal *controlplanev1.Principal, templateKey string) error {
		fromVersionRaw := strings.TrimSpace(c.QueryParam("from_version"))
		toVersionRaw := strings.TrimSpace(c.QueryParam("to_version"))
		if fromVersionRaw == "" {
			return errs.Validation{Field: "from_version", Msg: "is required"}
		}
		if toVersionRaw == "" {
			return errs.Validation{Field: "to_version", Msg: "is required"}
		}
		fromVersion, fromErr := strconv.Atoi(fromVersionRaw)
		if fromErr != nil || fromVersion <= 0 {
			return errs.Validation{Field: "from_version", Msg: "must be a positive integer"}
		}
		toVersion, toErr := strconv.Atoi(toVersionRaw)
		if toErr != nil || toVersion <= 0 {
			return errs.Validation{Field: "to_version", Msg: "must be a positive integer"}
		}
		item, err := h.cp.Service().DiffPromptTemplateVersions(c.Request().Context(), &controlplanev1.DiffPromptTemplateVersionsRequest{
			Principal:   principal,
			TemplateKey: templateKey,
			FromVersion: int32(fromVersion),
			ToVersion:   int32(toVersion),
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.PromptTemplateDiffResponse(item))
	})
}

func (h *staffHandler) ListPromptTemplateAuditEvents(c *echo.Context) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		limit, err := parseLimit(c, 200)
		if err != nil {
			return err
		}
		resp, err := h.cp.Service().ListPromptTemplateAuditEvents(c.Request().Context(), &controlplanev1.ListPromptTemplateAuditEventsRequest{
			Principal:   principal,
			Limit:       int32(limit),
			ProjectId:   optionalStringPtr(c.QueryParam("project_id")),
			TemplateKey: optionalStringPtr(c.QueryParam("template_key")),
			ActorId:     optionalStringPtr(c.QueryParam("actor_id")),
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, models.ItemsResponse[models.PromptTemplateAuditEvent]{Items: casters.PromptTemplateAuditEvents(resp.GetItems())})
	})
}
