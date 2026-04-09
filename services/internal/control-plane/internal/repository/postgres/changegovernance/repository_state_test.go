package changegovernance

import (
	"strings"
	"testing"

	domainrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/changegovernance"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

func TestDerivePackageState_IgnoresSupersededWavesAndEvidence(t *testing.T) {
	t.Parallel()

	aggregate := domainrepo.Aggregate{
		Package: domainrepo.Package{
			ID:                  "pkg-1",
			BundleAdmissibility: enumtypes.ChangeGovernanceBundleAdmissibilityRequiresDecomposition,
			PublicationState:    enumtypes.ChangeGovernancePublicationStateHiddenDraft,
		},
		Waves: []domainrepo.Wave{
			{
				ID:               "wave-active",
				WaveKey:          "docs",
				PublishOrder:     1,
				DominantIntent:   enumtypes.ChangeGovernanceDominantIntentDocsOnly,
				BoundedScopeKind: enumtypes.ChangeGovernanceBoundedScopeKindMechanicalBoundedScope,
				PublicationState: enumtypes.ChangeGovernanceWavePublicationStatePublished,
			},
			{
				ID:               "wave-old",
				WaveKey:          "schema",
				PublishOrder:     2,
				DominantIntent:   enumtypes.ChangeGovernanceDominantIntentSchema,
				BoundedScopeKind: enumtypes.ChangeGovernanceBoundedScopeKindCrossContext,
				PublicationState: enumtypes.ChangeGovernanceWavePublicationStateSuperseded,
			},
		},
		EvidenceBlocks: []domainrepo.EvidenceBlock{
			{
				ID:                "evidence-active",
				PackageID:         "pkg-1",
				WaveID:            "wave-active",
				State:             enumtypes.ChangeGovernanceEvidenceBlockStateVerified,
				VerificationState: enumtypes.ChangeGovernanceVerificationMinimumStateMet,
			},
			{
				ID:                "evidence-old",
				PackageID:         "pkg-1",
				WaveID:            "wave-old",
				State:             enumtypes.ChangeGovernanceEvidenceBlockStateMissing,
				VerificationState: enumtypes.ChangeGovernanceVerificationMinimumStateFailed,
			},
		},
	}

	state := derivePackageState(aggregate, "corr-1", nil)
	if got, want := state.BundleAdmissibility, enumtypes.ChangeGovernanceBundleAdmissibilityMechanicalBoundedScope; got != want {
		t.Fatalf("bundle admissibility = %q, want %q", got, want)
	}
	if got, want := state.PublicationState, enumtypes.ChangeGovernancePublicationStateWavesPublished; got != want {
		t.Fatalf("publication state = %q, want %q", got, want)
	}
	if got, want := state.EvidenceCompletenessState, enumtypes.ChangeGovernanceEvidenceCompletenessStateComplete; got != want {
		t.Fatalf("evidence completeness = %q, want %q", got, want)
	}
	if got, want := state.VerificationMinimumState, enumtypes.ChangeGovernanceVerificationMinimumStateMet; got != want {
		t.Fatalf("verification minimum = %q, want %q", got, want)
	}
}

func TestDeriveWaveSummaryStates_UsesWaveScopedEvidence(t *testing.T) {
	t.Parallel()

	waves := []domainrepo.Wave{
		{ID: "wave-1", WaveKey: "transport"},
		{ID: "wave-2", WaveKey: "docs"},
	}
	items := []domainrepo.EvidenceBlock{
		{
			ID:                "evidence-1",
			WaveID:            "wave-1",
			State:             enumtypes.ChangeGovernanceEvidenceBlockStatePresent,
			VerificationState: enumtypes.ChangeGovernanceVerificationMinimumStateInProgress,
		},
		{
			ID:                "evidence-2",
			WaveID:            "wave-2",
			State:             enumtypes.ChangeGovernanceEvidenceBlockStateVerified,
			VerificationState: enumtypes.ChangeGovernanceVerificationMinimumStateMet,
		},
	}

	summaries := deriveWaveSummaryStates(waves, items)
	if got, want := summaries["wave-1"].EvidenceCompletenessState, enumtypes.ChangeGovernanceEvidenceCompletenessStateComplete; got != want {
		t.Fatalf("wave-1 completeness = %q, want %q", got, want)
	}
	if got, want := summaries["wave-1"].VerificationMinimumState, enumtypes.ChangeGovernanceVerificationMinimumStateInProgress; got != want {
		t.Fatalf("wave-1 verification = %q, want %q", got, want)
	}
	if got, want := summaries["wave-2"].VerificationMinimumState, enumtypes.ChangeGovernanceVerificationMinimumStateMet; got != want {
		t.Fatalf("wave-2 verification = %q, want %q", got, want)
	}
}

func TestChangeGovernanceRevisionQueriesKeepWaveMapConsistent(t *testing.T) {
	t.Parallel()

	required := map[string]string{
		"stage":     "ROW_NUMBER() OVER",
		"supersede": "publication_state = 'superseded'",
		"list":      "CASE WHEN publication_state = 'superseded' THEN 1 ELSE 0 END",
	}
	if !strings.Contains(queryStageWavePublishOrders, required["stage"]) {
		t.Fatalf("stage_wave_publish_orders query must contain %q", required["stage"])
	}
	if !strings.Contains(querySupersedeMissingWaves, required["supersede"]) {
		t.Fatalf("supersede_missing_waves query must contain %q", required["supersede"])
	}
	if !strings.Contains(queryListWaves, required["list"]) {
		t.Fatalf("list_waves query must contain %q", required["list"])
	}
}

func TestInsertDraftIfAbsentQueryClearsLatestAtomically(t *testing.T) {
	t.Parallel()

	required := []string{
		"WITH existing_signal AS",
		"UPDATE change_governance_internal_drafts",
		"SET is_latest = false",
		"NOT EXISTS (SELECT 1 FROM existing_signal)",
	}
	for _, item := range required {
		if !strings.Contains(queryInsertDraftIfAbsent, item) {
			t.Fatalf("insert_draft_if_absent query must contain %q", item)
		}
	}
}
