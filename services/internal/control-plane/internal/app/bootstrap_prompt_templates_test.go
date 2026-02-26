package app

import (
	"context"
	"errors"
	"reflect"
	"testing"

	staffdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/staff"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type promptTemplateSeedSyncerStub struct {
	result       entitytypes.PromptTemplateSeedSyncResult
	err          error
	calls        int
	gotPrincipal staffdomain.Principal
	gotParams    querytypes.PromptTemplateSeedSyncParams
}

func (s *promptTemplateSeedSyncerStub) SyncPromptTemplateSeeds(_ context.Context, principal staffdomain.Principal, params querytypes.PromptTemplateSeedSyncParams) (entitytypes.PromptTemplateSeedSyncResult, error) {
	s.calls++
	s.gotPrincipal = principal
	s.gotParams = params
	return s.result, s.err
}

func TestSyncBootstrapPromptTemplateSeeds_Success(t *testing.T) {
	stub := &promptTemplateSeedSyncerStub{}

	if err := syncBootstrapPromptTemplateSeeds(context.Background(), stub, "owner-1", nil); err != nil {
		t.Fatalf("syncBootstrapPromptTemplateSeeds returned error: %v", err)
	}
	if stub.calls != 1 {
		t.Fatalf("expected one SyncPromptTemplateSeeds call, got %d", stub.calls)
	}
	if stub.gotPrincipal.UserID != "owner-1" {
		t.Fatalf("expected owner user id %q, got %q", "owner-1", stub.gotPrincipal.UserID)
	}
	if !stub.gotPrincipal.IsPlatformAdmin {
		t.Fatal("expected bootstrap principal to be platform admin")
	}
	if stub.gotParams.Mode != promptTemplateSeedSyncModeApply {
		t.Fatalf("expected sync mode %q, got %q", promptTemplateSeedSyncModeApply, stub.gotParams.Mode)
	}
	if stub.gotParams.Scope != "global" {
		t.Fatalf("expected sync scope %q, got %q", "global", stub.gotParams.Scope)
	}
	if !reflect.DeepEqual(stub.gotParams.IncludeLocales, bootstrapPromptTemplateLocales) {
		t.Fatalf("expected include locales %v, got %v", bootstrapPromptTemplateLocales, stub.gotParams.IncludeLocales)
	}
}

func TestSyncBootstrapPromptTemplateSeeds_EmptyOwnerID(t *testing.T) {
	stub := &promptTemplateSeedSyncerStub{}

	err := syncBootstrapPromptTemplateSeeds(context.Background(), stub, "   ", nil)
	if err == nil {
		t.Fatal("expected error for empty owner user id, got nil")
	}
	if stub.calls != 0 {
		t.Fatalf("expected zero sync calls, got %d", stub.calls)
	}
}

func TestSyncBootstrapPromptTemplateSeeds_PropagatesSyncError(t *testing.T) {
	stub := &promptTemplateSeedSyncerStub{err: errors.New("sync failed")}

	err := syncBootstrapPromptTemplateSeeds(context.Background(), stub, "owner-1", nil)
	if err == nil {
		t.Fatal("expected sync error, got nil")
	}
	if stub.calls != 1 {
		t.Fatalf("expected one SyncPromptTemplateSeeds call, got %d", stub.calls)
	}
}
