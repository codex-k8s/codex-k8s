package joblauncher

const (
	metadataLabelManagedBy          = "codex-k8s.dev/managed-by"
	metadataLabelNamespacePurpose   = "codex-k8s.dev/namespace-purpose"
	metadataLabelRuntimeMode        = "codex-k8s.dev/runtime-mode"
	metadataLabelProjectID          = "codex-k8s.dev/project-id"
	metadataLabelRunID              = "codex-k8s.dev/run-id"
	metadataLabelIssueNumber        = "codex-k8s.dev/issue-number"
	metadataLabelAgentKey           = "codex-k8s.dev/agent-key"
	metadataAnnotationCorrelationID = "codex-k8s.dev/correlation-id"
	metadataAnnotationNamespaceTTL  = "codex-k8s.dev/namespace-lease-ttl"
	metadataAnnotationNamespaceExp  = "codex-k8s.dev/namespace-lease-expires-at"
	metadataAnnotationNamespaceUpd  = "codex-k8s.dev/namespace-lease-updated-at"
)
