package handlers

import (
	"encoding/json"
	"net/http"
	"os/exec"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/nginx"
)

type NginxHandler struct {
	generator *nginx.Generator
}

func NewNginxHandler(g *nginx.Generator) *NginxHandler {
	return &NginxHandler{generator: g}
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
	out, err := exec.Command("podman", "exec", "nginx-proxy-manager", "nginx", "-s", "reload").CombinedOutput()
	if err != nil {
		http.Error(w, string(out), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "reloaded"})
}
