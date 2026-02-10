package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/export"
	"gopkg.in/yaml.v3"
)

func (h *StackHandler) ImportStack(w http.ResponseWriter, r *http.Request) {
	// Extract boundary from Content-Type header
	contentType := r.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "multipart/form-data" {
		writeImportError(w, http.StatusBadRequest, "invalid Content-Type: expected multipart/form-data", nil)
		return
	}
	boundary := params["boundary"]
	if boundary == "" {
		writeImportError(w, http.StatusBadRequest, "missing boundary in Content-Type", nil)
		return
	}

	// Get import size limit from env var (default 256MB)
	importMaxBytes := int64(256 << 20)
	if envVal := os.Getenv("STACK_IMPORT_MAX_BYTES"); envVal != "" {
		if parsed, err := strconv.ParseInt(envVal, 10, 64); err == nil {
			importMaxBytes = parsed
		}
	}

	// Create streaming multipart reader with size cap to prevent boundary-based memory exhaustion
	mr := multipart.NewReader(io.LimitReader(r.Body, importMaxBytes+4096), boundary)

	// Find the "file" part
	var filePart *multipart.Part
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			if isRequestTooLarge(err) {
				writeImportError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("Import payload exceeds %dMB limit", importMaxBytes>>20), map[string]interface{}{
					"max_bytes":      importMaxBytes,
					"received_bytes": importMaxBytes + 1,
				})
				return
			}
			writeImportError(w, http.StatusBadRequest, fmt.Sprintf("failed to read multipart: %v", err), nil)
			return
		}
		if part.FormName() == "file" {
			filePart = part
			break
		}
		part.Close()
	}

	if filePart == nil {
		writeImportError(w, http.StatusBadRequest, "missing 'file' field in multipart form", nil)
		return
	}
	defer filePart.Close()

	// Stream file part to temp file to preserve memory safety while enforcing deterministic size checks.
	tmpFile, err := os.CreateTemp("", "devarch-stack-import-*.yml")
	if err != nil {
		writeImportError(w, http.StatusInternalServerError, "failed to create temp file", nil)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	receivedBytes, err := io.Copy(tmpFile, filePart)
	if err != nil {
		if isRequestTooLarge(err) {
			writeImportError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("Import payload exceeds %dMB limit", importMaxBytes>>20), map[string]interface{}{
				"max_bytes":      importMaxBytes,
				"received_bytes": importMaxBytes + 1,
			})
			return
		}
		writeImportError(w, http.StatusBadRequest, fmt.Sprintf("failed to read uploaded file: %v", err), nil)
		return
	}

	if receivedBytes > importMaxBytes {
		writeImportError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("Import payload exceeds %dMB limit", importMaxBytes>>20), map[string]interface{}{
			"max_bytes":      importMaxBytes,
			"received_bytes": receivedBytes,
		})
		return
	}

	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		writeImportError(w, http.StatusInternalServerError, "failed to rewind uploaded file", nil)
		return
	}

	// Decode YAML from streaming reader
	var devarchFile export.DevArchFile
	if err := yaml.NewDecoder(tmpFile).Decode(&devarchFile); err != nil {
		writeImportError(w, http.StatusBadRequest, fmt.Sprintf("failed to decode YAML: %v", err), nil)
		return
	}

	if devarchFile.Version != 1 {
		writeImportError(w, http.StatusBadRequest, fmt.Sprintf("unsupported version %d — only version 1 is supported", devarchFile.Version), nil)
		return
	}

	if devarchFile.Stack.Name == "" {
		writeImportError(w, http.StatusBadRequest, "stack name is required", nil)
		return
	}

	if err := container.ValidateName(devarchFile.Stack.Name); err != nil {
		writeImportError(w, http.StatusBadRequest, fmt.Sprintf("invalid stack name: %v", err), nil)
		return
	}

	importer := export.NewImporter(h.db)
	result, err := importer.Import(&devarchFile)
	if err != nil {
		if strings.Contains(err.Error(), "not found in catalog") {
			writeImportError(w, http.StatusBadRequest, err.Error(), map[string]interface{}{"result": result})
			return
		}
		if strings.Contains(err.Error(), "locked") {
			writeImportError(w, http.StatusConflict, err.Error(), nil)
			return
		}
		writeImportError(w, http.StatusInternalServerError, fmt.Sprintf("import failed: %v", err), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func isRequestTooLarge(err error) bool {
	var maxErr *http.MaxBytesError
	return errors.As(err, &maxErr)
}

func writeImportError(w http.ResponseWriter, status int, message string, extra map[string]interface{}) {
	payload := map[string]interface{}{"error": message}
	for k, v := range extra {
		payload[k] = v
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
