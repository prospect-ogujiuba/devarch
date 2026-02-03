package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

type Client struct {
	httpClient *http.Client
	socketPath string
	baseURL    string
}

var defaultSocketPaths = []string{
	"/run/podman/podman.sock",
	"/var/run/docker.sock",
}

func NewClient() (*Client, error) {
	socketPath := os.Getenv("CONTAINER_HOST")
	if socketPath != "" {
		if len(socketPath) > 7 && socketPath[:7] == "unix://" {
			socketPath = socketPath[7:]
		}
	}

	if socketPath == "" {
		for _, path := range defaultSocketPaths {
			if _, err := os.Stat(path); err == nil {
				socketPath = path
				break
			}
		}
	}

	if socketPath == "" {
		uid := os.Getuid()
		rootlessPath := fmt.Sprintf("/run/user/%d/podman/podman.sock", uid)
		if _, err := os.Stat(rootlessPath); err == nil {
			socketPath = rootlessPath
		}
	}

	if socketPath == "" {
		return nil, fmt.Errorf("no container socket found")
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
		DisableKeepAlives: false,
		MaxIdleConns:      10,
		IdleConnTimeout:   90 * time.Second,
	}

	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		socketPath: socketPath,
		baseURL:    "http://d/v4.0.0",
	}, nil
}

func NewClientWithSocket(socketPath string) (*Client, error) {
	if _, err := os.Stat(socketPath); err != nil {
		return nil, fmt.Errorf("socket not found: %s", socketPath)
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
		DisableKeepAlives: false,
		MaxIdleConns:      10,
		IdleConnTimeout:   90 * time.Second,
	}

	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		socketPath: socketPath,
		baseURL:    "http://d/v4.0.0",
	}, nil
}

func (c *Client) SocketPath() string {
	return c.socketPath
}

func (c *Client) get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	return c.httpClient.Do(req)
}

func (c *Client) getJSON(ctx context.Context, path string, v interface{}) error {
	resp, err := c.get(ctx, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

func (c *Client) post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.httpClient.Do(req)
}

func (c *Client) postAction(ctx context.Context, path string) error {
	resp, err := c.post(ctx, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	resp, err := c.get(ctx, "/libpod/_ping")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping failed with status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) Info(ctx context.Context) (*SystemInfo, error) {
	var info SystemInfo
	if err := c.getJSON(ctx, "/libpod/info", &info); err != nil {
		return nil, err
	}
	return &info, nil
}
