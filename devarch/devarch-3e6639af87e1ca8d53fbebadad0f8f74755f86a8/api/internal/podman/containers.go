package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func (c *Client) ListContainers(ctx context.Context, all bool) ([]Container, error) {
	path := "/libpod/containers/json"
	if all {
		path += "?all=true"
	}

	var containers []Container
	if err := c.getJSON(ctx, path, &containers); err != nil {
		return nil, err
	}
	return containers, nil
}

func (c *Client) InspectContainer(ctx context.Context, nameOrID string) (*ContainerInspect, error) {
	path := fmt.Sprintf("/libpod/containers/%s/json", url.PathEscape(nameOrID))

	var inspect ContainerInspect
	if err := c.getJSON(ctx, path, &inspect); err != nil {
		return nil, err
	}
	return &inspect, nil
}

func (c *Client) BatchInspect(ctx context.Context, names []string) (map[string]*ContainerInspect, error) {
	result := make(map[string]*ContainerInspect, len(names))

	for _, name := range names {
		inspect, err := c.InspectContainer(ctx, name)
		if err != nil {
			continue
		}
		result[name] = inspect
	}

	return result, nil
}

func (c *Client) StartContainer(ctx context.Context, nameOrID string) error {
	path := fmt.Sprintf("/libpod/containers/%s/start", url.PathEscape(nameOrID))
	return c.postAction(ctx, path)
}

func (c *Client) StopContainer(ctx context.Context, nameOrID string, timeout int) error {
	path := fmt.Sprintf("/libpod/containers/%s/stop?timeout=%d", url.PathEscape(nameOrID), timeout)
	return c.postAction(ctx, path)
}

func (c *Client) RestartContainer(ctx context.Context, nameOrID string, timeout int) error {
	path := fmt.Sprintf("/libpod/containers/%s/restart?timeout=%d", url.PathEscape(nameOrID), timeout)
	return c.postAction(ctx, path)
}

func (c *Client) KillContainer(ctx context.Context, nameOrID string, signal string) error {
	path := fmt.Sprintf("/libpod/containers/%s/kill", url.PathEscape(nameOrID))
	if signal != "" {
		path += "?signal=" + signal
	}
	return c.postAction(ctx, path)
}

func (c *Client) RemoveContainer(ctx context.Context, nameOrID string, force bool, volumes bool) error {
	path := fmt.Sprintf("/libpod/containers/%s?force=%t&v=%t", url.PathEscape(nameOrID), force, volumes)

	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+path, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remove failed %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

type LogOptions struct {
	Follow     bool
	Stdout     bool
	Stderr     bool
	Since      time.Time
	Until      time.Time
	Timestamps bool
	Tail       int
}

func (c *Client) ContainerLogs(ctx context.Context, nameOrID string, opts LogOptions) (io.ReadCloser, error) {
	params := url.Values{}
	params.Set("follow", fmt.Sprintf("%t", opts.Follow))
	params.Set("stdout", fmt.Sprintf("%t", opts.Stdout))
	params.Set("stderr", fmt.Sprintf("%t", opts.Stderr))
	params.Set("timestamps", fmt.Sprintf("%t", opts.Timestamps))

	if !opts.Since.IsZero() {
		params.Set("since", opts.Since.Format(time.RFC3339))
	}
	if !opts.Until.IsZero() {
		params.Set("until", opts.Until.Format(time.RFC3339))
	}
	if opts.Tail > 0 {
		params.Set("tail", fmt.Sprintf("%d", opts.Tail))
	}

	path := fmt.Sprintf("/libpod/containers/%s/logs?%s", url.PathEscape(nameOrID), params.Encode())

	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("logs failed %d: %s", resp.StatusCode, string(body))
	}

	return resp.Body, nil
}

