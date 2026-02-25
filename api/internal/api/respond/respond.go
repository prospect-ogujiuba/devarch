package respond

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

// JSON wraps data in a success envelope and encodes to response
func JSON(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	envelope := SuccessEnvelope{Data: data}
	if err := json.NewEncoder(w).Encode(envelope); err != nil {
		slog.Error("error encoding response", "error", err)
	}
}

// Error wraps error details in an error envelope and encodes to response
func Error(w http.ResponseWriter, r *http.Request, statusCode int, code, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	envelope := ErrorEnvelope{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	if err := json.NewEncoder(w).Encode(envelope); err != nil {
		slog.Error("error encoding error response", "error", err)
	}
}

// BadRequest returns a 400 bad request error
func BadRequest(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, http.StatusBadRequest, "bad_request", message, nil)
}

// NotFound returns a 404 not found error
func NotFound(w http.ResponseWriter, r *http.Request, resource, identifier string) {
	message := fmt.Sprintf("%s '%s' not found", resource, identifier)
	Error(w, r, http.StatusNotFound, "not_found", message, nil)
}

// InternalError returns a 500 internal server error and logs the full error
func InternalError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("internal error", "error", err)
	Error(w, r, http.StatusInternalServerError, "internal_error", "An internal error occurred", nil)
}

// ContainerError returns a 500 error with the sanitized container error message visible to the client.
func ContainerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("container error", "error", err)
	Error(w, r, http.StatusInternalServerError, "container_error", sanitizeContainerError(err.Error()), nil)
}

// sanitizeContainerError strips lines that look like env var assignments to avoid leaking secrets.
func sanitizeContainerError(raw string) string {
	var safe []string
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "=") && !strings.HasPrefix(trimmed, "exit") {
			continue
		}
		safe = append(safe, trimmed)
	}
	if len(safe) == 0 {
		return "Container operation failed"
	}
	return strings.Join(safe, "; ")
}

// Conflict returns a 409 conflict error
func Conflict(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, http.StatusConflict, "conflict", message, nil)
}

// Unauthorized returns a 401 unauthorized error
func Unauthorized(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, http.StatusUnauthorized, "unauthorized", message, nil)
}

// Forbidden returns a 403 forbidden error
func Forbidden(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, http.StatusForbidden, "forbidden", message, nil)
}

// NoContent returns a 204 no content response
func NoContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// ValidationError returns a 400 validation error with details
func ValidationError(w http.ResponseWriter, r *http.Request, message string, details interface{}) {
	Error(w, r, http.StatusBadRequest, "validation_error", message, details)
}

// Action returns an action response wrapped in success envelope.
func Action(w http.ResponseWriter, r *http.Request, statusCode int, status string, opts ...func(*ActionResponse)) {
	resp := &ActionResponse{Status: status}
	for _, opt := range opts {
		opt(resp)
	}
	JSON(w, r, statusCode, resp)
}

// WithMessage sets the message field on an ActionResponse.
func WithMessage(msg string) func(*ActionResponse) {
	return func(r *ActionResponse) { r.Message = msg }
}

// WithOutput sets the output field on an ActionResponse.
func WithOutput(output string) func(*ActionResponse) {
	return func(r *ActionResponse) { r.Output = output }
}

// WithWarnings sets the warnings field on an ActionResponse.
func WithWarnings(warnings []string) func(*ActionResponse) {
	return func(r *ActionResponse) { r.Warnings = warnings }
}

// WithMetadata adds a key-value pair to the metadata map.
func WithMetadata(key string, value interface{}) func(*ActionResponse) {
	return func(r *ActionResponse) {
		if r.Metadata == nil {
			r.Metadata = make(map[string]interface{})
		}
		r.Metadata[key] = value
	}
}
