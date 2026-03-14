package missioncontrol

import (
	"testing"

	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

func TestResolveRolloutCapabilities_FullyEnabledCore(t *testing.T) {
	t.Parallel()

	caps, err := ResolveRolloutCapabilities(valuetypes.MissionControlRolloutState{
		CoreFeatureEnabled:  true,
		SchemaReady:         true,
		DomainReady:         true,
		WarmupVerified:      true,
		WritePathEnabled:    true,
		VoiceFeatureEnabled: true,
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

func TestResolveRolloutCapabilities_ReadPathAlwaysAvailableWithSchemaAndDomain(t *testing.T) {
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
	if caps.CanRunWarmup || caps.CanSubmitCoreCommand || caps.CanUseVoicePath {
		t.Fatalf("expected write-side capabilities to stay disabled, got %+v", caps)
	}
}

func TestValidateRolloutState_VoiceRequiresWritePath(t *testing.T) {
	t.Parallel()

	err := ValidateRolloutState(valuetypes.MissionControlRolloutState{
		CoreFeatureEnabled:  true,
		SchemaReady:         true,
		DomainReady:         true,
		WarmupVerified:      true,
		VoiceFeatureEnabled: true,
	})
	if err == nil {
		t.Fatal("expected voice-path validation error, got nil")
	}
}

func TestValidateRolloutState_WritePathRequiresWarmupVerification(t *testing.T) {
	t.Parallel()

	err := ValidateRolloutState(valuetypes.MissionControlRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
		WritePathEnabled:   true,
	})
	if err == nil {
		t.Fatal("expected write-path validation error, got nil")
	}
}
