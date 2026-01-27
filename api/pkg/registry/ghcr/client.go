package ghcr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/priz/devarch-api/pkg/registry"
)

const (
	ghcrAPIURL = "https://ghcr.io/v2"
	tokenURL   = "https://ghcr.io/token"
)

type Client struct {
	httpClient         *http.Client
	token              string
	rateLimitRemaining int
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimitRemaining: -1,
	}
}

func (c *Client) Name() string {
	return "ghcr"
}

func (c *Client) RateLimitRemaining() int {
	return c.rateLimitRemaining
}

func (c *Client) Authenticate(ctx context.Context) error {
	return nil
}

func (c *Client) getToken(ctx context.Context, repository string) (string, error) {
	url := fmt.Sprintf("%s?scope=repository:%s:pull&service=ghcr.io", tokenURL, repository)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get token: %d", resp.StatusCode)
	}

	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.Token, nil
}

func (c *Client) GetImageInfo(ctx context.Context, repository string) (*registry.ImageInfo, error) {
	return &registry.ImageInfo{
		Repository: repository,
	}, nil
}

func (c *Client) ListTags(ctx context.Context, repository string, opts registry.ListTagsOptions) ([]registry.TagInfo, error) {
	token, err := c.getToken(ctx, repository)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s/tags/list", ghcrAPIURL, repository)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if limit := resp.Header.Get("X-RateLimit-Remaining"); limit != "" {
		if v, err := strconv.Atoi(limit); err == nil {
			c.rateLimitRemaining = v
		}
	}

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("repository not found: %s", repository)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var listResp struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	tags := make([]registry.TagInfo, len(listResp.Tags))
	for i, t := range listResp.Tags {
		tags[i] = registry.TagInfo{
			Name: t,
		}
	}

	return tags, nil
}

func (c *Client) GetTagManifest(ctx context.Context, repository, tag string) (*registry.TagInfo, error) {
	token, err := c.getToken(ctx, repository)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s/manifests/%s", ghcrAPIURL, repository, tag)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("tag not found: %s", tag)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	digest := resp.Header.Get("Docker-Content-Digest")

	var manifest struct {
		Config struct {
			Size   int64  `json:"size"`
			Digest string `json:"digest"`
		} `json:"config"`
		Layers []struct {
			Size int64 `json:"size"`
		} `json:"layers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}

	var totalSize int64
	for _, layer := range manifest.Layers {
		totalSize += layer.Size
	}

	return &registry.TagInfo{
		Name:      tag,
		Digest:    digest,
		SizeBytes: totalSize,
	}, nil
}
