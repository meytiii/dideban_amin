package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestParseCIDRs(t *testing.T) {
	nets, err := parseCIDRs("192.168.1.0/24, 10.0.0.5, 2001:db8::/32, ::1")
	if err != nil {
		t.Fatalf("expected valid CIDRs to parse, got: %v", err)
	}
	if len(nets) != 4 {
		t.Fatalf("expected 4 nets parsed, got: %d", len(nets))
	}

	_, err = parseCIDRs("192.168.1.300/24")
	if err == nil {
		t.Fatalf("expected malformed CIDR to fail")
	}

	_, err = parseCIDRs("invalid_ip/24")
	if err == nil {
		t.Fatalf("expected non-IP CIDR to fail")
	}

	_, err = parseCIDRs("10.0.0.1/35")
	if err == nil {
		t.Fatalf("expected out-of-range prefix CIDR to fail")
	}
}

func TestAdminAllowedCIDRsStartupRefusal(t *testing.T) {
	defer ResetNetworkSecurityForTest()
	defer ResetCookieSecurityForTest()

	origAdminCIDRs := os.Getenv("HERASAT_ADMIN_ALLOWED_CIDRS")
	origAllowAny := os.Getenv("HERASAT_ADMIN_ALLOW_ANY")
	origDevMode := os.Getenv("HERASAT_DEV_MODE")
	origCookieSecure := os.Getenv("HERASAT_COOKIE_SECURE")
	origCert := os.Getenv("HERASAT_TLS_CERT")
	origKey := os.Getenv("HERASAT_TLS_KEY")
	origProxies := os.Getenv("HERASAT_TRUSTED_PROXIES")

	defer func() {
		_ = os.Setenv("HERASAT_ADMIN_ALLOWED_CIDRS", origAdminCIDRs)
		_ = os.Setenv("HERASAT_ADMIN_ALLOW_ANY", origAllowAny)
		_ = os.Setenv("HERASAT_DEV_MODE", origDevMode)
		_ = os.Setenv("HERASAT_COOKIE_SECURE", origCookieSecure)
		_ = os.Setenv("HERASAT_TLS_CERT", origCert)
		_ = os.Setenv("HERASAT_TLS_KEY", origKey)
		_ = os.Setenv("HERASAT_TRUSTED_PROXIES", origProxies)
	}()

	InitCookieSecurity(false)
	_ = os.Setenv("HERASAT_ADMIN_ALLOWED_CIDRS", "")
	_ = os.Setenv("HERASAT_ADMIN_ALLOW_ANY", "false")
	_ = os.Setenv("HERASAT_DEV_MODE", "0")
	_ = os.Setenv("HERASAT_COOKIE_SECURE", "false")
	_ = os.Setenv("HERASAT_TLS_CERT", "")
	_ = os.Setenv("HERASAT_TLS_KEY", "")
	_ = os.Setenv("HERASAT_TRUSTED_PROXIES", "")

	err := InitNetworkSecurity()
	if err == nil {
		t.Fatalf("expected InitNetworkSecurity to fail when HERASAT_ADMIN_ALLOWED_CIDRS is empty and allow any is false")
	}
	if !strings.Contains(err.Error(), "HERASAT_ADMIN_ALLOWED_CIDRS is required") {
		t.Fatalf("expected required CIDRs error message, got: %v", err)
	}

	_ = os.Setenv("HERASAT_ADMIN_ALLOW_ANY", "true")
	err = InitNetworkSecurity()
	if err != nil {
		t.Fatalf("expected InitNetworkSecurity to succeed when HERASAT_ADMIN_ALLOW_ANY=true, got: %v", err)
	}

	_ = os.Setenv("HERASAT_ADMIN_ALLOW_ANY", "false")
	_ = os.Setenv("HERASAT_DEV_MODE", "1")
	err = InitNetworkSecurity()
	if err != nil {
		t.Fatalf("expected InitNetworkSecurity to succeed when HERASAT_DEV_MODE=1, got: %v", err)
	}

	_ = os.Setenv("HERASAT_DEV_MODE", "0")
	_ = os.Setenv("HERASAT_ADMIN_ALLOWED_CIDRS", "127.0.0.1/32, 10.0.0.0/8")
	err = InitNetworkSecurity()
	if err != nil {
		t.Fatalf("expected InitNetworkSecurity to succeed with valid admin CIDRs, got: %v", err)
	}
}

