package http

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/kodex/services/external/api-gateway/internal/transport/http/casters"
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
