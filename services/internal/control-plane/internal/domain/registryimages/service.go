package registryimages

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/registry"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

const (
	defaultListRepositoriesLimit = 100
	defaultListTagsLimit         = 50
	defaultCleanupKeepTags       = 5
)

// RegistryClient describes operations required from registry API adapter.
type RegistryClient interface {
	ListRepositories(ctx context.Context) ([]string, error)
	ListTags(ctx context.Context, repository string) ([]string, error)
	ListTagInfos(ctx context.Context, repository string) ([]registry.TagInfo, error)
	DeleteTag(ctx context.Context, repository string, tag string) (registry.DeleteResult, error)
}

// Config controls image registry management behavior.
type Config struct {
	DefaultCleanupKeepTags int
}

// Service provides typed list/delete/cleanup operations for internal registry images.
type Service struct {
	cfg    Config
	client RegistryClient
}

// NewService creates image registry management service.
func NewService(cfg Config, client RegistryClient) (*Service, error) {
	if client == nil {
		return nil, fmt.Errorf("registry client is required")
	}
	if cfg.DefaultCleanupKeepTags <= 0 {
		cfg.DefaultCleanupKeepTags = defaultCleanupKeepTags
	}
	return &Service{
		cfg:    cfg,
		client: client,
	}, nil
}

// List returns repositories with tags from internal registry.
func (s *Service) List(ctx context.Context, filter querytypes.RegistryImageListFilter) ([]entitytypes.RegistryImageRepository, error) {
	limitRepositories := filter.LimitRepositories
	if limitRepositories <= 0 {
		limitRepositories = defaultListRepositoriesLimit
	}
	limitTags := filter.LimitTags
	if limitTags <= 0 {
		limitTags = defaultListTagsLimit
	}

	repositories, err := s.client.ListRepositories(ctx)
	if err != nil {
		return nil, err
	}

	needle := strings.ToLower(strings.TrimSpace(filter.Repository))
	items := make([]entitytypes.RegistryImageRepository, 0, limitRepositories)
	for _, repository := range repositories {
		if needle != "" && !strings.Contains(strings.ToLower(repository), needle) {
			continue
		}
		// Listing tag infos (digest+metadata) is expensive (requires multiple HTTP calls per tag).
		// For the main list endpoint, we only need repository and tags; digests are resolved on delete/cleanup.
		tags, err := s.client.ListTags(ctx, repository)
		if err != nil {
			return nil, err
		}
		repositoryItem := entitytypes.RegistryImageRepository{
			Repository: repository,
			TagCount:   len(tags),
			Tags:       mapTags(tags, limitTags),
		}
		items = append(items, repositoryItem)
		if len(items) >= limitRepositories {
			break
		}
	}
	return items, nil
}

// DeleteTag deletes one tag from internal registry.
func (s *Service) DeleteTag(ctx context.Context, params querytypes.RegistryImageDeleteParams) (entitytypes.RegistryImageDeleteResult, error) {
	repository := strings.TrimSpace(params.Repository)
	tag := strings.TrimSpace(params.Tag)
	if repository == "" {
		return entitytypes.RegistryImageDeleteResult{}, fmt.Errorf("repository is required")
	}
	if tag == "" {
		return entitytypes.RegistryImageDeleteResult{}, fmt.Errorf("tag is required")
	}

	result, err := s.client.DeleteTag(ctx, repository, tag)
	if err != nil {
		return entitytypes.RegistryImageDeleteResult{}, err
	}
	return entitytypes.RegistryImageDeleteResult{
		Repository: result.Repository,
		Tag:        result.Tag,
		Digest:     result.Digest,
		Deleted:    result.Deleted,
	}, nil
}

// Cleanup deletes stale tags and keeps latest N tags for each repository.
func (s *Service) Cleanup(ctx context.Context, filter querytypes.RegistryImageCleanupFilter) (entitytypes.RegistryImageCleanupResult, error) {
	limitRepositories := filter.LimitRepositories
	if limitRepositories <= 0 {
		limitRepositories = defaultListRepositoriesLimit
	}
	keepTags := filter.KeepTags
	if keepTags <= 0 {
		keepTags = s.cfg.DefaultCleanupKeepTags
	}
	prefix := strings.TrimSpace(filter.RepositoryPrefix)
	isDryRun := filter.DryRun

	repositories, err := s.client.ListRepositories(ctx)
	if err != nil {
		return entitytypes.RegistryImageCleanupResult{}, err
	}

	result := entitytypes.RegistryImageCleanupResult{
		Deleted: make([]entitytypes.RegistryImageDeleteResult, 0),
		Skipped: make([]entitytypes.RegistryImageDeleteResult, 0),
	}
	for _, repository := range repositories {
		if prefix != "" && !strings.HasPrefix(repository, prefix) {
			continue
		}
		result.RepositoriesScanned++
		tagInfos, err := s.client.ListTagInfos(ctx, repository)
		if err != nil {
			return entitytypes.RegistryImageCleanupResult{}, err
		}
		for idx, tagInfo := range tagInfos {
			keep := idx < keepTags
			if keep || isDryRun {
				result.TagsSkipped++
				result.Skipped = append(result.Skipped, entitytypes.RegistryImageDeleteResult{
					Repository: repository,
					Tag:        tagInfo.Tag,
					Digest:     tagInfo.Digest,
					Deleted:    false,
				})
				continue
			}
			deleteResult, err := s.client.DeleteTag(ctx, repository, tagInfo.Tag)
			if err != nil {
				return entitytypes.RegistryImageCleanupResult{}, err
			}
			item := entitytypes.RegistryImageDeleteResult{
				Repository: deleteResult.Repository,
				Tag:        deleteResult.Tag,
				Digest:     deleteResult.Digest,
				Deleted:    deleteResult.Deleted,
			}
			if deleteResult.Deleted {
				result.TagsDeleted++
				result.Deleted = append(result.Deleted, item)
			} else {
				result.TagsSkipped++
				result.Skipped = append(result.Skipped, item)
			}
		}
		if result.RepositoriesScanned >= limitRepositories {
			break
		}
	}

	return result, nil
}

func mapTagInfos(items []registry.TagInfo, limit int) []entitytypes.RegistryImageTag {
	if len(items) == 0 {
		return []entitytypes.RegistryImageTag{}
	}
	out := make([]entitytypes.RegistryImageTag, 0, minInt(len(items), limit))
	for idx := range items {
		if len(out) >= limit {
			break
		}
		item := items[idx]
		var createdAt *time.Time
		if item.CreatedAt != nil {
			value := item.CreatedAt.UTC()
			createdAt = &value
		}
		out = append(out, entitytypes.RegistryImageTag{
			Tag:             item.Tag,
			Digest:          item.Digest,
			CreatedAt:       createdAt,
			ConfigSizeBytes: item.ConfigSizeBytes,
		})
	}
	return out
}

func mapTags(tags []string, limit int) []entitytypes.RegistryImageTag {
	if len(tags) == 0 {
		return []entitytypes.RegistryImageTag{}
	}
	if limit <= 0 {
		limit = defaultListTagsLimit
	}
	out := make([]entitytypes.RegistryImageTag, 0, minInt(len(tags), limit))
	for _, tag := range tags {
		if len(out) >= limit {
			break
		}
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		out = append(out, entitytypes.RegistryImageTag{Tag: tag})
	}
	return out
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