func TestReverseProxyCookieSecureStartupRefusal(t *testing.T) {
	defer ResetNetworkSecurityForTest()
	defer ResetCookieSecurityForTest()

	origAdminCIDRs := os.Getenv("HERASAT_ADMIN_ALLOWED_CIDRS")
	origAllowAny := os.Getenv("HERASAT_ADMIN_ALLOW_ANY")
	origDevMode := os.Getenv("HERASAT_DEV_MODE")
	origCookieSecure := os.Getenv("HERASAT_COOKIE_SECURE")
	origCert := os.Getenv("HERASAT_TLS_CERT")
	origKey := os.Getenv("HERASAT_TLS_KEY")
	origProxies := os.Getenv("HERASAT_TRUSTED_PROXIES")

	defer func() {
		_ = os.Setenv("HERASAT_ADMIN_ALLOWED_CIDRS", origAdminCIDRs)
		_ = os.Setenv("HERASAT_ADMIN_ALLOW_ANY", origAllowAny)
		_ = os.Setenv("HERASAT_DEV_MODE", origDevMode)
		_ = os.Setenv("HERASAT_COOKIE_SECURE", origCookieSecure)
		_ = os.Setenv("HERASAT_TLS_CERT", origCert)
		_ = os.Setenv("HERASAT_TLS_KEY", origKey)
		_ = os.Setenv("HERASAT_TRUSTED_PROXIES", origProxies)
	}()

	_ = os.Setenv("HERASAT_ADMIN_ALLOW_ANY", "true")
	_ = os.Setenv("HERASAT_DEV_MODE", "0")
	_ = os.Setenv("HERASAT_TLS_CERT", "")
	_ = os.Setenv("HERASAT_TLS_KEY", "")
	_ = os.Setenv("HERASAT_TRUSTED_PROXIES", "")
	InitCookieSecurity(true)

	err := InitNetworkSecurity()
	if err == nil {
		t.Fatalf("expected InitNetworkSecurity to fail when cookie secure true without TLS and empty trusted proxies")
	}
	if !strings.Contains(err.Error(), "HERASAT_TRUSTED_PROXIES is empty") {
		t.Fatalf("expected empty trusted proxies error message, got: %v", err)
	}

	_ = os.Setenv("HERASAT_TRUSTED_PROXIES", "10.0.0.0/8, 172.16.0.0/12")
	err = InitNetworkSecurity()
	if err != nil {
		t.Fatalf("expected InitNetworkSecurity to succeed when trusted proxies configured, got: %v", err)
	}

	_ = os.Setenv("HERASAT_TRUSTED_PROXIES", "invalid-proxy-cidr/99")
	err = InitNetworkSecurity()
	if err == nil {
		t.Fatalf("expected InitNetworkSecurity to fail on invalid trusted proxy CIDR")
	}
}

func TestIsKnownPlaceholderKey(t *testing.T) {
	placeholders := []string{
		"CHANGE_ME_TO_EXACTLY_32_BYTES_!",
		"01234567890123456789012345678901",
		"12345678901234567890123456789012",
		"00000000000000000000000000000000",
		"11111111111111111111111111111111",
		"abcdefghijklmnopqrstuvwxyz012345",
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	}
	for _, p := range placeholders {
		if !isKnownPlaceholderKey(p) {
			t.Errorf("expected placeholder key %q to be rejected", p)
		}
	}

	validKey := "8f14e45fceea167a5a36dedd4bea2543"
	if isKnownPlaceholderKey(validKey) {
		t.Errorf("expected valid key %q not to be rejected as placeholder", validKey)
	}
}

