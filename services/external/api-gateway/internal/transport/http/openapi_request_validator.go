package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
)

type openAPIRequestValidator struct {
	router   routers.Router
	specPath string
}

func newOpenAPIRequestValidator(initCtx context.Context, specPath string) (*openAPIRequestValidator, error) {
	if initCtx == nil {
		return nil, fmt.Errorf("init context is nil")
	}

	resolvedPath, err := resolveOpenAPISpecPath(specPath)
	if err != nil {
		return nil, err
	}

	loader := &openapi3.Loader{
		Context:               initCtx,
		IsExternalRefsAllowed: true,
	}

	doc, err := loader.LoadFromFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("load openapi spec %q: %w", resolvedPath, err)
	}
	if err := doc.Validate(initCtx); err != nil {
		return nil, fmt.Errorf("validate openapi spec %q: %w", resolvedPath, err)
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, fmt.Errorf("build openapi router: %w", err)
	}

	return &openAPIRequestValidator{
		router:   router,
		specPath: resolvedPath,
	}, nil
}

func (v *openAPIRequestValidator) middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			if !shouldValidateOpenAPI(req.URL.Path) {
				return next(c)
			}

			route, pathParams, err := v.router.FindRoute(req)
			if err != nil {
				// Keep legacy handler behavior for requests that router cannot resolve
				// (e.g. non-contract content-type quirks). Echo routing will still return
				// proper 404/405 for unknown endpoints.
				return next(c)
			}

			validationReq, err := cloneRequestWithBufferedBody(req)
			if err != nil {
				return err
			}

			input := &openapi3filter.RequestValidationInput{
				Request:    validationReq,
				PathParams: pathParams,
				Route:      route,
				Options: &openapi3filter.Options{
					// Auth is handled by dedicated middleware; OpenAPI middleware validates
					// only request shape/parameters/body against the contract.
					AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
				},
			}
			if err := openapi3filter.ValidateRequest(req.Context(), input); err != nil {
				return errs.Validation{Field: "request", Msg: err.Error()}
			}

			recorder := newOpenAPIResponseRecorder(c.Response())
			c.SetResponse(recorder)

			if err := next(c); err != nil {
				return err
			}

			if err := validateOpenAPIResponse(req.Context(), c, input, recorder); err != nil {
				return fmt.Errorf("validate openapi response: %w", err)
			}

			return nil
		}
	}
}

func validateOpenAPIResponse(
	ctx context.Context,
	c *echo.Context,
	input *openapi3filter.RequestValidationInput,
	recorder *openAPIResponseRecorder,
) error {
	status := recorder.statusCode
	if status == 0 {
		status = http.StatusOK
	}

	validationResp := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: input,
		Status:                 status,
		Header:                 c.Response().Header().Clone(),
		Options:                input.Options,
	}
	validationResp.SetBodyBytes(recorder.body.Bytes())
	return openapi3filter.ValidateResponse(ctx, validationResp)
}

func cloneRequestWithBufferedBody(req *http.Request) (*http.Request, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if req.Body == nil {
		return req.Clone(req.Context()), nil
	}

	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body for openapi validation: %w", err)
	}
	_ = req.Body.Close()

	req.Body = io.NopCloser(bytes.NewReader(rawBody))
	req.ContentLength = int64(len(rawBody))

	cloned := req.Clone(req.Context())
	cloned.Body = io.NopCloser(bytes.NewReader(rawBody))
	cloned.ContentLength = int64(len(rawBody))

	return cloned, nil
}

func shouldValidateOpenAPI(path string) bool {
	return strings.HasPrefix(path, "/api/")
}

func resolveOpenAPISpecPath(configPath string) (string, error) {
	candidates := make([]string, 0, 3)
	if strings.TrimSpace(configPath) != "" {
		candidates = append(candidates, configPath)
	}
	candidates = append(candidates,
		"/app/api/server/api.yaml",
		"services/external/api-gateway/api/server/api.yaml",
		"api/server/api.yaml",
	)

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("openapi spec not found; checked: %s", strings.Join(candidates, ", "))
}

type openAPIResponseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func newOpenAPIResponseRecorder(writer http.ResponseWriter) *openAPIResponseRecorder {
	return &openAPIResponseRecorder{ResponseWriter: writer}
}

func (r *openAPIResponseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *openAPIResponseRecorder) Write(p []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	_, _ = r.body.Write(p)
	return r.ResponseWriter.Write(p)
}
