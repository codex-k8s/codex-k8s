package http

import (
	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
)

// NOTE(codex-k8s#226): Runtime errors endpoints are intentionally isolated because
// the current staff UI does not use them. Contracts are kept for backward sync
// with OpenAPI and server routing until owner approves deprecation/removal.
// TODO(codex-k8s#81): This endpoint is temporarily unused after staff UI removed
// "platform error" alerts. Decide later whether to keep, repurpose, or remove it.
func (h *staffHandler) ListRuntimeErrors(c *echo.Context) error {
	return listByResolvedResp(c, resolveRuntimeErrorsListFilters(5), h.listRuntimeErrorsCall, casters.RuntimeErrors)
}

// TODO(codex-k8s#81): This endpoint is temporarily unused after staff UI removed
// "platform error" alerts. Decide later whether to keep, repurpose, or remove it.
func (h *staffHandler) MarkRuntimeErrorViewed(c *echo.Context) error {
	return getByPathResp(c, "runtime_error_id", h.markRuntimeErrorViewedCall, casters.RuntimeError)
}
