package sse

import (
	"sync"
	"time"
)

// Event represents a server-sent event
type Event struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// Hub manages SSE client connections
type Hub struct {
	clients    map[chan Event]bool
	broadcast  chan Event
	register   chan chan Event
	unregister chan chan Event
	mu         sync.RWMutex
}

// NewHub creates a new SSE hub
func NewHub() *Hub {
	hub := &Hub{
		clients:    make(map[chan Event]bool),
		broadcast:  make(chan Event, 100),
		register:   make(chan chan Event),
		unregister: make(chan chan Event),
	}
	go hub.run()
	return hub
}

// run handles hub operations
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client)
			}
			h.mu.Unlock()

		case event := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client <- event:
				default:
					// Client buffer full, skip
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends an event to all connected clients
func (h *Hub) Broadcast(eventType string, data interface{}) {
	event := Event{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}
	select {
	case h.broadcast <- event:
	default:
		// Broadcast channel full, skip
	}
}

// ClientCount returns the number of connected clients
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Register registers a new client channel
func (h *Hub) Register(client chan Event) {
	h.register <- client
}

// Unregister unregisters a client channel
func (h *Hub) Unregister(client chan Event) {
	h.unregister <- client
}