func TestSecureHeadersMiddlewareHardening(t *testing.T) {
	_ = os.Setenv("HERASAT_ALLOWED_ORIGINS", "https://admin.herasat.local,https://portal.herasat.gov")
	defer os.Unsetenv("HERASAT_ALLOWED_ORIGINS")

	handler := secureHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://admin.herasat.local")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	if xss := resp.Header.Get("X-XSS-Protection"); xss != "" {
		t.Errorf("expected deprecated X-XSS-Protection to be removed, got: %s", xss)
	}

	csp := resp.Header.Get("Content-Security-Policy")
	if !strings.Contains(csp, "connect-src 'self' https://admin.herasat.local https://portal.herasat.gov") {
		t.Errorf("expected CSP connect-src to include allowed origins, got: %s", csp)
	}
	if !strings.Contains(csp, "form-action 'self' https://admin.herasat.local https://portal.herasat.gov") {
		t.Errorf("expected CSP form-action to include allowed origins, got: %s", csp)
	}

	if vary := resp.Header.Get("Vary"); !strings.Contains(vary, "Origin") {
		t.Errorf("expected Vary: Origin when allowed origins configured, got: %s", vary)
	}

	if acao := resp.Header.Get("Access-Control-Allow-Origin"); acao != "https://admin.herasat.local" {
		t.Errorf("expected CORS allow origin header matching request origin, got: %s", acao)
	}

	optReqDisallowed := httptest.NewRequest("OPTIONS", "/api/admin/reports", nil)
	optReqDisallowed.Header.Set("Origin", "https://evil-attacker.com")
	optRecDisallowed := httptest.NewRecorder()
	handler.ServeHTTP(optRecDisallowed, optReqDisallowed)
	if optRecDisallowed.Code != http.StatusForbidden {
		t.Errorf("expected preflight OPTIONS from disallowed origin to be rejected with 403, got: %d", optRecDisallowed.Code)
	}

	optReqAllowed := httptest.NewRequest("OPTIONS", "/api/admin/reports", nil)
	optReqAllowed.Header.Set("Origin", "https://admin.herasat.local")
	optRecAllowed := httptest.NewRecorder()
	handler.ServeHTTP(optRecAllowed, optReqAllowed)
	if optRecAllowed.Code != http.StatusOK {
		t.Errorf("expected preflight OPTIONS from allowed origin to succeed with 200, got: %d", optRecAllowed.Code)
	}
}

func TestAdminIPWhitelistMiddlewareCached(t *testing.T) {
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

	_ = os.Setenv("HERASAT_ADMIN_ALLOWED_CIDRS", "192.168.10.0/24,127.0.0.1/32")
	_ = os.Setenv("HERASAT_ADMIN_ALLOW_ANY", "false")
	_ = os.Setenv("HERASAT_DEV_MODE", "0")
	_ = os.Setenv("HERASAT_TRUSTED_PROXIES", "")

	if err := InitNetworkSecurity(); err != nil {
		t.Fatalf("InitNetworkSecurity failed: %v", err)
	}

	dummyHandler := AdminIPWhitelistMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("admin-ok"))
	})

	allowedReq := httptest.NewRequest("POST", "/api/admin/login", nil)
	allowedReq.RemoteAddr = "192.168.10.45:43210"
	recAllowed := httptest.NewRecorder()
	dummyHandler.ServeHTTP(recAllowed, allowedReq)
	if recAllowed.Code != http.StatusOK {
		t.Errorf("expected allowed CIDR IP to pass, got: %d", recAllowed.Code)
	}

	blockedReq := httptest.NewRequest("POST", "/api/admin/login", nil)
	blockedReq.RemoteAddr = "203.0.113.88:12345"
	recBlocked := httptest.NewRecorder()
	dummyHandler.ServeHTTP(recBlocked, blockedReq)
	if recBlocked.Code != http.StatusForbidden {
		t.Errorf("expected blocked CIDR IP to get 403, got: %d", recBlocked.Code)
	}
}
