package enum

// ChangeGovernanceRiskTier captures explicit package risk classification.
type ChangeGovernanceRiskTier string

const (
	ChangeGovernanceRiskTierLow      ChangeGovernanceRiskTier = "low"
	ChangeGovernanceRiskTierMedium   ChangeGovernanceRiskTier = "medium"
	ChangeGovernanceRiskTierHigh     ChangeGovernanceRiskTier = "high"
	ChangeGovernanceRiskTierCritical ChangeGovernanceRiskTier = "critical"
)

// ChangeGovernanceBundleAdmissibility captures publication discipline for one package.
type ChangeGovernanceBundleAdmissibility string

const (
	ChangeGovernanceBundleAdmissibilitySingleWave             ChangeGovernanceBundleAdmissibility = "single_wave"
	ChangeGovernanceBundleAdmissibilityMechanicalBoundedScope ChangeGovernanceBundleAdmissibility = "mechanical_bounded_scope"
	ChangeGovernanceBundleAdmissibilityRequiresDecomposition  ChangeGovernanceBundleAdmissibility = "requires_decomposition"
)

// ChangeGovernancePublicationState captures package-level publication state.
type ChangeGovernancePublicationState string

const (
	ChangeGovernancePublicationStateHiddenDraft    ChangeGovernancePublicationState = "hidden_draft"
	ChangeGovernancePublicationStateWaveMapDefined ChangeGovernancePublicationState = "wave_map_defined"
	ChangeGovernancePublicationStateWavesPublished ChangeGovernancePublicationState = "waves_published"
	ChangeGovernancePublicationStateReviewReady    ChangeGovernancePublicationState = "review_ready"
	ChangeGovernancePublicationStateReleaseDecided ChangeGovernancePublicationState = "release_decided"
	ChangeGovernancePublicationStateFeedbackOpen   ChangeGovernancePublicationState = "feedback_open"
	ChangeGovernancePublicationStateClosed         ChangeGovernancePublicationState = "closed"
)

// ChangeGovernanceEvidenceCompletenessState captures completeness summary.
type ChangeGovernanceEvidenceCompletenessState string

const (
	ChangeGovernanceEvidenceCompletenessStateNotStarted ChangeGovernanceEvidenceCompletenessState = "not_started"
	ChangeGovernanceEvidenceCompletenessStatePartial    ChangeGovernanceEvidenceCompletenessState = "partial"
	ChangeGovernanceEvidenceCompletenessStateComplete   ChangeGovernanceEvidenceCompletenessState = "complete"
	ChangeGovernanceEvidenceCompletenessStateGapped     ChangeGovernanceEvidenceCompletenessState = "gapped"
	ChangeGovernanceEvidenceCompletenessStateWaived     ChangeGovernanceEvidenceCompletenessState = "waived"
)

// ChangeGovernanceVerificationMinimumState captures verification summary.
type ChangeGovernanceVerificationMinimumState string

const (
	ChangeGovernanceVerificationMinimumStateNotStarted ChangeGovernanceVerificationMinimumState = "not_started"
	ChangeGovernanceVerificationMinimumStateInProgress ChangeGovernanceVerificationMinimumState = "in_progress"
	ChangeGovernanceVerificationMinimumStateMet        ChangeGovernanceVerificationMinimumState = "met"
	ChangeGovernanceVerificationMinimumStateFailed     ChangeGovernanceVerificationMinimumState = "failed"
	ChangeGovernanceVerificationMinimumStateWaived     ChangeGovernanceVerificationMinimumState = "waived"
)

// ChangeGovernanceWaiverState captures waiver summary.
type ChangeGovernanceWaiverState string

const (
	ChangeGovernanceWaiverStateNone      ChangeGovernanceWaiverState = "none"
	ChangeGovernanceWaiverStateRequested ChangeGovernanceWaiverState = "requested"
	ChangeGovernanceWaiverStateApproved  ChangeGovernanceWaiverState = "approved"
	ChangeGovernanceWaiverStateRejected  ChangeGovernanceWaiverState = "rejected"
	ChangeGovernanceWaiverStateExpired   ChangeGovernanceWaiverState = "expired"
)

// ChangeGovernanceReleaseReadinessState captures release gate summary.
type ChangeGovernanceReleaseReadinessState string

