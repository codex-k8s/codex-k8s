package githubratelimitwait

import (
	domainrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/githubratelimitwait"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	"github.com/codex-k8s/kodex/services/internal/control-plane/internal/repository/postgres/githubratelimitwait/dbmodel"
)

func waitFromDBModel(row dbmodel.WaitRow) domainrepo.Wait {
	item := domainrepo.Wait{
		ID:                     row.ID,
		ProjectID:              row.ProjectID,
		RunID:                  row.RunID,
		ContourKind:            enumtypes.GitHubRateLimitContourKind(row.ContourKind),
		SignalOrigin:           enumtypes.GitHubRateLimitSignalOrigin(row.SignalOrigin),
		OperationClass:         enumtypes.GitHubRateLimitOperationClass(row.OperationClass),
		State:                  enumtypes.GitHubRateLimitWaitState(row.State),
		LimitKind:              enumtypes.GitHubRateLimitLimitKind(row.LimitKind),
		Confidence:             enumtypes.GitHubRateLimitConfidence(row.Confidence),
		RecoveryHintKind:       enumtypes.GitHubRateLimitRecoveryHintKind(row.RecoveryHintKind),
		DominantForRun:         row.DominantForRun,
		SignalID:               row.SignalID,
		CorrelationID:          row.CorrelationID,
		ResumeActionKind:       enumtypes.GitHubRateLimitResumeActionKind(row.ResumeActionKind),
		ResumePayloadJSON:      row.ResumePayloadJSON,
		AutoResumeAttemptsUsed: int(row.AutoResumeAttemptsUsed),
		MaxAutoResumeAttempts:  int(row.MaxAutoResumeAttempts),
		FirstDetectedAt:        row.FirstDetectedAt.Time,
		LastSignalAt:           row.LastSignalAt.Time,
		CreatedAt:              row.CreatedAt.Time,
		UpdatedAt:              row.UpdatedAt.Time,
	}
	if row.RequestFingerprint.Valid {
		item.RequestFingerprint = row.RequestFingerprint.String
	}
	if row.ManualActionKind.Valid {
		item.ManualActionKind = enumtypes.GitHubRateLimitManualActionKind(row.ManualActionKind.String)
	}
	if row.ResumeNotBefore.Valid {
		value := row.ResumeNotBefore.Time
		item.ResumeNotBefore = &value
	}
	if row.LastResumeAttemptAt.Valid {
		value := row.LastResumeAttemptAt.Time
		item.LastResumeAttemptAt = &value
	}
	if row.ResolvedAt.Valid {
		value := row.ResolvedAt.Time
		item.ResolvedAt = &value
	}
	return item
}

func evidenceFromDBModel(row dbmodel.EvidenceRow) domainrepo.Evidence {
	item := domainrepo.Evidence{
		ID:          row.ID,
		WaitID:      row.WaitID,
		EventKind:   enumtypes.GitHubRateLimitEvidenceEventKind(row.EventKind),
		PayloadJSON: row.PayloadJSON,
		ObservedAt:  row.ObservedAt.Time,
		CreatedAt:   row.CreatedAt.Time,
	}
	if row.SignalID.Valid {
		item.SignalID = row.SignalID.String
	}
	if row.SignalOrigin.Valid {
		item.SignalOrigin = enumtypes.GitHubRateLimitSignalOrigin(row.SignalOrigin.String)
	}
	if row.ProviderStatusCode.Valid {
		value := int(row.ProviderStatusCode.Int32)
		item.ProviderStatusCode = &value
	}
	if row.RetryAfterSeconds.Valid {
		value := int(row.RetryAfterSeconds.Int32)
		item.RetryAfterSeconds = &value
	}
	if row.RateLimitLimit.Valid {
		value := int(row.RateLimitLimit.Int32)
		item.RateLimitLimit = &value
	}
	if row.RateLimitRemaining.Valid {
		value := int(row.RateLimitRemaining.Int32)
		item.RateLimitRemaining = &value
	}
	if row.RateLimitUsed.Valid {
		value := int(row.RateLimitUsed.Int32)
		item.RateLimitUsed = &value
	}
	if row.RateLimitResetAt.Valid {
		value := row.RateLimitResetAt.Time
		item.RateLimitResetAt = &value
	}
	if row.RateLimitResource.Valid {
		item.RateLimitResource = row.RateLimitResource.String
	}
	if row.GitHubRequestID.Valid {
		item.GitHubRequestID = row.GitHubRequestID.String
	}
	if row.DocumentationURL.Valid {
		item.DocumentationURL = row.DocumentationURL.String
	}
	if row.MessageExcerpt.Valid {
		item.MessageExcerpt = row.MessageExcerpt.String
	}
	if row.StderrExcerpt.Valid {
		item.StderrExcerpt = row.StderrExcerpt.String
	}
	return item
}
