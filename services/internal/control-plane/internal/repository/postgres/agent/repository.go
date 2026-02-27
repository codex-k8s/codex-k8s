package agent

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agent"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/find_effective_by_key.sql
	queryFindEffectiveByKey string
	//go:embed sql/list.sql
	queryList string
	//go:embed sql/get_by_id.sql
	queryGetByID string
	//go:embed sql/update_settings.sql
	queryUpdateSettings string
)

// Repository stores agent profiles in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL agent profile repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// FindEffectiveByKey resolves active agent profile by key with project override priority.
func (r *Repository) FindEffectiveByKey(ctx context.Context, projectID string, agentKey string) (domainrepo.Agent, bool, error) {
	var (
		item           domainrepo.Agent
		projectIDValue pgtype.Text
	)

	err := r.db.QueryRow(ctx, queryFindEffectiveByKey, agentKey, projectID).Scan(
		&item.ID,
		&item.AgentKey,
		&item.RoleKind,
		&projectIDValue,
		&item.Name,
	)
	if err == nil {
		if projectIDValue.Valid {
			item.ProjectID = projectIDValue.String
		}
		return item, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domainrepo.Agent{}, false, nil
	}
	return domainrepo.Agent{}, false, fmt.Errorf("find effective agent by key: %w", err)
}

// List returns active agents visible for requested project ids.
func (r *Repository) List(ctx context.Context, filter querytypes.AgentListFilter) ([]domainrepo.Agent, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}
	projectIDs := normalizeProjectIDs(filter.ProjectIDs)
	rows, err := r.db.Query(ctx, queryList, projectIDs, limit, filter.IncludeAllProjects)
	if err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}
	items, err := pgx.CollectRows(rows, rowToAgent)
	if err != nil {
		return nil, fmt.Errorf("collect agents: %w", err)
	}
	return items, nil
}

// GetByID loads one agent by id.
func (r *Repository) GetByID(ctx context.Context, agentID string) (domainrepo.Agent, bool, error) {
	item, err := queryOneAgent(ctx, r.db, queryGetByID, strings.TrimSpace(agentID))
	if err != nil {
		return domainrepo.Agent{}, false, fmt.Errorf("get agent by id: %w", err)
	}
	if item == nil {
		return domainrepo.Agent{}, false, nil
	}
	return *item, true, nil
}

// UpdateSettings updates settings with optimistic concurrency control.
func (r *Repository) UpdateSettings(ctx context.Context, params querytypes.AgentUpdateSettingsParams) (domainrepo.Agent, bool, error) {
	settingsJSON, err := json.Marshal(agentSettingsJSON{
		RuntimeMode:       strings.TrimSpace(params.Settings.RuntimeMode),
		TimeoutSeconds:    params.Settings.TimeoutSeconds,
		MaxRetryCount:     params.Settings.MaxRetryCount,
		PromptLocale:      strings.TrimSpace(params.Settings.PromptLocale),
		ApprovalsRequired: params.Settings.ApprovalsRequired,
	})
	if err != nil {
		return domainrepo.Agent{}, false, fmt.Errorf("marshal agent settings: %w", err)
	}

	item, err := queryOneAgent(ctx, r.db, queryUpdateSettings, strings.TrimSpace(params.AgentID), settingsJSON, params.ExpectedVersion)
	if err != nil {
		return domainrepo.Agent{}, false, fmt.Errorf("update agent settings: %w", err)
	}
	if item == nil {
		return domainrepo.Agent{}, false, nil
	}
	return *item, true, nil
}

type agentSettingsJSON struct {
	RuntimeMode       string `json:"runtime_mode"`
	TimeoutSeconds    int    `json:"timeout_seconds"`
	MaxRetryCount     int    `json:"max_retry_count"`
	PromptLocale      string `json:"prompt_locale"`
	ApprovalsRequired bool   `json:"approvals_required"`
}

func rowToAgent(row pgx.CollectableRow) (domainrepo.Agent, error) {
	var (
		item            domainrepo.Agent
		projectIDValue  pgtype.Text
		settingsJSONRaw []byte
	)
	if err := row.Scan(
		&item.ID,
		&item.AgentKey,
		&item.RoleKind,
		&projectIDValue,
		&item.Name,
		&item.IsActive,
		&settingsJSONRaw,
		&item.SettingsVersion,
	); err != nil {
		return domainrepo.Agent{}, err
	}
	if projectIDValue.Valid {
		item.ProjectID = strings.TrimSpace(projectIDValue.String)
	}
	item.Settings = parseAgentSettings(settingsJSONRaw)
	return item, nil
}

func queryOneAgent(ctx context.Context, db *pgxpool.Pool, query string, args ...interface{}) (*domainrepo.Agent, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	items, err := pgx.CollectRows(rows, rowToAgent)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func normalizeProjectIDs(projectIDs []string) []string {
	if len(projectIDs) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(projectIDs))
	seen := make(map[string]struct{}, len(projectIDs))
	for _, projectID := range projectIDs {
		value := strings.TrimSpace(projectID)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	if len(out) == 0 {
		return []string{}
	}
	return out
}

func parseAgentSettings(raw []byte) domainrepo.AgentSettings {
	out := domainrepo.AgentSettings{
		RuntimeMode:       "code-only",
		TimeoutSeconds:    3600,
		MaxRetryCount:     1,
		PromptLocale:      "en",
		ApprovalsRequired: false,
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "{}" {
		return out
	}
	var payload agentSettingsJSON
	if err := json.Unmarshal(raw, &payload); err != nil {
		return out
	}
	if value := strings.TrimSpace(payload.RuntimeMode); value != "" {
		out.RuntimeMode = value
	}
	if payload.TimeoutSeconds > 0 {
		out.TimeoutSeconds = payload.TimeoutSeconds
	}
	if payload.MaxRetryCount >= 0 {
		out.MaxRetryCount = payload.MaxRetryCount
	}
	if value := strings.TrimSpace(payload.PromptLocale); value != "" {
		out.PromptLocale = value
	}
	out.ApprovalsRequired = payload.ApprovalsRequired
	return out
}
