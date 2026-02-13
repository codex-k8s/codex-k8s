package webhook

import "strings"

var triggerKinds = [...]TriggerKind{
	TriggerKindIntake,
	TriggerKindIntakeRevise,
	TriggerKindVision,
	TriggerKindVisionRevise,
	TriggerKindPRD,
	TriggerKindPRDRevise,
	TriggerKindArch,
	TriggerKindArchRevise,
	TriggerKindDesign,
	TriggerKindDesignRevise,
	TriggerKindPlan,
	TriggerKindPlanRevise,
	TriggerKindDev,
	TriggerKindDevRevise,
	TriggerKindDocAudit,
	TriggerKindQA,
	TriggerKindRelease,
	TriggerKindPostDeploy,
	TriggerKindOps,
	TriggerKindSelfImprove,
	TriggerKindRethink,
}

var defaultLabelByTriggerKind = map[TriggerKind]string{
	TriggerKindIntake:       DefaultRunIntakeLabel,
	TriggerKindIntakeRevise: DefaultRunIntakeReviseLabel,
	TriggerKindVision:       DefaultRunVisionLabel,
	TriggerKindVisionRevise: DefaultRunVisionReviseLabel,
	TriggerKindPRD:          DefaultRunPRDLabel,
	TriggerKindPRDRevise:    DefaultRunPRDReviseLabel,
	TriggerKindArch:         DefaultRunArchLabel,
	TriggerKindArchRevise:   DefaultRunArchReviseLabel,
	TriggerKindDesign:       DefaultRunDesignLabel,
	TriggerKindDesignRevise: DefaultRunDesignReviseLabel,
	TriggerKindPlan:         DefaultRunPlanLabel,
	TriggerKindPlanRevise:   DefaultRunPlanReviseLabel,
	TriggerKindDev:          DefaultRunDevLabel,
	TriggerKindDevRevise:    DefaultRunDevReviseLabel,
	TriggerKindDocAudit:     DefaultRunDocAuditLabel,
	TriggerKindQA:           DefaultRunQALabel,
	TriggerKindRelease:      DefaultRunReleaseLabel,
	TriggerKindPostDeploy:   DefaultRunPostDeployLabel,
	TriggerKindOps:          DefaultRunOpsLabel,
	TriggerKindSelfImprove:  DefaultRunSelfImproveLabel,
	TriggerKindRethink:      DefaultRunRethinkLabel,
}

// NormalizeTriggerKind returns canonical TriggerKind for known values.
// Unknown values are preserved as normalized lower-case tokens to avoid silent rewrites.
func NormalizeTriggerKind(value string) TriggerKind {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return TriggerKindDev
	}
	for _, kind := range triggerKinds {
		if normalized == string(kind) {
			return kind
		}
	}
	return TriggerKind(normalized)
}

// IsKnownTriggerKind reports whether kind is part of canonical trigger-kind catalog.
func IsKnownTriggerKind(kind TriggerKind) bool {
	for _, known := range triggerKinds {
		if kind == known {
			return true
		}
	}
	return false
}

// IsReviseTriggerKind reports whether trigger kind belongs to revise loop.
func IsReviseTriggerKind(kind TriggerKind) bool {
	return strings.HasSuffix(string(kind), "_revise")
}

// DefaultTriggerLabel returns default run:* label for given trigger kind.
func DefaultTriggerLabel(kind TriggerKind) string {
	if label, ok := defaultLabelByTriggerKind[kind]; ok {
		return label
	}
	return DefaultRunDevLabel
}
