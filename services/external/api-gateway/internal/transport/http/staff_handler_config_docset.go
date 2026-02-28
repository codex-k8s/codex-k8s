package http

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

func (h *staffHandler) TransitionIssueStageLabel(c *echo.Context) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		var req models.TransitionIssueStageLabelRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}

		repositoryFullName := strings.TrimSpace(req.RepositoryFullName)
		if repositoryFullName == "" {
			return errs.Validation{Field: "repository_full_name", Msg: "is required"}
		}
		issueNumber := int(req.IssueNumber)
		if issueNumber <= 0 {
			return errs.Validation{Field: "issue_number", Msg: "must be a positive integer"}
		}
		targetLabel := strings.TrimSpace(req.TargetLabel)
		if targetLabel == "" {
			return errs.Validation{Field: "target_label", Msg: "is required"}
		}

		resp, err := h.cp.Service().TransitionIssueStageLabel(c.Request().Context(), &controlplanev1.TransitionIssueStageLabelRequest{
			Principal:          principal,
			RepositoryFullName: repositoryFullName,
			IssueNumber:        int32(issueNumber),
			TargetLabel:        targetLabel,
		})
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, casters.TransitionIssueStageLabelResponse(resp))
	})
}

func (h *staffHandler) ListConfigEntries(c *echo.Context) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		scope := strings.TrimSpace(c.QueryParam("scope"))
		if scope == "" {
			return errs.Validation{Field: "scope", Msg: "is required"}
		}

		limit, err := parseLimit(c, 200)
		if err != nil {
			return err
		}

		projectID := strings.TrimSpace(c.QueryParam("project_id"))
		repositoryID := strings.TrimSpace(c.QueryParam("repository_id"))

		resp, err := h.cp.Service().ListConfigEntries(c.Request().Context(), &controlplanev1.ListConfigEntriesRequest{
			Principal:    principal,
			Scope:        scope,
			ProjectId:    optionalStringPtr(projectID),
			RepositoryId: optionalStringPtr(repositoryID),
			Limit:        int32(limit),
		})
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, models.ItemsResponse[models.ConfigEntry]{Items: casters.ConfigEntries(resp.GetItems())})
	})
}

func (h *staffHandler) UpsertConfigEntry(c *echo.Context) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		var req models.UpsertConfigEntryRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}

		var valuePlain *string
		if req.ValuePlain != nil {
			valuePlain = req.ValuePlain
		}
		var valueSecret *string
		if req.ValueSecret != nil {
			valueSecret = req.ValueSecret
		}

		item, err := h.cp.Service().UpsertConfigEntry(c.Request().Context(), &controlplanev1.UpsertConfigEntryRequest{
			Principal:          principal,
			Scope:              req.Scope,
			Kind:               req.Kind,
			ProjectId:          req.ProjectID,
			RepositoryId:       req.RepositoryID,
			Key:                req.Key,
			ValuePlain:         valuePlain,
			ValueSecret:        valueSecret,
			SyncTargets:        req.SyncTargets,
			Mutability:         req.Mutability,
			IsDangerous:        req.IsDangerous,
			DangerousConfirmed: req.DangerousConfirmed,
		})
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, casters.ConfigEntry(item))
	})
}

func (h *staffHandler) DeleteConfigEntry(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePath("config_entry_id"), func(principal *controlplanev1.Principal, id string) error {
		if _, err := h.cp.Service().DeleteConfigEntry(c.Request().Context(), &controlplanev1.DeleteConfigEntryRequest{
			Principal:     principal,
			ConfigEntryId: id,
		}); err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})
}

func (h *staffHandler) ListDocsetGroups(c *echo.Context) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		docsetRef := strings.TrimSpace(c.QueryParam("docset_ref"))
		locale := strings.TrimSpace(c.QueryParam("locale"))
		resp, err := h.cp.Service().ListDocsetGroups(c.Request().Context(), &controlplanev1.ListDocsetGroupsRequest{
			Principal: principal,
			DocsetRef: docsetRef,
			Locale:    locale,
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, models.DocsetGroupItemsResponse{Groups: casters.DocsetGroups(resp.GetGroups())})
	})
}

func (h *staffHandler) ImportDocset(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePath("project_id"), func(principal *controlplanev1.Principal, projectID string) error {
		var req models.ImportDocsetRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		resp, err := h.cp.Service().ImportDocset(c.Request().Context(), &controlplanev1.ImportDocsetRequest{
			Principal:    principal,
			ProjectId:    projectID,
			RepositoryId: strings.TrimSpace(req.RepositoryID),
			DocsetRef:    strings.TrimSpace(req.DocsetRef),
			Locale:       strings.TrimSpace(req.Locale),
			GroupIds:     req.GroupIDs,
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.ImportDocsetResponse(resp))
	})
}

func (h *staffHandler) SyncDocset(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePath("project_id"), func(principal *controlplanev1.Principal, projectID string) error {
		var req models.SyncDocsetRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		resp, err := h.cp.Service().SyncDocset(c.Request().Context(), &controlplanev1.SyncDocsetRequest{
			Principal:    principal,
			ProjectId:    projectID,
			RepositoryId: strings.TrimSpace(req.RepositoryID),
			DocsetRef:    strings.TrimSpace(req.DocsetRef),
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.SyncDocsetResponse(resp))
	})
}
