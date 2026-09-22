package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCSRFProtectionMiddleware(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	protected := CSRFMiddleware(dummyHandler)

	validToken := "secure-csrf-token-1234567890abcdef"

	getReq := httptest.NewRequest(http.MethodGet, "/api/admin/reports", nil)
	getRec := httptest.NewRecorder()
	protected.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Errorf("Expected GET to pass without CSRF token, got %d", getRec.Code)
	}

	postNoHeader := httptest.NewRequest(http.MethodPost, "/api/admin/toggle-read", nil)
	postNoHeader.AddCookie(&http.Cookie{Name: "herasat_csrf", Value: validToken})
	recNoHeader := httptest.NewRecorder()
	protected.ServeHTTP(recNoHeader, postNoHeader)
	if recNoHeader.Code != http.StatusForbidden {
		t.Errorf("Expected POST without X-CSRF-Token header to return 403, got %d", recNoHeader.Code)
	}

	postNoCookie := httptest.NewRequest(http.MethodPost, "/api/admin/toggle-read", nil)
	postNoCookie.Header.Set("X-CSRF-Token", validToken)
	recNoCookie := httptest.NewRecorder()
	protected.ServeHTTP(recNoCookie, postNoCookie)
	if recNoCookie.Code != http.StatusForbidden {
		t.Errorf("Expected POST without herasat_csrf cookie to return 403, got %d", recNoCookie.Code)
	}

	postMismatch := httptest.NewRequest(http.MethodPost, "/api/admin/toggle-read", nil)
	postMismatch.Header.Set("X-CSRF-Token", "wrong-csrf-token")
	postMismatch.AddCookie(&http.Cookie{Name: "herasat_csrf", Value: validToken})
	recMismatch := httptest.NewRecorder()
	protected.ServeHTTP(recMismatch, postMismatch)
	if recMismatch.Code != http.StatusForbidden {
		t.Errorf("Expected POST with mismatched CSRF token to return 403, got %d", recMismatch.Code)
	}

	postMatching := httptest.NewRequest(http.MethodPost, "/api/admin/toggle-read", nil)
	postMatching.Header.Set("X-CSRF-Token", validToken)
	postMatching.AddCookie(&http.Cookie{Name: "herasat_csrf", Value: validToken})
	recMatching := httptest.NewRecorder()
	protected.ServeHTTP(recMatching, postMatching)
	if recMatching.Code != http.StatusOK {
		t.Errorf("Expected POST with matching CSRF token to return 200, got %d", recMatching.Code)
	}
}