func (c *Client) ContainerLogsString(ctx context.Context, nameOrID string, tail int) (string, error) {
	reader, err := c.ContainerLogs(ctx, nameOrID, LogOptions{
		Stdout: true,
		Stderr: true,
		Tail:   tail,
	})
	if err != nil {
		return "", err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Client) ContainerStats(ctx context.Context, nameOrID string) (*ContainerStats, error) {
	path := fmt.Sprintf("/libpod/containers/%s/stats?stream=false", url.PathEscape(nameOrID))

	var stats ContainerStats
	if err := c.getJSON(ctx, path, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

type StatsCallback func(stats *ContainerStats, err error) bool

func (c *Client) StreamContainerStats(ctx context.Context, nameOrID string, callback StatsCallback) error {
	path := fmt.Sprintf("/libpod/containers/%s/stats?stream=true", url.PathEscape(nameOrID))

	resp, err := c.get(ctx, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stats stream failed %d: %s", resp.StatusCode, string(body))
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var stats ContainerStats
		if err := decoder.Decode(&stats); err != nil {
			if err == io.EOF {
				return nil
			}
			if !callback(nil, err) {
				return nil
			}
			continue
		}

		if !callback(&stats, nil) {
			return nil
		}
	}
}

func (c *Client) AllContainerStats(ctx context.Context, names []string) (map[string]*ContainerStats, error) {
	result := make(map[string]*ContainerStats, len(names))

	for _, name := range names {
		stats, err := c.ContainerStats(ctx, name)
		if err != nil {
			continue
		}
		result[name] = stats
	}

	return result, nil
}

type ContainerHealth struct {
	Name         string
	Status       string
	Health       string
	RestartCount int
	StartedAt    *time.Time
	FinishedAt   *time.Time
	ExitCode     *int
	Error        string
}

func (c *Client) GetContainerHealth(ctx context.Context, nameOrID string) (*ContainerHealth, error) {
	inspect, err := c.InspectContainer(ctx, nameOrID)
	if err != nil {
		return nil, err
	}

	health := &ContainerHealth{
		Name:         inspect.Name,
		Status:       inspect.State.Status,
		RestartCount: inspect.RestartCount,
		Error:        inspect.State.Error,
	}

	if inspect.State.Health != nil {
		health.Health = inspect.State.Health.Status
	}

	if inspect.State.StartedAt != "" && inspect.State.StartedAt != "0001-01-01T00:00:00Z" {
		if t, err := time.Parse(time.RFC3339Nano, inspect.State.StartedAt); err == nil {
			health.StartedAt = &t
		}
	}

	if inspect.State.FinishedAt != "" && inspect.State.FinishedAt != "0001-01-01T00:00:00Z" {
		if t, err := time.Parse(time.RFC3339Nano, inspect.State.FinishedAt); err == nil {
			health.FinishedAt = &t
		}
	}

	if inspect.State.ExitCode != 0 {
		ec := inspect.State.ExitCode
		health.ExitCode = &ec
	}

	return health, nil
}

func (c *Client) Top(ctx context.Context, nameOrID string, psArgs string) ([][]string, error) {
	path := fmt.Sprintf("/libpod/containers/%s/top", url.PathEscape(nameOrID))
	if psArgs != "" {
		path += "?ps_args=" + url.QueryEscape(psArgs)
	}

	var result struct {
		Titles    []string   `json:"Titles"`
		Processes [][]string `json:"Processes"`
	}

	if err := c.getJSON(ctx, path, &result); err != nil {
		return nil, err
	}

	return append([][]string{result.Titles}, result.Processes...), nil
}

func (c *Client) WaitContainer(ctx context.Context, nameOrID string, condition string) (int, error) {
	path := fmt.Sprintf("/libpod/containers/%s/wait", url.PathEscape(nameOrID))
	if condition != "" {
		path += "?condition=" + condition
	}

	resp, err := c.post(ctx, path, nil)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	var result struct {
		StatusCode int    `json:"StatusCode"`
		Error      string `json:"Error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return -1, err
	}

	if result.Error != "" {
		return result.StatusCode, fmt.Errorf(result.Error)
	}

	return result.StatusCode, nil
}
