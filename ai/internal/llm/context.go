package llm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ContextBuilder struct {
	apiURL string
	apiKey string
	http   *http.Client
}

func NewContextBuilder(apiURL, apiKey string) *ContextBuilder {
	return &ContextBuilder{
		apiURL: apiURL,
		apiKey: apiKey,
		http:   &http.Client{},
	}
}

func (cb *ContextBuilder) ServiceAuthorContext() string {
	var parts []string

	cats := cb.fetchNames("/api/v1/categories", "data", "name")
	if len(cats) > 0 {
		parts = append(parts, "Available categories: "+strings.Join(cats, ", "))
	}

	names := cb.fetchNames("/api/v1/services", "data", "name")
	if len(names) > 0 {
		parts = append(parts, "Existing services (avoid name collisions): "+strings.Join(names, ", "))
	}

	if len(parts) == 0 {
		return ""
	}
	return "Context:\n" + strings.Join(parts, "\n")
}

func (cb *ContextBuilder) CLIAssistantContext() string {
	var parts []string

	status := cb.fetchJSON("/api/v1/status")
	if status != nil {
		if data, ok := status["data"].(map[string]interface{}); ok {
			if running, ok := data["running"].(float64); ok {
				stopped, _ := data["stopped"].(float64)
				parts = append(parts, fmt.Sprintf("Running containers: %.0f, Stopped: %.0f", running, stopped))
			}
		}
	}

	names := cb.fetchNames("/api/v1/services", "data", "name")
	if len(names) > 0 {
		if len(names) > 20 {
			names = names[:20]
		}
		parts = append(parts, "Services: "+strings.Join(names, ", "))
	}

	if len(parts) == 0 {
		return ""
	}
	return "Current environment:\n" + strings.Join(parts, "\n")
}

func (cb *ContextBuilder) DebugContext(target string) string {
	var parts []string

	// Container status via API
	status := cb.fetchJSON("/api/v1/services/" + target + "/status")
	if status != nil {
		if data, ok := status["data"].(map[string]interface{}); ok {
			if s, ok := data["status"].(string); ok {
				health, _ := data["health"].(string)
				parts = append(parts, fmt.Sprintf("Container status: %s (health: %s)", s, health))
			}
			if e, ok := data["error"].(string); ok && e != "" {
				parts = append(parts, "Error: "+e)
			}
		}
	}

	// Logs via API
	logs := cb.fetchJSON("/api/v1/services/" + target + "/logs?tail=50")
	if logs != nil {
		if data, ok := logs["data"].(map[string]interface{}); ok {
			if l, ok := data["logs"].(string); ok && l != "" {
				parts = append(parts, "Recent logs:\n"+l)
			}
		}
	}

	if len(parts) == 0 {
		return "No diagnostic data available for target: " + target
	}
	return strings.Join(parts, "\n\n")
}

func (cb *ContextBuilder) fetchJSON(path string) map[string]interface{} {
	req, err := http.NewRequest("GET", cb.apiURL+path, nil)
	if err != nil {
		return nil
	}
	if cb.apiKey != "" {
		req.Header.Set("X-API-Key", cb.apiKey)
	}

	resp, err := cb.http.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var result map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	json.Unmarshal(body, &result)
	return result
}

func (cb *ContextBuilder) FetchServices() []map[string]interface{} {
	result := cb.fetchJSON("/api/v1/services")
	if result == nil {
		return nil
	}

	data, ok := result["data"]
	if !ok {
		return nil
	}

	items, ok := data.([]interface{})
	if !ok {
		return nil
	}

	var services []map[string]interface{}
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			services = append(services, m)
		}
	}
	return services
}

func (cb *ContextBuilder) fetchNames(path, dataKey, nameKey string) []string {
	result := cb.fetchJSON(path)
	if result == nil {
		return nil
	}

	data, ok := result[dataKey]
	if !ok {
		return nil
	}

	items, ok := data.([]interface{})
	if !ok {
		return nil
	}

	var names []string
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			if name, ok := m[nameKey].(string); ok {
				names = append(names, name)
			}
		}
	}
	return names
}
