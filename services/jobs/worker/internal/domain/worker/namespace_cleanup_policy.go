package worker

import (
	"encoding/json"
	"fmt"
	"strings"
)

const defaultRunDebugLabel = "run:debug"

// namespaceCleanupSkipReason is written to flow_events when cleanup is intentionally skipped.
type namespaceCleanupSkipReason string

const (
	namespaceCleanupSkipReasonDisabledByConfig namespaceCleanupSkipReason = "cleanup_disabled_by_config"
	namespaceCleanupSkipReasonDebugLabel       namespaceCleanupSkipReason = "debug_label_present"
	namespaceCleanupSkipReasonSlotHasPeerRun   namespaceCleanupSkipReason = "slot_has_running_peer"
	namespaceCleanupSkipReasonPeerCheckFailed  namespaceCleanupSkipReason = "slot_peer_check_failed"
)

// runDebugPolicy decides whether full-env namespace cleanup should be skipped.
type runDebugPolicy struct {
	SkipCleanup bool
	Reason      namespaceCleanupSkipReason
}

// resolveRunDebugPolicy evaluates cleanup policy from config and issue labels in run payload.
func (s *Service) resolveRunDebugPolicy(runPayload json.RawMessage) runDebugPolicy {
	if !s.cfg.CleanupFullEnvNamespace {
		return runDebugPolicy{
			SkipCleanup: true,
			Reason:      namespaceCleanupSkipReasonDisabledByConfig,
		}
	}

	if hasIssueLabelInRunPayload(runPayload, s.cfg.RunDebugLabel) {
		return runDebugPolicy{
			SkipCleanup: true,
			Reason:      namespaceCleanupSkipReasonDebugLabel,
		}
	}

	return runDebugPolicy{}
}

// buildNamespaceCleanupCommand returns a kubectl command for manual namespace cleanup.
func buildNamespaceCleanupCommand(namespace string) string {
	trimmedNamespace := strings.TrimSpace(namespace)
	if trimmedNamespace == "" {
		return ""
	}
	return fmt.Sprintf("kubectl delete namespace %s", trimmedNamespace)
}
