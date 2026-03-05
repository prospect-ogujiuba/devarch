package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/priz/devarch-ai/internal/ramalama"
)

type Client struct {
	manager *ramalama.Manager
	http    *http.Client
}

func NewClient(manager *ramalama.Manager) *Client {
	return &Client{
		manager: manager,
		http:    &http.Client{},
	}
}

func (c *Client) Complete(messages []ChatMessage) (*ChatResponse, error) {
	if err := c.manager.EnsureRunning(); err != nil {
		return nil, fmt.Errorf("ensure LLM running: %w", err)
	}
	c.manager.TouchLastUse()

	cfg := c.manager.Config()
	req := ChatRequest{
		Model:       cfg.Model,
		Messages:    messages,
		MaxTokens:   cfg.MaxTokens,
		Temperature: 0.7,
		Stream:      false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Post(c.manager.BaseURL()+"/v1/chat/completions", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("LLM request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LLM error (status %d): %s", resp.StatusCode, string(b))
	}

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode LLM response: %w", err)
	}

	return &result, nil
}

func (c *Client) StreamComplete(messages []ChatMessage, onChunk func(StreamChunk)) error {
	if err := c.manager.EnsureRunning(); err != nil {
		return fmt.Errorf("ensure LLM running: %w", err)
	}
	c.manager.TouchLastUse()

	cfg := c.manager.Config()
	req := ChatRequest{
		Model:       cfg.Model,
		Messages:    messages,
		MaxTokens:   cfg.MaxTokens,
		Temperature: 0.7,
		Stream:      true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.http.Post(c.manager.BaseURL()+"/v1/chat/completions", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("LLM stream request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LLM error (status %d): %s", resp.StatusCode, string(b))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		onChunk(chunk)
	}

	return scanner.Err()
}
