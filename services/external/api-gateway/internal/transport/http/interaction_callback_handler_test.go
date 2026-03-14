package http

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/controlplane"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

func TestInteractionCallbackHandlerRejectsMissingToken(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/interactions/callback", http.NoBody)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	h := &interactionCallbackHandler{}
	err := h.Callback(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var unauthorized errs.Unauthorized
	if !errors.As(err, &unauthorized) {
		t.Fatalf("expected errs.Unauthorized, got %T", err)
	}
	if unauthorized.Msg != "missing mcp callback token" {
		t.Fatalf("unexpected unauthorized message: %q", unauthorized.Msg)
	}
}

func TestInteractionCallbackHandlerForwardsTypedRequest(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer()
	controlplanev1.RegisterControlPlaneServiceServer(grpcServer, &testInteractionCallbackServer{t: t})
	defer grpcServer.Stop()

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cpClient, err := controlplane.Dial(ctx, listener.Addr().String())
	if err != nil {
		t.Fatalf("dial control-plane: %v", err)
	}
	defer func() {
		_ = cpClient.Close()
	}()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/interactions/callback", strings.NewReader(`{
		"schema_version":"telegram-interaction-v1",
		"interaction_id":"interaction-1",
		"delivery_id":"delivery-1",
		"adapter_event_id":"event-1",
		"callback_kind":"option_selected",
		"occurred_at":"2026-03-13T15:04:05Z",
		"callback_handle":"handle-1",
		"responder_ref":"user-42",
		"provider_message_ref":{"message_id":"42"}
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer callback-token-1")
	rec := httptest.NewRecorder()
	echoCtx := e.NewContext(req, rec)

	h := newInteractionCallbackHandler(cpClient)
	if err := h.Callback(echoCtx); err != nil {
		t.Fatalf("Callback returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200", rec.Code)
	}

	var resp models.InteractionCallbackOutcome
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Accepted {
		t.Fatal("accepted = false, want true")
	}
	if resp.Classification != "applied" {
		t.Fatalf("classification = %q, want applied", resp.Classification)
	}
	if resp.InteractionState != "resolved" {
		t.Fatalf("interaction_state = %q, want resolved", resp.InteractionState)
	}
	if !resp.ResumeRequired {
		t.Fatal("resume_required = false, want true")
	}
	if resp.ContinuationAction != "edit_message" {
		t.Fatalf("continuation_action = %q, want edit_message", resp.ContinuationAction)
	}
}

func TestInteractionCallbackRateLimitMiddlewareLimitsByInteractionID(t *testing.T) {
	t.Parallel()

	e := echo.New()
	middleware := newInteractionCallbackRateLimitMiddleware(1024)
	handler := middleware(func(c *echo.Context) error {
		return c.NoContent(http.StatusAccepted)
	})

	for i := 0; i < interactionCallbackRateLimitBurst; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/interactions/callback", strings.NewReader(`{"interaction_id":"interaction-1"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.RemoteAddr = "192.0.2.10:12345"
		ctx := e.NewContext(req, rec)
		if err := handler(ctx); err != nil {
			t.Fatalf("unexpected error on request %d: %v", i+1, err)
		}
		if rec.Code != http.StatusAccepted {
			t.Fatalf("status code on request %d = %d, want 202", i+1, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/interactions/callback", strings.NewReader(`{"interaction_id":"interaction-1"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.RemoteAddr = "192.0.2.10:12345"
	ctx := e.NewContext(req, rec)
	if err := handler(ctx); err != nil {
		t.Fatalf("unexpected error on rate-limited request: %v", err)
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status code = %d, want 429", rec.Code)
	}
}

type testInteractionCallbackServer struct {
	controlplanev1.UnimplementedControlPlaneServiceServer
	t *testing.T
}

func (s *testInteractionCallbackServer) SubmitInteractionCallback(
	ctx context.Context,
	req *controlplanev1.SubmitInteractionCallbackRequest,
) (*controlplanev1.SubmitInteractionCallbackResponse, error) {
	s.t.Helper()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.t.Fatal("missing metadata")
	}
	authValues := md.Get("authorization")
	if len(authValues) != 1 || authValues[0] != "Bearer callback-token-1" {
		s.t.Fatalf("unexpected authorization metadata: %v", authValues)
	}

	if req.GetInteractionId() != "interaction-1" {
		s.t.Fatalf("interaction_id = %q, want interaction-1", req.GetInteractionId())
	}
	if req.GetDeliveryId() != "delivery-1" {
		s.t.Fatalf("delivery_id = %q, want delivery-1", req.GetDeliveryId())
	}
	if req.GetAdapterEventId() != "event-1" {
		s.t.Fatalf("adapter_event_id = %q, want event-1", req.GetAdapterEventId())
	}
	if req.GetCallbackKind() != "option_selected" {
		s.t.Fatalf("callback_kind = %q, want option_selected", req.GetCallbackKind())
	}
	if req.GetCallbackHandle() != "handle-1" {
		s.t.Fatalf("callback_handle = %q, want handle-1", req.GetCallbackHandle())
	}
	if req.GetResponderRef() != "user-42" {
		s.t.Fatalf("responder_ref = %q, want user-42", req.GetResponderRef())
	}
	if got, want := string(req.GetProviderMessageRefJson()), `{"message_id":"42"}`; got != want {
		s.t.Fatalf("provider_message_ref_json = %q, want %q", got, want)
	}

	return &controlplanev1.SubmitInteractionCallbackResponse{
		Accepted:           true,
		Classification:     "applied",
		InteractionState:   "resolved",
		ResumeRequired:     true,
		ContinuationAction: "edit_message",
	}, nil
}
