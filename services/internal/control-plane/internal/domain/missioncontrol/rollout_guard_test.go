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
		ReadPathEnabled:     true,
		RealtimeEnabled:     true,
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

func TestValidateRolloutState_ReadPathRequiresWarmup(t *testing.T) {
	t.Parallel()

	err := ValidateRolloutState(valuetypes.MissionControlRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
		ReadPathEnabled:    true,
	})
	if err == nil {
		t.Fatal("expected read-path validation error, got nil")
	}
}

func TestValidateRolloutState_VoiceRequiresWritePath(t *testing.T) {
	t.Parallel()

	err := ValidateRolloutState(valuetypes.MissionControlRolloutState{
		CoreFeatureEnabled:  true,
		SchemaReady:         true,
		DomainReady:         true,
		WarmupVerified:      true,
		ReadPathEnabled:     true,
		VoiceFeatureEnabled: true,
	})
	if err == nil {
		t.Fatal("expected voice-path validation error, got nil")
	}
}
