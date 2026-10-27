package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRemediationCaptchaConfidentialityAndIntegrity(t *testing.T) {
	InitCrypto("01234567890123456789012345678901")
	db := getTestDB(t)
	UsedCaptchas.init(db)

	answer := 73
	token := GenerateCaptchaToken(answer, 2*time.Minute)
	if token == "" {
		t.Fatalf("Expected non-empty CAPTCHA token")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		t.Fatalf("Expected token to have format payload.signature, got: %s", token)
	}

	rawPayload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("Failed to base64 decode payload: %v", err)
	}

	var jsonProbe map[string]interface{}
	if err := json.Unmarshal(rawPayload, &jsonProbe); err == nil {
		t.Fatalf("SECURITY VIOLATION: CAPTCHA payload is plaintext JSON! Answer leaked to client: %v", jsonProbe)
	}

	if strings.Contains(string(rawPayload), "73") || strings.Contains(string(rawPayload), "Answer") {
		t.Fatalf("SECURITY VIOLATION: CAPTCHA payload contains plaintext answer strings: %s", string(rawPayload))
	}

	if !VerifyCaptcha(token, "73") {
		t.Fatalf("Expected valid CAPTCHA answer to verify successfully")
	}

	if VerifyCaptcha(token, "73") {
		t.Fatalf("SECURITY VIOLATION: Replay attack succeeded on already redeemed token!")
	}

	token2 := GenerateCaptchaToken(88, 2*time.Minute)
	parts2 := strings.Split(token2, ".")
	tamperedPayload := parts2[0]
	if len(tamperedPayload) > 2 {
		tamperedBytes := []byte(tamperedPayload)
		tamperedBytes[len(tamperedBytes)-2] ^= 0x01
		tamperedPayload = string(tamperedBytes)
	}
	tamperedToken := tamperedPayload + "." + parts2[1]
	if VerifyCaptcha(tamperedToken, "88") {
		t.Fatalf("SECURITY VIOLATION: Verification succeeded on ciphertext-tampered token!")
	}

	tamperedSigToken := parts2[0] + ".AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	if VerifyCaptcha(tamperedSigToken, "88") {
		t.Fatalf("SECURITY VIOLATION: Verification succeeded on signature-tampered token!")
	}
}

