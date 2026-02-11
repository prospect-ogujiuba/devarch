package respond

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// JSON wraps data in a success envelope and encodes to response
func JSON(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	envelope := SuccessEnvelope{Data: data}
	if err := json.NewEncoder(w).Encode(envelope); err != nil {
		log.Printf("error encoding response: %v", err)
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
		log.Printf("error encoding error response: %v", err)
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
	log.Printf("internal error: %v", err)
	Error(w, r, http.StatusInternalServerError, "internal_error", "An internal error occurred", nil)
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
