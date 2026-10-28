package main

import (
	"database/sql"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNonEnumeratingLoginLockout(t *testing.T) {
	fakeUser := "non_existent_probe_user_12345"
	ip := "192.168.10.50"

	Lockouts.Reset(fakeUser, ip)

	for i := 1; i <= 4; i++ {
		failures, justLocked := Lockouts.RecordFailure(fakeUser, ip)
		if failures != i {
			t.Fatalf("expected failures %d, got %d", i, failures)
		}
		if justLocked {
			t.Fatalf("should not be locked on attempt %d", i)
		}
	}

	failures, justLocked := Lockouts.RecordFailure(fakeUser, ip)
	if failures != 5 || !justLocked {
		t.Fatalf("expected 5 failures and justLocked=true on attempt 5, got %d, %v", failures, justLocked)
	}

	locked, remaining := Lockouts.IsLocked(fakeUser, ip)
	if !locked || remaining <= 0 {
		t.Fatalf("expected fake user to be locked out identically to valid users")
	}

	Lockouts.Reset(fakeUser, ip)
}

func TestLockoutDecay(t *testing.T) {
	user := "decay_test_user_xyz"
	ip := "10.0.0.99"

	Lockouts.Reset(user, ip)

	Lockouts.RecordFailure(user, ip)
	Lockouts.RecordFailure(user, ip)
	failures, _ := Lockouts.RecordFailure(user, ip)
	if failures != 3 {
		t.Fatalf("expected 3 failures, got %d", failures)
	}

	if db := Lockouts.getDB(); db != nil {
		_, err := db.Exec(`UPDATE admin_login_attempts SET updated_at = NOW() - INTERVAL '16 minutes' WHERE username = $1`, user)
		if err != nil {
			t.Fatalf("failed updating updated_at: %v", err)
		}
	}

	key := user + "@" + ip
	Lockouts.mu.Lock()
	if info, ok := Lockouts.accounts[key]; ok {
		info.cachedAt = time.Now().Add(-16 * time.Minute)
	}
	Lockouts.mu.Unlock()

	failuresAfterDecay, justLocked := Lockouts.RecordFailure(user, ip)
	if failuresAfterDecay != 1 {
		t.Fatalf("expected failure count to decay back to 1 after 15m, got %d", failuresAfterDecay)
	}
	if justLocked {
		t.Fatalf("decayed failure should not lock account")
	}

	Lockouts.Reset(user, ip)
}

func TestTOTPDatabaseReplayProtection(t *testing.T) {
	connStr := os.Getenv("HERASAT_DB_CONN")
	if connStr == "" {
		if os.Getenv("HERASAT_REQUIRE_DB_TESTS") == "1" {
			t.Fatalf("HERASAT_REQUIRE_DB_TESTS=1 but HERASAT_DB_CONN is unset")
		}
		t.Skip("Skipping DB test: HERASAT_DB_CONN unset")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	oldDB := DB
	DB = db
	defer func() { DB = oldDB }()

	secret, err := GenerateBase32Secret()
	if err != nil {
		t.Fatalf("failed generating secret: %v", err)
	}

	step := time.Now().Unix() / 30
	code, err := GenerateTOTPCode(secret, step)
	if err != nil {
		t.Fatalf("failed generating TOTP: %v", err)
	}

	username := fmt.Sprintf("totp_replay_user_%d", time.Now().UnixNano())

	valid, err := VerifyTOTPCode(username, secret, code)
	if !valid || err != nil {
		t.Fatalf("first verification failed: %v", err)
	}

	valid2, err2 := VerifyTOTPCode(username, secret, code)
	if valid2 || err2 != ErrTOTPReplayed {
		t.Fatalf("expected ErrTOTPReplayed on immediate replay, got valid=%v, err=%v", valid2, err2)
	}

	totpReplays.mu.Lock()
	totpReplays.used = make(map[string]time.Time)
	totpReplays.mu.Unlock()

	valid3, err3 := VerifyTOTPCode(username, secret, code)
	if valid3 || err3 != ErrTOTPReplayed {
		t.Fatalf("expected ErrTOTPReplayed from database used_mfa_nonces after memory wipe, got valid=%v, err=%v", valid3, err3)
	}
}

func TestBootstrapPasswordGeneration(t *testing.T) {
	pwd := generateInitialPassword()
	if len(pwd) < 20 {
		t.Fatalf("expected initial password length >= 20, got %d", len(pwd))
	}

	hasUpper := strings.ContainsAny(pwd, "ABCDEFGHJKLMNPQRSTUVWXYZ")
	hasLower := strings.ContainsAny(pwd, "abcdefghijkmnopqrstuvwxyz")
	hasDigit := strings.ContainsAny(pwd, "23456789")
	hasSymbol := strings.ContainsAny(pwd, "!@#$%^&*()-_=+")

	if !hasUpper || !hasLower || !hasDigit || !hasSymbol {
		t.Fatalf("initial password missing required character classes: upper=%v, lower=%v, digit=%v, sym=%v",
			hasUpper, hasLower, hasDigit, hasSymbol)
	}
}

func TestProxyHeaderTrustEnforcement(t *testing.T) {
	SetTrustedProxies(nil)

	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "203.0.113.10:4321"
	req1.Header.Set("X-Forwarded-For", "198.51.100.5")
	req1.Header.Set("X-Real-IP", "198.51.100.5")

	ip1 := getIP(req1)
	if ip1 != "203.0.113.10" {
		t.Fatalf("expected untrusted RemoteAddr 203.0.113.10, but got spoofed IP: %s", ip1)
	}

	trustedCIDRs := parseTrustedProxies("10.0.0.0/8,172.16.0.0/12")
	SetTrustedProxies(trustedCIDRs)
	defer SetTrustedProxies(nil)

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "10.0.0.1:4321"
	req2.Header.Set("X-Forwarded-For", "198.51.100.5, 10.0.0.2")

	ip2 := getIP(req2)
	if ip2 != "198.51.100.5" {
		t.Fatalf("expected extracted client IP 198.51.100.5 behind trusted proxy, got: %s", ip2)
	}
}
