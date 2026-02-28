package http

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

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

func bindBody(c *echo.Context, target interface{}) error {
	if err := c.Bind(target); err != nil {
		return errs.Validation{Field: "body", Msg: "invalid JSON"}
	}
	return nil
}

func resolvePath(param string) func(c *echo.Context) (string, error) {
	return func(c *echo.Context) (string, error) {
		return requirePathParam(c, param)
	}
}

func resolvePathUnescaped(param string) func(c *echo.Context) (string, error) {
	return func(c *echo.Context) (string, error) {
		value, err := requirePathParam(c, param)
		if err != nil {
			return "", err
		}
		decoded, decodeErr := url.PathUnescape(value)
		if decodeErr != nil {
			return "", errs.Validation{Field: param, Msg: "must be valid URL-encoded path value"}
		}
		decoded = strings.TrimSpace(decoded)
		if decoded == "" {
			return "", errs.Validation{Field: param, Msg: "is required"}
		}
		return decoded, nil
	}
}

func resolveLimit(defLimit int) func(c *echo.Context) (int, error) {
	return func(c *echo.Context) (int, error) {
		return parseLimit(c, defLimit)
	}
}

func resolvePathLimit(param string, defLimit int) func(c *echo.Context) (pathLimit, error) {
	return func(c *echo.Context) (pathLimit, error) {
		id, err := requirePathParam(c, param)
		if err != nil {
			return pathLimit{}, err
		}
		limit, err := parseLimit(c, defLimit)
		if err != nil {
			return pathLimit{}, err
		}
		return pathLimit{id: id, limit: limit}, nil
	}
}

func resolveRunListFilters(defLimit int, includeWaitState bool) func(c *echo.Context) (runListFilterArg, error) {
	return func(c *echo.Context) (runListFilterArg, error) {
		limit, err := parseLimit(c, defLimit)
		if err != nil {
			return runListFilterArg{}, err
		}
		result := runListFilterArg{
			limit:       int32(limit),
			triggerKind: strings.TrimSpace(c.QueryParam("trigger_kind")),
			status:      strings.TrimSpace(c.QueryParam("status")),
			agentKey:    strings.TrimSpace(c.QueryParam("agent_key")),
		}
		if includeWaitState {
			result.waitState = strings.TrimSpace(c.QueryParam("wait_state"))
		}
		return result, nil
	}
}

func resolveRuntimeDeployListFilters(defLimit int) func(c *echo.Context) (runtimeDeployListArg, error) {
	return func(c *echo.Context) (runtimeDeployListArg, error) {
		limit, err := parseLimit(c, defLimit)
		if err != nil {
			return runtimeDeployListArg{}, err
		}
		return runtimeDeployListArg{
			limit:     int32(limit),
			status:    strings.TrimSpace(c.QueryParam("status")),
			targetEnv: strings.TrimSpace(c.QueryParam("target_env")),
		}, nil
	}
}

func resolveRuntimeErrorsListFilters(defLimit int) func(c *echo.Context) (runtimeErrorsListArg, error) {
	return func(c *echo.Context) (runtimeErrorsListArg, error) {
		limit, err := parseLimit(c, defLimit)
		if err != nil {
			return runtimeErrorsListArg{}, err
		}
		return runtimeErrorsListArg{
			limit:         int32(limit),
			state:         strings.TrimSpace(c.QueryParam("state")),
			level:         strings.TrimSpace(c.QueryParam("level")),
			source:        strings.TrimSpace(c.QueryParam("source")),
			runID:         strings.TrimSpace(c.QueryParam("run_id")),
			correlationID: strings.TrimSpace(c.QueryParam("correlation_id")),
		}, nil
	}
}

func parseOptionalPositiveInt(raw string, field string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(trimmed)
	if err != nil || value <= 0 {
		return 0, errs.Validation{Field: field, Msg: "must be a positive integer"}
	}
	if value > 1000 {
		value = 1000
	}
	return value, nil
}

func parseOptionalBool(raw string, field string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return false, nil
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, errs.Validation{Field: field, Msg: "must be a boolean"}
	}
}

