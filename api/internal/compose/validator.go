package compose

import (
	"database/sql"
	"fmt"

	"github.com/priz/devarch-api/pkg/models"
)

type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Validator struct {
	db *sql.DB
}

func NewValidator(db *sql.DB) *Validator {
	return &Validator{db: db}
}

func (v *Validator) Validate(service *models.Service) *ValidationResult {
	result := &ValidationResult{Valid: true}

	v.validateRequired(service, result)
	v.validatePortConflicts(service, result)
	v.validateDependencyCycles(service, result)

	result.Valid = len(result.Errors) == 0
	return result
}

func (v *Validator) validateRequired(service *models.Service, result *ValidationResult) {
	if service.ImageName == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "image_name",
			Message: "image name is required",
		})
	}
	if service.ImageTag == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "image_tag",
			Message: "image tag is required",
		})
	}
	if service.Name == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "name",
			Message: "service name is required",
		})
	}
}

func (v *Validator) validatePortConflicts(service *models.Service, result *ValidationResult) {
	for _, port := range service.Ports {
		if port.HostPort == 0 {
			continue
		}
		var conflictName string
		err := v.db.QueryRow(`
			SELECT s.name FROM service_ports p
			JOIN services s ON p.service_id = s.id
			WHERE p.host_port = $1 AND p.host_ip = $2 AND s.id != $3
			LIMIT 1
		`, port.HostPort, port.HostIP, service.ID).Scan(&conflictName)
		if err == nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "ports",
				Message: fmt.Sprintf("host port %d conflicts with service %s", port.HostPort, conflictName),
			})
		}
	}
}

func (v *Validator) validateDependencyCycles(service *models.Service, result *ValidationResult) {
	if len(service.Dependencies) == 0 {
		return
	}

	// BFS cycle detection from this service's dependencies
	visited := map[string]bool{service.Name: true}
	queue := make([]string, len(service.Dependencies))
	copy(queue, service.Dependencies)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "dependencies",
				Message: fmt.Sprintf("circular dependency detected involving %s", current),
			})
			return
		}
		visited[current] = true

		rows, err := v.db.Query(`
			SELECT s2.name FROM service_dependencies d
			JOIN services s1 ON d.service_id = s1.id
			JOIN services s2 ON d.depends_on_service_id = s2.id
			WHERE s1.name = $1
		`, current)
		if err != nil {
			continue
		}
		for rows.Next() {
			var dep string
			rows.Scan(&dep)
			queue = append(queue, dep)
		}
		rows.Close()
	}
}
