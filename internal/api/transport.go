package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/prospect-ogujiuba/devarch/internal/apply"
	"github.com/prospect-ogujiuba/devarch/internal/appsvc"
	"github.com/prospect-ogujiuba/devarch/internal/events"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
)

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type requestError struct {
	message string
}

func (e requestError) Error() string { return e.message }

func badRequest(message string) error { return requestError{message: message} }

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)
}

func writeError(w http.ResponseWriter, err error) {
	apiErr, status := mapAPIError(err)
	writeJSON(w, status, errorResponse{Error: apiErr})
}

func mapAPIError(err error) (apiError, int) {
	if err == nil {
		return apiError{Code: "internal_error", Message: http.StatusText(http.StatusInternalServerError)}, http.StatusInternalServerError
	}

	var notFound *appsvc.NotFoundError
	if errors.As(err, &notFound) {
		return apiError{Code: "not_found", Message: notFound.Error()}, http.StatusNotFound
	}

	var unsupported *appsvc.UnsupportedCapabilityError
	if errors.As(err, &unsupported) {
		return apiError{Code: "unsupported_capability", Message: unsupported.Error(), Details: unsupported}, http.StatusConflict
	}

	var requestErr requestError
	if errors.As(err, &requestErr) {
		return apiError{Code: "bad_request", Message: requestErr.Error()}, http.StatusBadRequest
	}
	if errors.Is(err, apply.ErrBlocked) {
		return apiError{Code: "apply_blocked", Message: err.Error()}, http.StatusConflict
	}

	var runtimeUnsupported *runtimepkg.UnsupportedOperationError
	if errors.As(err, &runtimeUnsupported) {
		return apiError{Code: "unsupported_capability", Message: runtimeUnsupported.Error(), Details: runtimeUnsupported}, http.StatusConflict
	}

	return apiError{Code: "internal_error", Message: err.Error()}, http.StatusInternalServerError
}

type sseWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func newSSEWriter(w http.ResponseWriter) (*sseWriter, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("response writer does not support streaming")
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	return &sseWriter{w: w, flusher: flusher}, nil
}

func (s *sseWriter) Comment(comment string) error {
	if s == nil {
		return fmt.Errorf("nil sse writer")
	}
	if _, err := fmt.Fprintf(s.w, ": %s\n\n", comment); err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

func (s *sseWriter) Envelope(envelope events.Envelope) error {
	if s == nil {
		return fmt.Errorf("nil sse writer")
	}
	encoded, err := events.MarshalEnvelope(envelope)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(s.w, "id: %d\nevent: %s\ndata: %s\n\n", envelope.Sequence, envelope.Kind, string(encoded)); err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

type execRequestFrame struct {
	Type        string   `json:"type,omitempty"`
	Command     []string `json:"command"`
	Interactive bool     `json:"interactive,omitempty"`
	TTY         bool     `json:"tty,omitempty"`
}

type execResultFrame struct {
	Type     string    `json:"type"`
	ExitCode int       `json:"exitCode,omitempty"`
	Stdout   string    `json:"stdout,omitempty"`
	Stderr   string    `json:"stderr,omitempty"`
	Error    *apiError `json:"error,omitempty"`
}

func readExecRequestFrame(conn *websocket.Conn) (runtimepkg.ExecRequest, error) {
	var frame execRequestFrame
	if err := conn.ReadJSON(&frame); err != nil {
		return runtimepkg.ExecRequest{}, badRequest("expected one JSON exec request frame")
	}
	if frame.Type != "" && frame.Type != "request" {
		return runtimepkg.ExecRequest{}, badRequest("exec frame type must be \"request\"")
	}
	if len(frame.Command) == 0 {
		return runtimepkg.ExecRequest{}, badRequest("exec command must not be empty")
	}
	return runtimepkg.ExecRequest{Command: append([]string(nil), frame.Command...), Interactive: frame.Interactive, TTY: frame.TTY}, nil
}

func writeExecResultFrame(conn *websocket.Conn, result *runtimepkg.ExecResult) error {
	if result == nil {
		return fmt.Errorf("nil exec result")
	}
	return conn.WriteJSON(execResultFrame{Type: "result", ExitCode: result.ExitCode, Stdout: result.Stdout, Stderr: result.Stderr})
}

func writeExecErrorFrame(conn *websocket.Conn, err error) error {
	apiErr, _ := mapAPIError(err)
	return conn.WriteJSON(execResultFrame{Type: "error", Error: &apiErr})
}
