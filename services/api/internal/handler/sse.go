package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leksa/datamapper-senyar/internal/sse"
)

// SSEHandler handles SSE connections
type SSEHandler struct {
	hub *sse.Hub
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(hub *sse.Hub) *SSEHandler {
	return &SSEHandler{hub: hub}
}

// Stream handles SSE stream connections
// @Summary Subscribe to real-time updates
// @Description Opens an SSE connection for real-time sync and feed updates
// @Tags events
// @Produce text/event-stream
// @Success 200 {string} string "SSE stream"
// @Router /api/v1/events [get]
func (h *SSEHandler) Stream(c *gin.Context) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no")

	// Create client channel
	clientChan := make(chan sse.Event, 10)
	h.hub.Register(clientChan)

	// Send initial connection event
	initialEvent := sse.Event{
		Type:      "connected",
		Data:      map[string]interface{}{"message": "Connected to event stream"},
		Timestamp: time.Now(),
	}
	sendSSEEvent(c, initialEvent)

	// Cleanup on disconnect
	notify := c.Writer.CloseNotify()

	// Heartbeat ticker (every 30 seconds)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-notify:
			h.hub.Unregister(clientChan)
			return

		case event := <-clientChan:
			sendSSEEvent(c, event)

		case <-ticker.C:
			// Send heartbeat
			heartbeat := sse.Event{
				Type:      "heartbeat",
				Data:      map[string]interface{}{"clients": h.hub.ClientCount()},
				Timestamp: time.Now(),
			}
			sendSSEEvent(c, heartbeat)
		}
	}
}

// sendSSEEvent sends a single SSE event
func sendSSEEvent(c *gin.Context, event sse.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	fmt.Fprintf(c.Writer, "event: %s\n", event.Type)
	fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	c.Writer.Flush()
}
