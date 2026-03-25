package changegovernance

import (
	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/changegovernance"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

type waveSummaryState struct {
	EvidenceCompletenessState enumtypes.ChangeGovernanceEvidenceCompletenessState
	VerificationMinimumState  enumtypes.ChangeGovernanceVerificationMinimumState
}

func activeWaves(items []domainrepo.Wave) []domainrepo.Wave {
	result := make([]domainrepo.Wave, 0, len(items))
	for _, item := range items {
		if item.PublicationState == enumtypes.ChangeGovernanceWavePublicationStateSuperseded {
			continue
		}
		result = append(result, item)
	}
	return result
}

func filterEvidenceBlocksForActiveWaves(waves []domainrepo.Wave, items []domainrepo.EvidenceBlock) []domainrepo.EvidenceBlock {
	activeWaveIDs := make(map[string]struct{}, len(waves))
	for _, wave := range activeWaves(waves) {
		activeWaveIDs[wave.ID] = struct{}{}
	}

	result := make([]domainrepo.EvidenceBlock, 0, len(items))
	for _, item := range items {
		if item.WaveID == "" {
			result = append(result, item)
			continue
		}
		if _, ok := activeWaveIDs[item.WaveID]; ok {
			result = append(result, item)
		}
	}
	return result
}

func deriveWaveSummaryStates(waves []domainrepo.Wave, items []domainrepo.EvidenceBlock) map[string]waveSummaryState {
	evidenceByWaveID := make(map[string][]domainrepo.EvidenceBlock, len(waves))
	for _, item := range items {
		if item.WaveID == "" {
			continue
		}
		evidenceByWaveID[item.WaveID] = append(evidenceByWaveID[item.WaveID], item)
	}

	result := make(map[string]waveSummaryState, len(waves))
	for _, wave := range waves {
		evidenceItems := evidenceByWaveID[wave.ID]
		result[wave.ID] = waveSummaryState{
			EvidenceCompletenessState: deriveEvidenceCompletenessState(evidenceItems),
			VerificationMinimumState:  deriveVerificationMinimumState(evidenceItems),
		}
	}
	return result
}
