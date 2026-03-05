package llm

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/priz/devarch-api/internal/container"
)

type ContextBuilder struct {
	db              *sql.DB
	containerClient *container.Client
}

func NewContextBuilder(db *sql.DB, containerClient *container.Client) *ContextBuilder {
	return &ContextBuilder{db: db, containerClient: containerClient}
}

func (cb *ContextBuilder) ServiceAuthorContext() string {
	var parts []string

	// Categories
	cats := cb.getCategories()
	if len(cats) > 0 {
		parts = append(parts, "Available categories: "+strings.Join(cats, ", "))
	}

	// Existing service names (collision avoidance)
	names := cb.getServiceNames()
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

	if cb.containerClient != nil {
		running, stopped := cb.containerClient.GetRunningCount()
		parts = append(parts, fmt.Sprintf("Running containers: %d, Stopped: %d", running, stopped))
	}

	stacks := cb.getStackNames()
	if len(stacks) > 0 {
		parts = append(parts, "Active stacks: "+strings.Join(stacks, ", "))
	}

	names := cb.getServiceNames()
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

	// Container status
	if cb.containerClient != nil {
		status, err := cb.containerClient.GetContainerStatus(target)
		if err == nil {
			parts = append(parts, fmt.Sprintf("Container status: %s (health: %s)", status.Status, status.Health))
			if status.Error != "" {
				parts = append(parts, "Error: "+status.Error)
			}
		}

		logs, err := cb.containerClient.GetLogs(target, "50")
		if err == nil && logs != "" {
			parts = append(parts, "Recent logs:\n"+logs)
		}
	}

	// Service config from DB
	svcConfig := cb.getServiceConfig(target)
	if svcConfig != "" {
		parts = append(parts, "Service config:\n"+svcConfig)
	}

	if len(parts) == 0 {
		return "No diagnostic data available for target: " + target
	}
	return strings.Join(parts, "\n\n")
}

func (cb *ContextBuilder) getCategories() []string {
	rows, err := cb.db.Query("SELECT name FROM categories ORDER BY name")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var cats []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			cats = append(cats, name)
		}
	}
	return cats
}

func (cb *ContextBuilder) getServiceNames() []string {
	rows, err := cb.db.Query("SELECT name FROM services ORDER BY name LIMIT 50")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			names = append(names, name)
		}
	}
	return names
}

func (cb *ContextBuilder) getStackNames() []string {
	rows, err := cb.db.Query("SELECT name FROM stacks WHERE enabled = true ORDER BY name LIMIT 20")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			names = append(names, name)
		}
	}
	return names
}

func (cb *ContextBuilder) getServiceConfig(name string) string {
	var image string
	var ports, env sql.NullString
	err := cb.db.QueryRow("SELECT image, ports, environment FROM services WHERE name = $1", name).Scan(&image, &ports, &env)
	if err != nil {
		return ""
	}

	var parts []string
	parts = append(parts, "Image: "+image)
	if ports.Valid && ports.String != "" {
		parts = append(parts, "Ports: "+ports.String)
	}
	if env.Valid && env.String != "" {
		parts = append(parts, "Environment: "+env.String)
	}
	return strings.Join(parts, "\n")
}
