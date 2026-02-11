package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/api/respond"
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
		respond.InternalError(w, r, err)
		return
	}
	respond.Action(w, r, http.StatusOK, "generated")
}

func (h *NginxHandler) GenerateOne(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if err := h.generator.GenerateProject(name); err != nil {
		respond.InternalError(w, r, err)
		return
	}
	respond.Action(w, r, http.StatusOK, "generated", respond.WithMetadata("project", name))
}

func (h *NginxHandler) Reload(w http.ResponseWriter, r *http.Request) {
	out, err := h.containerClient.Exec("nginx-proxy-manager", []string{"nginx", "-s", "reload"})
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("%s", out))
		return
	}
	respond.Action(w, r, http.StatusOK, "reloaded")
}
