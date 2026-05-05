package api

import "net/http"

func (s *Server) handleDoctor(w http.ResponseWriter, r *http.Request) {
	report, err := s.service.Doctor(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleRuntimeStatus(w http.ResponseWriter, r *http.Request) {
	report, err := s.service.RuntimeStatus(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleSocketStatus(w http.ResponseWriter, r *http.Request) {
	report, err := s.service.SocketStatus(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleSocketStart(w http.ResponseWriter, r *http.Request) {
	result, err := s.service.SocketStart(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleSocketStop(w http.ResponseWriter, r *http.Request) {
	result, err := s.service.SocketStop(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
