package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/proxy"
)

type ProxyHandler struct {
	generator *proxy.Generator
}

func NewProxyHandler(g *proxy.Generator) *ProxyHandler {
	return &ProxyHandler{generator: g}
}

type generateProxyRequest struct {
	ProxyType string `json:"proxy_type"`
}

// ListTypes returns supported proxy types.
func (h *ProxyHandler) ListTypes(w http.ResponseWriter, r *http.Request) {
	types := proxy.SupportedTypes()
	resp := make([]map[string]string, len(types))
	for i, t := range types {
		resp[i] = map[string]string{
			"id":   string(t),
			"name": proxyDisplayName(t),
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GenerateForService generates proxy config for a standalone service.
func (h *ProxyHandler) GenerateForService(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	proxyType, err := parseProxyType(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.generator.GenerateForService(proxyType, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GenerateForStack generates proxy config for all instances in a stack.
func (h *ProxyHandler) GenerateForStack(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	proxyType, err := parseProxyType(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.generator.GenerateForStack(proxyType, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GenerateForProject generates proxy config for a project.
func (h *ProxyHandler) GenerateForProject(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	proxyType, err := parseProxyType(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.generator.GenerateForProject(proxyType, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func parseProxyType(r *http.Request) (proxy.ProxyType, error) {
	var req generateProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return "", fmt.Errorf("invalid request body: %v", err)
	}
	pt := proxy.ProxyType(req.ProxyType)
	for _, supported := range proxy.SupportedTypes() {
		if pt == supported {
			return pt, nil
		}
	}
	return "", fmt.Errorf("unsupported proxy type %q — supported: nginx, caddy, traefik, haproxy, apache", req.ProxyType)
}

func proxyDisplayName(t proxy.ProxyType) string {
	switch t {
	case proxy.Nginx:
		return "Nginx"
	case proxy.Caddy:
		return "Caddy"
	case proxy.Traefik:
		return "Traefik"
	case proxy.HAProxy:
		return "HAProxy"
	case proxy.Apache:
		return "Apache"
	default:
		return string(t)
	}
}
