package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLimitBodyMiddlewareRejectsOversizedPayload(t *testing.T) {
	echoHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Request Entity Too Large", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	maxBytes := int64(256 << 10)
	limitedHandler := LimitBodyMiddleware(maxBytes, echoHandler)

	smallBody := bytes.Repeat([]byte("A"), 1024)
	reqSmall := httptest.NewRequest(http.MethodPost, "/api/admin/respond", bytes.NewReader(smallBody))
	recSmall := httptest.NewRecorder()
	limitedHandler.ServeHTTP(recSmall, reqSmall)
	if recSmall.Code != http.StatusOK {
		t.Errorf("Expected small body to return 200, got %d", recSmall.Code)
	}

	tenMB := 10 * 1024 * 1024
	largeBody := bytes.Repeat([]byte("B"), tenMB)
	reqLarge := httptest.NewRequest(http.MethodPost, "/api/admin/respond", bytes.NewReader(largeBody))
	recLarge := httptest.NewRecorder()
	limitedHandler.ServeHTTP(recLarge, reqLarge)
	if recLarge.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected 10MB body to return 413 RequestEntityTooLarge, got %d", recLarge.Code)
	}
}
