package podman

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
)

func (c *Client) ExecCreate(ctx context.Context, containerName string, config ExecConfig) (string, error) {
	path := fmt.Sprintf("/libpod/containers/%s/exec", containerName)
	body, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("marshal exec config: %w", err)
	}

	resp, err := c.post(ctx, path, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("exec create: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		var podmanErr struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &podmanErr); err == nil && podmanErr.Message != "" {
			return "", fmt.Errorf("exec create: %s", podmanErr.Message)
		}
		return "", fmt.Errorf("exec create: unexpected status %d: %s", resp.StatusCode, bytes.TrimSpace(body))
	}

	var result ExecCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode exec create response: %w", err)
	}
	return result.Id, nil
}

func (c *Client) ExecStart(execID string, tty bool) (net.Conn, *bufio.Reader, error) {
	conn, err := c.rawDial()
	if err != nil {
		return nil, nil, fmt.Errorf("raw dial: %w", err)
	}

	startConfig := ExecStartConfig{Detach: false, Tty: tty}
	body, err := json.Marshal(startConfig)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("marshal exec start config: %w", err)
	}

	req := fmt.Sprintf("POST /v4.0.0/libpod/exec/%s/start HTTP/1.1\r\nHost: d\r\nContent-Type: application/json\r\nConnection: Upgrade\r\nUpgrade: tcp\r\nContent-Length: %d\r\n\r\n%s", execID, len(body), body)

	if _, err := conn.Write([]byte(req)); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("write exec start request: %w", err)
	}

	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("read exec start response: %w", err)
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		conn.Close()
		var podmanErr struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &podmanErr); err == nil && podmanErr.Message != "" {
			return nil, nil, fmt.Errorf("exec start: %s", podmanErr.Message)
		}
		return nil, nil, fmt.Errorf("exec start: expected 101, got %d: %s", resp.StatusCode, bytes.TrimSpace(body))
	}

	return conn, reader, nil
}

func (c *Client) ExecResize(ctx context.Context, execID string, height, width int) error {
	path := fmt.Sprintf("/libpod/exec/%s/resize?h=%d&w=%d", execID, height, width)
	resp, err := c.post(ctx, path, nil)
	if err != nil {
		return fmt.Errorf("exec resize: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("exec resize: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) ExecInspect(ctx context.Context, execID string) (*ExecInspectResponse, error) {
	path := fmt.Sprintf("/libpod/exec/%s/json", execID)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("exec inspect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exec inspect: unexpected status %d", resp.StatusCode)
	}

	var result ExecInspectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode exec inspect response: %w", err)
	}
	return &result, nil
}
