package app

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/registry"
)

type registryJobImageChecker struct {
	client       *registry.Client
	internalHost string
}

func newRegistryJobImageChecker(scheme string, host string, timeout time.Duration) (*registryJobImageChecker, error) {
	normalizedHost := strings.TrimSpace(host)
	if normalizedHost == "" {
		return nil, fmt.Errorf("internal registry host is required")
	}

	normalizedScheme := strings.ToLower(strings.TrimSpace(scheme))
	if normalizedScheme == "" {
		normalizedScheme = "http"
	}

	client, err := registry.NewClient(normalizedScheme+"://"+normalizedHost, timeout)
	if err != nil {
		return nil, err
	}

	return &registryJobImageChecker{
		client:       client,
		internalHost: normalizedHost,
	}, nil
}

func (c *registryJobImageChecker) IsImageAvailable(ctx context.Context, imageRef string) (bool, error) {
	repository, tag := registry.SplitImageRef(imageRef)
	repositoryPath := registry.ExtractRepositoryPath(repository, c.internalHost)
	if repositoryPath == "" || strings.TrimSpace(tag) == "" {
		// Not an internal-registry image reference; treat as available and skip fallback switching.
		return true, nil
	}

	_, found, err := c.client.GetTagInfo(ctx, repositoryPath, tag)
	if err != nil {
		return false, err
	}
	return found, nil
}

func (c *registryJobImageChecker) ResolvePreviousImage(ctx context.Context, imageRef string) (string, bool, error) {
	repository, primaryTag := registry.SplitImageRef(imageRef)
	repositoryPath := registry.ExtractRepositoryPath(repository, c.internalHost)
	if repositoryPath == "" || strings.TrimSpace(primaryTag) == "" {
		return "", false, nil
	}

	tagInfos, err := c.client.ListTagInfos(ctx, repositoryPath)
	if err != nil {
		return "", false, err
	}
	if len(tagInfos) == 0 {
		return "", false, nil
	}

	sort.SliceStable(tagInfos, func(i, j int) bool {
		left := tagInfos[i]
		right := tagInfos[j]
		switch {
		case left.CreatedAt != nil && right.CreatedAt != nil:
			if !left.CreatedAt.Equal(*right.CreatedAt) {
				return left.CreatedAt.After(*right.CreatedAt)
			}
		case left.CreatedAt != nil:
			return true
		case right.CreatedAt != nil:
			return false
		}
		return left.Tag > right.Tag
	})

	for _, item := range tagInfos {
		tag := strings.TrimSpace(item.Tag)
		if tag == "" || tag == strings.TrimSpace(primaryTag) {
			continue
		}
		return repository + ":" + tag, true, nil
	}
	return "", false, nil
}
