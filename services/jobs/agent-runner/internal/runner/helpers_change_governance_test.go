package runner

import (
	"context"

	cpclient "github.com/codex-k8s/codex-k8s/services/jobs/agent-runner/internal/controlplane"
)

func (f *fakeSessionRestoreControlPlane) ReportChangeGovernanceDraftSignal(context.Context, cpclient.ReportChangeGovernanceDraftSignalParams) (cpclient.ReportChangeGovernanceDraftSignalResult, error) {
	return cpclient.ReportChangeGovernanceDraftSignalResult{}, nil
}

func (f *fakeSessionRestoreControlPlane) PublishChangeGovernanceWaveMap(context.Context, cpclient.PublishChangeGovernanceWaveMapParams) (cpclient.PublishChangeGovernanceWaveMapResult, error) {
	return cpclient.PublishChangeGovernanceWaveMapResult{}, nil
}

func (f *fakeSessionRestoreControlPlane) UpsertChangeGovernanceEvidenceSignal(context.Context, cpclient.UpsertChangeGovernanceEvidenceSignalParams) (cpclient.UpsertChangeGovernanceEvidenceSignalResult, error) {
	return cpclient.UpsertChangeGovernanceEvidenceSignalResult{}, nil
}

func (f *fakeOutputRecoveryControlPlane) ReportChangeGovernanceDraftSignal(context.Context, cpclient.ReportChangeGovernanceDraftSignalParams) (cpclient.ReportChangeGovernanceDraftSignalResult, error) {
	return cpclient.ReportChangeGovernanceDraftSignalResult{}, nil
}

func (f *fakeOutputRecoveryControlPlane) PublishChangeGovernanceWaveMap(context.Context, cpclient.PublishChangeGovernanceWaveMapParams) (cpclient.PublishChangeGovernanceWaveMapResult, error) {
	return cpclient.PublishChangeGovernanceWaveMapResult{}, nil
}

func (f *fakeOutputRecoveryControlPlane) UpsertChangeGovernanceEvidenceSignal(context.Context, cpclient.UpsertChangeGovernanceEvidenceSignalParams) (cpclient.UpsertChangeGovernanceEvidenceSignalResult, error) {
	return cpclient.UpsertChangeGovernanceEvidenceSignalResult{}, nil
}

func (f *fakeGitHubRateLimitControlPlane) ReportChangeGovernanceDraftSignal(context.Context, cpclient.ReportChangeGovernanceDraftSignalParams) (cpclient.ReportChangeGovernanceDraftSignalResult, error) {
	return cpclient.ReportChangeGovernanceDraftSignalResult{}, nil
}

func (f *fakeGitHubRateLimitControlPlane) PublishChangeGovernanceWaveMap(context.Context, cpclient.PublishChangeGovernanceWaveMapParams) (cpclient.PublishChangeGovernanceWaveMapResult, error) {
	return cpclient.PublishChangeGovernanceWaveMapResult{}, nil
}

func (f *fakeGitHubRateLimitControlPlane) UpsertChangeGovernanceEvidenceSignal(context.Context, cpclient.UpsertChangeGovernanceEvidenceSignalParams) (cpclient.UpsertChangeGovernanceEvidenceSignalResult, error) {
	return cpclient.UpsertChangeGovernanceEvidenceSignalResult{}, nil
}
