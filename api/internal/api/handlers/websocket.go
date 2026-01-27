package handlers

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	devsync "github.com/priz/devarch-api/internal/sync"
)

type WebSocketHandler struct {
	syncManager *devsync.Manager
	upgrader    websocket.Upgrader
	clients     map[*websocket.Conn]bool
	mu          sync.RWMutex
}

func NewWebSocketHandler(sm *devsync.Manager) *WebSocketHandler {
	h := &WebSocketHandler{
		syncManager: sm,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients: make(map[*websocket.Conn]bool),
	}

	go h.broadcastLoop()

	return h
}

func (h *WebSocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
	}()

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}
	}
}

func (h *WebSocketHandler) broadcastLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-ticker.C:
			h.broadcastStatus()
		case <-pingTicker.C:
			h.pingClients()
		}
	}
}

func (h *WebSocketHandler) broadcastStatus() {
	status := h.syncManager.GetStatusUpdate()
	if status == nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.clients {
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		err := conn.WriteJSON(status)
		if err != nil {
			log.Printf("websocket write error: %v", err)
			conn.Close()
			go func(c *websocket.Conn) {
				h.mu.Lock()
				delete(h.clients, c)
				h.mu.Unlock()
			}(conn)
		}
	}
}

func (h *WebSocketHandler) pingClients() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.clients {
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		err := conn.WriteMessage(websocket.PingMessage, nil)
		if err != nil {
			conn.Close()
		}
	}
}

func (h *WebSocketHandler) Broadcast(msg interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.clients {
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		err := conn.WriteJSON(msg)
		if err != nil {
			conn.Close()
		}
	}
}
