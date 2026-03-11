package agentsession

import (
	"strings"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentsession"
)

type upsertRecord struct {
	RunID              string
	CorrelationID      string
	ProjectID          string
	RepositoryFullName string
	AgentKey           string
	IssueNumber        *int
	BranchName         string
	PRNumber           *int
	PRURL              string
	TriggerKind        string
	TemplateKind       string
	TemplateSource     string
	TemplateLocale     string
	Model              string
	ReasoningEffort    string
	Status             string
	SessionID          string
	SessionJSON        []byte
	CodexSessionPath   string
	CodexSessionJSON   []byte
	StartedAt          time.Time
	FinishedAt         *time.Time
	ExpectedVersion    int64
	SnapshotChecksum   string
	SnapshotUpdatedAt  time.Time
}

func buildUpsertRecord(params domainrepo.UpsertParams, existing *domainrepo.Session) (upsertRecord, error) {
	record := upsertRecord{
		RunID:              strings.TrimSpace(params.RunID),
		CorrelationID:      strings.TrimSpace(params.CorrelationID),
		ProjectID:          strings.TrimSpace(params.ProjectID),
		RepositoryFullName: strings.TrimSpace(params.RepositoryFullName),
		AgentKey:           strings.TrimSpace(params.AgentKey),
		BranchName:         strings.TrimSpace(params.BranchName),
		PRURL:              strings.TrimSpace(params.PRURL),
		TriggerKind:        strings.TrimSpace(params.TriggerKind),
		TemplateKind:       strings.TrimSpace(params.TemplateKind),
		TemplateSource:     strings.TrimSpace(params.TemplateSource),
		TemplateLocale:     strings.TrimSpace(params.TemplateLocale),
		Model:              strings.TrimSpace(params.Model),
		ReasoningEffort:    strings.TrimSpace(params.ReasoningEffort),
		Status:             strings.TrimSpace(params.Status),
		SessionID:          strings.TrimSpace(params.SessionID),
		CodexSessionPath:   strings.TrimSpace(params.CodexSessionPath),
		ExpectedVersion:    params.ExpectedSnapshotVersion,
	}

	if len(params.SessionJSON) > 0 {
		record.SessionJSON = append([]byte(nil), params.SessionJSON...)
	} else {
		record.SessionJSON = []byte(`{}`)
	}

	if len(params.CodexSessionJSON) > 0 {
		record.CodexSessionJSON = append([]byte(nil), params.CodexSessionJSON...)
	}

	if params.IssueNumber != nil {
		value := *params.IssueNumber
		record.IssueNumber = &value
	}
	if params.PRNumber != nil {
		value := *params.PRNumber
		record.PRNumber = &value
	}
	if !params.StartedAt.IsZero() {
		record.StartedAt = params.StartedAt.UTC()
	}
	if params.FinishedAt != nil {
		value := params.FinishedAt.UTC()
		record.FinishedAt = &value
	}

	if existing != nil {
		if record.CorrelationID == "" {
			record.CorrelationID = strings.TrimSpace(existing.CorrelationID)
		}
		if record.ProjectID == "" {
			record.ProjectID = strings.TrimSpace(existing.ProjectID)
		}
		if record.IssueNumber == nil && existing.IssueNumber > 0 {
			value := existing.IssueNumber
			record.IssueNumber = &value
		}
		if record.PRNumber == nil && existing.PRNumber > 0 {
			value := existing.PRNumber
			record.PRNumber = &value
		}
		if record.PRURL == "" {
			record.PRURL = strings.TrimSpace(existing.PRURL)
		}
		if record.TriggerKind == "" {
			record.TriggerKind = strings.TrimSpace(existing.TriggerKind)
		}
		if record.TemplateKind == "" {
			record.TemplateKind = strings.TrimSpace(existing.TemplateKind)
		}
		if record.TemplateSource == "" {
			record.TemplateSource = strings.TrimSpace(existing.TemplateSource)
		}
		if record.TemplateLocale == "" {
			record.TemplateLocale = strings.TrimSpace(existing.TemplateLocale)
		}
		if record.Model == "" {
			record.Model = strings.TrimSpace(existing.Model)
		}
		if record.ReasoningEffort == "" {
			record.ReasoningEffort = strings.TrimSpace(existing.ReasoningEffort)
		}
		if record.SessionID == "" {
			record.SessionID = strings.TrimSpace(existing.SessionID)
		}
		if len(record.SessionJSON) == 0 {
			record.SessionJSON = append([]byte(nil), existing.SessionJSON...)
		}
		if record.CodexSessionPath == "" {
			record.CodexSessionPath = strings.TrimSpace(existing.CodexSessionPath)
		}
		if len(record.CodexSessionJSON) == 0 && len(existing.CodexSessionJSON) > 0 {
			record.CodexSessionJSON = append([]byte(nil), existing.CodexSessionJSON...)
		}
		if record.StartedAt.IsZero() {
			record.StartedAt = existing.StartedAt.UTC()
		}
		if record.FinishedAt == nil && !existing.FinishedAt.IsZero() {
			value := existing.FinishedAt.UTC()
			record.FinishedAt = &value
		}
	}

	checksum, err := agentdomain.ComputeSessionSnapshotChecksum(record.SessionJSON, record.CodexSessionJSON)
	if err != nil {
		return upsertRecord{}, err
	}
	record.SnapshotChecksum = checksum
	record.SnapshotUpdatedAt = time.Now().UTC()
	return record, nil
}

func snapshotStateFromSession(item domainrepo.Session) domainrepo.UpsertResult {
	return domainrepo.UpsertResult{
		SnapshotVersion:   item.SnapshotVersion,
		SnapshotChecksum:  strings.TrimSpace(item.SnapshotChecksum),
		SnapshotUpdatedAt: item.SnapshotUpdatedAt.UTC(),
	}
}

func isIdempotentReplay(item domainrepo.Session, expectedVersion int64, checksum string) bool {
	normalizedChecksum := strings.TrimSpace(checksum)
	return item.SnapshotVersion == expectedVersion+1 &&
		normalizedChecksum != "" &&
		strings.EqualFold(strings.TrimSpace(item.SnapshotChecksum), normalizedChecksum)
}
