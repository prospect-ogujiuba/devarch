package dockerhub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/priz/devarch-api/pkg/registry"
)

const (
	hubAPIURL      = "https://hub.docker.com/v2"
	registryAPIURL = "https://registry.hub.docker.com/v2"
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
	return "dockerhub"
}

func (c *Client) Authenticate(ctx context.Context) error {
	return nil
}

func (c *Client) RateLimitRemaining() int {
	return c.rateLimitRemaining
}

type repoResponse struct {
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	Description     string `json:"description"`
	StarCount       int    `json:"star_count"`
	PullCount       int64  `json:"pull_count"`
	IsPrivate       bool   `json:"is_private"`
	IsAutomated     bool   `json:"is_automated"`
	LastUpdated     string `json:"last_updated"`
}

func (c *Client) GetImageInfo(ctx context.Context, repository string) (*registry.ImageInfo, error) {
	url := fmt.Sprintf("%s/repositories/%s", hubAPIURL, repository)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("repository not found: %s", repository)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var repoResp repoResponse
	if err := json.NewDecoder(resp.Body).Decode(&repoResp); err != nil {
		return nil, err
	}

	info := &registry.ImageInfo{
		Repository:  repository,
		Description: repoResp.Description,
		StarCount:   repoResp.StarCount,
		PullCount:   repoResp.PullCount,
		IsOfficial:  repoResp.Namespace == "library",
	}

	if repoResp.LastUpdated != "" {
		if t, err := time.Parse(time.RFC3339, repoResp.LastUpdated); err == nil {
			info.LastUpdated = t
		}
	}

	return info, nil
}

type tagsResponse struct {
	Count   int `json:"count"`
	Results []struct {
		Name        string `json:"name"`
		FullSize    int64  `json:"full_size"`
		LastUpdated string `json:"last_updated"`
		Digest      string `json:"digest"`
		Images      []struct {
			Architecture string `json:"architecture"`
			OS           string `json:"os"`
			Variant      string `json:"variant"`
			Digest       string `json:"digest"`
			Size         int64  `json:"size"`
		} `json:"images"`
	} `json:"results"`
}

func (c *Client) ListTags(ctx context.Context, repository string, opts registry.ListTagsOptions) ([]registry.TagInfo, error) {
	pageSize := opts.PageSize
	if pageSize == 0 {
		pageSize = 25
	}
	page := opts.Page
	if page == 0 {
		page = 1
	}

	url := fmt.Sprintf("%s/repositories/%s/tags?page_size=%d&page=%d", hubAPIURL, repository, pageSize, page)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("repository not found: %s", repository)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var tagsResp tagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, err
	}

	tags := make([]registry.TagInfo, len(tagsResp.Results))
	for i, t := range tagsResp.Results {
		tag := registry.TagInfo{
			Name:      t.Name,
			Digest:    t.Digest,
			SizeBytes: t.FullSize,
		}

		if t.LastUpdated != "" {
			if pt, err := time.Parse(time.RFC3339, t.LastUpdated); err == nil {
				tag.PushedAt = pt
			}
		}

		for _, img := range t.Images {
			tag.Architectures = append(tag.Architectures, registry.ArchInfo{
				OS:           img.OS,
				Architecture: img.Architecture,
				Variant:      img.Variant,
				Digest:       img.Digest,
				SizeBytes:    img.Size,
			})
		}

		tags[i] = tag
	}

	return tags, nil
}

func (c *Client) GetTagManifest(ctx context.Context, repository, tag string) (*registry.TagInfo, error) {
	tags, err := c.ListTags(ctx, repository, registry.ListTagsOptions{PageSize: 100})
	if err != nil {
		return nil, err
	}

	for _, t := range tags {
		if t.Name == tag {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("tag not found: %s", tag)
}
