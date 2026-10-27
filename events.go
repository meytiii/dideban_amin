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
	role     string
	send     chan []byte
}

type sseBroadcastMsg struct {
	targetRoles []string
	msg         []byte
}

type SSEBroker struct {
	mu         sync.RWMutex
	clients    map[*SSEClient]bool
	register   chan *SSEClient
	unregister chan *SSEClient
	broadcast  chan sseBroadcastMsg
	quit       chan struct{}
}

var GlobalSSEBroker = NewSSEBroker()

func NewSSEBroker() *SSEBroker {
	b := &SSEBroker{
		clients:    make(map[*SSEClient]bool),
		register:   make(chan *SSEClient),
		unregister: make(chan *SSEClient),
		broadcast:  make(chan sseBroadcastMsg, 128),
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

		case bm := <-b.broadcast:
			b.mu.RLock()
			for client := range b.clients {
				if len(bm.targetRoles) > 0 {
					matched := false
					for _, r := range bm.targetRoles {
						if client.role == r {
							matched = true
							break
						}
					}
					if !matched {
						continue
					}
				}
				select {
				case client.send <- bm.msg:
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
	b.BroadcastToRoles(nil, eventName, data)
}

func (b *SSEBroker) BroadcastToRoles(roles []string, eventName string, data interface{}) {
	raw, err := json.Marshal(data)
	if err != nil {
		raw = []byte("{}")
	}
	msg := []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, string(raw)))

	select {
	case b.broadcast <- sseBroadcastMsg{targetRoles: roles, msg: msg}:
	default:
	}
}

func (b *SSEBroker) BroadcastNewReport(reportID int64) {
	b.BroadcastToRoles([]string{"superuser", "edare_kol"}, "new_report", map[string]interface{}{"report_id": reportID})
}

func (b *SSEBroker) BroadcastNewMessage(reportID int64) {
	b.Broadcast("new_message", map[string]interface{}{"report_id": reportID})
}

func (b *SSEBroker) BroadcastRegionalResponse(reportID int64) {
	b.Broadcast("regional_response", map[string]interface{}{"report_id": reportID})
}

func (b *SSEBroker) BroadcastSessionRevoked(username, tokenHash string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	msg := []byte("event: session_revoked\ndata: {}\n\n")
	for client := range b.clients {
		if client.username == username || hashToken(client.id) == tokenHash {
			select {
			case client.send <- msg:
			default:
			}
		}
	}
}

func AdminEventsHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	admin, ok := adminFromContext(r)
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
		role:     admin.Role,
		send:     make(chan []byte, 32),
	}

	GlobalSSEBroker.register <- client
	defer func() {
		select {
		case GlobalSSEBroker.unregister <- client:
		case <-time.After(500 * time.Millisecond):
		}
	}()

	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Time{})
	_ = rc.SetReadDeadline(time.Time{})

	flusher.Flush()

	heartbeatTicker := time.NewTicker(15 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case <-heartbeatTicker.C:
			currSess, exists := Sessions.Get(admin.Token)
			if !exists || currSess.Username != admin.Username || currSess.Role != admin.Role {
				return
			}
			_ = rc.SetWriteDeadline(time.Now().Add(10 * time.Second))
			_, err := w.Write([]byte("event: heartbeat\ndata: {}\n\n"))
			if err != nil {
				return
			}
			flusher.Flush()

		case msg, ok := <-client.send:
			if !ok {
				return
			}
			_ = rc.SetWriteDeadline(time.Now().Add(10 * time.Second))
			_, err := w.Write(msg)
			if err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
