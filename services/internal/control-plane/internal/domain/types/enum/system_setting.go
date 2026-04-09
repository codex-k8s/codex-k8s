package enum

import sharedsystemsettings "github.com/codex-k8s/kodex/libs/go/systemsettings"

// SystemSettingKey identifies one platform-owned runtime setting.
type SystemSettingKey string

const (
	SystemSettingKeyGitHubRateLimitWaitEnabled SystemSettingKey = sharedsystemsettings.GitHubRateLimitWaitEnabledKey
	SystemSettingKeyQualityGovernanceEnabled   SystemSettingKey = sharedsystemsettings.QualityGovernanceEnabledKey
)

// SystemSettingSection groups settings by operational area.
type SystemSettingSection string

const (
	SystemSettingSectionGitHubRateLimit   SystemSettingSection = "github_rate_limit"
	SystemSettingSectionQualityGovernance SystemSettingSection = "quality_governance"
)

// SystemSettingValueKind describes the typed contract for one setting value.
type SystemSettingValueKind string

const (
	SystemSettingValueKindBoolean SystemSettingValueKind = "boolean"
)

// SystemSettingReloadSemantics describes when changed values take effect.
type SystemSettingReloadSemantics string

const (
	SystemSettingReloadSemanticsHotReload       SystemSettingReloadSemantics = "hot_reload"
	SystemSettingReloadSemanticsNewRunsOnly     SystemSettingReloadSemantics = "new_runs_only"
	SystemSettingReloadSemanticsRestartRequired SystemSettingReloadSemantics = "restart_required"
)

// SystemSettingVisibility describes whether one setting is operator-visible.
type SystemSettingVisibility string

const (
	SystemSettingVisibilityStaffVisible      SystemSettingVisibility = sharedsystemsettings.VisibilityStaffVisible
	SystemSettingVisibilityInternalOnly      SystemSettingVisibility = sharedsystemsettings.VisibilityInternalOnly
	SystemSettingVisibilitySecretForbiddenWS SystemSettingVisibility = sharedsystemsettings.VisibilitySecretForbiddenWS
)

// SystemSettingSource describes who currently owns the effective value.
type SystemSettingSource string

const (
	SystemSettingSourceDefault SystemSettingSource = "default"
	SystemSettingSourceStaff   SystemSettingSource = "staff"
)

// SystemSettingChangeKind classifies audit trail events.
type SystemSettingChangeKind string

const (
	SystemSettingChangeKindSeeded  SystemSettingChangeKind = "seeded"
	SystemSettingChangeKindUpdated SystemSettingChangeKind = "updated"
	SystemSettingChangeKindReset   SystemSettingChangeKind = "reset"
)