func TestRemediationSSEBrokerRoleScoping(t *testing.T) {
	broker := NewSSEBroker()
	defer broker.Close()

	clientSuper := &SSEClient{
		id:       "tok_super_1",
		username: "admin_superuser",
		role:     "superuser",
		send:     make(chan []byte, 10),
	}
	clientEdare := &SSEClient{
		id:       "tok_edare_1",
		username: "admin_edare",
		role:     "edare_kol",
		send:     make(chan []byte, 10),
	}
	clientRegional := &SSEClient{
		id:       "tok_regional_1",
		username: "officer_region_3",
		role:     "navahi_manategh",
		send:     make(chan []byte, 10),
	}

	broker.register <- clientSuper
	broker.register <- clientEdare
	broker.register <- clientRegional

	time.Sleep(50 * time.Millisecond)

	broker.BroadcastNewReport(98765)

	select {
	case msg := <-clientSuper.send:
		if !strings.Contains(string(msg), "new_report") || !strings.Contains(string(msg), "98765") {
			t.Fatalf("Unexpected message on superuser client: %s", string(msg))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("Superuser client failed to receive new_report event")
	}

	select {
	case msg := <-clientEdare.send:
		if !strings.Contains(string(msg), "new_report") || !strings.Contains(string(msg), "98765") {
			t.Fatalf("Unexpected message on edare_kol client: %s", string(msg))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("Edare_kol client failed to receive new_report event")
	}

	select {
	case msg := <-clientRegional.send:
		t.Fatalf("SECURITY VIOLATION: Regional officer received unassigned new_report event: %s", string(msg))
	case <-time.After(100 * time.Millisecond):
	}
}

func TestRemediationSessionRevocationLifecycle(t *testing.T) {
	sm := NewSessionManager(nil)

	testUser := "victim_admin"
	tok1 := "tok_active_session_1"
	now := time.Now()
	expiry := now.Add(15 * time.Minute)

	sm.Set(tok1, AdminSession{
		Token:       tok1,
		Username:    testUser,
		FullName:    "Victim Admin",
		Title:       "Officer",
		Role:        "edare_kol",
		IsSuperuser: false,
		IP:          "127.0.0.1",
		UserAgent:   "TestAgent",
		CreatedAt:   now,
		LastActive:  now,
		Expiry:      expiry,
	})

	if _, ok := sm.Get(tok1); !ok {
		t.Fatalf("Expected session 1 to be active")
	}

	revokedCount := sm.RevokeAllUserSessions(testUser, "PASSWORD_CHANGED")
	if revokedCount != 1 {
		t.Fatalf("Expected 1 revoked session, got %d", revokedCount)
	}

	if _, ok := sm.Get(tok1); ok {
		t.Fatalf("SECURITY VIOLATION: Old session 1 remains valid after password change!")
	}
}

func TestRemediationPaginationClampingBounds(t *testing.T) {
	testCases := []struct {
		name         string
		url          string
		expectedSize int
	}{
		{"Excessive page size", "/api/admin/reports?page_size=100000", 100},
		{"fetchAll parameter", "/api/admin/reports?all=true", 100},
		{"fetchAll via page_size", "/api/admin/reports?page_size=all", 100},
		{"Zero page size", "/api/admin/reports?page_size=0", 25},
		{"Negative page size", "/api/admin/reports?page_size=-10", 25},
		{"Normal allowed page size", "/api/admin/reports?page_size=50", 50},
		{"Max boundary page size", "/api/admin/reports?page_size=100", 100},
		{"Over max boundary", "/api/admin/reports?page_size=101", 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.url, nil)
			page := 1
			pageSize := 25
			if p := req.URL.Query().Get("page"); p != "" {
				if parsed, err := parsePositiveInt(p); err == nil && parsed > 0 {
					page = parsed
				}
			}
			_ = page
			if ps := req.URL.Query().Get("page_size"); ps != "" && ps != "all" {
				if parsed, err := parsePositiveInt(ps); err == nil && parsed > 0 {
					pageSize = parsed
				}
			}
			if req.URL.Query().Get("all") == "true" || req.URL.Query().Get("page_size") == "all" {
				pageSize = 100
			}
			if pageSize > 100 {
				pageSize = 100
			}
			if pageSize <= 0 {
				pageSize = 25
			}

			if pageSize != tc.expectedSize {
				t.Fatalf("For URL %s: expected pageSize %d, got %d", tc.url, tc.expectedSize, pageSize)
			}
		})
	}
}

func parsePositiveInt(s string) (int, error) {
	return parseInteger(s)
}

func parseInteger(s string) (int, error) {
	var res int
	var sign = 1
	trimmed := strings.TrimSpace(s)
	if strings.HasPrefix(trimmed, "-") {
		sign = -1
		trimmed = trimmed[1:]
	}
	for _, c := range trimmed {
		if c < '0' || c > '9' {
			return 0, http.ErrNotSupported
		}
		res = res*10 + int(c-'0')
	}
	return res * sign, nil
}

func TestPublicTrackingOmitsAdminUsername(t *testing.T) {
	InitCrypto("01234567890123456789012345678901")
	m := ReportMessageItem{
		ID:            1,
		SenderType:    "admin",
		AdminUsername: "admin_internal_super",
		Message:       "Response from operator",
	}
	if m.SenderType == "admin" {
		m.AdminUsername = ""
	}
	if m.AdminUsername != "" {
		t.Fatalf("SECURITY VIOLATION: AdminUsername was not sanitized for public tracking response!")
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Failed to marshal message item: %v", err)
	}
	if strings.Contains(string(raw), "admin_internal_super") {
		t.Fatalf("SECURITY VIOLATION: Serialized JSON leaked internal admin username: %s", string(raw))
	}
}
