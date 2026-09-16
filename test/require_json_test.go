package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequireJSONMiddleware(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})

	wrapped := RequireJSONMiddleware(dummyHandler)

	// 1. Missing Content-Type
	{
		req := httptest.NewRequest("POST", "/api/admin/login", strings.NewReader(`{}`))
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
		if w.Result().StatusCode != http.StatusUnsupportedMediaType {
			t.Errorf("expected 415 for missing Content-Type, got %d", w.Result().StatusCode)
		}
	}

	// 2. text/plain Content-Type
	{
		req := httptest.NewRequest("POST", "/api/admin/login", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
		if w.Result().StatusCode != http.StatusUnsupportedMediaType {
			t.Errorf("expected 415 for text/plain Content-Type, got %d", w.Result().StatusCode)
		}
	}

	// 3. application/x-www-form-urlencoded Content-Type
	{
		req := httptest.NewRequest("POST", "/api/admin/login", strings.NewReader(`username=admin`))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
		if w.Result().StatusCode != http.StatusUnsupportedMediaType {
			t.Errorf("expected 415 for urlencoded Content-Type, got %d", w.Result().StatusCode)
		}
	}

	// 4. application/json Content-Type
	{
		req := httptest.NewRequest("POST", "/api/admin/login", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK for application/json Content-Type, got %d", w.Result().StatusCode)
		}
	}

	// 5. application/json; charset=utf-8 Content-Type
	{
		req := httptest.NewRequest("POST", "/api/admin/login", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK for application/json; charset=utf-8 Content-Type, got %d", w.Result().StatusCode)
		}
	}
}
