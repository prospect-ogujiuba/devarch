package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"github.com/priz/devarch-api/internal/api/respond"
	"github.com/priz/devarch-api/internal/podman"
	"github.com/priz/devarch-api/internal/security"
)

type ExecHandler struct {
	podmanClient *podman.Client
	upgrader     websocket.Upgrader
	secMode      security.Mode
}

func NewExecHandler(pc *podman.Client, allowedOrigins []string, secMode security.Mode) *ExecHandler {
	return &ExecHandler{
		podmanClient: pc,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				if len(allowedOrigins) == 0 || (len(allowedOrigins) == 1 && allowedOrigins[0] == "*") {
					return true
				}
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				for _, allowed := range allowedOrigins {
					if origin == allowed {
						return true
					}
				}
				return false
			},
		},
		secMode: secMode,
	}
}

type resizeMessage struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

func (h *ExecHandler) Handle(w http.ResponseWriter, r *http.Request) {
	containerName := chi.URLParam(r, "name")
	if containerName == "" {
		http.Error(w, "container name required", http.StatusBadRequest)
		return
	}

	if h.secMode.RequiresWSAuth() {
		token := r.URL.Query().Get("token")
		apiKey := os.Getenv("DEVARCH_API_KEY")
		if err := security.ValidateWSToken(token, []byte(apiKey)); err != nil {
			respond.Unauthorized(w, r, "unauthorized: invalid or missing ws token")
			return
		}
	}

	cmd := r.URL.Query().Get("cmd")
	if cmd == "" {
		cmd = "/bin/sh"
	}

	cols := 80
	rows := 24
	if v := r.URL.Query().Get("cols"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cols = n
		}
	}
	if v := r.URL.Query().Get("rows"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			rows = n
		}
	}

	ws, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("exec: websocket upgrade failed: %v", err)
		return
	}

	execID, err := h.podmanClient.ExecCreate(r.Context(), containerName, podman.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{cmd},
	})
	if err != nil {
		log.Printf("exec: create failed: %v", err)
		writeCloseError(ws, "Error: "+err.Error())
		return
	}

	rawConn, reader, err := h.podmanClient.ExecStart(execID, true)
	if err != nil {
		log.Printf("exec: start failed: %v", err)
		writeCloseError(ws, "Error: "+err.Error())
		return
	}

	if cols != 80 || rows != 24 {
		h.podmanClient.ExecResize(r.Context(), execID, rows, cols)
	}

	done := make(chan struct{}, 1)

	// container → browser
	go func() {
		defer func() { done <- struct{}{} }()
		buf := make([]byte, 4096)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				return
			}
			if err := ws.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				return
			}
		}
	}()

	// browser → container
	go func() {
		defer func() { done <- struct{}{} }()
		for {
			msgType, data, err := ws.ReadMessage()
			if err != nil {
				return
			}
			switch msgType {
			case websocket.TextMessage:
				if _, err := rawConn.Write(data); err != nil {
					return
				}
			case websocket.BinaryMessage:
				var msg resizeMessage
				if err := json.Unmarshal(data, &msg); err != nil {
					continue
				}
				if msg.Type == "resize" && msg.Cols > 0 && msg.Rows > 0 {
					h.podmanClient.ExecResize(context.Background(), execID, msg.Rows, msg.Cols)
				}
			}
		}
	}()

	// When either goroutine exits, close both ends to unblock the other
	<-done
	rawConn.Close()
	ws.Close()

	h.podmanClient.ExecInspect(context.Background(), execID)
}

func writeCloseError(ws *websocket.Conn, reason string) {
	const maxReason = 123
	if len(reason) > maxReason {
		reason = reason[:maxReason]
	}
	ws.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(4000, reason),
		time.Now().Add(time.Second),
	)
	// Wait for client to acknowledge close frame before TCP teardown,
	// otherwise the client sees code 1006 instead of 4000.
	ws.SetReadDeadline(time.Now().Add(time.Second))
	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			break
		}
	}
}