const (
	ChangeGovernanceReleaseReadinessStateNotReady           ChangeGovernanceReleaseReadinessState = "not_ready"
	ChangeGovernanceReleaseReadinessStateConditionallyReady ChangeGovernanceReleaseReadinessState = "conditionally_ready"
	ChangeGovernanceReleaseReadinessStateReady              ChangeGovernanceReleaseReadinessState = "ready"
	ChangeGovernanceReleaseReadinessStateBlocked            ChangeGovernanceReleaseReadinessState = "blocked"
	ChangeGovernanceReleaseReadinessStateReleased           ChangeGovernanceReleaseReadinessState = "released"
)

// ChangeGovernanceFeedbackState captures late feedback loop summary.
type ChangeGovernanceFeedbackState string

const (
	ChangeGovernanceFeedbackStateNone         ChangeGovernanceFeedbackState = "none"
	ChangeGovernanceFeedbackStateOpen         ChangeGovernanceFeedbackState = "open"
	ChangeGovernanceFeedbackStateReclassified ChangeGovernanceFeedbackState = "reclassified"
	ChangeGovernanceFeedbackStateClosed       ChangeGovernanceFeedbackState = "closed"
)

// ChangeGovernanceDraftKind captures hidden draft kind.
type ChangeGovernanceDraftKind string

const (
	ChangeGovernanceDraftKindInternalWorkingDraft ChangeGovernanceDraftKind = "internal_working_draft"
)

// ChangeGovernanceDraftState captures hidden draft acknowledgement state.
type ChangeGovernanceDraftState string

const (
	ChangeGovernanceDraftStateHiddenRecorded ChangeGovernanceDraftState = "hidden_recorded"
)

// ChangeGovernanceNextStepKind captures signal-driven next steps.
type ChangeGovernanceNextStepKind string

const (
	ChangeGovernanceNextStepKindWaveMapRequired ChangeGovernanceNextStepKind = "wave_map_required"
	ChangeGovernanceNextStepKindNoOp            ChangeGovernanceNextStepKind = "no_op"
)

// ChangeGovernanceSurfaceKind captures affected surface hint.
type ChangeGovernanceSurfaceKind string

const (
	ChangeGovernanceSurfaceKindDomain        ChangeGovernanceSurfaceKind = "domain"
	ChangeGovernanceSurfaceKindTransport     ChangeGovernanceSurfaceKind = "transport"
	ChangeGovernanceSurfaceKindSchema        ChangeGovernanceSurfaceKind = "schema"
	ChangeGovernanceSurfaceKindReleasePolicy ChangeGovernanceSurfaceKind = "release_policy"
	ChangeGovernanceSurfaceKindUI            ChangeGovernanceSurfaceKind = "ui"
	ChangeGovernanceSurfaceKindDocs          ChangeGovernanceSurfaceKind = "docs"
)

// ChangeGovernanceRiskDriver captures non-authoritative risk hints.
type ChangeGovernanceRiskDriver string

const (
	ChangeGovernanceRiskDriverBlastRadius    ChangeGovernanceRiskDriver = "blast_radius"
	ChangeGovernanceRiskDriverContractData   ChangeGovernanceRiskDriver = "contract_data"
	ChangeGovernanceRiskDriverSecurityPolicy ChangeGovernanceRiskDriver = "security_policy"
	ChangeGovernanceRiskDriverRuntimeRelease ChangeGovernanceRiskDriver = "runtime_release"
)

// ChangeGovernanceDominantIntent captures semantic wave intent.
type ChangeGovernanceDominantIntent string

const (
	ChangeGovernanceDominantIntentCodeBehavior       ChangeGovernanceDominantIntent = "code_behavior"
	ChangeGovernanceDominantIntentSchema             ChangeGovernanceDominantIntent = "schema"
	ChangeGovernanceDominantIntentTransport          ChangeGovernanceDominantIntent = "transport"
	ChangeGovernanceDominantIntentUI                 ChangeGovernanceDominantIntent = "ui"
	ChangeGovernanceDominantIntentOps                ChangeGovernanceDominantIntent = "ops"
	ChangeGovernanceDominantIntentMechanicalRefactor ChangeGovernanceDominantIntent = "mechanical_refactor"
	ChangeGovernanceDominantIntentDocsOnly           ChangeGovernanceDominantIntent = "docs_only"
)

// ChangeGovernanceBoundedScopeKind captures wave boundedness.
type ChangeGovernanceBoundedScopeKind string