func resolveRegistryImagesListFilters(defLimitRepositories int, defLimitTags int) func(c *echo.Context) (registryImagesListArg, error) {
	return func(c *echo.Context) (registryImagesListArg, error) {
		limitRepositories, err := parseOptionalPositiveInt(c.QueryParam("limit_repositories"), "limit_repositories")
		if err != nil {
			return registryImagesListArg{}, err
		}
		if limitRepositories == 0 {
			limitRepositories = defLimitRepositories
		}

		limitTags, err := parseOptionalPositiveInt(c.QueryParam("limit_tags"), "limit_tags")
		if err != nil {
			return registryImagesListArg{}, err
		}
		if limitTags == 0 {
			limitTags = defLimitTags
		}

		return registryImagesListArg{
			repository:        strings.TrimSpace(c.QueryParam("repository")),
			limitRepositories: int32(limitRepositories),
			limitTags:         int32(limitTags),
		}, nil
	}
}

func resolveRunEventsArg(defLimit int) func(c *echo.Context) (runEventsArg, error) {
	return func(c *echo.Context) (runEventsArg, error) {
		runID, err := requirePathParam(c, "run_id")
		if err != nil {
			return runEventsArg{}, err
		}

		limit, err := parseLimit(c, defLimit)
		if err != nil {
			return runEventsArg{}, err
		}

		includePayload, err := parseOptionalBool(c.QueryParam("include_payload"), "include_payload")
		if err != nil {
			return runEventsArg{}, err
		}

		return runEventsArg{
			runID:          runID,
			limit:          int32(limit),
			includePayload: includePayload,
		}, nil
	}
}

func resolveRunLogsArg(defTailLines int) func(c *echo.Context) (runLogsArg, error) {
	return func(c *echo.Context) (runLogsArg, error) {
		runID, err := requirePathParam(c, "run_id")
		if err != nil {
			return runLogsArg{}, err
		}

		tailLines := defTailLines
		if rawTailLines := strings.TrimSpace(c.QueryParam("tail_lines")); rawTailLines != "" {
			value, convErr := strconv.Atoi(rawTailLines)
			if convErr != nil || value <= 0 {
				return runLogsArg{}, errs.Validation{Field: "tail_lines", Msg: "must be a positive integer"}
			}
			if value > 2000 {
				value = 2000
			}
			tailLines = value
		}

		includeSnapshot, err := parseOptionalBool(c.QueryParam("include_snapshot"), "include_snapshot")
		if err != nil {
			return runLogsArg{}, err
		}

		return runLogsArg{
			runID:           runID,
			tailLines:       int32(tailLines),
			includeSnapshot: includeSnapshot,
		}, nil
	}
}

func withPrincipal(c *echo.Context, fn func(principal *controlplanev1.Principal) error) error {
	principal, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	return fn(principal)
}

func withPrincipalAndResolved[T any](
	c *echo.Context,
	resolve func(c *echo.Context) (T, error),
	fn func(principal *controlplanev1.Principal, value T) error,
) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		value, err := resolve(c)
		if err != nil {
			return err
		}
		return fn(principal, value)
	})
}

func withPrincipalAndTwoPaths(
	c *echo.Context,
	param1 string,
	param2 string,
	fn func(principal *controlplanev1.Principal, id1 string, id2 string) error,
) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		id1, err := requirePathParam(c, param1)
		if err != nil {
			return err
		}
		id2, err := requirePathParam(c, param2)
		if err != nil {
			return err
		}
		return fn(principal, id1, id2)
	})
}

type itemsGetter[Proto any] interface {
	GetItems() []Proto
}

type runItemsGetter interface {
	GetItems() []*controlplanev1.Run
}

type runListCallFn func(ctx context.Context, principal *controlplanev1.Principal, arg runListFilterArg) (runItemsGetter, error)

func listByLimitResp[Proto any, Resp itemsGetter[Proto], Out any](
	c *echo.Context,
	defLimit int,
	call func(ctx context.Context, principal *controlplanev1.Principal, limit int32) (Resp, error),
	cast func(items []Proto) []Out,
) error {
	return withPrincipalAndResolved(c, resolveLimit(defLimit), func(principal *controlplanev1.Principal, limit int) error {
		resp, err := call(c.Request().Context(), principal, int32(limit))
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, models.ItemsResponse[Out]{Items: cast(resp.GetItems())})
	})
}

