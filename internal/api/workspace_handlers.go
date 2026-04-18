package api

import "net/http"

func (s *Server) handleWorkspaces(w http.ResponseWriter, r *http.Request) {
	workspaces, err := s.service.Workspaces(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, workspaces)
}

func (s *Server) handleWorkspace(w http.ResponseWriter, r *http.Request) {
	workspace, err := s.service.Workspace(r.Context(), r.PathValue("name"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

func (s *Server) handleWorkspaceManifest(w http.ResponseWriter, r *http.Request) {
	manifest, err := s.service.WorkspaceManifest(r.Context(), r.PathValue("name"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, manifest)
}

func (s *Server) handleWorkspaceGraph(w http.ResponseWriter, r *http.Request) {
	graph, err := s.service.WorkspaceGraph(r.Context(), r.PathValue("name"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, graph)
}

func (s *Server) handleWorkspaceStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.service.WorkspaceStatus(r.Context(), r.PathValue("name"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleWorkspacePlan(w http.ResponseWriter, r *http.Request) {
	plan, err := s.service.WorkspacePlan(r.Context(), r.PathValue("name"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, plan)
}
