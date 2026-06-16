// Package websocket provides the HTTP handler that upgrades connections to WebSocket.
package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	gws "github.com/gorilla/websocket"
)

// ─── Upgrader ─────────────────────────────────────────────────────────────────

// newUpgrader creates a Gorilla WebSocket Upgrader.
// CheckOrigin is intentionally permissive for development.
// For production, restrict to known frontend origins.
func newUpgrader(readBufSize, writeBufSize int) gws.Upgrader {
	return gws.Upgrader{
		ReadBufferSize:  readBufSize,
		WriteBufferSize: writeBufSize,
		CheckOrigin: func(r *http.Request) bool {
			// TODO: In production, validate r.Header.Get("Origin") against allowlist.
			return true
		},
	}
}

// ─── WsHandler ───────────────────────────────────────────────────────────────

// WsHandler handles HTTP → WebSocket upgrade requests.
type WsHandler struct {
	hub            *Hub
	upgrader       gws.Upgrader
	clientSendBuf  int
}

// NewWsHandler creates a WsHandler with injected Hub and buffer configuration.
func NewWsHandler(hub *Hub, readBufSize, writeBufSize, clientSendBuf int) *WsHandler {
	return &WsHandler{
		hub:           hub,
		upgrader:      newUpgrader(readBufSize, writeBufSize),
		clientSendBuf: clientSendBuf,
	}
}

// ServeWS upgrades the HTTP connection to a WebSocket and starts the client pumps.
// Registered as: GET /ws
func (h *WsHandler) ServeWS(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WsHandler] Upgrade failed: %v", err)
		return // Upgrader already wrote HTTP error to c.Writer
	}

	client := newClient(h.hub, conn, h.clientSendBuf)

	// Start pumps in dedicated goroutines.
	// WritePump runs first so the send channel is drained before ReadPump exits.
	go client.WritePump()
	go client.ReadPump()
}
