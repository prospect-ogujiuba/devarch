package handlers

import (
	"encoding/json"
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
		http.Error(w, "invalid Content-Type: expected multipart/form-data", http.StatusBadRequest)
		return
	}
	boundary := params["boundary"]
	if boundary == "" {
		http.Error(w, "missing boundary in Content-Type", http.StatusBadRequest)
		return
	}

	// Get import size limit from env var (default 256MB)
	importMaxBytes := int64(256 << 20)
	if envVal := os.Getenv("STACK_IMPORT_MAX_BYTES"); envVal != "" {
		if parsed, err := strconv.ParseInt(envVal, 10, 64); err == nil {
			importMaxBytes = parsed
		}
	}

	// Create streaming multipart reader
	mr := multipart.NewReader(r.Body, boundary)

	// Find the "file" part
	var filePart *multipart.Part
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read multipart: %v", err), http.StatusBadRequest)
			return
		}
		if part.FormName() == "file" {
			filePart = part
			break
		}
		part.Close()
	}

	if filePart == nil {
		http.Error(w, "missing 'file' field in multipart form", http.StatusBadRequest)
		return
	}
	defer filePart.Close()

	// Apply size limit to the file part
	limitedReader := io.LimitReader(filePart, importMaxBytes)

	// Decode YAML from streaming reader
	var devarchFile export.DevArchFile
	if err := yaml.NewDecoder(limitedReader).Decode(&devarchFile); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			http.Error(w, fmt.Sprintf("payload too large: exceeds %d bytes", importMaxBytes), http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, fmt.Sprintf("failed to decode YAML: %v", err), http.StatusBadRequest)
		return
	}

	if devarchFile.Version != 1 {
		http.Error(w, fmt.Sprintf("unsupported version %d — only version 1 is supported", devarchFile.Version), http.StatusBadRequest)
		return
	}

	if devarchFile.Stack.Name == "" {
		http.Error(w, "stack name is required", http.StatusBadRequest)
		return
	}

	if err := container.ValidateName(devarchFile.Stack.Name); err != nil {
		http.Error(w, fmt.Sprintf("invalid stack name: %v", err), http.StatusBadRequest)
		return
	}

	importer := export.NewImporter(h.db)
	result, err := importer.Import(&devarchFile)
	if err != nil {
		if strings.Contains(err.Error(), "not found in catalog") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":  err.Error(),
				"result": result,
			})
			return
		}
		if strings.Contains(err.Error(), "locked") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("import failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
