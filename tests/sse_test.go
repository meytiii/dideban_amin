package main

import (
	"context"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSSEBrokerClientLifecycleAndBroadcast(t *testing.T) {
	broker := NewSSEBroker()
	defer broker.Close()

	client1 := &SSEClient{
		id:       "tok1",
		username: "admin1",
		send:     make(chan []byte, 32),
	}
	client2 := &SSEClient{
		id:       "tok2",
		username: "admin2",
		send:     make(chan []byte, 32),
	}

	broker.register <- client1
	broker.register <- client2
	time.Sleep(50 * time.Millisecond)

	broker.mu.RLock()
	count := len(broker.clients)
	broker.mu.RUnlock()
	if count != 2 {
		t.Fatalf("expected 2 clients, got %d", count)
	}

	broker.Broadcast("test_event", map[string]bool{"ok": true})

	select {
	case msg := <-client1.send:
		if !strings.Contains(string(msg), "test_event") {
			t.Fatalf("client1 got %s, want test_event", string(msg))
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for broadcast on client1")
	}

	select {
	case msg := <-client2.send:
		if !strings.Contains(string(msg), "test_event") {
			t.Fatalf("client2 got %s, want test_event", string(msg))
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for broadcast on client2")
	}

	broker.unregister <- client1
	time.Sleep(50 * time.Millisecond)

	broker.mu.RLock()
	count = len(broker.clients)
	broker.mu.RUnlock()
	if count != 1 {
		t.Fatalf("expected 1 client after unregister, got %d", count)
	}

	broker.unregister <- client2
	time.Sleep(50 * time.Millisecond)

	broker.mu.RLock()
	count = len(broker.clients)
	broker.mu.RUnlock()
	if count != 0 {
		t.Fatalf("expected 0 clients, got %d", count)
	}
}

func TestSSESlowClientDoesNotBlockBroker(t *testing.T) {
	broker := NewSSEBroker()
	defer broker.Close()

	slowClient := &SSEClient{
		id:       "slow1",
		username: "slow_admin",
		send:     make(chan []byte, 1),
	}
	broker.register <- slowClient
	time.Sleep(50 * time.Millisecond)

	done := make(chan bool)
	go func() {
		for i := 0; i < 200; i++ {
			broker.Broadcast("spam", map[string]int{"i": i})
		}
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("broker blocked on slow client broadcast")
	}

	broker.unregister <- slowClient
}

func TestAdminEventsHandlerStreamAndContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest("GET", "/api/admin/events", nil).WithContext(ctx)
	req = req.WithContext(contextWithAdmin(req.Context(), AdminSession{
		Username: "test_sse_admin",
		Title:    "مدیر سامانه",
		Role:     "superuser",
		Token:    "test_token_123",
	}))

	rec := httptest.NewRecorder()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		AdminEventsHandler(rec, req)
	}()

	time.Sleep(100 * time.Millisecond)

	GlobalSSEBroker.BroadcastNewReport(12345)

	time.Sleep(100 * time.Millisecond)
	cancel()
	wg.Wait()

	body := rec.Body.String()
	if !strings.Contains(body, "event: new_report") {
		t.Fatalf("missing new_report event in SSE stream: %s", body)
	}
	if !strings.Contains(body, "12345") {
		t.Fatalf("missing report ID in new_report event: %s", body)
	}
}
