package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	defaultHTTPTimeout       = 15 * time.Second
	defaultCatalogPageSize   = 100
	manifestAcceptHeader     = "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json"
	registryDigestHeaderName = "Docker-Content-Digest"
)

// TagInfo describes one repository tag in internal registry.
type TagInfo struct {
	Tag             string
	Digest          string
	CreatedAt       *time.Time
	ConfigSizeBytes int64
}

// DeleteResult describes one tag deletion result.
type DeleteResult struct {
	Repository string
	Tag        string
	Digest     string
	Deleted    bool
}

// Client provides typed access to Docker Registry HTTP API v2.
type Client struct {
	baseURL *url.URL
	http    *http.Client
}

// NewClient creates registry API client.
func NewClient(baseURL string, timeout time.Duration) (*Client, error) {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return nil, fmt.Errorf("registry base url is required")
	}
	if !strings.Contains(trimmed, "://") {
		trimmed = "http://" + trimmed
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("parse registry base url: %w", err)
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return nil, fmt.Errorf("registry host is required")
	}
	if timeout <= 0 {
		timeout = defaultHTTPTimeout
	}
	return &Client{
		baseURL: parsed,
		http:    &http.Client{Timeout: timeout},
	}, nil
}

// ListRepositories lists all repositories from registry catalog.
func (c *Client) ListRepositories(ctx context.Context) ([]string, error) {
	repositories := make([]string, 0)
	last := ""
	for {
		query := url.Values{}
		query.Set("n", fmt.Sprintf("%d", defaultCatalogPageSize))
		if last != "" {
			query.Set("last", last)
		}
		endpoint := c.buildURL("/v2/_catalog", query)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("build catalog request: %w", err)
		}
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request catalog: %w", err)
		}

		body, readErr := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read catalog response: %w", readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close catalog response: %w", closeErr)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("catalog request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		var payload struct {
			Repositories []string `json:"repositories"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("decode catalog response: %w", err)
		}
		if len(payload.Repositories) == 0 {
			break
		}
		for _, repository := range payload.Repositories {
			repository = strings.TrimSpace(repository)
			if repository == "" {
				continue
			}
			repositories = append(repositories, repository)
		}
		if len(payload.Repositories) < defaultCatalogPageSize {
			break
		}
		last = strings.TrimSpace(payload.Repositories[len(payload.Repositories)-1])
		if last == "" {
			break
		}
	}

	sort.Strings(repositories)
	return repositories, nil
}

// ListTagInfos lists tags for one repository with digest and metadata.
func (c *Client) ListTagInfos(ctx context.Context, repository string) ([]TagInfo, error) {
	repoPath, err := encodeRepository(repository)
	if err != nil {
		return nil, err
	}

	endpoint := c.buildURL("/v2/"+repoPath+"/tags/list", nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build tags request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request tags list: %w", err)
	}
	body, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if readErr != nil {
		return nil, fmt.Errorf("read tags response: %w", readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close tags response: %w", closeErr)
	}
	if resp.StatusCode == http.StatusNotFound {
		return []TagInfo{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tags request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode tags response: %w", err)
	}
	if len(payload.Tags) == 0 {
		return []TagInfo{}, nil
	}

	items := make([]TagInfo, 0, len(payload.Tags))
	for _, tag := range payload.Tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		digest, err := c.resolveTagDigest(ctx, repoPath, tag)
		if err != nil {
			return nil, fmt.Errorf("resolve digest for %s:%s: %w", repository, tag, err)
		}
		if digest == "" {
			continue
		}
		createdAt, configSizeBytes, err := c.loadTagMetadata(ctx, repoPath, digest)
		if err != nil {
			return nil, fmt.Errorf("load metadata for %s@%s: %w", repository, digest, err)
		}
		items = append(items, TagInfo{
			Tag:             tag,
			Digest:          digest,
			CreatedAt:       createdAt,
			ConfigSizeBytes: configSizeBytes,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		left := items[i]
		right := items[j]
		if left.CreatedAt != nil && right.CreatedAt != nil {
			if left.CreatedAt.Equal(*right.CreatedAt) {
				return left.Tag > right.Tag
			}
			return left.CreatedAt.After(*right.CreatedAt)
		}
		if left.CreatedAt != nil {
			return true
		}
		if right.CreatedAt != nil {
			return false
		}
		return left.Tag > right.Tag
	})

	return items, nil
}

// DeleteTag deletes repository tag by deleting its manifest digest.
func (c *Client) DeleteTag(ctx context.Context, repository string, tag string) (DeleteResult, error) {
	repository = strings.TrimSpace(repository)
	tag = strings.TrimSpace(tag)
	if repository == "" {
		return DeleteResult{}, fmt.Errorf("repository is required")
	}
	if tag == "" {
		return DeleteResult{}, fmt.Errorf("tag is required")
	}

	repoPath, err := encodeRepository(repository)
	if err != nil {
		return DeleteResult{}, err
	}

	digest, err := c.resolveTagDigest(ctx, repoPath, tag)
	if err != nil {
		return DeleteResult{}, fmt.Errorf("resolve digest for %s:%s: %w", repository, tag, err)
	}
	if digest == "" {
		return DeleteResult{
			Repository: repository,
			Tag:        tag,
			Deleted:    false,
		}, nil
	}

	endpoint := c.buildURL("/v2/"+repoPath+"/manifests/"+url.PathEscape(digest), nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return DeleteResult{}, fmt.Errorf("build delete request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return DeleteResult{}, fmt.Errorf("delete manifest request: %w", err)
	}
	body, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if readErr != nil {
		return DeleteResult{}, fmt.Errorf("read delete response: %w", readErr)
	}
	if closeErr != nil {
		return DeleteResult{}, fmt.Errorf("close delete response: %w", closeErr)
	}

	switch resp.StatusCode {
	case http.StatusAccepted:
		return DeleteResult{
			Repository: repository,
			Tag:        tag,
			Digest:     digest,
			Deleted:    true,
		}, nil
	case http.StatusNotFound:
		return DeleteResult{
			Repository: repository,
			Tag:        tag,
			Digest:     digest,
			Deleted:    false,
		}, nil
	default:
		return DeleteResult{}, fmt.Errorf("delete manifest failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

func (c *Client) resolveTagDigest(ctx context.Context, repoPath string, tag string) (string, error) {
	endpoint := c.buildURL("/v2/"+repoPath+"/manifests/"+url.PathEscape(tag), nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("build digest request: %w", err)
	}
	req.Header.Set("Accept", manifestAcceptHeader)
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("request digest: %w", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("digest request failed: status=%d", resp.StatusCode)
	}
	digest := strings.TrimSpace(resp.Header.Get(registryDigestHeaderName))
	if digest == "" {
		return "", fmt.Errorf("registry response does not include %s", registryDigestHeaderName)
	}
	return digest, nil
}

func (c *Client) loadTagMetadata(ctx context.Context, repoPath string, digest string) (*time.Time, int64, error) {
	endpoint := c.buildURL("/v2/"+repoPath+"/manifests/"+url.PathEscape(digest), nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("build manifest request: %w", err)
	}
	req.Header.Set("Accept", manifestAcceptHeader)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request manifest: %w", err)
	}
	body, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if readErr != nil {
		return nil, 0, fmt.Errorf("read manifest response: %w", readErr)
	}
	if closeErr != nil {
		return nil, 0, fmt.Errorf("close manifest response: %w", closeErr)
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("manifest request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var manifest struct {
		Config struct {
			Digest string `json:"digest"`
			Size   int64  `json:"size"`
		} `json:"config"`
	}
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, 0, fmt.Errorf("decode manifest response: %w", err)
	}
	configDigest := strings.TrimSpace(manifest.Config.Digest)
	if configDigest == "" {
		return nil, manifest.Config.Size, nil
	}

	blobURL := c.buildURL("/v2/"+repoPath+"/blobs/"+url.PathEscape(configDigest), nil)
	blobReq, err := http.NewRequestWithContext(ctx, http.MethodGet, blobURL, nil)
	if err != nil {
		return nil, manifest.Config.Size, fmt.Errorf("build blob request: %w", err)
	}
	blobResp, err := c.http.Do(blobReq)
	if err != nil {
		return nil, manifest.Config.Size, fmt.Errorf("request blob: %w", err)
	}
	blobBody, blobReadErr := io.ReadAll(blobResp.Body)
	blobCloseErr := blobResp.Body.Close()
	if blobReadErr != nil {
		return nil, manifest.Config.Size, fmt.Errorf("read blob response: %w", blobReadErr)
	}
	if blobCloseErr != nil {
		return nil, manifest.Config.Size, fmt.Errorf("close blob response: %w", blobCloseErr)
	}
	if blobResp.StatusCode != http.StatusOK {
		if blobResp.StatusCode == http.StatusNotFound {
			return nil, manifest.Config.Size, nil
		}
		return nil, manifest.Config.Size, fmt.Errorf("blob request failed: status=%d body=%s", blobResp.StatusCode, strings.TrimSpace(string(blobBody)))
	}

	var configBlob struct {
		Created string `json:"created"`
	}
	if err := json.Unmarshal(blobBody, &configBlob); err != nil {
		return nil, manifest.Config.Size, nil
	}
	createdRaw := strings.TrimSpace(configBlob.Created)
	if createdRaw == "" {
		return nil, manifest.Config.Size, nil
	}
	createdAt, err := time.Parse(time.RFC3339Nano, createdRaw)
	if err != nil {
		return nil, manifest.Config.Size, nil
	}
	return &createdAt, manifest.Config.Size, nil
}

func (c *Client) buildURL(path string, query url.Values) string {
	u := *c.baseURL
	basePath := strings.TrimRight(u.Path, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u.Path = basePath + path
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	} else {
		u.RawQuery = ""
	}
	return u.String()
}

func encodeRepository(repository string) (string, error) {
	repository = strings.TrimSpace(repository)
	if repository == "" {
		return "", fmt.Errorf("repository is required")
	}
	parts := strings.Split(repository, "/")
	encoded := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return "", fmt.Errorf("repository %q is invalid", repository)
		}
		encoded = append(encoded, url.PathEscape(part))
	}
	return strings.Join(encoded, "/"), nil
}
