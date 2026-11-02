package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAdminAuthMiddlewareJSONErrors(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	wrapped := AdminAuthMiddleware(dummyHandler)

	{
		req := httptest.NewRequest(http.MethodGet, "/api/admin/reports", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for missing session, got %d", rec.Code)
		}
		if !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
			t.Fatalf("expected application/json, got %s", rec.Header().Get("Content-Type"))
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}
		if body["error"] != "session_missing" {
			t.Fatalf("expected error session_missing, got %v", body["error"])
		}
	}

	{
		req := httptest.NewRequest(http.MethodGet, "/api/admin/reports", nil)
		req.AddCookie(&http.Cookie{Name: "herasat_admin_session", Value: "non-existent-token-123456789"})
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for invalid session, got %d", rec.Code)
		}
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if body["error"] != "session_invalid" {
			t.Fatalf("expected error session_invalid, got %v", body["error"])
		}
	}

	{
		idleToken := "test-idle-token-xyz"
		now := time.Now()
		Sessions.mu.Lock()
		Sessions.sessions[idleToken] = AdminSession{
			Token:      idleToken,
			Username:   "test_idle_admin",
			Role:       "superuser",
			CreatedAt:  now.Add(-30 * time.Minute),
			LastActive: now.Add(-20 * time.Minute),
			Expiry:     now.Add(10 * time.Minute),
		}
		delete(Sessions.cache, hashToken(idleToken))
		Sessions.mu.Unlock()

		req := httptest.NewRequest(http.MethodGet, "/api/admin/reports", nil)
		req.AddCookie(&http.Cookie{Name: "herasat_admin_session", Value: idleToken})
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for idle expired session, got %d", rec.Code)
		}
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if body["error"] != "session_expired" {
			t.Fatalf("expected error session_expired, got %v", body["error"])
		}
		if body["reason"] != "idle" {
			t.Fatalf("expected reason idle, got %v", body["reason"])
		}
	}

	{
		absToken := "test-abs-token-xyz"
		now := time.Now()
		Sessions.mu.Lock()
		Sessions.sessions[absToken] = AdminSession{
			Token:      absToken,
			Username:   "test_abs_admin",
			Role:       "superuser",
			CreatedAt:  now.Add(-13 * time.Hour),
			LastActive: now.Add(-2 * time.Minute),
			Expiry:     now.Add(10 * time.Minute),
		}
		delete(Sessions.cache, hashToken(absToken))
		Sessions.mu.Unlock()

		req := httptest.NewRequest(http.MethodGet, "/api/admin/reports", nil)
		req.AddCookie(&http.Cookie{Name: "herasat_admin_session", Value: absToken})
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for absolute expired session, got %d", rec.Code)
		}
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if body["error"] != "session_expired" {
			t.Fatalf("expected error session_expired, got %v", body["error"])
		}
		if body["reason"] != "absolute" {
			t.Fatalf("expected reason absolute, got %v", body["reason"])
		}
	}

	{
		kickedToken := "test-kicked-token-xyz"
		Sessions.mu.Lock()
		Sessions.kickedTokens[kickedToken] = "CONCURRENT_LOGIN"
		delete(Sessions.cache, hashToken(kickedToken))
		Sessions.mu.Unlock()

		req := httptest.NewRequest(http.MethodGet, "/api/admin/reports", nil)
		req.AddCookie(&http.Cookie{Name: "herasat_admin_session", Value: kickedToken})
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for concurrent login kick, got %d", rec.Code)
		}
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if body["error"] != "concurrent_login" {
			t.Fatalf("expected error concurrent_login, got %v", body["error"])
		}
	}

	{
		termToken := "test-term-token-xyz"
		Sessions.mu.Lock()
		Sessions.kickedTokens[termToken] = "TERMINATED_BY_SUPERUSER"
		delete(Sessions.cache, hashToken(termToken))
		Sessions.mu.Unlock()

		req := httptest.NewRequest(http.MethodGet, "/api/admin/reports", nil)
		req.AddCookie(&http.Cookie{Name: "herasat_admin_session", Value: termToken})
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for terminated session, got %d", rec.Code)
		}
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if body["error"] != "session_terminated" {
			t.Fatalf("expected error session_terminated, got %v", body["error"])
		}
	}
}

func TestCSRFMiddlewareJSONErrors(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	wrapped := CSRFMiddleware(dummyHandler)

	{
		req := httptest.NewRequest(http.MethodPost, "/api/admin/report/update", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for missing CSRF, got %d", rec.Code)
		}
		if !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
			t.Fatalf("expected application/json, got %s", rec.Header().Get("Content-Type"))
		}
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if body["error"] != "csrf_missing" {
			t.Fatalf("expected csrf_missing, got %v", body["error"])
		}
	}

	{
		req := httptest.NewRequest(http.MethodPost, "/api/admin/report/update", nil)
		req.AddCookie(&http.Cookie{Name: "herasat_csrf", Value: "token-a"})
		req.Header.Set("X-CSRF-Token", "token-b")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for invalid CSRF, got %d", rec.Code)
		}
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if body["error"] != "csrf_invalid" {
			t.Fatalf("expected csrf_invalid, got %v", body["error"])
		}
	}

	{
		req := httptest.NewRequest(http.MethodPost, "/api/admin/report/update", nil)
		req.AddCookie(&http.Cookie{Name: "herasat_csrf", Value: "token-a"})
		req.Header.Set("X-CSRF-Token", "token-a")
		req.Header.Set("Origin", "https://attacker-domain.evil")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for invalid Origin, got %d", rec.Code)
		}
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if body["error"] != "origin_invalid" {
			t.Fatalf("expected origin_invalid, got %v", body["error"])
		}
	}
}

func TestAdminLogoutExemptFromCSRFAndSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/logout", nil)
	rec := httptest.NewRecorder()

	LogoutHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for logout without session or CSRF, got %d", rec.Code)
	}

	var foundSessionClear, foundCSRFClear bool
	for _, c := range rec.Result().Cookies() {
		if c.Name == "herasat_admin_session" && c.MaxAge < 0 {
			foundSessionClear = true
		}
		if c.Name == "herasat_csrf" && c.MaxAge < 0 {
			foundCSRFClear = true
		}
	}
	if !foundSessionClear {
		t.Fatalf("expected herasat_admin_session cookie to be cleared")
	}
	if !foundCSRFClear {
		t.Fatalf("expected herasat_csrf cookie to be cleared")
	}
}

func TestAdminIPWhitelistPathFormatting(t *testing.T) {
	defer ResetNetworkSecurityForTest()

	origAdminCIDRs := os.Getenv("HERASAT_ADMIN_ALLOWED_CIDRS")
	origAllowAny := os.Getenv("HERASAT_ADMIN_ALLOW_ANY")
	origDevMode := os.Getenv("HERASAT_DEV_MODE")
	origProxies := os.Getenv("HERASAT_TRUSTED_PROXIES")

	defer func() {
		_ = os.Setenv("HERASAT_ADMIN_ALLOWED_CIDRS", origAdminCIDRs)
		_ = os.Setenv("HERASAT_ADMIN_ALLOW_ANY", origAllowAny)
		_ = os.Setenv("HERASAT_DEV_MODE", origDevMode)
		_ = os.Setenv("HERASAT_TRUSTED_PROXIES", origProxies)
	}()

	_ = os.Setenv("HERASAT_ADMIN_ALLOWED_CIDRS", "192.168.1.100/32")
	_ = os.Setenv("HERASAT_ADMIN_ALLOW_ANY", "false")
	_ = os.Setenv("HERASAT_DEV_MODE", "0")
	_ = os.Setenv("HERASAT_TRUSTED_PROXIES", "")

	if err := InitNetworkSecurity(); err != nil {
		t.Fatalf("InitNetworkSecurity failed: %v", err)
	}

	dummyHandler := AdminIPWhitelistMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	{
		reqAPI := httptest.NewRequest(http.MethodGet, "/api/admin/reports", nil)
		reqAPI.RemoteAddr = "10.0.0.5:12345"
		recAPI := httptest.NewRecorder()
		dummyHandler.ServeHTTP(recAPI, reqAPI)

		if recAPI.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for blocked IP on API, got %d", recAPI.Code)
		}
		if !strings.Contains(recAPI.Header().Get("Content-Type"), "application/json") {
			t.Fatalf("expected application/json for API path, got %s", recAPI.Header().Get("Content-Type"))
		}
		var body map[string]any
		_ = json.Unmarshal(recAPI.Body.Bytes(), &body)
		if body["error"] != "ip_not_allowed" {
			t.Fatalf("expected ip_not_allowed error, got %v", body["error"])
		}
	}

	{
		reqPage := httptest.NewRequest(http.MethodGet, "/admin/", nil)
		reqPage.RemoteAddr = "10.0.0.5:12345"
		recPage := httptest.NewRecorder()
		dummyHandler.ServeHTTP(recPage, reqPage)

		if recPage.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for blocked IP on page, got %d", recPage.Code)
		}
		if strings.Contains(recPage.Header().Get("Content-Type"), "application/json") {
			t.Fatalf("expected non-JSON content type for HTML page, got %s", recPage.Header().Get("Content-Type"))
		}
	}
}

