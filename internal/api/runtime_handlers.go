package api

import (
	"net/http"
	"strconv"
	"time"

	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
)

func (s *Server) handleWorkspaceApply(w http.ResponseWriter, r *http.Request) {
	result, err := s.service.ApplyWorkspace(r.Context(), r.PathValue("name"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleWorkspaceLogs(w http.ResponseWriter, r *http.Request) {
	request, resource, err := logsRequestFromHTTP(r)
	if err != nil {
		writeError(w, err)
		return
	}
	chunks, err := s.service.WorkspaceLogs(r.Context(), r.PathValue("name"), resource, request)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, chunks)
}

func (s *Server) handleWorkspaceEvents(w http.ResponseWriter, r *http.Request) {
	stream, cancel, err := s.service.SubscribeWorkspaceEvents(r.Context(), r.PathValue("name"), 32)
	if err != nil {
		writeError(w, err)
		return
	}
	defer cancel()

	writer, err := newSSEWriter(w)
	if err != nil {
		writeError(w, err)
		return
	}
	_ = writer.Comment("connected")
	for {
		select {
		case <-r.Context().Done():
			return
		case envelope, ok := <-stream:
			if !ok {
				return
			}
			if err := writer.Envelope(envelope); err != nil {
				return
			}
		}
	}
}

func (s *Server) handleWorkspaceExec(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	request, err := readExecRequestFrame(conn)
	if err != nil {
		_ = writeExecErrorFrame(conn, err)
		return
	}
	result, err := s.service.ExecWorkspace(r.Context(), r.PathValue("name"), r.PathValue("resource"), request)
	if err != nil {
		_ = writeExecErrorFrame(conn, err)
		return
	}
	_ = writeExecResultFrame(conn, result)
}

func logsRequestFromHTTP(r *http.Request) (runtimepkg.LogsRequest, string, error) {
	query := r.URL.Query()
	resource := query.Get("resource")
	if resource == "" {
		return runtimepkg.LogsRequest{}, "", badRequest("resource query parameter is required")
	}
	request := runtimepkg.LogsRequest{}
	if rawTail := query.Get("tail"); rawTail != "" {
		tail, err := strconv.Atoi(rawTail)
		if err != nil || tail < 0 {
			return runtimepkg.LogsRequest{}, "", badRequest("tail query parameter must be a non-negative integer")
		}
		request.Tail = tail
	}
	if rawSince := query.Get("since"); rawSince != "" {
		since, err := time.Parse(time.RFC3339, rawSince)
		if err != nil {
			return runtimepkg.LogsRequest{}, "", badRequest("since query parameter must be RFC3339")
		}
		request.Since = &since
	}
	if rawFollow := query.Get("follow"); rawFollow != "" {
		follow, err := strconv.ParseBool(rawFollow)
		if err != nil {
			return runtimepkg.LogsRequest{}, "", badRequest("follow query parameter must be true or false")
		}
		if follow {
			return runtimepkg.LogsRequest{}, "", badRequest("follow is not supported on the JSON logs endpoint")
		}
	}
	return request, resource, nil
}
