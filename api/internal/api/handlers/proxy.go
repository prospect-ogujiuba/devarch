package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/api/respond"
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

// ListTypes godoc
// @Summary      List supported proxy types
// @Description  Returns all supported reverse proxy types (nginx, caddy, traefik, haproxy, apache)
// @Tags         proxy
// @Produce      json
// @Success      200 {object} respond.SuccessEnvelope{data=[]map[string]string}
// @Router       /proxy/types [get]
// @Security     ApiKeyAuth
func (h *ProxyHandler) ListTypes(w http.ResponseWriter, r *http.Request) {
	types := proxy.SupportedTypes()
	resp := make([]map[string]string, len(types))
	for i, t := range types {
		resp[i] = map[string]string{
			"id":   string(t),
			"name": proxyDisplayName(t),
		}
	}
	respond.JSON(w, r, http.StatusOK,resp)
}

// GenerateForService godoc
// @Summary      Generate proxy config for service
// @Description  Generates reverse proxy configuration for a standalone service
// @Tags         proxy
// @Accept       json
// @Produce      json
// @Param        name path string true "Service name"
// @Param        request body generateProxyRequest true "Proxy generation request"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      422 {object} respond.ErrorEnvelope
// @Router       /services/{name}/proxy-config [post]
// @Security     ApiKeyAuth
func (h *ProxyHandler) GenerateForService(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	proxyType, err := parseProxyType(r)
	if err != nil {
		respond.BadRequest(w, r, err.Error())
		return
	}

	result, err := h.generator.GenerateForService(proxyType, name)
	if err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "unprocessable_entity", err.Error(), nil)
		return
	}

	respond.JSON(w, r, http.StatusOK,result)
}

// GenerateForStack godoc
// @Summary      Generate proxy config for stack
// @Description  Generates reverse proxy configuration for all instances in a stack
// @Tags         proxy
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        request body generateProxyRequest true "Proxy generation request"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      422 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/proxy-config [post]
// @Security     ApiKeyAuth
func (h *ProxyHandler) GenerateForStack(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	proxyType, err := parseProxyType(r)
	if err != nil {
		respond.BadRequest(w, r, err.Error())
		return
	}

	result, err := h.generator.GenerateForStack(proxyType, name)
	if err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "unprocessable_entity", err.Error(), nil)
		return
	}

	respond.JSON(w, r, http.StatusOK,result)
}

// GenerateForProject godoc
// @Summary      Generate proxy config for project
// @Description  Generates reverse proxy configuration for all services in a project
// @Tags         proxy
// @Accept       json
// @Produce      json
// @Param        name path string true "Project name"
// @Param        request body generateProxyRequest true "Proxy generation request"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      422 {object} respond.ErrorEnvelope
// @Router       /projects/{name}/proxy-config [post]
// @Security     ApiKeyAuth
func (h *ProxyHandler) GenerateForProject(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	proxyType, err := parseProxyType(r)
	if err != nil {
		respond.BadRequest(w, r, err.Error())
		return
	}

	result, err := h.generator.GenerateForProject(proxyType, name)
	if err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "unprocessable_entity", err.Error(), nil)
		return
	}

	respond.JSON(w, r, http.StatusOK,result)
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
