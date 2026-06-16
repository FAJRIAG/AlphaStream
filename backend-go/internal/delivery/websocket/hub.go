// Package websocket provides a thread-safe WebSocket Hub for AlphaStream.
// The Hub manages all active client connections and routes broadcast messages.
//
// Concurrency model:
//   - Hub.clients map is protected by sync.RWMutex
//   - Broadcast messages are sent to each client's buffered `send` channel
//   - If a client's channel is full (slow consumer), the message is dropped (non-blocking select)
//     to prevent one slow client from blocking all others
//   - Hub listens on a `done` channel for graceful shutdown
package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/alphastream/backend-go/internal/domain/entity"
)

// Hub maintains the set of active WebSocket clients and broadcasts messages to them.
// It implements the usecase.IBroadcaster interface.
type Hub struct {
	// mu protects the clients map from concurrent read/write access.
	mu sync.RWMutex

	// clients holds all currently connected WebSocket clients.
	// The map value is unused (struct{}{}); the Client pointer is the key.
	clients map[*Client]struct{}

	// broadcast receives WsMessages from the usecase layer for delivery.
	broadcast chan entity.WsMessage

	// done signals the Hub goroutine to stop cleanly on shutdown.
	done chan struct{}
}

// NewHub creates and returns a new Hub ready to run.
// Call hub.Run() in a dedicated goroutine after creation.
func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*Client]struct{}),
		broadcast: make(chan entity.WsMessage, 512), // Buffered to absorb burst traffic
		done:      make(chan struct{}),
	}
}

// Run starts the Hub's main event loop.
// It must be called in a dedicated goroutine (go hub.Run()).
// It processes register, unregister, and broadcast events until Shutdown() is called.
func (h *Hub) Run() {
	log.Println("[Hub] WebSocket Hub started")
	for {
		select {
		case msg := <-h.broadcast:
			h.deliverBroadcast(msg)

		case <-h.done:
			log.Println("[Hub] WebSocket Hub shutting down")
			h.closeAllClients()
			return
		}
	}
}

// Broadcast enqueues a WsMessage for delivery to all connected clients.
// This method is safe for concurrent use from any goroutine.
// It implements usecase.IBroadcaster.
func (h *Hub) Broadcast(msg entity.WsMessage) {
	if h.ClientCount() == 0 {
		return // No clients connected, silent discard
	}
	select {
	case h.broadcast <- msg:
	default:
		// If the broadcast channel is full (burst overload), drop the message.
		// This prevents the calling goroutine from blocking.
		log.Printf("[Hub] Broadcast channel full, dropping message type=%s", msg.Type)
	}
}

// RegisterClient adds a new client to the Hub's registry.
func (h *Hub) RegisterClient(c *Client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
	log.Printf("[Hub] Client registered. Total clients: %d", h.clientCount())
}

// UnregisterClient removes a client from the Hub and closes its send channel.
func (h *Hub) UnregisterClient(c *Client) {
	h.mu.Lock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send) // Signal writePump to exit
	}
	h.mu.Unlock()
	log.Printf("[Hub] Client unregistered. Total clients: %d", h.clientCount())
}

// Shutdown signals the Hub to stop its Run() loop and close all client connections.
// Safe to call multiple times (channel close is idempotent via sync.Once pattern).
func (h *Hub) Shutdown() {
	select {
	case <-h.done:
		// Already closed — do nothing
	default:
		close(h.done)
	}
}

// ClientCount returns the number of currently connected clients.
func (h *Hub) ClientCount() int {
	return h.clientCount()
}

// GetActiveSubscriptions returns a slice of all unique symbols currently subscribed by connected clients.
func (h *Hub) GetActiveSubscriptions() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	activeMap := make(map[string]struct{})
	for client := range h.clients {
		client.mu.RLock()
		for sym := range client.subscriptions {
			activeMap[sym] = struct{}{}
		}
		client.mu.RUnlock()
	}

	activeList := make([]string, 0, len(activeMap))
	for sym := range activeMap {
		activeList = append(activeList, sym)
	}
	return activeList
}

// ─── Private Methods ──────────────────────────────────────────────────────────

// deliverBroadcast serializes a WsMessage to JSON and enqueues it in each client's
// send channel. Uses subscription filtering and a non-blocking select to drop messages.
func (h *Hub) deliverBroadcast(msg entity.WsMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[Hub] Failed to marshal broadcast message: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		// Filter by subscription if symbol is set
		if msg.Symbol != "" && !client.IsSubscribed(msg.Symbol) {
			continue
		}

		select {
		case client.send <- data:
			// Message enqueued successfully
		default:
			// Client's send buffer is full → slow consumer → drop message.
			// The client will be cleaned up by its own writePump on next write error.
			log.Printf("[Hub] Slow client detected, dropping message")
		}
	}
}

// closeAllClients closes all client send channels, triggering their writePumps to exit.
func (h *Hub) closeAllClients() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		close(client.send)
		delete(h.clients, client)
	}
}

// clientCount returns the number of clients using a read lock.
func (h *Hub) clientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
