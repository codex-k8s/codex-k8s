package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	"github.com/labstack/echo/v5"
)

func TestResolvePathUnescaped_DecodesTemplateKeyFromEncodedURL(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/staff/prompt-templates/global%2Freviewer%2Fwork%2Fen/versions", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/api/v1/staff/prompt-templates/:template_key/versions")
	ctx.SetPathValues(echo.PathValues{{Name: "template_key", Value: "global%2Freviewer%2Fwork%2Fen"}})

	got, err := resolvePathUnescaped("template_key")(ctx)
	if err != nil {
		t.Fatalf("resolvePathUnescaped returned error: %v", err)
	}
	if got != "global/reviewer/work/en" {
		t.Fatalf("expected decoded template key %q, got %q", "global/reviewer/work/en", got)
	}
}

func TestResolvePathUnescaped_ReturnsValidationErrorForInvalidEscape(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/staff/prompt-templates/template-key/versions", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/api/v1/staff/prompt-templates/:template_key/versions")
	ctx.SetPathValues(echo.PathValues{{Name: "template_key", Value: "global%2Freviewer%2Fwork%2Fen%ZZ"}})

	_, err := resolvePathUnescaped("template_key")(ctx)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validation errs.Validation
	if !errors.As(err, &validation) {
		t.Fatalf("expected errs.Validation, got %T", err)
	}
	if validation.Field != "template_key" {
		t.Fatalf("expected validation field %q, got %q", "template_key", validation.Field)
	}
}
