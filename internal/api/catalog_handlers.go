package api

import "net/http"

func (s *Server) handleCatalogTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := s.service.CatalogTemplates(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, templates)
}

func (s *Server) handleCatalogTemplate(w http.ResponseWriter, r *http.Request) {
	template, err := s.service.CatalogTemplate(r.Context(), r.PathValue("name"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, template)
}
