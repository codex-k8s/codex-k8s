package http

import (
	"net/http"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/controlplane"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
	"github.com/labstack/echo/v5"
)

type runNamespaceHandler struct {
	cp *controlplane.Client
}

func newRunNamespaceHandler(cp *controlplane.Client) *runNamespaceHandler {
	return &runNamespaceHandler{cp: cp}
}

func (h *runNamespaceHandler) DeleteRunNamespaceByToken(c *echo.Context) error {
	token := strings.TrimSpace(c.Param("token"))
	if token == "" {
		return errs.Validation{Field: "token", Msg: "is required"}
	}

	resp, err := h.cp.DeleteRunNamespaceByToken(c.Request().Context(), token)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, casters.RunNamespaceCleanup(resp))
}
