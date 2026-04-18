package spec

import (
	"fmt"
	"os"
	"sort"
	"strings"

	embeddedschemas "github.com/prospect-ogujiuba/devarch/schemas"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

const (
	ManifestFilename    = "devarch.workspace.yaml"
	WorkspaceSchemaFile = "workspace.schema.json"
	TemplateSchemaFile  = "template.schema.json"
	PlanSchemaFile      = "plan.schema.json"
)

// ValidationError captures one schema validation failure.
type ValidationError struct {
	Field   string
	Message string
}

// ValidationErrors groups schema validation failures for one document.
type ValidationErrors struct {
	Schema string
	Errors []ValidationError
}

func (e *ValidationErrors) Error() string {
	if e == nil || len(e.Errors) == 0 {
		return "document validation failed"
	}

	first := e.Errors[0]
	if first.Field == "" {
		return fmt.Sprintf("document validation failed for %s: %s", e.Schema, first.Message)
	}

	return fmt.Sprintf("document validation failed for %s at %s: %s", e.Schema, first.Field, first.Message)
}

// LoadSchema reads an embedded schema file from the root schemas package.
func LoadSchema(name string) ([]byte, error) {
	data, err := embeddedschemas.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read embedded schema %s: %w", name, err)
	}
	return data, nil
}

// ValidateWorkspaceBytes validates a workspace manifest document.
func ValidateWorkspaceBytes(data []byte) error {
	return validateDocument(WorkspaceSchemaFile, data)
}

// ValidateWorkspaceFile validates a workspace manifest file from disk.
func ValidateWorkspaceFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read workspace file %s: %w", path, err)
	}
	return ValidateWorkspaceBytes(data)
}

// ValidateTemplateBytes validates a template document.
func ValidateTemplateBytes(data []byte) error {
	return validateDocument(TemplateSchemaFile, data)
}

// ValidateTemplateFile validates a template file from disk.
func ValidateTemplateFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read template file %s: %w", path, err)
	}
	return ValidateTemplateBytes(data)
}

// ValidatePlanBytes validates a plan document.
func ValidatePlanBytes(data []byte) error {
	return validateDocument(PlanSchemaFile, data)
}

// ValidatePlanFile validates a plan file from disk.
func ValidatePlanFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read plan file %s: %w", path, err)
	}
	return ValidatePlanBytes(data)
}

func validateDocument(schemaName string, data []byte) error {
	document, err := decodeDocument(data)
	if err != nil {
		return err
	}

	schemaBytes, err := LoadSchema(schemaName)
	if err != nil {
		return err
	}

	result, err := gojsonschema.Validate(
		gojsonschema.NewStringLoader(string(schemaBytes)),
		gojsonschema.NewGoLoader(document),
	)
	if err != nil {
		return fmt.Errorf("validate document against %s: %w", schemaName, err)
	}
	if result.Valid() {
		return nil
	}

	validationErrors := make([]ValidationError, 0, len(result.Errors()))
	for _, resultErr := range result.Errors() {
		field := strings.TrimPrefix(resultErr.Field(), "(root).")
		if field == "(root)" {
			field = ""
		}

		validationErrors = append(validationErrors, ValidationError{
			Field:   field,
			Message: resultErr.Description(),
		})
	}

	sort.Slice(validationErrors, func(i, j int) bool {
		if validationErrors[i].Field != validationErrors[j].Field {
			return validationErrors[i].Field < validationErrors[j].Field
		}
		return validationErrors[i].Message < validationErrors[j].Message
	})

	return &ValidationErrors{
		Schema: schemaName,
		Errors: validationErrors,
	}
}

func decodeDocument(data []byte) (any, error) {
	var document any
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("decode document: %w", err)
	}
	return document, nil
}