const (
	ChangeGovernanceBoundedScopeKindSingleContext          ChangeGovernanceBoundedScopeKind = "single_context"
	ChangeGovernanceBoundedScopeKindCrossContext           ChangeGovernanceBoundedScopeKind = "cross_context"
	ChangeGovernanceBoundedScopeKindMechanicalBoundedScope ChangeGovernanceBoundedScopeKind = "mechanical_bounded_scope"
)

// ChangeGovernanceWavePublicationState captures wave-level lifecycle.
type ChangeGovernanceWavePublicationState string

const (
	ChangeGovernanceWavePublicationStatePlanned    ChangeGovernanceWavePublicationState = "planned"
	ChangeGovernanceWavePublicationStatePublished  ChangeGovernanceWavePublicationState = "published"
	ChangeGovernanceWavePublicationStateReviewed   ChangeGovernanceWavePublicationState = "reviewed"
	ChangeGovernanceWavePublicationStateMerged     ChangeGovernanceWavePublicationState = "merged"
	ChangeGovernanceWavePublicationStateSuperseded ChangeGovernanceWavePublicationState = "superseded"
)

// ChangeGovernanceVerificationTargetKind captures expected verification targets.
type ChangeGovernanceVerificationTargetKind string

const (
	ChangeGovernanceVerificationTargetKindUnit             ChangeGovernanceVerificationTargetKind = "unit"
	ChangeGovernanceVerificationTargetKindIntegration      ChangeGovernanceVerificationTargetKind = "integration"
	ChangeGovernanceVerificationTargetKindContract         ChangeGovernanceVerificationTargetKind = "contract"
	ChangeGovernanceVerificationTargetKindRegression       ChangeGovernanceVerificationTargetKind = "regression"
	ChangeGovernanceVerificationTargetKindRollback         ChangeGovernanceVerificationTargetKind = "rollback"
	ChangeGovernanceVerificationTargetKindReleaseReadiness ChangeGovernanceVerificationTargetKind = "release_readiness"
)

// ChangeGovernanceEvidenceScopeKind captures evidence attachment scope.
type ChangeGovernanceEvidenceScopeKind string

const (
	ChangeGovernanceEvidenceScopeKindPackage ChangeGovernanceEvidenceScopeKind = "package"
	ChangeGovernanceEvidenceScopeKindWave    ChangeGovernanceEvidenceScopeKind = "wave"
)

// ChangeGovernanceEvidenceBlockKind captures evidence family.
type ChangeGovernanceEvidenceBlockKind string

const (
	ChangeGovernanceEvidenceBlockKindIntentContract   ChangeGovernanceEvidenceBlockKind = "intent_contract"
	ChangeGovernanceEvidenceBlockKindVerification     ChangeGovernanceEvidenceBlockKind = "verification"
	ChangeGovernanceEvidenceBlockKindReviewWaiver     ChangeGovernanceEvidenceBlockKind = "review_waiver"
	ChangeGovernanceEvidenceBlockKindReleaseReadiness ChangeGovernanceEvidenceBlockKind = "release_readiness"
	ChangeGovernanceEvidenceBlockKindRuntimeFeedback  ChangeGovernanceEvidenceBlockKind = "runtime_feedback"
)

// ChangeGovernanceEvidenceBlockState captures evidence completeness state.
type ChangeGovernanceEvidenceBlockState string

const (
	ChangeGovernanceEvidenceBlockStateMissing  ChangeGovernanceEvidenceBlockState = "missing"
	ChangeGovernanceEvidenceBlockStatePresent  ChangeGovernanceEvidenceBlockState = "present"
	ChangeGovernanceEvidenceBlockStateVerified ChangeGovernanceEvidenceBlockState = "verified"
	ChangeGovernanceEvidenceBlockStateWaived   ChangeGovernanceEvidenceBlockState = "waived"
	ChangeGovernanceEvidenceBlockStateStale    ChangeGovernanceEvidenceBlockState = "stale"
)

// ChangeGovernanceEvidenceSourceKind captures evidence provenance.
type ChangeGovernanceEvidenceSourceKind string

