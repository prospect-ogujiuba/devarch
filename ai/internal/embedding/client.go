package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
}

func (c *Client) Embed(texts []string) ([][]float32, error) {
	if err := c.manager.EnsureRunning(); err != nil {
		return nil, fmt.Errorf("ensure embed container running: %w", err)
	}
	c.manager.TouchLastUse()

	cfg := c.manager.Config()
	req := embeddingRequest{
		Model: cfg.Model,
		Input: texts,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Post(c.manager.BaseURL()+"/v1/embeddings", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embedding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding error (status %d): %s", resp.StatusCode, string(b))
	}

	var result embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
	}

	embeddings := make([][]float32, len(result.Data))
	for _, d := range result.Data {
		embeddings[d.Index] = d.Embedding
	}
	return embeddings, nil
}

func (c *Client) EmbedSingle(text string) ([]float32, error) {
	results, err := c.Embed([]string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return results[0], nil
}
