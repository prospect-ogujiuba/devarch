package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/nginx"
)

type NginxHandler struct {
	generator       *nginx.Generator
	containerClient *container.Client
}

func NewNginxHandler(g *nginx.Generator, containerClient *container.Client) *NginxHandler {
	return &NginxHandler{generator: g, containerClient: containerClient}
}

func (h *NginxHandler) GenerateAll(w http.ResponseWriter, r *http.Request) {
	if err := h.generator.GenerateAll(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "generated"})
}

func (h *NginxHandler) GenerateOne(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if err := h.generator.GenerateProject(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "generated", "project": name})
}

func (h *NginxHandler) Reload(w http.ResponseWriter, r *http.Request) {
	out, err := h.containerClient.Exec("nginx-proxy-manager", []string{"nginx", "-s", "reload"})
	if err != nil {
		http.Error(w, out, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "reloaded"})
}