const (
	ChangeGovernanceEvidenceSourceKindAgentSignal      ChangeGovernanceEvidenceSourceKind = "agent_signal"
	ChangeGovernanceEvidenceSourceKindGitHubWebhook    ChangeGovernanceEvidenceSourceKind = "github_webhook"
	ChangeGovernanceEvidenceSourceKindStaffCommand     ChangeGovernanceEvidenceSourceKind = "staff_command"
	ChangeGovernanceEvidenceSourceKindWorkerFeedback   ChangeGovernanceEvidenceSourceKind = "worker_feedback"
	ChangeGovernanceEvidenceSourceKindBackfillInferred ChangeGovernanceEvidenceSourceKind = "backfill_inferred"
)

// ChangeGovernanceDecisionScopeKind captures decision target scope.
type ChangeGovernanceDecisionScopeKind string

const (
	ChangeGovernanceDecisionScopeKindPackage       ChangeGovernanceDecisionScopeKind = "package"
	ChangeGovernanceDecisionScopeKindWave          ChangeGovernanceDecisionScopeKind = "wave"
	ChangeGovernanceDecisionScopeKindEvidenceBlock ChangeGovernanceDecisionScopeKind = "evidence_block"
)

// ChangeGovernanceDecisionKind captures decision family.
type ChangeGovernanceDecisionKind string

const (
	ChangeGovernanceDecisionKindRiskClassification ChangeGovernanceDecisionKind = "risk_classification"
	ChangeGovernanceDecisionKindReclassification   ChangeGovernanceDecisionKind = "reclassification"
	ChangeGovernanceDecisionKindWaiver             ChangeGovernanceDecisionKind = "waiver"
	ChangeGovernanceDecisionKindReleaseReadiness   ChangeGovernanceDecisionKind = "release_readiness"
)

// ChangeGovernanceDecisionState captures decision lifecycle state.
type ChangeGovernanceDecisionState string

const (
	ChangeGovernanceDecisionStateProposed   ChangeGovernanceDecisionState = "proposed"
	ChangeGovernanceDecisionStateApproved   ChangeGovernanceDecisionState = "approved"
	ChangeGovernanceDecisionStateRejected   ChangeGovernanceDecisionState = "rejected"
	ChangeGovernanceDecisionStateSuperseded ChangeGovernanceDecisionState = "superseded"
)

// ChangeGovernanceDecisionActorKind captures actor role for decisions.
type ChangeGovernanceDecisionActorKind string

const (
	ChangeGovernanceDecisionActorKindOwner    ChangeGovernanceDecisionActorKind = "owner"
	ChangeGovernanceDecisionActorKindReviewer ChangeGovernanceDecisionActorKind = "reviewer"
	ChangeGovernanceDecisionActorKindOperator ChangeGovernanceDecisionActorKind = "operator"
	ChangeGovernanceDecisionActorKindSystem   ChangeGovernanceDecisionActorKind = "system"
)

// ChangeGovernanceFeedbackGapKind captures feedback gap category.
type ChangeGovernanceFeedbackGapKind string

const (
	ChangeGovernanceFeedbackGapKindUnderClassified      ChangeGovernanceFeedbackGapKind = "under_classified"
	ChangeGovernanceFeedbackGapKindMissingEvidence      ChangeGovernanceFeedbackGapKind = "missing_evidence"
	ChangeGovernanceFeedbackGapKindVerificationBypass   ChangeGovernanceFeedbackGapKind = "verification_bypass"
	ChangeGovernanceFeedbackGapKindSilentWaiverAttempt  ChangeGovernanceFeedbackGapKind = "silent_waiver_attempt"
	ChangeGovernanceFeedbackGapKindSemanticMix          ChangeGovernanceFeedbackGapKind = "semantic_mix"
	ChangeGovernanceFeedbackGapKindLateReclassification ChangeGovernanceFeedbackGapKind = "late_reclassification"
)

// ChangeGovernanceFeedbackSourceKind captures feedback origin.
type ChangeGovernanceFeedbackSourceKind string

const (
	ChangeGovernanceFeedbackSourceKindReview      ChangeGovernanceFeedbackSourceKind = "review"
	ChangeGovernanceFeedbackSourceKindRelease     ChangeGovernanceFeedbackSourceKind = "release"
	ChangeGovernanceFeedbackSourceKindPostdeploy  ChangeGovernanceFeedbackSourceKind = "postdeploy"
	ChangeGovernanceFeedbackSourceKindRemediation ChangeGovernanceFeedbackSourceKind = "remediation"
	ChangeGovernanceFeedbackSourceKindWorkerSweep ChangeGovernanceFeedbackSourceKind = "worker_sweep"
	ChangeGovernanceFeedbackSourceKindBackfill    ChangeGovernanceFeedbackSourceKind = "backfill"
)

