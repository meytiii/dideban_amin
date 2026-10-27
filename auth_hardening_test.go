package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"unicode/utf8"
)

func TestCompoundAccountLockoutByIP(t *testing.T) {
	username := "admin_target_test"
	ip1 := "10.0.0.1"
	ip2 := "10.0.0.2"

	Lockouts.Reset(username, ip1)
	Lockouts.Reset(username, ip2)

	for i := 0; i < 5; i++ {
		Lockouts.RecordFailure(username, ip1)
	}

	locked1, _ := Lockouts.IsLocked(username, ip1)
	if !locked1 {
		t.Fatalf("expected (username, ip1) to be locked after 5 failures")
	}

	locked2, _ := Lockouts.IsLocked(username, ip2)
	if locked2 {
		t.Fatalf("SECURITY FLAW: (username, ip2) must not be locked by ip1 failures")
	}

	Lockouts.Reset(username, ip1)
	unlocked1, _ := Lockouts.IsLocked(username, ip1)
	if unlocked1 {
		t.Fatalf("expected (username, ip1) to be unlocked after Reset")
	}
}

func TestMFAPendingTokenClientBindingAndReplay(t *testing.T) {
	InitCrypto("01234567890123456789012345678901")

	username := "mfa_bound_user"
	ip := "192.168.1.50"
	ua := "HerasatClient/2.0"

	token, err := GenerateMFAPendingToken(username, ip, ua)
	if err != nil {
		t.Fatalf("failed to generate bound MFA token: %v", err)
	}

	_, err = ValidateMFAPendingToken(token, "192.168.1.99", ua)
	if err == nil {
		t.Fatalf("expected validation to fail on IP mismatch")
	}

	_, err = ValidateMFAPendingToken(token, ip, "AttackerBrowser/1.0")
	if err == nil {
		t.Fatalf("expected validation to fail on User-Agent mismatch")
	}

	payload, err := ValidateMFAPendingToken(token, ip, ua)
	if err != nil {
		t.Fatalf("expected validation to succeed with matching IP and UA: %v", err)
	}
	if payload.Username != username {
		t.Fatalf("expected username %s, got %s", username, payload.Username)
	}

	_, err = ValidateMFAPendingToken(token, ip, ua)
	if err == nil {
		t.Fatalf("SECURITY VIOLATION: replaying single-use MFA token succeeded")
	}
}

func TestPasswordPolicyUnicodeAndRestrictions(t *testing.T) {
	if err := ValidatePasswordPolicy("رمزعبوربزرگ@A123"); err != nil {
		t.Fatalf("expected valid Unicode multi-byte password to pass, got: %v", err)
	}

	if utf8.RuneCountInString("رمزعبوربزرگ@A123") < 12 {
		t.Fatalf("expected rune count >= 12")
	}

	if err := ValidatePasswordPolicy("Herasat@2026Secure", "any_user"); err == nil {
		t.Fatalf("expected default Herasat password to be rejected")
	}

	if err := ValidatePasswordPolicy("@2026SecureP@ss!", "any_user"); err == nil {
		t.Fatalf("expected default 2026Secure password to be rejected")
	}

	if err := ValidatePasswordPolicy("AdminSpecial#2026", "any_user"); err == nil {
		t.Fatalf("expected password containing admin to be rejected")
	}

	if err := ValidatePasswordPolicy("MySecret#custom_admin#2026", "custom_admin"); err == nil {
		t.Fatalf("expected password containing username to be rejected")
	}

	if err := ValidatePasswordPolicy("VerySecure#Token99!", "different_user"); err != nil {
		t.Fatalf("expected valid compliant password to pass, got: %v", err)
	}
}

func TestRestrictedSessionMiddlewareEnforcement(t *testing.T) {
	sessionManager := NewSessionManager(nil)
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	token2FA := "test-session-token-2fa"
	sessionManager.Set(token2FA, AdminSession{
		Token:            token2FA,
		Username:         "officer1",
		Role:             "edare_kol",
		Requires2FASetup: true,
		CreatedAt:        time.Now(),
		LastActive:       time.Now(),
		Expiry:           time.Now().Add(15 * time.Minute),
	})

	testSessionsBackup := Sessions
	Sessions = sessionManager
	defer func() {
		Sessions = testSessionsBackup
	}()

	wrappedHandler := AdminAuthMiddleware(dummyHandler)

	reqReports := httptest.NewRequest("GET", "/api/admin/reports", nil)
	reqReports.AddCookie(&http.Cookie{Name: "herasat_admin_session", Value: token2FA})
	rrReports := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rrReports, reqReports)

	if rrReports.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for restricted 2FA session accessing /api/admin/reports, got %d", rrReports.Code)
	}

	var resp2FA map[string]string
	_ = json.Unmarshal(rrReports.Body.Bytes(), &resp2FA)
	if resp2FA["error"] != "2fa_setup_required" {
		t.Fatalf("expected error '2fa_setup_required', got: %v", resp2FA)
	}

	reqSetup := httptest.NewRequest("GET", "/api/admin/2fa/setup", nil)
	reqSetup.AddCookie(&http.Cookie{Name: "herasat_admin_session", Value: token2FA})
	rrSetup := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rrSetup, reqSetup)

	if rrSetup.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for restricted 2FA session accessing /api/admin/2fa/setup, got %d", rrSetup.Code)
	}

	tokenPwd := "test-session-token-pwd"
	sessionManager.Set(tokenPwd, AdminSession{
		Token:                  tokenPwd,
		Username:               "officer2",
		Role:                   "edare_kol",
		RequiresPasswordChange: true,
		CreatedAt:              time.Now(),
		LastActive:             time.Now(),
		Expiry:                 time.Now().Add(15 * time.Minute),
	})

	reqReportsPwd := httptest.NewRequest("GET", "/api/admin/reports", nil)
	reqReportsPwd.AddCookie(&http.Cookie{Name: "herasat_admin_session", Value: tokenPwd})
	rrReportsPwd := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rrReportsPwd, reqReportsPwd)

	if rrReportsPwd.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for password-change-required session, got %d", rrReportsPwd.Code)
	}

	var respPwd map[string]string
	_ = json.Unmarshal(rrReportsPwd.Body.Bytes(), &respPwd)
	if respPwd["error"] != "password_change_required" {
		t.Fatalf("expected error 'password_change_required', got: %v", respPwd)
	}

	reqChange := httptest.NewRequest("POST", "/api/admin/change-my-password", nil)
	reqChange.AddCookie(&http.Cookie{Name: "herasat_admin_session", Value: tokenPwd})
	rrChange := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rrChange, reqChange)

	if rrChange.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for password-change session accessing change-my-password, got %d", rrChange.Code)
	}
}

func TestInitialAdminPasswordGenerator(t *testing.T) {
	pwd := generateInitialPassword()
	if len(pwd) < 24 {
		t.Fatalf("expected generated initial password length >= 24, got %d", len(pwd))
	}
	if err := ValidatePasswordComplexity(pwd); err != nil {
		t.Fatalf("expected generated password to satisfy complexity policy, got: %v", err)
	}
}
