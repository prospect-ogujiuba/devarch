package api

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/prospect-ogujiuba/devarch/internal/appsvc"
)

// Server wires the thin Phase 4 HTTP surface over the shared application
// service boundary.
type Server struct {
	service  *appsvc.Service
	mux      *http.ServeMux
	upgrader websocket.Upgrader
}

func NewServer(service *appsvc.Service) *Server {
	server := &Server{
		service: service,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(*http.Request) bool { return true },
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/catalog/templates", server.handleCatalogTemplates)
	mux.HandleFunc("GET /api/catalog/templates/{name}", server.handleCatalogTemplate)
	mux.HandleFunc("GET /api/workspaces", server.handleWorkspaces)
	mux.HandleFunc("GET /api/workspaces/{name}", server.handleWorkspace)
	mux.HandleFunc("GET /api/workspaces/{name}/manifest", server.handleWorkspaceManifest)
	mux.HandleFunc("GET /api/workspaces/{name}/graph", server.handleWorkspaceGraph)
	mux.HandleFunc("GET /api/workspaces/{name}/status", server.handleWorkspaceStatus)
	mux.HandleFunc("GET /api/workspaces/{name}/plan", server.handleWorkspacePlan)
	mux.HandleFunc("POST /api/workspaces/{name}/apply", server.handleWorkspaceApply)
	mux.HandleFunc("GET /api/workspaces/{name}/logs", server.handleWorkspaceLogs)
	mux.HandleFunc("GET /api/workspaces/{name}/events", server.handleWorkspaceEvents)
	mux.HandleFunc("GET /api/workspaces/{name}/resources/{resource}/exec", server.handleWorkspaceExec)
	server.mux = mux
	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