func listByPathLimitResp[Proto any, Resp itemsGetter[Proto], Out any](
	c *echo.Context,
	param string,
	defLimit int,
	call func(ctx context.Context, principal *controlplanev1.Principal, id string, limit int32) (Resp, error),
	cast func(items []Proto) []Out,
) error {
	return withPrincipalAndResolved(c, resolvePathLimit(param, defLimit), func(principal *controlplanev1.Principal, value pathLimit) error {
		resp, err := call(c.Request().Context(), principal, value.id, int32(value.limit))
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, models.ItemsResponse[Out]{Items: cast(resp.GetItems())})
	})
}

func listByResolvedResp[Arg any, Proto any, Resp itemsGetter[Proto], Out any](
	c *echo.Context,
	resolve func(c *echo.Context) (Arg, error),
	call func(ctx context.Context, principal *controlplanev1.Principal, arg Arg) (Resp, error),
	cast func(items []Proto) []Out,
) error {
	return withPrincipalAndResolved(c, resolve, func(principal *controlplanev1.Principal, arg Arg) error {
		resp, err := call(c.Request().Context(), principal, arg)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, models.ItemsResponse[Out]{Items: cast(resp.GetItems())})
	})
}

func getByPathResp[Proto any, Out any](
	c *echo.Context,
	param string,
	call func(ctx context.Context, principal *controlplanev1.Principal, id string) (Proto, error),
	cast func(item Proto) Out,
) error {
	return withPrincipalAndResolved(c, resolvePath(param), func(principal *controlplanev1.Principal, id string) error {
		item, err := call(c.Request().Context(), principal, id)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, cast(item))
	})
}

func withPrincipalAndResolvedJSON[Req any, Proto any, Out any](
	c *echo.Context,
	resolve func(c *echo.Context) (Req, error),
	call func(ctx context.Context, principal *controlplanev1.Principal, req Req) (Proto, error),
	cast func(item Proto) Out,
) error {
	return withPrincipalAndResolved(c, resolve, func(principal *controlplanev1.Principal, req Req) error {
		item, err := call(c.Request().Context(), principal, req)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, cast(item))
	})
}

func (h *staffHandler) listRunsByFilter(
	c *echo.Context,
	includeWaitState bool,
	call runListCallFn,
) error {
	return withPrincipalAndResolved(c, resolveRunListFilters(200, includeWaitState), func(principal *controlplanev1.Principal, arg runListFilterArg) error {
		resp, err := call(c.Request().Context(), principal, arg)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, models.ItemsResponse[models.Run]{Items: casters.Runs(resp.GetItems())})
	})
}

func createByBodyResp[Req any, Proto any, Out any](
	c *echo.Context,
	statusCode int,
	call func(ctx context.Context, principal *controlplanev1.Principal, req Req) (Proto, error),
	cast func(item Proto) Out,
) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		var req Req
		if err := bindBody(c, &req); err != nil {
			return err
		}
		item, err := call(c.Request().Context(), principal, req)
		if err != nil {
			return err
		}
		return c.JSON(statusCode, cast(item))
	})
}

func (h *staffHandler) deleteWith1Param(c *echo.Context, paramName string, fn func(ctx context.Context, principal *controlplanev1.Principal, id string) error) error {
	return withPrincipalAndResolved(c, resolvePath(paramName), func(principal *controlplanev1.Principal, id string) error {
		if err := fn(c.Request().Context(), principal, id); err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})
}

func (h *staffHandler) deleteWith2Params(
	c *echo.Context,
	param1 string,
	param2 string,
	fn func(ctx context.Context, principal *controlplanev1.Principal, id1 string, id2 string) error,
) error {
	return withPrincipalAndTwoPaths(c, param1, param2, func(principal *controlplanev1.Principal, id1 string, id2 string) error {
		if err := fn(c.Request().Context(), principal, id1, id2); err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})
}
