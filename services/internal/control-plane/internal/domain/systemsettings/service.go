package systemsettings

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	sharedsystemsettings "github.com/codex-k8s/codex-k8s/libs/go/systemsettings"
	systemsettingrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/systemsetting"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

// Service owns typed platform settings catalog, cache and persistence semantics.
type Service struct {
	repo    systemsettingrepo.Repository
	catalog map[enumtypes.SystemSettingKey]valuetypes.SystemSettingCatalogEntry

	mu      sync.RWMutex
	records map[enumtypes.SystemSettingKey]entitytypes.SystemSettingRecord
}

// NewService constructs the system settings domain service.
func NewService(repo systemsettingrepo.Repository) (*Service, error) {
	return newService(repo, defaultCatalog())
}

func newService(repo systemsettingrepo.Repository, catalog map[enumtypes.SystemSettingKey]valuetypes.SystemSettingCatalogEntry) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("system settings repository is required")
	}
	if len(catalog) == 0 {
		return nil, fmt.Errorf("system settings catalog is required")
	}
	return &Service{
		repo:    repo,
		catalog: catalog,
		records: make(map[enumtypes.SystemSettingKey]entitytypes.SystemSettingRecord),
	}, nil
}

// RefreshCache reloads current persisted settings snapshot into in-memory cache.
func (s *Service) RefreshCache(ctx context.Context) error {
	items, err := s.repo.List(ctx)
	if err != nil {
		return err
	}
	s.replaceCache(items)
	return nil
}

// List returns all staff-visible settings merged with current cache snapshot.
func (s *Service) List() []entitytypes.SystemSetting {
	keys := make([]enumtypes.SystemSettingKey, 0, len(s.catalog))
	for key := range s.catalog {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i int, j int) bool { return keys[i] < keys[j] })

	out := make([]entitytypes.SystemSetting, 0, len(keys))
	for _, key := range keys {
		item, ok := s.getSettingForSurface(key, sharedsystemsettings.ExposureSurfaceStaff)
		if ok {
			out = append(out, item)
		}
	}
	return out
}

// Get returns one staff-visible setting by key.
func (s *Service) Get(key string) (entitytypes.SystemSetting, error) {
	parsed, err := s.requireCatalogEntry(key)
	if err != nil {
		return entitytypes.SystemSetting{}, err
	}
	item, ok := s.getSettingForSurface(parsed.Key, sharedsystemsettings.ExposureSurfaceStaff)
	if !ok {
		return entitytypes.SystemSetting{}, notFoundError(key)
	}
	return item, nil
}

// UpdateBoolean persists a boolean setting value and updates the local cache immediately.
func (s *Service) UpdateBoolean(ctx context.Context, params querytypes.SystemSettingBooleanWriteParams) (entitytypes.SystemSetting, error) {
	entry, err := s.requireCatalogEntry(string(params.Key))
	if err != nil {
		return entitytypes.SystemSetting{}, err
	}
	if entry.ValueKind != enumtypes.SystemSettingValueKindBoolean {
		return entitytypes.SystemSetting{}, fmt.Errorf("system setting %q is not boolean", params.Key)
	}

	record, err := s.repo.UpsertBoolean(ctx, params)
	if err != nil {
		return entitytypes.SystemSetting{}, err
	}
	s.upsertCache(record)
	return s.mergeSetting(entry, record), nil
}

// Reset restores one setting to catalog default and increments version/audit state.
func (s *Service) Reset(ctx context.Context, key string, actorUserID string, actorEmail string) (entitytypes.SystemSetting, error) {
	entry, err := s.requireCatalogEntry(key)
	if err != nil {
		return entitytypes.SystemSetting{}, err
	}

	record, err := s.repo.UpsertBoolean(ctx, querytypes.SystemSettingBooleanWriteParams{
		Key:          entry.Key,
		BooleanValue: entry.DefaultBooleanValue,
		Source:       enumtypes.SystemSettingSourceDefault,
		ChangeKind:   enumtypes.SystemSettingChangeKindReset,
		ActorUserID:  strings.TrimSpace(actorUserID),
		ActorEmail:   strings.TrimSpace(actorEmail),
	})
	if err != nil {
		return entitytypes.SystemSetting{}, err
	}
	s.upsertCache(record)
	return s.mergeSetting(entry, record), nil
}

// GitHubRateLimitWaitEnabled returns current effective rollout toggle.
func (s *Service) GitHubRateLimitWaitEnabled() bool {
	item, ok := s.getSettingForSurface(
		enumtypes.SystemSettingKeyGitHubRateLimitWaitEnabled,
		sharedsystemsettings.ExposureSurfaceStaff,
	)
	if !ok {
		return false
	}
	return item.BooleanValue
}

