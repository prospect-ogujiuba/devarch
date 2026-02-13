package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/priz/devarch-api/internal/podman"
)

type ImageHandler struct {
	podmanClient *podman.Client
}

func NewImageHandler(pc *podman.Client) *ImageHandler {
	return &ImageHandler{podmanClient: pc}
}

func (h *ImageHandler) List(w http.ResponseWriter, r *http.Request) {
	allStr := r.URL.Query().Get("all")
	all, _ := strconv.ParseBool(allStr)

	images, err := h.podmanClient.ListImages(r.Context(), all)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}

func (h *ImageHandler) Inspect(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name query parameter required", http.StatusBadRequest)
		return
	}

	image, err := h.podmanClient.InspectImage(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(image)
}

func (h *ImageHandler) Remove(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name query parameter required", http.StatusBadRequest)
		return
	}

	forceStr := r.URL.Query().Get("force")
	force, _ := strconv.ParseBool(forceStr)

	if err := h.podmanClient.RemoveImage(r.Context(), name, force); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ImageHandler) Pull(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Reference string `json:"reference"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Reference == "" {
		http.Error(w, "reference required", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err := h.podmanClient.PullImage(r.Context(), body.Reference, func(report podman.ImagePullReport) {
		data, err := json.Marshal(report)
		if err != nil {
			return
		}
		fmt.Fprintf(w, "%s\n", data)
		flusher.Flush()
	})

	if err != nil {
		log.Printf("image pull error: %v", err)
	}
}

func (h *ImageHandler) History(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name query parameter required", http.StatusBadRequest)
		return
	}

	entries, err := h.podmanClient.ImageHistory(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (h *ImageHandler) Prune(w http.ResponseWriter, r *http.Request) {
	danglingStr := r.URL.Query().Get("dangling")
	dangling := true
	if danglingStr != "" {
		dangling, _ = strconv.ParseBool(danglingStr)
	}

	reports, err := h.podmanClient.PruneImages(r.Context(), dangling)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}
