package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type SSEClient struct {
	id       string
	username string
	send     chan []byte
}

type SSEBroker struct {
	mu         sync.RWMutex
	clients    map[*SSEClient]bool
	register   chan *SSEClient
	unregister chan *SSEClient
	broadcast  chan []byte
	quit       chan struct{}
}

var GlobalSSEBroker = NewSSEBroker()

func NewSSEBroker() *SSEBroker {
	b := &SSEBroker{
		clients:    make(map[*SSEClient]bool),
		register:   make(chan *SSEClient),
		unregister: make(chan *SSEClient),
		broadcast:  make(chan []byte, 128),
		quit:       make(chan struct{}),
	}
	go b.run()
	return b
}

func (b *SSEBroker) run() {
	for {
		select {
		case <-b.quit:
			b.mu.Lock()
			for client := range b.clients {
				close(client.send)
				delete(b.clients, client)
			}
			b.mu.Unlock()
			return

		case client := <-b.register:
			b.mu.Lock()
			b.clients[client] = true
			b.mu.Unlock()

		case client := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client.send)
			}
			b.mu.Unlock()

		case msg := <-b.broadcast:
			b.mu.RLock()
			for client := range b.clients {
				select {
				case client.send <- msg:
				default:
				}
			}
			b.mu.RUnlock()
		}
	}
}

func (b *SSEBroker) Close() {
	close(b.quit)
}

func (b *SSEBroker) Broadcast(eventName string, data interface{}) {
	raw, err := json.Marshal(data)
	if err != nil {
		raw = []byte("{}")
	}
	msg := []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, string(raw)))

	select {
	case b.broadcast <- msg:
	default:
	}
}

func (b *SSEBroker) BroadcastNewReport(reportID int64) {
	b.Broadcast("new_report", map[string]interface{}{"report_id": reportID})
}

func (b *SSEBroker) BroadcastNewMessage(reportID int64) {
	b.Broadcast("new_message", map[string]interface{}{"report_id": reportID})
}

func (b *SSEBroker) BroadcastRegionalResponse(reportID int64) {
	b.Broadcast("regional_response", map[string]interface{}{"report_id": reportID})
}

func (b *SSEBroker) BroadcastSessionRevoked(username, tokenHash string) {
	b.Broadcast("session_revoked", map[string]string{"username": username})
}

func AdminEventsHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	admin, ok := r.Context().Value("admin").(AdminSession)
	if !ok || admin.Username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	client := &SSEClient{
		id:       admin.Token,
		username: admin.Username,
		send:     make(chan []byte, 32),
	}

	GlobalSSEBroker.register <- client
	defer func() {
		select {
		case GlobalSSEBroker.unregister <- client:
		case <-time.After(500 * time.Millisecond):
		}
	}()

	flusher.Flush()

	heartbeatTicker := time.NewTicker(15 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case <-heartbeatTicker.C:
			_, err := w.Write([]byte("event: heartbeat\ndata: {}\n\n"))
			if err != nil {
				return
			}
			flusher.Flush()

		case msg, ok := <-client.send:
			if !ok {
				return
			}
			_, err := w.Write(msg)
			if err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