func TestRateLimitMiddlewareAdminJSON(t *testing.T) {
	limiter := NewIPRateLimiter(1.0, 1)
	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
	wrapped := RateLimitMiddleware(limiter, dummyHandler)

	{
		req := httptest.NewRequest(http.MethodPost, "/api/admin/login", nil)
		req.RemoteAddr = "198.51.100.12:4444"
		rec1 := httptest.NewRecorder()
		wrapped(rec1, req)
		if rec1.Code != http.StatusOK {
			t.Fatalf("expected 200 on first hit, got %d", rec1.Code)
		}

		rec2 := httptest.NewRecorder()
		wrapped(rec2, req)
		if rec2.Code != http.StatusTooManyRequests {
			t.Fatalf("expected 429 on rate limit hit, got %d", rec2.Code)
		}
		if !strings.Contains(rec2.Header().Get("Content-Type"), "application/json") {
			t.Fatalf("expected application/json for /api/admin/ path, got %s", rec2.Header().Get("Content-Type"))
		}
		var body map[string]any
		_ = json.Unmarshal(rec2.Body.Bytes(), &body)
		if body["error"] != "rate_limited" {
			t.Fatalf("expected rate_limited error, got %v", body["error"])
		}
	}

	{
		limiterPub := NewIPRateLimiter(1.0, 1)
		wrappedPub := RateLimitMiddleware(limiterPub, dummyHandler)
		reqPub := httptest.NewRequest(http.MethodPost, "/api/submit", nil)
		reqPub.RemoteAddr = "198.51.100.13:4444"
		rec1 := httptest.NewRecorder()
		wrappedPub(rec1, reqPub)
		rec2 := httptest.NewRecorder()
		wrappedPub(rec2, reqPub)
		if rec2.Code != http.StatusTooManyRequests {
			t.Fatalf("expected 429 on public rate limit hit, got %d", rec2.Code)
		}
		if strings.Contains(rec2.Header().Get("Content-Type"), "application/json") {
			t.Fatalf("expected plain text for public route rate limit, got %s", rec2.Header().Get("Content-Type"))
		}
	}
}

func TestSessionTerminatedParallelPersistence(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	wrapped := AdminAuthMiddleware(dummyHandler)

	termToken := "parallel-term-test-token"
	th := hashToken(termToken)

	Sessions.mu.Lock()
	delete(Sessions.cache, th)
	delete(Sessions.sessions, termToken)
	if Sessions.tombstones != nil {
		delete(Sessions.tombstones, th)
	}
	Sessions.kickedTokens[termToken] = "TERMINATED_BY_SUPERUSER"
	Sessions.mu.Unlock()

	if db := Sessions.getDB(); db != nil {
		_, _ = db.Exec(`DELETE FROM admin_sessions WHERE token_hash = $1`, th)
		_, _ = db.Exec(`INSERT INTO admin_session_kicks (token_hash, reason, created_at)
			VALUES ($1, 'TERMINATED_BY_SUPERUSER', NOW())
			ON CONFLICT (token_hash) DO UPDATE SET reason = 'TERMINATED_BY_SUPERUSER'`, th)
	}

	var wg sync.WaitGroup
	errCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/api/admin/reports", nil)
			req.AddCookie(&http.Cookie{Name: "herasat_admin_session", Value: termToken})
			rec := httptest.NewRecorder()
			wrapped.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				mu.Lock()
				errCount++
				mu.Unlock()
				return
			}
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["error"] != "session_terminated" {
				mu.Lock()
				errCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if errCount > 0 {
		t.Fatalf("expected all 10 parallel requests to get session_terminated, but %d failed", errCount)
	}
}

func TestConcurrentLoginParallelPersistence(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	wrapped := AdminAuthMiddleware(dummyHandler)

	token1 := "concurrent-token-first-12345"
	session1 := AdminSession{
		Token:      token1,
		Username:   "concurrent_user_test",
		Role:       "superuser",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		Expiry:     time.Now().Add(1 * time.Hour),
	}
	Sessions.Set(token1, session1)

	token2 := "concurrent-token-second-67890"
	session2 := AdminSession{
		Token:      token2,
		Username:   "concurrent_user_test",
		Role:       "superuser",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		Expiry:     time.Now().Add(1 * time.Hour),
	}
	Sessions.Set(token2, session2)

	var wg sync.WaitGroup
	errCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/api/admin/reports", nil)
			req.AddCookie(&http.Cookie{Name: "herasat_admin_session", Value: token1})
			rec := httptest.NewRecorder()
			wrapped.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				mu.Lock()
				errCount++
				mu.Unlock()
				return
			}
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["error"] != "concurrent_login" {
				mu.Lock()
				errCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if errCount > 0 {
		t.Fatalf("expected all 10 parallel requests to get concurrent_login, but %d failed", errCount)
	}
}

func TestCSRFAuditThrottling(t *testing.T) {
	csrfAuditMu.Lock()
	csrfAuditThrottle = make(map[string]time.Time)
	csrfAuditMu.Unlock()

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	wrapped := CSRFMiddleware(dummyHandler)

	req1 := httptest.NewRequest(http.MethodPost, "/api/admin/respond", nil)
	req1.RemoteAddr = "192.168.1.55:1234"
	rec1 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusForbidden {
		t.Fatalf("expected 403 on missing CSRF, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/admin/respond", nil)
	req2.RemoteAddr = "192.168.1.55:1234"
	rec2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusForbidden {
		t.Fatalf("expected 403 on missing CSRF second time, got %d", rec2.Code)
	}

	csrfAuditMu.Lock()
	throttleCount := len(csrfAuditThrottle)
	csrfAuditMu.Unlock()

	if throttleCount != 1 {
		t.Fatalf("expected exactly 1 throttled audit record, got %d", throttleCount)
	}
}

