package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuditKeyDerivation(t *testing.T) {
	key := GetAuditHMACKey()
	if len(key) != 32 {
		t.Fatalf("expected 32-byte audit HMAC key, got %d bytes", len(key))
	}

	captchaKey := GetHMACKey()
	if bytes.Equal(key, captchaKey) {
		t.Fatalf("audit HMAC key should be domain-separated from captcha key")
	}
}

func TestAuditChainHMACComputation(t *testing.T) {
	now := time.Now().UTC()
	prev := genesisAuditHash
	target := int64(42)

	h1 := computeAuditHash(prev, "admin", "Support", "VIEW_REPORT", &target, "127.0.0.1", "detail_a", now)
	if len(h1) != 64 {
		t.Fatalf("expected 64-char hex string, got %d chars", len(h1))
	}

	h2 := computeAuditHash(prev, "admin", "Support", "VIEW_REPORT", &target, "127.0.0.1", "detail_b", now)
	if h1 == h2 {
		t.Fatalf("expected different hashes when detail differs")
	}

	h3 := computeAuditHash(prev, "hacker", "Support", "VIEW_REPORT", &target, "127.0.0.1", "detail_a", now)
	if h1 == h3 {
		t.Fatalf("expected different hashes when username differs")
	}

	h4 := computeAuditHash(h1, "admin", "Support", "VIEW_REPORT", &target, "127.0.0.1", "detail_a", now)
	if h1 == h4 {
		t.Fatalf("expected different hashes when prevHash differs")
	}
}

func TestNormalizeAndValidateUsername(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
	}{
		{"ValidSimple", "admin_herasat", "admin_herasat", false},
		{"ValidMixedCase", "Admin_Herasat", "admin_herasat", false},
		{"ValidWithHyphenAndDot", "edare.kol-01", "edare.kol-01", false},
		{"ValidTrimmed", "  admin123  ", "admin123", false},
		{"TooShort", "ab", "", true},
		{"TooLong", "a123456789012345678901234567890123", "", true},
		{"SpacesInside", "admin user", "", true},
		{"DisallowedSymbolAt", "admin@domain", "", true},
		{"DisallowedSlash", "admin/root", "", true},
		{"DisallowedBackslash", "admin\\root", "", true},
		{"HomoglyphCyrillic", "adm\u0456n", "", true},
		{"PersianLetters", "مدیر_حراست", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := normalizeAndValidateUsername(tc.input)
			if tc.expectErr && err == nil {
				t.Fatalf("expected error for input %q, got nil", tc.input)
			}
			if !tc.expectErr && err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}
			if !tc.expectErr && res != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, res)
			}
		})
	}
}

func TestAuditLegacyHashCompatibility(t *testing.T) {
	now := time.Now().UTC()
	prev := genesisAuditHash
	target := int64(10)

	legacyHash := computeLegacyAuditHash(prev, "admin", "Role", "LOGIN", &target, "1.2.3.4", now)
	if len(legacyHash) != 64 {
		t.Fatalf("expected 64-char legacy hash, got %d chars", len(legacyHash))
	}

	raw := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s", prev, "admin", "Role", "LOGIN", "10", "1.2.3.4", now.Format(time.RFC3339Nano))
	expected := sha256.Sum256([]byte(raw))
	if hex.EncodeToString(expected[:]) != legacyHash {
		t.Fatalf("legacy hash computation mismatch")
	}
}

func TestVerifyAuditLogsHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit-logs/verify", nil)
	rr := httptest.NewRecorder()

	VerifyAuditLogsHandler(rr, req)

	if rr.Code != http.StatusOK && rr.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected response status code: %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode json response: %v", err)
	}

	if _, ok := resp["status"]; !ok {
		t.Fatalf("response missing status field: %v", resp)
	}
}
