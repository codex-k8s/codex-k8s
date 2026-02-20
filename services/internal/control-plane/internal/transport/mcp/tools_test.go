package mcp

import (
	"context"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
)

type testToolInput struct {
	Status string `json:"status"`
}

type testToolOutput struct {
	OK bool `json:"ok"`
}

func TestAddTool_DoesNotPanicOnNilInputSchema(t *testing.T) {
	t.Parallel()

	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	run := func(_ context.Context, _ mcpdomain.SessionContext, _ testToolInput) (testToolOutput, error) {
		return testToolOutput{OK: true}, nil
	}

	mustNotPanic(t, func() {
		addTool(server, mcpdomain.ToolName("test.tool.nil-schema"), "test", run)
	})
}

func TestAddToolWithInputSchema_DoesNotPanicOnTypedNilSchema(t *testing.T) {
	t.Parallel()

	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	run := func(_ context.Context, _ mcpdomain.SessionContext, _ testToolInput) (testToolOutput, error) {
		return testToolOutput{OK: true}, nil
	}

	var typedNilSchema *jsonschema.Schema
	mustNotPanic(t, func() {
		addToolWithInputSchema(server, mcpdomain.ToolName("test.tool.typed-nil-schema"), "test", typedNilSchema, run)
	})
}

func mustNotPanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()
	fn()
}