// CurrentGitHubRateLimitRolloutState maps the typed setting into existing rollout guard shape.
func (s *Service) CurrentGitHubRateLimitRolloutState() valuetypes.GitHubRateLimitRolloutState {
	enabled := s.GitHubRateLimitWaitEnabled()
	return valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: enabled,
		SchemaReady:        enabled,
		DomainReady:        enabled,
		WorkerReady:        enabled,
		RunnerReady:        enabled,
		TransportReady:     enabled,
		UIFeatureEnabled:   enabled,
	}
}

func defaultCatalog() map[enumtypes.SystemSettingKey]valuetypes.SystemSettingCatalogEntry {
	return map[enumtypes.SystemSettingKey]valuetypes.SystemSettingCatalogEntry{
		enumtypes.SystemSettingKeyGitHubRateLimitWaitEnabled: {
			Key:                 enumtypes.SystemSettingKeyGitHubRateLimitWaitEnabled,
			Section:             enumtypes.SystemSettingSectionGitHubRateLimit,
			ValueKind:           enumtypes.SystemSettingValueKindBoolean,
			ReloadSemantics:     enumtypes.SystemSettingReloadSemanticsHotReload,
			Visibility:          enumtypes.SystemSettingVisibilityStaffVisible,
			DefaultBooleanValue: false,
		},
	}
}

func (s *Service) requireCatalogEntry(key string) (valuetypes.SystemSettingCatalogEntry, error) {
	parsed, err := s.requireCatalogKey(key)
	if err != nil {
		return valuetypes.SystemSettingCatalogEntry{}, err
	}
	entry, ok := s.catalog[parsed]
	if !ok {
		return valuetypes.SystemSettingCatalogEntry{}, notFoundError(key)
	}
	return entry, nil
}

func (s *Service) requireCatalogKey(key string) (enumtypes.SystemSettingKey, error) {
	trimmed := strings.TrimSpace(key)
	parsed := enumtypes.SystemSettingKey(trimmed)
	if parsed == "" {
		return "", errs.Validation{Field: "setting_key", Msg: "is required"}
	}
	if _, ok := s.catalog[parsed]; !ok {
		return "", notFoundError(trimmed)
	}
	return parsed, nil
}

func (s *Service) getSettingForSurface(
	key enumtypes.SystemSettingKey,
	surface sharedsystemsettings.ExposureSurface,
) (entitytypes.SystemSetting, bool) {
	entry, ok := s.catalog[key]
	if !ok {
		return entitytypes.SystemSetting{}, false
	}
	if !sharedsystemsettings.IsVisibleOnSurface(string(entry.Visibility), surface) {
		return entitytypes.SystemSetting{}, false
	}

	s.mu.RLock()
	record, found := s.records[key]
	s.mu.RUnlock()
	return s.mergeSetting(entry, recordOrDefault(record, found, entry)), true
}

func (s *Service) mergeSetting(entry valuetypes.SystemSettingCatalogEntry, record entitytypes.SystemSettingRecord) entitytypes.SystemSetting {
	var updatedAt *time.Time
	if !record.UpdatedAt.IsZero() {
		value := record.UpdatedAt.UTC()
		updatedAt = &value
	}

	return entitytypes.SystemSetting{
		Key:                 entry.Key,
		Section:             entry.Section,
		ValueKind:           entry.ValueKind,
		ReloadSemantics:     entry.ReloadSemantics,
		Visibility:          entry.Visibility,
		BooleanValue:        record.BooleanValue,
		DefaultBooleanValue: entry.DefaultBooleanValue,
		Source:              record.Source,
		Version:             record.Version,
		UpdatedAt:           updatedAt,
		UpdatedByUserID:     record.UpdatedByUserID,
		UpdatedByEmail:      record.UpdatedByEmail,
	}
}

func recordOrDefault(record entitytypes.SystemSettingRecord, found bool, entry valuetypes.SystemSettingCatalogEntry) entitytypes.SystemSettingRecord {
	if found {
		return record
	}
	return entitytypes.SystemSettingRecord{
		Key:          entry.Key,
		ValueKind:    entry.ValueKind,
		BooleanValue: entry.DefaultBooleanValue,
		Source:       enumtypes.SystemSettingSourceDefault,
		Version:      0,
	}
}

func (s *Service) replaceCache(items []entitytypes.SystemSettingRecord) {
	next := make(map[enumtypes.SystemSettingKey]entitytypes.SystemSettingRecord, len(items))
	for _, item := range items {
		next[item.Key] = item
	}

	s.mu.Lock()
	s.records = next
	s.mu.Unlock()
}

func (s *Service) upsertCache(item entitytypes.SystemSettingRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[item.Key] = item
}

func notFoundError(key string) error {
	return errs.NotFound{Msg: fmt.Sprintf("system setting %q not found", strings.TrimSpace(key))}
}
