package embedding

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Notifier struct {
	aiURL  string
	client *http.Client
}

func NewNotifier() *Notifier {
	aiURL := os.Getenv("DEVARCH_AI_URL")
	if aiURL == "" {
		return nil
	}
	return &Notifier{
		aiURL: aiURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (n *Notifier) IndexService(name string, data map[string]interface{}) {
	if n == nil {
		return
	}
	go n.index("service", name, data)
}

func (n *Notifier) DeleteService(name string) {
	if n == nil {
		return
	}
	go n.delete("service", name)
}

func (n *Notifier) index(sourceType, sourceID string, data map[string]interface{}) {
	content, err := json.Marshal(data)
	if err != nil {
		slog.Warn("embedding notifier: marshal failed", "error", err)
		return
	}

	body, _ := json.Marshal(map[string]string{
		"source_type": sourceType,
		"source_id":   sourceID,
		"content":     string(content),
	})

	resp, err := n.client.Post(n.aiURL+"/api/v1/ai/embed/index", "application/json", bytes.NewReader(body))
	if err != nil {
		slog.Warn("embedding notifier: index request failed", "source", sourceID, "error", err)
		return
	}
	resp.Body.Close()
	slog.Debug("embedding notifier: indexed", "source_type", sourceType, "source_id", sourceID, "status", resp.StatusCode)
}

func (n *Notifier) delete(sourceType, sourceID string) {
	body, _ := json.Marshal(map[string]string{
		"source_type": sourceType,
		"source_id":   sourceID,
	})

	req, err := http.NewRequest(http.MethodDelete, n.aiURL+"/api/v1/ai/embed/index", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		slog.Warn("embedding notifier: delete request failed", "source", sourceID, "error", err)
		return
	}
	resp.Body.Close()
	slog.Debug("embedding notifier: deleted", "source_type", sourceType, "source_id", sourceID, "status", resp.StatusCode)
}
