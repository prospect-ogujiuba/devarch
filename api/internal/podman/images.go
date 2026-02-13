package podman

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) ListImages(ctx context.Context, all bool) ([]Image, error) {
	path := fmt.Sprintf("/libpod/images/json?all=%t", all)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list images: unexpected status %d", resp.StatusCode)
	}

	var images []Image
	if err := json.NewDecoder(resp.Body).Decode(&images); err != nil {
		return nil, fmt.Errorf("decode images: %w", err)
	}
	return images, nil
}

func (c *Client) InspectImage(ctx context.Context, nameOrID string) (*ImageInspect, error) {
	path := fmt.Sprintf("/libpod/images/%s/json", nameOrID)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("inspect image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("inspect image: unexpected status %d", resp.StatusCode)
	}

	var result ImageInspect
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode image inspect: %w", err)
	}
	return &result, nil
}

func (c *Client) RemoveImage(ctx context.Context, nameOrID string, force bool) error {
	path := fmt.Sprintf("/libpod/images/%s?force=%t", nameOrID, force)
	resp, err := c.delete(ctx, path)
	if err != nil {
		return fmt.Errorf("remove image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remove image: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) PullImage(ctx context.Context, reference string, callback func(ImagePullReport)) error {
	path := fmt.Sprintf("/libpod/images/pull?reference=%s", reference)
	resp, err := c.post(ctx, path, nil)
	if err != nil {
		return fmt.Errorf("pull image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pull image: unexpected status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var report ImagePullReport
		if err := json.Unmarshal(scanner.Bytes(), &report); err != nil {
			continue
		}
		if callback != nil {
			callback(report)
		}
	}
	return scanner.Err()
}

func (c *Client) ImageHistory(ctx context.Context, nameOrID string) ([]ImageHistoryEntry, error) {
	path := fmt.Sprintf("/libpod/images/%s/history", nameOrID)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("image history: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("image history: unexpected status %d", resp.StatusCode)
	}

	var entries []ImageHistoryEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("decode image history: %w", err)
	}
	return entries, nil
}

func (c *Client) PruneImages(ctx context.Context, dangling bool) ([]ImagePruneReport, error) {
	path := fmt.Sprintf("/libpod/images/prune?dangling=%t", dangling)
	resp, err := c.post(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("prune images: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prune images: unexpected status %d", resp.StatusCode)
	}

	var reports []ImagePruneReport
	if err := json.NewDecoder(resp.Body).Decode(&reports); err != nil {
		return nil, fmt.Errorf("decode prune reports: %w", err)
	}
	return reports, nil
}
