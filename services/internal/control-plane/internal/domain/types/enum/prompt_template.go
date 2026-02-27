package enum

// PromptTemplateScope defines template ownership scope.
type PromptTemplateScope string

const (
	PromptTemplateScopeGlobal  PromptTemplateScope = "global"
	PromptTemplateScopeProject PromptTemplateScope = "project"
)

// PromptTemplateKind defines supported template variants.
type PromptTemplateKind string

const (
	PromptTemplateKindWork   PromptTemplateKind = "work"
	PromptTemplateKindRevise PromptTemplateKind = "revise"
)

// PromptTemplateStatus defines lifecycle state of one stored version.
type PromptTemplateStatus string

const (
	PromptTemplateStatusDraft    PromptTemplateStatus = "draft"
	PromptTemplateStatusActive   PromptTemplateStatus = "active"
	PromptTemplateStatusArchived PromptTemplateStatus = "archived"
)

// PromptTemplateSource defines provenance of stored prompt content.
type PromptTemplateSource string

const (
	PromptTemplateSourceProjectOverride PromptTemplateSource = "project_override"
	PromptTemplateSourceGlobalOverride  PromptTemplateSource = "global_override"
	PromptTemplateSourceRepoSeed        PromptTemplateSource = "repo_seed"
)