// ChangeGovernanceFeedbackSeverity captures feedback priority.
type ChangeGovernanceFeedbackSeverity string

const (
	ChangeGovernanceFeedbackSeverityMedium   ChangeGovernanceFeedbackSeverity = "medium"
	ChangeGovernanceFeedbackSeverityHigh     ChangeGovernanceFeedbackSeverity = "high"
	ChangeGovernanceFeedbackSeverityCritical ChangeGovernanceFeedbackSeverity = "critical"
)

// ChangeGovernanceFeedbackRecordState captures feedback lifecycle.
type ChangeGovernanceFeedbackRecordState string

const (
	ChangeGovernanceFeedbackRecordStateOpen         ChangeGovernanceFeedbackRecordState = "open"
	ChangeGovernanceFeedbackRecordStateAcknowledged ChangeGovernanceFeedbackRecordState = "acknowledged"
	ChangeGovernanceFeedbackRecordStateReclassified ChangeGovernanceFeedbackRecordState = "reclassified"
	ChangeGovernanceFeedbackRecordStateClosed       ChangeGovernanceFeedbackRecordState = "closed"
)

// ChangeGovernanceFeedbackSuggestedAction captures typed CTA for one gap.
type ChangeGovernanceFeedbackSuggestedAction string

const (
	ChangeGovernanceFeedbackSuggestedActionReclassify      ChangeGovernanceFeedbackSuggestedAction = "reclassify"
	ChangeGovernanceFeedbackSuggestedActionRequestEvidence ChangeGovernanceFeedbackSuggestedAction = "request_evidence"
	ChangeGovernanceFeedbackSuggestedActionRecordWaiver    ChangeGovernanceFeedbackSuggestedAction = "record_waiver"
	ChangeGovernanceFeedbackSuggestedActionBlockRelease    ChangeGovernanceFeedbackSuggestedAction = "block_release"
	ChangeGovernanceFeedbackSuggestedActionCloseGap        ChangeGovernanceFeedbackSuggestedAction = "close_gap"
)

// ChangeGovernanceProjectionKind captures persisted projection family.
type ChangeGovernanceProjectionKind string

const (
	ChangeGovernanceProjectionKindPackageList         ChangeGovernanceProjectionKind = "package_list"
	ChangeGovernanceProjectionKindPackageDetail       ChangeGovernanceProjectionKind = "package_detail"
	ChangeGovernanceProjectionKindOperatorGapQueue    ChangeGovernanceProjectionKind = "operator_gap_queue"
	ChangeGovernanceProjectionKindReleaseGate         ChangeGovernanceProjectionKind = "release_gate"
	ChangeGovernanceProjectionKindGitHubStatusComment ChangeGovernanceProjectionKind = "github_status_comment"
)

// ChangeGovernanceArtifactKind captures linked artifact family.
type ChangeGovernanceArtifactKind string

const (
	ChangeGovernanceArtifactKindIssue          ChangeGovernanceArtifactKind = "issue"
	ChangeGovernanceArtifactKindPullRequest    ChangeGovernanceArtifactKind = "pull_request"
	ChangeGovernanceArtifactKindRun            ChangeGovernanceArtifactKind = "run"
	ChangeGovernanceArtifactKindAgentSession   ChangeGovernanceArtifactKind = "agent_session"
	ChangeGovernanceArtifactKindDocument       ChangeGovernanceArtifactKind = "document"
	ChangeGovernanceArtifactKindServiceComment ChangeGovernanceArtifactKind = "service_comment"
	ChangeGovernanceArtifactKindReleaseNote    ChangeGovernanceArtifactKind = "release_note"
)

// ChangeGovernanceArtifactRelationKind captures lineage semantics for one artifact link.
type ChangeGovernanceArtifactRelationKind string

const (
	ChangeGovernanceArtifactRelationKindPrimaryContext   ChangeGovernanceArtifactRelationKind = "primary_context"
	ChangeGovernanceArtifactRelationKindEvidenceSource   ChangeGovernanceArtifactRelationKind = "evidence_source"
	ChangeGovernanceArtifactRelationKindDecisionFollowup ChangeGovernanceArtifactRelationKind = "decision_followup"
	ChangeGovernanceArtifactRelationKindFeedbackSource   ChangeGovernanceArtifactRelationKind = "feedback_source"
)
