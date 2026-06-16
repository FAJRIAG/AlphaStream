// Package websocket provides the Client type representing a single WebSocket connection.
//
// Each Client runs two goroutines:
//   - readPump:  reads messages from the browser (subscriptions, pings)
//   - writePump: writes broadcast messages from the hub to the browser
//
// The Client is responsible for its own cleanup: when either pump exits,
// it unregisters itself from the Hub to prevent resource leaks.
package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/alphastream/backend-go/internal/domain/entity"
)

// ─── Constants ────────────────────────────────────────────────────────────────

const (
	// writeWait is the maximum time allowed to write a message to the client.
	writeWait = 10 * time.Second

	// pongWait is the maximum time to wait for a pong response from the client.
	pongWait = 60 * time.Second

	// pingPeriod is how often the server sends ping frames to the client.
	// Must be less than pongWait to allow time for the pong to arrive.
	pingPeriod = (pongWait * 9) / 10

	// maxMessageSize is the maximum size (bytes) of a message received from a client.
	maxMessageSize = 4096
)

// ─── Client ───────────────────────────────────────────────────────────────────

// Client is a middleman between the WebSocket connection and the Hub.
// It is unexported to prevent direct construction outside this package.
type Client struct {
	hub  *Hub
	conn *websocket.Conn

	// send is a buffered channel of outbound messages.
	// The Hub writes to it; writePump reads from it.
	send chan []byte

	mu            sync.RWMutex
	subscriptions map[string]struct{}
}

// newClient creates a Client and registers it with the Hub.
func newClient(hub *Hub, conn *websocket.Conn, sendBufferSize int) *Client {
	c := &Client{
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, sendBufferSize),
		subscriptions: make(map[string]struct{}),
	}
	hub.RegisterClient(c)
	return c
}

// ReadPump pumps messages from the WebSocket connection to the Hub.
//
// The application runs ReadPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.UnregisterClient(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Printf("[Client] SetReadDeadline error: %v", err)
		return
	}
	c.conn.SetPongHandler(func(string) error {
		// Reset the read deadline on every pong to keep the connection alive.
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, rawMsg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseNormalClosure,
			) {
				log.Printf("[Client] ReadPump unexpected close: %v", err)
			}
			return // Exit pump → defer triggers UnregisterClient
		}

		c.handleClientMessage(rawMsg)
	}
}

// WritePump pumps messages from the Hub to the WebSocket connection.
//
// A goroutine running WritePump is started for each connection. The application
// ensures that there is at most one writer to a connection by executing all
// writes from this goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("[Client] SetWriteDeadline error: %v", err)
				return
			}

			if !ok {
				// Hub closed the send channel (client was unregistered).
				// Send a CloseMessage and exit cleanly.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write the message as a single text frame.
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("[Client] WritePump write error: %v", err)
				return
			}

		case <-ticker.C:
			// Send a periodic ping to detect dead connections.
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ─── Client Message Handling ──────────────────────────────────────────────────

// IsSubscribed returns true if the client is subscribed to the given symbol.
func (c *Client) IsSubscribed(symbol string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, subscribed := c.subscriptions[symbol]
	return subscribed
}

// handleClientMessage processes a raw JSON message received from the browser client.
// Currently handles SUBSCRIBE, UNSUBSCRIBE, and PING message types.
func (c *Client) handleClientMessage(raw []byte) {
	var msg entity.WsMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("[Client] Failed to parse client message: %v", err)
		c.sendError("PARSE_ERROR", "Invalid JSON message format")
		return
	}

	switch msg.Type {
	case entity.WsMsgPing:
		c.sendPong()
	case entity.WsMsgSubscribe:
		if msg.Symbol != "" {
			c.mu.Lock()
			c.subscriptions[msg.Symbol] = struct{}{}
			c.mu.Unlock()
			log.Printf("[Client] Subscribe request for symbol: %s", msg.Symbol)
		}
	case entity.WsMsgUnsubscribe:
		if msg.Symbol != "" {
			c.mu.Lock()
			delete(c.subscriptions, msg.Symbol)
			c.mu.Unlock()
			log.Printf("[Client] Unsubscribe request for symbol: %s", msg.Symbol)
		}
	default:
		log.Printf("[Client] Unknown message type: %s", msg.Type)
	}
}

// sendPong sends a PONG response to an explicit PING from the client.
func (c *Client) sendPong() {
	pong := entity.NewWsMessage(entity.WsMsgPong, "", nil)
	data, err := json.Marshal(pong)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
		// Drop if channel is full
	}
}

// sendError sends an error message back to the client.
func (c *Client) sendError(code, message string) {
	errMsg := entity.NewWsErrorMessage(code, message)
	data, err := json.Marshal(errMsg)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
	}
}
