package missioncontrol

import (
	"testing"

	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

func TestResolveRolloutCapabilities_FullyEnabledWithSchemaAndDomain(t *testing.T) {
	t.Parallel()

	caps, err := ResolveRolloutCapabilities(valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	if err != nil {
		t.Fatalf("ResolveRolloutCapabilities() error = %v", err)
	}
	if !caps.CanRunWarmup || !caps.CanServeSnapshot || !caps.CanOpenRealtime || !caps.CanSubmitCoreCommand || !caps.CanUseVoicePath {
		t.Fatalf("unexpected capabilities: %+v", caps)
	}
}

func TestResolveRolloutCapabilities_AllDisabledByDefault(t *testing.T) {
	t.Parallel()

	caps, err := ResolveRolloutCapabilities(valuetypes.MissionControlRolloutState{})
	if err != nil {
		t.Fatalf("ResolveRolloutCapabilities() error = %v", err)
	}
	if caps.CanRunWarmup || caps.CanServeSnapshot || caps.CanOpenRealtime || caps.CanSubmitCoreCommand || caps.CanUseVoicePath {
		t.Fatalf("expected zero-value rollout to keep all capabilities disabled, got %+v", caps)
	}
}

func TestResolveRolloutCapabilities_AllCapabilitiesAvailableWithSchemaAndDomain(t *testing.T) {
	t.Parallel()

	caps, err := ResolveRolloutCapabilities(valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	if err != nil {
		t.Fatalf("ResolveRolloutCapabilities() error = %v", err)
	}
	if !caps.CanServeSnapshot || !caps.CanOpenRealtime {
		t.Fatalf("expected read/realtime path to be available, got %+v", caps)
	}
	if !caps.CanRunWarmup {
		t.Fatalf("expected warmup path to be available, got %+v", caps)
	}
	if !caps.CanSubmitCoreCommand || !caps.CanUseVoicePath {
		t.Fatalf("expected command and voice capabilities to be available, got %+v", caps)
	}
}

func TestValidateRolloutState_AllowsSchemaAndDomain(t *testing.T) {
	t.Parallel()

	err := ValidateRolloutState(valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	if err != nil {
		t.Fatalf("ValidateRolloutState() error = %v", err)
	}
}

func TestValidateRolloutState_DomainRequiresSchema(t *testing.T) {
	t.Parallel()

	err := ValidateRolloutState(valuetypes.MissionControlRolloutState{
		DomainReady: true,
	})
	if err == nil {
		t.Fatal("expected schema validation error, got nil")
	}
}
