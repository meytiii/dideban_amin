package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetServerTimeHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/server-time", nil)
	w := httptest.NewRecorder()

	GetServerTimeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	xServerTime := w.Header().Get("X-Server-Time")
	if xServerTime == "" {
		t.Fatalf("expected X-Server-Time header, got empty")
	}

	var resp struct {
		ServerTime       string `json:"server_time"`
		ServerTimeUnixMs int64  `json:"server_time_unix_ms"`
		Timezone         string `json:"timezone"`
	}

	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response json: %v", err)
	}

	if resp.ServerTime == "" {
		t.Fatalf("expected non-empty server_time")
	}
	if resp.ServerTimeUnixMs <= 0 {
		t.Fatalf("expected positive server_time_unix_ms, got %d", resp.ServerTimeUnixMs)
	}
	if resp.Timezone == "" {
		t.Fatalf("expected non-empty timezone")
	}
}

func TestGetServerTimeHandlerMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/server-time", nil)
	w := httptest.NewRecorder()

	GetServerTimeHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", w.Code)
	}
}
