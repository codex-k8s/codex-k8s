package postgres

import (
	"context"
	"testing"
	"time"
)

type testContextKey string

func TestNormalizeOpenInput_Defaults(t *testing.T) {
	t.Parallel()

	inputCtx := context.Background()
	ctx, params := normalizeOpenInput(inputCtx, OpenParams{})
	if ctx == nil {
		t.Fatalf("expected non-nil context")
	}
	if ctx != inputCtx {
		t.Fatalf("expected original context to be preserved")
	}
	if params.PingTimeout != defaultPingTimeout {
		t.Fatalf("unexpected default ping timeout: got %s want %s", params.PingTimeout, defaultPingTimeout)
	}
}

func TestNormalizeOpenInput_PreservesValues(t *testing.T) {
	t.Parallel()

	inputCtx := context.WithValue(context.Background(), testContextKey("k"), "v")
	inputParams := OpenParams{PingTimeout: 250 * time.Millisecond}

	ctx, params := normalizeOpenInput(inputCtx, inputParams)
	if ctx != inputCtx {
		t.Fatalf("expected original context to be preserved")
	}
	if params.PingTimeout != inputParams.PingTimeout {
		t.Fatalf("expected ping timeout to be preserved: got %s want %s", params.PingTimeout, inputParams.PingTimeout)
	}
}

func TestBuildDSN(t *testing.T) {
	t.Parallel()

	got := BuildDSN(OpenParams{
		Host:     "localhost",
		Port:     5432,
		DBName:   "codex",
		User:     "user",
		Password: "pass",
		SSLMode:  "disable",
	})

	want := "host=localhost port=5432 dbname=codex user=user password=pass sslmode=disable"
	if got != want {
		t.Fatalf("unexpected dsn: got %q want %q", got, want)
	}
}
