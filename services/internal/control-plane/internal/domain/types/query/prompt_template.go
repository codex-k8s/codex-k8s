package query

import enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"

// PromptTemplateKey identifies one template stream.
type PromptTemplateKey struct {
	Scope   enumtypes.PromptTemplateScope
	ScopeID string
	Role    string
	Kind    enumtypes.PromptTemplateKind
	Locale  string
}

// PromptTemplateKeyListFilter keeps list filters for template key index.
type PromptTemplateKeyListFilter struct {
	Limit     int
	Scope     string
	ProjectID string
	Role      string
	Kind      string
	Locale    string
}

// PromptTemplateVersionListFilter keeps filters for version list endpoint.
type PromptTemplateVersionListFilter struct {
	Key   PromptTemplateKey
	Limit int
}

// PromptTemplateVersionCreateParams describes one create-version request.
type PromptTemplateVersionCreateParams struct {
	Key             PromptTemplateKey
	BodyMarkdown    string
	ExpectedVersion int
	Source          enumtypes.PromptTemplateSource
	ChangeReason    string
	UpdatedByUserID string
}

// PromptTemplateVersionActivateParams describes one activate-version request.
type PromptTemplateVersionActivateParams struct {
	Key             PromptTemplateKey
	Version         int
	ExpectedVersion int
	ChangeReason    string
	UpdatedByUserID string
}

// PromptTemplateVersionLookup resolves one specific version.
type PromptTemplateVersionLookup struct {
	Key     PromptTemplateKey
	Version int
}

// PromptTemplatePreviewLookup resolves preview target.
type PromptTemplatePreviewLookup struct {
	Key     PromptTemplateKey
	Version int
}

// PromptTemplateAuditListFilter keeps filters for prompt template audit endpoint.
type PromptTemplateAuditListFilter struct {
	Limit       int
	ProjectID   string
	TemplateKey string
	ActorID     string
}

// PromptTemplateSeedSyncParams describes seed sync execution mode.
type PromptTemplateSeedSyncParams struct {
	Mode            string
	Scope           string
	ProjectID       string
	IncludeLocales  []string
	ForceOverwrite  bool
	UpdatedByUserID string
}

// PromptTemplateSeedCreateParams describes create-if-missing seed upsert.
type PromptTemplateSeedCreateParams struct {
	Key             PromptTemplateKey
	BodyMarkdown    string
	UpdatedByUserID string
}
