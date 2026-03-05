package respond

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type successEnvelope struct {
	Data interface{} `json:"data"`
}

type errorEnvelope struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(successEnvelope{Data: data}); err != nil {
		slog.Error("encode response", "error", err)
	}
}

func Error(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorEnvelope{Error: errorDetail{Code: code, Message: message}})
}

func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "bad_request", message)
}

func InternalError(w http.ResponseWriter, err error) {
	slog.Error("internal error", "error", err)
	Error(w, http.StatusInternalServerError, "internal_error", "An internal error occurred")
}
